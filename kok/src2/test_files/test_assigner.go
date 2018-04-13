package test_files

/*
import (
	"elevfsm"
	"strconv"
	"elevio"
	"fmt"
	"os/exec"
	"strings"
	"ordersync"
)

type hallRequestsType [4][2]bool

//String to be passed to orderAssigner program
var orderAssignerArgument string
// [NUM_FLOORS][UP, DOWN]
var assignedHallRequests hallRequestsType

func main() {

	//VARIABLES DECLARED FOR TEST PURPOSES
	var test_hallRequests hallRequestsType
	var statusOverview []ordersync.StatusMsg

	elevatorId := "two"

	test_hallRequests[0][0] = false
	test_hallRequests[0][1] = false
	test_hallRequests[1][0] = true
	test_hallRequests[1][1] = false
	test_hallRequests[2][0] = false
	test_hallRequests[2][1] = false
	test_hallRequests[3][0] = false
	test_hallRequests[3][1] = true

	statusOverview[0].id="one"
	statusOverview[0].status.State = elevfsm.Elevator_Driving
	statusOverview[0].status.Floor = 2
	statusOverview[0].status.Direction = elevio.MD_Up
	statusOverview[0].status.CabRequests[0] = false
	statusOverview[0].status.CabRequests[1] = false
	statusOverview[0].status.CabRequests[2] = true
	statusOverview[0].status.CabRequests[3] = true

	statusOverview[1].id="two"
	statusOverview[1].status.State = elevfsm.Elevator_Waiting
	statusOverview[1].status.Floor = 0
	statusOverview[1].status.Direction = elevio.MD_Stop
	statusOverview[1].status.CabRequests[0] = false
	statusOverview[1].status.CabRequests[1] = false
	statusOverview[1].status.CabRequests[2] = false
	statusOverview[1].status.CabRequests[3] = false

	fmt.Print(reAssignHallRequests(test_hallRequests, statusOverview, elevatorId))
}

func stateString(peer ordersync.StatusMsg) string {
	peerState := ordersync.GetPeerState(peer)

	if peerState == elevfsm.Elevator_Waiting {
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
	for i := 1; i < 4; i++ {
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

func reAssignHallRequests(hallRequests hallRequestsType, statusOverview []ordersync.StatusMsg, elevatorId string) hallRequestsType {

	orderAssignerArgument = ""

	if len(hallRequests) < 1 || len(statusOverview) < 1 {
		return assignedHallRequests
	}

	//Insert accepted hall requests
	orderAssignerArgument = `{"hallRequests":[[`
	orderAssignerArgument += hallRequestString(hallRequests[0])
	orderAssignerArgument += `]`

	for i := 1; i < len(hallRequests); i++  {
		orderAssignerArgument += `,[`
		orderAssignerArgument += hallRequestString(hallRequests[i])
		orderAssignerArgument += `]`
	}

	//Insert elevator states
	orderAssignerArgument += `],"states":{"`+statusOverview[0].id+`":{"behaviour":"`
	orderAssignerArgument += stateString(statusOverview[0])
	orderAssignerArgument += `","floor":`+strconv.Itoa(ordersync.GetPeerFloor(statusOverview[0]))+`,"direction":"`
	orderAssignerArgument += directionString(statusOverview[0])
	orderAssignerArgument += `","cabRequests":[`
	orderAssignerArgument += cabRequestsString(statusOverview[0])
	orderAssignerArgument += `]}`

	for i := 1; i < len(statusOverview); i++ {
		orderAssignerArgument += `,"`+statusOverview[i].id+`":{"behaviour":"`
		orderAssignerArgument += stateString(statusOverview[i])
		orderAssignerArgument += `","floor":`+strconv.Itoa(ordersync.GetPeerFloor(statusOverview[i]))+`,"direction":"`
		orderAssignerArgument += directionString(statusOverview[i])
		orderAssignerArgument += `","cabRequests":[`
		orderAssignerArgument += cabRequestsString(statusOverview[i])
		orderAssignerArgument += `]}`
	}
	orderAssignerArgument += `}}`

	output, err := exec.Command(`./hall_request_assigner`,
		`--input`, orderAssignerArgument).CombinedOutput()
	if err != nil {
		fmt.Println(err.Error())
	}

	modifiedOutputString := strings.Split(string(output),elevatorId)[1]
	modifiedOutputString = modifiedOutputString[4:]
	for i := 0; i < 4; i++ {
		for j := 0; j < 2 ; j++ {
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

	return assignedHallRequests
}

*/