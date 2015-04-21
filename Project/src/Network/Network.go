package Network

import(	
	"net"
	"encoding/json"
	"time"
	"Source"
	//"fmt"
)


func recieveUdpMessage(master bool, responseChannel chan Source.Message, terminate chan bool, terminated chan int){
	
	buffer := make([]byte,1024)
	raddr,_ := net.ResolveUDPAddr("udp", ":26969")
	if(master){
		raddr,_ = net.ResolveUDPAddr("udp", ":27000")
	}
	
	recievesock,_ := net.ListenUDP("udp", raddr)
	//recievesock.SetReadDeadline(time.Now().Add(time.Second)) SPØR OM HJELP.
	var recMsg Source.Message
	for {
		select{
			case <- terminate:
				println("STOPPER RECEIVEEEE from master")
				recievesock.Close()
				terminated <- 1
				return
			default:
				mlen , _,_:= recievesock.ReadFromUDP(buffer)
				//if(err == nil){
					json.Unmarshal(buffer[:mlen], &recMsg)
					responseChannel <- recMsg
				//}
				
		}
	
	}
}

func sendUdpMessage(msg Source.Message){
	baddr,err := net.ResolveUDPAddr("udp", "129.241.187.255:27000")

	if(msg.FromMaster){
		baddr,err = net.ResolveUDPAddr("udp", "129.241.187.255:26969")
	}

	sendSock, err := net.DialUDP("udp", nil ,baddr) // connection
	buf,_ := json.Marshal(msg)
	_,err = sendSock.Write(buf)
	if err != nil{
		panic(err)
	}
	
}

func Slave(elevator Source.ElevatorInfo, externalOrderChannel chan Source.ButtonMessage, updateElevatorInfoChannel chan Source.ElevatorInfo, handleOrderChannel chan Source.Message, removeExternalOrderChannel chan Source.Message, bestElevatorChannel chan Source.Message, removeElevator chan int, completedOrderChannel chan Source.ButtonMessage){
	
	println("Starter SALVE")
	msgFromMasterChannel := make(chan Source.Message,1)
	terminate := make(chan bool, 1)
	terminated := make(chan int, 1)
	go recieveUdpMessage(false, msgFromMasterChannel, terminate, terminated)
    	
	for{
		select{
	
			case messageFromMaster := <- msgFromMasterChannel:
				println("Message <- respChan")
				handleOrderChannel  <- messageFromMaster
				if (messageFromMaster.NewOrder && messageFromMaster.MessageTo == elevator.ID) {
					sendUdpMessage(Source.Message{false, true, false, false, false, elevator.ID, -1, elevator, messageFromMaster.Button})
				}

			case newOrder := <-externalOrderChannel:
				println("newOrder <- extOrdChan")
				for i := 0; i < 4; i++ {
					sendUdpMessage(Source.Message{true, false, false, false, false, elevator.ID, -1, elevator, newOrder})
					select {
						case reply := <-msgFromMasterChannel:
							println("Feil ID")
							i = 4
							msgFromMasterChannel <- reply 
							break
						
						case <-time.After(200*time.Millisecond):
							println("time.After()")
							if (i < 3) {						
								break
							} else {
								msg := Source.Message{true, false, false, false, false, elevator.ID, 2, elevator, newOrder}
								handleOrderChannel <- msg
								terminate <- true
								sendUdpMessage(Source.Message{false, false, true, false, true, elevator.ID, -1, elevator, Source.ButtonMessage{0,0,0}}) // DUMMY MESSAGE
								<- terminated
								go master( elevator, externalOrderChannel, updateElevatorInfoChannel, handleOrderChannel, removeExternalOrderChannel, bestElevatorChannel, removeElevator, completedOrderChannel)
								return
							}
						
					}
				}

			case completedOrder := <-completedOrderChannel:
				println("Sending Completed Order")
				for i := 0; i < 4; i++ {
					sendUdpMessage(Source.Message{false, false, false, true, false, elevator.ID, -1, elevator, completedOrder}) // FJERNER ORDREN EIN ANNA PLASS?
					select {
						case reply := <-msgFromMasterChannel:
							i = 4
							msgFromMasterChannel <- reply 
							break 
						case <-time.After(200*time.Millisecond):
							if (i < 3) {						
								break
							} else {
								go master( elevator, externalOrderChannel, updateElevatorInfoChannel, handleOrderChannel, removeExternalOrderChannel, bestElevatorChannel, removeElevator, completedOrderChannel)
								return
							}
					}
				}
				
			//Skifter retn. el. etg.: 
			case elevatorInfo :=  <- updateElevatorInfoChannel:
				println("Sending elev info from salve")
				sendUdpMessage(Source.Message{false, false, false, false, true, elevator.ID, -1, elevatorInfo, Source.ButtonMessage{0,0,0}})
				handleOrderChannel <- Source.Message{false, false, false, false, true, elevator.ID, -1, elevatorInfo, Source.ButtonMessage{0,0,0}}
				
		}
	}
}


func master( elevator Source.ElevatorInfo, externalOrderChannel chan Source.ButtonMessage, updateElevatorInfoChannel chan Source.ElevatorInfo, handleOrderChannel chan Source.Message, removeExternalOrderChannel chan Source.Message, bestElevatorChannel chan Source.Message, removeElevator chan int, completedOrderChannel chan Source.ButtonMessage){
	println("Starter MASTAH")

	messageFromSlaveChannel := make(chan Source.Message, 1)
	messageFromMasterChannel := make(chan Source.Message, 1)

	terminateFromMaster := make(chan bool, 1)
	terminatedFromMaster := make(chan int, 1)
	terminateFromSlave := make(chan bool, 1)
	terminatedFromSlave := make(chan int, 1)
	go recieveUdpMessage(false, messageFromMasterChannel, terminateFromMaster, terminatedFromMaster)
	go recieveUdpMessage(true, messageFromSlaveChannel, terminateFromSlave, terminatedFromSlave)

	for {
		select{

			case messageFromMaster := <- messageFromMasterChannel:
				println("TO MASTERAA", messageFromMaster.MessageFrom)
				if(messageFromMaster.MessageFrom < elevator.ID){
						terminateFromSlave <- true
						terminateFromMaster <- true
						sendUdpMessage(Source.Message{false, false, true, false, true, -1, -1, Source.ElevatorInfo{-1, -1, -1}, Source.ButtonMessage{0,0,0}})
						sendUdpMessage(Source.Message{false, false, false, false, true, -1, -1, Source.ElevatorInfo{-1, -1, -1}, Source.ButtonMessage{0,0,0}})
						<- terminatedFromSlave
						<- terminatedFromMaster
						go Slave( elevator, externalOrderChannel, updateElevatorInfoChannel, handleOrderChannel, removeExternalOrderChannel, bestElevatorChannel, removeElevator, completedOrderChannel)	
						return	
				}
			case messageFromSlave := <- messageFromSlaveChannel:
				println("MESSAGE FROM SLAVE")
				handleOrderChannel <- messageFromSlave
				if(messageFromSlave.NewOrder){
				 	sendUdpMessage(Source.Message{false, false, true, false, false, elevator.ID, -1, elevator, Source.ButtonMessage{0,0,0}})
					 <- messageFromMasterChannel
				}else if (messageFromSlave.UpdatedElevInfo){
					sendUdpMessage(Source.Message{false, false, true, false, true, elevator.ID, -1, messageFromSlave.ElevInfo, Source.ButtonMessage{0,0,0}})
					 <- messageFromMasterChannel
				} else if (messageFromSlave.CompletedOrder) {
					sendUdpMessage(Source.Message{false, false, true, true, false, elevator.ID, -1, elevator, messageFromSlave.Button})
					 <- messageFromMasterChannel
				}              
	                
		    case newOrder := <-externalOrderChannel:
	        	// SEND PÅ EIN KANAL ELLER GJER SÅNN : calculatedElev := Queue.CalculateCost(message.Button) // RETURNS kva heis som tar ordren
				newOrderMessage := Source.Message{true, false, false, false, false, elevator.ID, -1, elevator, newOrder}
				handleOrderChannel <- newOrderMessage
				println("New Order ")
		        
			case distributedOrder := <- bestElevatorChannel:
				println("distributeOrder", distributedOrder.MessageTo)
				if( distributedOrder.MessageTo == elevator.ID){
					break
				}
				sendUdpMessage(distributedOrder)
				println("distributeOrder", distributedOrder.MessageTo)
				 <- messageFromMasterChannel
				if(elevator.ID != distributedOrder.MessageTo){				
					ack:
					for{
						select {
							case reply := <- messageFromSlaveChannel:
								if(reply.MessageFrom == distributedOrder.MessageTo && reply.AcceptedOrder){
									reply.FromMaster = true
									sendUdpMessage(reply)
									<- messageFromMasterChannel
									handleOrderChannel <- reply
									break ack
								} else{
									println("IKKJE ACK PÅ NY ORDRE")
								}
													
							case <-time.After(200*time.Millisecond):
								// NO ACK.
								println("No ACK")
								removeElevator <- distributedOrder.MessageTo
								distributedOrder.FromMaster = false
								//handleOrderChannel <- distributeOrder
								break ack

						}
					}
				}

			case completedOrder := <-completedOrderChannel:
				println("COMPLETEd")
				sendUdpMessage(Source.Message{false, false, true, true, false, elevator.ID, -1, elevator, completedOrder})
				 <- messageFromMasterChannel	
				//handleOrderChannel <- Source.Message{false, false, true, true, false, elevator.ID, -1, elevator, completedOrder}

			case elevatorInfo :=  <- updateElevatorInfoChannel:
				println("SENDING UPDATE")
				sendUdpMessage(Source.Message{false, false, true, false, true, elevator.ID, -1, elevatorInfo, Source.ButtonMessage{0,0,0}})
				 <- messageFromMasterChannel
				handleOrderChannel <- Source.Message{false, false, true, false, true, elevator.ID, -1, elevatorInfo, Source.ButtonMessage{0,0,0}} 

     	}      
    }
}