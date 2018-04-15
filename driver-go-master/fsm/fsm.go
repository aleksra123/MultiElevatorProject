package fsm

import (
	"fmt"
	"time"

	"../costfunction"
	"../elevio"
	"../requests"
)

var Elev elevio.Elevator
var Door_timer = time.NewTimer(3 * time.Second)
var Power_timer = time.NewTimer(5* time.Second)
var AckMat = [elevio.NumElevators][elevio.NumFloors][elevio.NumButtons - 1]int{} //acknowledged orders matrise
var BP = [2]int{-10, 0}
var CurrElev = [elevio.NumElevators]elevio.Elevator{} //liste med elevs som main bruker i sentmsg, se main:68

//var Teller int = 0


func RecievedMSG(floor int, button int, pos int, e elevio.Elevator, activeE int, mypos int) {

	var index int
	CurrElev[pos].Position = pos
	if floor != -10   {

					for i := 0; i < activeE; i++ {
						CurrElev[i].AcceptedOrders[floor][button] = e.AcceptedOrders[floor][button]
					}
					 index = costfunction.CostCalc(CurrElev, activeE)
					 CurrElev[index].Requests[floor][button] = true
					 if pos == mypos {

						 Power_timer.Reset(5 * time.Second)
					 }
					 SetHallLights(CurrElev[index])
					 OnRequestButtonPress(floor, elevio.ButtonType(button), index, activeE, mypos)
 }
}

func SetHallLights(es elevio.Elevator) {

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

func SetAllLights(es elevio.Elevator) {

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
func AddCabRequest(pos int, floor int) {
	CurrElev[pos].Requests[floor][elevio.BT_Cab] = true
}

func Init( pos int) {
	elevio.SetMotorDirection(elevio.MD_Down)
	CurrElev[pos].State = elevio.Moving //
	CurrElev[pos].Dir = elevio.MD_Down // failsafes in case of package loss
	CurrElev[pos].FirstTime = true     //
}

func Updatepos(pos int){
	CurrElev[pos].Position = pos
}

func CopyInfo(pos int, lost int){
	fmt.Printf("Dette er lost: %d og dette er pos: %d\n", lost, pos)
	if pos+1 > lost {
		CurrElev[pos-1].AcceptedOrders = CurrElev[pos].AcceptedOrders
		CurrElev[pos-1].Requests = CurrElev[pos].Requests
		CurrElev[pos-1].State = CurrElev[pos].State
		CurrElev[pos-1].Dir = CurrElev[pos].Dir
		CurrElev[pos-1].Position = CurrElev[pos].Position
		CurrElev[pos-1].Floor = CurrElev[pos].Floor
	}
}

func Online(pos int, mypos int){

	CurrElev[pos].Position = pos
	CurrElev[pos].FirstTime = true
}


func OnRequestButtonPress(btn_floor int, btn_type elevio.ButtonType, pos int, activeElevs int, mypos int) {

	switch CurrElev[pos].State {

	case elevio.DoorOpen:
		fmt.Printf("case door open?\n")
		if CurrElev[pos].Floor == btn_floor {
			for i := 0; i < activeElevs; i++ {
			CurrElev[i].AcceptedOrders = requests.ClearAtCurrentFloor(CurrElev[pos], activeElevs).AcceptedOrders
			SetHallLights(CurrElev[i])
			}
			CurrElev[pos] = requests.ClearRequests(CurrElev[pos])

			if pos == mypos {
				SetAllLights(CurrElev[pos])
				Door_timer.Reset(3 * time.Second)
			}
		} else {
			CurrElev[pos].Requests[btn_floor][btn_type] = true

		}

	case elevio.Moving:
		fmt.Printf("case moving ?\n")
		CurrElev[pos].Requests[btn_floor][btn_type] = true

	case elevio.Idle:
		fmt.Printf("case idle\n")

		if CurrElev[pos].Floor == btn_floor {
			if pos == mypos {
				elevio.SetDoorOpenLamp(true)
			}
			CurrElev[pos].State = elevio.DoorOpen
			for i := 0; i < activeElevs; i++ {
			CurrElev[i].AcceptedOrders = requests.ClearAtCurrentFloor(CurrElev[pos], activeElevs).AcceptedOrders
			SetHallLights(CurrElev[i])
			}

			CurrElev[pos] = requests.ClearRequests(CurrElev[pos])
			if pos == mypos {
				SetAllLights(CurrElev[pos])
				Door_timer.Reset(3 * time.Second)
			}

		} else {
			CurrElev[pos].Requests[btn_floor][btn_type] = true
			CurrElev[pos].Dir = requests.ChooseDirection(CurrElev[pos])
			//fmt.Printf("direction for heis 0 : %d\n", CurrElev[pos].Dir)
			if pos == mypos {
				fmt.Printf("check!\n")
				elevio.SetMotorDirection(CurrElev[pos].Dir)
				Power_timer.Reset(5 * time.Second)
			}
			CurrElev[pos].State = elevio.Moving

		}

	}
}

func OnFloorArrival(newFloor int, pos int, activeElevs int, mypos int) {

	CurrElev[pos].Floor = newFloor
	elevio.SetFloorIndicator(CurrElev[mypos].Floor)


	if CurrElev[pos].FirstTime {

		CurrElev[pos].State = elevio.Moving
		CurrElev[pos].Dir = elevio.MD_Down

		if newFloor == 0 {
			fmt.Printf("kommer begge hit?\n")
			if pos == mypos {
				elevio.SetMotorDirection(elevio.MD_Stop)
			}
			CurrElev[pos].State = elevio.Idle
			CurrElev[pos].Dir = elevio.MD_Stop
		}
		//CurrElev[pos].Dir = requests.ChooseDirection(CurrElev[pos])
		// if mypos == pos {
		// 	elevio.SetMotorDirection(CurrElev[mypos].Dir)
		// }
		// if CurrElev[pos].Dir != elevio.MD_Stop{
		// 	CurrElev[pos].State = elevio.Moving
		// }
	}
	if pos == mypos{
		Power_timer.Reset(5* time.Second)
	}

	if requests.ShouldStop(CurrElev[pos]) && !CurrElev[pos].FirstTime{
			if pos == mypos {
				Power_timer.Stop()
			}
			fmt.Printf("skal ikke komme hit\n")
			fmt.Printf("first time til heis 0:, %d\n", CurrElev[0].FirstTime)
			CurrElev[mypos].AcceptedOrders = requests.ClearAtCurrentFloor(CurrElev[pos], activeElevs).AcceptedOrders
			for i := 0; i < activeElevs; i++ {
				CurrElev[i].AcceptedOrders = CurrElev[mypos].AcceptedOrders
			}
			SetHallLights(CurrElev[mypos])

			CurrElev[pos].Requests = requests.ClearRequests(CurrElev[pos]).Requests
		if pos == mypos {

			 SetAllLights(CurrElev[pos])
		}
		if !CurrElev[pos].FirstTime{
			//fmt.Printf("ikke first timer\n")
			CurrElev[pos].Dir = elevio.MD_Stop

			CurrElev[pos].State = elevio.DoorOpen
			if pos == mypos {

				elevio.SetDoorOpenLamp(true)
				Door_timer.Reset(3 * time.Second)
				elevio.SetMotorDirection(elevio.MD_Stop)
			}
		}
	}
	if newFloor == 0 {
		CurrElev[pos].FirstTime = false

	}
}

func Powerout(pos int, activeElevs int){
	CurrElev[pos].State = elevio.Undefined

}

func OnDoorTimeout(pos int, mypos int) {
	// CurrElev[pos].State = state
	switch CurrElev[pos].State {
	case elevio.DoorOpen:
		// fmt.Printf("kommer begge her ? OnDoorTimeout\n")
		// fmt.Printf("Requests: %+v\n", CurrElev[pos].Requests)
		CurrElev[pos].Dir = requests.ChooseDirection(CurrElev[pos])
		if pos == mypos{
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(CurrElev[pos].Dir)
		}
		if CurrElev[pos].Dir == elevio.MD_Stop {

			CurrElev[pos].State = elevio.Idle
		} else {

			CurrElev[pos].State = elevio.Moving
			if pos == mypos {
				Power_timer.Reset(8 * time.Second)

			}
		}
	}
}
