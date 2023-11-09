package mr

import (
	"fmt"
	"hash/fnv"
	"log"
	"net/rpc"
	"time"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// main/mrworker.go calls this function.
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {
	// Your worker implementation here.

	for { // maping
		var mapReq MapTaskReq
		var mapRes MapTeskRes
		if call("Coordinator.GetMapTask", &mapReq, &mapRes) {
			if mapRes.State == 0 || mapRes.State == 1 {
				var doneReq = MapTaskDoneReq{TaskId: mapRes.TaskId}
				var doneRes MapTaskDoneRes
				if call("Coordinator.MapTeskDone", &doneReq, &doneRes) {
					if doneRes.State == 1 {
						break
					}
				}
			} else if mapRes.State == 2 {

			} else if mapRes.State == 3 {
				break
			} else {
				panic(mapRes.State)
			}
		}

		time.Sleep(200 * time.Millisecond)
	}

	for { // reducing
		var rdcTaskReq ReduceTaskReq
		var rdcTaskRes ReduceTaskRes
		if call("Coordinator.GetReduceTask", &rdcTaskReq, &rdcTaskRes) {
			if rdcTaskRes.State == 0 || rdcTaskRes.State == 1 {
				var doneReq = ReduceTaskDoneReq{TaskId: rdcTaskRes.TaskId}
				var doneRes ReduceTaskDoneRes
				if call("Coordinator.ReduceTaskDone", &doneReq, &doneRes) {
					if doneRes.State == 1 {
						break
					}
				}
			} else if rdcTaskRes.State == 2 {

			} else if rdcTaskRes.State == 3 {
				break
			} else {
				panic(rdcTaskRes.State)
			}
		}
	}

	// uncomment to send the Example RPC to the coordinator.
	// CallExample()

}

// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
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

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
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
