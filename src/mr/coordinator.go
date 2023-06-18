package mr

import (
	"log"
	"time"
)
import "net"
import "os"
import "net/rpc"
import "net/http"
import "sync"

type Status string

type StatusMapTask struct {
	NumberOfTask int
	Status       Status
}

const (
	Idle       Status = "Idle"
	Processing Status = "Processing"
	Complete   Status = "Complete"
)

type Coordinator struct {
	// Your definitions here.
	NumberOfReducer   int
	FileTaskAssigned  map[string]*StatusMapTask
	ReducePhaseStatus map[int]Status
	mu                sync.Mutex
}

// Your code here -- RPC handlers for the worker to call.

// an example RPC handler.
// the RPC argument and reply types are defined in rpc.go.
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

func (c *Coordinator) GetIntermediaKey(args *GetReduceIntermediaKeyArgs, reply *GetReduceIntermediaKeyReply) error {
	c.mu.Lock()
	reply.ReduceIntermediaKey = -1
	reply.NumberOfMapPhase = -1
	for index, taskStatus := range c.ReducePhaseStatus {
		if taskStatus == Idle {
			reply.ReduceIntermediaKey = index
			reply.NumberOfMapPhase = len(c.FileTaskAssigned)

			c.ReducePhaseStatus[index] = Processing
			c.printReduceStatus()
			//do need check status of reduce task
			go c.checkStatusOfReduceTask(index)
			break
		}
	}
	c.mu.Unlock()
	return nil
}

// RPC handler
// return the file that not handle yet
func (c *Coordinator) GetFile(args *FileNotDoneYetArgs, reply *FileNotDoneYetReply) error {
	c.mu.Lock()
	for nameFile, taskStatus := range c.FileTaskAssigned {
		if taskStatus.Status == Idle {
			// update Corrdinator
			c.FileTaskAssigned[nameFile].Status = Processing

			// update response for worker
			reply.NameFile = nameFile
			reply.MapTask = taskStatus.NumberOfTask
			reply.NReduce = c.NumberOfReducer

			c.printMapStatus()

			go c.checkStatusOfTask(nameFile)
			break
		}
	}
	c.mu.Unlock()

	return nil
}

//func (c *Coordinator) AssignReduceTask(args *AssignReduceTaskArgs, reply *AssignReduceTaskReply) error {
//
//	return nil
//}

func (c *Coordinator) checkStatusOfReduceTask(key int) {
	time.Sleep(10 * time.Second)
	c.mu.Lock()
	if c.ReducePhaseStatus[key] == Processing {
		// update coordinator if worker not done job
		log.Println("Worker not done reduce job, reassign task name ", key)
		c.ReducePhaseStatus[key] = Idle
		c.printReduceStatus()
	}
	c.mu.Unlock()
}

func (c *Coordinator) checkStatusOfTask(nameFile string) {
	time.Sleep(10 * time.Second)
	c.mu.Lock()
	if c.FileTaskAssigned[nameFile].Status == Processing {
		// update coordinator if worker not done job
		log.Println("Worker not done job, reassign task name ", nameFile)
		c.FileTaskAssigned[nameFile].Status = Idle
		c.printMapStatus()
	}
	c.mu.Unlock()
}

func (c *Coordinator) NotifyDoneJob(args *NotifyDoneJobArgs, reply *NotifyDoneJobReply) error {
	c.mu.Lock()
	if c.FileTaskAssigned[args.NameFile].Status == Processing {
		log.Println("Worker done job task ", args.NameFile)
		c.FileTaskAssigned[args.NameFile].Status = Complete
		c.printMapStatus()
	}
	c.mu.Unlock()
	return nil
}

func (c *Coordinator) NotifyDoneReduceJob(args *NotifyDoneReduceJobArgs, reply *NotifyDoneReduceJobReply) error {
	c.mu.Lock()
	if c.ReducePhaseStatus[args.IntermediaKey] == Processing {
		log.Println("Worker done reduce job task ", args.IntermediaKey)
		c.ReducePhaseStatus[args.IntermediaKey] = Complete
		c.printReduceStatus()
	}
	c.mu.Unlock()
	return nil
}

func (c *Coordinator) GetStatusMapPhase(args *GetStatusMasterArgs, reply *GetStatusMasterReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.printMapStatus()
	for _, v := range c.FileTaskAssigned {
		if v.Status == Idle || v.Status == Processing {
			reply.IsDone = false
			return nil
		}
	}
	reply.IsDone = true
	return nil
}

func (c *Coordinator) GetStatusReducePhase(args *GetStatusMasterArgs, reply *GetStatusMasterReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, v := range c.ReducePhaseStatus {
		if v == Idle || v == Processing {
			reply.IsDone = false
			return nil
		}
	}
	reply.IsDone = true
	return nil
}

func (c *Coordinator) GetStatusMaster(args *GetStatusMasterArgs, reply *GetStatusMasterReply) error {
	reply.IsDone = c.Done()
	return nil
}

func (c *Coordinator) printReduceStatus() {
	log.Printf("=== Map status ===\n")
	for k, v := range c.ReducePhaseStatus {
		log.Printf("key: %v, status: %v \n", k, v)
	}
	log.Printf("==================\n")
}

func (c *Coordinator) printMapStatus() {
	log.Printf("=== Map status ===\n")
	for k, v := range c.FileTaskAssigned {
		log.Printf("key: %s, status: %v, mapTask: %v \n", k, string(v.Status), v.NumberOfTask)
	}
	log.Printf("==================\n")
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

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, v := range c.ReducePhaseStatus {
		if v != Complete {
			return ret
		}
	}
	ret = true
	return ret
}

//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{
		NumberOfReducer:   nReduce,
		FileTaskAssigned:  make(map[string]*StatusMapTask),
		ReducePhaseStatus: make(map[int]Status),
	}

	for i, filename := range files {
		c.FileTaskAssigned[filename] = &StatusMapTask{
			NumberOfTask: i,
			Status:       Idle,
		}
	}

	for i := 0; i < nReduce; i++ {
		c.ReducePhaseStatus[i] = Idle
	}
	c.printMapStatus()
	// Your code here.

	c.server()
	return &c
}
