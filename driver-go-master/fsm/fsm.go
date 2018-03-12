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


func RecievedMSG(floor int, button int, position int, activeE int) {
	var index int
	fmt.Printf("posisjon. %d\n", position) // posisjon i lista, bør være 0 hvis du bare har en heis
	if floor != -10 { // se main:118
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
					CurrElev[i].AcceptedOrders[floor][button] = 1 // Er egentlig veldig skeptisk til denne måten å akseptere
																												// ordre på, siden vi bare sender en gang og så går alt på
																												// logiske operasjoner.
																												//har en mistanke om at hver heis som kjører lager sin egen
																												// CurrElev liste hvor elementene ikke nødvendigvis er like
																												//de andre heisene sin CurrElev liste...men vet ikke
					//fmt.Printf("Accepted av i : %d\n", CurrElev[i].AcceptedOrders)
				}
				index = costfunction.CostCalc(CurrElev, floor, button, activeE)
				fmt.Printf("index: %d\n", index)
				//kostfunksjonen vår! ser ikke ut til å gi rett index hver gang, men den er alltid den samme for begge heiser

				CurrElev[index].Requests[floor][button] = true
				for i := 0; i < activeE; i++ {
					//fmt.Printf("Accepted av i : %+v\n", CurrElev[i].Requests)
					AckMat[i][floor][button] = 0 //clearer ordren fra AckMat

				}
				//fmt.Printf("Detter er AckMat[1]: %+v \n", AckMat[1])
				SetAllLights(CurrElev[index]) //tror denne bare burde sette lys på en heis, men det settes på begge..

			}
		}
		for i := 0; i < elevio.NumFloors; i++ {
			for j := 0; j < elevio.NumButtons; j++ {
				if CurrElev[index].Requests[i][j] { //sjekker requests for den beste heisen, if true kjør OnRequestButtonPress
					OnRequestButtonPress(i, elevio.ButtonType(j), index)
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
			fmt.Printf("cab\n")
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

func OnRequestButtonPress(btn_floor int, btn_type elevio.ButtonType, pos int) {
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
			CurrElev[pos].Dir = requests.ChooseDirection(CurrElev[pos]) //tror denne får inn noe feil ved flere hesier
																																	//ettersom en av heisene nesten alltid kjører opp
			elevio.SetMotorDirection(CurrElev[pos].Dir)
			CurrElev[pos].State = elevio.Moving
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
				//fmt.Printf("Cleared av i: %+v\n", CurrElev[i].AcceptedOrders)
				SetAllLights(CurrElev[i]) // prøver å cleare lysene til alle heisene, hvis du får "index out of range" error
																	// så troooor jeg det er denne for-løkken, men aner ikke hvorfor
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
