package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"fmt"
	"os"
)
import "strconv"

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.

type FileNotDoneYetArgs struct {
	Id int
}

type FileNotDoneYetReply struct {
	Name string
}

type GetStatusMasterArgs struct {
}

type GetStatusMasterReply struct {
	isDone bool
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	fmt.Println(s)
	return s
}

func workerSock(idWorker int) string {
	s := "/var/tmp/5840-mr-worker-"
	s += strconv.Itoa(idWorker)
	fmt.Println(s)
	return s
}
