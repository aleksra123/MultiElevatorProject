package main

import "./elevio"

//import "../config"
import "fmt"
import "time"
import "flag"
import "../Network-go-master/network/bcast"
import "../Network-go-master/network/peers"

//import "../Network-go-master/network/localip"

func main() {
	const (
		numFloors  = 4
		numButtons = 3
	)
	elevio.Init("localhost:20016", numFloors)

	var d elevio.MotorDirection = elevio.MD_Up
	//elevio.SetMotorDirection(d)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	type ElevMsg struct {
		ElevatorID  string
		OrderMatrix [numFloors][numButtons - 1]int
		//Iter        int
	}

	// type AckMsg struct {
	// 	ThisElevID	string
	// 	AckedOrders [numFloors][numButtons-1]config.Acknowledge
	// }

	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15231, id, peerTxEnable)
	go peers.Receiver(15231, peerUpdateCh)

	msgTrans := make(chan ElevMsg)
	msgRec := make(chan ElevMsg)

	go bcast.Transmitter(25000, msgTrans)
	go bcast.Receiver(25000, msgRec)

	var OM = [numFloors][numButtons - 1]int{}

	testmsg := ElevMsg{id, OM}
	go func() {

		for {
			msgTrans <- testmsg
			time.Sleep(4 * time.Second)
		}
	}()

	for {
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			elevio.SetButtonLamp(a.Button, a.Floor, true)
			testmsg.OrderMatrix[a.Floor][a.Button] = 1
			//fmt.Println(testmsg.orderMatrix)
			//msgTrans <- testmsg

		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case a := <-msgRec:
			fmt.Printf("Received: %#v\n", a)

		// case a := <-drv_floors:
		// 	fmt.Printf("%+v\n", a)
		// 	if a == numFloors-1 {
		// 		d = elevio.MD_Down
		// 	} else if a == 0 {
		// 		d = elevio.MD_Up
		// 	}
		// 	elevio.SetMotorDirection(d)

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
