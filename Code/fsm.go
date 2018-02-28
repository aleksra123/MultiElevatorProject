import(
  "fmt"
  "time"
  //con_load.go
  //elevator.go
  //elevator_io_device.go
  //requests.go
  //timer.go
)

var Elevator elevator
var ElevOutputDevice outputDevice

func setAllLights(es Elevator)  {
  for floor := 0; floor < NumFloors; floor++ {
    for btn := 0; btn < NumButtons; btn++ {
      outputDevice.requestButtonLight(floor, btn, es.request[floor][btn])
    }
  }
}

func fsm_onInitBetweenFloors(){
  outputDevice.motorDirection(D_Down)
  elevator.dirn = D_Down
  elevator.behaviour = EB.Moving
}

func fsm_onRequestButtonPress(btn_floor int, btn_type Button)  {
  fmt.Println(btn_floor, elevio_button_toString(btn_type)) //Mangler to første argumenter
  elevator_print(elevator)
  switch elevator.behaviour {
  case EB_DoorOpen:
    if elevator.floor == btn_floor {
      timer_start(elevator.config.doorOpenDuration_s)
    }
    else{
      elevator.requests[btn_floor][btn_type] = 1
    }
    break
  case EB_Moving:
    elevator.requests[btn_floor][btn_type] = 1
    break
  case EB_Idle:
    if elevator.floor == btn_floor {
      outputDevice.doorLight(1)
      timer_start(elevator.config.doorOpenDuration_s)
      elevator.behaviour = EB_DoorOpen
    }
    else {
      elevator.requests[btn_floor][btn_type] = 1
      elevator.dirn = requests_chooseDirection(elevator)
      outputDevice.motorDirection(elevator.dirn)
      elevator.behaviour = EB_Moving
    }
    break
  }
  setAllLights(elevator)
  fmt.Println("\nNew state:\n")
  elevator_print(elevator)
}

func fsm_onFloorArrival(newFloor int){
  fmt.Println(newFloor) //Er noe rart her også
  elevator_print(elevator)
  elevator.floor = newFloor
  outputDevice.floorIndicator(elevator.floor)
  switch elevator.behaviour {
  case EB_Moving:
    if requests_shouldStop(elevator) {
      outputDevice.motorDirection(D_Stop)
      outputDevice.doorLight(1)
      elevator = requests_clearAtCurrentFloor(elevator)
      timer_start(elevator.config.doorOpenDuration_s)
      setAllLights(elevator)
      elevator.behaviour = EB_DoorOpen
    }
    break
  default:
    break
  }
  fmt.Println("\nNew state:\n")
  elevator_print(elevator)
}

func fsm_onDoorTimeout()  {
  fmt.Println() //Weird shit igjen
  elevator_print(elevator)
  switch elevator.behaviour {
  case EB_DoorOpen:
    elevator.dirn = requests_chooseDirection(elevator)
    outputDevice.doorLight(0)
    outputDevice.motorDirection(elevator.dirn)
    if elevator.dirn == D_Stop {
      elevator.behaviour = EB_Idle
    }
    else {
      elevator.behaviour = EB_Moving
    }
    break
  default:
    break
  }
  fmt.Println("\nNew state:\n")
  elevator_print(elevator)
}
