package main
import (
	. "fmt" // Using '.' to avoid prefixing functions with their package names
	// This is probably not a good idea for large projects...
	"runtime"

)

var i int = 0;


func thread_1Func(channel chan int, doneChannel chan int) {

	for j := 0; j < 1000000; j++ {
		<- channel
		i += 1	
		channel <-1
	}
	doneChannel <-1

}

func thread_2Func(channel chan int,doneChannel chan int) {

	for j := 0; j < 999999; j++ {
		<- channel
		i -= 1	
		channel <- 1
	}

	doneChannel <- 1
}


func main() {
	runtime.GOMAXPROCS(runtime.NumCPU()) 	
	
	
	someChannel := make(chan int,1)
	doneChannel := make(chan int)
	
	
	someChannel <-1	
	go thread_1Func(someChannel,doneChannel)
	go thread_2Func(someChannel,doneChannel)
				

	<- doneChannel
	<- doneChannel

	Println(i)
}
