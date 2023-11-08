package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import "os"
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
type MapTaskReq struct {

}

// State 
// 0 非法
// 1 成功获取 Map 任务
// 2 没有新的 Map 任务，但是 Map 阶段还未完成
// 3 没有新的 Map 任务，Map 阶段已完成，该请求 Reduce 任务了
type MapTeskRes struct {
	Key string
	Value string
	State int
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
