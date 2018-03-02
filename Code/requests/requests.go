package requests

import (
	. "../elevio" //Explicit ?
)

func Check_above(e Elevator) bool {
	for floor := e.Floor + 1; floor < NumFloors; floor++ {
		for button := 0; button < NumButtons; button++ {
			if e.Requests[floor][button] { // ==true --> order
				return true
			}
		}
	}
	return false
}

func Check_below(e Elevator) bool {
	for floor := 0; floor < e.Floor; floor++ {
		for button := 0; button < NumButtons; button++ {
			if e.Requests[floor][button] {
				return true
			}
		}
	}
	return false
}

func ChooseDirection(e Elevator) MotorDirection {
	switch e.Dir {
	case MD_Up:
		if Check_above(e) {
			return MD_Up
		} else if Check_below(e) {
			return MD_Down
		} else {
			return MD_Stop
		}
	case MD_Down: //Compared to C-code. Is this redundant?
		if Check_below(e) {
			return MD_Down
		} else if Check_above(e) {
			return MD_Up
		} else {
			return MD_Stop
		}
	case MD_Stop: //Only one request. Arbitrary if we check up or down first
		if Check_below(e) {
			return MD_Down
		} else if Check_above(e) {
			return MD_Up
		} else {
			return MD_Stop
		}
	default:
		return MD_Stop
	}
}

func ShouldStop(e Elevator) bool {
	switch e.Dir {
	case MD_Down:
		return e.Requests[e.Floor][BT_HallDown] || e.Requests[e.Floor][BT_Cab] || !Check_below(e)
	case MD_Up:
		return e.Requests[e.Floor][BT_HallUp] || e.Requests[e.Floor][BT_Cab] || !Check_above(e)
	case MD_Stop:
	default:
		return true
	}
	return true
}

func ClearAtCurrentFloor(e Elevator) Elevator {
	for button := 0; button < NumButtons; button++ {
		e.Requests[e.Floor][button] = false
	} //Clears both directions
	return e //Why does it return an elevator-type?
}
