package Queue

import(
	"FileHandler"
	"driver" // MÅ FINNA EI LØYSING, TRENGER KUN BUTTONMESSAGE
	"math"
)

var NumOfElevs int
var NumOfFloors int
const UP int = 0
const DOWN int = 1

var allQueues[5][] FileHandler.Directions
var queue[] FileHandler.Directions
var numberInQueue[5] int
var ordersInDirection[2] int 


func queueInit(){
	
	queueList := FileHandler.Read(&NumOfElevs, &NumOfFloors)
	
	for j:=0;j<len(queueList);j+=2{
		q := FileHandler.Directions{queueList[j], queueList[j+1]}
		queue = append(queue, q)
		//internalOrders = append(internalOrders,q)
	}
	//AddQueue(queue, ourID)
	ordersInDirection[UP] = 0
	ordersInDirection[DOWN] = 0
	for i,_ := range queue {
		ordersInDirection[UP] += queue[i].UP
		ordersInDirection[DOWN] = queue[i].DOWN
	}
	
	/*
	for i:=0;i<NumOfElevs;i++ {
		numInQueue = append(numInQueue,0}
	}
	*/
	addQueue(1, queue)
	addQueue(0, queue)
}

func addQueue(elevatorID int, queue []FileHandler.Directions) {
	allQueues[elevatorID] = queue
}



func Queue(addOrderChannel chan driver.ButtonMessage, removeOrderChannel chan int, setDirectionChannel chan int, checkDirectionChannel chan int, checkOrdersChannel chan int, stop chan int){//, findBestElevator chan ButtonMessage ){
	turn := false
	direction := 0
	currentFloor := 0
	queueInit()
	
	for{
	/*println("\t\t\t\tORDERS DOWN ", ordersInDirection[DOWN] ,"\n\n\n")
	println("\t\t\t\tORDERS UP", ordersInDirection[UP] ,"\n\n\n")
	println("\t\t\t\tTURN", turn ,"\n\n\n")*/
		select{
			case newOrder := <- addOrderChannel:
				addOrder(1, newOrder, currentFloor, direction)

			case <- removeOrderChannel: 
				println("QUEUE: REMOVEORDER")
				removeOrder(1, currentFloor, direction, &turn)
				
				
			case movingDirection := <- checkDirectionChannel:
				direction = movingDirection
				println("QUEUE: CHECKDIRECTION", direction)
				/*if(checkIfOrdersInDirection(1, currentFloor, direction, &turn) != -1){
					setDirectionChannel <- direction
					break
				} else{*/
					newDirection := setDirection(1, currentFloor, &turn)
					direction = newDirection
					setDirectionChannel <- newDirection
					break
				//}

			case floor := <- checkOrdersChannel:
				println("QUEUE: CHECKORDER", direction)
				currentFloor = floor
				if(checkOrders(1, currentFloor, direction, &turn)){
					if(turn){
						stop <- int(math.Abs(float64(1-direction)))
					}else{
						stop <- direction
					}
					break
				} 
				
			//case <- findBestElevatorChannel:

		}
	}
}

func setDirection(elevatorID int, currentFloor int, turn *bool) int{
    *turn = false
	if(ordersInDirection[UP] != 0){
		if(checkIfOrdersInDirection(elevatorID, currentFloor, UP, turn) != -1){
        	return UP
		} else{
			*turn = true
			return DOWN
		}
    }else if(ordersInDirection[DOWN] != 0){
		if(checkIfOrdersInDirection(elevatorID, currentFloor, DOWN, turn) != -1){
        	return DOWN
		}else{

			*turn = true
			return UP
		}
    }
	println("SETDIRECTION DONE: -1")
	return -1
}

func checkIfOrdersInDirection(elevatorID int, currentFloor int, direction int, turn *bool ) int{
	orderedFloor := -1
	//println("DIRECTION: ",direction)
	/*if(numberInQueue [elevatorID] == 0 || direction == -1){
			return orderedFloor
		}*/

	if(*turn){	
		if(direction == UP){
			/*if(currentFloor == NumOfFloors-1){
				return orderedFloor
			}*/
			
			for floor := currentFloor ;floor >= 0; floor--{
				if(allQueues[elevatorID][floor].UP == 1){
					orderedFloor = floor
				}
			}
		}else if (direction == DOWN){
			/*if(currentFloor == 0){
				return orderedFloor
			}*/
			
			for floor := currentFloor ;floor < NumOfFloors; floor++{
				if(allQueues[elevatorID][floor].DOWN == 1){
					orderedFloor = floor
				}
			}
			println("ORDERESFLOOR: ",orderedFloor)
			return orderedFloor
		} 
	} else if (!*turn){
		if(direction == UP){
			if(currentFloor == NumOfFloors-1){
				return orderedFloor
			}
			for floor := currentFloor ;floor < NumOfFloors; floor++{
				if(allQueues[elevatorID][floor].UP == 1){
					orderedFloor = floor
				}
			}
		} else if (direction == DOWN){
			if(currentFloor == 0){
				return orderedFloor
			}
			for floor := currentFloor ;floor >= 0; floor--{
				if(allQueues[elevatorID][floor].DOWN == 1){
					orderedFloor = floor
				}
			}
		}
	}
	println("ORDERESFLOOR: ",orderedFloor)
	return orderedFloor
}

func checkOrders(elevatorID int, currentFloor int, direction int, turn *bool) bool{
	if (direction == UP) {	
		if(allQueues[elevatorID][currentFloor].UP == 1 || (*turn && (checkIfOrdersInDirection(elevatorID, currentFloor, DOWN, turn) == currentFloor))){
			//println(allQueues[elevatorID][currentFloor].UP)
			return true
		}
	} else if (direction == DOWN) {
		if(allQueues[elevatorID][currentFloor].DOWN == 1 || (*turn && checkIfOrdersInDirection(elevatorID, currentFloor, UP, turn) == currentFloor)){
			//println(allQueues[elevatorID][currentFloor].DOWN)
			return true
		}
	}
	return false
}



func addOrder(elevatorID int, order driver.ButtonMessage, currentFloor int, movingDirection int) {
	//Antar at vi vet hvilken heis vi skal bruk
	var directionOfOrder = -1
	
	
	if ( currentFloor - order.Floor > 0 ) {
        directionOfOrder = DOWN
    } else if ( currentFloor - order.Floor < 0 ) {
        directionOfOrder = UP
    } else if ( currentFloor - order.Floor == 0) {
		if movingDirection == -1 {
			directionOfOrder = UP
		}
		directionOfOrder = int(math.Abs(float64(1-movingDirection)))
	}
	
	
    if( order.Button == driver.BUTTON_CALL_UP && allQueues[elevatorID][order.Floor].UP == 0 ) {
        allQueues[elevatorID][order.Floor].UP = 1
        numberInQueue[elevatorID] ++
        ordersInDirection[UP]++
		println("ORDERED ADDED")
    } else if( order.Button == driver.BUTTON_CALL_DOWN && allQueues[elevatorID][order.Floor].DOWN==0 ) {
        allQueues[elevatorID][order.Floor].DOWN = 1
        numberInQueue[elevatorID] ++
        ordersInDirection[DOWN] ++
    } else if(order.Button == driver.BUTTON_COMMAND) {
    	
    	if directionOfOrder == UP {
			if( allQueues[elevatorID][order.Floor].UP == 0){            	
				numberInQueue[elevatorID] ++
            	ordersInDirection[directionOfOrder] ++
            	allQueues[elevatorID][order.Floor].UP = 1
			}	
		} else if directionOfOrder == DOWN {
			if( allQueues[elevatorID][order.Floor].DOWN == 0){
            	numberInQueue[elevatorID] ++
            	ordersInDirection[directionOfOrder] ++
            	allQueues[elevatorID][order.Floor].DOWN = 1
			}
		}	
   	}
	
}
	
func removeOrder(elevatorID int, floor int, movingDirection int, turn *bool) {
	numberInQueue[elevatorID]--
	if ((movingDirection == UP && !(*turn)) || ((movingDirection == DOWN) && *turn)) {
		allQueues[elevatorID][floor].UP = 0
		ordersInDirection[UP]--
		//println("Sletter ordre opp") 
	}else if ((movingDirection == DOWN) && !(*turn) || ((movingDirection == UP) && *turn)) {
		allQueues[elevatorID][floor].DOWN = 0
		ordersInDirection[DOWN]--
		//println("Sletter ordre ned")  
	} //else error 
}
	

func findBestElevator(elevatorID int, order driver.ButtonMessage) int{
	min := 0
	for elev:= 0; elev < NumOfElevs; elev++{
			if(numberInQueue [elev] <= min){
				min = elev
			}
	}
	return min
}


/*func ClearAllExternalOrders(elevatorID int) {
	numberInQueue [elevatorID] = 0
	for floor := 0; floor < NumOfFloors; floor++ {
		if(internalOrders[floor][UP]){
			numberInQueue [elevatorID]++
		}
		if(internalOrders[floor][DOWN]){
			numberInQueue [elevatorID]++
		}
	}
	queue = internalOrders
}
*/


