package cluster

import "testing"

//import "log"
import (
	"fmt"
	"strconv"
)

func Test(t *testing.T) {
	c := 0
	s := New(1, "./d1.json")
	s1 := New(2, "./d1.json")
	s2 := New(3, "./d1.json")
	s3 := New(4, "./d1.json")
	s4 := New(5, "./d1.json")
	//-------------------------tests----------------

	//1.basic test send and receive

	s2.Outbox() <- &Envelope{Pid: -1, Msg: "Msg test1"}
	s4.Outbox() <- &Envelope{Pid: 2, Msg: "Msg test 2"}
	s3.Outbox() <- &Envelope{Pid: 2, Msg: "Msg test 2"}
	s.Outbox() <- &Envelope{Pid: 5, Msg: "Msg test 2"}
	s2.Outbox() <- &Envelope{Pid: 4, Msg: "Msg test 2"}
	s1.Outbox() <- &Envelope{Pid: 3, Msg: "Msg test 2"}
	s3.Outbox() <- &Envelope{Pid: 5, Msg: "Msg test 2"}

	c += 10

	//2.broadcast from every server to everyone else

	s.Outbox() <- &Envelope{Pid: -1 /*cluster.BROADCAST*/, Msg: "broadcast from 1."}
	s1.Outbox() <- &Envelope{Pid: -1 /*cluster.BROADCAST*/, Msg: "broadcast from 2."}
	s2.Outbox() <- &Envelope{Pid: -1 /*cluster.BROADCAST*/, Msg: "broadcast from 3."}
	c += 12
	//3.round--robin test

	for i := 0; i < 25; i++ {
		msg := "round-robin msg no:"
		s.Outbox() <- &Envelope{Pid: 2, Msg: msg + " from 1"}
		s1.Outbox() <- &Envelope{Pid: 3, Msg: msg + " from 2"}
		s2.Outbox() <- &Envelope{Pid: 1, Msg: msg + " from 3"}
		c += 3
	}

	//4.send arbitrary long no of messages

	for i := 0; i < 500; i++ {
		Msg1 := strconv.Itoa(i)
		s4.Outbox() <- &Envelope{Pid: -1 /*cluster.BROADCAST*/, Msg: "Broadcast:" + Msg1}
		c += 4
	}
	fmt.Println(c)
	fmt.Println("c=")

	c1 := 0
	for {
		if c == c1 {
			fmt.Println("Messages sent:")
			fmt.Println(c)
			fmt.Println("Messages received:")
			fmt.Println(c1)
			break
		}

		select {

		case <-s.Inbox():
			fmt.Println("received111111111")
			c1++
		case <-s1.Inbox():
			fmt.Println("receieved222222")
			c1++
		case <-s2.Inbox():
			fmt.Println("receieved3333333")
			c1++
		case <-s3.Inbox():
			fmt.Println("receieved444444")
			c1++
		case <-s4.Inbox():
			fmt.Println("receiev555555")
			c1++

		}
	}

}
