package main

import (
	"fmt"
	//"time"
	"net"
	//"strconv"
)



func server_TCP() (string){
	
	buffer := make([]byte, 1024)
	addr, _ := net.ResolveTCPAddr("tcp","129.241.187.136:34933")
	listner,erro := net.ListenTCP("tcp",addr)
		
	if erro != nil {
		fmt.Println(erro.Error())
	}		
	con,err   := listner.AcceptTCP()
	if err != nil {

		fmt.Println(err.Error())
	}
	rlen, _ := con.Read(buffer)
	con.Close()
	return string(buffer[0:rlen])
	
}

/*func client_TCP() (string){

	addr, _ := net.ResolveTCPAddr("tcp",":34933")
	conn,_  := net.DialTCP("tcp",nil,addr)
	conn.Write([]byte("Halo"))
	
	reply:= make([]byte,1024)
	conn.Read(reply)


	conn.Close()
	return string(reply[0:1023])
}
*/

func main (){

	//fmt.Println(server_TCP())
	//fmt.Println(client_TCP())
	
	
	
}
