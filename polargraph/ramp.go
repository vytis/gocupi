package polargraph

// Manages the trapezoidal interpolation

import (
	"fmt"
)

type TrapezoidRamp struct {
	origin           float64 // positions currently interpolating from
	destination      float64 // position currently interpolating towards
	directionForward bool    // true - moving forward

	entrySpeed  float64 // speed at beginning at origin
	cruiseSpeed float64 // maximum speed reached
	exitSpeed   float64 // target speed when we reach destination

	acceleration float64 // acceleration, only differs from Settings.Acceleration_MM_S2 when decelerating and there is not enough distance to hit exit speed

	distance float64 // total distance travelled
	time     float64 // total time to go from origin to destination

	accelTime  float64 // time accelerating
	accelDist  float64 // distance covered while accelerating
	cruiseTime float64 // time cruising at max speed
	cruiseDist float64 // distance covered while cruising
	decelTime  float64 // time decelerating
	decelDist  float64 // distance covered while decelerating
}

func (data *TrapezoidRamp) WriteData() {
	fmt.Println("Origin:", data.origin, "Dest:", data.destination)
	if data.directionForward {
		fmt.Println("Dir: Forward")
	} else {
		fmt.Println("Dir: Backward")
	}
	fmt.Println()

	fmt.Println("Entry", data.entrySpeed, "Cruise", data.cruiseSpeed, "Exit", data.exitSpeed)

	fmt.Println("Taccel", data.accelTime, "Tcruise", data.cruiseTime, "Tdecel", data.decelTime)
	fmt.Println("Daccel", data.accelDist, "Dcruise", data.cruiseDist, "Ddecel", data.decelDist)

	fmt.Println("Total distance", data.distance)
}

// Next calculates a ramp to move to nextPosition
func (data *TrapezoidRamp) Next(nextPosition float64) (next TrapezoidRamp) {

	var maxSpeed = Settings.MaxSpeed_MM_S
	var maxAcceleration = Settings.Acceleration_MM_S2

	// special case of not going anywhere
	if data.destination == nextPosition {
		next.origin = nextPosition
		next.destination = nextPosition
		next.directionForward = true

		next.entrySpeed = data.entrySpeed
		next.cruiseSpeed = data.entrySpeed
		next.exitSpeed = data.entrySpeed

		next.acceleration = 0

		next.distance = 0
		next.time = 0

		next.accelTime = 0
		next.accelDist = 0
		next.cruiseTime = 0
		next.cruiseDist = 0
		next.decelTime = 0
		next.decelDist = 0

		return
	}

	// entry speed is whatever the previous exit speed was
	next.entrySpeed = data.exitSpeed
	next.exitSpeed = 0 // For now no lookahead

	next.origin = data.destination
	next.destination = nextPosition
	next.distance = next.destination - next.origin
	next.directionForward = next.distance >= 0

	if next.directionForward {
		var accelerationTime = (maxSpeed - next.entrySpeed) / maxAcceleration
		var accelerationDistance = accelerationTime * (next.entrySpeed + (maxAcceleration*accelerationTime)/2.0)

		var decelerationTime = (maxSpeed - next.exitSpeed) / maxAcceleration
		var decelerationDistance = decelerationTime * ((maxAcceleration*decelerationTime)/2.0 - next.exitSpeed)

		var cruisingDistance = next.distance - accelerationDistance - decelerationDistance
		var cruisingTime = cruisingDistance / maxSpeed

		if cruisingDistance >= 0 {
			next.cruiseSpeed = maxSpeed
			next.acceleration = maxAcceleration

			next.accelDist = accelerationDistance
			next.accelTime = accelerationTime

			next.cruiseDist = cruisingDistance
			next.cruiseTime = cruisingTime

			next.decelDist = decelerationDistance
			next.decelTime = decelerationTime

			next.time = accelerationTime + cruisingTime + decelerationTime
		} else {

		}
	}
	return
	// data.direction = data.destination.Minus(origin)
	// data.distance = data.direction.Len()
	// data.direction = data.direction.Normalized()

	// nextDirection := nextDest.Minus(dest)
	// if nextDirection.Len() == 0 ||
	// 	origin.PenUp == false && dest.PenUp == true {
	// 	// if there is no next direction or we have to stop for pen movement, make the exit speed 0 by pretending the next move will be backwards from current direction
	// 	nextDirection = Coordinate{X: -data.direction.X, Y: -data.direction.Y}
	// } else {
	// 	nextDirection = nextDirection.Normalized()
	// }
	// cosAngle := data.direction.DotProduct(nextDirection)
	// cosAngle = math.Pow(cosAngle, 3) // use cube in order to make it smaller for non straight lines
	// data.exitSpeed = Settings.MaxSpeed_MM_S * math.Max(cosAngle, 0.0)

	// data.cruiseSpeed = Settings.MaxSpeed_MM_S

	// data.accelTime = (data.cruiseSpeed - data.entrySpeed) / Settings.Acceleration_MM_S2
	// data.accelDist = 0.5*Settings.Acceleration_MM_S2*data.accelTime*data.accelTime + data.entrySpeed*data.accelTime

	// data.decelTime = (data.cruiseSpeed - data.exitSpeed) / Settings.Acceleration_MM_S2
	// data.decelDist = 0.5*-Settings.Acceleration_MM_S2*data.decelTime*data.decelTime + data.cruiseSpeed*data.decelTime

	// data.cruiseDist = data.distance - (data.accelDist + data.decelDist)
	// data.cruiseTime = data.cruiseDist / data.cruiseSpeed

	// data.acceleration = Settings.Acceleration_MM_S2

	// // we dont have enough room to reach max velocity, have to calculate what max speed we can reach
	// if data.distance < data.accelDist+data.decelDist {

	// 	// equation derived by http://www.numberempire.com/equationsolver.php from equations:
	// 	// distanceAccel = 0.5 * accel * timeAccel^2 + entrySpeed * timeAccel
	// 	// distanceDecel = 0.5 * -accel * timeDecel^2 + maxSpeed * timeDecel
	// 	// totalDistance = distanceAccel + distanceDecel
	// 	// maxSpeed = entrySpeed + accel * timeAccel
	// 	// maxSpeed = exitSpeed + accel * timeDecel
	// 	data.decelTime = (math.Sqrt2*math.Sqrt(data.exitSpeed*data.exitSpeed+data.entrySpeed*data.entrySpeed+2*Settings.Acceleration_MM_S2*data.distance) - 2*data.exitSpeed) / (2 * Settings.Acceleration_MM_S2)
	// 	data.cruiseTime = 0
	// 	data.cruiseSpeed = data.exitSpeed + Settings.Acceleration_MM_S2*data.decelTime
	// 	data.accelTime = (data.cruiseSpeed - data.entrySpeed) / Settings.Acceleration_MM_S2

	// 	// don't have enough room to accelerate to exitSpeed over the given distance, have to change exit speed
	// 	if data.decelTime < 0 || data.accelTime < 0 {

	// 		if data.exitSpeed > data.entrySpeed { // need to accelerate to max exit speed possible

	// 			data.decelDist = 0
	// 			data.decelTime = 0
	// 			data.cruiseDist = 0
	// 			data.cruiseTime = 0

	// 			// determine time it will take to travel distance at the given acceleration
	// 			data.accelTime = (math.Sqrt(data.entrySpeed*data.entrySpeed+2*Settings.Acceleration_MM_S2*data.distance) - data.entrySpeed) / Settings.Acceleration_MM_S2
	// 			data.exitSpeed = data.entrySpeed + Settings.Acceleration_MM_S2*data.accelTime
	// 			data.cruiseSpeed = data.exitSpeed
	// 			data.accelDist = data.distance
	// 		} else { // need to decelerate to exit speed, by changing acceleration

	// 			//fmt.Println("Warning, unable to decelerate to target exit speed using acceleration, try adding -slowfactor=2")

	// 			data.accelDist = 0
	// 			data.accelTime = 0
	// 			data.cruiseDist = 0
	// 			data.cruiseTime = 0

	// 			// determine time it will take to reach exit speed over the given distance
	// 			data.decelTime = 2.0 * data.distance / (data.exitSpeed + data.entrySpeed)
	// 			data.acceleration = (data.entrySpeed - data.exitSpeed) / data.decelTime
	// 			data.cruiseSpeed = data.entrySpeed
	// 			data.decelDist = data.distance
	// 		}
	// 	} else {
	// 		data.accelDist = 0.5*Settings.Acceleration_MM_S2*data.accelTime*data.accelTime + data.entrySpeed*data.accelTime
	// 		data.cruiseDist = 0
	// 		data.decelDist = 0.5*-Settings.Acceleration_MM_S2*data.decelTime*data.decelTime + data.cruiseSpeed*data.decelTime
	// 	}
	// }

	// data.time = data.accelTime + data.cruiseTime + data.decelTime
	// data.slices = data.time / (TimeSlice_US / 1000000)
}
