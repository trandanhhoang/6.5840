package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
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

type FileNotDoneYetArgs struct{}

type FileNotDoneYetReply struct {
	NameFile string
	MapTask  int
	NReduce  int
}

type GetReduceIntermediaKeyArgs struct{}

type GetReduceIntermediaKeyReply struct {
	NumberOfMapPhase    int
	ReduceIntermediaKey int
}

type GetStatusMasterArgs struct {
}

type GetStatusMasterReply struct {
	IsDone bool
}

type NotifyDoneJobArgs struct {
	NameFile string
}

type NotifyDoneJobReply struct {
	IsNotify bool
}

type NotifyDoneReduceJobArgs struct {
	IntermediaKey int
}

type NotifyDoneReduceJobReply struct {
	IsNotify bool
}

//type AssignReduceTaskArgs struct {
//}
//
//type AssignReduceTaskReply struct {
//}

type StatusOfTaskArgs struct {
}

type StatusOfTaskReply struct {
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	//fmt.Println(s)
	return s
}
