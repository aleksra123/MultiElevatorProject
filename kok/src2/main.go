package main

import (
	"assigner"
	"configFile"
	"elevfsm"
	"elevio"
	"elevorders"
	"fmt"
	"network/localip"
	"network/peers"
	"ordersync"
	"os"
)

func main() {

	var id string

	var elevaddr string

	if len(os.Args) < 2 {
		elevaddr = `localhost:15657`
	} else {
		elevaddr = `localhost:` + os.Args[1]
		//elevaddr = `localhost:44194`
	}

	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	fmt.Println("************************************** \n Elevaddr: ", elevaddr, "\n Elevator ID: ", id, "\n**************************************\n")

	fmt.Print("")

	peersTxPort := 49091
	peersRxPort := 49091

	chNewOrderElev := make(chan elevio.ButtonEvent)
	chNewOrderSync := make(chan elevio.ButtonEvent)

	chAcceptedHallOrders := make(chan configFile.HallRequestType)
	chAssignedHallOrders := make(chan configFile.HallRequestType)

	chPeerUpdate := make(chan peers.PeerUpdate)
	chPeerUpdateAssigner := make(chan peers.PeerUpdate)
	chPeerUpdateSync := make(chan peers.PeerUpdate)

	chOrderServed := make(chan int)

	chTxEnable := make(chan bool)

	go peers.Transmitter(peersTxPort, id, chTxEnable)
	go peers.Receiver(peersRxPort, chPeerUpdate)

	chTxEnable <- true

	elevfsm.Init(elevaddr, chNewOrderElev)
	go elevfsm.StateMachine(chNewOrderElev, chNewOrderSync, chOrderServed)
	go elevfsm.PrintFSMstate()

	elevorders.InitElevatorOrders(chNewOrderElev, chAssignedHallOrders)
	go elevorders.RunElevatorOrders(chNewOrderElev, chAssignedHallOrders)

	assigner.InitializeOrderAssigner(id)
	go assigner.DecideReassign(chPeerUpdateAssigner, chAcceptedHallOrders, chAssignedHallOrders)

	go ordersync.RunSync(id, chNewOrderSync, chPeerUpdateSync, chAcceptedHallOrders, chOrderServed)

	fmt.Println("\nElevator software now running \n")

	for {
		select {
		case peerUpdate := <-chPeerUpdate: //<-chPeerUpdate:
			//fmt.Println("///////////////////// \n Peer update: ", peerUpdate.Peers, "///////////////////// \n")
			chPeerUpdateAssigner <- peerUpdate
			chPeerUpdateSync <- peerUpdate
			//fmt.Println("\n ======================================================= \n")

		}
	}
}
