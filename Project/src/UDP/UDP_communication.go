
package UDP


import(
	"time"	
	"net"
	"encoding/json"
)

var boolvar bool

type UDPMessage struct{
	Message string
	MessageNumber int 
	
}


func recieveUdpMessage(rec_channel chan UDPMessage){
	
	buffer := make([]byte,1024) 
	raddr,_ := net.ResolveUDPAddr("udp", ":25555")
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
	
	baddr,err := net.ResolveUDPAddr("udp", "129.241.187.255:25555")
	sendSock, err := net.DialUDP("udp", nil ,baddr) // connection
	//time.Sleep(200*time.Millisecond)
	buf,_ := json.Marshal(msg)
	_,err = sendSock.Write(buf)
	if err != nil{
		panic(err)
	}
	
}


