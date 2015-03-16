
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

func master(SlaveResponseChannel chan SlaveMessage, MasterResponseChannel chan MasterMessage, terminate chan bool, externalOrderChannel chan ButtonMessage){
	go recieveUdpMessageSlave(SlaveResponseChannel,terminate)
	go recieveUdpMessageMaster(MasterResponseChannel,terminate)
	var elevator = Elevator{2,0,0}
	for {
                select{
			case message:= <-SlaveResponseChannel:
                                sendUdpMessageFromMaster(MasterMessage{message.ElevInfo.ElevatorID,false,message.Button})
                                fmt.Println("SENDER ACK")
                                if message.NewOrder{
                                        externalOrderChannel <- message.Button
                                } else{
                                	fmt.Println("Slave i'm alive)  
                        	}                
	                
                        case newOrder := <-externalOrderChannel:
                                // SEND PÅ EIN KANAL ELLER GJER SÅNN : elevatorID := CalculateCost(message.Button) // RETURNS kva heis som tar ordren
                                sendUdpMessageFromMaster(MasterMessage{elevator.ElevInfo.ElevatorID,true,newOrder})
                                //TURN ON LIGHT
				//SEND PÅ KANAL ELLER BRUKA FUNKSJONEN DIREKTE??
                               
                                for{
                            		if(elevator.ElevatorID == elevatorID){
                              			//ADD IN QUEUEUEU
                              			//SEND PÅ KANAL ELLER BRUKA FUNKSJONEN DIREKTE??             
            							break                                                    
                        			}
        							select {
	        							case reply := <-SlaveResponseChannel:
		        							if(reply.ElevInfo.ElevatorID == elevatorID){
                                           	  // Har fått ack av heisen som tar bestillingo.
                                           	  //ADD ORDER
                                       			break
                                        	} 
						       			case <-time.After(200*time.Millisecond):
                                    		// NO ACK.
								       	 	//removeElvatorFromCalculateCost(elevatorID)
                                        	//elevatorID := CalculateCost(message.Button) // RETURNS kva heis som tar ordren
                                        	sendUdpMessageFromMaster(MasterMessage{elevatorID,true,newOrder})
    
					        }
				        }
				        
			case masterMessage := <- MasterResponseChannel:
				if(masterMessage.ElevatorID < elevator.ElevatorID){
					terminate <- true
					terminate <- true
					go Slave(MasterResponseChannel,externalOrderChannel, terminate , SlaveResponseChannel)
					return
				} //else if(masterMessage.ElevatorID == elevatord.ElevatorID)
					//ERROR
				//}
					
                        case <-time.After(time.Second):
                                sendUdpMessageMaster(MasterMessage{2,false,ButtonMessage{0,0}})
                                 fmt.Println("SENDER I'm alive")
                                        
                }           
        }
}

func Slave(MasterResponseChannel chan MasterMessage, externalOrderChannel chan ButtonMessage, terminate chan bool, SlaveResponseChannel chan SlaveMessage){
 	var elevator = Elevator{2,0,0}
	var prevFloor int = 0
	go recieveUdpMessageMaster(MasterResponseChannel,terminate)
    	sensorChannel := make(chan int,1)
    	go driver.ReadSensors(sensorChannel)
	for{
		select{
			case currentFloor := <-sensorChannel:
				prevFloor = currentFloor
				movingDirection := prevFloor-currentFloor
				
			case reply:= <-MasterResponseChannel: 
				fmt.Println("i'm alive frå master")
				if(reply.NewOrder){
					if(reply.ElevatorID == elevator.ElevatorID){
			        		Queue.AddOrder(reply.Button,reply.elevatorID, currentFloor, movingDirection)
                                        } 
                                	driver.Elev_set_button_lamp(reply.Button, reply.Floor, 1) 
                                } 
			case newOrder := <-externalOrderChannel:
				for i := 0; i < 4; i++ {
					sendUdpMessageSlave(SlaveMessage{elevator, newOrder, true})
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
								go master(SlaveResponseChannel,MasterResponseChannel,terminate,externalOrderChannel)
								
								return
							}
					}
				}
					
			case <-time.After(3*time.Second):
                fmt.Println("STARTER MASTER")
				terminate <- true
				go master(SlaveResponseChannel,MasterResponseChannel,terminate,externalOrderChannel)
                                
				return
			}
	}
}



