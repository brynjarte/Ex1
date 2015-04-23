package Queue

import (
	"Source"
	"math"
	"time"
	"FileHandler"
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

var numOrdersInDirection = make(map[int] [2]int)
var allExternalQueues  = make(map[int] []Source.ButtonMessage)
var allElevatorsInfo = make(map[int] Source.ElevatorInfo)
var queue = linkedList{nil,nil,0}


func queueInit(elevator Source.ElevatorInfo){
	orderInEmptyQueue := make(chan int, 1)
	fetchMyQueue(elevator.ID, elevator.CurrentFloor, elevator.Direction, orderInEmptyQueue)

	updateElevInfo(elevator)
}


func Queue(elevatorInfo Source.ElevatorInfo, addOrderChannel chan Source.ButtonMessage, removeOrderChannel chan int, nextOrderChannel chan int, checkOrdersChannel chan int, orderInEmptyQueue chan int, finishedRemoving chan int, fromUdpToQueue chan Source.Message, bestElevatorChannel chan Source.Message, removeElevatorChannel chan int, completedOrderChannel chan Source.ButtonMessage, orderRemovedChannel chan Source.ButtonMessage, requestQueueChannel chan int, receiveQueueChannel chan Source.Message){

	direction := -1
	deletedOrderChannel := make(chan Source.ButtonMessage, 1)
	queueInit(elevatorInfo)

	for{
		select{
			case newOrder := <- addOrderChannel:
				go addOrder(elevatorInfo.ID, newOrder, elevatorInfo.CurrentFloor, elevatorInfo.Direction, orderInEmptyQueue)
			
			case <- removeOrderChannel: 
				go removeOrder(finishedRemoving, orderRemovedChannel, deletedOrderChannel)				
	
			case floor := <- checkOrdersChannel:
				
				//println("QUEUE: CHECKORDER DIRECTION:", elevatorInfo.Direction)
				elevatorInfo.CurrentFloor = floor
				nextOrderedFloor := checkOrders()
				nextOrderChannel <- nextOrderedFloor
				//println("Next or floor", nextOrderedFloor)
				direction = nextOrderedFloor - elevatorInfo.CurrentFloor
				if(direction > 0){
					elevatorInfo.Direction = Source.UP
				}else if (direction < 0){
					elevatorInfo.Direction = Source.DOWN
				}
				go updateElevInfo(elevatorInfo)

			case newUpdate := <- fromUdpToQueue:
				println("NEW INFO", newUpdate.UpdatedElevInfo)
				if(newUpdate.FromMaster){	
					if(newUpdate.NewOrder && newUpdate.MessageTo == elevatorInfo.ID){
						addOrderChannel <- newUpdate.Button
						go recieveExternalQueue(newUpdate.MessageTo, newUpdate.Button)
					/*} else if (newUpdate.NewOrder && newUpdate.MessageTo != elevatorInfo.ID && newUpdate.AcceptedOrder){
						go recieveExternalQueue(newUpdate.MessageTo, newUpdate.Button)
					*/}else if (!newUpdate.NewOrder && newUpdate.AcceptedOrder) {
						go recieveExternalQueue(newUpdate.MessageFrom, newUpdate.Button)
					} else if( newUpdate.CompletedOrder && newUpdate.MessageTo != elevatorInfo.ID){
						go recieveExternalQueue(newUpdate.MessageFrom, newUpdate.Button)
					} else if (newUpdate.UpdatedElevInfo && newUpdate.MessageTo != elevatorInfo.ID){
						go updateElevInfo(newUpdate.ElevInfo)
					} //else if(newUpdate.ExternalQueue != nil){
						//mergeExternalQueues(newUpdate.ExternalQueue)
					//}
				} else{
					if(newUpdate.NewOrder ){
						go findBestElevator(elevatorInfo.ID, newUpdate, bestElevatorChannel, addOrderChannel)
					} else if(newUpdate.AcceptedOrder){ 
						go recieveExternalQueue(newUpdate.MessageFrom, newUpdate.Button) 
					} else if(newUpdate.CompletedOrder){
						go recieveExternalQueue(newUpdate.MessageFrom, newUpdate.Button)
					} else if (newUpdate.UpdatedElevInfo){
						go updateElevInfo(newUpdate.ElevInfo)
						println("NEW INFO", newUpdate.ElevInfo.ID)
					}	
				
				}
			case lostElevator := <- removeElevatorChannel:
				go removeElevator(lostElevator, elevatorInfo, bestElevatorChannel, addOrderChannel)

			case <- requestQueueChannel:
				go getExternalQueues(elevatorInfo, requestQueueChannel, receiveQueueChannel)
		}
	}
}

func getExternalQueues(elevator Source.ElevatorInfo, requestQueueChannel chan int, receiveQueueChannel chan Source.Message) {
	queueMessage := Source.Message{false, true, true, false, false, elevator.ID, -1, elevator, Source.ButtonMessage{-1,-1,-1}}	
	for elev := range allExternalQueues {
		queueMessage.MessageTo = elev 
		for orders := 0; orders < len(allExternalQueues[elev]); orders++ {
			queueMessage.Button = allExternalQueues[elev][orders]
			receiveQueueChannel <- queueMessage
			
		}
	}
	queueMessage.MessageTo = -1
	receiveQueueChannel <- queueMessage
}

func removeElevator(lostElevator int, elevatorInfo Source.ElevatorInfo, bestElevatorChannel chan Source.Message, addOrderChannel chan Source.ButtonMessage){
	unDistributedOrder := Source.Message{true, false, false, false, false, -1, -1, elevatorInfo, Source.ButtonMessage{-1, -1, -1}}

	delete(numOrdersInDirection, lostElevator)
	delete(allElevatorsInfo, lostElevator)
	
	for orders := 0; orders < len(allExternalQueues[lostElevator]); orders++ {
		order := allExternalQueues[lostElevator][orders]
		unDistributedOrder.Button = order
		//orderRemovedChannel <- order KAN DETTA VER LURT
		go findBestElevator(lostElevator, unDistributedOrder, bestElevatorChannel, addOrderChannel)
		time.Sleep(50*time.Microsecond) // NØDVENDIG?
		
	}		
	
	delete(allExternalQueues, lostElevator) 
}


func checkOrders() int {
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
		saveAndSendQueue()
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
		saveAndSendQueue()
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
			saveAndSendQueue()
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
					saveAndSendQueue()
					return
				} else {
					nodePointer = (*nodePointer).next
				}
			 }
		}
		queue.last.next = &newOrder
		queue.last = &newOrder
		queue.length++
		saveAndSendQueue()
	}
}

func compareOrders(oldOrder Source.ButtonMessage, newOrder Source.ButtonMessage, currentFloor int, direction int) Source.ButtonMessage {
	
	if newOrder.Button == Source.BUTTON_COMMAND {
		if newOrder.Floor < currentFloor {
			//direction DOWN
			if (oldOrder.Floor >  newOrder.Floor && oldOrder.Button != Source.BUTTON_CALL_UP) {
				return oldOrder
			} else if (oldOrder.Floor <= newOrder.Floor || oldOrder.Button == Source.BUTTON_CALL_UP && direction != Source.UP) {
				return newOrder
			} 
		} else if newOrder.Floor > currentFloor {
			//direction UP
			if (oldOrder.Floor <  newOrder.Floor && oldOrder.Button != Source.BUTTON_CALL_DOWN) {
				return oldOrder
			} else if (oldOrder.Floor >= newOrder.Floor || oldOrder.Button == Source.BUTTON_CALL_DOWN && direction != Source.DOWN) {
				return newOrder
			}	
		} else if (newOrder.Floor == currentFloor) {
			if (direction == Source.UP) {
				if (oldOrder.Floor > newOrder.Floor) {
					return oldOrder
				} else if (oldOrder.Floor <= newOrder.Floor) {
					return newOrder
				}
			} else if (direction == Source.DOWN) {
				if (oldOrder.Floor < newOrder.Floor) {
					return oldOrder
				} else if (oldOrder.Floor >= newOrder.Floor) {
					return newOrder
				}
			}
		} 
	} else if newOrder.Button == Source.BUTTON_CALL_DOWN {
		if direction == Source.UP {
			if (oldOrder.Button == Source.BUTTON_CALL_DOWN && oldOrder.Floor < newOrder.Floor){
				return newOrder
			} else if (oldOrder.Button != Source.BUTTON_CALL_DOWN && oldOrder.Floor < currentFloor) {
				return newOrder
			} else {
				return oldOrder
			}
		} else if direction == Source.DOWN {
			if (oldOrder.Floor > newOrder.Floor || newOrder.Floor >= currentFloor) {
				return oldOrder
			} else {
				return newOrder
			}
		}
	} else if newOrder.Button== Source.BUTTON_CALL_UP {
		if direction == Source.DOWN {
			if (oldOrder.Button == Source.BUTTON_CALL_UP && oldOrder.Floor > newOrder.Floor) {
				return newOrder
			}  else if (oldOrder.Button != Source.BUTTON_CALL_UP && oldOrder.Floor > currentFloor) {			
				return newOrder
			} else {
				return oldOrder
			}
		} else if direction == Source.UP {
			if (oldOrder.Floor < newOrder.Floor  || newOrder.Floor <= currentFloor) {
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
				//deletedOrderChannel <- orderRemoved
				orderRemovedChannel <- orderRemoved
				finishedRemoving <- 1
				saveAndSendQueue()
				return
			}
		} else {
			orderRemoved := queue.head.value
			queue.head = nil
			queue.length = 0
			//deletedOrderChannel <- orderRemoved
			orderRemovedChannel <- orderRemoved
			finishedRemoving <- 1
			saveAndSendQueue()
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

	numUP := numOrdersInDirection[elevatorID][Source.UP]
	numDOWN := numOrdersInDirection[elevatorID][Source.DOWN]
	var temp [2] int
	temp[Source.UP] = numUP
	temp[Source.DOWN] = numDOWN 

	if (button.Value == 1 && button.Button != Source.BUTTON_COMMAND) {
		for order := 0; order < len(allExternalQueues[elevatorID]); order++ {
			if (allExternalQueues[elevatorID][order].Floor == button.Floor && allExternalQueues[elevatorID][order].Button == button.Button ) {
					return
			}
		} 
		if(button.Button == Source.UP){
			temp[Source.UP]++
			numOrdersInDirection[elevatorID] = temp
		} else if (button.Button == Source.DOWN) {
			temp[Source.DOWN]++
		}
		allExternalQueues[elevatorID] = append(allExternalQueues[elevatorID], button)
	} else if (button.Value == 0 && button.Button != Source.BUTTON_COMMAND) {
		for order := 0; order < len(allExternalQueues[elevatorID]); order++ {
			if (allExternalQueues[elevatorID][order].Floor == button.Floor && allExternalQueues[elevatorID][order].Button == button.Button) {
				allExternalQueues[elevatorID] = append(allExternalQueues[elevatorID][:order],allExternalQueues[elevatorID][order+1:]...)
				if(button.Button == Source.UP){
					temp[Source.UP]--
					numOrdersInDirection[elevatorID] = temp
				} else if (button.Button == Source.DOWN) {
					temp[Source.DOWN]--
				}
				return 			
			}
		}
	}	
}

func updateElevInfo(newElevInfo Source.ElevatorInfo){
	if(newElevInfo.CurrentFloor == Source.NumOfFloors-1){
		newElevInfo.Direction = Source.DOWN
	} else if(newElevInfo.CurrentFloor == 0){
		newElevInfo.Direction = Source.UP
	}
	allElevatorsInfo[newElevInfo.ID] = newElevInfo 
}

func findBestElevator(myElevatorID int, order Source.Message, bestElevatorChannel chan Source.Message, addOrderChannel chan Source.ButtonMessage){
	bestElevator := -1
	bestCost := 100
	calculateCost:
	for elevator := range allElevatorsInfo{
		println("Elevator nr. ", elevator)
		directionOfOrder := order.Button.Floor - allElevatorsInfo[elevator].CurrentFloor
		if(directionOfOrder > 0){
					directionOfOrder = Source.UP
				}else{
					directionOfOrder = Source.DOWN
				}

		if (numOrdersInDirection[elevator][Source.UP] == 0 && numOrdersInDirection[elevator][Source.DOWN] == 0) {
			bestElevator = elevator
			break calculateCost
		} else if(allElevatorsInfo[elevator].Direction == directionOfOrder && allElevatorsInfo[elevator].Direction == order.Button.Button){
			tempCost := int(math.Abs(float64(order.Button.Floor - allElevatorsInfo[elevator].CurrentFloor)))
			if(tempCost < bestCost){
				bestCost = tempCost
				bestElevator = elevator
			}
		} else if (allElevatorsInfo[elevator].Direction != directionOfOrder && allElevatorsInfo[elevator].Direction == order.Button.Button) {
			tempCost := numOrdersInDirection[elevator][Source.UP] + numOrdersInDirection[elevator][Source.DOWN]
			if(tempCost < bestCost){
				bestCost = tempCost
				bestElevator = elevator
			}
		} else if (allElevatorsInfo[elevator].Direction != directionOfOrder) || (allElevatorsInfo[elevator].Direction == directionOfOrder && allElevatorsInfo[elevator].Direction != order.Button.Button) {
			tempCost :=  numOrdersInDirection[elevator][allElevatorsInfo[elevator].Direction] + int(math.Abs(float64(order.Button.Floor - allElevatorsInfo[elevator].CurrentFloor)))
			if(tempCost < bestCost){
				bestCost = tempCost
				bestElevator = elevator
			}
		}
	}

	if(bestElevator != myElevatorID){	
		bestElevatorChannel <- Source.Message{true, false, true, false, false, myElevatorID, bestElevator, order.ElevInfo, order.Button}
	}else if (bestElevator == myElevatorID){
		bestElevatorChannel <- Source.Message{true, false, true, false, false, myElevatorID, bestElevator, order.ElevInfo, order.Button}
		go recieveExternalQueue(myElevatorID, order.Button)
		addOrderChannel <- order.Button
	}
		
		
}

func mergeExternalQueues(extQueue map[int][]Source.ButtonMessage) {
	
	for elev := 0; elev < Source.NumOfElevs; elev++ {
		temp := allExternalQueues[elev]
		if (temp == nil) {
			allExternalQueues[elev] = append(allExternalQueues[elev], extQueue[elev]...)
		}
	}
}


func fetchMyQueue(elevatorID int, currentFloor int, movingDirection int, orderInEmptyQueue chan int) {
	dummy := 0
	q := FileHandler.Read(&dummy, &dummy)
	
	clearAllOrders()
		
	for j:=0; j < len(q); j+=2 {
		order := Source.ButtonMessage{q[j],q[j+1],0}
		println("flr", q[j], "but", q[j+1])
		addOrder(elevatorID, order, currentFloor, movingDirection, orderInEmptyQueue)
	}
}



func saveAndSendQueue() {
	
	qList :=[]int(nil)
	
	if (queue.length != 0) {
		qList = append(qList, queue.head.value.Floor)
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
	}	
	FileHandler.Write(Source.NumOfElevs, Source.NumOfFloors, len(qList), qList)
	//UDP.sendQueue()
	
}
