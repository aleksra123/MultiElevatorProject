package main

import (
	"flag"
	"time"
	"strconv"

	"./network/bcast"
	"./network/peers"
	"./elevio"
	"./fsm"
	//"./backup"

)
//
// This system consists of 6 modules. The network module, the fsm(finite state machine) module, the request module,
// the elevio module and the costfunction module all have their roots in the given project resources material. However the fsm has been altered quite
// substantially to handle several elevators, and the request, elevio and costfunction modules have been tweaked to fit our needs.
// The backup module writes caborders to file, and allows us to retrieve them after a reboot.

func main() {

	var activeElevs = 0
	//port := ":" + os.Args[2]
	//elevio.Init(port, elevio.NumFloors) // this is nescessary to run the test.sh shell.
	elevio.Init("localhost:15657", elevio.NumFloors)


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
		Orgsender int
		Receiver int
	}

	type ElevMsg struct {
		ElevatorID   string
		ButtonPushed [2]int
		ElevList     [elevio.NumElevators]elevio.Elevator
		ListPos      int
		Msgtype      int
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
	var pos int = -1

	var sentmsg = ElevMsg{}
	sentmsg.Msgtype = -1
	sentmsg.ElevList = fsm.CurrElev
	var ackmsg = AckMsg{}
	var CorrectAck bool = false
	var MotorFailure bool

	for {
		select {
		case a := <-drv_buttons:
			if pos != -1 && !fsm.CurrElev[pos].FirstTime{
				if a.Button != 2  && activeElevs > 1{

					sentmsg.ButtonPushed[0] = a.Floor
					sentmsg.ButtonPushed[1] = int(a.Button)
					sentmsg.Msgtype = 3
					for i := 0; i < 10; i++ {
						msgTrans <- sentmsg
						time.Sleep(5*time.Millisecond)
						if  CorrectAck{
							CorrectAck = false
							break
						}
					}

				} else if a.Button == 2{

					elevio.SetButtonLamp(a.Button, a.Floor, true)
					//fsm.OnRequestButtonPress(a.Floor, a.Button, pos, activeElevs, pos)
					//backup.UpdateBackup(fsm.CurrElev[pos])
					sentmsg.ButtonPushed[0] = a.Floor
					sentmsg.ButtonPushed[1] = int(a.Button)
					sentmsg.ListPos = pos
					sentmsg.Msgtype = 1
					for i := 0; i < 10; i++ {
						msgTrans <- sentmsg
						time.Sleep(5*time.Millisecond)
						if  CorrectAck{
							CorrectAck = false
							break
						}
					}
				}
			}

		case a := <- msgAckR:
			if a.Orgsender == pos{
				CorrectAck = true
				}

		case a := <-drv_floors:
			//fsm.OnFloorArrival(a.Floor, pos, activeElevs, pos)
			sentmsg.Msgtype = 2
			if pos != -1 {
				sentmsg.ListPos = pos
			}
			sentmsg.ElevList[pos].Floor = a
			for i := 0; i < 10; i++ {
				msgTrans <- sentmsg
				time.Sleep(5*time.Millisecond)
				if  CorrectAck{
					CorrectAck = false
					break
				}
			}
			// if MotorFailure{
			// 	peerTxEnable <- true
			// 	MotorFailure = false
			// }

		case p := <-peerUpdateCh:
			peers.UpdatePeers(p)
			prev := activeElevs
			activeElevs = len(p.Peers)
			var teller int

			if prev > activeElevs  {
				lost, _ :=  strconv.Atoi(p.Lost[0])
				fsm.TransferRequests(lost, activeElevs, pos)
				fsm.CopyInfo_Lost(lost, activeElevs)
			}

			if activeElevs > 1 { // && prev < activeElevs må ha med!
				new, _ := strconv.Atoi(p.New)
				fsm.CopyInfo_New(new, activeElevs )
			}

			for _, i := range p.Peers {
				if i == id {
					pos = teller
					sentmsg.ListPos = pos
					sentmsg.Msgtype = 6
					sentmsg.ElevList[pos].Position = pos
					for i := 0; i < 10; i++ {
						msgTrans <- sentmsg
						time.Sleep(5*time.Millisecond)
						if  CorrectAck{
							CorrectAck = false
							break
						}
					}
					if !initialized {
						fsm.Init(pos, activeElevs) // må flyttes ???
						initialized = true
						sentmsg.Msgtype = 7
						for i := 0; i < 10; i++ {
							msgTrans <- sentmsg
							time.Sleep(5*time.Millisecond)
							if  CorrectAck{
								CorrectAck = false
								break
							}
						}
					}
				}
				teller++
			}
			teller = 0


		case a := <-msgRec:
			ackmsg.Orgsender = a.ListPos
			ackmsg.Receiver = pos
			msgAckT <- ackmsg
			// for i := 0; i < 5; i++ {
			// 	msgAckT <- ackmsg
			// }



			if a.Msgtype == 1 {
				fsm.AddCabRequest(a.ListPos, a.ButtonPushed[0], pos)
				// if a.ListPos != pos {
				fsm.OnRequestButtonPress(a.ButtonPushed[0], elevio.ButtonType(a.ButtonPushed[1]), a.ListPos, activeElevs, pos)
				//}
			}

			if a.Msgtype == 2 {
				// if a.ListPos != pos {
				fsm.OnFloorArrival(a.ElevList[a.ListPos].Floor, a.ListPos, activeElevs, pos)
				//}
			}

			if a.Msgtype == 3{
				for i := 0; i < activeElevs; i++ {
					sentmsg.ElevList[i].AcceptedOrders[a.ButtonPushed[0]][a.ButtonPushed[1]] = 1
					a.ElevList[i].AcceptedOrders[a.ButtonPushed[0]][a.ButtonPushed[1]] = 1
				}
				fsm.RecievedMSG(a.ButtonPushed[0], a.ButtonPushed[1], a.ListPos, a.ElevList[a.ListPos], activeElevs, pos)
			}

			if a.Msgtype == 4 {
				// if a.ListPos != pos {
				fsm.OnDoorTimeout(a.ListPos, pos )
				//}
			}

			if a.Msgtype == 5 {
				fsm.Powerout(a.ListPos, activeElevs)
			}

			if a.Msgtype == 6 {
				fsm.Updatepos(a.ListPos)
			}

			if a.Msgtype == 7 {
				fsm.Online(a.ListPos, pos)

			}

		case a := <-drv_obstr:
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
				//fsm.Power_timer.Stop()
			}

		case a := <-drv_stop:
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
				fsm.Power_timer.Stop()
			}

		case <-drv_timeout:
			 //fsm.OnDoorTimeout(pos, pos )
		   sentmsg.ListPos = pos
			 sentmsg.Msgtype = 4
			 msgTrans <- sentmsg


		case <- drv_powerout:
			fsm.Power_timer.Stop()
			MotorFailure = true
			initialized = false
			peerTxEnable <- false
			return



		}
	}

}
