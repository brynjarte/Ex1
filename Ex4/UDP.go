package main

import (
	"fmt"
	"time"
	"net"
	"encoding/json"
)

type UDPmessage struct {
	message string
	messageNumber int
}

var boolvar bool

//For å laga ein struct: var msg UDPmessage
//Berre la inn noke random i structen.


func receive_UDP(){
	
	boolvar = true

	addr, _ := net.ResolveUDPAddr("udp",":25555")
	sock, _ := net.ListenUDP("udp",addr)
	buffer := make([]byte, 1024)

	for (boolvar) {
		rlen,_,_ := sock.ReadFromUDP(buffer)
		var receiveMsg UDPmessage
		err := json.Unmarshal(buffer[:rlen], &receiveMsg)

		if err!=nil {
			println("error:", err)
		}
		
		println(receiveMsg.message)
	}
	
}

func send_UDP(){
	
	
	addr, _ := net.ResolveUDPAddr("udp","129.241.187.255:25555")
	con,_ := net.DialUDP("udp",nil,addr)
	sendMessage := UDPmessage{"Det er mottat!",1}
	time.Sleep(1*time.Second)
	buf,_ := json.Marshal(sendMessage)

	println(buf)


	var receiveMsg []UDPmessage
	json.Unmarshal(buf, &receiveMsg)
	fmt.Println("%+v",receiveMsg)






	time.Sleep(100*time.Millisecond)
	
	_,err := con.Write(buf)
	if err != nil {
		fmt.Println(err)
	}
	
}


func main (){
	
	go receive_UDP()
	send_UDP()
	time.Sleep(100*time.Second)
	boolvar = false
	

}
