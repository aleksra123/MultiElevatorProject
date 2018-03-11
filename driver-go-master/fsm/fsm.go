package fsm

import (
	"fmt"
	"time"

	"../costfunction"
	"../elevio"
	"../requests"
)

//Might have to change package location later.

var Elev elevio.Elevator

//var outputDevice ElevOutputDevice ??

var Door_timer = time.NewTimer(3 * time.Second)

var AckMat = [elevio.NumElevators][elevio.NumFloors][elevio.NumButtons - 1]int{}

//var Elevlist = [elevio.NumElevators]elevio.Elevator{}
var BP = [2]int{-10, 0}
var CurrElev = [elevio.NumElevators]elevio.Elevator{}

var firstTime bool = false
var teller int

func RecievedMSG(floor int, button int, e elevio.Elevator, position int, activeE int) {
	var index int
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
					CurrElev[i].AcceptedOrders[floor][button] = 1
					fmt.Printf("Accepted av i : %d\n", CurrElev[i].AcceptedOrders)
				}
				index = costfunction.CostCalc(CurrElev, floor, button, activeE)
				fmt.Printf("index: %d\n", index)

				CurrElev[index].Requests[floor][button] = true
				for i := 0; i < activeE; i++ {
					fmt.Printf("Accepted av i : %+v\n", CurrElev[i].Requests)
					AckMat[i][floor][button] = 0

				}
				//fmt.Printf("Detter er AckMat[1]: %+v \n", AckMat[1])
				SetAllLights(CurrElev[index])

			}
		}

		//fmt.Printf("Received: %#v\n", AckMat[position])

		for i := 0; i < elevio.NumFloors; i++ {
			for j := 0; j < elevio.NumButtons; j++ {
				if CurrElev[index].Requests[i][j] { // må bruke elevlist[] ved flere heiser ??

					OnRequestButtonPress(i, elevio.ButtonType(j), index)
				}
			}
		}
	}
}

func SetAllLights(es elevio.Elevator) {

	var btn elevio.ButtonType = elevio.BT_Cab
	for floor := 0; floor < elevio.NumFloors; floor++ {
		if es.Requests[floor][btn] == true {
			fmt.Printf("cab\n")
			elevio.SetButtonLamp(btn, floor, true)
		} else {
			elevio.SetButtonLamp(btn, floor, false)
		}
	}
	for floor := 0; floor < elevio.NumFloors; floor++ {
		for bttn := 0; bttn < elevio.NumButtons-1; bttn++ {
			if es.AcceptedOrders[floor][bttn] == 1 {
				//fmt.Printf("teller : %d\n", teller)
				//fmt.Printf("SAL, if : %+v\n", es.AcceptedOrders)
				elevio.SetButtonLamp(elevio.ButtonType(bttn), floor, true)
			} else {
				elevio.SetButtonLamp(elevio.ButtonType(bttn), floor, false)
			}
		}
	}
	teller++
}

func OnInitBetweenFloors() {
	elevio.SetMotorDirection(elevio.MD_Down)
	Elev.Dir = elevio.MD_Down
	Elev.State = elevio.Moving
	firstTime = true
}

func OnRequestButtonPress(btn_floor int, btn_type elevio.ButtonType, pos int) {
	//fmt.Println(btn_floor, elevio_button_toString(btn_type)) //Mangler to første argumenter
	//Elev_print(Elev)//fmt.Println(btn_floor,
	fmt.Printf("pos i ORBP: %d\n", pos)
	switch CurrElev[pos].State {

	case elevio.DoorOpen:

		if CurrElev[pos].Floor == btn_floor {
			CurrElev[pos] = requests.ClearAtCurrentFloor(CurrElev[pos])
			SetAllLights(CurrElev[pos])
			Door_timer.Reset(3 * time.Second)
		} else {
			CurrElev[pos].Requests[btn_floor][btn_type] = true

		}

	case elevio.Moving:
		CurrElev[pos].Requests[btn_floor][btn_type] = true

	case elevio.Idle:
		if CurrElev[pos].Floor == btn_floor {

			elevio.SetDoorOpenLamp(true)
			CurrElev[pos].State = elevio.DoorOpen
			CurrElev[pos] = requests.ClearAtCurrentFloor(CurrElev[pos])
			SetAllLights(CurrElev[pos])
			Door_timer.Reset(3 * time.Second)

		} else {

			CurrElev[pos].Requests[btn_floor][btn_type] = true
			CurrElev[pos].Dir = requests.ChooseDirection(CurrElev[pos])
			elevio.SetMotorDirection(CurrElev[pos].Dir)
			CurrElev[pos].State = elevio.Moving
		}

	}
}

func OnFloorArrival(newFloor int, pos int, activeE int) {
	//fmt.Println(newFloor) //Er noe rart her også
	//Elev_print(Elev)
	if firstTime {
		Elev.State = elevio.Idle
		elevio.SetMotorDirection(elevio.MD_Stop)
		firstTime = false
	}
	//fmt.Printf("\nOFA\n")
	CurrElev[pos].Floor = newFloor
	elevio.SetFloorIndicator(CurrElev[pos].Floor)
	//fmt.Printf("state:   %d\n", Elevlist[pos].State)
	switch CurrElev[pos].State {
	case elevio.Moving:
		//fmt.Printf("OFA, case moving\n")
		if requests.ShouldStop(CurrElev[pos]) {
			//fmt.Printf("OFA, case moving, if should stop\n")
			//fmt.Printf("POS: %d\n", pos)
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			// for i := 0; i < activeE; i++ {
			// 	fmt.Printf("Accepted: %+v\n", CurrElev[i].AcceptedOrders)
			// }
			CurrElev[pos] = requests.ClearAtCurrentFloor(CurrElev[pos])
			for i := 0; i < activeE; i++ {

				fmt.Printf("Cleared av i: %+v\n", CurrElev[i].AcceptedOrders)
				SetAllLights(CurrElev[i])
			}
			fmt.Println("slutt\n")
			Door_timer.Reset(3 * time.Second)

			CurrElev[pos].State = elevio.DoorOpen
		}
	}
	//fmt.Println("\nNew state:\n")
	//Elev_print(Elev)
}

func OnDoorTimeout(pos int) {
	switch CurrElev[pos].State {
	case elevio.DoorOpen:
		CurrElev[pos].Dir = requests.ChooseDirection(CurrElev[pos])
		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(CurrElev[pos].Dir)
		if CurrElev[pos].Dir == elevio.MD_Stop {
			CurrElev[pos].State = elevio.Idle
		} else {
			CurrElev[pos].State = elevio.Moving
		}
	}
}
