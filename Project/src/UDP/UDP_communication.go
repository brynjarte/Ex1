
package UDP


import(	
	"net"
	"encoding/json"
	"time"
	"fmt"
)

var boolvar bool

type UDPMessage struct{
	Message string
	MessageNumber int 
	
}



func RecieveUdpMessage(rec_channel chan UDPMessage){
	
	buffer := make([]byte,1024) 
	raddr,_ := net.ResolveUDPAddr("udp", ":26969")
	recievesock,_ := net.ListenUDP("udp", raddr)
	var rec_msg UDPMessage
	for {
		mlen , _,_ := recievesock.ReadFromUDP(buffer)
		json.Unmarshal(buffer[:mlen], &rec_msg)
		rec_channel <- rec_msg 
		//fmt.Println(rec_msg.Message, rec_msg.MessageNumber)
	}
}



func sendUdpMessage(msg UDPMessage){
	
	baddr,err := net.ResolveUDPAddr("udp", "192.168.1.69:26969")
	sendSock, err := net.DialUDP("udp", nil ,baddr) // connection
	//time.Sleep(200*time.Millisecond)
	buf,_ := json.Marshal(msg)
	_,err = sendSock.Write(buf)
	if err != nil{
		panic(err)
	}
	
}


func master(){
	
	msg := UDPMessage{"I'm alive",1}
	for {
		sendUdpMessage(msg)
		time.Sleep(1*time.Second)
	}
}

func Slave(rec_channel chan UDPMessage){
	go RecieveUdpMessage(rec_channel)
	for{
		select{
			case <-rec_channel: 
				fmt.Println("KONTAKT MED MASTER")	
			case <-time.After(3*time.Second):
				go master()
				fmt.Println("Startar ny master")
			}
	}
}





