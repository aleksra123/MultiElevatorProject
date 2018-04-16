package costfunction

import (
	"../elevio"
	"../requests"
	"math"
)

var TRAVEL_TIME float64 = 2.5
var DOOR_OPEN_TIME float64 = 3

func FakeClearAtCurrentFloor(e_old elevio.Elevator) elevio.Elevator {
	e := e_old
	for btn := 0; btn < elevio.NumButtons; btn++ {
		if e.Requests[e.Floor][btn] {
			e.Requests[e.Floor][btn] = false
		}
	}
	return e
}

func timeToIdle(e elevio.Elevator) float64 {
	var duration float64 = 0

	switch e.State {
	case elevio.Idle:
		e.Dir = requests.ChooseDirection(e)
		for floor := 0; floor < elevio.NumFloors; floor++ {
			for button := 0; button < elevio.NumButtons-1; button++ {
				if e.AcceptedOrders[floor][button] == 1 {
					duration = duration + math.Abs(float64(e.Floor - floor))
				}
			}
		}

		if e.Dir == elevio.MD_Stop {
			return duration
		}

	case elevio.Moving:
		duration += TRAVEL_TIME / 2
		e.Floor += int(e.Dir)

	case elevio.DoorOpen:
		duration -= DOOR_OPEN_TIME / 2
	}

	for {
		if requests.ShouldStop(e) {
			e = FakeClearAtCurrentFloor(e)
			duration += DOOR_OPEN_TIME
			e.Dir = requests.ChooseDirection(e)
			if e.Dir == elevio.MD_Stop {
				return duration
			}
		}
		e.Floor += int(e.Dir)
		duration += TRAVEL_TIME
	}
}

func CostCalc(elevlist [elevio.NumElevators]elevio.Elevator, activeElevators int , lost int) int {
	var minCost float64 = 500
	var index int
	var time float64
	for i := 0; i < activeElevators; i++ {
		elevlist[i].State = 0
		if i != lost {
			time = timeToIdle(elevlist[i])
			if time < minCost {
				minCost = time
				index = i
			}
		}
	}
	return index
}
