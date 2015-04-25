package main  

import (
	"Elevator"
)

func main() {
	
	runningElevator := make(chan int,1)

	go Elevator.RunElevator()

	<- runningElevator
}

