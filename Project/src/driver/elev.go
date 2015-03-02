package driver

const (
	DIRN_DOWN = -1
	DIRN_STOP = 0
	DIRN_UP = 1

	N_FLOORS = 4
	N_BUTTONS = 3

	BUTTON_CALL_UP = 0
	BUTTON_CALL_DOWN = 1
	BUTTON_COMMAND = 2

	ERROR = 20132
)


var lamp_channel_matrix = [N_FLOORS][N_BUTTONS] int {
	{LIGHT_UP1, LIGHT_DOWN1, LIGHT_COMMAND1},
	{LIGHT_UP2, LIGHT_DOWN2, LIGHT_COMMAND2},
	{LIGHT_UP3, LIGHT_DOWN3, LIGHT_COMMAND3},
	{LIGHT_UP4, LIGHT_DOWN4, LIGHT_COMMAND4},
}

var button_channel_matrix = [N_FLOORS][N_BUTTONS] int {
	{BUTTON_UP1, BUTTON_DOWN1, BUTTON_COMMAND1},
	{BUTTON_UP2, BUTTON_DOWN2, BUTTON_COMMAND2},
	{BUTTON_UP3, BUTTON_DOWN3, BUTTON_COMMAND3},
	{BUTTON_UP4, BUTTON_DOWN4, BUTTON_COMMAND4},
}

type ButtonMessage struct {
	Floor int
	Button int
}	

func elev_init() int{
	

	if (io_init() == 0) {
		return 0
	}
	
	for i := 0; i < N_FLOORS; i++ {
		if i != 0 {
			elev_set_button_lamp(BUTTON_CALL_DOWN, i, 0)
		}
		if i != (N_FLOORS - 1) {
			elev_set_button_lamp(BUTTON_CALL_UP, i, 0)
		}

		elev_set_button_lamp(BUTTON_COMMAND, i, 0)
	}

	elev_set_stop_lamp(0)
	elev_set_door_open_lamp(0)
	elev_set_floor_indicator(0)

	//NewtorkChannel := make(chan int, 1)
	//ReadSensorsChannel := make(chan int, 1)
	

	return 1
}

func elev_set_motor_direction(dirn int) {
	if dirn == 0 {
		io_write_analog(MOTOR, 0)
	} else if dirn > 0 {
		io_clear_bit(MOTORDIR)
		io_write_analog(MOTOR, 2800)
	} else if (dirn < 0) {
		io_set_bit(MOTORDIR)
		io_write_analog(MOTOR, 2800)
	}
} 

func elev_set_door_open_lamp(value int) {
	if value != 0 {
		io_set_bit(LIGHT_DOOR_OPEN)
	} else {
		io_clear_bit(LIGHT_DOOR_OPEN)
	}
}

func elev_get_obstruction_signal() int {
	return io_read_bit(OBSTRUCTION)
}

func elev_get_stop_signal() int {
	return io_read_bit(STOP)
}

func elev_set_stop_lamp(value int) {
	if value != 0 {
		io_set_bit(LIGHT_STOP)
	} else {
		io_clear_bit(LIGHT_STOP)
	}
}

func elev_set_floor_indicator(floor int) int {
	if (floor < 0 || floor >= N_FLOORS) {
		return ERROR;
	}

	if (floor & 0x02) != 0 {
		io_set_bit(LIGHT_FLOOR_IND1)
	} else {
		io_clear_bit(LIGHT_FLOOR_IND1)
	}

	if (floor & 0x01) != 0 {
		io_set_bit(LIGHT_FLOOR_IND2)
	} else {
		io_clear_bit(LIGHT_FLOOR_IND2)
	}
	
	return 0
}


func elev_set_button_lamp(button int, floor int, value int) int{

	if(floor < 0){
		return ERROR
	}
	if(floor >= N_FLOORS){
		return ERROR
	}
	if((button == BUTTON_CALL_UP) && (floor == N_FLOORS -1)){
		return ERROR
	}
	if((button == BUTTON_CALL_DOWN) && (floor == 0)){
		return ERROR
	}
	if((button != BUTTON_CALL_UP) && (button != BUTTON_CALL_DOWN) && (button != BUTTON_COMMAND)){
		return ERROR
	}

	if(value != 0){
		io_set_bit(lamp_channel_matrix[floor][button])
	} else {
		io_clear_bit(lamp_channel_matrix[floor][button])
	}

	return 0
}


func elev_get_button_signal(button int, floor int) int{

	if(floor < 0){
		return ERROR
	}
	if(floor >= N_FLOORS){
		return ERROR
	}
	if((button == BUTTON_CALL_UP) && (floor == N_FLOORS -1)){
		return ERROR
	}
	if((button == BUTTON_CALL_DOWN) && (floor == 0)){
		return ERROR
	}
	if((button != BUTTON_CALL_UP) && (button != BUTTON_CALL_DOWN) && (button != BUTTON_COMMAND)){
		return ERROR
	}

	if(io_read_bit(button_channel_matrix[floor][button]) != 0){
		return 1
	} else {
		return 0
	}
}

func readButtons(ReadButtonsChannel chan ButtonMessage) { 
	var buttonPressed ButtonMessage
	buttonPressed.Floor = -1
	for{    	
		for  i := 0; i < 3; i++  {
   
			if ( elev_get_button_signal( BUTTON_CALL_UP, i ) == 1) {
				buttonPressed.Floor =  i
				buttonPressed.Button = BUTTON_CALL_UP
			} else if ( elev_get_button_signal( BUTTON_CALL_DOWN, i+1 ) == 1) {
				buttonPressed.Floor =  i+1
				buttonPressed.Button = BUTTON_CALL_DOWN
			} 
		} 
    
		for i := 0; i < 4; i++ {
        
			if ( elev_get_button_signal( BUTTON_COMMAND, i ) == 1 ) {
				buttonPressed.Floor =  i
				buttonPressed.Button = BUTTON_COMMAND
			}
		}
	
		if (buttonPressed.Floor != -1) {

			ReadButtonsChannel<- buttonPressed
			buttonPressed.Floor = -1
		}
	}
}
	
func readSensors(sensorChannel chan int){
	for{
		if (io_read_bit(SENSOR_FLOOR1) != 0) {
			sensorChannel <- 0
		} else if (io_read_bit(SENSOR_FLOOR2) != 0) {
			sensorChannel <- 1
		} else if (io_read_bit(SENSOR_FLOOR3) != 0) {
			sensorChannel <- 2
		} else if (io_read_bit(SENSOR_FLOOR4) != 0) {
			sensorChannel <- 3
		} 
	}
}
/*
func Elevator(sensorChannel chan int, readButtonsChannel chan ButtonMessage, recChannel chan UDPMessage,sensorChannel chan int){
	err := elev_init()
	if(err == 0){
		return
	}
	go UDP.recieveUdpMessage(recChannel)
	go readSensors(sensorChannel)
	go readButtons(readButtonsChannel)	
	for{
		select{
			case currentFloor := <- sensorChannel:
				elev_set_floor_indicator(currentFloor)
			case buttonPushed := readButtonsChannel:
				select{
					case sjekk ka heis så er best
				//LEGG TIL I KØ
				elev_set_button_lamp(buttonPushed.Button, buttonPushed.Floor,1)
			case msg := <- rec_channel:
				/*select{
					case sjekk ka heis så er best
				//LEGG TIL I KØ
		}
	}
}*/

				
