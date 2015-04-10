package FileHandler

import ( 
	//"io"
	"os"
	"fmt"
	//"encoding/binary"
	//"bytes"
	"Queue"
)


func Read(NumOfElevs int, NumOfFloors int) {//([][]bool) {
	
	fd, err := os.Open("maxmekker.txt")
	if err != nil {
		panic(err)
	}

	var buf int

	_, err = fmt.Fscanf(fd, "%d\n", &buf)

	NumOfElevs = buf

	_, err = fmt.Fscanf(fd, "%d\n", &buf)

	NumOfFloors = buf

	fmt.Println(NumOfElevs)
	fmt.Println(NumOfFloors)

	//var queue [NumOfFloors][2]bool
	
	for i:=0; i < NumOfFloors; i++{
		_, err = fmt.Fscanf(fd, "%d", &buf)
		Queue.AddOrderFromBackup(0,i,buf==1)
		_, err = fmt.Fscanf(fd, "%d\n", &buf)
		Queue.AddOrderFromBackup(1,i,buf==1)
		//fmt.Println(queue[i][0], queue[i][1])
	}
	for{}
}

func Write() {
	// open output file
    fo, err := os.Create("output.txt")
    if err != nil {
        panic(err)
    }
    // close fo on exit and check for its returned error
    defer func() {
        if err := fo.Close(); err != nil {
            panic(err)
        }
    }()

	/*for {
	        // write a chunk
        if _, err := fo.Write(buf[:n]); err != nil {
            panic(err)
        }
	}*/
}
