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

//Might have to change package location later.

var Elev elevio.Elevator

//var outputDevice ElevOutputDevice ??

var Door_timer = time.NewTimer(3 * time.Second)

func setAllLights(es elevio.Elevator) {
	for floor := 0; floor < NumFloors; floor++ {
		for btn := 0; btn < NumButtons; btn++ {
			if Elev.Requests[floor][btn] == true {
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
	Elev.State = elevio.Moving
}

func OnRequestButtonPress(btn_floor int, btn_type elevio.ButtonType) {
	//fmt.Println(btn_floor, elevio_button_toString(btn_type)) //Mangler to første argumenter
	//Elev_print(Elev)
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
	setAllLights(Elev) //
	//fmt.Println("\nNew state:\n")
	//Elev_print(Elev)
}

func OnFloorArrival(newFloor int) {
	fmt.Println(newFloor) //Er noe rart her også
	//Elev_print(Elev)
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
	}
	//fmt.Println("\nNew state:\n")
	//Elev_print(Elev)
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
