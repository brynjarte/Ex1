package main

import (
	"fmt"
	//"time"
	"net"
	//"strconv"
)



func server_TCP() (string){
	
	buffer := make([]byte, 1024)
	//addr, _ := net.ResolveTCPAddr("tcp","129.241.187.136:34933")
	listner,erro := net.ListenTCP("tcp",:200+//LABPLASS)
		
	if erro != nil {
		fmt.Println(erro.Error())
	}		
	for{
		con,err   := listner.AcceptTCP()
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	rlen, _ := con.Read(buffer)
	fmt.Println("Got connection!")
	
	con.Close() // KANSKJE UT AV FOR- LØKKA?

	return string(buffer[0:rlen])
	
}

/*func client_TCP() (string){

	//addr, _ := net.ResolveTCPAddr("tcp","129.241.187.136:33546")
	//conn,_  := net.DialTCP("tcp",nil,addr)
	conn,error  := net.Dial("tcp","129.241.187.136:33546")
	if error != nil{
		fmt.Println("Connection error",error)
	}

	conn.Write([]byte("Halo"))
	reply:= make([]byte,1024)
	rlen,_ := conn.Read(reply)
	

	conn.Close()
	return string(reply[0:rlen])
}
*/

func main (){

	//fmt.Println(server_TCP())
	//fmt.Println(client_TCP())
	
	
	
}
