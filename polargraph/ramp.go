package polargraph

import (
	"math"
)

// Manages the trapezoidal interpolation
type AxisRamp struct {
	origin           float64 // positions currently interpolating from
	destination      float64 // position currently interpolating towards
	directionForward bool    // true - moving forward

	entrySpeed  float64 // speed at beginning at origin
	cruiseSpeed float64 // maximum speed reached
	exitSpeed   float64 // target speed when we reach destination

	accelDist  float64 // distance covered while accelerating
	cruiseDist float64 // distance covered while cruising
	decelDist  float64 // distance covered while decelerating

	accelTime  float64 // time accelerating
	cruiseTime float64 // time cruising at max speed
	decelTime  float64 // time decelerating

	acceleration float64 // acceleration, only differs from Settings.Acceleration_MM_S2 when decelerating and there is not enough distance to hit exit speed
	deceleration float64

	distance float64 // total distance travelled
	time     float64 // total time to go from origin to destination

}

type TrapezoidRamp struct {
	left  AxisRamp
	right AxisRamp
	time  float64 // total time to go from origin to destination

	accelTime  float64 // time accelerating
	cruiseTime float64 // time cruising at max speed
	decelTime  float64 // time decelerating
}

func (data *TrapezoidRamp) WriteData() {
	// fmt.Println("[L] Origin:", data.left.origin, "Dest:", data.left.destination)
	// if data.left.directionForward {
	// 	fmt.Println("[L] Dir: Forward")
	// } else {
	// 	fmt.Println("[L] Dir: Backward")
	// }
	// fmt.Println("[L] Entry", data.left.entrySpeed, "Cruise", data.left.cruiseSpeed, "Exit", data.left.exitSpeed)
	// fmt.Println("[L] Distance", data.left.distance)

	// fmt.Println("[R] Origin:", data.right.origin, "Dest:", data.right.destination)
	// if data.right.directionForward {
	// 	fmt.Println("[R] Dir: Forward")
	// } else {
	// 	fmt.Println("[R] Dir: Backward")
	// }
	// fmt.Println("[R] Entry", data.right.entrySpeed, "Cruise", data.right.cruiseSpeed, "Exit", data.right.exitSpeed)
	// fmt.Println("[R] Distance", data.right.distance)

	// fmt.Println()

	// fmt.Println("Taccel", data.accelTime, "Tcruise", data.cruiseTime, "Tdecel", data.decelTime)
	// fmt.Println("Daccel", data.accelDist, "Dcruise", data.cruiseDist, "Ddecel", data.decelDist)
}

func (data *LinearInterpolater) Setup(origin, dest, nextDest Coordinate) {

	data.origin = origin
	data.destination = dest
	data.movement = data.destination.Minus(data.origin)
	data.distance = data.movement.Len()

	data.time = data.distance / Settings.MaxSpeed_MM_S
	data.slices = math.Ceil(data.time / (TimeSlice_US / 1000000))
}

func (data *TrapezoidRamp) Next(origin, dest, nextDest Coordinate) (next TrapezoidRamp) {

	left := data.left.Next(origin.X, dest.X, nextDest.X, dest.PenUp)
	right := data.right.Next(origin.Y, dest.Y, nextDest.Y, dest.PenUp)

	if right.time == 0 {
		right.accelTime = left.accelTime

		right.cruiseTime = left.cruiseTime
		right.decelTime = left.decelTime

	} else {
		scale = left.time / right.time
		left = left.Scale(scale)
		right = right.Scale(scale)

	}

	next.accelTime = left.accelTime
	next.cruiseTime = left.accelTime
	return
}

func (data *AxisRamp) Scale(scale float64) (next AxisRamp) {
	next = *data
	next.entrySpeed = next.entrySpeed * scale
	next.accelTime = next.accelTime / scale
	next.cruiseSpeed = next.cruiseSpeed * scale
	next.cruiseTime = next.cruiseTime / scale
	next.exitSpeed = next.exitSpeed * scale
	next.decelTime = next.decelTime / scale

	return
}

func (data *AxisRamp) Next(origin, dest, nextDest float64, penUp bool) (next AxisRamp) {

	// entry speed is whatever the previous exit speed was
	next.entrySpeed = data.exitSpeed

	// special case of not going anywhere
	if origin == dest {
		next.origin = origin
		next.destination = data.destination
		next.distance = 0
		next.exitSpeed = data.entrySpeed
		next.cruiseSpeed = data.entrySpeed
		next.accelDist = 0
		next.accelTime = 0
		next.cruiseDist = 0
		next.cruiseTime = 0
		next.decelDist = 0
		next.decelTime = 0
		next.acceleration = Settings.Acceleration_MM_S2
		return
	}

	next.origin = origin
	next.destination = dest

	next.directionForward = (next.destination - next.origin) > 0
	next.distance = math.Abs(next.destination - next.origin)

	next.exitSpeed = 0

	next.cruiseSpeed = Settings.MaxSpeed_MM_S

	next.accelTime = (next.cruiseSpeed - next.entrySpeed) / Settings.Acceleration_MM_S2
	next.accelDist = 0.5*Settings.Acceleration_MM_S2*next.accelTime*next.accelTime + next.entrySpeed*next.accelTime

	next.decelTime = (next.cruiseSpeed - next.exitSpeed) / Settings.Acceleration_MM_S2
	next.decelDist = 0.5*-Settings.Acceleration_MM_S2*next.decelTime*next.decelTime + next.cruiseSpeed*next.decelTime

	next.cruiseDist = next.distance - (next.accelDist + next.decelDist)
	next.cruiseTime = next.cruiseDist / next.cruiseSpeed

	data.acceleration = Settings.Acceleration_MM_S2

	// we dont have enough room to reach max velocity, have to calculate what max speed we can reach
	if next.distance < next.accelDist+next.decelDist {

		// equation derived by http://www.numberempire.com/equationsolver.php from equations:
		// distanceAccel = 0.5 * accel * timeAccel^2 + entrySpeed * timeAccel
		// distanceDecel = 0.5 * -accel * timeDecel^2 + maxSpeed * timeDecel
		// totalDistance = distanceAccel + distanceDecel
		// maxSpeed = entrySpeed + accel * timeAccel
		// maxSpeed = exitSpeed + accel * timeDecel
		next.decelTime = (math.Sqrt2*math.Sqrt(next.exitSpeed*next.exitSpeed+next.entrySpeed*next.entrySpeed+2*Settings.Acceleration_MM_S2*next.distance) - 2*data.exitSpeed) / (2 * Settings.Acceleration_MM_S2)
		next.cruiseTime = 0
		next.cruiseSpeed = next.exitSpeed + Settings.Acceleration_MM_S2*next.decelTime
		next.accelTime = (next.cruiseSpeed - next.entrySpeed) / Settings.Acceleration_MM_S2

		// don't have enough room to accelerate to exitSpeed over the given distance, have to change exit speed
		if next.decelTime < 0 || next.accelTime < 0 {

			if next.exitSpeed > next.entrySpeed { // need to accelerate to max exit speed possible

				next.decelDist = 0
				next.decelTime = 0
				next.cruiseDist = 0
				next.cruiseTime = 0

				// determine time it will take to travel distance at the given acceleration
				next.accelTime = (math.Sqrt(next.entrySpeed*next.entrySpeed+2*Settings.Acceleration_MM_S2*next.distance) - next.entrySpeed) / Settings.Acceleration_MM_S2
				next.exitSpeed = next.entrySpeed + Settings.Acceleration_MM_S2*next.accelTime
				next.cruiseSpeed = next.exitSpeed
				next.accelDist = next.distance
			} else { // need to decelerate to exit speed, by changing acceleration

				//fmt.Println("Warning, unable to decelerate to target exit speed using acceleration, try adding -slowfactor=2")

				next.accelDist = 0
				next.accelTime = 0
				next.cruiseDist = 0
				next.cruiseTime = 0

				// determine time it will take to reach exit speed over the given distance
				next.decelTime = 2.0 * next.distance / (next.exitSpeed + next.entrySpeed)
				next.acceleration = (next.entrySpeed - next.exitSpeed) / next.decelTime
				next.cruiseSpeed = next.entrySpeed
				next.decelDist = next.distance
			}
		} else {
			next.accelDist = 0.5*Settings.Acceleration_MM_S2*next.accelTime*next.accelTime + next.entrySpeed*next.accelTime
			next.cruiseDist = 0
			next.decelDist = 0.5*-Settings.Acceleration_MM_S2*next.decelTime*next.decelTime + next.cruiseSpeed*next.decelTime
		}
	}

	next.time = next.accelTime + next.cruiseTime + next.decelTime
	return
}

// func (data *AxisRamp) Next(nextPosition float64) (next AxisRamp) {

// 	next.origin = data.destination
// 	next.destination = nextPosition
// 	next.directionForward = (next.origin-next.destination > 0)
// 	next.entrySpeed = data.exitSpeed
// 	next.distance = math.Abs(next.origin - next.destination)

// 	var speedup = Settings.MaxSpeed_MM_S - next.entrySpeed
// 	var maxAccelDistance = 0.5 * math.Pow(speedup, 2) / Settings.Acceleration_MM_S2

// 	if next.distance/2.0 < maxAccelDistance {
// 		next.accelTime = math.Sqrt(next.distance / Settings.Acceleration_MM_S2)
// 	} else {
// 		next.accelTime = speedup / Settings.Acceleration_MM_S2
// 	}

// 	var slowdown = Settings.MaxSpeed_MM_S - next.exitSpeed
// 	var maxDecelDistance = 0.5 * math.Pow(slowdown, 2) / Settings.Acceleration_MM_S2

// 	if next.distance/2.0 < maxDecelDistance {
// 		next.accelTime = math.Sqrt(next.distance / Settings.Acceleration_MM_S2)
// 	} else {
// 		next.accelTime = slowdown / Settings.Acceleration_MM_S2
// 	}

// 	next.accelDist = 0.5 * Settings.Acceleration_MM_S2 * math.Pow(next.accelTime, 2)

// 	next.decelDist = 0.5 * Settings.Acceleration_MM_S2 * math.Pow(next.decelTime, 2)

// 	// next.cruiseSpeed = Settings.Acceleration_MM_S2 * next.accelTime + next.entrySpeed
// 	// next.cruiseTime =
// 	return
// }

// Next calculates a ramp to move to nextPosition
// func (data *TrapezoidRamp) Next(left float64, right float64) (next TrapezoidRamp) {

// 	var maxSpeed = Settings.MaxSpeed_MM_S
// 	var maxAcceleration = Settings.Acceleration_MM_S2

// 	// special case of not going anywhere
// 	if data.destination == nextPosition {
// 		next.origin = nextPosition
// 		next.destination = nextPosition
// 		next.directionForward = true

// 		next.entrySpeed = data.entrySpeed
// 		next.cruiseSpeed = data.entrySpeed
// 		next.exitSpeed = data.entrySpeed

// 		next.acceleration = 0

// 		next.distance = 0
// 		next.time = 0

// 		next.accelTime = 0
// 		next.accelDist = 0
// 		next.cruiseTime = 0
// 		next.cruiseDist = 0
// 		next.decelTime = 0
// 		next.decelDist = 0

// 		return
// 	}

// 	// entry speed is whatever the previous exit speed was
// 	next.entrySpeed = data.exitSpeed
// 	next.exitSpeed = 0 // For now no lookahead

// 	next.origin = data.destination
// 	next.destination = nextPosition
// 	next.distance = next.destination - next.origin
// 	next.directionForward = next.distance >= 0

// 	if next.directionForward {
// 		var accelerationTime = (maxSpeed - next.entrySpeed) / maxAcceleration
// 		var accelerationDistance = accelerationTime * (next.entrySpeed + (maxAcceleration*accelerationTime)/2.0)

// 		var decelerationTime = (maxSpeed - next.exitSpeed) / maxAcceleration
// 		var decelerationDistance = decelerationTime * ((maxAcceleration*decelerationTime)/2.0 - next.exitSpeed)

// 		var cruisingDistance = next.distance - accelerationDistance - decelerationDistance
// 		var cruisingTime = cruisingDistance / maxSpeed

// 		if cruisingDistance >= 0 {
// 			next.cruiseSpeed = maxSpeed
// 			next.acceleration = maxAcceleration

// 			next.accelDist = accelerationDistance
// 			next.accelTime = accelerationTime

// 			next.cruiseDist = cruisingDistance
// 			next.cruiseTime = cruisingTime

// 			next.decelDist = decelerationDistance
// 			next.decelTime = decelerationTime

// 			next.time = accelerationTime + cruisingTime + decelerationTime
// 		} else {

// 		}
// 	}
// 	return
// 	// data.direction = data.destination.Minus(origin)
// 	// data.distance = data.direction.Len()
// 	// data.direction = data.direction.Normalized()

// 	// nextDirection := nextDest.Minus(dest)
// 	// if nextDirection.Len() == 0 ||
// 	// 	origin.PenUp == false && dest.PenUp == true {
// 	// 	// if there is no next direction or we have to stop for pen movement, make the exit speed 0 by pretending the next move will be backwards from current direction
// 	// 	nextDirection = Coordinate{X: -data.direction.X, Y: -data.direction.Y}
// 	// } else {
// 	// 	nextDirection = nextDirection.Normalized()
// 	// }
// 	// cosAngle := data.direction.DotProduct(nextDirection)
// 	// cosAngle = math.Pow(cosAngle, 3) // use cube in order to make it smaller for non straight lines
// 	// data.exitSpeed = Settings.MaxSpeed_MM_S * math.Max(cosAngle, 0.0)

// 	// data.cruiseSpeed = Settings.MaxSpeed_MM_S

// 	// data.accelTime = (data.cruiseSpeed - data.entrySpeed) / Settings.Acceleration_MM_S2
// 	// data.accelDist = 0.5*Settings.Acceleration_MM_S2*data.accelTime*data.accelTime + data.entrySpeed*data.accelTime

// 	// data.decelTime = (data.cruiseSpeed - data.exitSpeed) / Settings.Acceleration_MM_S2
// 	// data.decelDist = 0.5*-Settings.Acceleration_MM_S2*data.decelTime*data.decelTime + data.cruiseSpeed*data.decelTime

// 	// data.cruiseDist = data.distance - (data.accelDist + data.decelDist)
// 	// data.cruiseTime = data.cruiseDist / data.cruiseSpeed

// 	// data.acceleration = Settings.Acceleration_MM_S2

// 	// // we dont have enough room to reach max velocity, have to calculate what max speed we can reach
// 	// if data.distance < data.accelDist+data.decelDist {

// 	// 	// equation derived by http://www.numberempire.com/equationsolver.php from equations:
// 	// 	// distanceAccel = 0.5 * accel * timeAccel^2 + entrySpeed * timeAccel
// 	// 	// distanceDecel = 0.5 * -accel * timeDecel^2 + maxSpeed * timeDecel
// 	// 	// totalDistance = distanceAccel + distanceDecel
// 	// 	// maxSpeed = entrySpeed + accel * timeAccel
// 	// 	// maxSpeed = exitSpeed + accel * timeDecel
// 	// 	data.decelTime = (math.Sqrt2*math.Sqrt(data.exitSpeed*data.exitSpeed+data.entrySpeed*data.entrySpeed+2*Settings.Acceleration_MM_S2*data.distance) - 2*data.exitSpeed) / (2 * Settings.Acceleration_MM_S2)
// 	// 	data.cruiseTime = 0
// 	// 	data.cruiseSpeed = data.exitSpeed + Settings.Acceleration_MM_S2*data.decelTime
// 	// 	data.accelTime = (data.cruiseSpeed - data.entrySpeed) / Settings.Acceleration_MM_S2

// 	// 	// don't have enough room to accelerate to exitSpeed over the given distance, have to change exit speed
// 	// 	if data.decelTime < 0 || data.accelTime < 0 {

// 	// 		if data.exitSpeed > data.entrySpeed { // need to accelerate to max exit speed possible

// 	// 			data.decelDist = 0
// 	// 			data.decelTime = 0
// 	// 			data.cruiseDist = 0
// 	// 			data.cruiseTime = 0

// 	// 			// determine time it will take to travel distance at the given acceleration
// 	// 			data.accelTime = (math.Sqrt(data.entrySpeed*data.entrySpeed+2*Settings.Acceleration_MM_S2*data.distance) - data.entrySpeed) / Settings.Acceleration_MM_S2
// 	// 			data.exitSpeed = data.entrySpeed + Settings.Acceleration_MM_S2*data.accelTime
// 	// 			data.cruiseSpeed = data.exitSpeed
// 	// 			data.accelDist = data.distance
// 	// 		} else { // need to decelerate to exit speed, by changing acceleration

// 	// 			//fmt.Println("Warning, unable to decelerate to target exit speed using acceleration, try adding -slowfactor=2")

// 	// 			data.accelDist = 0
// 	// 			data.accelTime = 0
// 	// 			data.cruiseDist = 0
// 	// 			data.cruiseTime = 0

// 	// 			// determine time it will take to reach exit speed over the given distance
// 	// 			data.decelTime = 2.0 * data.distance / (data.exitSpeed + data.entrySpeed)
// 	// 			data.acceleration = (data.entrySpeed - data.exitSpeed) / data.decelTime
// 	// 			data.cruiseSpeed = data.entrySpeed
// 	// 			data.decelDist = data.distance
// 	// 		}
// 	// 	} else {
// 	// 		data.accelDist = 0.5*Settings.Acceleration_MM_S2*data.accelTime*data.accelTime + data.entrySpeed*data.accelTime
// 	// 		data.cruiseDist = 0
// 	// 		data.decelDist = 0.5*-Settings.Acceleration_MM_S2*data.decelTime*data.decelTime + data.cruiseSpeed*data.decelTime
// 	// 	}
// 	// }

// 	// data.time = data.accelTime + data.cruiseTime + data.decelTime
// 	// data.slices = data.time / (TimeSlice_US / 1000000)
// }
