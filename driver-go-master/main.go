package main

import (
	"flag"
	"fmt"
	"os"

	"../Network-go-master/network/bcast"
	"../Network-go-master/network/peers"
	"./elevio"
	"./fsm"
)

func main() {

	activeElevs := 1         // HAS to be non-zero initialized. Is however promptly updated to correct value bc of peerupdate
	port := ":" + os.Args[2] // this is nescessary to run the test.sh shell. if you want to run normally use the line below
	//elevio.Init("localhost:(same_port_as_in:_sim.con)", elevio.NumFloors)
	elevio.Init(port, elevio.NumFloors)

	fsm.Elev.State = elevio.Undefined
	//var d elevio.MotorDirection = elevio.MD_Stop

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
		elevio.SetMotorDirection(elevio.MD_Stop)
		fsm.Elev.State = elevio.Idle
	default:
		fsm.OnInitBetweenFloors()
	}

	type ElevMsg struct {
		ElevatorID   string
		ButtonPushed [2]int
		ThisElev     [elevio.NumElevators]elevio.Elevator
	}

	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15231, string(id), peerTxEnable)
	go peers.Receiver(15231, peerUpdateCh)

	msgTrans := make(chan ElevMsg)
	msgRec := make(chan ElevMsg)
	go bcast.Transmitter(25000, msgTrans)
	go bcast.Receiver(25000, msgRec)

	var pos int
	var sentmsg = ElevMsg{id, fsm.BP, fsm.CurrElev}

	// go func() {
	// 	for {
	// 		msgTrans <- sentmsg
	// 		time.Sleep(250 * time.Millisecond)
	// 	}
	// }()

	for {
		select {
		case a := <-drv_buttons:

			if a.Button != 2 {
				//fmt.Printf("%+v\n", a)
				//sentmsg.OrderMatrix[a.Floor][int(a.Button)] = 1
				sentmsg.ButtonPushed[1] = int(a.Button)
				sentmsg.ButtonPushed[0] = a.Floor
				msgTrans <- sentmsg
			} else {
				sentmsg.ThisElev[pos].Requests[a.Floor][a.Button] = true
				elevio.SetButtonLamp(a.Button, a.Floor, true)
				//fmt.Println("orders in queue: \n", sentmsg.ThisElev.Requests)
				fsm.OnRequestButtonPress(a.Floor, a.Button, pos)
				sentmsg.ThisElev[pos].Requests[a.Floor][a.Button] = true
			}

		case a := <-drv_floors:
			fsm.OnFloorArrival(a, pos, activeElevs)
			msgTrans <- sentmsg
			//fmt.Printf("loop or no?\n")

		case p := <-peerUpdateCh:
			peers.UpdatePeers(p)
			activeElevs = len(p.Peers)
			var teller int
			for _, i := range p.Peers {
				if i == id {
					pos = teller
					sentmsg.ThisElev[pos].Position = pos
					fmt.Printf("pos: %d\n", pos)

				}
				teller++
			}
			teller = 0

		case a := <-msgRec:
			fsm.RecievedMSG(a.ButtonPushed[0], a.ButtonPushed[1], a.ThisElev[pos], pos, activeElevs)
			sentmsg.ButtonPushed[0] = -10 // same as init value so we dont keep sending the same buttonpress forever

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} //else {
			// 	elevio.SetMotorDirection(d)
			// }

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < elevio.NumFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		case <-drv_timeout:
			fsm.OnDoorTimeout(pos)
		}
	}

}
