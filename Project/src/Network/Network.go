package Network

import(	
	"net"
	"encoding/json"
	"time"
	"Source"
	"fmt"
)


func recieveUdpMessage(master bool, responseChannel chan Source.Message, terminate chan bool, terminated chan int){
	
	buffer := make([]byte, 4098)
	raddr, err := net.ResolveUDPAddr("udp", ":26969")
	Source.ErrorChannel <- err
	if(master){
		raddr, err = net.ResolveUDPAddr("udp", ":27000")
		Source.ErrorChannel <- err
	}
	
	recievesock, err := net.ListenUDP("udp", raddr)
	Source.ErrorChannel <- err
	var recMsg Source.Message
	for {
		err := recievesock.SetReadDeadline(time.Now().Add(50*time.Millisecond))
		Source.ErrorChannel <- err
		select{
			case <- terminate:
				err := recievesock.Close()
				Source.ErrorChannel <- err
				terminated <- 1
				return
			default:
				
				mlen , _, err := recievesock.ReadFromUDP(buffer)
				Source.ErrorChannel <- err
				if(mlen > 0){
					println("RECEIVEEEE")
					err := json.Unmarshal(buffer[:mlen], &recMsg)
					Source.ErrorChannel <- err
					responseChannel <- recMsg
				}
				
		}
	
	}
}

func sendUdpMessage(msg Source.Message){
	baddr,err := net.ResolveUDPAddr("udp", "129.241.187.255:27000")

	if(msg.FromMaster){
		baddr,err = net.ResolveUDPAddr("udp", "129.241.187.255:26969")
	}
	println("MSG", msg.UpdatedElevInfo)
	sendSock, err := net.DialUDP("udp", nil ,baddr) // connection
	
	buf, err:= json.Marshal(msg)
	
	_,err = sendSock.Write(buf)

	if( err != nil){
		Source.ErrorChannel <- err
	}	

}

func Slave(elevator Source.ElevatorInfo, externalOrderChannel chan Source.ButtonMessage, updateElevatorInfoChannel chan Source.ElevatorInfo, handleOrderChannel chan Source.Message, bestElevatorChannel chan Source.Message, removeElevator chan int, completedOrderChannel chan Source.ButtonMessage, requestQueueChannel chan int, receiveQueueChannel chan Source.Message){

	fmt.Println("\x1b[31;1mStarter SALVE!\x1b[0m")
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
								msg := Source.Message{true, true, true, false, false, elevator.ID, elevator.ID, elevator, newOrder}
								handleOrderChannel <- msg
								terminate <- true
								<- terminated
								go master( elevator, externalOrderChannel, updateElevatorInfoChannel, handleOrderChannel, bestElevatorChannel, removeElevator, completedOrderChannel, requestQueueChannel, receiveQueueChannel)
								return
							}
						
					}
				}

			case completedOrder := <-completedOrderChannel:
				println("Sending Completed Order")
				for i := 0; i < 4; i++ {
					sendUdpMessage(Source.Message{false, false, false, true, false, elevator.ID, -1, elevator, completedOrder}) 
					select {
						case reply := <-msgFromMasterChannel:
							i = 4
							msgFromMasterChannel <- reply 
							break 
						case <-time.After(200*time.Millisecond):
							if (i < 3) {						
								break
							} else {
								terminate <- true
								<- terminated
								go master( elevator, externalOrderChannel, updateElevatorInfoChannel, handleOrderChannel, bestElevatorChannel, removeElevator, completedOrderChannel, requestQueueChannel, receiveQueueChannel)
								return
							}
					}
				}
				
			//Skifter retn. el. etg.: 
			case elevatorInfo :=  <- updateElevatorInfoChannel:
				println("Sending elev info from salve")
				updatedElevInfo := Source.Message{false, false, false, false, true, elevator.ID, -1, elevatorInfo, Source.ButtonMessage{0,0,0}}
				sendUdpMessage(updatedElevInfo)
				handleOrderChannel <- updatedElevInfo
		}
	}
}


func master( elevator Source.ElevatorInfo, externalOrderChannel chan Source.ButtonMessage, updateElevatorInfoChannel chan Source.ElevatorInfo, handleOrderChannel chan Source.Message, bestElevatorChannel chan Source.Message, removeElevator chan int, completedOrderChannel chan Source.ButtonMessage, requestQueueChannel chan int, receiveQueueChannel chan Source.Message){
	fmt.Println("\x1b[31;1mStarter MASTAH!\x1b[0m")
	println("Vår id,",elevator.ID)
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
						fmt.Println("\x1b[31;1mEVIG RUNDDANS?????????!\x1b[0m")
						requestQueueChannel <- 1
						sendQueue: 
						for {
							select {
								case order := <-receiveQueueChannel:
									if (order.MessageTo == -1) {
										break sendQueue
									}
									sendUdpMessage(order)
							}
						}

						terminateFromSlave <- true
						terminateFromMaster <- true
						<- terminatedFromSlave
						<- terminatedFromMaster
						go Slave( elevator, externalOrderChannel, updateElevatorInfoChannel, handleOrderChannel, bestElevatorChannel, removeElevator, completedOrderChannel, requestQueueChannel, receiveQueueChannel)	
						return	
				} else if (messageFromMaster.AcceptedOrder && !messageFromMaster.NewOrder){
					handleOrderChannel <- messageFromMaster
				}
			case messageFromSlave := <- messageFromSlaveChannel:
				println("MESSAGE FROM SLAVE", messageFromSlave.UpdatedElevInfo)
				handleOrderChannel <- messageFromSlave
				if(messageFromSlave.NewOrder){
				 	sendUdpMessage(Source.Message{false, false, true, false, false, elevator.ID, -1, elevator, Source.ButtonMessage{0,0,0}})
					 <- messageFromMasterChannel
				}else if (messageFromSlave.UpdatedElevInfo){
					sendUdpMessage(Source.Message{false, false, true, false, true, elevator.ID, -1, messageFromSlave.ElevInfo, Source.ButtonMessage{0,0,0}})
					 <- messageFromMasterChannel
				} else if (messageFromSlave.CompletedOrder) {
					sendUdpMessage(Source.Message{false, false, true, true, false, messageFromSlave.MessageFrom, -1, elevator, messageFromSlave.Button})
					 <- messageFromMasterChannel
				}              
	                
			case newOrder := <-externalOrderChannel:
				newOrderMessage := Source.Message{true, false, false, false, false, elevator.ID, -1, elevator, newOrder}
				handleOrderChannel <- newOrderMessage
				println("New Order ")
		        
			case distributedOrder := <- bestElevatorChannel:
				println("distributeOrder", distributedOrder.MessageTo)
				if( distributedOrder.MessageTo == elevator.ID){
					distributedOrder.AcceptedOrder = true
					handleOrderChannel <- distributedOrder
					sendUdpMessage(distributedOrder)
					break
				}
				sendUdpMessage(distributedOrder)
				println("distributeOrder", distributedOrder.MessageTo)
				 <- messageFromMasterChannel		
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
				

			case completedOrder := <-completedOrderChannel:
				println("COMPLETEd")
				sendUdpMessage(Source.Message{false, false, true, true, false, elevator.ID, -1, elevator, completedOrder})
				 <- messageFromMasterChannel	
				handleOrderChannel <- Source.Message{false, false, true, true, false, elevator.ID, -1, elevator, completedOrder}

			case elevatorInfo :=  <- updateElevatorInfoChannel:
				println("SENDING UPDATE")
				sendUdpMessage(Source.Message{false, false, true, false, true, elevator.ID, -1, elevatorInfo, Source.ButtonMessage{0,0,0}})
				<- messageFromMasterChannel
				handleOrderChannel <- Source.Message{false, false, true, false, true, elevator.ID, -1, elevatorInfo, Source.ButtonMessage{0,0,0}} 

     	}      
    }
}
