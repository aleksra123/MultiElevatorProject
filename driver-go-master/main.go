package main

import "./elevio"
import "fmt"
import "time"
import "flag"
import "os"
import "../Network-go-master/network/bcast"
import "../Network-go-master/network/peers"
import "strconv"
import "./fsm"
import "./costfunction"

func main() {
	const (
		numFloors    = 4
		numButtons   = 3
		numElevators = 3
	)
	activeElevs := numElevators //needs to be non-zero initialized, changes when peers are updated (see case p := <- peerUpdateCh)
	port := ":" + os.Args[2]    // this is nescessary to run the test.sh shell. if you want to run normally use the line below
	//elevio.Init("localhost:(same_port_as_in:_sim.con)", numFloors)
	elevio.Init(port, numFloors)
	fsm.Elev.State = elevio.Undefined
	var d elevio.MotorDirection = elevio.MD_Stop

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	drv_timeout := fsm.Door_timer.C

	fsm.Elev.State = elevio.Undefined

	select { //Init
	case <-drv_floors:
		elevio.SetMotorDirection(elevio.MD_Stop)
		fsm.Elev.State = elevio.Idle
	default:
		fsm.OnInitBetweenFloors()
	}

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	type ElevMsg struct {
		ElevatorID   string
		OrderMatrix  [numFloors][numButtons - 1]int
		ButtonPushed [2]int
		ThisElev     elevio.Elevator
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

	var elevlist = [numElevators]elevio.Elevator{}
	var OM = [numFloors][numButtons - 1]int{}
	var AckMat = [numElevators][numFloors][numButtons - 1]int{}
	var BP = [2]int{}
	var AccOrders = [numFloors][numButtons - 1]int{} //Accepted orders, alle heisene ser samme, denne inneholder ikke cab orders.
	// Cab orders blir lagt direkte inn i ThisElev.Queue. Hver heis har sin egen
	// queue og sender ikke den til noen (foreløpig). AccOrders må sendes til vår (hentes inn av)  kommende kostfunksjon.
	var CurrElev elevio.Elevator
	var ID int
	ID, _ = strconv.Atoi(id)
	sentmsg := ElevMsg{id, OM, BP, CurrElev}
	//lololo
	go func() {
		for {
			for i := 0; i < numFloors; i++ {
				for j := 0; j < numButtons-1; j++ {
					var counter int
					for k := 0; k < activeElevs; k++ {
						if AckMat[k][i][j] == 2 {
							counter++
						}
					}
					if counter == activeElevs {
						AccOrders[i][j] = 1
						costfunction.CostCalc(elevlist, i, j, activeElevs)
						fmt.Printf("elvlist av 0: %+v\n", elevlist) // elevlist does not keep changes after being run in costfunction (make it work!)
					}
				}
			}
			fmt.Println(AccOrders)
			fsm.Elev.AcceptedOrders = AccOrders

			time.Sleep(4 * time.Second)
		}
	}()

	for {
		select {
		case a := <-drv_buttons:
			if a.Button != 2 {
				fmt.Printf("%+v\n", a)
				sentmsg.OrderMatrix[a.Floor][int(a.Button)] = 1
				sentmsg.ButtonPushed[1] = int(a.Button)
				sentmsg.ButtonPushed[0] = a.Floor
				msgTrans <- sentmsg
			} else {
				sentmsg.ThisElev.Requests[a.Floor][a.Button] = true
				elevio.SetButtonLamp(a.Button, a.Floor, true)
				fmt.Println("orders in queue: \n", sentmsg.ThisElev.Requests)
				fsm.OnRequestButtonPress(a.Floor, a.Button)
				elevlist[ID].Requests[a.Floor][a.Button] = true

			}

		case a := <-drv_floors:
			fsm.OnFloorArrival(a)
			fmt.Printf("loop or no?\n")

		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			fmt.Printf(" Length of Peers:    %d\n", len(p.Peers))
			activeElevs = len(p.Peers)
			//fmt.Println("antall peers: %s\n",antallpeers)

		case a := <-msgRec:
			fmt.Printf("ID: %d\n", ID)
			if AckMat[ID-1][a.ButtonPushed[0]][a.ButtonPushed[1]] != 2 {
				AckMat[ID-1][a.ButtonPushed[0]][a.ButtonPushed[1]] = a.OrderMatrix[a.ButtonPushed[0]][a.ButtonPushed[1]]
				//fmt.Printf("Received: %#v\n", AckMat[ID-1])
				for i := 0; i < activeElevs; i++ {
					if AckMat[i][a.ButtonPushed[0]][a.ButtonPushed[1]] == AckMat[ID-1][a.ButtonPushed[0]][a.ButtonPushed[1]]-1 {
						AckMat[i][a.ButtonPushed[0]][a.ButtonPushed[1]]++
						fmt.Printf("we incremented! \n")
					}
				}
				var counter int
				for i := 0; i < activeElevs; i++ {
					if AckMat[i][a.ButtonPushed[0]][a.ButtonPushed[1]] == 1 {
						counter++
					}
				}
				if counter == activeElevs {
					for i := 0; i < activeElevs; i++ {
						AckMat[i][a.ButtonPushed[0]][a.ButtonPushed[1]] = 2
					}
				}
			}
			elevio.SetButtonLamp(elevio.ButtonType(a.ButtonPushed[1]), a.ButtonPushed[0], true)

			fmt.Printf("Received: %#v\n", AckMat[ID-1])

			for i := 0; i < numFloors; i++ {
				for j := 0; j < numButtons; j++ {
					if elevlist[ID].Requests[i][j] {
						fmt.Printf("rett før anders cumme")
						fsm.OnRequestButtonPress(i, elevio.ButtonType(j))
					}
				}
			}

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
		case <-drv_timeout:
			fsm.OnDoorTimeout()
		}
	}

}
