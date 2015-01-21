package main
import (
	. "fmt" // Using '.' to avoid prefixing functions with their package names
	// This is probably not a good idea for large projects...
	"runtime"
	"time"
)

var i int = 0;

func thread_1Func() {
	for j := 0; j < 1000000; j++ {
		i += 1	
	}
}

func thread_2Func() {
	for j := 0; j < 1000000; j++ {
		i -= 1	
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU()) 	
						
	go thread_1Func()
	time.Sleep(100*time.Millisecond)
	go thread_2Func()			
	time.Sleep(100*time.Millisecond)
	
	
	Println(i)
}
