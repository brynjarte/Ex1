package UDP

import(	
	"net"
	"encoding/json"
	"time"
	"driver"
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
	Button driver.ButtonMessage
}

type Elevator struct {
	ID int	
	CurrentFloor int
	Direction int
	//numberInQueue int
}*/



func recieveUdpMessage(responseChannel chan Source.Message){
	
	buffer := make([]byte,1024)
	raddr,_ := net.ResolveUDPAddr("udp", ":26969")
	recievesock,_ := net.ListenUDP("udp", raddr)
	var recMsg Source.Message
	for {
		mlen , _,_ := recievesock.ReadFromUDP(buffer)
		json.Unmarshal(buffer[:mlen], &recMsg)
		responseChannel <- recMsg 
	}
}

func sendUdpMessage(msg Source.Message){

	baddr,err := net.ResolveUDPAddr("udp", "129.241.187.255:26969")
	sendSock, err := net.DialUDP("udp", nil ,baddr) // connection
	//time.Sleep(200*time.Millisecond)
	buf,_ := json.Marshal(msg)
	_,err = sendSock.Write(buf)
	if err != nil{
		panic(err)
	}
	
}

func Slave(completedOrderChannel chan Source.Message, externalOrderChannel chan driver.ButtonMessage, /*elevInfoChannel chan Elevator,*/ addOrderChannel chan Source.Message, removeExternalOrderChannel chan Source.Message){
	println("Starter SALVE")
	responseChannel := make(chan Source.Message,1)
	checkIdChannel := make(chan int, 1)
 	var elevator = Source.Elevator{2,0,0}
	go recieveUdpMessage(responseChannel)
    	
	for{
		select{
	
			case message := <-responseChannel:
				println("Message <- respChan")
				if (message.FromMaster || message.CompletedOrder) && !(message.ElevInfo.ID == elevator.ID) {
					if (message.NewOrder) {
                        addOrderChannel <- message
						if (message.MessageTo == elevator.ID) {
							sendUdpMessage(Source.Message{false, false, false, false, -1, elevator, message.Button})
						}
					}else if message.CompletedOrder {
						removeExternalOrderChannel <- message
					} 
				}
        
			case newOrder := <-externalOrderChannel:
				println("newOrder <- extOrdChan")
				for i := 0; i < 4; i++ {
					sendUdpMessage(Source.Message{true, false, false, false, -1, elevator, newOrder})
					
					select {
						case reply := <-responseChannel:
							if !(reply.ElevInfo.ID == elevator.ID) {
								checkIdChannel <- 1
							}
							select {
								case  <- checkIdChannel:
									println("Feil ID")
									i = 4
									responseChannel <- reply 
									break
							

								case <-time.After(200*time.Millisecond):
									println("time.After()")
									if (i < 3) {						
										break
									} else {
										msg := Source.Message{true, false, false, false, elevator.ID, elevator, newOrder}
										addOrderChannel <- msg
										go master(completedOrderChannel, externalOrderChannel, /*elevInfoChannel, */addOrderChannel, removeExternalOrderChannel, responseChannel)
										return
									}
							}
					}
				}

			/*case completedOrder := <-completedOrderChannel:
				for i := 0; i < 4; i++ {
					sendUdpMessage(Message{false, false, true, false, -1, elevator, completedOrder.Button}) // FJERNER ORDREN EIN ANNA PLASS?
					select {
						case reply := <-responseChannel:
							i = 4
							responseChannel <- reply 
							break 
						case <-time.After(200*time.Millisecond):
							if (i < 3) {						
								break
							} else {
								go master(completedOrderChannel, externalOrderChannel, elevInfoChannel, addOrderChannel, removeOrderChannel, responseChannel)
								return
							}
					}
				}
				
			//Skifter retn. el. etg.: 
			case elevatorInfo := <- elevInfoChannel:
				sendUdpMessage(Message{false, false, false, true, -1, elevatorInfo, driver.ButtonMessage{0,0,0}})*/
		}
	}
}


func master(completedOrderChannel chan Source.Message, externalOrderChannel chan driver.ButtonMessage,/* elevInfoChannel chan Elevator,*/ addOrderChannel chan Source.Message, removeExternalOrderChannel chan Source.Message, responseChannel chan Source.Message){
	println("Starter MASTAH")
	var elevator = Source.Elevator{2,0,0}

	for {
        select{


			case message:= <-responseChannel:
				if message.FromMaster{
					if(message.ElevInfo.ID < elevator.ID){
						go Slave(completedOrderChannel, externalOrderChannel, /*elevInfoChannel,*/ addOrderChannel, removeExternalOrderChannel)	
						return
					}
				}
					
				if message.ElevInfo.ID == elevator.ID {
					break
				}
		        if message.NewOrder{
	              	externalOrderChannel <- message.Button
		        }else if message.CompletedOrder {
					removeExternalOrderChannel <- message
				}/*else{
	        		elevInfoChannel <- message.ElevInfo
				}       */         
	                
		    case newOrder := <-externalOrderChannel:
	        	// SEND PÅ EIN KANAL ELLER GJER SÅNN : calculatedElev := Queue.CalculateCost(message.Button) // RETURNS kva heis som tar ordren
				calculatedElev := 1
				msg := Source.Message{true, true, false, false, calculatedElev, elevator, newOrder}
				println("Sendes denne heeele tiden?")
		        sendUdpMessage(msg)
 
				for{
					if(elevator.ID == calculatedElev){
						addOrderChannel <- msg       //LEGGER TIL I KØ OG SETTER LYS
						break                                                    
					}
					select {
						case reply := <-responseChannel:
							if(reply.ElevInfo.ID == calculatedElev){
								addOrderChannel <- msg
								break
													} 
						case <-time.After(200*time.Millisecond):
							// NO ACK.
							//Queue.removeElvatorFromCalculateCost(calculatedElev)
							//calculatedElev := Queue.CalculateCost(message.Button) // RETURNS kva heis som tar ordren
							// HEIS UTE AV NETTVERK, KVA GJER ME MED BESTILLINGAR????
							calculatedElev := 2
							sendUdpMessage(Source.Message{true, true, false, false, calculatedElev, elevator, newOrder})
					}
				}

			/*case completedOrder := <-completedOrderChannel:
				sendUdpMessage(Message{false, true, true, false, -1, elevator, completedOrder.Button}) // FJERNER ORDREN EIN ANNA PLASS?*/

			//case elevInfo := <-elevInfoChannel:
				//OPPDATER INFO OM HEISAR.
//				break

     	}      
    }
}
