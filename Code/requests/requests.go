package requests

import(
  . "../elevio" //Explicit ?
)

func check_above(e Elevator) int {
  for floor := e.Floor + 1; floor < _numFloors; floor++ {
    for button := 0; button < _numButtons; button++ {
      if(e.Requests[floor][button]){ // ==true --> order
        return 1
      }
    }
  }
  return 0
}

func check_below(e Elevator) int {
  for floor := 0; floor < e.Floor; floor++{
    for button := 0; button < _numButtons; button++ {
      if(e.Requests[floor][button]){
        return 1
      }
    }
  }
  return 0;
}

func chooseDirection(e Elevator) MotorDirection {
  switch e.Dir{
  case MD_Up:
    if check_above(e) {
      return MD_Up
    } else if check_below(e) {
      return MD_Down
    } else {
      return MD_Stop
    }
  case MD_Down: //Compared to C-code. Is this redundant?
    if check_below(e) {
      return MD_Down
    } else if check_above(e) {
      return MD_Up
    } else {
      return MD_Stop
    }
  case MD_Stop: //Only one request. Arbitrary if we check up or down first
    if check_below(e) {
      return MD_Down
    } else if check_above(e) {
      return MD_Up
    } else {
      return MD_Stop
    }
  default:
    return MD_Stop
  }
}

func shouldStop(e Elevator) int{
  switch e.Dir {
  case MD_Down:
    return e.Requests[e.Floor][BT_HallDown] || e.Requests[e.Floor][BT_Cab] || !check_below(e)
  case MD_Up:
    return e.Requests[e.Floor][BT_HallUp] || e.Requests[e.Floor][BT_Cab] || !check_above(e)
  case MD_Stop:
  default:
    return 1
  }
}

func clearAtCurrentFloor(e Elevator) Elevator{
  for button := 0; button < _numButtons; button++ {
    e.Requests[e.Floor][button]  = 0
  }
  return e //Why does it return an elevator-type?
}
