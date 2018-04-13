package ordersync

import (
	"configFile"
	"elevfsm"
	"elevio"
	"elevorders"
	"fmt"
	"network/bcast"
	"network/peers"
	"time"
)

var activePeers []string

var acceptedOrders configFile.HallRequestType

type StatusMsg struct {
	Id                   string
	Status               elevfsm.ElevatorStatus
	CabRequests          configFile.CabRequestType
	HallRequestStatus    [configFile.NUM_FLOORS][configFile.NUM_DIR]RequestStatus
	PeersPendingRequests [configFile.NUM_FLOORS][configFile.NUM_DIR][]string
}

var networkOverview []StatusMsg
var myStatus StatusMsg

type RequestStatus int

const (
	requestEmpty     RequestStatus = 0
	requestPending                 = 1
	requestAccepted                = 2
	requestCompleted               = 3
)

type requestDir int

const (
	requestUp   requestDir = 0
	requestDown            = 1
)

func GetPeerState(thisPeer StatusMsg) elevfsm.ElevatorState {
	return thisPeer.Status.State
}

func GetPeerDirection(thisPeer StatusMsg) elevio.MotorDirection {
	return thisPeer.Status.Direction
}

func GetPeerCabRequests(thisPeer StatusMsg) configFile.CabRequestType {
	return thisPeer.CabRequests
}

func GetPeerFloor(thisPeer StatusMsg) int {
	return thisPeer.Status.Floor
}

func GetPeerId(thisPeer StatusMsg) string {
	return thisPeer.Id
}

func GetNetworkOverview() []StatusMsg {
	return networkOverview
}

func addControlPanelRequest(chNewOrder <-chan elevio.ButtonEvent, elevatorId string) {
	for {
		select {
		case newOrder := <-chNewOrder:
			fmt.Println("\n--> Order received in SYNC. Active peers in SYNC: ", activePeers)
			if newOrder.Button == elevio.BT_Cab {
				fmt.Println("\n------> CAB ORDER, quitting SYNC\n")
				break
			} else if myStatus.HallRequestStatus[newOrder.Floor][newOrder.Button] == requestEmpty || myStatus.HallRequestStatus[newOrder.Floor][newOrder.Button] == requestCompleted {
				fmt.Println("\n------> HALL ORDER registred in SYNC\n")
				myStatus.HallRequestStatus[newOrder.Floor][newOrder.Button] = requestPending
				myStatus.PeersPendingRequests[newOrder.Floor][newOrder.Button] = append(myStatus.PeersPendingRequests[newOrder.Floor][newOrder.Button], elevatorId)
				fmt.Println("------> HALL ORDER STATUS ARRAY in SYNC: ", myStatus.HallRequestStatus)
				fmt.Println("------> PEERS PENDING REQUEST ARRAY in SYNC: ", myStatus.PeersPendingRequests)
			}
		}
	}
}

func broadcastMyStatus(chStatusTx chan<- StatusMsg) {
	for {
		myStatus.Status = elevfsm.GetStatus()
		myStatus.CabRequests = elevorders.GetCabOrders()
		chStatusTx <- myStatus
		//fmt.Println("+++ Broadcasted my Status on the network")
		//fmt.Println("\n Active peers: ", activePeers, "\n")
		<-time.After(100 * time.Millisecond)
	}
}

func acceptRequest(floor int, requestDir int) {

	if floor < 0 || floor > configFile.NUM_FLOORS-1 || requestDir < 0 || requestDir > configFile.NUM_DIR-1 {
		return
	}
	myStatus.HallRequestStatus[floor][requestDir] = requestAccepted
	myStatus.PeersPendingRequests[floor][requestDir] = myStatus.PeersPendingRequests[floor][requestDir][:0]

	acceptedOrders[floor][requestDir] = true
	/* TODO: sett lys */
}

func updateRequestOverview(chStatusRx <-chan StatusMsg) {
	for {
		select {

		case receivedStatus := <-chStatusRx:
			//fmt.Println("+++ Received a Status on the network")
			for floorIter := 0; floorIter < configFile.NUM_FLOORS; floorIter++ {
				for dirIter := 0; dirIter < configFile.NUM_DIR; dirIter++ {

					//Must accept order if another elevator has accepted
					if receivedStatus.HallRequestStatus[floorIter][dirIter] == requestAccepted && myStatus.HallRequestStatus[floorIter][dirIter] != requestCompleted && myStatus.HallRequestStatus[floorIter][dirIter] != requestAccepted {
						acceptRequest(floorIter, dirIter)

						//Add pending request (if not accepted)
					} else if receivedStatus.HallRequestStatus[floorIter][dirIter] == requestPending && myStatus.HallRequestStatus[floorIter][dirIter] != requestAccepted {

						if !stringInSlice(receivedStatus.Id, myStatus.PeersPendingRequests[floorIter][dirIter]) {
							myStatus.PeersPendingRequests[floorIter][dirIter] = append(myStatus.PeersPendingRequests[floorIter][dirIter], receivedStatus.Id)
						}

						if myStatus.HallRequestStatus[floorIter][dirIter] == requestEmpty {
							myStatus.HallRequestStatus[floorIter][dirIter] = requestPending
						}

						//Add elevator Id's that are pending from incoming message to our list
						for _, peerId := range receivedStatus.PeersPendingRequests[floorIter][dirIter] {
							if !stringInSlice(peerId, myStatus.PeersPendingRequests[floorIter][dirIter]) && myStatus.HallRequestStatus[floorIter][dirIter] != requestAccepted {
								myStatus.PeersPendingRequests[floorIter][dirIter] = append(myStatus.PeersPendingRequests[floorIter][dirIter], peerId)
							}
						}

						// Set order as empty if another elevator has completed it
					} else if receivedStatus.HallRequestStatus[floorIter][dirIter] == requestCompleted && myStatus.HallRequestStatus[floorIter][dirIter] == requestAccepted {
						myStatus.HallRequestStatus[floorIter][dirIter] = requestCompleted
						elevio.SetButtonLamp(elevio.ButtonType(dirIter), floorIter, false)
					}

				}
			}
		}
	}
}

func transitionRequestStatus(chAcceptedOrders chan<- configFile.HallRequestType) {
	for {
		//fmt.Println("In transitionRequestStatus")
		//<-time.After(50*time.Millisecond)

		//fmt.Println("Active peers on network: ", activePeers)

		for floorIter := 0; floorIter < configFile.NUM_FLOORS; floorIter++ {
			//TODO

			if len(activePeers) < 2 {
				break
			}

			for dirIter := 0; dirIter < configFile.NUM_DIR; dirIter++ {

				peerCounter := 0

				for _, peerId := range activePeers {

					if stringInSlice(peerId, myStatus.PeersPendingRequests[floorIter][dirIter]) {
						peerCounter++
					}
				}

				//Accept if all active peers have registered the request, and there are more than 1 elevator on the network

				if peerCounter == len(activePeers) && peerCounter != 0 && myStatus.HallRequestStatus[floorIter][dirIter] != requestCompleted && myStatus.HallRequestStatus[floorIter][dirIter] != requestAccepted {
					acceptRequest(floorIter, dirIter)

					chAcceptedOrders <- acceptedOrders

					elevio.SetButtonLamp(elevio.ButtonType(dirIter), floorIter, true)
				}
			}
		}
	}
}

func stringInSlice(a string, slice []string) bool {
	for _, b := range slice {
		if b == a {
			return true
		}
	}
	return false
}

func updateNetworkOverview(chStatusRx <-chan StatusMsg) {
	for {

		select {
		case receivedStatus := <-chStatusRx:
			if !stringInSlice(receivedStatus.Id, activePeers) {
				activePeers = append(activePeers, receivedStatus.Id)
				networkOverview = append(networkOverview, receivedStatus)
				break
			}
			for i, thisPeer := range networkOverview {
				if receivedStatus.Id == thisPeer.Id {
					networkOverview[i] = receivedStatus
				}
			}
		default:
			//First make sure only active peers are in the overview list.
			for i, thisPeer := range networkOverview {
				if thisPeer.Id == myStatus.Id {
					networkOverview[i] = myStatus
				}
				if !stringInSlice(thisPeer.Id, activePeers) {
					networkOverview[i] = networkOverview[len(networkOverview)-1] // Replace it with the last one.
					networkOverview = networkOverview[:len(networkOverview)-1]   // Chop off the last one.
				}
			}
		}
	}
}

func updatePeers(chPeerUpdates <-chan peers.PeerUpdate) {
	for {
		select {
		case inputPeers := <-chPeerUpdates:
			activePeers = inputPeers.Peers
			//TODO: TEMPORARY FIX!
			//activePeers = append(activePeers, myStatus.Id)

			fmt.Println("\n/\\ \nActive peers: ", activePeers, "\n")
		}
	}
}

func InitializeSync(elevatorId string, chNewOrders <-chan elevio.ButtonEvent, chPeerUpdates <-chan peers.PeerUpdate, chAcceptedOrders chan<- configFile.HallRequestType) {

	myStatus.Id = elevatorId

	//TODO: TEMPORARY FIX!
	activePeers = append(activePeers, myStatus.Id)

	networkOverview = append(networkOverview, myStatus)

	go updatePeers(chPeerUpdates)

	chStatusTx := make(chan StatusMsg)
	chStatusRx := make(chan StatusMsg)
	chRequestOverviewRx := make(chan StatusMsg)
	chNetworkOverviewRx := make(chan StatusMsg)

	go bcast.Transmitter(30070, chStatusTx)
	go bcast.Receiver(30070, chStatusRx)

	go addControlPanelRequest(chNewOrders, elevatorId)

	go broadcastMyStatus(chStatusTx)

	go updateRequestOverview(chRequestOverviewRx)
	go updateNetworkOverview(chNetworkOverviewRx)

	go transitionRequestStatus(chAcceptedOrders)

	fmt.Print("Sync module initialized")

	for {
		select {
		case statusRx := <-chStatusRx:
			//fmt.Println("\n !!! Recieved Status update on network !!! \n")
			chRequestOverviewRx <- statusRx
			chNetworkOverviewRx <- statusRx
		}
	}
}
