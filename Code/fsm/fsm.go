package fsm

import (
	"fmt"
	"time"

	"../elevio"
	"../requests"
)

const (
	NumFloors    = 4
	NumButtons   = 3
	NumElevators = 3
)

var Elev elevio.Elevator

var Door_timer = time.NewTimer(3 * time.Second)

func setAllLights(e elevio.Elevator) {
	for floor := 0; floor < NumFloors; floor++ {
		for btn := 0; btn < NumButtons; btn++ {
			if e.Requests[floor][btn] == true {
				elevio.SetButtonLamp(elevio.ButtonType(btn), floor, true)
			} else {
				elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
			}
		}
	}
}

func OnInitBetweenFloors() {
	elevio.SetMotorDirection(elevio.MD_Down)
	Elev.Dir = elevio.MD_Down
	Elev.State = elevio.Undefined
}

func OnRequestButtonPress(btn_floor int, btn_type elevio.ButtonType) {
	switch Elev.State {
	case elevio.DoorOpen:
		if Elev.Floor == btn_floor {
			Door_timer.Reset(3 * time.Second)
		} else {
			Elev.Requests[btn_floor][btn_type] = true

		}

	case elevio.Moving:
		Elev.Requests[btn_floor][btn_type] = true

	case elevio.Idle:
		if Elev.Floor == btn_floor {
			elevio.SetDoorOpenLamp(true)
			Door_timer.Reset(3 * time.Second)
			Elev.State = elevio.DoorOpen
		} else {
			Elev.Requests[btn_floor][btn_type] = true
			Elev.Dir = requests.ChooseDirection(Elev)
			elevio.SetMotorDirection(Elev.Dir)
			Elev.State = elevio.Moving
		}

	}
	setAllLights(Elev)
}

func OnFloorArrival(newFloor int) {
	fmt.Println("Floor:",newFloor)
	Elev.Floor = newFloor
	elevio.SetFloorIndicator(Elev.Floor)
	switch Elev.State {
	case elevio.Moving:
		if requests.ShouldStop(Elev) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			Elev = requests.ClearAtCurrentFloor(Elev)
			Door_timer.Reset(3 * time.Second)
			setAllLights(Elev)
			Elev.State = elevio.DoorOpen
		}
	case elevio.Undefined:
		 elevio.SetMotorDirection(elevio.MD_Stop)
		 Elev.State = elevio.Idle
	}
}

func OnDoorTimeout() {
	switch Elev.State {
	case elevio.DoorOpen:
		Elev.Dir = requests.ChooseDirection(Elev)
		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(Elev.Dir)
		if Elev.Dir == elevio.MD_Stop {
			Elev.State = elevio.Idle
		} else {
			Elev.State = elevio.Moving
		}
	}
}

func StateToString(elev elevio.Elevator) string{
	switch elev.State {
	case elevio.Idle:
		return "Idle"
	case elevio.Moving:
		return "Moving"
	case elevio.DoorOpen:
		return "Door open"
	default:
		return "Invalid state"
	}
}
