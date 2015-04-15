
package Elevator

import(
	//"UDP"
	"driver"
	"Queue"	
	"time"
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
    emptyQueue := make(chan int, 1)
    stop := make(chan int, 1)

	//driver
	newOrderChannel := make(chan driver.ButtonMessage,1)
	floorReachedChannel := make(chan int,1)
	setSpeedChannel := make(chan int,1)
	stopChannel := make(chan int,1)
	stoppedChannel := make(chan int,1)
	setButtonLightChannel := make(chan driver.ButtonMessage,1) 

	//Queue
	addOrderChannel := make(chan driver.ButtonMessage,1)
	removeOrderChannel := make(chan int,1)
	setDirectionChannel := make(chan int,1)
	checkDirectionChannel := make(chan int,1)
	checkOrdersChannel := make(chan int, 1)
    
	//UDP

	
	direction := 1
	currentFloor := 0
	prevFloor := 0

	go driver.Drivers(newOrderChannel, floorReachedChannel, setSpeedChannel, stopChannel, stoppedChannel, setButtonLightChannel)
  	go Queue.Queue(addOrderChannel, removeOrderChannel, setDirectionChannel, checkDirectionChannel, checkOrdersChannel, stop)
    

    emptyQueue <- 1
 
	for{
		select{
			case arrivedAtFloor := <- floorReachedChannel:// FLOOR REACHED				
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
                    break
												 	
			case dir := <- stop:

                //println("StopChannel")
				stopChannel <- dir
                removeOrderChannel <- currentFloor
				wait <- 1
                break

			case <- wait:
                println("ELEVATOR :WAIT")
				doorOpen:
				for{
                   // println("WAIT")
					select{
						case <- stoppedChannel:
                            //println("STOPPED")
							        
                            run <- 1
							break doorOpen
						case newInternalOrder := <- newOrderChannel:
                            go newOrder( newInternalOrder, addOrderChannel , setButtonLightChannel)
					}
				}
   	
			case <- run:
                println("ELEVATOR :RUN")
				checkDirectionChannel <- direction
				movingDirection := <- setDirectionChannel
				if(movingDirection == -1){
					emptyQueue <- 1
					break
				} else{
					checkOrdersChannel <- currentFloor
					select{
						case <-stop:
							stop <- 1
							break
						
                        case <- time.After(30*time.Millisecond):
							setSpeedChannel <- movingDirection
							direction = movingDirection
                            break                         
                          
					}
				}
					
            case <-emptyQueue:
                println("ELEVATOR :EMPTYQUEUE")
                emptyQueue:
                for{
                    select{
                        case newInternalOrder := <- newOrderChannel:
                             go newOrder( newInternalOrder, addOrderChannel , setButtonLightChannel)
                            run <- 1
                            break emptyQueue
                    }
                }

			case newInternalOrder := <- newOrderChannel:
			    go newOrder( newInternalOrder, addOrderChannel , setButtonLightChannel)
			
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


func newOrder( newInternalOrder driver.ButtonMessage, addOrderChannel chan driver.ButtonMessage, setButtonLightChannel chan driver.ButtonMessage){
	newInternalOrder.Light = 1
	//if(newInternalOrder.Button == driver.BUTTON_COMMAND){
		addOrderChannel <- newInternalOrder
		setButtonLightChannel <- newInternalOrder
	//} else{
		//Send til networkchannel
	//}
}


