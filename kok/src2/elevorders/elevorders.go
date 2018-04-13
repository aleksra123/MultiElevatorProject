package elevorders

import (
	"elevio"
	"configFile"
	"fmt"
	"time"
)

type OrderArrayType [configFile.NUM_FLOORS][configFile.NUM_ORDER_TYPES]bool
var orderArray OrderArrayType


func SetCabOrders(chNewOrders <-chan elevio.ButtonEvent) {
	for {
		select {
		case newOrder := <-chNewOrders:
			fmt.Println("\n--> Order received in ELEVORDERS\n")
			if newOrder.Button == elevio.BT_Cab {
				fmt.Println("\n---> Setting Cab order in ELEVORDERS\n")
				orderArray[newOrder.Floor][elevio.BT_Cab] = true
				/* TODO: sett lys */
				elevio.SetButtonLamp(elevio.BT_Cab, newOrder.Floor, true)
			}
		}
	}
}

func GetCabOrders() configFile.CabRequestType {
	var currCabOrders configFile.CabRequestType
	for floorIterator := 0; floorIterator < configFile.NUM_FLOORS; floorIterator++ {
		currCabOrders[floorIterator] = orderArray[floorIterator][elevio.BT_Cab]
	}
	return currCabOrders
}

func deleteOrder(inputFloor int, inputOrderType elevio.ButtonType)  {
	orderArray[inputFloor][inputOrderType] = false
	/* TODO: fjern lys */
	elevio.SetButtonLamp(inputOrderType, inputFloor, false)
}

func GetOrderArray() OrderArrayType {
	return orderArray
}

func clearOrderArray() {
	for i := 0; i < configFile.NUM_FLOORS; i++ {
		for j := 0; j < configFile.NUM_ORDER_TYPES; j++ {
			orderArray[i][j] = false
		}
	}
}

func InitElevatorOrders(chNewOrders <-chan elevio.ButtonEvent, chAssignedHallOrders <-chan configFile.HallRequestType) {
	clearOrderArray()
	//Get orders from file? <- If program restarts after crash etc.
}

func RunElevatorOrders(chNewOrderElev <-chan elevio.ButtonEvent, chAssignedHallOrders <-chan configFile.HallRequestType) {

	go SetCabOrders(chNewOrderElev)
	go UpdateHallOrders(chAssignedHallOrders)

	for{
		select{
		case <- time.After(5*time.Second):
		}
	}
}

func ServeFloor(inputFloor int, chOrderServed chan<- int){

	for i := 0; i < configFile.NUM_ORDER_TYPES; i++ {
			orderArray[inputFloor][i] = false
			elevio.SetButtonLamp(elevio.ButtonType(i), inputFloor, false)
	}
	chOrderServed <- inputFloor

}


func CheckStop(inputMotorDir elevio.MotorDirection, inputFloor int) bool {
	//Should stop if: order at floor, no orders further on in same direction, or at end of elevator shaft.

	inputElevDir := 0 //Case: elevator is standing still
	if inputMotorDir == elevio.MD_Down {
		inputElevDir = -1
	} else if inputMotorDir == elevio.MD_Up {
		inputElevDir = 1
	}

	//Case: reached end of elevator shaft
	if inputFloor == 0 || inputFloor == configFile.NUM_FLOORS-1 {
		return true
	}

	for i := 0; i < configFile.NUM_ORDER_TYPES; i++ {
		if orderArray[inputFloor][i] == true {
			return true
		}
	}

	for nextFloor := inputFloor + inputElevDir; nextFloor >= 0 && nextFloor < configFile.NUM_FLOORS; nextFloor += inputElevDir {
		for i := 0; i < configFile.NUM_ORDER_TYPES; i++ {
			if orderArray[nextFloor][i] == true {
				return false
			}
		}
	}

	return true
}

func OrderAtFloor(inputFloor int) bool {
	returnVal := false
	for i := 0; i < configFile.NUM_ORDER_TYPES; i++ {
		if orderArray[inputFloor][i] == true {
			returnVal = true
		}
	}
	//fmt.Println("RETURN VALUE FROM ORDERATFLOOR: ", returnVal)
	return returnVal
}

func GetDirection(inputFloor int, inputDirection int) int {

	currDirection := inputDirection

	//If elevator is waiting, use dir up as default
	if currDirection == configFile.MOTOR_DIR_STOP {
		currDirection = configFile.MOTOR_DIR_UP
	}

	//Check orders in current direction
	for nextFloor := inputFloor + currDirection; nextFloor >= 0 && nextFloor < configFile.NUM_FLOORS; nextFloor += currDirection {
		for i := 0; i < configFile.NUM_ORDER_TYPES; i++ {
			if orderArray[nextFloor][i] == true {
				fmt.Println("\n---------> In elevorders.GetDirection, returning: ", currDirection, "\n")
				return currDirection
			}
		}
	}

	var oppDirection int
	if currDirection == configFile.MOTOR_DIR_UP {
		oppDirection = configFile.MOTOR_DIR_DOWN
	} else {
		oppDirection = configFile.MOTOR_DIR_UP
	}

	//Check orders in opposite direction
	for nextFloor := inputFloor + oppDirection; nextFloor >= 0 && nextFloor < configFile.NUM_FLOORS; nextFloor += oppDirection {
		for i := 0; i < configFile.NUM_ORDER_TYPES; i++ {
			if orderArray[nextFloor][i] == true {
				fmt.Println("\n---------> In elevorders.GetDirection, returning: ", oppDirection, "\n")
				return oppDirection
			}
		}
	}
	//fmt.Println("\n---------> In elevorders.GetDirection, returning: ", configFile.MOTOR_DIR_STOP, "\n")
	return configFile.MOTOR_DIR_STOP
}


func UpdateHallOrders(chAssignedHallOrders <-chan configFile.HallRequestType)  {
	for {
		select {
		case assignedHallOrders := <- chAssignedHallOrders:
			//Copy assigned hall orders
			for floorIter := 0; floorIter < configFile.NUM_FLOORS; floorIter++ {
				for dirIter := 0; dirIter < 2; dirIter++ {
					orderArray[floorIter][dirIter] = assignedHallOrders[floorIter][dirIter]
				}
			}

		default:
			break
		}
	}
}
