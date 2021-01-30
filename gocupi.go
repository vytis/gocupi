package main

import (
	"errors"
	"flag"
	"fmt"
	"sort"
	"strconv"
	"strings"

	p "github.com/vytis/gocupi/polargraph"
)

// set flag usage variable so that entire help will be output
func init() {
	flag.Usage = PrintGenericHelp
}

// main
func main() {
	p.Settings.Read()
	p.Config.Read()

	toImageFlag := flag.Bool("toimage", false, "Output result to an image file instead of to the stepper")
	toChartFlag := flag.Bool("tochart", false, "Output a chart of the movement and velocity")
	countFlag := flag.Bool("count", false, "Outputs the time it would take to draw")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		PrintGenericHelp()
		return
	}

	plotCoords := make(chan p.Coordinate, 1024)
	var err error
	var params []float64

	switch args[0] {

	case "help":
		if len(args) != 2 {
			PrintGenericHelp()
		} else {
			PrintCommandHelp(args[1])
		}
		return

	case "spool":
		if len(args) == 3 {

			leftSpool := strings.ToLower(args[1]) == "l"
			if params, err = GetArgsAsFloats(args[2:], 1, true); err != nil {
				fmt.Println("ERROR: ", err)
				fmt.Println()
				PrintCommandHelp("spool")
				return
			}

			p.MoveSpool(leftSpool, params[0])
		} else {
			p.InteractiveMoveSpool()
		}
		return

	case "svg":
		if len(args) < 2 {
			fmt.Println("ERROR: ", fmt.Sprint("Expected at least 2 parameters and saw ", len(args)-1))
			fmt.Println()
			PrintCommandHelp("svg")
			return
		}

		optimize := false
		if len(args) > 2 {
			lastFlag := args[2]
			if lastFlag == "optimize" {
				optimize = true
			}
		}

		fmt.Println("Generating svg path")
		input, width, height := p.ParseSvgFile(args[1])
		var data []p.Coordinate
		if optimize {
			data = p.OptimizeTravel(input)
		} else {
			data = input
		}

		go p.GenerateSvgPath(data, plotCoords)

		if *toImageFlag {
			svgFileName := strings.Replace(args[1], ".svg", ".png", -1)
			fmt.Println("Outputting to image ", svgFileName)
			p.DrawToImageExact(svgFileName, width, height, plotCoords)
			return
		}

	default:
		PrintGenericHelp()
		return
	}

	// output the max speed and acceleration
	fmt.Println()
	fmt.Printf("MaxSpeed: %.3f mm/s Accel: %.3f mm/s^2", p.Settings.MaxSpeed_MM_S, p.Settings.Acceleration_MM_S2)
	fmt.Println()

	stepData := make(chan int8, 1024)
	go p.GenerateSteps(plotCoords, stepData)
	switch {
	case *countFlag:
		p.CountSteps(stepData)
	case *toChartFlag:
		p.WriteStepsToChart(stepData)
	default:
		p.WriteStepsToSerial(stepData)
	}
}

// Parse a series of numbers as floats
func GetArgsAsFloats(args []string, expectedCount int, preventZero bool) ([]float64, error) {

	if len(args) < expectedCount {
		return nil, errors.New(fmt.Sprint("Expected at least ", expectedCount, " numeric parameters and only saw ", len(args)))
	}

	numbers := make([]float64, expectedCount)

	var err error
	for argIndex := 0; argIndex < expectedCount; argIndex++ {
		if numbers[argIndex], err = strconv.ParseFloat(args[argIndex], 64); err != nil {
			return nil, errors.New(fmt.Sprint("Unable to parse ", args[argIndex], " as a float: ", err))
		}

		if preventZero && numbers[argIndex] == 0 {
			return nil, errors.New(fmt.Sprint("0 is not a valid value for parameter ", argIndex))
		}
	}

	return numbers, nil
}

// output the help for a specific command
func PrintCommandHelp(command string) {

	helpText, ok := CommandHelp[command]
	if !ok {
		fmt.Println("Unrecognized command: " + command)
		PrintGenericHelp()
	}
	fmt.Println(helpText)
	fmt.Println()
}

// output help summary
func PrintGenericHelp() {
	fmt.Println(`
General Usage: (flags) COMMAND PARAMETERS...

All distance numbers are in millimeters
All angles are in radians

Flags:
-toimage, outputs data to an image of what the render should look like
-tochart, outputs a graph of velocity and position
-count, outputs number of steps and render time

Commands:`)

	// output list of possible commands
	var keys []string
	for k := range CommandHelp {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	first := true
	for _, k := range keys {
		if !first {
			fmt.Print(", ")
		} else {
			first = false
		}
		fmt.Print(k)
	}
	fmt.Println()
	fmt.Println("help COMMAND to view help for a specific command")
	fmt.Println()
}

var CommandHelp = map[string]string{
	`spool`: `Directly control spool movement, useful for initial setup. If you ommit the L/R d parameters then you enter an interactive mode where you can repeatedly type the options to enter several spool commands in a row.

spool [L|R] d
	L|R - designing either the left or right spool
	d - distance to extend line, negative numbers retract`,

	`svg`: `Draw an svg file. Must be made up of only straight lines, curves are not currently supported in the svg parser.

svg "path" [optimize]
	path - path to svg file
	optimize - if the flag is passed then pen travel will be optimized to reduce unnecessary movements`,
}
