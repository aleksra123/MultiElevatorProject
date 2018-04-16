package requests

import (
	."../elevio"
	
)

func Check_above(e Elevator) bool {
	for floor := e.Floor + 1; floor < NumFloors; floor++ {
		for button := 0; button < NumButtons; button++ {
			if e.Requests[floor][button] {
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
	case MD_Down:
		if Check_below(e) {
			return MD_Down
		} else if Check_above(e) {
			return MD_Up
		} else {
			return MD_Stop
		}
	case MD_Stop:
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

func ClearAtCurrentFloor(e Elevator, activeElevs int) Elevator {

	switch e.Dir {
	case MD_Down:

		if activeElevs == 1{
			if !Check_below(e) {
				e.AcceptedOrders[e.Floor][BT_HallUp] = 0
			}
			e.AcceptedOrders[e.Floor][BT_HallDown] = 0
			return e
		}else {
			if !Check_below(e) && !Check_atFloor(e) {
				e.AcceptedOrders[e.Floor][BT_HallUp] = 0
			}
			e.AcceptedOrders[e.Floor][BT_HallDown] = 0
			return e
		}

	case MD_Up:

		if activeElevs == 1 {
			if !Check_above(e)  {
			 e.AcceptedOrders[e.Floor][BT_HallDown] = 0
		 }
		 e.AcceptedOrders[e.Floor][BT_HallUp] = 0
		 return e
		} else {
		 if !Check_above(e) && !Check_atFloor(e) {
			e.AcceptedOrders[e.Floor][BT_HallDown] = 0
		}
		e.AcceptedOrders[e.Floor][BT_HallUp] = 0
		return e
	}

	case MD_Stop:
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
		if !Check_below(e) && !Check_atFloor(e){
			e.Requests[e.Floor][BT_HallUp] = false
		}
		return e
	case MD_Up:
		e.Requests[e.Floor][BT_HallUp] = false
		if !Check_above(e) && !Check_atFloor(e){
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
