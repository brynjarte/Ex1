package main  

import (
	//"driver"
	"UDP"
)

func main() {
	
	responseChannel := make(chan UDP.MasterMessage,1)
	externalOrderChannel := make(chan UDP.ButtonMessage,1)
	terminate := make(chan bool,1)
	SlaveResponseChannel := make(chan UDP.SlaveMessage,1)

	someChannel := make(chan int,1)
/*
	driver.Elev_init()
	go driver.ReadButtons()
	
	go func () {
		for {
		read := <- driver.ReadButtonsChannel
		println(read.Button)
		println(read.Floor)
		}	
	}()*/
	//p1 := UDP.Process{true,false,0}
	/*p2 := UDP.Process{false,true,0}
	
	go UDP.ProcessPair(p1,someChannel2)
	go UDP.ProcessPair(p2,someChannel2)	*/
	//go UDP.Slave(someChannel2)
	go UDP.Slave(responseChannel,externalOrderChannel,terminate,SlaveResponseChannel)
	<- someChannel
	


}

