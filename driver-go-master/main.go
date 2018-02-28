package main

import "./elevio"
import "fmt"
import "../Network-go-master/network/bcast"

//import "../Network-go-master/network/localip"
//import "../Network-go-master/network/peers"

func main() {
	const (
		numFloors  = 4
		numButtons = 3
	)
	elevio.Init("localhost:15657", numFloors)

	var d elevio.MotorDirection = elevio.MD_Up
	elevio.SetMotorDirection(d)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	type HelloMsg struct {
		Elevator    int
		orderMatrix [numFloors]int
		Iter        int
	}

	msgTrans := make(chan HelloMsg)
	msgRec := make(chan HelloMsg)

	go bcast.Transmitter(25000, msgTrans)
	go bcast.Receiver(25000, msgRec)

	OM := [numFloors]int{}

	testmsg := HelloMsg{1, OM, 3}
	//fmt.Println(testmsg.orderMatrix)

	// for {
	// 	select {
	// 	case a := <-msgRec:
	// 		fmt.Printf("Received: %#v\n", a)
	// 	}
	// }
	//var OrderQueue = []elevio.ButtonEvent{}

	for {
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			elevio.SetButtonLamp(a.Button, a.Floor, true)
			//OrderQueue = append(OrderQueue, a)
			//fmt.Println(OrderQueue)
			OM[a.Floor] = 1
			testmsg.orderMatrix[a.Floor] = OM[a.Floor]
			//fmt.Println(testmsg.orderMatrix)
			testmsg.Elevator = 5
			testmsg.Iter = 100
			msgTrans <- testmsg

		case a := <-msgRec:
			fmt.Printf("Received: %#v\n", a)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			if a == numFloors-1 {
				d = elevio.MD_Down
			} else if a == 0 {
				d = elevio.MD_Up
			}
			elevio.SetMotorDirection(d)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(d)
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		}
	}
}
