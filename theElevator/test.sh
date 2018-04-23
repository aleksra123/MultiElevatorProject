gnome-terminal -- ./SimElevatorServer --port 23010
gnome-terminal -- ./SimElevatorServer --port 23011
gnome-terminal -- ./SimElevatorServer --port 23012


gnome-terminal -- go run main.go -id=1 23010
gnome-terminal -- go run main.go -id=2 23011
gnome-terminal -- go run main.go -id=3 23012
