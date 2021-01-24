package polargraph

// Handles sending data over serial to the arduino

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	serial "github.com/tarm/goserial"
)

// Output the coordinates to the screen
func OutputCoords(plotCoords <-chan Coordinate) {

	for coord := range plotCoords {
		fmt.Println(coord)
	}

	fmt.Println("Done plotting")
}

// Takes in coordinates and outputs stepData
func GenerateSteps(plotCoords <-chan Coordinate, stepData chan<- int8) {

	defer close(stepData)

	polarSystem := PolarSystemFromSettings()
	previousPolarPos := PolarCoordinate{LeftDist: Settings.StartingLeftDist_MM, RightDist: Settings.StartingRightDist_MM}
	startingLocation := previousPolarPos.ToCoord(polarSystem)

	fmt.Println("Start Location", startingLocation, "Initial Polar", previousPolarPos)

	if startingLocation.IsNaN() {
		panic(fmt.Sprint("Starting location is not a valid number, setup has impossible values"))
	}

	// setup 0,0 as the initial location of the plot head
	polarSystem.XOffset = startingLocation.X
	polarSystem.YOffset = startingLocation.Y

	//var interp PositionInterpolater = new(LinearInterpolater)
	var interp = new(TrapezoidInterpolater)

	target, chanOpen := <-plotCoords
	if !chanOpen {
		return
	}
	origin := target

	var currentPenUp bool = true // arduino code defaults to pen up on ResetCommand
	var anotherTarget bool = true

	for anotherTarget {
		nextTarget, chanOpen := <-plotCoords
		if !chanOpen {
			anotherTarget = false
			nextTarget = target
		}

		if target.PenUp != currentPenUp {
			// send twice in order to preserve alignment of always sending 2 values at a time over serial
			if target.PenUp {
				stepData <- PenUpCommand
				stepData <- PenUpCommand
			} else {
				stepData <- PenDownCommand
				stepData <- PenDownCommand
			}
			currentPenUp = target.PenUp
		}

		interp.Setup(origin, target, nextTarget)

		//fmt.Println("Slices", interp.Slices(), "------------------------")

		for slice := 1.0; slice <= interp.Slices(); slice++ {

			sliceTarget := interp.Position(slice)
			polarSliceTarget := sliceTarget.ToPolar(polarSystem)

			// calc number of steps that will be made this time slice, have to precision that can be sent in a single value from StepsMaxValue to -StepsMaxValue
			sliceSteps := polarSliceTarget.
				Minus(previousPolarPos).
				Scaled(StepsFixedPointFactor/Settings.StepSize_MM).
				Ceil().
				Clamp(StepsMaxValue, -StepsMaxValue)
			previousPolarPos = previousPolarPos.
				Add(sliceSteps.Scaled(Settings.StepSize_MM / StepsFixedPointFactor))

			stepData <- int8(-sliceSteps.LeftDist)
			stepData <- int8(sliceSteps.RightDist)
		}
		origin = target
		target = nextTarget
	}
	fmt.Println("Done generating steps")
}

// Count steps
func CountSteps(stepData <-chan int8) {

	sliceCount := 0
	penTransition := 0
	for step := range stepData {

		if step == PenUpCommand || step == PenDownCommand {
			penTransition++
		} else {
			sliceCount++
		}
	}
	// since data is sent once for left and right spools, have to divide by 2
	sliceCount = sliceCount >> 1
	penTransition = penTransition >> 1
	penTransitionCooldown_US := 650000.0 // as defined in the microcontroller code
	fmt.Println("Steps", sliceCount, "Pen Transitions", penTransition, "Time", time.Duration(float64(sliceCount)*TimeSlice_US+float64(penTransition)*penTransitionCooldown_US)*time.Microsecond)
}

// Sends the given stepData to the stepper driver
func WriteStepsToSerial(stepData <-chan int8) {

	fmt.Println("Opening com port ", Settings.SerialPortPath)
	c := &serial.Config{Name: Settings.SerialPortPath, Baud: 57600}
	s, err := serial.OpenPort(c)
	if err != nil {
		panic(err)
	}
	defer s.Close()

	// buffers to use during serial communication
	writeData := make([]byte, 128)
	readData := make([]byte, 1)

	previousSend := time.Now()
	var totalSends int = 0
	var byteData int8 = 0

	// send a -128 to force the arduino to restart and rerequest data
	s.Write([]byte{ResetCommand})

	var pauseAfterWrite = false

	for stepDataOpen := true; stepDataOpen; {
		// wait for next data request
		n, err := s.Read(readData)
		if err != nil {
			panic(err)
		}
		if n != 1 {
			panic(err)
		}

		dataToWrite := int(readData[0])
		for i := 0; i < dataToWrite; i += 2 {

			if pauseAfterWrite {
				// want to fill remainder of buffer with 0s before writing it to serial
				writeData[i] = byte(0)
				writeData[i+1] = byte(0)
			} else {
				// even if stepData is closed and empty, receiving from it will return default value 0 for byteData and false for stepDataOpen
				byteData, stepDataOpen = <-stepData
				writeData[i] = byte(byteData)
				byteData, stepDataOpen = <-stepData
				writeData[i+1] = byte(byteData)
			}
		}

		totalSends++
		if totalSends >= 100 {
			curTime := time.Now()

			fmt.Println("Sent 100 messages after", curTime.Sub(previousSend))
			totalSends = 0

			previousSend = curTime
		}

		s.Write(writeData)

		if pauseAfterWrite {
			pauseAfterWrite = false

			fmt.Println("Press any key to continue...")
			reader := bufio.NewReader(os.Stdin)
			reader.ReadString('\n')
		}
	}
}

// Used to manually adjust length of each step
func InteractiveMoveSpool() {

	for {

		var side string
		var distance float64
		fmt.Print("Input L/R DIST:")
		if _, err := fmt.Scanln(&side, &distance); err != nil {
			return
		}

		leftSpool := strings.ToLower(side) == "l"
		fmt.Println("Moving ", side, distance)

		MoveSpool(leftSpool, distance)
	}
}

// Move a specific spool a given distance
func MoveSpool(leftSpool bool, distance float64) {

	alignStepData := make(chan int8, 1024)
	go WriteStepsToSerial(alignStepData)

	interp := new(TrapezoidInterpolater)
	interp.Setup(Coordinate{}, Coordinate{X: distance, Y: 0}, Coordinate{})
	position := 0.0

	for slice := 1.0; slice <= interp.Slices(); slice++ {

		sliceTarget := interp.Position(slice)

		// calc integer number of steps that will be made this time slice
		sliceSteps := math.Ceil((sliceTarget.X - position) * (StepsFixedPointFactor / Settings.StepSize_MM))
		position = position + sliceSteps*(Settings.StepSize_MM/StepsFixedPointFactor)

		if leftSpool {
			alignStepData <- int8(-sliceSteps)
			alignStepData <- 0
		} else {
			alignStepData <- 0
			alignStepData <- int8(sliceSteps)
		}
	}

	close(alignStepData)
}
