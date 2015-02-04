package main


import(
	"fmt"
	"net"
)

func Recieve(){
	buffer := make([]byte,1024) 
	raddr,_ := net.ResolveTCPAddr("tcp", ":34933")
	TCP_listener,_ := net.ListenTCP("tcp",raddr)
	TCP_socket,_ := TCP_listener.Accept()
	
	mlen ,_ := TCP_socket.Read(buffer)
	 
	fmt.Println(string(buffer[:mlen]))
}

func main(){
	buffer := make([]byte,1024) 
	raddr,_ := net.ResolveTCPAddr("tcp", ":34933")
	TCP_listener,_ := net.ListenTCP("tcp",raddr)
	TCP_socket,_ := TCP_listener.Accept()
	
	mlen ,_ := TCP_socket.Read(buffer)
	 
	fmt.Println(string(buffer[:mlen]))
	
		
}
