package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

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
// 0 成功获取 Map 任务。此任务第一次被执行
// 1 成功获取 Map 任务。属于重试任务
// 2 没有新的 Map 任务，但是 Map 阶段还未完成。需要隔一段时间继续请求，因为可能有失败的任务
// 3 没有新的 Map 任务，Map 阶段已完成，该请求 Reduce 任务了
type MapTeskRes struct {
	Key       string
	TaskId    int // 任务编号
	ReduceNum int
	State     int // 任务状态
}

type MapTaskDoneReq struct {
	TaskId int
}

type MapTaskDoneRes struct {
	State int // 0 需要继续请求 Map 任务，1 需要请求 Reduce 任务
}

type ReduceTaskReq struct {
}

type ReduceTaskRes struct {
	MapNumber int
	TaskId    int
	State     int // 0 新的 R 任务，第一次执行；1 重试的 R 任务；2 没有新任务，但是 R 阶段未完成；3 没有新任务了，退出程序
}

type ReduceTaskDoneReq struct {
	TaskId int
}

type ReduceTaskDoneRes struct {
	State int // 0 需要继续请求 R 任务，1 没有新任务了，退出程序
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
