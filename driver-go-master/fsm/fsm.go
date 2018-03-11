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

var AckMat = [elevio.NumElevators][elevio.NumFloors][elevio.NumButtons - 1]int{}
var Elevlist = [elevio.NumElevators]elevio.Elevator{}
var BP = [2]int{-10, 0}
var CurrElev elevio.Elevator

func RecievedMSG(floor int, button int, e elevio.Elevator, position int, activeE int) {
	if floor != -10 {
		if AckMat[position][floor][button] != 2 {
			AckMat[position][floor][button] = 1
			//fmt.Printf("Received: %#v\n", AckMat[ID-1])
			for i := 0; i < activeE; i++ {
				if AckMat[i][floor][button] == AckMat[position][floor][button]-1 {
					AckMat[i][floor][button]++
					//fmt.Printf("we incremented! \n")
				}
			}
			var counter int
			for i := 0; i < activeE; i++ {
				if AckMat[i][floor][button] == 1 {
					counter++
				}
			}
			if counter == activeE {
				for i := 0; i < activeE; i++ {
					AckMat[i][floor][button] = 2
				}
				//var index int = costfunction.CostCalc(Elevlist, floor, button, activeE)
				Elev.AcceptedOrders[floor][button] = 1
				Elev.Requests[floor][button] = true // må bruke elevlist[] ved flere heiser ??'
				for i := 0; i < activeE; i++ {
					AckMat[i][floor][button] = 0
				}
				SetAllLights(Elev)
			}
		}

		fmt.Printf("Received: %#v\n", AckMat[position])

		for i := 0; i < elevio.NumFloors; i++ {
			for j := 0; j < elevio.NumButtons; j++ {
				if Elev.Requests[i][j] { // må bruke elevlist[] ved flere heiser ??
					OnRequestButtonPress(i, elevio.ButtonType(j))
				}
			}
		}
	}
}

func SetAllLights(es elevio.Elevator) {
	var btn elevio.ButtonType = elevio.BT_Cab
	for floor := 0; floor < NumFloors; floor++ {
		if Elev.Requests[floor][btn] == true {
			fmt.Printf("cab\n")
			elevio.SetButtonLamp(btn, floor, true)
		} else {
			elevio.SetButtonLamp(btn, floor, false)
		}
	}
	for floor := 0; floor < NumFloors; floor++ {
		for bttn := 0; bttn < NumButtons-1; bttn++ {
			if Elev.AcceptedOrders[floor][bttn] == 1 {
				elevio.SetButtonLamp(elevio.ButtonType(bttn), floor, true)
			} else {
				elevio.SetButtonLamp(elevio.ButtonType(bttn), floor, false)
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
	//setAllLights(Elev) //
	//fmt.Println("\nNew state:\n")
	//Elev_print(Elev)
}

func OnFloorArrival(newFloor int) {
	//fmt.Println(newFloor) //Er noe rart her også
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
			SetAllLights(Elev)
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
