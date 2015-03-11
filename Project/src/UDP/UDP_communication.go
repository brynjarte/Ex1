
package UDP


import(	
	"net"
	"encoding/json"
	//"time"
        //"fmt"
)

type SlaveMessage struct{
	ElevInfo Elevator
	Button ButtonMessage
	NewOrder bool
}

type ButtonMessage struct {
	Floor int
	Button int
}

type Elevator struct {
	ElevatorID int	
	CurrentFloor int
	Direction int
	//numberInQueue int
}

type MasterMessage struct {
	ElevatorID int
	NewOrder bool
	Button ButtonMessage
}


func RecieveUdpMessageSlave(responseChannel chan SlaveMessage, terminate chan bool){
	
	buffer := make([]byte,1024) 
	raddr,_ := net.ResolveUDPAddr("udp", ":26969")
	recievesock,_ := net.ListenUDP("udp", raddr)
	var recMsg SlaveMessage
	for {
		select{
			case <- terminate:
				return
			default:
				mlen , _,_ := recievesock.ReadFromUDP(buffer)
				json.Unmarshal(buffer[:mlen], &recMsg)
				responseChannel <- recMsg 
		}
	}
}

func RecieveUdpMessageMaster(responseChannel chan MasterMessage,terminate chan bool){

	buffer := make([]byte,1024) 
	raddr,_ := net.ResolveUDPAddr("udp", ":29696")
	recievesock,_ := net.ListenUDP("udp", raddr)
	var recMsg MasterMessage
	for {
		select{
			case <- terminate:
				return
			default:
				mlen , _,_ := recievesock.ReadFromUDP(buffer)
				json.Unmarshal(buffer[:mlen], &recMsg)
				responseChannel <- recMsg 
		}
	}
}



func SendUdpMessageSlave(msg SlaveMessage){

	baddr,err := net.ResolveUDPAddr("udp", "129.241.187.255:26969")
	sendSock, err := net.DialUDP("udp", nil ,baddr) // connection
	//time.Sleep(200*time.Millisecond)
	buf,_ := json.Marshal(msg)
	_,err = sendSock.Write(buf)
	if err != nil{
		panic(err)
	}
	
}

func SendUdpMessageMaster(msg MasterMessage){
   
	baddr,err := net.ResolveUDPAddr("udp", "129.241.187.255:29696")
	sendSock, err := net.DialUDP("udp", nil ,baddr) // connection
	//time.Sleep(200*time.Millisecond)
	buf,_ := json.Marshal(msg)
	_,err = sendSock.Write(buf)
	if err != nil{
		panic(err)
	}
	
}





