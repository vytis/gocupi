/*
  Gocupi Arduino Code
  Reads movement commands over serial and controls two stepper motors 
*/
#include <SPI.h>
#include "TMC5072_register.h"
#include <stddef.h>
#include <stdbool.h>
#include <stdint.h>

// comment out to disable PENUP support
#define ENABLE_PENUP

// Constants and global variables
// --------------------------------------
const int LED_PINS_COUNT = 2;
const int LED_PINS[LED_PINS_COUNT] = {
  2,3}; // the pins of all of the leds, first 3 are status lights, 5th is receive indicator
const int LEFT_STEP_PIN = 6;
const int LEFT_DIR_PIN = 7;
const int RIGHT_STEP_PIN = 8;
const int RIGHT_DIR_PIN = 9;
const int chipCS = 10;
const int MOTOR_ENABLE = 5;

#ifdef ENABLE_PENUP
#include <Servo.h>
Servo penUpServo;
char penTransitionDirection; // -1, 0, 1
const int PENUP_SERVO_PIN = 4;
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

typedef struct Data {
    byte value[4];
} Data;


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
  const unsigned long MSTEPS_256 = 0x0;
  const unsigned long MSTEPS_128 = 0x1;
  const unsigned long MSTEPS_64  = 0x2;
  const unsigned long MSTEPS_32  = 0x3;
  const unsigned long MSTEPS_16  = 0x4;
  const unsigned long MSTEPS_8   = 0x5;
  const unsigned long MSTEPS_4   = 0x6;
  const unsigned long MSTEPS_2   = 0x7;
  const unsigned long MSTEPS_1   = 0x8;

  const Data chop = {0x10 + MSTEPS_16, 0x1, 0x0, 0xC5};
  const Data ihold = {0x0, 0x06, 0x1F, 0x0};
  const Data zerowait = {0x0, 0x0, 0x27, 0x10};
  const Data pwm = {0x01, 0x20, 0x0, 0x0};
  
  // m1 to step/dir
  sendData(TMC5072_CHOPCONF_1,  chop);
  sendData(TMC5072_IHOLD_IRUN_1, ihold);
  sendData(TMC5072_TZEROWAIT_1, zerowait);
  sendData(TMC5072_PWMCONF_1, pwm);


  sendData(TMC5072_CHOPCONF_2,  chop);
  sendData(TMC5072_IHOLD_IRUN_2, ihold);
  sendData(TMC5072_TZEROWAIT_2, zerowait);
  sendData(TMC5072_PWMCONF_2, pwm);

  sendData(TMC5072_GCONF, {0x0, 0x0, 0x0, 0x6});
}

void sendData(byte address, Data datagram) {
  //TMC5130 takes 40 bit data: 8 address and 32 data

//  delay(100);
  unsigned long i_datagram;

  digitalWrite(chipCS,LOW);
  delayMicroseconds(10);
  Serial.print(datagram.value[0], HEX);
  Serial.print(datagram.value[1], HEX);
  Serial.print(datagram.value[2], HEX);
  Serial.print(datagram.value[3], HEX);

  SPI.transfer(address + 0x80);  

  i_datagram |= SPI.transfer(datagram.value[0]);
  i_datagram <<= 8;
  i_datagram |= SPI.transfer(datagram.value[1]);
  i_datagram <<= 8;
  i_datagram |= SPI.transfer(datagram.value[2]);
  i_datagram <<= 8;
  i_datagram |= SPI.transfer(datagram.value[3]);

  delayMicroseconds(10);
  digitalWrite(chipCS,HIGH);
}

void readData(unsigned long address) {
  //TMC5130 takes 40 bit data: 8 address and 32 data

//  delay(100);
  unsigned long i_datagram;

  digitalWrite(chipCS,LOW);
//  delayMicroseconds(10);

  SPI.transfer(address);
  SPI.transfer(0); SPI.transfer(0); SPI.transfer(0); SPI.transfer(0);
  delayMicroseconds(10);

  
  
  SPI.transfer(address);

  i_datagram |= SPI.transfer(0);
  i_datagram <<= 8;
  i_datagram |= SPI.transfer(0);
  i_datagram <<= 8;
  i_datagram |= SPI.transfer(0);
  i_datagram <<= 8;
  i_datagram |= SPI.transfer(0);
  digitalWrite(chipCS,HIGH);

//  Serial.print("Received: ");
//  Serial.print(i_datagram,HEX);
//  Serial.print(" from register: ");
//  Serial.println(address,HEX);
}

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