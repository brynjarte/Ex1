package Queue

import (
	"Source"
	"math"
	"time"
	"FileHandler"
)


const (
	UP = 0
	DOWN = 1
)


type node struct{
	value Source.ButtonMessage
	next *node
}

type linkedList struct{
	head *node
	last *node
	length int
}

var allExternalQueues  = make(map[int] []Source.ButtonMessage)
var allElevatorsInfo = make(map[int] Source.Elevator)


var queue = linkedList{nil,nil,0}
var qList [] int


func queueInit(elevator Source.Elevator){
	orderInEmptyQueue := make(chan int, 1)
	fetchMyQueue(elevator.ID, elevator.CurrentFloor, elevator.Direction, orderInEmptyQueue)
	
	updateElevInfo(elevator)
}


func Queue(elevator Source.Elevator, addOrderChannel chan Source.ButtonMessage, removeOrderChannel chan int, nextOrderChannel chan int, checkOrdersChannel chan int, orderInEmptyQueue chan int, finishedRemoving chan int, fromUdpToQueue chan Source.Message, bestElevatorChannel chan Source.Message, removeElevatorChannel chan int, completedOrderChannel chan Source.ButtonMessage, orderRemovedChannel chan Source.ButtonMessage){

	direction := -1
	deletedOrderChannel := make(chan Source.ButtonMessage, 1)
	queueInit(elevator)

	for{
		select{
			case newOrder := <- addOrderChannel:
				go addOrder(elevator.ID, newOrder, elevator.CurrentFloor, elevator.Direction, orderInEmptyQueue)
			
			case <- removeOrderChannel: 
				println("QUEUE: REMOVEORDER")
				go removeOrder(finishedRemoving, orderRemovedChannel, deletedOrderChannel)				
	
			case floor := <- checkOrdersChannel:
				
				println("QUEUE: CHECKORDER DIRECTION:", elevator.Direction)
				elevator.CurrentFloor = floor
				nextOrderedFloor := checkOrders(1)
				nextOrderChannel <- nextOrderedFloor
				println("Next or floor", nextOrderedFloor)
				direction = nextOrderedFloor - elevator.CurrentFloor
				if(direction > 0){
					elevator.Direction = UP
				}else{
					elevator.Direction = DOWN
				}
				go updateElevInfo(elevator)

			case newUpdate := <- fromUdpToQueue:
				println("QUEUE: mesaage")
				if(newUpdate.FromMaster){		
					if(newUpdate.NewOrder && newUpdate.MessageTo == elevator.ID){
						addOrderChannel <- newUpdate.Button
						go recieveExternalQueue(newUpdate.MessageTo, newUpdate.Button)
					} else if (newUpdate.NewOrder && newUpdate.MessageTo != elevator.ID && newUpdate.AcceptedOrder){
						go recieveExternalQueue(newUpdate.MessageTo, newUpdate.Button)
					} else if( newUpdate.CompletedOrder && newUpdate.MessageTo != elevator.ID){
						go recieveExternalQueue(newUpdate.MessageTo, newUpdate.Button)
					} else if (newUpdate.UpdatedElevInfo && newUpdate.MessageTo != elevator.ID){
						go updateElevInfo(newUpdate.ElevInfo)
					}
				} else{
					if(newUpdate.NewOrder ){
						if(newUpdate.AcceptedOrder){ 
							go recieveExternalQueue(newUpdate.MessageTo, newUpdate.Button)
						} else{
							go findBestElevator(elevator.ID, newUpdate, bestElevatorChannel, addOrderChannel)
						}
					} else if(newUpdate.CompletedOrder){
						go recieveExternalQueue(newUpdate.MessageTo, newUpdate.Button)
					} else if (newUpdate.UpdatedElevInfo){
						go updateElevInfo(newUpdate.ElevInfo)
					}	
					 
				}
			case lostElevator := <- removeElevatorChannel:
				println("QUEUE: Removie")
				go removeElevator(lostElevator, elevator, bestElevatorChannel, addOrderChannel)

		}
	}
}


func removeElevator(lostElevator int, elevator Source.Elevator, bestElevatorChannel chan Source.Message, addOrderChannel chan Source.ButtonMessage){
	unDistributedOrder := Source.Message{true, false, false, false, false, -1, -1, elevator, Source.ButtonMessage{-1, -1, -1}}

	for orders := 0; orders < len(allExternalQueues[lostElevator]); orders++ {
		order := allExternalQueues[lostElevator][orders]
		unDistributedOrder.Button = order
		go findBestElevator(lostElevator, unDistributedOrder, bestElevatorChannel, addOrderChannel)
		time.Sleep(50*time.Microsecond) // NØDVENDIG?
	}		

	delete(allExternalQueues, lostElevator) 
	delete(allElevatorsInfo, lostElevator)
}


func checkOrders(elevatorID int) int {
	if queue.head == nil {
		return -1	
	}else {
		return queue.head.value.Floor
	}
}

func addOrder(elevatorID int, order Source.ButtonMessage, currentFloor int, movingDirection int, orderInEmptyQueue chan int) {

	var newOrder = node{order, nil}
	
	if (queue.length == 0) {
		queue.head = &newOrder
		queue.last = &newOrder
		queue.length = 1
		orderInEmptyQueue <- 1
		go saveAndSendQueue()
		return
	} else if (queue.length == 1) {
		if equalOrders(queue.head.value, order) {
			return
		} else {
			queue.length++
			if equalOrders(compareOrders(queue.head.value, order, currentFloor, movingDirection), order) {
				newOrder.next = queue.last
				queue.head = &newOrder
			} else {
				queue.head.next = &newOrder
				queue.last = &newOrder
			}
		go saveAndSendQueue()
		return
		}
	} else {

		var nodePointer *node = queue.head
		if equalOrders(nodePointer.value, order) {
			return
		} else if equalOrders(compareOrders((*nodePointer).value, order, currentFloor, movingDirection), order) {
			newOrder.next = queue.head
			queue.head = &newOrder
			queue.length++
			go saveAndSendQueue()
			return
		}
		for i:=0; i < queue.length-1; i++ {
			 if equalOrders((*nodePointer).next.value, order) {
			 	return
			 } else {
			 	if equalOrders(compareOrders((*nodePointer).next.value, order, currentFloor, movingDirection), order) {
					newOrder.next = (*nodePointer).next
					(*nodePointer).next = &newOrder
					queue.length++
					go saveAndSendQueue()
					return
				} else {
					nodePointer = (*nodePointer).next
				}
			 }
		}
		queue.last.next = &newOrder
		queue.last = &newOrder
		queue.length++
		go saveAndSendQueue()
	}
}

func compareOrders(oldOrder Source.ButtonMessage, newOrder Source.ButtonMessage, currentFloor int, direction int) Source.ButtonMessage {
	
	if newOrder.Button == Source.BUTTON_COMMAND {
		if newOrder.Floor < currentFloor {
			//direction DOWN
			if oldOrder.Floor >  newOrder.Floor {
				return oldOrder
			} else if oldOrder.Floor < newOrder.Floor {
				return newOrder
			} 
		} else if newOrder.Floor > currentFloor {
			//direction UP
			if oldOrder.Floor <  newOrder.Floor {
				return oldOrder
			} else if oldOrder.Floor > newOrder.Floor {
				return newOrder
			} else if newOrder.Floor == currentFloor {
				return oldOrder
			}	
		}
	} else if newOrder.Button == Source.BUTTON_CALL_DOWN {
		if direction == UP {
			if (oldOrder.Button == Source.BUTTON_CALL_DOWN && oldOrder.Floor < newOrder.Floor){
				return newOrder
			} else if (oldOrder.Button != Source.BUTTON_CALL_DOWN && oldOrder.Floor < currentFloor) {
				return newOrder
			} else {
				return oldOrder
			}
		} else if direction == DOWN {
			if (oldOrder.Floor > newOrder.Floor || newOrder.Floor > currentFloor) {
				return oldOrder
			} else {
				return newOrder
			}
		}
	} else if newOrder.Button== Source.BUTTON_CALL_UP {
		if direction == DOWN {
			if (oldOrder.Button == Source.BUTTON_CALL_UP && oldOrder.Floor > newOrder.Floor) {
				return newOrder
			}  else if (oldOrder.Button != Source.BUTTON_CALL_UP && oldOrder.Floor > currentFloor) {			
				return newOrder
			} else {
				return oldOrder
			}
		} else if direction == UP {
			if (oldOrder.Floor < newOrder.Floor  || newOrder.Floor < currentFloor) {
				return oldOrder
			} else {
				return newOrder
			}
		}
	}
	return oldOrder
}

func equalOrders(oldOrder Source.ButtonMessage, newOrder Source.ButtonMessage) bool {
	return (oldOrder.Floor == newOrder.Floor && oldOrder.Button== newOrder.Button)
}

func removeOrder(finishedRemoving chan int, orderRemovedChannel chan Source.ButtonMessage, deletedOrderChannel chan Source.ButtonMessage ) {
	for {
		if (queue.length > 1) {
			if (queue.head.value.Floor == queue.head.next.value.Floor) {
				orderRemoved := queue.head.value
				queue.head = queue.head.next
				queue.length--
				orderRemovedChannel <- orderRemoved
				//deletedOrderChannel <- orderRemoved
			} else {
				orderRemoved := queue.head.value
				queue.head = queue.head.next
				queue.length--
				println("Sletter for siste gang, ikke siste ordre")
				//deletedOrderChannel <- orderRemoved
				orderRemovedChannel <- orderRemoved
				finishedRemoving <- 1
				go saveAndSendQueue()
				return
			}
		} else {
			orderRemoved := queue.head.value
			queue.head = nil
			queue.length = 0
			println("Sletter siste ordre")
			//deletedOrderChannel <- orderRemoved
			orderRemovedChannel <- orderRemoved
			finishedRemoving <- 1
			go saveAndSendQueue()
			return
		}
	}
}

func clearAllOrders(){
	queue.head = nil
	queue.last = nil
	queue.length = 0
}

func PrintQueue() {
	//println("KØ: ", queue.length)
	if queue.length == 0 {
		return
	}
	println("Element 1:\nEtasje: ", queue.head.value.Floor, "\tKnapp: ", queue.head.value.Button,"\n")
	var newOrder *node
	newOrder = queue.head.next
	for i:=1 ; i < queue.length; i++ {
		println("Element", i+1,":\nEtasje: ", newOrder.value.Floor, "\tKnapp: ", newOrder.value.Button,"\n")
		newOrder = newOrder.next
	}
}

func recieveExternalQueue(elevatorID int, button Source.ButtonMessage) {
	
	if (button.Value == 1 && button.Button != Source.BUTTON_COMMAND) {
		allExternalQueues[elevatorID] = append(allExternalQueues[elevatorID], button)
	} else if (button.Value == 0 && button.Button != Source.BUTTON_COMMAND) {
		for i := 0; i < len(allExternalQueues[elevatorID]); i++ {
			if (allExternalQueues[elevatorID][i].Floor == button.Floor && allExternalQueues[elevatorID][i].Button == button.Button) {
				allExternalQueues[elevatorID] = append(allExternalQueues[elevatorID][:(i-1)],allExternalQueues[elevatorID][i:]...)  			
			}
		}
	}	
}

func updateElevInfo(newElevInfo Source.Elevator){
	allElevatorsInfo[newElevInfo.ID] = newElevInfo //ER DETTA RETT?
}

func findBestElevator(myElevatorID int, order Source.Message, bestElevatorChannel chan Source.Message, addOrderChannel chan Source.ButtonMessage){
	bestElevator := -1
	bestCost := 100

	for elevator := range allElevatorsInfo{
		directionOfOrder := order.Button.Floor - allElevatorsInfo[elevator].CurrentFloor
		if(directionOfOrder > 0){
					directionOfOrder = UP
				}else{
					directionOfOrder = DOWN
				}

		if(allElevatorsInfo[elevator].Direction == directionOfOrder){
			tempCost := int(math.Abs(float64(order.Button.Floor - allElevatorsInfo[elevator].CurrentFloor)))
			if(tempCost < bestCost){
				bestCost = tempCost
				bestElevator = elevator
			}
		} else {
			tempCost := 3 + int(math.Abs(float64(order.Button.Floor - allElevatorsInfo[elevator].CurrentFloor)))
			if(tempCost < bestCost){
				bestCost = tempCost
				bestElevator = elevator
			}
		}
	}
	if(bestElevator != myElevatorID){	
		bestElevatorChannel <- Source.Message{true, false, true, false, false, myElevatorID, bestElevator, order.ElevInfo, order.Button}
	}else{
		addOrderChannel <- order.Button
		bestElevatorChannel <- Source.Message{true, false, true, false, false, myElevatorID, bestElevator, order.ElevInfo, order.Button}
	}
		
		
}




func fetchMyQueue(elevatorID int, currentFloor int, movingDirection int, orderInEmptyQueue chan int) {
	q := FileHandler.Read(&NumOfElevs, &NumOfFloors)
	
	clearAllOrders()
		
	for j:=0; j < len(q); j+=2 {
		order := Source.ButtonMessage{q[j],q[j+1],0}
		println("flr", q[j], "but", q[j+1])
		addOrder(elevatorID, order, currentFloor, movingDirection, orderInEmptyQueue)
	}
}



func saveAndSendQueue() {
	
	qList := append([]int(nil), queue.head.value.Floor)
	qList = append(qList, queue.head.value.Button)
	
	var newOrder *node
	newOrder = queue.head.next
	for i:=1 ; i < queue.length; i++ {
		if (newOrder.value.Button == Source.BUTTON_COMMAND) {
			qList = append(qList, newOrder.value.Floor)
			qList = append(qList, newOrder.value.Button)
		}
		newOrder = newOrder.next
	}
	
	FileHandler.Write(NumOfElevs, NumOfFloors, len(qList), qList)
	//UDP.sendQueue()
	
}
