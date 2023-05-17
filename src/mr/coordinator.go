package mr

import (
	"fmt"
	"log"
)
import "net"
import "os"
import "net/rpc"
import "net/http"
import "sync"

type Status string

const (
	Idle       Status = "Idle"
	Processing Status = "Processing"
	Complete   Status = "Complete"
)

type Coordinator struct {
	// Your definitions here.
	ReduceWorker     []int
	MapWorker        []int
	FileTaskAssigned map[string]Status
	mu               sync.Mutex
}

// Your code here -- RPC handlers for the worker to call.

// an example RPC handler.
// the RPC argument and reply types are defined in rpc.go.
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

// RPC handler
// return the file that not handle yet
func (c *Coordinator) GetFile(args *FileNotDoneYetArgs, reply *FileNotDoneYetReply) error {
	c.mu.Lock()
	fmt.Printf("Worker id %d want task \n", args.Id)
	c.MapWorker = append(c.MapWorker, args.Id)
	for nameFile, taskStatus := range c.FileTaskAssigned {
		if taskStatus == Idle {
			reply.Name = nameFile
			c.FileTaskAssigned[nameFile] = Processing
			break
		}
	}
	c.printMapStatus()
	c.mu.Unlock()
	return nil
}

func (c *Coordinator) GetStatusMaster(args *GetStatusMasterArgs, reply *GetStatusMasterReply) error {
	reply.isDone = c.Done()
	return nil
}

func (c *Coordinator) printMapStatus() {
	fmt.Printf("=== Map status ===\n")
	for k, v := range c.FileTaskAssigned {
		fmt.Printf("key: %s, value: %t \n", k, v)
	}
	fmt.Printf("==================\n")
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
//
func (c *Coordinator) Done() bool {
	ret := false
	// Your code here.
	//main/mrcoordinator.go expects mr/coordinator.go
	//to implement Done() method that returns true when
	//the MapReduce job is completely finished;
	//at that point, mrcoordinator.go will exit.
	return ret
}

//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{
		FileTaskAssigned: make(map[string]Status),
	}

	for _, filename := range os.Args[1:] {
		c.FileTaskAssigned[filename] = Idle
		//file, err := os.Open(filename)
		//if err != nil {
		//	log.Fatalf("cannot open %v", filename)
		//}
		//content, err := ioutil.ReadAll(file)
		//if err != nil {
		//	log.Fatalf("cannot read %v", filename)
		//}
		//file.Close()
		//kva := mapf(filename, string(content))
		//intermediate = append(intermediate, kva...)
	}
	c.printMapStatus()
	// Your code here.

	c.server()
	return &c
}
