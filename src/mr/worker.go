package mr

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	random "math/rand"
	"net"
	"net/http"
	"os"
	"time"
)
import "log"
import "net/rpc"
import "hash/fnv"

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// TODO read file back
//dec := json.NewDecoder(file)
//for {
//var kv KeyValue
//if err := dec.Decode(&kv); err != nil {
//break
//}
//kva = append(kva, kv)
//}

type ReduceWorker int

type Args struct {
	nameFile string
}

type Answer struct {
}

func (r *ReduceWorker) ReduceTask(args *Args, ans *Answer) error {
	fmt.Println("You get reduce task ", args.nameFile)
	return nil
}

func serverWorker(idWorker int) {
	reduceWorker := new(ReduceWorker)
	rpc.Register(reduceWorker)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := workerSock(idWorker)
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

func ReducePhase(reducef func(string, []string) string, idWorker int) {
	serverWorker(idWorker)
}

func genId() int {
	random.Seed(time.Now().UnixNano())
	return random.Intn(90) + 100
}

//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {
	idWorker := genId()
	// Your worker implementation here.
	go ReducePhase(reducef, idWorker)

	// uncomment to send the Example RPC to the coordinator.
	CallExample()
	MapPhase(mapf, idWorker)

	for !isCorDone() {
		time.Sleep(10 * time.Second)
	}
}

func isCorDone() bool {
	return false
}

func MapPhase(mapf func(string, string) []KeyValue, idWorker int) {
	nameFile, err := CallGetFile(idWorker)
	if err != nil {
		return
	}

	// read file
	intermediate := []KeyValue{}
	file, err := os.Open(nameFile)
	if err != nil {
		log.Fatalf("cannot open %v", nameFile)
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("cannot read %v", nameFile)
	}
	file.Close()
	kva := mapf(nameFile, string(content))
	intermediate = append(intermediate, kva...)

	//fmt.Print(intermediate)

	//write file
	oname := "mr-0-0"
	ofile, _ := os.Create(oname)
	defer ofile.Close()

	enc := json.NewEncoder(ofile)
	for _, kv := range intermediate {
		err := enc.Encode(&kv)
		if err != nil {
			log.Fatal("cannot Encode ", err)
			return
		}
	}

	//for i := 0; i < len(intermediate); i++ {
	//	enc := json.NewEncoder(file)
	//	// this is the correct format for each line of Reduce output.
	//	fmt.Fprintf(ofile, "%v %v\n", intermediate[i].Key, intermediate[i].Value)
	//}
}

//
// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
//
func CallExample() {
	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.Example", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("reply.Y %v\n", reply.Y)
	} else {
		fmt.Printf("call failed!\n")
	}
}

// TODO hoang
func CallGetFile(idWorker int) (string, error) {
	// declare an argument structure.
	args := FileNotDoneYetArgs{
		Id: idWorker,
	}

	// fill in the argument(s).

	// declare a reply structure.
	reply := FileNotDoneYetReply{}

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.GetFile", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("reply.Name %v\n", reply.Name)
	} else {
		fmt.Printf("call failed!\n")
		return "", errors.New("call failed!")
	}
	return reply.Name, nil
}

//
// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
