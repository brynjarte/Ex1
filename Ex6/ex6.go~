package main  

import (
	"os"
)

//type sessionp session struct{}

type session struct {
	source int
	target int
	in int  // Sequence number of most recently acked msg
	out int // Sequence number of most recently sent msg 
	ack int // Sequence number of most recent ack
}



type messageType struct {
	data int
	ack int
	takeOver int

}

type message struct {
	asession session
	MSG_TYPE messageType
	value int // DATA

}

var my_session session
//var in_msg_queue *sessionPointer

func session_init(){
	
my_session.source = os.Getpid()
my_session.target = 1
my_session.in = 0
my_session.ack = 0
my_session.out = 0
in_msg_queue = nil

}

func main() {



}
