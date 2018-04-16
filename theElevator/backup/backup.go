package backup

import (
  "io/ioutil"
  "encoding/json"
  "fmt"

  "../elevio"
)

func UpdateBackup(CabList elevio.Elevator){
	var backupList []elevio.ButtonEvent
  var temp elevio.ButtonEvent
		for floor:=0; floor < elevio.NumFloors; floor++ {
      if CabList.Requests[floor][elevio.BT_Cab] {
        temp.Floor = floor
        temp.Button = elevio.BT_Cab
				backupList = append(backupList,temp)
      }
		}
    b1, _ := json.Marshal(backupList)
    ioutil.WriteFile("Backup", b1, 0644)
}

func ReadBackup(CabList elevio.Elevator) elevio.Elevator{
	var backup []elevio.ButtonEvent
	c, err := ioutil.ReadFile("Backup")

	if err != nil {
		fmt.Println("No file exists yet, please continue")

	}else{

	json.Unmarshal(c, &backup)
	
	for _, order := range backup {
		CabList.Requests = AddCabOrder(CabList, order).Requests
		}
	}
  return CabList
}

func AddCabOrder(CabList elevio.Elevator, order elevio.ButtonEvent) elevio.Elevator{

  CabList.Requests[order.Floor][elevio.BT_Cab] = true
  return CabList
}
