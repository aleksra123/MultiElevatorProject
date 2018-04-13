package elevfsm

import (
	"elevio"
	"fmt"
	"elevorders"
	"time"
	"configFile"
)

var 	initialized		bool	= false

type ElevatorState int
const (
	Elevator_Initializing	ElevatorState 	= 0
	Elevator_Waiting						= 1
	Elevator_Driving						= 2
	Elevator_DoorOpen						= 3
)

type ElevatorStatus struct {
	State				ElevatorState
	Floor				int
	Direction			elevio.MotorDirection
}

var elevatorStatus ElevatorStatus

func Init(inputElevAddr string, chNewOrder chan<- elevio.ButtonEvent) {
	if initialized == false {
		elevio.Init(inputElevAddr, configFile.NUM_FLOORS)
		initialized = true
	}
}

func GetStatus() ElevatorStatus {
	return elevatorStatus
}

func stateString() string {

	if elevatorStatus.State == Elevator_Waiting {
		return `idle`
	} else if elevatorStatus.State == Elevator_Driving {
		return `moving`
	} else if elevatorStatus.State == Elevator_DoorOpen {
		return `doorOpen`
	}

	return ""
}

func PrintFSMstate()  {
	var currOrderArray elevorders.OrderArrayType
	oddNum := false
	for {
		select {
		case <- time.After(time.Second):
			currOrderArray = elevorders.GetOrderArray()
			fmt.Print("\rState: ",stateString(),", Floor: ", elevatorStatus.Floor, " Order Array: ", currOrderArray)
			if oddNum {
				fmt.Print(" +")
				oddNum = false
			} else {
				fmt.Print(" -")
				oddNum = true
			}
		}
	}
}

func StateMachine(chNewOrderElev chan<- elevio.ButtonEvent, chNewOrderSync chan<- elevio.ButtonEvent, chOrderServed chan<- int) {

	if initialized == false {
		fmt.Print("\nERROR! State machine running without initialization.\n")
		return
	}

	chButtons := make(chan elevio.ButtonEvent)
	chFloors  := make(chan int)
	chObstr   := make(chan bool)
	chStop    := make(chan bool)

	go elevio.PollButtons(chButtons)
	go elevio.PollFloorSensor(chFloors)
	go elevio.PollObstructionSwitch(chObstr)
	go elevio.PollStopButton(chStop)


	elevatorStatus = ElevatorStatus{Elevator_Initializing, -1, elevio.MD_Down}
	elevio.SetMotorDirection(elevatorStatus.Direction)

	timerRunning := false

	for {
		/*
		if elevatorStatus.State != Elevator_Initializing {
			select {

			case inputButtonEvent := <-chButtons:
				// Pass buttonEvent to module handling orders

				chNewOrderElev <- inputButtonEvent
				chNewOrderSync <- inputButtonEvent
				fmt.Println("\n**** BUTTON EVENT in MAIN FOR LOOP\n")

			default:
				break
				//case inputObstr := <- chObstr:
				// Check specifications for what actions should be made
			}
		} */

		switch elevatorStatus.State {

		case Elevator_Initializing:
			select {

			case inputFloor := <-chFloors:

				elevio.SetMotorDirection(elevio.MD_Stop)
				elevatorStatus.Direction = elevio.MD_Stop

				elevio.SetFloorIndicator(inputFloor)
				elevatorStatus.Floor = inputFloor

				elevatorStatus.State = Elevator_Waiting
				fmt.Println("\n")

			//case inputObstr := <-chObstr:
			//	fmt.Printf("Obstruction, state = INITIALIZING \n")

			}

		case Elevator_Waiting:

			//fmt.Println("In ELEVATOR WAITING")


			select {

			case inputButtonEvent := <-chButtons:
				// Pass buttonEvent to module handling orders

				chNewOrderElev <- inputButtonEvent
				chNewOrderSync <- inputButtonEvent
				//fmt.Println("\n-> BUTTON EVENT in FSM when IDLE\n")

			default:
				//Check if the doors should open
				orderAtFloor := elevorders.OrderAtFloor(elevatorStatus.Floor)
				//fmt.Println("Order at floor ", elevatorStatus.Floor, ": ", orderAtFloor, "\n")

				//Check if elevator should drive
				nextElevDir := elevorders.GetDirection(elevatorStatus.Floor, int(elevatorStatus.Direction))
				//fmt.Println("\n@@@@@ Checking next direction. Result: ", nextElevDir, " \n")

				if orderAtFloor {
					elevatorStatus.State = Elevator_DoorOpen
					fmt.Println("\n Transitioning from IDLE to DOOR_OPEN \n")
				} else if nextElevDir != configFile.MOTOR_DIR_STOP {
					elevio.SetMotorDirection(elevio.MotorDirection(nextElevDir))
					elevatorStatus.Direction = elevio.MotorDirection(nextElevDir)

					elevatorStatus.State = Elevator_Driving
					fmt.Println("\n Transitioning from IDLE to DRIVING \n")
				}


				//case inputObstr := <- chObstr:
				// Check specifications for what actions should be made
			}




		case Elevator_Driving:

			select {
			case inputFloor := <-chFloors:
				elevio.SetFloorIndicator(inputFloor)
				elevatorStatus.Floor = inputFloor

				shouldStop := elevorders.CheckStop(elevatorStatus.Direction, inputFloor)

				fmt.Println("\nElevator driving, floor sensor input: ",inputFloor, " Should stop: ", shouldStop, "\n")

				if shouldStop {
					elevatorStatus.Direction = elevio.MD_Stop
					elevio.SetMotorDirection(elevio.MD_Stop)

					elevatorStatus.State = Elevator_Waiting
					fmt.Println("\n")
					break
				}

			case inputButtonEvent := <-chButtons:
				// Pass buttonEvent to module handling orders
				chNewOrderElev <- inputButtonEvent
				chNewOrderSync <- inputButtonEvent

			}


		case Elevator_DoorOpen:

			select {
			case inputButtonEvent := <-chButtons:
				// Pass buttonEvent to module handling orders
				chNewOrderElev <- inputButtonEvent
				chNewOrderSync <- inputButtonEvent
			default:
				if !timerRunning {
					elevio.SetDoorOpenLamp(true)
					elevorders.ServeFloor(elevatorStatus.Floor, chOrderServed)
					go func() {
						//doorTimer := time.NewTimer(3*time.Second)
						timerRunning = true
						//<-doorTimer.C
						<-time.After(3*time.Second)
						timerRunning = false
						elevio.SetDoorOpenLamp(false)
						elevatorStatus.State = Elevator_Waiting
					}()
				}
			}

		}
	}

}
