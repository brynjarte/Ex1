
package UDP


import(	
	"net"
	"encoding/json"
	"time"
    "fmt"
)

type SlaveMessage struct{
	ElevInfo Elevator
	Button ButtonMessage
	NewOrder bool

}

type ButtonMessage struct {
	Floor int
	Button int
	Light int
}

type Elevator struct {
	ElevatorID int	
	CurrentFloor int
	Direction int
	//numberInQueue int
}

type MasterMessage struct {
	ElevatorID int
	Ack bool
	NewOrder bool
	Button ButtonMessage

}


func recieveUdpMessageSlave(responseChannel chan SlaveMessage, terminate chan bool){
	
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

func recieveUdpMessageMaster(responseChannel chan MasterMessage,terminate chan bool){

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



func sendUdpMessageFromSlave(msg SlaveMessage){

	baddr,err := net.ResolveUDPAddr("udp", "129.241.187.255:26969")
	sendSock, err := net.DialUDP("udp", nil ,baddr) // connection
	//time.Sleep(200*time.Millisecond)
	buf,_ := json.Marshal(msg)
	_,err = sendSock.Write(buf)
	if err != nil{
		panic(err)
	}
	
}

func sendUdpMessageFromMaster(msg MasterMessage){
   
	baddr,err := net.ResolveUDPAddr("udp", "129.241.187.255:29696")
	sendSock, err := net.DialUDP("udp", nil ,baddr) // connection
	//time.Sleep(200*time.Millisecond)
	buf,_ := json.Marshal(msg)
	_,err = sendSock.Write(buf)
	if err != nil{
		panic(err)
	}
	
}

func master(SlaveResponseChannel chan SlaveMessage, MasterResponseChannel chan MasterMessage, terminate chan bool, externalOrderChannel chan ButtonMessage,AddToQueueChannel chan ButtonMessage,RemoveFromQueueChannel chan ButtonMessage,SetLightChannel chan ButtonMessage, CalculateCostChannel chan MasterMessage, elevatorInfoChannel chan Elevator){
	go recieveUdpMessageSlave(SlaveResponseChannel,terminate)
	go recieveUdpMessageMaster(MasterResponseChannel,terminate)
	var elevator = Elevator{2,0,0}
	for {
        select{
			case message:= <-SlaveResponseChannel:
		        sendUdpMessageFromMaster(MasterMessage{message.ElevInfo.ElevatorID,true,false,message.Button})
		        fmt.Println("SENDER ACK")
		        if message.NewOrder{
		                externalOrderChannel <- message.Button
		        } else{
		        	elevatorInfoChannel <- message.ElevInfo
				}                
	                
		    case newOrder := <-externalOrderChannel:
	        	// SEND PÅ EIN KANAL ELLER GJER SÅNN : elevatorID := CalculateCost(message.Button) // RETURNS kva heis som tar ordren
		        CalculateCostChannel <- MasterMessage{-1,false,true,newOrder}
		        msg := <-CalculateCostChannel
		        sendUdpMessageFromMaster(msg)
 
				for{
					if(elevator.ElevatorID == elevatorID){
						AddToQueueChannel <- msg.Button // HUSK Å SKU PÅ LYS            
						break                                                    
					}
					select {
						case reply := <-SlaveResponseChannel:
							if(reply.ElevInfo.ElevatorID == elevatorID){
								// Har fått ack av heisen som tar bestillingo.
								//ADD ORDER in masterQUEUE
								SetLightChannel <- reply.Button
								break
													} 
						case <-time.After(200*time.Millisecond):
							// NO ACK.
							//removeElvatorFromCalculateCost(elevatorID)
							//elevatorID := CalculateCost(message.Button) // RETURNS kva heis som tar ordren
							sendUdpMessageFromMaster(MasterMessage{elevatorID,false,true,newOrder})
					}
				}
				        
			case masterMessage := <- MasterResponseChannel:
				if(masterMessage.ElevatorID < elevator.ElevatorID){
					terminate <- true
					terminate <- true
					go Slave(MasterResponseChannel,externalOrderChannel, terminate , SlaveResponseChannel,AddToQueueChannel,RemoveFromQueueChannel,SetLightChannel)
					return
				} //else if(masterMessage.ElevatorID == elevatord.ElevatorID)
					//ERROR
				//}
					
            case <-time.After(time.Second):
                sendUdpMessageFromMaster(MasterMessage{2,true,false,ButtonMessage{0,0,0}})
                fmt.Println("SENDER I'm alive")
                                        
                }           
        }
}

func Slave(MasterResponseChannel chan MasterMessage, externalOrderChannel chan ButtonMessage, terminate chan bool, SlaveResponseChannel chan SlaveMessage, AddToQueueChannel chan ButtonMessage, RemoveFromQueueChannel chan ButtonMessage,SetLightChannel chan ButtonMessage, elevatorInfoChannel chan Elevator){

 	CalculateCostChannel := make(chan ButtonMessage,1)// LAGAST I EVENTHANDLER
 	var elevator = Elevator{2,0,0}
	go recieveUdpMessageMaster(MasterResponseChannel,terminate)
    	
	for{
		select{				
			case message:= <-MasterResponseChannel:
				if message.Ack{
					break
				} else if message.newOrder{
					if message.ElevatorID == elevator.ElevatorID{
						AddToQueueChannel <- message.ButtonMessage // HUSK Å SETT PÅ LYS
						sendUdpMessageFromSlave(SlaveMessage{elevator, message.newOrder, false})
					} 
				} else{
					SetLightChannel <- message.ButtonMessage
				} 				
                                 
			case newOrder := <-externalOrderChannel:
				for i := 0; i < 4; i++ {
					sendUdpMessageFromSlave(SlaveMessage{elevator, newOrder, true})
					select {
						case reply := <-MasterResponseChannel:
							i = 4
							MasterResponseChannel <- reply 
							break 
						case <-time.After(200*time.Millisecond):
							if (i < 3) {						
								break
							} else {
								terminate <- true
								AddToQueueChannel <- newOrder // LEGGER TIL SJØLV
								go master(SlaveResponseChannel,MasterResponseChannel,terminate,externalOrderChannel,AddToQueueChannel,RemoveFromQueueChannel,SetLightChannel,CalculateCostChannel)
								
								return
							}
					}
				}
			case doneOrder := <-RemoveFromQueueChannel:
				for i := 0; i < 4; i++ {
					sendUdpMessageFromSlave(SlaveMessage{elevator, doneOrder, false})
					select {
						case reply := <-MasterResponseChannel:
							i = 4
							MasterResponseChannel <- reply 
							break 
						case <-time.After(200*time.Millisecond):
							if (i < 3) {						
								break
							} else {
								terminate <- true
								go master(SlaveResponseChannel,MasterResponseChannel,terminate,externalOrderChannel,AddToQueueChannel,RemoveFromQueueChannel,SetLightChannel,CalculateCostChannel)
								return
							}
					}
				}
				
			case elevatorInfo: <- ElevatorinfoChannel:
				sendUdpMessageFromSlave(SlaveMessage{elevatorInfo, ButtonMessage{0,0,0}, false})
		
			case <-time.After(3*time.Second):
                fmt.Println("STARTER MASTER")
				terminate <- true
				go master(SlaveResponseChannel,MasterResponseChannel,terminate,externalOrderChannel,AddToQueueChannel,RemoveFromQueueChannel,SetLightChannel)
                                
				return
			}
	}
}



