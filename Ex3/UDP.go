package main

import (
	"fmt"
	"time"
	"net"

)

type UDPmessage {
	message string
	messageNumber int
}

//For å laga ein struct: var msg UDPmessage
// Berre la inn noke random i structen.


func receive_UDP() (string){

	addr, _ := net.ResolveUDPAddr("udp",":30000")
	sock, _ := net.ListenUDP("udp",addr)
	buffer := make([]byte, 1024)
	rlen,_,_ := sock.ReadFromUDP(buffer)
	
	return string(buffer[0:rlen])
	
}

func send_UDP(){
	
	
	addr, _ := net.ResolveUDPAddr("udp","129.241.187.255:20005")
	con,_ := net.DialUDP("udp",nil,addr)
	buf := []byte("JAJA")
	time.Sleep(100*time.Millisecond)
	_,err := con.Write(buf)

	if err != nil {
		fmt.Println(err)
	}
	
}


func main (){
	
	send_UDP()
	fmt.Println(receive_UDP())


}
