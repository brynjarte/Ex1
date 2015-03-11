package Elevator

import(
	"UDP"
	"driver"
	"Queue"	
	"time"
   	"fmt"
)

func master(SlaveResponseChannel chan SlaveMessage, MasterResponseChannel chan MasterMessage, terminate chan bool, externalOrderChannel chan ButtonMessage){
	go recieveUdpMessageSlave(SlaveResponseChannel,terminate)
	go recieveUdpMessageMaster(MasterResponseChannel,terminate)
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
                                sendUdpMessageMaster(MasterMessage{elevatorID,true,newOrder})
                                //TURN ON LIGHT
                               
                                for{
                            		if(elevator.ElevatorID == elevatorID){
                              			//ADD IN QUEUEUEU             
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
                                        	sendUdpMessageMaster(MasterMessage{elevatorID,true,newOrder})
    
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



