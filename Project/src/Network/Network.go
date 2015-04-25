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
				mlen , _, _ := recievesock.ReadFromUDP(buffer)
				if(mlen > 0){
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
	sendSock, err := net.DialUDP("udp", nil ,baddr) 
	buf, err:= json.Marshal(msg)
	_,err = sendSock.Write(buf)

	if( err != nil){
		Source.ErrorChannel <- err
	}	
}

func Slave(elevator Source.ElevatorInfo, externalOrderChannel chan Source.ButtonMessage, handleOrderChannel chan Source.Message, bestElevatorChannel chan Source.Message, completedOrderChannel chan Source.ButtonMessage, updateElevatorInfoChannel chan Source.ElevatorInfo, removeElevatorChannel chan int, requestQueueChannel chan int, receiveQueueChannel chan Source.Message){
	fmt.Println("\x1b[35;1mStarter slave\x1b[0m")
	msgFromMasterChannel := make(chan Source.Message,1)
	terminate := make(chan bool, 1)
	terminated := make(chan int, 1)
	go recieveUdpMessage(false, msgFromMasterChannel, terminate, terminated)
	go syncQueues(false, elevator , requestQueueChannel , receiveQueueChannel, msgFromMasterChannel)
    	
	for{
		select{
			case messageFromMaster := <- msgFromMasterChannel:
				handleOrderChannel  <- messageFromMaster
				if (messageFromMaster.NewOrder && messageFromMaster.MessageTo == elevator.ID) {
					sendUdpMessage(Source.Message{false, true, false, false, false, elevator.ID, -1, elevator, messageFromMaster.Button, nil})
				}

			case newOrder := <-externalOrderChannel:
				for i := 0; i < 4; i++ {
					sendUdpMessage(Source.Message{true, false, false, false, false, elevator.ID, -1, elevator, newOrder, nil})
					select {
						case reply := <-msgFromMasterChannel:
							i = 4
							msgFromMasterChannel <- reply 
							break
						
						case <-time.After(200*time.Millisecond):
							if (i < 3) {						
								break
							} else {
								msg := Source.Message{true, true, true, false, false, elevator.ID, elevator.ID, elevator, newOrder, nil}
								handleOrderChannel <- msg
								terminate <- true
								<- terminated
								go master(elevator, externalOrderChannel, handleOrderChannel, bestElevatorChannel, completedOrderChannel, updateElevatorInfoChannel, removeElevatorChannel, requestQueueChannel, receiveQueueChannel)
								return
							}
					}
				}

			case completedOrder := <-completedOrderChannel:
				for i := 0; i < 4; i++ {
					sendUdpMessage(Source.Message{false, false, false, true, false, elevator.ID, -1, elevator, completedOrder, nil}) 
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
								go master(elevator, externalOrderChannel, handleOrderChannel, bestElevatorChannel, completedOrderChannel, updateElevatorInfoChannel, removeElevatorChannel, requestQueueChannel, receiveQueueChannel)
								return
							}
					}
				}
				
			case elevatorInfo :=  <- updateElevatorInfoChannel:
				elevator = elevatorInfo
				updatedElevInfo := Source.Message{false, false, false, false, true, elevator.ID, -1, elevator, Source.ButtonMessage{0,0,0}, nil}
				sendUdpMessage(updatedElevInfo)
				handleOrderChannel <- updatedElevInfo

			case <- time.After(time.Second):
				sendUdpMessage(Source.Message{false, false, false, false, true, elevator.ID, -1, elevator, Source.ButtonMessage{0,0,0}, nil})
				select {
						case reply := <-msgFromMasterChannel:
							msgFromMasterChannel <- reply 
							break 
                        case <- time.After(200*time.Millisecond):
								terminate <- true
								<- terminated
								go master(elevator, externalOrderChannel, handleOrderChannel, bestElevatorChannel, completedOrderChannel, updateElevatorInfoChannel, removeElevatorChannel, requestQueueChannel, receiveQueueChannel)
								return
				}
		}
	}
}

func master(elevator Source.ElevatorInfo, externalOrderChannel chan Source.ButtonMessage, handleOrderChannel chan Source.Message, bestElevatorChannel chan Source.Message, completedOrderChannel chan Source.ButtonMessage, updateElevatorInfoChannel chan Source.ElevatorInfo, removeElevatorChannel chan int, requestQueueChannel chan int, receiveQueueChannel chan Source.Message){
	fmt.Println("\x1b[31;1mStarter master\x1b[0m")
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
				go syncQueues(true, elevator, requestQueueChannel, receiveQueueChannel, messageFromMasterChannel)
				if(messageFromMaster.MessageFrom < elevator.ID){
					terminateFromSlave <- true
					terminateFromMaster <- true
					<- terminatedFromSlave
					<- terminatedFromMaster
					go Slave(elevator, externalOrderChannel, handleOrderChannel, bestElevatorChannel, completedOrderChannel, updateElevatorInfoChannel, removeElevatorChannel, requestQueueChannel, receiveQueueChannel)
					return
				}	

			case messageFromSlave := <- messageFromSlaveChannel:
				handleOrderChannel <- messageFromSlave
				if(messageFromSlave.NewOrder){
				 	sendUdpMessage(Source.Message{false, false, true, false, false, elevator.ID, -1, elevator, Source.ButtonMessage{0,0,0}, nil})
					<- messageFromMasterChannel
				}else if (messageFromSlave.UpdatedElevInfo){
					sendUdpMessage(Source.Message{false, false, true, false, true, elevator.ID, -1, messageFromSlave.ElevInfo, Source.ButtonMessage{0,0,0}, nil})
					sendUdpMessage(Source.Message{false, false, true, false, true, elevator.ID, -1, elevator, Source.ButtonMessage{0,0,0}, nil})
					<- messageFromMasterChannel
					<- messageFromMasterChannel
				} else if (messageFromSlave.CompletedOrder) {
					sendUdpMessage(Source.Message{false, false, true, true, false, messageFromSlave.MessageFrom, -1, elevator, messageFromSlave.Button, nil})
					<- messageFromMasterChannel
				} else if (messageFromSlave.AllExternalOrders != nil) {
					sendUdpMessage(Source.Message{false, false, true, false, false, elevator.ID, -1, elevator, Source.ButtonMessage{-1,-1,-1}, messageFromSlave.AllExternalOrders})
				}
	                
			case newOrder := <-externalOrderChannel:
				newOrderMessage := Source.Message{true, false, false, false, false, elevator.ID, -1, elevator, newOrder, nil}
				handleOrderChannel <- newOrderMessage

		        
			case distributedOrder := <- bestElevatorChannel:
				if(distributedOrder.MessageTo == elevator.ID){
					distributedOrder.AcceptedOrder = true
					handleOrderChannel <- distributedOrder
					sendUdpMessage(distributedOrder)
					<- messageFromMasterChannel	
					break
				}
				sendUdpMessage(distributedOrder)
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
							} 
				
						case <-time.After(200*time.Millisecond):
							removeElevatorChannel <- distributedOrder.MessageTo
							distributedOrder.FromMaster = false
							break ack
					}
				}

			case completedOrder := <-completedOrderChannel:
				sendUdpMessage(Source.Message{false, false, true, true, false, elevator.ID, -1, elevator, completedOrder, nil})
				<- messageFromMasterChannel	
				handleOrderChannel <- Source.Message{false, false, true, true, false, elevator.ID, -1, elevator, completedOrder, nil}

			case elevatorInfo :=  <- updateElevatorInfoChannel:
				elevator = elevatorInfo
				sendUdpMessage(Source.Message{false, false, true, false, true, elevator.ID, -1, elevatorInfo, Source.ButtonMessage{0,0,0}, nil})
				<- messageFromMasterChannel
				handleOrderChannel <- Source.Message{false, false, true, false, true, elevator.ID, -1, elevatorInfo, Source.ButtonMessage{0,0,0}, nil} 
			
			case <- time.After(1500*time.Millisecond):
				removeElevatorChannel <- -1
				sendUdpMessage(Source.Message{false, false, true, false, true, elevator.ID, -1, elevator, Source.ButtonMessage{0,0,0}, nil})
				<-messageFromMasterChannel
		}      
	}
}


func syncQueues(master bool, levator Source.ElevatorInfo, requestQueueChannel chan int, receiveQueueChannel chan Source.Message, messageFromMasterChannel chan Source.Message){
	 
	requestQueueChannel <- 1
	order := <-receiveQueueChannel
	sendUdpMessage(order)
	if (master) {	
		<- messageFromMasterChannel
	}
}
