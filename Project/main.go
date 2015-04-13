package main  

import (
	//"driver"
	"Elevator"
)

func main() {
	
//	responseChannel := make(chan UDP.MasterMessage,1)
//	externalOrderChannel := make(chan UDP.ButtonMessage,1)
//	terminate := make(chan bool,1)
//	SlaveResponseChannel := make(chan UDP.SlaveMessage,1)

	someChannel := make(chan int,1)
	
	//var queue [4][2] bool

	go Elevator.Elevator()

	<- someChannel
	


}

