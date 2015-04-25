
package Elevator

import(
	"Network"
	"driver"
	"Queue"	
	"Source"
	"FileHandler"
	"time"
	"fmt"
)


func Elevator(elevatorID int){

	elevatorStatus := Source.ElevatorInfo{elevatorID, -1, -1}

	//STATES
	//EventHandler
	updateElevatorInfoChannel := make(chan Source.ElevatorInfo,1)

	//ELEV
	wait := make(chan int, 1)
	run := make(chan int, 1)
	stop := make(chan int, 1)
	orderInEmptyQueue := make(chan int, 1)

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
	fromElevToQueue := make(chan Source.Message, 1)
	orderRemovedChannel := make(chan Source.ButtonMessage, 1)
	
	requestQueueChannel := make(chan int, 1)
	receiveQueueChannel := make(chan Source.Message, 10)
	//UDP

	completedOrderChannel := make(chan Source.ButtonMessage, 1)
	externalOrderChannel := make(chan Source.ButtonMessage, 10)
	handleOrderChannel := make(chan Source.Message, 1)
	bestElevatorChannel := make(chan Source.Message, 1)
	removeElevatorChannel := make(chan int, 1)
	nextOrderedFloor := 100

	

	go errorHandling(elevatorStatus.ID)
	go Source.SourceInit()
	go driver.Drivers(newOrderChannel, floorReachedChannel, setSpeedChannel, stopChannel, stoppedChannel, setButtonLightChannel, initFinished)
	<- initFinished
	println("Finished inting")
  	go Queue.Queue(elevatorStatus, addOrderChannel, removeOrderChannel, nextOrderChannel, checkOrdersChannel, orderInEmptyQueue, finishedRemoving, fromElevToQueue, bestElevatorChannel, removeElevatorChannel, completedOrderChannel, orderRemovedChannel, requestQueueChannel, receiveQueueChannel)
   	go handleOrders(elevatorStatus.ID, addOrderChannel , setButtonLightChannel, newOrderChannel, externalOrderChannel, handleOrderChannel, fromElevToQueue, orderRemovedChannel, completedOrderChannel)
	go Network.Slave(elevatorStatus, externalOrderChannel, updateElevatorInfoChannel, handleOrderChannel, bestElevatorChannel, removeElevatorChannel, completedOrderChannel, requestQueueChannel, receiveQueueChannel)

	go readFromBackup(newOrderChannel)
	
	//run <- 1
	prevFloor := 10

	
	for{
		select{
			case arrivedAtFloor := <- floorReachedChannel:// FLOOR REACHED		
				println("FLOORREACHED")
				prevFloor = elevatorStatus.CurrentFloor
				elevatorStatus.CurrentFloor = arrivedAtFloor

				checkOrdersChannel <- elevatorStatus.CurrentFloor
				nextOrder := <- nextOrderChannel
				nextOrderedFloor = nextOrder
				direction := prevFloor - elevatorStatus.CurrentFloor

				fmt.Println("\x1b[31;1mNextOrder: \x1b[0m", nextOrderedFloor)

				if(direction < 0){
					elevatorStatus.Direction = Source.UP
				}else if (direction > 0) {
					elevatorStatus.Direction = Source.DOWN
				}
				
				if(elevatorStatus.CurrentFloor < nextOrderedFloor && elevatorStatus.Direction == Source.DOWN){
					run <- 1
				} else if(elevatorStatus.CurrentFloor > nextOrderedFloor && elevatorStatus.Direction == Source.UP){
					run <- 1
				}
					
				updateElevatorInfoChannel <- elevatorStatus
				//println("currentFloor:", elevatorStatus.CurrentFloor, "\nOrderedFloor", nextOrderedFloor, "Direccion:", elevatorStatus.Direction)
				//println("NEXTORDEEER", nextOrderedFloor, "Current floor " , currentFloor)
				if(elevatorStatus.CurrentFloor == nextOrderedFloor ){
					stop <- elevatorStatus.Direction
				}
												 	
			case <- stop:
                println("ELEVATOR: StopChannel")
				stopChannel <- elevatorStatus.Direction
				wait <- 1
                break

			case <- wait:
				
                println("ELEVATOR :WAIT")
			 	wait:
				for{
					select{
						case <- stoppedChannel:
							removeOrderChannel <- elevatorStatus.CurrentFloor
							println("Removing")
							<- finishedRemoving
							println("order removed")
							run <- 1
							break wait
					}
				}
							
			case <- run:
        		println("ELEVATOR :RUN")
				checkOrdersChannel <- elevatorStatus.CurrentFloor
				orderedFloor := <- nextOrderChannel
				nextOrderedFloor = orderedFloor
				println("\x1b[31;1mELEVATOR :NEXT ORDER: \x1b[0m", nextOrderedFloor)
				if(nextOrderedFloor == -1){
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
				println("\x1b[33;1mOrder in empty queueueue\x1b[0m")
				run <- 1

		}
	}
}


func handleOrders(elevatorID int, addOrderChannel chan Source.ButtonMessage, setButtonLightChannel chan Source.ButtonMessage, newOrderChannel chan Source.ButtonMessage, externalOrderChannel chan Source.ButtonMessage, handleOrderChannel chan Source.Message, fromElevToQueue chan Source.Message, orderRemovedChannel chan Source.ButtonMessage, completedOrderChannel chan Source.ButtonMessage){
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
					fromElevToQueue <- newExternalOrder
					break
				}
				fromElevToQueue <- newExternalOrder  
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
					fromElevToQueue <- Source.Message{false, false, false, true, false, elevatorID, -1, Source.ElevatorInfo{elevatorID, -1, -1}, orderRemoved, nil}
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

func errorHandling(elevatorID int){
	//defer Elevator(elevatorID)
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



