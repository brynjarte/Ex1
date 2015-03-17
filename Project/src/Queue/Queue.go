package Queue

import(
	"UDP"
	"driver"
)

const elevators int = 3
const floors int = 3
const UP int = 0
const DOWN int = 1

var externalOrders [floors][2*elevators] bool  // 0 = UP, 1 = DOWN
var internalOrders [floors+1][2] bool
var queue [floors+1][2] bool
var numberInQueue [elevators] int  
var ordersInDirection[2] int;

func queueInit(){

	ordersInDirection[0] = 0
	ordersInDirection[1] = 0

	for floor:=0 ;floor < floors+1; floor++{
		internalOrders[floor][UP] = false
		internalOrders[floor][DOWN] = false
		queue[floor][UP] = false
		queue[floor][DOWN] = false
		if (floor!=floors){
			for elev:= 0; elev < 2*elevators; elev++{
				externalOrders[floor][elev] = false
			}
		}	
	}	
	for elev:= 0; elev < elevators; elev++{
			numberInQueue [elev] = 0
	}
}

func AddOrder(order UDP.ButtonMessage,elevatorID int, currentFloor int, movingDirection int){
	var directionOfOrder int = -1
	if(externalOrders[order.Floor][order.Button + 2*elevatorID]) {				// Legger kun til ordre om den er ny
		return
	}
	numberInQueue [elevatorID] ++
	externalOrders[order.Floor][order.Button + 2*elevatorID] = true
	
	if ( currentFloor - order.Floor > 0 ) {
        directionOfOrder = DOWN
    } else if ( currentFloor - order.Floor < 0 ) {
        directionOfOrder = UP
    } else {
		directionOfOrder = movingDirection
	}
    		
    if( order.Button == driver.BUTTON_CALL_UP && queue[order.Floor][UP]==false ) {
        queue[order.Floor][UP] = true
        numberInQueue[elevatorID] ++
        ordersInDirection[UP]++
    } else if( order.Button == driver.BUTTON_CALL_DOWN && queue[order.Floor][DOWN]==false ) {
        queue[order.Floor][DOWN] = true
        numberInQueue[elevatorID] ++
        ordersInDirection[DOWN] ++
    } else if(order.Button == driver.BUTTON_COMMAND) {
        internalOrders[order.Floor][directionOfOrder] = true
    	if( queue[order.Floor][directionOfOrder]==false){
            	numberInQueue[elevatorID] ++
            	ordersInDirection[directionOfOrder] ++
            	queue[order.Floor][directionOfOrder]=true
		}		
   	}
}
	

	

func CalculateCost(order UDP.ButtonMessage) int{
	min := 0
	for elev:= 0; elev < elevators; elev++{
			if(numberInQueue [elev] <= min){
				min = elev
			}
	}
	return min
}


func ClearAllExternalOrders(elevatorID int) {
	numberInQueue [elevatorID] = 0
	for floor := 0; floor < floors+1; floor++ {
		if(internalOrders[floor][UP]){
			numberInQueue [elevatorID]++
		}
		if(internalOrders[floor][DOWN]){
			numberInQueue [elevatorID]++
		}
	}
	queue = internalOrders
}



