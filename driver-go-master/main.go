package main

import (
	"flag"
	"fmt"
	//"os"
	"time"

	"../Network-go-master/network/bcast"
	"../Network-go-master/network/peers"
	"./elevio"
	"./fsm"
	"./requests"
)



func main() {

	for i := 0; i < 20; i++ {
		fmt.Printf("\n")
	}


	activeElevs := 1         // HAS to be non-zero initialized. Is however promptly updated to correct value bc of peerupdate
	//port := ":" + os.Args[2] // this is nescessary to run the test.sh shell. if you want to run normally use the line below
	elevio.Init("localhost:15657", elevio.NumFloors)
	//elevio.Init(port, elevio.NumFloors)

	fsm.Elev.State = elevio.Undefined
	//var d elevio.MotorDirection = elevio.MD_Stop

	//Event channels
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	drv_timeout := fsm.Door_timer.C
	drv_powerout := fsm.Power_timer.C

	fsm.Door_timer.Stop()
	fsm.Power_timer.Stop()


	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)


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

	go bcast.Transmitter(25000, msgTrans, msgAckT )
	go bcast.Receiver(25000, msgRec, msgAckR)

	var initialized bool = false
	// fsm.Start_blank()
	var pos int = 0 // blir oppdatert (nesten) med en gang heisen kommer online
	fsm.Init(pos)
	var sentmsg = ElevMsg{}
	sentmsg.Msgtype = -1
	sentmsg.ButtonPushed = fsm.BP
	sentmsg.ElevList = fsm.CurrElev
	var ackmsg = AckMsg{}
	var CorrectAck bool = false
	//go ready()

	for {
		select {
		case a := <-drv_buttons:
			if pos != -1 && !fsm.CurrElev[pos].FirstTime{
				if a.Button != 2 {

					sentmsg.ButtonPushed[0] = a.Floor
					sentmsg.ButtonPushed[1] = int(a.Button)
					sentmsg.Msgtype = 3
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
					sentmsg.ButtonPushed[0] = a.Floor
					sentmsg.ButtonPushed[1] = int(a.Button)
					sentmsg.Msgtype = 1
					msgTrans <- sentmsg
					elevio.SetButtonLamp(a.Button, a.Floor, true)
					//fsm.OnRequestButtonPress(a.Floor, a.Button, pos, activeElevs, pos)
				}
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
					sentmsg.Msgtype = 6
					sentmsg.ElevList[pos].Position = pos
					msgTrans <- sentmsg
					if !initialized {
						fmt.Printf("test\n")
						fsm.Init(pos)
						sentmsg.Msgtype = 7
						msgTrans <- sentmsg
						initialized = true
					}



					fmt.Printf("pos: %d\n", pos)
				}
				teller++
			}
			teller = 0


		case a := <-msgRec:
			ackmsg.Orgsend = a.ListPos
			ackmsg.Receiver = pos
			msgAckT <- ackmsg

			if a.Msgtype == 7 {
				fsm.Online(a.ListPos, pos)
			}

			if a.Msgtype == 1 {
				fsm.AddCabRequest(a.ListPos, a.ButtonPushed[0])
				fsm.OnRequestButtonPress(a.ButtonPushed[0], elevio.ButtonType(a.ButtonPushed[1]), a.ListPos, activeElevs, pos)
			}

			if a.Msgtype == 2 {
				fsm.OnFloorArrival(a.ElevList[a.ListPos].Floor, a.ListPos, activeElevs, pos)
			}

			if a.Msgtype == 3{

				if a.ButtonPushed[0] != -10 {
					fmt.Printf("fra heis:  %d\n", a.ListPos)
					for i := 0; i < activeElevs; i++ {
						sentmsg.ElevList[i].AcceptedOrders[a.ButtonPushed[0]][a.ButtonPushed[1]] = 1
						a.ElevList[i].AcceptedOrders[a.ButtonPushed[0]][a.ButtonPushed[1]] = 1
					}
				}
				fsm.RecievedMSG(a.ButtonPushed[0], a.ButtonPushed[1], a.ListPos, a.ElevList[a.ListPos], activeElevs, pos)
				sentmsg.ButtonPushed[0] = -10 // same as init value so we dont keep sending the same buttonpress forever
			}

			if a.Msgtype == 4 {
				fsm.OnDoorTimeout(a.ListPos, pos )
			}

			if a.Msgtype == 5 {
				fsm.Powerout(a.ListPos, activeElevs)
			}

			if a.Msgtype == 6 {
				fsm.Updatepos(a.ListPos)
			}




		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(requests.ChooseDirection(fsm.CurrElev[pos]))
				fsm.Power_timer.Stop()
				peerTxEnable <- true
				// lag en reboot function i fsm
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < elevio.NumFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}

		case <-drv_timeout:

			 sentmsg.ListPos = pos
			 //sentmsg.ElevList[pos].State = elevio.DoorOpen
			 sentmsg.Msgtype = 4
			 msgTrans <- sentmsg

		case <- drv_powerout:
			sentmsg.ListPos = pos
			sentmsg.Msgtype = 5
			msgTrans <- sentmsg
			peerTxEnable <- false
			pos = -1

		}
	}

}
