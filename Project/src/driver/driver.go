package driver

import (
    "math"
    "time"
	"Source"
)

const (
	DIRN_DOWN = -1
	DIRN_STOP = 0
	DIRN_UP = 1

	N_FLOORS = 4 // MÅ ENKELT KUNNA ENDRAST
	N_BUTTONS = 3 

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

func elev_init(sensorChannel chan int) int{

	if (io_init() == 0) {
		return 0
	}
	
	for i := 0; i < N_FLOORS; i++ {
		if i != 0 {
			elev_set_button_lamp(Source.ButtonMessage{i,Source.BUTTON_CALL_DOWN, 0})
		}
		if i != (N_FLOORS - 1) {
			elev_set_button_lamp(Source.ButtonMessage{i,Source.BUTTON_CALL_UP, 0})
		}

		elev_set_button_lamp(Source.ButtonMessage{i,Source.BUTTON_COMMAND, 0})
	}

	elev_set_door_open_lamp(0)
	elev_set_floor_indicator(0)
	select{
		case floor := <- sensorChannel:
				sensorChannel <- floor
				return 1
		case <- time.After(10*time.Millisecond):
				elev_set_speed(-300)
	}
	
	for{
		select{
			case floor := <- sensorChannel:
				sensorChannel <- floor
				elev_set_speed(219)
				<- time.After(7*time.Millisecond)
				elev_set_speed(0)
				return 1
		}
	}

	return 1
}

func elev_set_speed(speed int){

    // If to start (speed > 0)
    if (speed > 0){
        io_clear_bit(MOTORDIR)
    } else if (speed < 0){
        io_set_bit(MOTORDIR)
	}


    absSpeed := math.Abs(float64(speed))
    speed = int(absSpeed)
    // Write new setting to motor.
    io_write_analog(MOTOR, 2048 + 4*speed)
}

func elev_set_door_open_lamp(value int) {
	if value != 0 {
		io_set_bit(LIGHT_DOOR_OPEN)
	} else {
		io_clear_bit(LIGHT_DOOR_OPEN)
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


func elev_set_button_lamp(buttonPushed Source.ButtonMessage) int{

	if(buttonPushed.Floor < 0){
		return ERROR
	}
	if(buttonPushed.Floor >= N_FLOORS){
		return ERROR
	}
	if((buttonPushed.Button == Source.BUTTON_CALL_UP) && (buttonPushed.Floor == N_FLOORS -1)){
		return ERROR
	}
	if((buttonPushed.Button == Source.BUTTON_CALL_DOWN) && (buttonPushed.Floor == 0)){
		return ERROR
	}
	if((buttonPushed.Button != Source.BUTTON_CALL_UP) && (buttonPushed.Button != Source.BUTTON_CALL_DOWN) && (buttonPushed.Button != Source.BUTTON_COMMAND)){
		return ERROR
	}

	if(buttonPushed.Value != 0){
		io_set_bit(lamp_channel_matrix[buttonPushed.Floor][buttonPushed.Button])
	} else {
		io_clear_bit(lamp_channel_matrix[buttonPushed.Floor][buttonPushed.Button])
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
	if((button == Source.BUTTON_CALL_UP) && (floor == N_FLOORS -1)){
		return ERROR
	}
	if((button == Source.BUTTON_CALL_DOWN) && (floor == 0)){
		return ERROR
	}
	if((button != Source.BUTTON_CALL_UP) && (button != Source.BUTTON_CALL_DOWN) && (button != Source.BUTTON_COMMAND)){
		return ERROR
	}

	if(io_read_bit(button_channel_matrix[floor][button]) != 0){
		return 1
	} else {
		return 0
	}
}

func readButtons(NewOrderChannel chan Source.ButtonMessage) {
	var buttonPressed Source.ButtonMessage
	lastButtonPressed := Source.ButtonMessage{-1, -1, -1}
	for{   
		time.Sleep(50*time.Millisecond) 	
		buttonPressed.Floor = -1
		for  i := 0; i < 3; i++  {
   
			if ( elev_get_button_signal( Source.BUTTON_CALL_UP, i ) == 1) {
				buttonPressed.Floor =  i
				buttonPressed.Button = Source.BUTTON_CALL_UP	
			} else if ( elev_get_button_signal( Source.BUTTON_CALL_DOWN, i+1 ) == 1) {
				buttonPressed.Floor =  i+1
				buttonPressed.Button = Source.BUTTON_CALL_DOWN
			} 
		} 
    
		for i := 0; i < 4; i++ {
        
			if ( elev_get_button_signal( Source.BUTTON_COMMAND, i ) == 1 ) {
				for ; elev_get_button_signal( Source.BUTTON_COMMAND, i ) == 1 ; {
				}
				buttonPressed.Floor =  i
				buttonPressed.Button = Source.BUTTON_COMMAND
			}
		}
	
		if (buttonPressed.Floor != -1 && lastButtonPressed != buttonPressed) {
			lastButtonPressed = buttonPressed
			NewOrderChannel <- buttonPressed
		}
	}
}
	
func readSensors(sensorChannel chan int){
	lastFloor := -1
	for{
		time.Sleep(30*time.Microsecond)
		if (io_read_bit(SENSOR_FLOOR1) != 0 && lastFloor != 0 ) {
			lastFloor = 0
			sensorChannel <- lastFloor

		} else if (io_read_bit(SENSOR_FLOOR2) != 0 && lastFloor != 1 ) {
			lastFloor = 1
			sensorChannel <- lastFloor

		} else if (io_read_bit(SENSOR_FLOOR3) != 0 && lastFloor != 2 ) {
			lastFloor = 2
			sensorChannel <- lastFloor
			
		} else if (io_read_bit(SENSOR_FLOOR4) != 0 && lastFloor != 3 ) {
			lastFloor = 3
			sensorChannel <- lastFloor
		} 
	}
}

func clearExternalLights() {
		for i := 0; i < N_FLOORS; i++ {
			if i != 0 {
				elev_set_button_lamp(Source.ButtonMessage{i, Source.BUTTON_CALL_DOWN, 0})
			}
			if i != (N_FLOORS - 1) {
				elev_set_button_lamp(Source.ButtonMessage{i, Source.BUTTON_CALL_UP, 0})
			}
		}
}

func setExternalLights(externalOrders [][] bool, elevatorID int) {
		for floor := 0; floor < N_FLOORS; floor++ {
			if floor != 0 {
				if (externalOrders[floor][1+2*elevatorID]) {
					elev_set_button_lamp(Source.ButtonMessage{floor, Source.BUTTON_CALL_DOWN, 1})
				}
			}
			if floor != (N_FLOORS - 1) {
				if (externalOrders[floor][2*elevatorID]) {
					elev_set_button_lamp(Source.ButtonMessage{floor, Source.BUTTON_CALL_UP, 1})
				}
			}
		}
}


func stop(stoppedChannel chan int, direction int){

	if (direction == Source.UP) {
		elev_set_speed(-219)
	} else if (direction == Source.DOWN) {
		elev_set_speed(219)
	}
	<- time.After(7*time.Millisecond)
	elev_set_speed(0)
	elev_set_door_open_lamp(1)
		
	//println("SLEEEPING")	

	//FORSLAGcloseDOooR()
	// resetDoorChannel <- 1
		
	<- time.After(3*time.Second)
	elev_set_door_open_lamp(0)
	stoppedChannel <- 1


}

func closeDoor(stoppedChannel chan int, resetDoorChannel chan int){
	
	for{
		select{
			case <- time.After(3*time.Second):
				elev_set_door_open_lamp(0)
				stoppedChannel <- 1
				<- resetDoorChannel
			case <- resetDoorChannel:
				//Reset door timer
		}
	}
	return

}

func Drivers(newOrderChannel chan Source.ButtonMessage, floorReachedChannel chan int, setSpeedChannel chan int, stopChannel chan int, stoppedChannel chan int,setButtonLightChannel chan Source.ButtonMessage, initFinished chan int){

	sensorChannel := make(chan int, 1)
	go readSensors(sensorChannel)
	elev_init(sensorChannel)
	initFinished <- 1
	go readButtons(newOrderChannel)
	currentFloor := -1
	direction := -1
	
	for{
		select{
			case movingDirection := <- setSpeedChannel:
				println("DRIVER: SETSPEED" )
				direction = movingDirection
				setSpeed(direction)

			case <- stopChannel:
				println("DRIVER: STOP")
				go stop(stoppedChannel, direction)	
				
			case button := <- setButtonLightChannel:
				println("DRIVER: setbutton")
				go elev_set_button_lamp(button)

			case floor:= <- sensorChannel:
				println("DRIVER: SENSORCHANNEL")
				currentFloor = floor
				elev_set_floor_indicator(currentFloor)
				floorReachedChannel <- currentFloor
				
			default:
				time.Sleep(30*time.Microsecond)			
			}

		}
}

func setSpeed(direction int){
	
	if(direction == Source.UP){
		elev_set_speed(150)
	} else if(direction == Source.DOWN){
		elev_set_speed(-150)
	} 		
}



