package utils

import (
	"errors"
	"sync"
	"time"
)

type PoolConfig struct {
	IdleNum int
	Factory func() (interface{}, error)
	Close   func(interface{}) error
}

var (
	//ErrClosed 任务池已经关闭Error
	ErrClosed = errors.New("pool is closed")
)

type taskPool struct {
	mu      sync.Mutex
	tasks   chan *idleTask
	factory func() (interface{}, error)
	close   func(interface{}) error
}

type idleTask struct {
	task interface{}
	t    time.Time
}

func NewTaskPool(poolConfig *PoolConfig) (Pool, error) {
	c := &taskPool{
		tasks:   make(chan *idleTask, poolConfig.IdleNum),
		factory: poolConfig.Factory,
		close:   poolConfig.Close,
	}
	return c, nil
}

func (c *taskPool) getTasks() chan *idleTask {
	c.mu.Lock()
	tasks := c.tasks
	c.mu.Unlock()
	return tasks
}

func (c *taskPool) Get() (interface{}, error) {
	tasks := c.getTasks()
	if tasks == nil {
		return nil, ErrClosed
	}
	for {
		select {
		case wrapTask := <-tasks:
			if wrapTask == nil {
				return nil, ErrClosed
			}
			return wrapTask.task, nil
		default:
			task, err := c.factory()
			if err != nil {
				return nil, err
			}

			return task, nil
		}
	}
}

func (c *taskPool) Put(task interface{}) error {
	if task == nil {
		return errors.New("task is nil. rejecting")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.tasks == nil {
		return c.Close(task)
	}

	select {
	case c.tasks <- &idleTask{task: task, t: time.Now()}:
		return nil
	default:
		//连接池已满，直接关闭该链接
		return c.Close(task)
	}
}

//Close 关闭单条连接
func (c *taskPool) Close(task interface{}) error {
	if task == nil {
		return errors.New("task is nil. rejecting")
	}
	return c.close(task)
}

//Release 释放任务池中所有任务
func (c *taskPool) Release() {
	c.mu.Lock()
	tasks := c.tasks
	c.factory = nil
	closeFun := c.close
	c.close = nil
	c.mu.Unlock()

	if tasks == nil {
		return
	}

	close(tasks)
	for wrapTask := range tasks {
		closeFun(wrapTask)
	}
}

//Nums 任务池中已有的任务
func (c *taskPool) Nums() int {
	return len(c.getTasks())
}

//Pool 基本方法
type Pool interface {
	Get() (interface{}, error)

	Put(interface{}) error

	Close(interface{}) error

	Release()

	Nums() int
}
