package main

import (
	"fmt"

	"./elevio"
	"./fsm"
	"./requests"
)

func main() {

	elevio.Init("localhost:15657") //200+arb.plass or 15657
	fsm.Elev.State = elevio.Undefined

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	drv_timeout := fsm.Door_timer.C

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	select { //Init
	case <-drv_floors:
		fmt.Println("case 1")
		elevio.SetMotorDirection(elevio.MD_Stop)
		fsm.Elev.State = elevio.Idle
	default:
		fmt.Println("case default")
		//fsm.OnInitBetweenFloors()
		elevio.SetMotorDirection(elevio.MD_Down)
		fsm.Elev.Dir = elevio.MD_Down
		fsm.Elev.State = elevio.Moving
	}

	for {
		select {
		case a := <-drv_buttons:
			fsm.OnRequestButtonPress(a.Floor, a.Button)

			//fmt.Printf("%+v\n", a)

		case a := <-drv_floors:
			fsm.OnFloorArrival(a)

			// fmt.Printf("%+v\n", a)
			// if a == numFloors-1 {
			//     d = elevio.MD_Down
			// } else if a == 0 {
			//     d = elevio.MD_Up
			// }
			// elevio.SetMotorDirection(d)

		case a := <-drv_obstr: //Is this a part of the assigment?
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				fsm.Elev.Dir = requests.ChooseDirection(fsm.Elev)
			}

		case a := <-drv_stop: //Is this a part of the assigment?
			fmt.Printf("%+v\n", a)
			for f := 0; f < elevio.NumFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}

		case <-drv_timeout:
			fsm.OnDoorTimeout()
		default:
			//fmt.Println(fsm.Elev.State)
		}
	}
}
