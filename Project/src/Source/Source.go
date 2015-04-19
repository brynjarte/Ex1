package Source

const (
	BUTTON_CALL_UP = 0
	BUTTON_CALL_DOWN = 1
	BUTTON_COMMAND = 2
)
type ButtonMessage struct {
	Floor int
	Button int
	Value int
}

type Message struct{
	NewOrder bool
	FromMaster bool
	CompletedOrder bool
	UpdatedElevInfo bool
	MessageFrom int
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
