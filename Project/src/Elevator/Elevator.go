package Elevator

import(
	"Network"
	"driver"
	"Queue"	
	"Source"
	"FileHandler"
	"time"
	"fmt"
	"net"
	"errors"
)

func RunElevator(){
	
	elevatorStatus := Source.ElevatorInfo{getID(), -1, -1}
	println(elevatorStatus.ID)
	
	//ELEV
	stop := make(chan int, 1)
	wait := make(chan int, 1)
	run := make(chan int, 1)
	orderInEmptyQueue := make(chan int, 1)

	noOrdersLeftChannel := make(chan int, 1)
	removedOrder := make(chan int, 1)

	//driver
	newOrderChannel := make(chan Source.ButtonMessage,1)
	floorReachedChannel := make(chan int)
	setSpeedChannel := make(chan int, 1)
	stopChannel := make(chan int, 1)
	stoppedChannel := make(chan int, 1)
	setButtonLightChannel := make(chan Source.ButtonMessage,1) 
	initFinished := make(chan int)
	
	//Queue
	addOrderChannel := make(chan Source.ButtonMessage, 1)
	removeOrderChannel := make(chan int, 1)
	nextOrderChannel := make(chan int, 1)
	checkOrdersChannel := make(chan int, 1)
	finishedRemoving := make(chan int, 1)
	networkMessageChannel := make(chan Source.Message, 1)
	orderRemovedChannel := make(chan Source.ButtonMessage, 1)
	
	//Network
	completedOrderChannel := make(chan Source.ButtonMessage, 1)
	externalOrderChannel := make(chan Source.ButtonMessage, 1) //Hadde buf = 10, sjekk om ok nå
	handleOrderChannel := make(chan Source.Message, 1)
	bestElevatorChannel := make(chan Source.Message, 1)
	removeElevatorChannel := make(chan int, 1)
	updateElevatorInfoChannel := make(chan Source.ElevatorInfo,1)
	requestQueueChannel := make(chan int, 1)
	receiveQueueChannel := make(chan Source.Message, 1) //Hadde buf = 10, sjekk om ok nå

	nextOrderedFloor := 100

	go errorHandling()
	go Source.SourceInit()
	go driver.Drivers(newOrderChannel, floorReachedChannel, setSpeedChannel, stopChannel, stoppedChannel, setButtonLightChannel, initFinished)
	<- initFinished
  	go Queue.Queue(elevatorStatus, addOrderChannel, removeOrderChannel, nextOrderChannel, checkOrdersChannel, orderInEmptyQueue, finishedRemoving, networkMessageChannel, bestElevatorChannel, removeElevatorChannel, completedOrderChannel, orderRemovedChannel, requestQueueChannel, receiveQueueChannel)
   	go handleOrders(elevatorStatus.ID, addOrderChannel, setButtonLightChannel, newOrderChannel,externalOrderChannel, handleOrderChannel, networkMessageChannel, orderRemovedChannel, completedOrderChannel)
	go Network.Slave(elevatorStatus, externalOrderChannel, handleOrderChannel, bestElevatorChannel, completedOrderChannel, updateElevatorInfoChannel, removeElevatorChannel, requestQueueChannel, receiveQueueChannel)
	go readFromBackup(newOrderChannel)
	
	prevFloor := 10
	
	for{
		select{
			case arrivedAtFloor := <- floorReachedChannel:		
				prevFloor = elevatorStatus.CurrentFloor
				elevatorStatus.CurrentFloor = arrivedAtFloor

				checkOrdersChannel <- elevatorStatus.CurrentFloor
				nextOrder := <- nextOrderChannel
				nextOrderedFloor = nextOrder
				direction := prevFloor - elevatorStatus.CurrentFloor

				if(direction < 0){
					elevatorStatus.Direction = Source.UP
				}else if (direction > 0) {
					elevatorStatus.Direction = Source.DOWN
				}	
				updateElevatorInfoChannel <- elevatorStatus
				if(elevatorStatus.CurrentFloor == nextOrderedFloor ){
					stop <- elevatorStatus.Direction
				}
												 	
			case <- stop:
				stopChannel <- elevatorStatus.Direction
				wait <- 1
                break

			case <- wait:
				
			 	wait:
				for{
					select{
						case <- stoppedChannel:
							removeOrderChannel <- elevatorStatus.CurrentFloor
							<- finishedRemoving
							removedOrder <- 1
							run <- 1
							break wait
					}
				}
							
			case <- run:
				checkOrdersChannel <- elevatorStatus.CurrentFloor
				orderedFloor := <- nextOrderChannel
				nextOrderedFloor = orderedFloor
				if(nextOrderedFloor == -1){
					noOrdersLeftChannel <- 1
					break
				}else{
					if(nextOrderedFloor > elevatorStatus.CurrentFloor){
					 	setSpeedChannel <- 0
					}else if(nextOrderedFloor < elevatorStatus.CurrentFloor){
						setSpeedChannel <- 1
					}else{
						stop <- elevatorStatus.Direction
					}
				}

            case <- orderInEmptyQueue:
				go orderDeadline(noOrdersLeftChannel, removedOrder)
				run <- 1
		}
	}
}


func handleOrders(elevatorID int, addOrderChannel chan Source.ButtonMessage, setButtonLightChannel chan Source.ButtonMessage, newOrderChannel chan Source.ButtonMessage, externalOrderChannel chan Source.ButtonMessage, handleOrderChannel chan Source.Message, networkMessageChannel chan Source.Message, orderRemovedChannel chan Source.ButtonMessage, completedOrderChannel chan Source.ButtonMessage){
	for{
		select{
			case newOrder := <- newOrderChannel:
				newOrder.Value = 1
				if(newOrder.Button == Source.BUTTON_COMMAND){
					addOrderChannel <- newOrder
					setButtonLightChannel <- newOrder
				} else{
					externalOrderChannel <- newOrder
				}

			case newExternalOrder := <- handleOrderChannel:	
				if (newExternalOrder.AllExternalOrders != nil) {
					for elev := range newExternalOrder.AllExternalOrders {
						if (elev != fmt.Sprint(elevatorID)) {
							for order := 0; order < len(newExternalOrder.AllExternalOrders[elev]); order++ {
								setButtonLightChannel <- newExternalOrder.AllExternalOrders[elev][order]
							}
						}
					}
					networkMessageChannel <- newExternalOrder
					break
				}
				networkMessageChannel <- newExternalOrder  
				if(newExternalOrder.CompletedOrder && elevatorID != newExternalOrder.MessageTo){
					newExternalOrder.Button.Value = 0
					setButtonLightChannel <- newExternalOrder.Button
				} else if(newExternalOrder.AcceptedOrder){
				 	newExternalOrder.Button.Value = 1
					setButtonLightChannel <- newExternalOrder.Button
				}

			case orderRemoved := <- orderRemovedChannel:
				orderRemoved.Value = 0
				if (orderRemoved.Button != Source.BUTTON_COMMAND) {
					completedOrderChannel <- orderRemoved
					networkMessageChannel <- Source.Message{false, false, false, true, false, elevatorID, -1, Source.ElevatorInfo{elevatorID, -1, -1}, orderRemoved, nil}
				}
				setButtonLightChannel <- orderRemoved
		}
	}
}


func readFromBackup(newOrderChannel chan Source.ButtonMessage) {
	dummy := 0
	q := FileHandler.Read(&dummy, &dummy)
	for j:=0; j < len(q); j+=2 {
		order := Source.ButtonMessage{q[j],q[j+1], 1}
		newOrderChannel <- order
		time.Sleep(20*time.Millisecond)
	}
}

func errorHandling(){
	for{
		select{
			case err := <- Source.ErrorChannel:
				if( err != nil){
					FileHandler.ErrorLog(err)
					panic(err)
					return
				}
		}
	}
}

func getID() int {
	addrs, err := net.InterfaceAddrs()
	Source.ErrorChannel <- err

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipv4 := ipnet.IP.To4(); ipv4 != nil {
				return int(ipv4[3])															
			}
		}
	} 	
	err = errors.New("GET_ID: COULD_NOT_RETRIEVE_IP")
	Source.ErrorChannel <- err
	return -1
}

func orderDeadline(noOrders chan int, removedOrder chan int){
	for{
		select{
			case <- noOrders:
				return
			case <- removedOrder:
				break
			case <- time.After(20*time.Second):
				err := errors.New("ORDER_DEADLINE: TIMED_OUT")
				Source.ErrorChannel <- err 
		}
	}
}
