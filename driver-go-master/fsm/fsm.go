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

var AckMat = [elevio.NumElevators][elevio.NumFloors][elevio.NumButtons - 1]int{} //acknowledged orders matrise
var BP = [2]int{-10, 0}
var CurrElev = [elevio.NumElevators]elevio.Elevator{} //liste med elevs som main bruker i sentmsg, se main:68

var firstTime bool = false //trengte dette + litt i OnFloorArrival for å intialisere når vi har flere heiser

func UpdateAllElevs(floor int, pos int){

	//CurrElev[pos].Floor = floor
	//fmt.Printf("come at me bro pls\n")
}

func RecievedMSG(floor int, button int, position int, activeE int) {
	var index int
	//fmt.Printf("posisjon. %d\n", position) // posisjon i lista, bør være 0 hvis du bare har en heis
	if floor != -10 {                      // se main:118
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
					//fmt.Printf("Received: %#v\n", AckMat[i])
					CurrElev[i].AcceptedOrders[floor][button] = 1
					fmt.Printf("AccOrders: %+v\n", CurrElev[i].AcceptedOrders)
					// Er egentlig veldig skeptisk til denne måten å akseptere
					// ordre på, siden vi bare sender en gang og så går alt på
					// logiske operasjoner.
					//har en mistanke om at hver heis som kjører lager sin egen
					// CurrElev liste hvor elementene ikke nødvendigvis er like
					//de andre heisene sin CurrElev liste...men vet ikke
					//fmt.Printf("Accepted av i : %d\n", CurrElev[i].AcceptedOrders)
				}
				index = costfunction.CostCalc(CurrElev, floor, button, activeE)
				fmt.Printf("index: %d\n", index)
				//kostfunksjonen vår ser ikke ut til å gi rett index hver gang

				CurrElev[index].Requests[floor][button] = true

				//fmt.Printf("Detter er AckMat[1]: %+v \n", AckMat[1])
				SetAllLights(CurrElev[index])
				for i := 0; i < activeE; i++ {
					//CurrElev[i].AcceptedOrders[floor][button] = 0
					AckMat[i][floor][button] = 0
				}

			}
		}
		for i := 0; i < elevio.NumFloors; i++ {
			for j := 0; j < elevio.NumButtons; j++ {
				if CurrElev[index].Requests[i][j] { //sjekker requests for den beste heisen, if true kjør OnRequestButtonPress
					OnRequestButtonPress(i, elevio.ButtonType(j), index, activeE)
					fmt.Printf("index2: %d\n", index)


				}
			}
		}
	}
}

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

	switch CurrElev[pos].State {

	case elevio.DoorOpen:

		if CurrElev[pos].Floor == btn_floor {
			// for i := 0; i < activeE; i++ {
			// CurrElev[i] = requests.ClearAtCurrentFloor(CurrElev[pos])
			// SetAllLights(CurrElev[i])
			// }
			CurrElev[pos] = requests.ClearRequests(CurrElev[pos])
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
			// for i := 0; i < activeE; i++ {
			// CurrElev[i] = requests.ClearAtCurrentFloor(CurrElev[pos])
			// SetAllLights(CurrElev[i])
			// }
			CurrElev[pos] = requests.ClearRequests(CurrElev[pos])
			CurrElev[pos] = requests.ClearAtCurrentFloor(CurrElev[pos])
			SetAllLights(CurrElev[pos])
			Door_timer.Reset(3 * time.Second)

		} else {
			fmt.Printf("idle else\n")
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


	//fmt.Printf("on floor arrival\n")
  //fmt.Printf("pos i OFA %d\n", pos)
	if Elev.Dir == elevio.MD_Up && newFloor == 3 {
		fmt.Printf("stop då bro")
		elevio.SetMotorDirection(elevio.MD_Stop)
	} else if Elev.Dir == elevio.MD_Down && newFloor == 0 {
		elevio.SetMotorDirection(elevio.MD_Stop)
	}
	//fmt.Printf("\nOFA\n")
	CurrElev[pos].Floor = newFloor
	fmt.Printf("Info om heis %d: State: %d, Floor: %d\n", pos, CurrElev[pos].State, CurrElev[pos].Floor)
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
			// for i := 0; i < activeE; i++ {
			// CurrElev[i] = requests.ClearAtCurrentFloor(CurrElev[pos])
			// SetAllLights(CurrElev[i])
			// }
			//fmt.Printf( "requests: %+v\n", CurrElev[pos].Requests)
			//CurrElev[pos] = requests.ClearRequests(CurrElev[pos])
			//for i := 0; i < activeE; i++ {
			CurrElev[pos] = requests.ClearAtCurrentFloor(CurrElev[pos])
			//}
			CurrElev[pos] = requests.ClearRequests(CurrElev[pos])
			SetAllLights(CurrElev[pos])
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
