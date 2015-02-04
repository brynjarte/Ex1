
package main


import(
	"fmt"
	"net"
	"time"
)

func Recieve(){
	buffer := make([]byte,1024) 
	raddr,_ := net.ResolveUDPAddr("udp", ":20004")
	recievesock,_ := net.ListenUDP("udp", raddr)
	for  {
	
	mlen ,_,_ := recievesock.ReadFromUDP(buffer)
	
	fmt.Println(string(buffer[:mlen]))
	}
}

func Send(){
	baddr,err := net.ResolveUDPAddr("udp", "129.241.187.255:20004")
	sendSock, err := net.DialUDP("udp", nil ,baddr) // connection
	send_msg := []byte("JS")
	time.Sleep(1*time.Second)
	_,err = sendSock.Write(send_msg)
	//fmt.Println(err)
	if err != nil{
		panic(err)
	}
	
}

func main(){
	
	go Recieve()
	time.Sleep(1*time.Second)
	Send()
	
	time.Sleep(100*time.Second)
	

		
}
