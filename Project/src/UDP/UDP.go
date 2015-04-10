package UDP

import(	
	"net"
	"encoding/json"
	"time"
	//"fmt"
)

type Message struct{
	NewOrder bool
	FromMaster bool
	CompletedOrder bool
	UpdatedElevInfo bool
	MessageTo int

	ElevInfo Elevator
	Button ButtonMessage
}

type ButtonMessage struct {
	Floor int
	Button int
	Light int
}

type Elevator struct {
	ID int	
	CurrentFloor int
	Direction int
	//numberInQueue int
}


func recieveUdpMessage(responseChannel chan Message){
	
	buffer := make([]byte,1024)
	raddr,_ := net.ResolveUDPAddr("udp", ":26969")
	recievesock,_ := net.ListenUDP("udp", raddr)
	var recMsg Message
	for {
				mlen , _,_ := recievesock.ReadFromUDP(buffer)
				json.Unmarshal(buffer[:mlen], &recMsg)
				responseChannel <- recMsg 
	}
}

func sendUdpMessage(msg Message){

	baddr,err := net.ResolveUDPAddr("udp", "129.241.187.255:26969")
	sendSock, err := net.DialUDP("udp", nil ,baddr) // connection
	//time.Sleep(200*time.Millisecond)
	buf,_ := json.Marshal(msg)
	_,err = sendSock.Write(buf)
	if err != nil{
		panic(err)
	}
	
}

func Slave(CompletedOrderChannel chan Message, externalOrderChannel chan ButtonMessage, ElevInfoChannel chan Elevator, AddOrderChannel chan Message, RemoveOrderChannel chan Message){
	
	ResponseChannel := make(chan Message,1)
 	var elevator = Elevator{2,0,0}
	go recieveUdpMessage(ResponseChannel)
    	
	for{
		select{
	
			case message := <-ResponseChannel:
				
				if (message.FromMaster || message.CompletedOrder) && !(message.ElevInfo.ID == elevator.ID) {
					if message.NewOrder{
						if message.MessageTo == elevator.ID{
							AddOrderChannel <- message
							sendUdpMessage(Message{false, false, false, false, -1, elevator, message.Button})
						}
					}else if message.CompletedOrder {
						RemoveOrderChannel <- message
					} 
				}
        
			case newOrder := <-externalOrderChannel:
				for i := 0; i < 4; i++ {
					sendUdpMessage(Message{true, false, false, false, -1, elevator, newOrder})
					
					select {
						case reply := <-ResponseChannel:
							i = 4
							ResponseChannel <- reply 
							break 
						case <-time.After(200*time.Millisecond):
							if (i < 3) {						
								break
							} else {
								msg := Message{true, false, false, false, elevator.ID, elevator, newOrder}
								AddOrderChannel <- msg
								go master(CompletedOrderChannel, externalOrderChannel, ElevInfoChannel, AddOrderChannel, RemoveOrderChannel, ResponseChannel)
								return
							}
					}
				}

			case completedOrder := <-CompletedOrderChannel:
				for i := 0; i < 4; i++ {
					sendUdpMessage(Message{false, false, true, false, -1, elevator, completedOrder.Button}) // FJERNER ORDREN EIN ANNA PLASS?
					select {
						case reply := <-ResponseChannel:
							i = 4
							ResponseChannel <- reply 
							break 
						case <-time.After(200*time.Millisecond):
							if (i < 3) {						
								break
							} else {
								go master(CompletedOrderChannel, externalOrderChannel, ElevInfoChannel, AddOrderChannel, RemoveOrderChannel, ResponseChannel)
								return
							}
					}
				}
				
			//Skifter retn. el. etg.: 
			case elevatorInfo := <- ElevInfoChannel:
				sendUdpMessage(Message{false, false, false, true, -1, elevatorInfo, ButtonMessage{0,0,0}})
		}
	}
}


func master(CompletedOrderChannel chan Message, ExternalOrderChannel chan ButtonMessage, ElevInfoChannel chan Elevator, AddOrderChannel chan Message, RemoveOrderChannel chan Message, ResponseChannel chan Message){

	var elevator = Elevator{2,0,0}

	for {
        select{


			case message:= <-ResponseChannel:
				if message.FromMaster{
					if(message.ElevInfo.ID < elevator.ID){
						go Slave(CompletedOrderChannel, ExternalOrderChannel, ElevInfoChannel, AddOrderChannel, RemoveOrderChannel)	
						return
					}
				}
					
				if message.ElevInfo.ID == elevator.ID {
					break
				}
		        if message.NewOrder{
	              	ExternalOrderChannel <- message.Button
		        }else if message.CompletedOrder {
					RemoveOrderChannel <- message
				}else{
	        		ElevInfoChannel <- message.ElevInfo
				}                
	                
		    case newOrder := <-ExternalOrderChannel:
	        	// SEND PÅ EIN KANAL ELLER GJER SÅNN : calculatedElev := Queue.CalculateCost(message.Button) // RETURNS kva heis som tar ordren
				calculatedElev := 2
				msg := Message{true, true, false, false, calculatedElev, elevator, newOrder}
		        sendUdpMessage(msg)
 
				for{
					if(elevator.ID == calculatedElev){
						AddOrderChannel <- msg       //LEGGER TIL I KØ OG SETTER LYS
						break                                                    
					}
					select {
						case reply := <-ResponseChannel:
							if(reply.ElevInfo.ID == calculatedElev){
								AddOrderChannel <- msg
								break
													} 
						case <-time.After(200*time.Millisecond):
							// NO ACK.
							//Queue.removeElvatorFromCalculateCost(calculatedElev)
							//calculatedElev := Queue.CalculateCost(message.Button) // RETURNS kva heis som tar ordren
							// HEIS UTE AV NETTVERK, KVA GJER ME MED BESTILLINGAR????
							calculatedElev := 2
							sendUdpMessage(Message{true, true, false, false, calculatedElev, elevator, newOrder})
					}
				}

			case completedOrder := <-CompletedOrderChannel:
				sendUdpMessage(Message{false, true, true, false, -1, elevator, completedOrder.Button}) // FJERNER ORDREN EIN ANNA PLASS?

			case elevInfo := <-ElevInfoChannel:
				//OPPDATER INFO OM HEISAR.
				break

     	}      
    }
}
