package fsm

import (
	"fmt"
	"time"

	"../costfunction"
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

var firstTime bool = false

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
				var index int = costfunction.CostCalc(Elevlist, floor, button, activeE)
				fmt.Printf("index: %d\n", index)
				Elevlist[index].AcceptedOrders[floor][button] = 1
				Elevlist[index].Requests[floor][button] = true // må bruke elevlist[] ved flere heiser ??'
				for i := 0; i < activeE; i++ {
					AckMat[i][floor][button] = 0
				}
				SetAllLights(Elevlist[index])
			}
		}

		//fmt.Printf("Received: %#v\n", AckMat[position])

		for i := 0; i < elevio.NumFloors; i++ {
			for j := 0; j < elevio.NumButtons; j++ {
				if Elevlist[position].Requests[i][j] { // må bruke elevlist[] ved flere heiser ??

					OnRequestButtonPress(i, elevio.ButtonType(j), position)
				}
			}
		}
	}
}

func SetAllLights(es elevio.Elevator) {
	var btn elevio.ButtonType = elevio.BT_Cab
	for floor := 0; floor < NumFloors; floor++ {
		if es.Requests[floor][btn] == true {
			fmt.Printf("cab\n")
			elevio.SetButtonLamp(btn, floor, true)
		} else {
			elevio.SetButtonLamp(btn, floor, false)
		}
	}
	for floor := 0; floor < NumFloors; floor++ {
		for bttn := 0; bttn < NumButtons-1; bttn++ {
			if es.AcceptedOrders[floor][bttn] == 1 {
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
	firstTime = true
}

func OnRequestButtonPress(btn_floor int, btn_type elevio.ButtonType, pos int) {
	//fmt.Println(btn_floor, elevio_button_toString(btn_type)) //Mangler to første argumenter
	//Elev_print(Elev)

	switch Elevlist[pos].State {

	case elevio.DoorOpen:

		if Elevlist[pos].Floor == btn_floor {
			Door_timer.Reset(3 * time.Second)
		} else {
			Elevlist[pos].Requests[btn_floor][btn_type] = true

		}

	case elevio.Moving:
		Elevlist[pos].Requests[btn_floor][btn_type] = true

	case elevio.Idle:
		if Elevlist[pos].Floor == btn_floor {

			elevio.SetDoorOpenLamp(true)
			Door_timer.Reset(3 * time.Second)
			Elevlist[pos].State = elevio.DoorOpen
		} else {

			Elevlist[pos].Requests[btn_floor][btn_type] = true
			Elevlist[pos].Dir = requests.ChooseDirection(Elevlist[pos])
			elevio.SetMotorDirection(Elevlist[pos].Dir)
			Elevlist[pos].State = elevio.Moving
		}

	}
	//setAllLights(Elev) //
	//fmt.Println("\nNew state:\n")
	//Elev_print(Elev)
}

func OnFloorArrival(newFloor int, pos int) {
	//fmt.Println(newFloor) //Er noe rart her også
	//Elev_print(Elev)
	if firstTime {
		Elev.State = elevio.Idle
		elevio.SetMotorDirection(elevio.MD_Stop)
		firstTime = false
	}
	fmt.Printf("\nOFA\n")
	Elevlist[pos].Floor = newFloor
	elevio.SetFloorIndicator(Elevlist[pos].Floor)
	fmt.Printf("state:   %d\n", Elevlist[pos].State)
	switch Elevlist[pos].State {
	case elevio.Moving:
		fmt.Printf("OFA, case moving\n")
		if requests.ShouldStop(Elevlist[pos]) {
			fmt.Printf("OFA, case moving, if should stop\n")
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			Elevlist[pos] = requests.ClearAtCurrentFloor(Elevlist[pos])
			fmt.Printf("Position: %d\n", pos)

			Door_timer.Reset(3 * time.Second)
			SetAllLights(Elevlist[pos])
			Elevlist[pos].State = elevio.DoorOpen
		}
	}
	//fmt.Println("\nNew state:\n")
	//Elev_print(Elev)
}

func OnDoorTimeout(pos int) {
	switch Elevlist[pos].State {
	case elevio.DoorOpen:
		Elevlist[pos].Dir = requests.ChooseDirection(Elevlist[pos])
		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(Elevlist[pos].Dir)
		if Elevlist[pos].Dir == elevio.MD_Stop {
			Elevlist[pos].State = elevio.Idle
		} else {
			Elevlist[pos].State = elevio.Moving
		}
	}
}
