
package Elevator

import(
	//"UDP"
	"driver"
	"Queue"	
	//"time"
   	//"fmt"
)


func Elevator(){
	/*elev1 := UDP.Elevator{0,0,0}
	elev2 := UDP.Elevator{1,0,0}
	elev3 := UDP.Elevator{2,0,0}

	//STATES
	//EventHandler
	
	ElevDirection := make(chan int,1) // 0 NEd, 1 opp
	elevatorInfoChannel := make(chan UDP.Elevator,1)
	UpdateElevatorInfoChannel := make(chan UDP.Elevator,1)// FÅR INN OPPDATERING FRÅ NETTVERKET*/

	//ELEV
	wait := make(chan int, 1)
	run := make(chan int, 1)
	//emptyQueue := make(chan int, 1)
	stop := make(chan int, 1)

	//driver
	newOrderChannel := make(chan driver.ButtonMessage,1)
	floorReachedChannel := make(chan int, 1)
	setSpeedChannel := make(chan int, 1)
	stopChannel := make(chan int, 1)
	stoppedChannel := make(chan int, 1)
	setButtonLightChannel := make(chan driver.ButtonMessage,1) 

	//Queue
	addOrderChannel := make(chan driver.ButtonMessage,1)
	removeOrderChannel := make(chan int, 1)
	nextOrderChannel := make(chan int, 1)
	checkOrdersChannel := make(chan int, 1)
    orderInEmptyQueueChannel := make(chan int, 1)
	//UDP

	nextOrderedFloor := 100
	direction := 1
	currentFloor := 0
	prevFloor := 4 // FÅR ALLTID NED RETNING ETTER INTIT

	go driver.Drivers(newOrderChannel, floorReachedChannel, setSpeedChannel, stopChannel, stoppedChannel, setButtonLightChannel)
  	go Queue.Queue(addOrderChannel, removeOrderChannel, nextOrderChannel, checkOrdersChannel, orderInEmptyQueueChannel)
   	go handleOrders(addOrderChannel , setButtonLightChannel, newOrderChannel)


	for{
		select{
			case arrivedAtFloor := <- floorReachedChannel:// FLOOR REACHED		
				
				println("ELEVATOR :FLOORREACHED")		
				prevFloor = currentFloor
				currentFloor = arrivedAtFloor
				direction = prevFloor-currentFloor
		
				if(direction < 0){
					direction = 0
				}else{
					direction = 1
				}
				/*elev1.CurrentFloor = currentFloor
				elev1.Direction = movingDirection
				elevatorInfoChannel <- elev1*/

				checkOrdersChannel <- currentFloor
				nextOrder := <- nextOrderChannel
				nextOrderedFloor = nextOrder
				println("NEXTORDEEER", nextOrderedFloor, "Current floor " , currentFloor)
				if(currentFloor == nextOrderedFloor ){
					stop <- direction
				}
												 	
			case <- stop:

                println("ELEVATOR: StopChannel")
				stopChannel <- direction
				wait <- 1
                break

			case <- wait:
				
				
                println("ELEVATOR :WAIT")
			 	wait:
				for{
					select{
						case <- stoppedChannel:
							removeOrderChannel <- currentFloor	
							println("STOPPED")
							run <- 1
							break wait
					}
				}
							
			case <- run:
        		println("ELEVATOR :RUN")
				checkOrdersChannel <- currentFloor
				orderedFloor := <- nextOrderChannel
				if(orderedFloor == -1){
					break
				}else{
					if(orderedFloor > currentFloor){
					 	setSpeedChannel <- 0
					}else if(orderedFloor < currentFloor){
						setSpeedChannel <- 1
					}else{
						stop <- direction
					}
				}
					
            case <- orderInEmptyQueueChannel:
                println("ELEVATOR :EMPTYQUEUE")
				run <- 1
				break
                

/*
		  	case updatedElevInfo := <- UpdateElevatorInfoChannel:

				if(elev1.ID == updatedElevInfo.ID){
					elev1 = updatedElevInfo
				} else if (elev2.ID == updatedElevInfo.ID){
					elev2 = updatedElevInfo
				} else if (elev3.ID == updatedElevInfo.ID){
					elev3 = updatedElevInfo
				}
				*/
/*			case newReceivedOrder := <- addOrderChannel:

				if(newReceivedOrder.MessageTo == elev1.ID){
					Queue.AddOrder( newReceivedOrder.Button, newReceivedOrder.ID, currentFloor, movingDirection)
  			  	} 
				driver.Elev_set_button_lamp(newReceivedOrder.Button) 
              	
		*/	
		}
	}
}


func handleOrders(addOrderChannel chan driver.ButtonMessage, setButtonLightChannel chan driver.ButtonMessage, newOrderChannel chan driver.ButtonMessage){
	for{
		newOrder := <- newOrderChannel
		newOrder.Light = 1
		//if(newInternalOrder.Button == driver.BUTTON_COMMAND){
			addOrderChannel <- newOrder
			setButtonLightChannel <- newOrder
			println("ORDER ADDED")
			//Queue.PrintQueue()
		//} else{
			//Send til networkchannel
		//}
	}
}


