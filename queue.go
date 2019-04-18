/***
**
**队列的实现(先进，先出)
**
 */
package main

import (
	"fmt"
	"sync"
)

var mutex sync.Mutex
var rmutex sync.RWMutex

//定义一个节点类型
type Node struct {
	data interface{}
	next *Node
}
type Queue struct {
	head  *Node //头指针
	tail  *Node //尾指针
	total int   //大小
}

//创建一个队列对象
func New() *Queue {
	fmt.Println("hello")
	rmutex.Lock()
	queue := &Queue{}
	queue.head = nil
	queue.tail = nil
	queue.total = 0
	rmutex.Unlock()
	return queue
}

//在尾部添加数据
func (this *Queue) Add(data interface{}) *Queue {
	rmutex.Lock()
	//当中间还没有数据时
	node := &Node{}
	node.data = data
	node.next = nil
	if this.total == 0 {
		this.head = node
		this.tail = node
	} else {
		this.tail.next = node
		this.tail = node
	}
	this.total++
	rmutex.Unlock()
	return this
}

//在头部移除数据
func (this *Queue) Pop() interface{} {
	rmutex.RLock()
	//当没有数据时直接返回空
	if this.total == 0 {
		return nil
	}
	node := this.head
	this.head = node.next
	this.total--
	rmutex.RUnlock()
	return node.data

}

func (this *Queue) Total() int {
	return this.total
}

//打印队列中的数据
func (this *Queue) Println() {
	fmt.Print("[")
	for head := this.head; head != nil; head = head.next {
		fmt.Print("", head.data, " ")
	}
	fmt.Println("]")
}
