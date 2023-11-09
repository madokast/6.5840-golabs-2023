package mr

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type Coordinator struct {
	// Your definitions here.
	mapTasks    []MapTask
	reduceTasks []ReduceTask // length nReduce
	state       int          // 集群状态 0 初始化，1 map 进行中，2 reduce 中，3 结束
	rw          sync.RWMutex
}

type MapTask struct {
	key       string
	state     int       // 0 新任务，1 已分发未完成，2 已完成
	startTime time.Time // 如果已分发，表示开始的时间
}

type ReduceTask struct {
	state     int       // 0 新任务，1 已分发未完成，2 已完成
	startTime time.Time // 如果已分发，表示开始的时间
}

// Your code here -- RPC handlers for the worker to call.

// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

func (c *Coordinator) GetMapTask(args *MapTaskReq, reply *MapTeskRes) error {
	defer func() { log.Printf("GetMapTask %+v %+v", args, reply) }()
	c.rw.Lock()
	defer c.rw.Unlock()
	reply.ReduceNum = len(c.reduceTasks)
	switch c.state {
	case 0:
		reply.State = 2
		return nil
	case 1:
		now := time.Now()
		for i := range c.mapTasks {
			if c.mapTasks[i].state == 0 {
				reply.Key = c.mapTasks[i].key
				reply.TaskId = i
				reply.State = 0
				c.mapTasks[i].startTime = now
				c.mapTasks[i].state = 1
				return nil
			} else if c.mapTasks[i].state == 1 {
				if now.Sub(c.mapTasks[i].startTime) >= 10*time.Second {
					reply.Key = c.mapTasks[i].key
					reply.TaskId = i
					reply.State = 1
					c.mapTasks[i].startTime = now
					c.mapTasks[i].state = 1
					return nil
				}
			}
		}
		reply.State = 2
		return nil
	case 2:
		fallthrough
	case 3:
		reply.State = 3
		return nil
	default:
		log.Fatal("unknown state", c.state)
		return errors.New(fmt.Sprint("unknown state", c.state))
	}
}

func (c *Coordinator) MapTeskDone(args *MapTaskDoneReq, reply *MapTaskDoneRes) error {
	defer func() { log.Printf("MapTeskDone %+v %+v", args, reply) }()
	c.rw.Lock()
	defer c.rw.Unlock()

	c.mapTasks[args.TaskId].state = 2
	for i := range c.mapTasks {
		if c.mapTasks[i].state != 2 {
			reply.State = 0
			return nil
		}
	}
	reply.State = 1
	c.state = 2 // into reduce
	return nil
}

func (c *Coordinator) GetReduceTask(args *ReduceTaskReq, reply *ReduceTaskRes) error {
	defer func() { log.Printf("GetReduceTask %+v %+v", args, reply) }()
	c.rw.Lock()
	defer c.rw.Unlock()
	reply.MapNumber = len(c.mapTasks)
	switch c.state {
	case 0:
		fallthrough
	case 1:
		log.Fatal("reduce task req in init or map phase")
		return errors.New("reduce task req in init or map phase")
	case 2:
		now := time.Now()
		for i := range c.reduceTasks {
			if c.reduceTasks[i].state == 0 {
				reply.TaskId = i
				reply.State = 0
				c.reduceTasks[i].state = 1
				c.reduceTasks[i].startTime = now
				return nil
			} else if c.reduceTasks[i].state == 1 {
				if now.Sub(c.reduceTasks[i].startTime) >= 10*time.Second {
					reply.TaskId = i
					reply.State = 1
					c.reduceTasks[i].startTime = now
					return nil
				}
			}
		}
		reply.State = 2
		return nil
	case 3:
		reply.State = 3
		return nil
	default:
		log.Fatal("unknown state", c.state)
		return errors.New(fmt.Sprint("unknown state", c.state))
	}
}

func (c *Coordinator) ReduceTaskDone(args *ReduceTaskDoneReq, reply *ReduceTaskDoneRes) error {
	defer func() { log.Printf("ReduceTaskDone %+v %+v", args, reply) }()
	c.rw.Lock()
	defer c.rw.Unlock()

	c.reduceTasks[args.TaskId].state = 2
	for i := range c.reduceTasks {
		if c.reduceTasks[i].state != 2 {
			reply.State = 0
			return nil
		}
	}
	reply.State = 1
	c.state = 3
	return nil
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

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	c.rw.Lock()
	defer c.rw.Unlock()

	switch c.state {
	case 0:
		fallthrough
	case 1:
		return false
	case 2:
		for i := range c.reduceTasks {
			if c.reduceTasks[i].state != 2 {
				return false
			}
		}
		c.state = 3
		return true
	case 3:
		return true
	default:
		panic(fmt.Sprint("unknown state", c.state))
	}
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := &Coordinator{}

	// Your code here.
	// 构建 mapTask
	c.mapTasks = make([]MapTask, len(files))
	for i, filename := range files {
		c.mapTasks[i].key = filename
	}

	// 构建 reduceTask
	c.reduceTasks = make([]ReduceTask, nReduce)

	c.server()
	c.state = 1
	log.Println("master init done")
	return c
}
