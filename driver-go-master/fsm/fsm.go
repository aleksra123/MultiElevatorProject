package fsm

import (
	"fmt"
	"time"

	//"../costfunction"
	"../elevio"
	"../requests"
)

var Elev elevio.Elevator

var Door_timer = time.NewTimer(3 * time.Second)

var AckMat = [elevio.NumElevators][elevio.NumFloors][elevio.NumButtons - 1]int{} //acknowledged orders matrise
var BP = [2]int{-10, 0}
var CurrElev = [elevio.NumElevators]elevio.Elevator{} //liste med elevs som main bruker i sentmsg, se main:68

var firstTime bool = false //trengte dette + litt i OnFloorArrival for å intialisere når vi har flere heiser



func RecievedMSG(floor int, button int, position int, e elevio.Elevator, activeE int) {
	var index int

	//CurrElev[position].Floor = e.Floor
	//CurrElev[position].Position = position
	//CurrElev[e.Position].Position = e.Position

	 // fmt.Printf("pos til heis 0: %d\n", CurrElev[0].Position)
	 // fmt.Printf("pos til heis 1: %d\n", CurrElev[1].Position)
	if floor != -10   {


					for i := 0; i < activeE; i++ {
						CurrElev[i].AcceptedOrders[floor][button] = e.AcceptedOrders[floor][button]
					}

					fmt.Printf("AO: %+v\n", e.AcceptedOrders)
					// if CurrElev[position].AcceptedOrders[floor][button] == 1 { //floor != e.Floor && () !(floor == e.Floor && e.State == elevio.Idle) &&
					// 	index = costfunction.CostCalc(CurrElev, activeE)
					// 	fmt.Printf("index: %d\n", index)
					// 	CurrElev[index].Requests[floor][button] = true
					// 	// fmt.Printf("Request: %+v\n", CurrElev[index].Requests)
					// 	SetAllLights(CurrElev[index])
					//  }
					 index = 0
					 CurrElev[index].Requests[floor][button] = true
					 SetAllLights(CurrElev[index])
					 //else if e.State == elevio.Idle{
					// 	elevio.SetDoorOpenLamp(true)
					// 	CurrElev[position].State = elevio.DoorOpen
					// 	Door_timer.Reset(3 * time.Second)
					// }
		for i := 0; i < elevio.NumFloors; i++ {
			for j := 0; j < elevio.NumButtons; j++ {
				if CurrElev[index].Requests[i][j] {
					OnRequestButtonPress(i, elevio.ButtonType(j), index, activeE)
			}
		}
	}
 }
 // prev.Floor = floor
 // prev.Button = elevio.ButtonType(button)
}

func NewFloor(e elevio.Elevator, pos int){

	CurrElev[pos].Floor = e.Floor

	// fmt.Printf("floor til heis 0: %d\n", CurrElev[0].Floor)
	// fmt.Printf("floor til heis 1: %d\n", CurrElev[1].Floor)
	//fmt.Printf("possss, %d\n", pos)
}

func ArrivedAtOrderedFloor(e elevio.Elevator, pos int, activeElevs int){

	CurrElev[pos].Floor = e.Floor
	for i := 0; i < activeElevs; i++ {
		CurrElev[i].AcceptedOrders[e.Floor][0] = 0
		CurrElev[i].AcceptedOrders[e.Floor][1] = 0
		SetAllLights(CurrElev[i])
	}
	fmt.Printf("floor til heis 0: %d\n", CurrElev[0].Floor)
	fmt.Printf("floor til heis 1: %d\n", CurrElev[1].Floor)
	fmt.Printf("AccOrders: %+v\n", CurrElev[0].AcceptedOrders)
	fmt.Printf("AccOrders: %+v\n", CurrElev[1].AcceptedOrders)
}
//
// func PosUpdate( pos int) {
//
// 		CurrElev[pos].Position = pos
// }

func SetAllLights(es elevio.Elevator) {
	//stress å cleare en etasjes lys hvis en aen heis ar cab order der, fikke det ikke helt til
	var btn elevio.ButtonType = elevio.BT_Cab
	for floor := 0; floor < elevio.NumFloors; floor++ {
		if es.Requests[floor][btn] == true {

			elevio.SetButtonLamp(btn, floor, true)
		} else {
			elevio.SetButtonLamp(btn, floor, false)
		}
	}
	for floor := 0; floor < elevio.NumFloors; floor++ {
		for bttn := 0; bttn < elevio.NumButtons-1; bttn++ {
			if es.AcceptedOrders[floor][bttn] == 1 {
				//fmt.Printf("SAL, if : %+v\n", es.AcceptedOrders)
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


func OnRequestButtonPress(btn_floor int, btn_type elevio.ButtonType, pos int, activeE int) {


	 // fmt.Printf("Floor til heis 0: %d\n", CurrElev[0].Floor)
	 // fmt.Printf("AccOrders: %+v\n", CurrElev[0].AcceptedOrders)
	 // fmt.Printf("AccOrders: %+v\n", CurrElev[1].AcceptedOrders)



	switch CurrElev[pos].State {

	case elevio.DoorOpen:

		if CurrElev[pos].Floor == btn_floor {
			for i := 0; i < activeE; i++ {
			CurrElev[i].AcceptedOrders = requests.ClearAtCurrentFloor(CurrElev[pos]).AcceptedOrders
			SetAllLights(CurrElev[i])
			}
			//CurrElev[pos] = requests.ClearRequests(CurrElev[pos])
			CurrElev[pos] = requests.ClearRequests(CurrElev[pos])
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
			for i := 0; i < activeE; i++ {
			CurrElev[i].AcceptedOrders = requests.ClearAtCurrentFloor(CurrElev[pos]).AcceptedOrders
			SetAllLights(CurrElev[i])
			}
			//CurrElev[pos] = requests.ClearRequests(CurrElev[pos])
			CurrElev[pos] = requests.ClearRequests(CurrElev[pos])
			SetAllLights(CurrElev[pos])
			Door_timer.Reset(3 * time.Second)

		} else {

			CurrElev[pos].Requests[btn_floor][btn_type] = true
			CurrElev[pos].Dir = requests.ChooseDirection(CurrElev[pos])

			//ettersom en av heisene nesten alltid kjører opp
			elevio.SetMotorDirection(CurrElev[pos].Dir)
			CurrElev[pos].State = elevio.Moving
			//fmt.Printf("pos i ORBP: %d\n", pos)
		}

	}
}

func OnFloorArrival(newFloor int, pos int, activeE int) {
	//fmt.Println(newFloor)

	if firstTime {

		Elev.State = elevio.Idle
		elevio.SetMotorDirection(elevio.MD_Stop)
		firstTime = false

	}

	CurrElev[pos].Floor = newFloor

	elevio.SetFloorIndicator(CurrElev[pos].Floor)


	if  newFloor == 3 {

		elevio.SetMotorDirection(elevio.MD_Stop)
	} else if  newFloor == 0 {
		elevio.SetMotorDirection(elevio.MD_Stop)
	}

	switch CurrElev[pos].State {
	case elevio.Moving:
		//fmt.Printf("OFA, case moving\n")
		if requests.ShouldStop(CurrElev[pos]) {

			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)

			for i := 0; i < activeE; i++ {

			CurrElev[i].AcceptedOrders = requests.ClearAtCurrentFloor(CurrElev[pos]).AcceptedOrders
			SetAllLights(CurrElev[i])
			// CurrElev[i].AcceptedOrders[newFloor][0] = 0
			// CurrElev[i].AcceptedOrders[newFloor][1] = 0

			}

			CurrElev[pos] = requests.ClearRequests(CurrElev[pos])
			SetAllLights(CurrElev[pos])
			Door_timer.Reset(3 * time.Second)

			CurrElev[pos].State = elevio.DoorOpen
		}
	}
	 fmt.Printf("AccOrders i OnFloorArrival: %+v\n", CurrElev[0].AcceptedOrders)
	 fmt.Printf("AccOrders i OnFloorArrival: %+v\n", CurrElev[1].AcceptedOrders)
	 // fmt.Printf("Requests: %+v\n", CurrElev[0].Requests)
	 // fmt.Printf("Requests: %+v\n", CurrElev[1].Requests)
	 // fmt.Printf("pos til heis 0: %d\n", CurrElev[0].Position)
	 // fmt.Printf("pos til heis 1: %d\n", CurrElev[1].Position)
	 // fmt.Printf("Floor til heis 0: %d\n", CurrElev[0].Floor)
 	 // fmt.Printf("Floor til heis 1: %d\n", CurrElev[1].Floor)

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
