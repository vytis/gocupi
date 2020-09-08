/*
  Gocupi Arduino Code
  Reads movement commands over serial and controls two stepper motors 
*/
#include <SPI.h>
#include <stdint.h>
#include "TMC_API.h"
#include "TMC5072.h"

// #define TMC5072_FIELD_READ(buffer, address, mask, shift) \
// 	FIELD_GET(tmc5072_readInt(address), mask, shift)
// #define TMC5072_FIELD_WRITE(buffer, address, mask, shift, value) \
// 	(tmc5072_writeInt(buffer, address, FIELD_SET(tmc5072_readInt(address), mask, shift, value)))


// comment out to disable PENUP support
#define ENABLE_PENUP

// Constants and global variables
// --------------------------------------
const int LED_PINS_COUNT = 2;
const int LED_PINS[LED_PINS_COUNT] = {
  16,17}; // the pins of all of the leds, first 3 are status lights, 5th is receive indicator

const int PENUP_SERVO_PIN = 2;
const int INT_PIN = 3; // ENC1A
const int INT_PP_PIN = 4; // ENC1B

const int LEFT_STEP_PIN = 5; // ENC1N

const int LEFT_DIR_PIN = 6; // ENC2A
const int RIGHT_DIR_PIN = 7; // ENC2B
const int RIGHT_STEP_PIN = 8; // ENC2N

const int MOTOR_ENABLE = 9;

const int chipCS = 10;
const int SDI = 11;
const int SDO = 12;
const int CLOCK = 13; //SCK

#ifdef ENABLE_PENUP
#include <Servo.h>
Servo penUpServo;
char penTransitionDirection; // -1, 0, 1
const long PENUP_TRANSITION_US = 524288; // time to go from pen up to down, or down to up
const int PENUP_TRANSITION_US_LOG = 19; // 2^19 = 524288
const long PENUP_COOLDOWN_US = 650000;
const long PENUP_ANGLE = 140;
const long PENDOWN_ANGLE = 40;
#endif

const unsigned int TIME_SLICE_US = 1024; // number of microseconds per time step
const unsigned int TIME_SLICE_US_LOG = 10; // log base 2 of TIME_SLICE_US
const unsigned int POS_FACTOR = 32; // fixed point factor each position is multiplied by
const unsigned int POS_FACTOR_LOG = 5; // log base 2 of POS_FACTOR, used after multiplying two fixed point numbers together

const char RESET_COMMAND = 0x80; // -128, command to reset
const char PENUP_COMMAND = 0x81; // -127, command to lift pen
const char PENDOWN_COMMAND = 0x7F; // 127, command to lower pen

const unsigned int MOVE_DATA_CAPACITY = 1024;
char moveData[MOVE_DATA_CAPACITY]; // buffer of move data, circular buffer
unsigned int moveDataStart = 0; // where data is currently being read from
unsigned int moveDataLength = 0; // the number of items in the moveDataBuffer
unsigned int moveDataRequestPending = 0; // number of bytes requested

char leftDelta, rightDelta; // delta in the current slice
long leftStartPos, rightStartPos; // start position for this slice
long leftCurPos, rightCurPos; // current position of the spools

unsigned long curTime; // current time in microseconds
unsigned long sliceStartTime; // start of current slice in microseconds

// typedef TMC5072TypeDef TMC5072;
// typedef ConfigurationTypeDef Config;

// TMC5072TypeDef state = {0};

uint8_t buffer[4];

void tmc5072_readWriteArray(uint8_t channel, uint8_t *data, size_t length) {
  //TMC5130 takes 40 bit data: 8 address and 32 data
  digitalWrite(chipCS,LOW);

  // for (size_t i = 0; i < length; i++)
  // {
  //   data[i] = SPI.transfer(data[i]);  
  // }


  // i_datagram |= SPI.transfer(datagram.value[0]);
  // i_datagram <<= 8;
  // i_datagram |= SPI.transfer(datagram.value[1]);
  // i_datagram <<= 8;
  // i_datagram |= SPI.transfer(datagram.value[2]);
  // i_datagram <<= 8;
  // i_datagram |= SPI.transfer(datagram.value[3]);
  SPI.beginTransaction(SPISettings(4000000, MSBFIRST, SPI_MODE3));
  SPI.transfer(data, length);
  SPI.endTransaction();
  digitalWrite(chipCS,HIGH);
}

// setup
// --------------------------------------
void setup() {
  Serial.begin(57600);
  Serial.setTimeout(0);

  pinMode(chipCS,OUTPUT);
  pinMode(MOTOR_ENABLE,OUTPUT);
  digitalWrite(chipCS,HIGH);
  digitalWrite(MOTOR_ENABLE,HIGH);

  SPI.setBitOrder(MSBFIRST);
  SPI.setClockDivider(SPI_CLOCK_DIV32);
  SPI.setDataMode(SPI_MODE3);
  SPI.begin();

  setupMotors();

  // setup pins
  for(int ledIndex = 0; ledIndex < LED_PINS_COUNT; ledIndex++) {
    pinMode(LED_PINS[ledIndex], OUTPUT);
    digitalWrite(LED_PINS[ledIndex], HIGH);
  }	
  pinMode(LEFT_STEP_PIN, OUTPUT);
  pinMode(LEFT_DIR_PIN, OUTPUT);
  pinMode(RIGHT_STEP_PIN, OUTPUT);
  pinMode(RIGHT_DIR_PIN, OUTPUT);	

#ifdef ENABLE_PENUP
  penUpServo.attach(PENUP_SERVO_PIN);
  penUpServo.write(PENUP_ANGLE);
  delay(1000);
  penUpServo.write(PENDOWN_ANGLE);
  delay(1000);
  penUpServo.write(PENUP_ANGLE);
#endif  

  ResetMovementVariables();

  delay(500);
  UpdateReceiveLed(false);
  UpdateStatusLeds(0);
  digitalWrite(MOTOR_ENABLE,LOW);

}

void setupMotors() 
{
  // tmc5072_init(buffer, 0, state.config, tmc5072_defaultRegisterResetState);

  

  int32_t gconf = 0;
  FIELD_SET(gconf, TMC5072_STEPDIR1_ENABLE_MASK, TMC5072_STEPDIR1_ENABLE_SHIFT, 1l);
  FIELD_SET(gconf, TMC5072_STEPDIR2_ENABLE_MASK, TMC5072_STEPDIR2_ENABLE_SHIFT, 1l);
  tmc5072_writeInt(TMC5072_GCONF, gconf);

  int32_t chop = 0;
  FIELD_SET(chop, TMC5072_TOFF_MASK, TMC5072_TOFF_SHIFT, 5l);
  FIELD_SET(chop, TMC5072_HSTRT_MASK, TMC5072_HSTRT_SHIFT, 0l);
  FIELD_SET(chop, TMC5072_HEND_MASK, TMC5072_HEND_SHIFT, 13l);
  FIELD_SET(chop, TMC5072_CHM_MASK, TMC5072_CHM_SHIFT, 0l);
  FIELD_SET(chop, TMC5072_RNDTF_MASK, TMC5072_RNDTF_SHIFT, 1l);
  FIELD_SET(chop, TMC5072_TBL_MASK, TMC5072_TBL_SHIFT, 1l);
  FIELD_SET(chop, TMC5072_MRES_MASK, TMC5072_MRES_SHIFT, 6l);

  int32_t ihold = 0;
  FIELD_SET(ihold, TMC5072_IHOLD_MASK, TMC5072_IHOLD_SHIFT, 1);
  FIELD_SET(ihold, TMC5072_IRUN_MASK, TMC5072_IRUN_SHIFT, 20);
  FIELD_SET(ihold, TMC5072_IHOLDDELAY_MASK, TMC5072_IHOLDDELAY_SHIFT, 6);

  int32_t zerowait = 0;
  FIELD_SET(zerowait, TMC5072_TZEROWAIT_MASK, TMC5072_TZEROWAIT_SHIFT, 10000);

  int32_t pwmconf = 0;
  FIELD_SET(pwmconf, TMC5072_PWM_FREQ_MASK, TMC5072_PWM_FREQ_SHIFT, 2);
  FIELD_SET(pwmconf, TMC5072_FREEWHEEL_MASK, TMC5072_FREEWHEEL_SHIFT, 0);

  for (size_t i = 0; i < 2; i++) {
    tmc5072_writeInt(TMC5072_CHOPCONF(i), chop);
    tmc5072_writeInt(TMC5072_IHOLD_IRUN(i), ihold);
    tmc5072_writeInt(TMC5072_TZEROWAIT(i), zerowait);
    tmc5072_writeInt(TMC5072_PWMCONF(i), pwmconf);
  }
}

// void sendData(byte address, Data datagram) {
//   //TMC5130 takes 40 bit data: 8 address and 32 data

// //  delay(100);
//   unsigned long i_datagram;

//   digitalWrite(chipCS,LOW);
//   delayMicroseconds(10);
//   Serial.print(datagram.value[0], HEX);
//   Serial.print(datagram.value[1], HEX);
//   Serial.print(datagram.value[2], HEX);
//   Serial.print(datagram.value[3], HEX);

//   SPI.transfer(address + 0x80);  

//   i_datagram |= SPI.transfer(datagram.value[0]);
//   i_datagram <<= 8;
//   i_datagram |= SPI.transfer(datagram.value[1]);
//   i_datagram <<= 8;
//   i_datagram |= SPI.transfer(datagram.value[2]);
//   i_datagram <<= 8;
//   i_datagram |= SPI.transfer(datagram.value[3]);

//   delayMicroseconds(10);
//   digitalWrite(chipCS,HIGH);
// }




// void readData(unsigned long address) {
//   //TMC5130 takes 40 bit data: 8 address and 32 data

// //  delay(100);
//   unsigned long i_datagram;

//   digitalWrite(chipCS,LOW);
// //  delayMicroseconds(10);

//   SPI.transfer(address);
//   SPI.transfer(0); SPI.transfer(0); SPI.transfer(0); SPI.transfer(0);
//   delayMicroseconds(10);

  
  
//   SPI.transfer(address);

//   i_datagram |= SPI.transfer(0);
//   i_datagram <<= 8;
//   i_datagram |= SPI.transfer(0);
//   i_datagram <<= 8;
//   i_datagram |= SPI.transfer(0);
//   i_datagram <<= 8;
//   i_datagram |= SPI.transfer(0);
//   digitalWrite(chipCS,HIGH);

// //  Serial.print("Received: ");
// //  Serial.print(i_datagram,HEX);
// //  Serial.print(" from register: ");
// //  Serial.println(address,HEX);
// }

// Reset all movement variables
// --------------------------------------
void ResetMovementVariables()
{
  leftDelta = rightDelta = leftStartPos = rightStartPos = leftCurPos = rightCurPos = 0;
  sliceStartTime = curTime;

#ifdef ENABLE_PENUP
  penTransitionDirection = 0;
  penUpServo.write(PENUP_ANGLE);
#endif  
}

// Main execution loop
// --------------------------------------
void loop() {
  curTime = micros();
  if (curTime < sliceStartTime) { // protect against 70 minute overflow
    sliceStartTime = 0;
  }

  long curSliceTime = curTime - sliceStartTime;

#ifdef ENABLE_PENUP
  if (penTransitionDirection) {
    UpdatePenTransition(curSliceTime);
    if (!penTransitionDirection) {
      sliceStartTime = curTime;
    }
  } else {	
#endif
    // move to next slice if necessary
    while(curSliceTime > TIME_SLICE_US) {
      SetSliceVariables();
      curSliceTime -= TIME_SLICE_US;
      sliceStartTime += TIME_SLICE_US;

#ifdef ENABLE_PENUP	
      if (penTransitionDirection) {
        sliceStartTime = curTime;
        return;
      }
#endif      
    }
	
    UpdateStepperPins(curSliceTime);
#ifdef ENABLE_PENUP    
  }
#endif  

  ReadSerialMoveData();
  RequestMoreSerialMoveData();
}

// Update stepper pins
// --------------------------------------
void UpdateStepperPins(long curSliceTime) {
  long leftTarget = ((long(leftDelta) * curSliceTime) >> TIME_SLICE_US_LOG) + leftStartPos;
  long rightTarget = ((long(rightDelta) * curSliceTime) >> TIME_SLICE_US_LOG) + rightStartPos;

  int leftSteps = (leftTarget - leftCurPos) >> POS_FACTOR_LOG;
  int rightSteps = (rightTarget - rightCurPos) >> POS_FACTOR_LOG;

  boolean leftPositiveDir = true;
  if (leftSteps < 0) {
    leftPositiveDir = false;
    leftSteps = -leftSteps;
  }
  boolean rightPositiveDir = true;
  if (rightSteps < 0) {
    rightPositiveDir = false;
    rightSteps = -rightSteps;
  }

  do {
    if (leftSteps) {
      Step(LEFT_STEP_PIN, LEFT_DIR_PIN, leftPositiveDir);
      if (leftPositiveDir) {
        leftCurPos += POS_FACTOR;
      } else {
        leftCurPos -= POS_FACTOR;
      }
      leftSteps--;
      
      UpdateStatusLeds(leftCurPos >> 13);
    }

    if (rightSteps) {
      Step(RIGHT_STEP_PIN, RIGHT_DIR_PIN, rightPositiveDir);
      if (rightPositiveDir) {
        rightCurPos += POS_FACTOR;
      } else {
        rightCurPos -= POS_FACTOR;
      }
      rightSteps--;
    }

    if (leftSteps || rightSteps) {
//      delayMicroseconds(50); // delay a small amount of time before refiring the steps to smooth things out
    } else {
      break;
    }
  } while(true);
}

// Update pen position
// --------------------------------------
#ifdef ENABLE_PENUP
void UpdatePenTransition(long curSliceTime) {
	
  //int targetAngle = ((float)(PENDOWN_ANGLE - PENUP_ANGLE) * ((float)curSliceTime / (float)PENUP_TRANSITION_US)) + PENUP_ANGLE;
  //if (targetAngle > PENDOWN_ANGLE) {
    //targetAngle = PENDOWN_ANGLE;
    
    if (curSliceTime > PENUP_COOLDOWN_US) {
      penTransitionDirection = 0; // are done moving the pen servo
    }
  //}

  if (penTransitionDirection == 1) {
	//targetAngle = 180 - targetAngle;
    penUpServo.write(PENUP_ANGLE);
  } else if (penTransitionDirection == -1) {
    penUpServo.write(PENDOWN_ANGLE);
  }

  //penUpServo.write(targetAngle);
}
#endif

// Update status leds
// --------------------------------------
void UpdateStatusLeds(int value) {
  // output the time to the leds in binary
  digitalWrite(LED_PINS[0], value & 0x1);
  digitalWrite(LED_PINS[1], value & 0x2);
  digitalWrite(LED_PINS[2], value & 0x4);
}

// Update receive leds
// --------------------------------------
void UpdateReceiveLed(boolean value) {
  digitalWrite(LED_PINS[3], value);
}

// Step
// --------------------------------------
void Step(int stepPin, int dirPin, boolean dir) {
  digitalWrite(dirPin, dir);

  digitalWrite(stepPin, LOW);
  digitalWrite(stepPin, HIGH);
}

// Set all variables based on the data currently in the buffer
// --------------------------------------
void SetSliceVariables() {
  // set slice start pos to previous slice start plus previous delta
  leftStartPos = leftStartPos + long(leftDelta);
  rightStartPos = rightStartPos + long(rightDelta);

  if (moveDataLength < 2) {
    leftDelta = rightDelta = 0;
  } else {
    leftDelta = MoveDataGet();
    rightDelta = MoveDataGet();
    
#ifdef ENABLE_PENUP	
    if (leftDelta == PENUP_COMMAND) {
      leftDelta = rightDelta = 0;
      penTransitionDirection = 1;
    } else if (leftDelta == PENDOWN_COMMAND) {
      leftDelta = rightDelta = 0;
      penTransitionDirection = -1;
    }
#else
    if (leftDelta == PENUP_COMMAND || leftDelta == PENDOWN_COMMAND) {
       leftDelta = rightDelta = 0;
    }
#endif    
  }
}                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    

// Stop everything and blink the status led value times
// --------------------------------------
void Blink(char value) {
 int counts = value;
  if (counts<0) counts=-counts;

  UpdateReceiveLed(false);
  for(int i=0;i<counts;i++) {
   delay(1000);
   UpdateReceiveLed(true);
   delay(1000);
   UpdateReceiveLed(false);
    
  }
  delay(100000);
}

// Read serial data if its available
// --------------------------------------
void ReadSerialMoveData() {     

  if(Serial.available()) {
    char value = Serial.read();
    
    // Check if this value is the sentinel reset value
    if (value == RESET_COMMAND) {
      ResetMovementVariables();
      moveDataRequestPending = 0;
      moveDataLength = 0;
      UpdateReceiveLed(false);
//      digitalWrite(MOTOR_ENABLE,HIGH);

      return;
    }
//    digitalWrite(MOTOR_ENABLE,LOW);


    MoveDataPut(value);
    moveDataRequestPending--;

    if (!moveDataRequestPending) {
      UpdateReceiveLed(false);
    }
  }
}

// Put a value onto the end of the move data buffer
// --------------------------------------
void MoveDataPut(char value) {
  int writePosition = moveDataStart + moveDataLength;
  if (writePosition >= MOVE_DATA_CAPACITY) {
    writePosition = writePosition - MOVE_DATA_CAPACITY;
  }

  moveData[writePosition] = value;

  if (moveDataLength == MOVE_DATA_CAPACITY) { // full, overwrite existing data
    moveDataStart++;
    if (moveDataStart == MOVE_DATA_CAPACITY) {
      moveDataStart = 0;
    }
  } 
  else {
    moveDataLength++;
  }
}

// Return a piece of data sitting in the moveData buffer, removing it from the buffer
// --------------------------------------
char MoveDataGet() {
  if (moveDataLength == 0) {
    return 0;
  }

  char result = moveData[moveDataStart];
  moveDataStart++;
  if (moveDataStart == MOVE_DATA_CAPACITY) {
    moveDataStart = 0;
  }
  moveDataLength--;

  return result;
}

// Return the amount of data sitting in the moveData buffer
// --------------------------------------
void RequestMoreSerialMoveData() {
  if (moveDataRequestPending > 0 || MOVE_DATA_CAPACITY - moveDataLength < 128)
    return;

  // request 128 bytes of data
  Serial.write(128);
  moveDataRequestPending = 128;
  UpdateReceiveLed(true);
}


// ----------------
int32_t tmc5072_writeDatagram(uint8_t address, uint8_t x1, uint8_t x2, uint8_t x3, uint8_t x4)
{
	uint8_t data[5] = { address | TMC5072_WRITE_BIT, x1, x2, x3, x4 };
	tmc5072_readWriteArray(0, &data[0], 5);

	int32_t value = ((uint32_t)x1 << 24) | ((uint32_t)x2 << 16) | (x3 << 8) | x4;
  return value;
	// Write to the shadow register and mark the register dirty
	// address = TMC_ADDRESS(address);
	// tmc5072->config->shadowRegister[address] = value;
	// tmc5072->registerAccess[address] |= TMC_ACCESS_DIRTY;
}

void tmc5072_writeInt(uint8_t address, int32_t value)
{
	tmc5072_writeDatagram(address, BYTE(value, 3), BYTE(value, 2), BYTE(value, 1), BYTE(value, 0));
}

int32_t tmc5072_readInt(uint8_t address)
{
	address = TMC_ADDRESS(address);

	// register not readable -> shadow register copy
	// if(!TMC_IS_READABLE(tmc5072->registerAccess[address]))
	// 	return tmc5072->config->shadowRegister[address];

	uint8_t data[5] = { 0, 0, 0, 0, 0 };

	data[0] = address;
	tmc5072_readWriteArray(0, &data[0], 5);

	data[0] = address;
	tmc5072_readWriteArray(0, &data[0], 5);

	return ((uint32_t)data[1] << 24) | ((uint32_t)data[2] << 16) | (data[3] << 8) | data[4];
}