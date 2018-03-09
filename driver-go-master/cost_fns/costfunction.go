package cost_fns

import (
	"../elevio"
	"../requests"
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
