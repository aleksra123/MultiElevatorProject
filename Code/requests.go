//package something

import(
  "elevator_io.go"
)

func requests_above(e Elevator) int {
  for floor := e.Floor + 1; floor < _numFloors; floor++ {
    for button := 0; button < _numButtons; button++ {
      if(e.Queue[floor][button]){ // ==true --> order
        return 1
      }
    }
  }
  return 0
}

func requests_below(e Elevator) int {
  for floor := 0; floor < e.Floor; floor++{
    for button := 0; button < _numButtons; button++ {
      if(e.Queue[floor][button]){
        return 1
      }
    }
  }
  return 0;
}

func requests_chooseDirection(e Elevator) MotorDirection {
  switch e.Dir{
  case MD_Up:
    if requests_above(e) {
      return MD_Up
    } else if requests_below(e) {
      return MD_Down
    } else {
      return MD_Stop
    }
  case MD_Down: //Compared to C-code. Is this redundant?
    if requests_below(e) {
      return MD_Down
    } else if requests_above(e) {
      return MD_Up
    } else {
      return MD_Stop
    }
  case MD_Stop: //Only one request. Arbitrary if we check up or down first
    if requests_below(e) {
      return MD_Down
    } else if requests_above(e) {
      return MD_Up
    } else {
      return MD_Stop
    }
  default:
    return MD_Stop
  }
}

func requests_shouldStop(e Elevator) int{
  switch e.Dir {
  case MD_Down:
    return e.Queue[e.Floor][BT_HallDown] || e.Queue[e.Floor][BT_Cab] || !requests_below(e)
  case MD_Up:
    return e.Queue[e.Floor][BT_HallUp] || e.Queue[e.Floor][BT_Cab] || !requests_above(e)
  case MD_Stop:
  default:
    return 1
  }
}

func requests_clearAtCurrentFloor(e Elevator) Elevator{
  for button := 0; button < _numButtons; button++ {
    e.Queue[e.Floor][button]  = 0
  }
  return e //Why does it return an elevator-type?
}
