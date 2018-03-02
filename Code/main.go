package main

import (
  "fmt"
  "./elevio"
  "./fsm"
  "./requests"
  )

func main(){

    elevio.Init("localhost:15657")
    fsm.elevator.State = Undefined

    var d elevio.MotorDirection = elevio.MD_Up
    //elevio.SetMotorDirection(d)

    drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors  := make(chan typean int)
    drv_obstr   := make(chan bool)
    drv_stop    := make(chan bool)
    drv_timeout := make(chan bool) //??

    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)
    go elevio.PollObstructionSwitch(drv_obstr)
    go elevio.PollStopButton(drv_stop)

    <- drv_floors
    fsm.elevator.State = Idle

    for {
        select {
        case a := <- drv_buttons:
            fsm.onRequestButtonPress(a.Floor, a.Button)

            //fmt.Printf("%+v\n", a)

        case a := <- drv_floors:
            fsm.onFloorArrival(a)

            // fmt.Printf("%+v\n", a)
            // if a == numFloors-1 {
            //     d = elevio.MD_Down
            // } else if a == 0 {
            //     d = elevio.MD_Up
            // }
            // elevio.SetMotorDirection(d)


        case a := <- drv_obstr: //Is this a part of the assigment?
            fmt.Printf("%+v\n", a)
            if a {
                elevio.SetMotorDirection(elevio.MD_Stop)
            } else {
                fsm.elevator.Dir = requests.chooseDirection(fsm.elevator)
            }

        case a := <- drv_stop: //Is this a part of the assigment?
            fmt.Printf("%+v\n", a)
            for f := 0; f < numFloors; f++ {
                for b := elevio.ButtonType(0); b < 3; b++ {
                    elevio.SetButtonLamp(b, f, false)
                }
            }
        }
    }
}
