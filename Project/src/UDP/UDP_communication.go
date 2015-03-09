
package UDP


import(	
	"net"
	"encoding/json"
	"time"
        "fmt"
)

var boolvar bool

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



func sendUdpMessageSlave(msg SlaveMessage){

	baddr,err := net.ResolveUDPAddr("udp", "129.241.187.255:26969")
	sendSock, err := net.DialUDP("udp", nil ,baddr) // connection
	//time.Sleep(200*time.Millisecond)
	buf,_ := json.Marshal(msg)
	_,err = sendSock.Write(buf)
	if err != nil{
		panic(err)
	}
	
}

func sendUdpMessageMaster(msg MasterMessage){
   
	baddr,err := net.ResolveUDPAddr("udp", "129.241.187.255:29696")
	sendSock, err := net.DialUDP("udp", nil ,baddr) // connection
	//time.Sleep(200*time.Millisecond)
	buf,_ := json.Marshal(msg)
	_,err = sendSock.Write(buf)
	if err != nil{
		panic(err)
	}
	
}


func master(SlaveResponseChannel chan SlaveMessage, terminate chan bool, externalOrderChannel chan ButtonMessage){
	go RecieveUdpMessageSlave(SlaveResponseChannel,terminate)
	var elevator = Elevator{2,0,0}
	for {
                select{
			case message:= <-SlaveResponseChannel:
                                sendUdpMessageMaster(MasterMessage{message.ElevInfo.ElevatorID,false,message.Button})
                                fmt.Println("SENDER ACK")
                                if message.NewOrder{
                                        externalOrderChannel <- message.Button
                                }                  
	                
                        case newOrder := <-externalOrderChannel:
                                //elevatorID := CalculateCost(message.Button) // RETURNS kva heis som tar ordren
                                var elevatorID int = 1
                                
                                        for{
                                                if(elevator.ElevatorID == elevatorID){
                                                        //ADD IN QUEUEUEU
                                                        //TURN ON LIGHT                                                    
                                                        break
                                                }
                                        
                                                sendUdpMessageMaster(MasterMessage{elevatorID,true,newOrder})
					        select {
						        case reply := <-SlaveResponseChannel:
							        if(reply.ElevInfo.ElevatorID == elevator.ElevatorID){
                                                                         // Har fått ack.
                                                                       break
                                                                } 
						        case <-time.After(200*time.Millisecond):
                                                                        // NO ACK.
								        //removeElvatorFromCalculateCost(elevatorID)
                                                                        //elevatorID := CalculateCost(message.Button) // RETURNS kva heis som tar ordren
								        
							        
					        }
				        }
                        case <-time.After(time.Second):
                                sendUdpMessageMaster(MasterMessage{2,false,ButtonMessage{0,0}})
                                 fmt.Println("SENDER I'm alive")
                                        
                }           
        }
}

func Slave(MasterResponseChannel chan MasterMessage, externalOrderChannel chan ButtonMessage, terminate chan bool, SlaveResponseChannel chan SlaveMessage){
	var elevator = Elevator{2,0,0}
        go RecieveUdpMessageMaster(MasterResponseChannel,terminate)
	for{
		select{
			case /*reply:=*/ <-MasterResponseChannel: 
				fmt.Println("i'm alive frå master")
				
				/*f(reply.ElevatorID == elevator.ElevatorID){
					fmt.Println("i'm alive frå master")
                                        if(reply.NewOrder){
				        	//ADD IN QUEUEUEU
                                                //TURN ON LIGHT
                                        }
                                } */
			case newOrder := <-externalOrderChannel:
				for i := 0; i < 4; i++ {
					sendUdpMessageSlave(SlaveMessage{elevator, newOrder, true})
					select {
						case reply := <-MasterResponseChannel:
							if(reply.ElevatorID == elevator.ElevatorID){
                                                                i = 4
								if(reply.NewOrder){
									//ADD IN QUEUEUEU
                                                                        //TURN ON LIGHT
                                                                }
                                                        } 
						case <-time.After(200*time.Millisecond):
							if (i < 3) {						
								break
							} else {
								terminate <- true
								go master(SlaveResponseChannel,terminate,externalOrderChannel)
								
								return
							}
					}
				}
					
			case <-time.After(3*time.Second):
                                fmt.Println("STARTER MASTER")
				terminate <- true
				go master(SlaveResponseChannel,terminate,externalOrderChannel)
                                
				return
			}
	}
}





