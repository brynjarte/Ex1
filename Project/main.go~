package main  

import (
	"Elevator"
)

func main() {
	
	someChannel := make(chan int,1)

	go Elevator.Elevator(2)

	<- someChannel
	
}

