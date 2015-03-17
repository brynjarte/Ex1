<<<<<<< HEAD
package Elevator

import(
	"UDP"
	"driver"
	"Queue"	
	"time"
   	"fmt"
)



func Elevator(){
	elev1 := UDP.Elevator{0,0,0}
	elev2 := UDP.Elevator{1,0,0}
	elev3 := UDP.Elevator{2,0,0}
	elevatorInfoChannel := make(chan UDP.Elevator,1)
	UpdateElevatorInfoChannel := make(chan UDP.Elevator,1)
	sensorChannel := make(chan int,1)
	
	var prevFloor int = 0
    go driver.ReadSensors(sensorChannel)
    
	for{
		select{
			case currentFloor := <-sensorChannel:// FLOOR REACHED
				movingDirection := prevFloor-currentFloor
				prevFloor = currentFloor
				elev1.CurrentFloor = currentFloor
				elev1.Direction = movingDirection
				elevatorInfoChannel <- elev1
				//SJEKK OM ME SKAL STOPPA.

			case newExternalOrder := <- addOrderChannel:
				if(newExternalOrder.ElevatorID == elevator.ElevatorID){
		        		Queue.AddOrder(newExternalOrder.Button,newExternalOrder.elevatorID, currentFloor, movingDirection)
                                        } 
                                	driver.Elev_set_button_lamp(newExternalOrder.Button, newExternalOrder.Floor, 1)           	
           	case updatedElevInfo := <- UpdateElevatorInfoChannel:
				if(elev1.ElevatorID == updatedElevInfo.ElevatorID){
					elev1 = updatedElevInfo
				} else if (elev2.ElevatorID == updatedElevInfo.ElevatorID){
					elev2 = updatedElevInfo
				} else if (elev3.ElevatorID == updatedElevInfo.ElevatorID){
					elev3 = updatedElevInfo
				}
						
			}
		}
	}
}


