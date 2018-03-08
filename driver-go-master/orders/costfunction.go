import ".../elevio"

func duplicateOrder(order elevio.Keypress, elevList [NumElevators]elevio.Elev, id int) bool {
	return (AcceptedOrdersMatrix[order.Floor][order.Btn]) //returns true if order already exists
}

func costCalculator(order Keypress, elevList [NumElevators]Elev, id int, onlineList [NumElevators]bool) int {
	if order.Btn == BtnInside {
		return id
	}
	minCost := (NumButtons * NumFloors) * NumElevators
	bestElevator := id
	for elevator := 0; elevator < NumElevators; elevator++ {
		if !onlineList[elevator] {
			// Disregarding offline elevators
			continue
		}
		cost := order.Floor - elevList[elevator].Floor

		if cost == 0 && elevList[elevator].State != Moving {
			bestElevator = elevator
			return bestElevator
		}

		if cost < 0 {
			cost = -cost
			if elevList[elevator].Dir == DirUp {
				cost += 3
			}
		}
		else if cost > 0 {
			if elevList[elevator].Dir == DirDown {
				cost += 3
			}
		}
		if cost == 0 && elevList[elevator].State == Moving {
			cost += 4
		}

		if elevList[elevator].State == DoorOpen {
			cost++
		}

		if cost < minCost {
			minCost = cost
			bestElevator = elevator
		}
	}
	return bestElevator
}
