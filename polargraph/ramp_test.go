package polargraph

import (
	"math"
	"testing"
)

func init() {
	Settings.MaxSpeed_MM_S = 10
	Settings.Acceleration_MM_S2 = 5

}

func TestNoMovement(t *testing.T) {
	initial := TrapezoidRamp{}

	next := initial.Next(0)

	if !next.directionForward {
		t.Error("Must be forwards")
	}
	if next.time == 0 {
		t.Error("wrong time: ", next.time)
	}

}

func TestSimpleMovement(t *testing.T) {
	initial := TrapezoidRamp{}

	next := initial.Next(100)

	if !next.directionForward {
		t.Error("Must be forwards")
	}

	if math.Abs(next.time-12) > 0.001 {
		t.Error("wrong time: ", next.time)
	}

	if math.Abs(next.cruiseDist-80) > 0.001 {
		t.Error("wrong distance: ", next.cruiseDist)
	}

}

func TestTouchingMaxSpeed(t *testing.T) {
	initial := TrapezoidRamp{}

	next := initial.Next(20)

	if !next.directionForward {
		t.Error("Must be forwards")
	}

	if math.Abs(next.time-4) > 0.001 {
		t.Error("wrong time: ", next.time)
	}

	if math.Abs(next.cruiseDist) > 0.001 {
		t.Error("wrong distance: ", next.cruiseDist)
	}

}

func TestNotReachingMaxSpeed(t *testing.T) {
	initial := TrapezoidRamp{}

	next := initial.Next(5)

	if !next.directionForward {
		t.Error("Must be forwards")
	}

	if math.Abs(next.time-2) > 0.001 {
		t.Error("wrong time: ", next.time)
	}

	if math.Abs(next.cruiseDist) > 0.001 {
		t.Error("wrong cruiseDist: ", next.cruiseDist)
	}

	if math.Abs(next.acceleration) > 0.001 {
		t.Error("wrong acceleration: ", next.acceleration)
	}

	if math.Abs(next.accelDist-2.5) > 0.001 {
		t.Error("wrong accelDist: ", next.accelDist)
	}

	if math.Abs(next.accelTime-1) > 0.001 {
		t.Error("wrong accelTime: ", next.accelTime)
	}

}

// func TestShortMovement(t *testing.T) {
// 	initial := TrapezoidRamp{}

// 	first := initial.Next(100)
// 	next := first.Next(120)

// 	if !next.directionForward {
// 		t.Error("Must be forwards")
// 	}

// 	if next.time == 12 {
// 		t.Error("wrong time: ", next.time)
// 	}

// 	if next.cruiseDist == 80 {
// 		t.Error("wrong time: ", next.time)
// 	}

// }
