package configFile

// CONSTANTS
const (
	NUM_FLOORS      int = 4
	NUM_DIR         int = 2
	NUM_ORDER_TYPES int = 3
	MOTOR_DIR_UP    int = 1
	MOTOR_DIR_DOWN  int = -1
	MOTOR_DIR_STOP  int = 0
	REQUEST_UP      int = 0
	REQUEST_DOWN    int = 1
	REQUEST_CAB     int = 2
)

// DATATYPES
type HallRequestType [4][2]bool
type CabRequestType [4]bool
