package assigner

import (
	"configFile"
	"elevfsm"
	"elevio"
	"fmt"
	"network/peers"
	"ordersync"
	"os/exec"
	"strconv"
	"strings"
)

// [NUM_FLOORS][UP, DOWN]
var acceptedHallRequests configFile.HallRequestType

var elevatorId string

func DecideReassign(chPeerUpdate <-chan peers.PeerUpdate, chAcceptedOrders <-chan configFile.HallRequestType, chAssignedHallOrders chan<- configFile.HallRequestType) {
	for {
		select {
		case <-chPeerUpdate:
			fmt.Println("\n Peer update ---> reassignHallRequests")
			chAssignedHallOrders <- reassignHallRequests(acceptedHallRequests, ordersync.GetNetworkOverview(), elevatorId)

		case inputAcceptedHallRequests := <-chAcceptedOrders:

			acceptedHallRequests = inputAcceptedHallRequests
			fmt.Println("\n Accepted hall requests: ", acceptedHallRequests)

			assignedHallRequests := reassignHallRequests(acceptedHallRequests, ordersync.GetNetworkOverview(), elevatorId)
			fmt.Println("\n---Assigned hall requests: ", assignedHallRequests)
			chAssignedHallOrders <- assignedHallRequests

		}
	}
}

func InitializeOrderAssigner(inputElevatorId string) {
	elevatorId = inputElevatorId

}

func stateString(peer ordersync.StatusMsg) string {
	peerState := ordersync.GetPeerState(peer)

	if peerState == elevfsm.Elevator_Waiting || peerState == elevfsm.Elevator_Initializing {
		return `idle`
	} else if peerState == elevfsm.Elevator_Driving {
		return `moving`
	} else if peerState == elevfsm.Elevator_DoorOpen {
		return `doorOpen`
	}

	return ""
}

func directionString(peer ordersync.StatusMsg) string {
	peerDir := ordersync.GetPeerDirection(peer)

	if peerDir == elevio.MD_Down {
		return `down`
	} else if peerDir == elevio.MD_Up {
		return `up`
	} else if peerDir == elevio.MD_Stop {
		return `stop`
	}

	return ""
}

func cabRequestsString(peer ordersync.StatusMsg) string {
	returnString := ""

	peerCabRequests := ordersync.GetPeerCabRequests(peer)

	if peerCabRequests[0] == true {
		returnString += `true`
	} else {
		returnString += `false`
	}
	for i := 1; i < configFile.NUM_FLOORS; i++ {
		if peerCabRequests[i] == true {
			returnString += `,true`
		} else {
			returnString += `,false`
		}
	}
	return returnString
}

func hallRequestString(hallRequestFloor [2]bool) string {
	returnString := ""
	for i := 0; i < 2; i++ {
		if hallRequestFloor[i] == true {
			returnString += `true`
		} else {
			returnString += `false`
		}
		if i == 0 {
			returnString += `,`
		}
	}
	return returnString
}

func reassignHallRequests(hallRequests configFile.HallRequestType, networkOverview []ordersync.StatusMsg, elevatorId string) configFile.HallRequestType {
	fmt.Println("++++++++++++ Network overview from Assigner: ", networkOverview)
	//String to be passed to orderAssigner program
	orderAssignerArgument := ""

	//Must take all orders if alone on network
	if len(networkOverview) < 1 {
		return acceptedHallRequests
	}

	//Insert accepted hall requests
	orderAssignerArgument = `{"hallRequests":[[`
	orderAssignerArgument += hallRequestString(hallRequests[0])
	orderAssignerArgument += `]`

	for i := 1; i < len(hallRequests); i++ {
		orderAssignerArgument += `,[`
		orderAssignerArgument += hallRequestString(hallRequests[i])
		orderAssignerArgument += `]`
	}

	//Insert elevator states
	orderAssignerArgument += `],"states":{"` + ordersync.GetPeerId(networkOverview[0]) + `":{"behaviour":"`
	orderAssignerArgument += stateString(networkOverview[0])
	orderAssignerArgument += `","floor":` + strconv.Itoa(ordersync.GetPeerFloor(networkOverview[0])) + `,"direction":"`
	orderAssignerArgument += directionString(networkOverview[0])
	orderAssignerArgument += `","cabRequests":[`
	orderAssignerArgument += cabRequestsString(networkOverview[0])
	orderAssignerArgument += `]}`

	for i := 1; i < len(networkOverview); i++ {
		orderAssignerArgument += `,"` + ordersync.GetPeerId(networkOverview[i]) + `":{"behaviour":"`
		orderAssignerArgument += stateString(networkOverview[i])
		orderAssignerArgument += `","floor":` + strconv.Itoa(ordersync.GetPeerFloor(networkOverview[i])) + `,"direction":"`
		orderAssignerArgument += directionString(networkOverview[i])
		orderAssignerArgument += `","cabRequests":[`
		orderAssignerArgument += cabRequestsString(networkOverview[i])
		orderAssignerArgument += `]}`
	}
	orderAssignerArgument += `}}`

	output, err := exec.Command("./hall_request_assigner", "--input", orderAssignerArgument, "--clearRequestType", "all").CombinedOutput()
	if err != nil {
		fmt.Println(fmt.Sprint(err))
	}

	fmt.Println("\nOutput from hall assigner: ", string(output), "\n")

	var assignedHallRequests configFile.HallRequestType

	outputSlice := strings.Split(string(output), elevatorId)

	if len(outputSlice) > 1 {
		modifiedOutputString := outputSlice[1]

		//fmt.Println("HALL ORDERS SLICE OUTPUT: ", outputSlice)
		//fmt.Println("MODIFIED HALL ORDERS OUTPUT: ", modifiedOutputString)

		modifiedOutputString = modifiedOutputString[4:]

		for i := 0; i < 4; i++ {
			for j := 0; j < 2; j++ {
				if modifiedOutputString[0] == 'f' {
					assignedHallRequests[i][j] = false
					modifiedOutputString = modifiedOutputString[6:]
				} else if modifiedOutputString[0] == 't' {
					assignedHallRequests[i][j] = true
					modifiedOutputString = modifiedOutputString[5:]
				}
			}
			modifiedOutputString = modifiedOutputString[2:]
		}
	}

	//fmt.Println("RETURNED HALL ORDERS: ", assignedHallRequests)
	return assignedHallRequests
}
