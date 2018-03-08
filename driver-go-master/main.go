package main

import "./elevio"

//import "../config"
import "fmt"
//import "time"
import "flag"
import "os"
import "../Network-go-master/network/bcast"
import "../Network-go-master/network/peers"
import "strconv"

//import "../Network-go-master/network/localip"
//project-proggeskuta/driver-go-master/
func main() {
	const (
		numFloors  = 4
		numButtons = 3
		numElevators = 3
	)

	activeElevs := numElevators
	activeElevs = 2 //becuase 2 in shell script atm

	port := ":" + os.Args[2]
	//elevio.Init(port, numFloors)
	elevio.Init(port, numFloors)

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
		ElevatorID   string
		OrderMatrix  [numFloors][numButtons - 1]int
		ButtonPushed [2]int


		//Iter        int
	}

	var id string

	//id = "5"
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

	var OM = [numFloors][numButtons - 1]int{}
	var AckMat = [numElevators][numFloors][numButtons - 1]int{}
	var BP = [2]int{}



	testmsg := ElevMsg{id, OM, BP}
	// go func() {
	//
	// 		for {
	// 			msgTrans <- testmsg
	// 			time.Sleep(4 * time.Second)
	// 		}
	// }()
	// go func() {
	// 	fmt.Printf("komme eg hit")
	// 	for i := 0; i < numFloors; i++{
	// 		fmt.Printf("inni go func: %#v\n", AckMat)
	// 		for j := 0; j < (numButtons-1); j++ {
	// 			if AckMat[i][j] == 1 {
	// 				elevio.SetButtonLamp(elevio.ButtonType(i), j, true)
	// 			}
	// 		}
	// 	}
	// }()


	for {
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			//elevio.SetButtonLamp(a.Button, a.Floor, true)
			testmsg.OrderMatrix[a.Floor][int(a.Button)] = 1
			testmsg.ButtonPushed[1] = int(a.Button)
			testmsg.ButtonPushed[0] = a.Floor
			msgTrans <- testmsg
			//fmt.Printf("Received: %#v\n", AckMat)
			//fmt.Println(testmsg.orderMatrix)
			//msgTrans <- testmsg

		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case a := <-msgRec:
			//fmt.Printf("Received: %#v\n", a)
			var ID int

			ID, _ = strconv.Atoi(a.ElevatorID)
			fmt.Printf("ID: %d\n", ID)
			AckMat[ID-1][a.ButtonPushed[0]][a.ButtonPushed[1]] = a.OrderMatrix[a.ButtonPushed[0]][a.ButtonPushed[1]]
			//fmt.Printf("Received: %#v\n", AckMat[ID-1])
			for i := 0; i < activeElevs; i++ {
				if AckMat[i][a.ButtonPushed[0]][a.ButtonPushed[1]] < AckMat[ID-1][a.ButtonPushed[0]][a.ButtonPushed[1]]{
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
			if counter == activeElevs{
				for i := 0; i < activeElevs; i++ {
					AckMat[i][a.ButtonPushed[0]][a.ButtonPushed[1]] = 2
				}
			}
			fmt.Printf("Received: %#v\n", AckMat[ID-1])
			//elevio.SetButtonLamp(elevio.ButtonType(a.ButtonPushed[1]), a.ButtonPushed[0], true)


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
