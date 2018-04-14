package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"../Network-go-master/network/bcast"
	"../Network-go-master/network/peers"
	"./elevio"
	"./fsm"
	//"./requests"
)



func main() {

	for i := 0; i < 20; i++ {
		fmt.Printf("\n")
	}


	// Lettest å kjøre test.sh for å komme i gang, men gjerne koble ut den ene heisen ettersom koden
	// foreløpig bare funker med en. hehe

	activeElevs := 1         // HAS to be non-zero initialized. Is however promptly updated to correct value bc of peerupdate
	port := ":" + os.Args[2] // this is nescessary to run the test.sh shell. if you want to run normally use the line below
	//elevio.Init("localhost:(same_port_as_in:_sim.con)", elevio.NumFloors)
	elevio.Init(port, elevio.NumFloors)

	fsm.Elev.State = elevio.Undefined
	//var d elevio.MotorDirection = elevio.MD_Stop

	//Event channels
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

	type AckMsg struct {
		Orgsend int
		Receiver int
	}

	type ElevMsg struct {
		ElevatorID   string
		ButtonPushed [2]int
		ElevList     [elevio.NumElevators]elevio.Elevator //liste med alle heisene
		ListPos      int
		Msgtype      int    // 1 e BP, 2 e floor og 3 e pos
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

	msgAckT := make(chan AckMsg)
	msgAckR := make(chan AckMsg)
	//correctAck:= make(chan bool)

	// msgTransElevator := make(chan elevio.Elevator)
	// msgRecElevator := make(chan elevio.Elevator)

	go bcast.Transmitter(25000, msgTrans, msgAckT )
	go bcast.Receiver(25000, msgRec, msgAckR)

	var pos int // blir oppdatert (nesten) med en gang heisen kommer online
	var sentmsg = ElevMsg{}
	sentmsg.ButtonPushed = fsm.BP
	sentmsg.ElevList = fsm.CurrElev
	var ackmsg = AckMsg{}
	var CorrectAck bool = false
	//go ready()

	for {
		select {
		case a := <-drv_buttons:

			if a.Button != 2 {

				sentmsg.ButtonPushed[0] = a.Floor
				sentmsg.ButtonPushed[1] = int(a.Button)
				//sentmsg.Msgtype = 1
				sentmsg.Msgtype = 3
				//msgTrans <- sentmsg
				for i := 0; i < 3; i++ {
						msgTrans <- sentmsg
						time.Sleep(5*time.Millisecond)

					if  CorrectAck{
							fmt.Printf("WE HAVE ACKED!\n")
							CorrectAck = false
							break
					}
				}

			} else {
				sentmsg.ElevList[pos].Requests[a.Floor][a.Button] = true
				elevio.SetButtonLamp(a.Button, a.Floor, true)
				fsm.OnRequestButtonPress(a.Floor, a.Button, pos, activeElevs, pos)
			}

		case a := <- msgAckR:
			if a.Orgsend == pos{
				CorrectAck = true
				}

		case a := <-drv_floors: // til info så kjøres denne bare en gang når man kommer til en etasje, ikke loop
			sentmsg.Msgtype = 2
			sentmsg.ElevList[pos].Floor = a
			msgTrans <- sentmsg


		case p := <-peerUpdateCh:
			peers.UpdatePeers(p)
			activeElevs = len(p.Peers)
			var teller int
			for _, i := range p.Peers {
				if i == id {
					pos = teller
					sentmsg.ListPos = pos
					sentmsg.ElevList[pos].Position = pos
					sentmsg.Msgtype = 3
					msgTrans <- sentmsg
					fmt.Printf("pos: %d\n", pos)
				}
				teller++
			}
			teller = 0


		case a := <-msgRec:
			ackmsg.Orgsend = a.ListPos
			ackmsg.Receiver = pos
			msgAckT <- ackmsg


			if a.Msgtype == 2 {
				fsm.OnFloorArrival(a.ElevList[a.ListPos].Floor, a.ListPos, activeElevs, pos)
			}


			if a.Msgtype == 3{
				if a.ButtonPushed[0] != -10 {
					fmt.Printf("bp av 0, %d\n", a.ButtonPushed)
					for i := 0; i < activeElevs; i++ {
						sentmsg.ElevList[i].AcceptedOrders[a.ButtonPushed[0]][a.ButtonPushed[1]] = 1
						a.ElevList[i].AcceptedOrders[a.ButtonPushed[0]][a.ButtonPushed[1]] = 1
					}


					//fmt.Printf("AO: %+v\n", a.ElevList[a.ListPos].AcceptedOrders)
				}

				fsm.RecievedMSG(a.ButtonPushed[0], a.ButtonPushed[1], a.ListPos, a.ElevList[a.ListPos], activeElevs, pos)
				sentmsg.ButtonPushed[0] = -10 // same as init value so we dont keep sending the same buttonpress forever
			}

			if a.Msgtype == 4 {
				fsm.OnDoorTimeout(a.ListPos, pos)
			}

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
			fmt.Printf("blir dette printet hos begge 2?\n")
			 sentmsg.ListPos = pos
			 sentmsg.Msgtype = 4
			 msgTrans <- sentmsg

		}
	}

}
