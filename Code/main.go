package main

import (
	"fmt"
	"os"

	"./elevio"
	"./fsm"
)

func main() {
	port := ":" + os.Args[2]
	elevio.Init(port)
	// elevio.Init("localhost:20019") //200+arb.plass or 15657

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	// drv_obstr := make(chan bool)
	// drv_stop := make(chan bool)
	drv_timeout := fsm.Door_timer.C

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	// go elevio.PollObstructionSwitch(drv_obstr)
	// go elevio.PollStopButton(drv_stop)

	select { //Init
		case <-drv_floors:
			elevio.SetMotorDirection(elevio.MD_Stop)
			fsm.Elev.State = elevio.Idle
		default:
			fsm.OnInitBetweenFloors()
	}

	for {
		fmt.Println("Current state:", fsm.StateToString(fsm.Elev))
		select {
		case a := <-drv_buttons:
			fsm.OnRequestButtonPress(a.Floor, a.Button)

		case a := <-drv_floors:
			fsm.OnFloorArrival(a)

		// Is this a part of the assigment?
		// case a := <-drv_obstr:
		// 	fmt.Printf("%+v\n", a)
		// 	if a {
		// 		elevio.SetMotorDirection(elevio.MD_Stop)
		// 	} else {
		// 		fsm.Elev.Dir = requests.ChooseDirection(fsm.Elev)
		// 	}
		//
		// case a := <-drv_stop:
		// 	fmt.Printf("%+v\n", a)
		// 	for f := 0; f < elevio.NumFloors; f++ {
		// 		for b := elevio.ButtonType(0); b < 3; b++ {
		// 			elevio.SetButtonLamp(b, f, false)
		// 		}
		// 	}

		case <-drv_timeout:
			fsm.OnDoorTimeout()
		}
	}
}
