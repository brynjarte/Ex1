package main  

import (
	//"driver"
	"UDP"
)

func main() {
	
	someChannel := make(chan int,1)
	someChannel2 := make(chan UDP.UDPMessage,1)
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
	p1 := UDP.Process{true,false,0}
	p2 := UDP.Process{false,true,0}
	
	go UDP.ProcessPair(p1,someChannel2)
	go UDP.ProcessPair(p2,someChannel2)	
	<- someChannel
	


}
