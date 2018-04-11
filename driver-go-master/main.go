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
)

func main() {

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
		orgsend int
		receiver int
	}

	type ElevMsg struct {
		ElevatorID   string
		ButtonPushed [2]int
		ElevList     [elevio.NumElevators]elevio.Elevator //liste med alle heisene
		ListPos      int                                  //posisjonen i ElevList, oppdateres i peerupdate.
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

	// msgTransElevator := make(chan elevio.Elevator)
	// msgRecElevator := make(chan elevio.Elevator)

	go bcast.Transmitter(25000, msgTrans, msgAckT )
	go bcast.Receiver(25000, msgRec, msgAckR)

	var pos int // blir oppdatert (nesten) med en gang heisen kommer online
	var sentmsg = ElevMsg{id, fsm.BP, fsm.CurrElev, pos}
	var ackmsg = AckMsg{}
	var correctAck bool


	for {
		select {
		case a := <-drv_buttons:

			if a.Button != 2 { //hvis det ikke er cab, så sender vi det
				//fmt.Printf("%+v\n", a)

				sentmsg.ButtonPushed[0] = a.Floor
				sentmsg.ButtonPushed[1] = int(a.Button)
				//msgTrans <- sentmsg
				for i := 0; i < 3; i++ {
						fmt.Printf("melding før trans\n")
						msgTrans <- sentmsg

						time.Sleep(5*time.Millisecond)
						fmt.Printf("melding før if set\n")
						if correctAck{
							fmt.Printf("WE HAVE ACKED!\n")
							break

						}
					}

					sentmsg.ElevList[pos].AcceptedOrders[a.Floor][a.Button] = 1
					msgTrans <- sentmsg




			} else { // cab trykk, oppdaterer Requests med en gang og setter lys
				sentmsg.ElevList[pos].Requests[a.Floor][a.Button] = true
				elevio.SetButtonLamp(a.Button, a.Floor, true)
				//fmt.Println("orders in queue: \n", sentmsg.ElevList.Requests)
				fsm.OnRequestButtonPress(a.Floor, a.Button, pos, activeElevs) //får den til å kjøre

			}

		case a := <- msgAckR:
			//fmt.Printf("kommer vi hit ? msgackr\n")
			if a.orgsend == pos{
					correctAck = true
				}


		case a := <-drv_floors: // til info så kjøres denne bare en gang når man kommer til en etasje, ikke loop
			fmt.Printf("case floors\n")
			fsm.OnFloorArrival(a, pos, activeElevs)
			//sentmsg.ElevList[pos].State = fsm.GetState(pos)

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
					//ackmsg.myid = pos
					//sentmsg.ListPos = pos //vett egentlig ikkje om elvator structen trenge Position men
					msgTrans <- sentmsg
					fmt.Printf("pos: %d\n", pos)
				}
				teller++
			}
			teller = 0

		// 	//msgAckT <- AckMsg{myid, a.sender}

		case a := <-msgRec:
			//fmt.Printf("kommer vi hit ? msgRec\n")
			ackmsg.orgsend = a.ListPos
			ackmsg.receiver = pos
			msgAckT <- ackmsg

			fsm.RecievedMSG(a.ButtonPushed[0], a.ButtonPushed[1], a.ListPos, a.ElevList[a.ListPos], activeElevs)
			sentmsg.ButtonPushed[0] = -10 // same as init value so we dont keep sending the same buttonpress forever
			// trengs egentlig bare når vi sender melidnger på heartbeat, ikke knappetrykk

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
