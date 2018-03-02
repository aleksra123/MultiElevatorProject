package fsm

import(
  "fmt"
  "time"
  "../elevio"
)

//Might have to change package location later.

var elevator elevio.Elevator
//var outputDevice ElevOutputDevice ??

func setAllLights(es elevio.Elevator)  {
  for floor := 0; floor < elevio._numFloors; floor++ {
    for btn := 0; btn < elevio._numButtons; btn++ {
      if elevator.Requests[floor][button] == true {
        elevio.SetButtonLamp(btn, floor, true)
      } else{
        elevio.SetButtonLamp(btn, floor, false)
      }
    }
  }
}

func onInitBetweenFloors(){
  elevio.SetMotorDirection(elevio.MD_Down)
  elevator.Dir = elevio.MD_Down
  elevator.State = elevio.Moving
}

func onRequestButtonPress(btn_floor int, btn_type ButtonType)  {
  //fmt.Println(btn_floor, elevio_button_toString(btn_type)) //Mangler to første argumenter
  //elevator_print(elevator)
  switch elevator.State {
  case elevator.DoorOpen:
    if elevator.Floor == btn_floor {
      timer_start(elevator.config.doorOpenDuration_s) //???
    }
    else{
      elevator.Requests[btn_floor][btn_type] = 1

    }
    break
  case elevator.Moving:
    elevator.Requests[btn_floor][btn_type] = 1
    break
  case elevio.Idle:
    if elevator.Floor == btn_floor {
      elevio.SetDoorOpenLamp(true)
      timer_start(elevator.config.doorOpenDuration_s) //??
      elevator.State = elevio.DoorOpen
    }
    else {
      elevator.Requests[btn_floor][btn_type] = 1
      elevator.Dir = requests.chooseDirection(elevator)
      elevio.SetMotorDirection(elevator.Dir)
      elevator.State = elevio.Moving
    }
    break
  }
  setAllLights(elevator) //
  //fmt.Println("\nNew state:\n")
  //elevator_print(elevator)
}

func onFloorArrival(newFloor int){
  fmt.Println(newFloor) //Er noe rart her også
  //elevator_print(elevator)
  elevator.Floor = newFloor
  elevio.SetFloorIndicator(elevator.Floor)
  switch elevator.State {
  case elevator.Moving:
    if requests.shouldStop(elevator) {
      elevio.SetMotorDirection(elevio.MD_Stop)
      elevio.SetDoorOpenLamp(true)
      elevator = Requests.clearAtCurrentFloor(elevator)
      timer_start(elevator.config.doorOpenDuration_s) //???
      setAllLights(elevator)
      elevator.State = elevio.DoorOpen
    }
    break
  default:
    break
  }
  //fmt.Println("\nNew state:\n")
  //elevator_print(elevator)
}

func onDoorTimeout()  {
  //fmt.Println() //Weird shit igjen
  //elevator_print(elevator)
  switch elevator.State {
  case elevio.DoorOpen:
    elevator.Dir = requests.chooseDirection(elevator)
    elevio.SetDoorOpenLamp(false)
    elevio.SetMotorDirection(elevator.Dir)
    if elevator.Dir == elevio.MD_Stop {
      elevator.state = elevio.Idle
    }
    else {
      elevator.State = elevio.Moving
    }
    break
  default:
    break
  }
  //fmt.Println("\nNew state:\n")
  //elevator_print(elevator)
}
