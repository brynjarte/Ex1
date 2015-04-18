package source

type ButtonMessage struct {
	Floor int
	Button int
	Light int
}

type Message struct{
	NewOrder bool
	FromMaster bool
	CompletedOrder bool
	UpdatedElevInfo bool
	MessageTo int

	ElevInfo Elevator
	Button ButtonMessage
}

type Elevator struct {
	ID int	
	CurrentFloor int
	Direction int
	//numberInQueue int
}