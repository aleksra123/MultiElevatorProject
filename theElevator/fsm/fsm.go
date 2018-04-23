package fsm

import (
	"time"


	"../backup"
	"../costfunction"
	"../elevio"
	"../requests"
)

var Elev elevio.Elevator
var Door_timer = time.NewTimer(3 * time.Second)
var Power_timer = time.NewTimer(5 * time.Second)
var AckMat = [elevio.NumElevators][elevio.NumFloors][elevio.NumButtons - 1]int{}
var BP [2]int
var CurrElev = [elevio.NumElevators]elevio.Elevator{}

func RecievedMSG(floor int, button int, pos int, e elevio.Elevator, activeE int, mypos int) {

	var index int
	CurrElev[pos].Position = pos
	if floor != -10 {

		for i := 0; i < activeE; i++ {
			CurrElev[i].AcceptedOrders[floor][button] = e.AcceptedOrders[floor][button]
		}
		index = costfunction.CostCalc(CurrElev, activeE, -1, floor)
		CurrElev[index].Requests[floor][button] = true
		if index == mypos {

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
				elevio.SetButtonLamp(elevio.ButtonType(bttn), floor, true)
			} else {
				elevio.SetButtonLamp(elevio.ButtonType(bttn), floor, false)
			}
		}
	}

}
func AddCabRequest(pos int, floor int, mypos int) {
	CurrElev[pos].Requests[floor][elevio.BT_Cab] = true

}

func Init(pos int, activeElevs int) {
	elevio.SetDoorOpenLamp(false)

	elevio.SetMotorDirection(elevio.MD_Down)
	CurrElev[pos].Requests = backup.ReadBackup(CurrElev[pos]).Requests
	var empty elevio.Elevator
	SetHallLights(empty)
	backup.UpdateBackup(empty)
	CurrElev[pos].State = elevio.Moving //
	CurrElev[pos].Dir = elevio.MD_Down  // failsafes in case of package loss
	CurrElev[pos].FirstTime = true      //

}

func Updatepos(pos int) {
	CurrElev[pos].Position = pos
}

func CopyInfo_Lost(lost int, activeElevs int) {

	for i := 1; i < activeElevs+1; i++ {
		if i == lost {
			CurrElev[lost-1].AcceptedOrders = CurrElev[lost].AcceptedOrders
			CurrElev[lost-1].Requests = CurrElev[lost].Requests
			CurrElev[lost-1].State = CurrElev[lost].State
			CurrElev[lost-1].Dir = CurrElev[lost].Dir
			CurrElev[lost-1].Position = CurrElev[lost].Position
			CurrElev[lost-1].Floor = CurrElev[lost].Floor
		} else if i > lost {
			CurrElev[lost].AcceptedOrders = CurrElev[lost+1].AcceptedOrders
			CurrElev[lost].Requests = CurrElev[lost+1].Requests
			CurrElev[lost].State = CurrElev[lost+1].State
			CurrElev[lost].Dir = CurrElev[lost+1].Dir
			CurrElev[lost].Position = CurrElev[lost+1].Position
			CurrElev[lost].Floor = CurrElev[lost+1].Floor
		}
	}
}

func CopyInfo_New(new int, activeElevs int) {

	for i := 1; i < activeElevs+1; i++ {
		if i == new {
			for j := activeElevs-1; j > new-1; j--{
				CurrElev[j] = CurrElev[j-1]
			}
		}
		// 	CurrElev[new].AcceptedOrders = CurrElev[new-1].AcceptedOrders
		// 	CurrElev[new].Requests = CurrElev[new-1].Requests
		// 	CurrElev[new].State = CurrElev[new-1].State
		// 	CurrElev[new].Dir = CurrElev[new-1].Dir
		// 	CurrElev[new].Position = CurrElev[new-1].Position
		// 	CurrElev[new].Floor = CurrElev[new-1].Floor
		// } else if i > new {
		// 	CurrElev[new+1].AcceptedOrders = CurrElev[new].AcceptedOrders
		// 	CurrElev[new+1].Requests = CurrElev[new].Requests
		// 	CurrElev[new+1].State = CurrElev[new].State
		// 	CurrElev[new+1].Dir = CurrElev[new].Dir
		// 	CurrElev[new+1].Position = CurrElev[new].Position
		// 	CurrElev[new+1].Floor = CurrElev[new].Floor
		// }
	}
}

func TransferRequests(lost int, activeElevs int, pos int) {
	index := 0
	for floor := 0; floor < elevio.NumFloors; floor++ {
		for button := 0; button < elevio.NumButtons-1; button++ {
			if CurrElev[lost-1].Requests[floor][button] {
				index = costfunction.CostCalc(CurrElev, activeElevs+1, lost-1, floor)

				OnRequestButtonPress(floor, elevio.ButtonType(button), index, activeElevs, pos)

			}
		}
	}
}

func Online(pos int, mypos int) {

	CurrElev[pos].Position = pos

}

func OnRequestButtonPress(btn_floor int, btn_type elevio.ButtonType, pos int, activeElevs int, mypos int) {

	switch CurrElev[pos].State {

	case elevio.DoorOpen:
		if CurrElev[pos].Floor == btn_floor {
			for i := 0; i < activeElevs; i++ {
				CurrElev[i].AcceptedOrders = requests.ClearAtCurrentFloor(CurrElev[pos], activeElevs).AcceptedOrders
				SetHallLights(CurrElev[i])
			}
			CurrElev[pos] = requests.ClearRequests(CurrElev[pos])

			if pos == mypos {
				SetAllLights(CurrElev[pos])
				Door_timer.Reset(3 * time.Second)
				Power_timer.Stop()
			}
		} else {
			CurrElev[pos].Requests[btn_floor][btn_type] = true

		}

	case elevio.Moving:

		CurrElev[pos].Requests[btn_floor][btn_type] = true

	case elevio.Idle:
		if CurrElev[pos].Floor == btn_floor {
			if pos == mypos {
				elevio.SetDoorOpenLamp(true)
				Power_timer.Stop()
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
			if pos == mypos {
				elevio.SetMotorDirection(CurrElev[pos].Dir)
				Power_timer.Reset(8 * time.Second)
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
			CurrElev[pos].State = elevio.Idle
			CurrElev[pos].Dir = elevio.MD_Stop
			if pos == mypos {
				elevio.SetMotorDirection(elevio.MD_Stop)
				for floor := 0; floor < elevio.NumFloors; floor++ {
					if CurrElev[pos].Requests[floor][elevio.BT_Cab] {
						OnRequestButtonPress(floor, elevio.BT_Cab, pos, activeElevs, pos)

					}
				}
			}
		}
	}
	if pos == mypos && !CurrElev[pos].FirstTime {
		Power_timer.Reset(5 * time.Second) //bruk 8 sek i stedet
	}

	if requests.ShouldStop(CurrElev[pos]) && !CurrElev[pos].FirstTime {
		if pos == mypos {
			Power_timer.Stop()
		}
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

func Powerout(pos int, activeElevs int) {
	CurrElev[pos].State = elevio.Undefined

}

func OnDoorTimeout(pos int, mypos int) {
	switch CurrElev[pos].State {
	case elevio.DoorOpen:
		CurrElev[pos].Dir = requests.ChooseDirection(CurrElev[pos])
		if pos == mypos {
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
