
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

	//driver
	newOrderChannel := make(chan driver.ButtonMessage,1)
	sensorChannel := make(chan int,1)
	setSpeedChannel := make(chan int,1)
	stopChannel := make(chan int,1)
	stoppedChannel := make(chan int,1)
	setButtonLightChannel := make(chan driver.ButtonMessage,1) 
	setFloorLightChannel := make(chan int,1)


	//Queue
	addOrderChannel := make(chan driver.ButtonMessage,1)
	removeOrderChannel := make(chan int,1)
	setDirectionChannel := make(chan int,1)
	checkDirectionChannel := make(chan int,1)
	checkOrdersChannel := make(chan int, 1)

	//UDP

	
	direction := 0
	currentFloor := 0
	prevFloor := 0
	go driver.Drivers(newOrderChannel, sensorChannel, setSpeedChannel, stopChannel, stoppedChannel, setButtonLightChannel, setFloorLightChannel)
  	go Queue.Queue(addOrderChannel, removeOrderChannel, setDirectionChannel, checkDirectionChannel, checkOrdersChannel, stopChannel)
    
	for{
		select{
			case arrivedAtFloor := <- sensorChannel:// FLOOR REACHED
				if(currentFloor != arrivedAtFloor){				
					prevFloor = currentFloor
					currentFloor = arrivedAtFloor
					direction = prevFloor-currentFloor
			
					/*elev1.CurrentFloor = currentFloor
					elev1.Direction = movingDirection
					elevatorInfoChannel <- elev1*/

					setFloorLightChannel <- currentFloor
				}
				checkOrdersChannel <- currentFloor										 	
			
			case <- stoppedChannel:
				removeOrderChannel <- currentFloor
				checkDirectionChannel <- direction
				newDirection := <- setDirectionChannel
				direction = newDirection
				setSpeedChannel <- direction
				checkOrdersChannel <- currentFloor

		/*	case newReceivedOrder := <- addOrderChannel:
				if(newReceivedOrder.MessageTo == elev1.ID){
		        		Queue.AddOrder( newReceivedOrder.Button, newReceivedOrder.ID, currentFloor, movingDirection)
                } 
              	driver.Elev_set_button_lamp(newReceivedOrder.Button) 
              	*/
   	
			case newInternalOrder := <- newOrderChannel:
				newInternalOrder.Light = 1
				if(newInternalOrder.Button == driver.BUTTON_COMMAND){
					addOrderChannel <- newInternalOrder
					setButtonLightChannel <- newInternalOrder
				} else{
					//Send til networkchannel
				}
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
			
		}
	}
}



