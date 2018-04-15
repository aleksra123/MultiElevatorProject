package requests

import (
	. "../elevio" //Explicit ?
	"fmt"
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

func Check_atFloor(e Elevator) bool {
	if e.AcceptedOrders[e.Floor][BT_HallUp] == 1 && e.AcceptedOrders[e.Floor][BT_HallDown] == 1 {
		return true
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
		// fmt.Printf("SS dooown\n")
		return e.Requests[e.Floor][BT_HallDown] || e.Requests[e.Floor][BT_Cab] || !Check_below(e)
	case MD_Up:
		// fmt.Printf("SS case UP i Should stop\n")
		return e.Requests[e.Floor][BT_HallUp] || e.Requests[e.Floor][BT_Cab] || !Check_above(e)
	case MD_Stop:
		// fmt.Printf("SS stop\n")
	default:
		// fmt.Printf("test\n")
		return true
	}
	return true
}

func ClearAtCurrentFloor(e Elevator) Elevator { // kanskje ta inn active elevs for a cleare samme floor ved en elev

	switch e.Dir {
	case MD_Down:
		// fmt.Printf("CACF case down\n")
		if !Check_below(e) && !Check_atFloor(e){
			e.AcceptedOrders[e.Floor][BT_HallUp] = 0
		}
		e.AcceptedOrders[e.Floor][BT_HallDown] = 0
		return e

	case MD_Up:
		// fmt.Printf("CACF case up\n")
		if !Check_above(e) && !Check_atFloor(e){
			fmt.Printf("no pls\n")
			e.AcceptedOrders[e.Floor][BT_HallDown] = 0
		}
		e.AcceptedOrders[e.Floor][BT_HallUp] = 0
		return e

	case MD_Stop:
		// fmt.Printf("CACF case stop\n")
		e.AcceptedOrders[e.Floor][BT_HallUp] = 0
		e.AcceptedOrders[e.Floor][BT_HallDown] = 0
		return e
	default:
		return e
	}


}

func ClearRequests(e Elevator) Elevator {
	e.Requests[e.Floor][BT_Cab] = false
	switch e.Dir {
	case MD_Down:
		e.Requests[e.Floor][BT_HallDown] = false
		if !Check_below(e) {
			e.Requests[e.Floor][BT_HallUp] = false
		}
		return e
	case MD_Up:
		e.Requests[e.Floor][BT_HallUp] = false
		if !Check_above(e) {
			e.Requests[e.Floor][BT_HallDown] = false
		}
		return e
	case MD_Stop:
		e.Requests[e.Floor][BT_HallUp] = false
		e.Requests[e.Floor][BT_HallDown] = false
		return e
	default:
		return e
	}
}
