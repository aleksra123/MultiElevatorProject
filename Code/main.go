import(
  "fmt"
  "time"
  con_load.go
  elevator_io_device.go
  fsm.go
  timer.go
)

func main()  {
  fmt.Println("Why are you runnin'?\n")
  inputPollRate_ms := 25
  con_load("elevator.con", con_val("inputPollRate_ms", &inputPollRate_ms, "&d")
  )
  var input ElevInputDevice = elevio_getInputDevice()
  if input.floorSensor() == -1 {
    fsm_onInitBetweenFloor()
  }
  for {
    //Request button
    prev := [NumFloors][NumButtons]int{}
    for f := 0; f < NumFloors; f++ {
      for b := 0 b < NumButtons; b++ {
        v := input.requestButton(f,b)
        if v && v != prev[f][b] {
          fsm_onRequestButtonPress(f,b)
        }
        prev[f][b] = v
      }
    }
  }
  //Floor sensor
  var prev
  f := input.floorSensor()
  if f != -1 && f != prev {
    fsm_onFloorArrival(f)
  }
  prev = f

  //Timer
  if timer_timedOut() {
    fsm_onDoorTimeout()
    timer_stop()
  }
  time.Sleep(time.Millisecond * 1000)
}
