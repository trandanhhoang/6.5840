package mr

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"time"
)
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

type ReduceWorker int

type Args struct {
	nameFile string
}

type Answer struct {
}

func (r *ReduceWorker) ReduceTask(args *Args, ans *Answer) error {
	log.Println("You get reduce task ", args.nameFile)
	return nil
}

func ReducePhase(reducef func(string, []string) string) {
	numberOfMapPhase, reduceIntermediaKey, err := callGetIntermediaKey()

	if numberOfMapPhase == -1 || reduceIntermediaKey == -1 || err != nil {
		time.Sleep(1 * time.Second)
		return
	}

	intermediate := make([]KeyValue, 0)
	// read file
	for i := 0; i < numberOfMapPhase; i++ {
		nameFile := "mr-" + strconv.Itoa(i) + "-" + strconv.Itoa(reduceIntermediaKey)
		file, err := os.Open(nameFile)
		if err != nil {
			log.Fatalf("cannot open %v", nameFile)
		}
		dec := json.NewDecoder(file)
		for {
			var kv KeyValue
			if err := dec.Decode(&kv); err != nil {
				break
			}
			intermediate = append(intermediate, kv)
		}
		file.Close()
	}

	oname := "mr-out-" + strconv.Itoa(reduceIntermediaKey)
	ofile, _ := os.Create(oname)

	// sort key
	//sort.SliceStable(kva, func(i, j int) bool {
	//	return kva[i].Key < kva[j].Key
	//})
	log.Println("hoang check", intermediate)
	sort.Sort(ByKey(intermediate))
	//log.Println("hoang check after sort", intermediate)
	// reduce phase
	i := 0
	for i < len(intermediate) {
		j := i + 1
		for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
			j++
		}
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, intermediate[k].Value)
		}
		output := reducef(intermediate[i].Key, values)

		// this is the correct format for each line of Reduce output.
		fmt.Fprintf(ofile, "%v %v\n", intermediate[i].Key, output)

		i = j
	}

	defer ofile.Close()
	// write to out file

	// call coordinator done reduce task
	time.Sleep(1 * time.Second)
	notifyDoneReduceJob(reduceIntermediaKey)
}

func notifyDoneReduceJob(reduceIntermediaKey int) {
	args := NotifyDoneReduceJobArgs{
		IntermediaKey: reduceIntermediaKey,
	}
	reply := NotifyDoneReduceJobReply{}

	ok := call("Coordinator.NotifyDoneReduceJob", &args, &reply)
	if ok {
		// reply.Y should be 100.
		log.Printf("reply.IsNotify %v\n", reply.IsNotify)
	} else {
		log.Printf("call failed!\n")
	}
}

// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

//	log.Fprintf(ofile, "%v %v\n", intermediate[i].Key, intermediate[i].Value)

func callGetIntermediaKey() (int, int, error) {
	args := GetReduceIntermediaKeyArgs{}
	reply := GetReduceIntermediaKeyReply{}

	ok := call("Coordinator.GetIntermediaKey", &args, &reply)
	if ok {
		log.Printf("reply.numberOfTaskPhase %v reply.ReduceIntermediaKey %v \n", reply.NumberOfMapPhase, reply.ReduceIntermediaKey)
	} else {
		log.Printf("call failed!\n")
		return -1, -1, errors.New("call failed!")
	}
	return reply.NumberOfMapPhase, reply.ReduceIntermediaKey, nil
}

//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {
	// uncomment to send the Example RPC to the coordinator.
	//CallExample()

	for !isStatusMapPhaseDone() {
		MapPhase(mapf)
	}

	for !isStatusReducePhaseDone() {
		ReducePhase(reducef)
	}

	for !isCorDone() {
		time.Sleep(1 * time.Second)
	}
	log.Println("worker done")
}

func isStatusReducePhaseDone() bool {
	args := GetStatusMasterArgs{}
	reply := GetStatusMasterReply{}

	ok := call("Coordinator.GetStatusReducePhase", &args, &reply)
	if ok {
		log.Printf("isStatusReducePhaseDone %v\n", reply.IsDone)
		return reply.IsDone
	} else {
		log.Printf("call failed!\n")
	}
	return true
}

func isStatusMapPhaseDone() bool {
	args := GetStatusMasterArgs{}
	reply := GetStatusMasterReply{}

	ok := call("Coordinator.GetStatusMapPhase", &args, &reply)
	if ok {
		log.Printf("isStatusMapPhaseDone %v\n", reply.IsDone)
		return reply.IsDone
	} else {
		log.Printf("call failed!\n")
	}
	return true
}

func isCorDone() bool {
	args := GetStatusMasterArgs{}
	reply := GetStatusMasterReply{}

	ok := call("Coordinator.GetStatusMaster", &args, &reply)
	if ok {
		log.Printf("isStatusMapPhaseDone %v\n", reply.IsDone)
		return reply.IsDone
	} else {
		log.Printf("call failed!\n")
	}
	return true
}

type Intermedia struct {
	listKeyValue []KeyValue
}

func MapPhase(mapf func(string, string) []KeyValue) {
	nameFile, idMapPhase, nReduce, err := callGetFile()
	if nameFile == "" || err != nil {
		time.Sleep(1 * time.Second)
		return
	}

	listIntermedia := []Intermedia{}
	for i := 0; i < nReduce; i++ {
		listIntermedia = append(listIntermedia, Intermedia{
			listKeyValue: []KeyValue{},
		})
	}
	// read file
	file, err := os.Open(nameFile)
	if err != nil {
		log.Fatalf("cannot open %v", nameFile)
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("cannot read %v", nameFile)
	}
	file.Close()

	//map phase -> create intermedia file
	kva := mapf(nameFile, string(content))

	// calculate reduce number
	for _, kv := range kva {
		reduceNumber := ihash(kv.Key) % nReduce
		listIntermedia[reduceNumber].listKeyValue = append(listIntermedia[reduceNumber].listKeyValue, kv)
	}

	for i := 0; i < nReduce; i++ {
		fileWrite, err := getFile(idMapPhase, i)
		if err != nil {
			log.Fatal("cannot open file ", err)
			return
		}
		enc := json.NewEncoder(fileWrite)
		for _, kv := range listIntermedia[i].listKeyValue {
			err = enc.Encode(&kv)
			if err != nil {
				log.Fatal("cannot Encode ", err)
				return
			}
		}
		fileWrite.Close()
	}

	time.Sleep(1 * time.Second)
	notifyDoneMapJob(nameFile)
}

func notifyDoneMapJob(nameFile string) {
	args := NotifyDoneJobArgs{
		NameFile: nameFile,
	}
	reply := NotifyDoneJobReply{}

	ok := call("Coordinator.NotifyDoneJob", &args, &reply)
	if ok {
		// reply.Y should be 100.
		log.Printf("reply.IsNotify %v\n", reply.IsNotify)
	} else {
		log.Printf("call failed!\n")
	}
}

func getFile(idMapPhase string, nReduce int) (*os.File, error) {
	filename := "mr-" + idMapPhase + "-" + strconv.Itoa(nReduce)
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		log.Println("Failed to open file:", err)
		return nil, err
	}
	return file, nil
}

func CallExample() {
	args := ExampleArgs{}
	args.X = 99
	reply := ExampleReply{}

	ok := call("Coordinator.Example", &args, &reply)
	if ok {
		log.Printf("reply.Y %v\n", reply.Y)
	} else {
		log.Printf("call failed!\n")
	}
}

func callGetFile() (string, string, int, error) {
	args := FileNotDoneYetArgs{}
	reply := FileNotDoneYetReply{}

	ok := call("Coordinator.GetFile", &args, &reply)
	if ok {
		log.Printf("reply.NameFile %v, reply.MapTask %v, reply.NumberReduceTask %v \n", reply.NameFile, reply.MapTask, reply.NReduce)
	} else {
		log.Printf("call failed!\n")
		return "", "", 0, errors.New("call failed!")
	}
	return reply.NameFile, strconv.Itoa(reply.MapTask), reply.NReduce, nil
}

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

	log.Println(err)
	return false
}
