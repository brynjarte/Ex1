package Queue

import (
	"Source"
	"math"
	"time"
	"FileHandler"
	"fmt"
	"strconv"
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
var allExternalQueues  = make(map[string] []Source.ButtonMessage)
var allElevatorsInfo = make(map[int] Source.ElevatorInfo)
var lastUpdatedElevatorsInfo = make(map[int] time.Time) 
var queue = linkedList{nil,nil,0}
var master bool

func queueInit(elevator Source.ElevatorInfo, removeElevatorChannel chan int){
	updateElevInfo(elevator.ID, elevator, removeElevatorChannel)
}


func Queue(elevatorInfo Source.ElevatorInfo, addOrderChannel chan Source.ButtonMessage, removeOrderChannel chan int, nextOrderChannel chan int, checkOrdersChannel chan int, orderInEmptyQueue chan int, finishedRemoving chan int, networkMessageChannel chan Source.Message, bestElevatorChannel chan Source.Message, removeElevatorChannel chan int, completedOrderChannel chan Source.ButtonMessage, orderRemovedChannel chan Source.ButtonMessage, requestQueueChannel chan int, receiveQueueChannel chan Source.Message){
	
	direction := -1
	queueInit(elevatorInfo, removeElevatorChannel)

	for{
		select{
			case newOrder := <- addOrderChannel:
				go addOrder(elevatorInfo, newOrder, orderInEmptyQueue)
			
			case <- removeOrderChannel: 
				go removeOrder(finishedRemoving, orderRemovedChannel)			
	
			case floor := <- checkOrdersChannel:
				
				elevatorInfo.CurrentFloor = floor
				nextOrderedFloor := checkOrders()
				nextOrderChannel <- nextOrderedFloor
				direction = nextOrderedFloor - elevatorInfo.CurrentFloor
				if(direction > 0 || elevatorInfo.CurrentFloor == 0){
					elevatorInfo.Direction = Source.UP
				}else if (direction < 0 || elevatorInfo.CurrentFloor == Source.NumOfFloors-1){
					elevatorInfo.Direction = Source.DOWN
				}
				go updateElevInfo(elevatorInfo.ID, elevatorInfo, removeElevatorChannel)

			case newUpdate := <- networkMessageChannel:
				if (newUpdate.AllExternalOrders != nil) {
					go mergeExternalQueues(newUpdate.AllExternalOrders)
				}
				if(newUpdate.FromMaster){	
					master = false
					if(newUpdate.NewOrder && newUpdate.MessageTo == elevatorInfo.ID){
						addOrderChannel <- newUpdate.Button
						go receiveExternalQueue(elevatorInfo.ID, newUpdate.Button)
					} else if (newUpdate.AcceptedOrder) {
						go receiveExternalQueue(newUpdate.MessageFrom, newUpdate.Button)
					} else if( newUpdate.CompletedOrder && newUpdate.MessageTo != elevatorInfo.ID){
						go receiveExternalQueue(newUpdate.MessageFrom, newUpdate.Button)
					} else if (newUpdate.UpdatedElevInfo){
						go updateElevInfo(elevatorInfo.ID, newUpdate.ElevInfo, removeElevatorChannel)
					} 
				} else{
					master = true
					if(newUpdate.NewOrder){
						go findBestElevator(elevatorInfo.ID, newUpdate, bestElevatorChannel, addOrderChannel)
					} else if(newUpdate.AcceptedOrder || newUpdate.CompletedOrder){ 
						go receiveExternalQueue(newUpdate.MessageFrom, newUpdate.Button) 
					} else if (newUpdate.UpdatedElevInfo){
						go updateElevInfo(elevatorInfo.ID, newUpdate.ElevInfo, removeElevatorChannel)
					} else if(newUpdate.AllExternalOrders != nil){
						mergeExternalQueues(newUpdate.AllExternalOrders)
					}	
				}

			case disconnectedElevator := <- removeElevatorChannel:
				go removeElevator(disconnectedElevator, elevatorInfo, bestElevatorChannel, addOrderChannel)

			case <- requestQueueChannel:
				queueMessage := Source.Message{false, false, false, false, false, elevatorInfo.ID, -1, elevatorInfo, Source.ButtonMessage{-1,-1,-1}, allExternalQueues}	
				receiveQueueChannel <- queueMessage
		}
	}
}


func removeElevator(disconnectedElevator int, elevatorInfo Source.ElevatorInfo, bestElevatorChannel chan Source.Message, addOrderChannel chan Source.ButtonMessage){
	unDistributedOrder := Source.Message{true, false, false, false, false, -1, -1, elevatorInfo, Source.ButtonMessage{-1, -1, -1}, nil}

	temp := make(map [string][]Source.ButtonMessage)
	for elev,orders := range allExternalQueues {
		temp[elev] = orders
	}  

	if (disconnectedElevator == -1){
		master = true
		for elev := range allElevatorsInfo {
			if(elev != elevatorInfo.ID){			
				delete(numOrdersInDirection, elev)
				delete(allElevatorsInfo, elev)
				delete(allExternalQueues, fmt.Sprint(elev))
             }
        }
        for elev := range temp {
			if(elev != fmt.Sprint(elevatorInfo.ID)){
				for orders := 0; orders < len(temp[elev]); orders++ {
					order := temp[elev][orders]
					unDistributedOrder.Button = order
                    elevID, _ := strconv.Atoi(elev)
					go findBestElevator(elevID, unDistributedOrder, bestElevatorChannel, addOrderChannel)
					time.Sleep(50*time.Microsecond)
				}
			}
		}
		return
	}

	delete(numOrdersInDirection, disconnectedElevator)
	delete(allElevatorsInfo, disconnectedElevator)
	delete(allExternalQueues, fmt.Sprint(disconnectedElevator)) 
	for orders := 0; orders < len(temp[fmt.Sprint(disconnectedElevator)]); orders++ {
		order := temp[fmt.Sprint(disconnectedElevator)][orders]
		unDistributedOrder.Button = order
		go findBestElevator(disconnectedElevator, unDistributedOrder, bestElevatorChannel, addOrderChannel)
		time.Sleep(50*time.Microsecond)
	}		
}

func checkOrders() int {
	if queue.head == nil {
		return -1	
	} else {
		return queue.head.value.Floor
	}
}

func addOrder(elevator Source.ElevatorInfo, order Source.ButtonMessage, orderInEmptyQueue chan int) {
	currentFloor := elevator.CurrentFloor
	movingDirection := elevator.Direction

	var newOrder = node{order, nil}
	
	if (queue.length == 0) {
		queue.head = &newOrder
		queue.last = &newOrder
		queue.length = 1
		orderInEmptyQueue <- 1
		go saveQueue()
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
		go saveQueue()
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
			go saveQueue()
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
					go saveQueue()
					return
				} else {
					nodePointer = (*nodePointer).next
				}
			 }
		}
		queue.last.next = &newOrder
		queue.last = &newOrder
		queue.length++
		go saveQueue()
	}
}

func compareOrders(oldOrder Source.ButtonMessage, newOrder Source.ButtonMessage, currentFloor int, direction int) Source.ButtonMessage {
	
	if newOrder.Button == Source.BUTTON_COMMAND {
		if newOrder.Floor < currentFloor {
			if (oldOrder.Floor >  newOrder.Floor && oldOrder.Button != Source.BUTTON_CALL_UP) {
				return oldOrder
			} else if (oldOrder.Floor <= newOrder.Floor || oldOrder.Button == Source.BUTTON_CALL_UP && direction != Source.UP) {
				return newOrder
			} 
		} else if newOrder.Floor > currentFloor {
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

func removeOrder(finishedRemoving chan int, orderRemovedChannel chan Source.ButtonMessage) {
	
	for {
		if (queue.length > 1) {
			if (queue.head.value.Floor == queue.head.next.value.Floor) {
				orderRemoved := queue.head.value
				queue.head = queue.head.next
				queue.length--
				orderRemovedChannel <- orderRemoved
			} else {
				orderRemoved := queue.head.value
				queue.head = queue.head.next
				queue.length--
				orderRemovedChannel <- orderRemoved
				finishedRemoving <- 1
				go saveQueue()
				return
			}
		} else {
			orderRemoved := queue.head.value
			queue.head = nil
			queue.length = 0
			orderRemovedChannel <- orderRemoved
			finishedRemoving <- 1
			go saveQueue()
			return
		}
	}
}

func receiveExternalQueue(elevatorID int, button Source.ButtonMessage) {
	numUP := numOrdersInDirection[elevatorID][Source.UP]
	numDOWN := numOrdersInDirection[elevatorID][Source.DOWN]
	var temp [2] int
	temp[Source.UP] = numUP
	temp[Source.DOWN] = numDOWN 

	if (button.Value == 1 && button.Button != Source.BUTTON_COMMAND) {
		for order := 0; order < len(allExternalQueues[fmt.Sprint(elevatorID)]); order++ {
			if (allExternalQueues[fmt.Sprint(elevatorID)][order].Floor == button.Floor && allExternalQueues[fmt.Sprint(elevatorID)][order].Button == button.Button ) {
				return
			}
		} 
		if(button.Button == Source.UP){
			temp[Source.UP]++
			numOrdersInDirection[elevatorID] = temp
		} else if (button.Button == Source.DOWN) {
			temp[Source.DOWN]++
			numOrdersInDirection[elevatorID] = temp
		}
		allExternalQueues[fmt.Sprint(elevatorID)] = append(allExternalQueues[fmt.Sprint(elevatorID)], button)
	} else if (button.Value == 0 && button.Button != Source.BUTTON_COMMAND) {
		for order := 0; order < len(allExternalQueues[fmt.Sprint(elevatorID)]); order++ {
			if (allExternalQueues[fmt.Sprint(elevatorID)][order].Floor == button.Floor && allExternalQueues[fmt.Sprint(elevatorID)][order].Button == button.Button) {
				allExternalQueues[fmt.Sprint(elevatorID)] = append(allExternalQueues[fmt.Sprint(elevatorID)][:order],allExternalQueues[fmt.Sprint(elevatorID)][order+1:]...)
				if(button.Button == Source.UP){
					temp[Source.UP]--
					numOrdersInDirection[elevatorID] = temp
				} else if (button.Button == Source.DOWN) {
					temp[Source.DOWN]--
					numOrdersInDirection[elevatorID] = temp
				}
				return 			
			}
		}
	}	
}

func updateElevInfo(myElevatorID int, newElevInfo Source.ElevatorInfo, removeElevatorChannel chan int){
	
	if(newElevInfo.CurrentFloor == Source.NumOfFloors-1){
		newElevInfo.Direction = Source.DOWN
	} else if(newElevInfo.CurrentFloor == 0){
		newElevInfo.Direction = Source.UP
	}
	lastUpdatedElevatorsInfo[newElevInfo.ID] = time.Now().Add(7*time.Second) //Checking for disconnected slaves
	allElevatorsInfo[newElevInfo.ID] = newElevInfo 
	for elevator := range lastUpdatedElevatorsInfo{
	 	if(time.Now().After(lastUpdatedElevatorsInfo[elevator]) && elevator != myElevatorID && master){
			removeElevatorChannel <- elevator
		}
	}
}

func findBestElevator(myElevatorID int, order Source.Message, bestElevatorChannel chan Source.Message, addOrderChannel chan Source.ButtonMessage){
	bestElevator := myElevatorID
	bestCost := 100

	for elevator := range allExternalQueues {
		for ord := range allExternalQueues[elevator] {
			if (equalOrders(allExternalQueues[elevator][ord], order.Button)) {
				return			
			} 
		}
	}
	calculateCost:
	for elevator := range allElevatorsInfo{
		directionOfOrder := order.Button.Floor - allElevatorsInfo[elevator].CurrentFloor
		if(directionOfOrder > 0){
					directionOfOrder = Source.UP
				}else{
					directionOfOrder = Source.DOWN
				}

		if (numOrdersInDirection[elevator][Source.UP] == 0 && numOrdersInDirection[elevator][Source.DOWN] == 0) {
			tempCost := int(math.Abs(float64(order.Button.Floor - allElevatorsInfo[elevator].CurrentFloor)))
			if(tempCost < bestCost){
				bestCost = tempCost
				bestElevator = elevator
			}			
			if (allElevatorsInfo[elevator].CurrentFloor == order.Button.Floor) {
				break calculateCost
			}
		} else if(allElevatorsInfo[elevator].Direction == directionOfOrder && allElevatorsInfo[elevator].Direction == order.Button.Button){
			tempCost := int(math.Abs(float64(order.Button.Floor - allElevatorsInfo[elevator].CurrentFloor))) + numOrdersInDirection[elevator][Source.UP] + numOrdersInDirection[elevator][Source.DOWN]
			if(tempCost < bestCost){
				bestCost = tempCost
				bestElevator = elevator
			}
		} else if (allElevatorsInfo[elevator].Direction != directionOfOrder && allElevatorsInfo[elevator].Direction == order.Button.Button) {
			tempCost := numOrdersInDirection[elevator][Source.UP] + numOrdersInDirection[elevator][Source.DOWN] + int(math.Abs(float64(order.Button.Floor - allElevatorsInfo[elevator].CurrentFloor)))
			if(tempCost < bestCost){
				bestCost = tempCost
				bestElevator = elevator
			}
		} else if (allElevatorsInfo[elevator].Direction != directionOfOrder) || (allElevatorsInfo[elevator].Direction == directionOfOrder && allElevatorsInfo[elevator].Direction != order.Button.Button) {
			tempCost :=  numOrdersInDirection[elevator][Source.UP] + numOrdersInDirection[elevator][Source.DOWN] + int(math.Abs(float64(order.Button.Floor - allElevatorsInfo[elevator].CurrentFloor)))
			if(tempCost < bestCost){
				bestCost = tempCost
				bestElevator = elevator
			}
		}
	}

	if(bestElevator != myElevatorID){	
		bestElevatorChannel <- Source.Message{true, false, true, false, false, myElevatorID, bestElevator, order.ElevInfo, order.Button, nil}
	}else if (bestElevator == myElevatorID){
		bestElevatorChannel <- Source.Message{true, false, true, false, false, myElevatorID, bestElevator, order.ElevInfo, order.Button, nil}
		go receiveExternalQueue(myElevatorID, order.Button)
		addOrderChannel <- order.Button
	}
}


func mergeExternalQueues(extQueue map[string][]Source.ButtonMessage) {
	
	for elev := range extQueue {
		temp := allExternalQueues[elev]
		if (temp == nil ) {
			for order := 0; order < len(extQueue[elev]); order++ {
				allExternalQueues[elev] = append(allExternalQueues[elev], extQueue[elev][order])
				elevID, err := strconv.Atoi(elev)
				Source.ErrorChannel <- err
				go receiveExternalQueue(elevID, extQueue[elev][order])
			}
		} else {
			for order := 0; order < len(extQueue[elev]); order++ {
				if (!orderInQueue(elev, extQueue[elev][order])) {
					allExternalQueues[elev] = append(allExternalQueues[elev], extQueue[elev][order])
					elevID, err := strconv.Atoi(elev)
					Source.ErrorChannel <- err
					go receiveExternalQueue(elevID, extQueue[elev][order])
				}
			}
		}
	}
}

func orderInQueue(elevatorID string, order Source.ButtonMessage) bool {

    for oldOrder := 0; oldOrder < len(allExternalQueues[elevatorID]); oldOrder++ {
        if equalOrders(allExternalQueues[elevatorID][oldOrder], order) {
            return true
        }
    }
    return false
}

func saveQueue() {
	
	qList :=[]int(nil)
	
	if (queue.length != 0) {
		if (queue.head.value.Button == Source.BUTTON_COMMAND) {
			qList = append(qList, queue.head.value.Floor)
			qList = append(qList, queue.head.value.Button)
		}	
		var newOrder *node
		newOrder = queue.head.next
		for i:=1 ; i < queue.length; i++ {
			if (newOrder.value.Button == Source.BUTTON_COMMAND) {
				qList = append(qList, newOrder.value.Floor)
				qList = append(qList, newOrder.value.Button)
				newOrder = newOrder.next
			}
		}
	}	
	FileHandler.Write(Source.NumOfElevs, Source.NumOfFloors, len(qList)/2, qList)
}
