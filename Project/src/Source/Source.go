package Source

import( 
	"FileHandler"
)

const (
	BUTTON_CALL_UP = 0
	BUTTON_CALL_DOWN = 1
	BUTTON_COMMAND = 2
	
	UP = 0
	DOWN = 1
)

var NumOfFloors int
var NumOfElevs int

type ButtonMessage struct {
	Floor int
	Button int
	Value int
}

type Message struct{
	NewOrder bool
	AcceptedOrder bool
	FromMaster bool
	CompletedOrder bool
	UpdatedElevInfo bool
	MessageFrom int
	MessageTo int

	ElevInfo ElevatorInfo
	Button ButtonMessage
}

type ElevatorInfo struct {
	ID int	
	CurrentFloor int
	Direction int
	//numberInQueue int
}

func SourceInit() {
	FileHandler.Read(&NumOfElevs, &NumOfFloors)	
	println("NF", NumOfFloors,"\nNE",NumOfElevs)
}
