package Source

import "driver"

/*
type ButtonMessage struct {
	Floor int
	Button int
	Light int
}*/

type Message struct{
	NewOrder bool
	FromMaster bool
	CompletedOrder bool
	UpdatedElevInfo bool
	MessageTo int

	ElevInfo Elevator
	Button driver.ButtonMessage
}

type Elevator struct {
	ID int	
	CurrentFloor int
	Direction int
	//numberInQueue int
}
