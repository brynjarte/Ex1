package UDP

import(	
	"net"
	"encoding/json"
	"time"
    "Source"
	//"fmt"
)

/*type Message struct{
	NewOrder bool
	FromMaster bool
	CompletedOrder bool
	UpdatedElevInfo bool
	MessageTo int

	ElevInfo Elevator
	Button Source.ButtonMessage
}

type Elevator struct {
	ID int	
	CurrentFloor int
	Direction int
	//numberInQueue int
}*/





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
					println("MESSAGE RECEIVED")
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

func Slave(completedOrderChannel chan Source.Message, externalOrderChannel chan Source.ButtonMessage, updateElevatorInfoChannel chan Source.Elevator, addOrderChannel chan Source.Message, removeExternalOrderChannel chan Source.Message, newElevInfoChannel chan Source.Elevator, fromUdpToQueue chan Source.Message){
	
	println("Starter SALVE")
	msgFromMasterChannel := make(chan Source.Message,1)
	terminate := make(chan bool, 1)
	terminated := make(chan int, 1)
	var elevator = Source.Elevator{2,0,0}
	go recieveUdpMessage(false, msgFromMasterChannel, terminate, terminated)
    	
	for{
		select{
	
			case messageFromMaster := <- msgFromMasterChannel:
				println("Message <- respChan")
				fromUdpToQueue <- messageFromMaster
				if (messageFromMaster.NewOrder) {
					if (messageFromMaster.MessageTo == elevator.ID) {
						sendUdpMessage(Source.Message{false, false, false, false, elevator.ID, -1, elevator, messageFromMaster.Button})
					}
				}
				
        
			case newOrder := <-externalOrderChannel:
				println("newOrder <- extOrdChan")
				for i := 0; i < 4; i++ {
					sendUdpMessage(Source.Message{true, false, false, false, elevator.ID, -1, elevator, newOrder})
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
								msg := Source.Message{true, false, false, false, elevator.ID, 2, elevator, newOrder}
								addOrderChannel <- msg
								terminate <- true
								sendUdpMessage(Source.Message{false, true, false, true, elevator.ID, -1, elevator, Source.ButtonMessage{0,0,0}}) // DUMMY MESSAGE
								<- terminated
								go master(completedOrderChannel, externalOrderChannel, updateElevatorInfoChannel, addOrderChannel, removeExternalOrderChannel, newElevInfoChannel, fromUdpToQueue)
								return
							}
						
					}
				}

			/*case completedOrder := <-completedOrderChannel:
				for i := 0; i < 4; i++ {
					sendUdpMessage(Message{false, false, true, false, -1, elevator, completedOrder.Button}) // FJERNER ORDREN EIN ANNA PLASS?
					select {
						case reply := <-msgFromMasterChannel:
							i = 4
							msgFromMasterChannel <- reply 
							break 
						case <-time.After(200*time.Millisecond):
							if (i < 3) {						
								break
							} else {
								go master(completedOrderChannel, externalOrderChannel, elevInfoChannel, addOrderChannel, removeOrderChannel, newElevInfoChannel)
								return
							}
					}
				}
				*/
			//Skifter retn. el. etg.: 
			case elevatorInfo :=  <- updateElevatorInfoChannel:
				sendUdpMessage(Source.Message{false, false, false, true, elevator.ID, -1, elevatorInfo, Source.ButtonMessage{0,0,0}})
				
		}
	}
}


func master(completedOrderChannel chan Source.Message, externalOrderChannel chan Source.ButtonMessage, updateElevatorInfoChannel chan Source.Elevator, addOrderChannel chan Source.Message, removeExternalOrderChannel chan Source.Message, newElevInfoChannel chan Source.Elevator, fromUdpToQueue chan Source.Message){
	println("Starter MASTAH")
	var elevator = Source.Elevator{2,0,0}

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
				println("TO MASTERAA")
				if(messageFromMaster.MessageFrom < elevator.ID){
						terminateFromSlave <- true
						terminateFromMaster <- true
						sendUdpMessage(Source.Message{false, true, false, true, -1, -1, Source.Elevator{-1, -1, -1}, Source.ButtonMessage{0,0,0}})
						sendUdpMessage(Source.Message{false, false, false, true, -1, -1, Source.Elevator{-1, -1, -1}, Source.ButtonMessage{0,0,0}})
						<- terminatedFromSlave
						<- terminatedFromMaster
						go Slave(completedOrderChannel, externalOrderChannel, updateElevatorInfoChannel, addOrderChannel, removeExternalOrderChannel, newElevInfoChannel, fromUdpToQueue)	
						return	
				}
			case messageFromSlave := <- messageFromSlaveChannel:
				println("MESSAGE FROM SLAVE")

		        if messageFromSlave.NewOrder{
	              	externalOrderChannel <- messageFromSlave.Button

		        }else if messageFromSlave.CompletedOrder {
					removeExternalOrderChannel <- messageFromSlave

				}else if (messageFromSlave.UpdatedElevInfo){
	        		newElevInfoChannel <- messageFromSlave.ElevInfo
					sendUdpMessage(Source.Message{false, true, false, true, elevator.ID, -1, messageFromSlave.ElevInfo, Source.ButtonMessage{0,0,0}})
				}               
	                
		    case newOrder := <-externalOrderChannel:
	        	// SEND PÅ EIN KANAL ELLER GJER SÅNN : calculatedElev := Queue.CalculateCost(message.Button) // RETURNS kva heis som tar ordren
				calculatedElev := 1
				msg := Source.Message{true, true, false, false, elevator.ID, calculatedElev, elevator, newOrder}
				println("Sendes denne heeele tiden?")
		        //sendUdpMessage(msg)
 				label:
				for{
					if(elevator.ID == calculatedElev){
						addOrderChannel <- msg       //LEGGER TIL I KØ OG SETTER LYS
						break                                                    
					}
					select {
						case reply := <- messageFromSlaveChannel:
							if(reply.MessageFrom == calculatedElev){
								addOrderChannel <- msg
								break label
													} 
						case <-time.After(200*time.Millisecond):
							// NO ACK.
							println("No ACK")
							//Queue.removeElvatorFromCalculateCost(calculatedElev)
							//calculatedElev := Queue.CalculateCost(message.Button) // RETURNS kva heis som tar ordren
							// HEIS UTE AV NETTVERK, KVA GJER ME MED BESTILLINGAR????
							calculatedElev := 2
							msg = Source.Message{true, true, false, false, elevator.ID, calculatedElev, elevator, newOrder}
							//sendUdpMessage(msg)
							if(elevator.ID == calculatedElev){
								addOrderChannel <- msg       //LEGGER TIL I KØ OG SETTER LYS
								break label                                                   
							}
							//<-msgFromMasterChannel
					}
				}

		/*	case completedOrder := <-completedOrderChannel:
				sendUdpMessage(Message{false, true, true, false, -1, elevator, completedOrder.Button}) // FJERNER ORDREN EIN ANNA PLASS?*/

			case elevatorInfo :=  <- updateElevatorInfoChannel:
				println("SENDING UPDATE")
				sendUdpMessage(Source.Message{false, true, false, true, elevator.ID, -1, elevatorInfo, Source.ButtonMessage{0,0,0}})
     	}      
    }
}
