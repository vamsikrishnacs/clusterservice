package cluster

import (
	"encoding/json"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

const (
	BROADCAST = -1
)

type Envelope struct {
	// On the sender side, Pid identifies the receiving peer. If instead, Pid is
	// set to cluster.BROADCAST, the message is sent to all peers. On the receiver side, the
	// Id is always set to the original sender. If the Id is not found, the message is silently dropped
	Pid int

	// An id that globally and uniquely identifies the message, meant for duplicate detection at
	// higher levels. It is opaque to this package.
	MsgId int64

	// the actual message.
	Msg string
}

/*type Server interface {
// Id of this server
Pid() int
// array of other servers' ids in the same cluster
Peers() []int
// the channel to use to send messages to other peers
// Note that there are no guarantees of message delivery, and messages
// are silently dropped
Outbox() chan *Envelope
// the channel to receive messages from other peers.
Inbox() chan *Envelope
}
*/

type Server struct {
	serverID      int
	serverAddress string
	peers         []int
	peerAddress   map[string]string
	outbox        chan *Envelope
	inbox         chan *Envelope
	peersock      []*zmq.Socket
	c             int
}

// Id of this server

func (serv *Server) Pid() int {
	return serv.serverID
}

// array of other servers' ids in the same cluster
func (serv *Server) Peers() []int {
	return serv.peers
}

// the channel to use to send messages to other peers
// Note that there are no guarantees of message delivery, and messages are silently dropped
func (serv *Server) Outbox() chan *Envelope {
	return serv.outbox
}

// the channel to receive messages from other peers.
func (s *Server) Inbox() chan *Envelope {
	return s.inbox
}

type cc struct {
	Ser     map[string]string
	Send    int
	Receive int
	Peer    []int
}

func New(id int, path string) Server {

	//1.reading the configuration file
	file, e := ioutil.ReadFile(path)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	var c cc
	json.Unmarshal(file, &c)

	id1 := strconv.Itoa(id)
	//2.initialising

	o := make(chan *Envelope, 1000)
	i := make(chan *Envelope, 1000)
	var tempSocks = make([]*zmq.Socket, len(c.Peer))

	s := Server{id, c.Ser[id1], c.Peer, c.Ser, o, i, tempSocks, 0}

	s.connect()
	//3.handling server sockets
	go handleServerrep(&s)
	go handleServerreq(&s)

	fmt.Println(s.serverID)
	fmt.Println(s.peers)
	for i, _ := range s.peerAddress {
		fmt.Println(i + ":" + s.peerAddress[i])
	}
	return s
}

//SUBROUTINE RECEIVE

func handleServerrep(s *Server) {

	//1.point-point messaging

	socketrep, _ := zmq.NewSocket(zmq.REP)
	/*
	err := socket.Connect(PROTOCOL + addr)

	*/
	err := socketrep.Bind("tcp://" + s.serverAddress)
	if err != nil {
		fmt.Println(err)
	}

	//socket.Bind("tcp://127.0.0.1:6000")
	s.c = 0
	for {

		msg, _ := socketrep.Recv(0)
		var msg1 Envelope
		json.Unmarshal([]byte(msg), &msg1)
		s.inbox <- &msg1 //&Envelope{1,1,msg}
		s.c = s.c + 1
		count := strconv.Itoa(s.c)
		println("Got", string(msg)+"--msg count:"+count)
		socketrep.Send("Ack ok", 0)
	}

}

//SUBROUTINE SEND

func handleServerreq(s *Server /*,pid int*/) {

	//send msgs
	for {
		select {
		case envelope := <-s.Outbox():
			// fmt.Printf("Msg send request to-- %d: '%s'\n", envelope.Pid, envelope.Msg)
			if envelope.Pid == -1 {
				Sendbroadcastmessage(envelope, s)
			} else {
				Sendmessage(envelope, s)
			}
		case <-time.After(10 * time.Second):
			// fmt.Println("Waited and waited. Ab thak gaya\n")
		}
	}
}

func (s *Server) connect() {

	for j1 := range s.peers {
		j := strconv.Itoa(j1 + 1)

		if (j1 + 1) != s.serverID {

			sock, _ := zmq.NewSocket(zmq.REQ)
			s.peersock[j1] = sock

			//    fmt.Println("CONNECTED")
			err := s.peersock[j1].Connect("tcp://" + s.peerAddress[j])
			if err != nil {
				fmt.Println(err)
			}
		}

	}

}

func Sendmessage(envlope *Envelope, s *Server) {

	peerid := (envlope.Pid)
	//1.point-point messaging

	envlope.Pid = s.serverID
	ms, _ := json.Marshal(*envlope)

	_, err1 := s.peersock[peerid-1].Send(string(ms), 0)
	if err1 != nil {
		fmt.Println("error")
	}

	fmt.Println("Sending")
	_, err := s.peersock[peerid-1].Recv(0)
	if err == nil {
		fmt.Println("Ack")
	}

}

func Sendbroadcastmessage(envlope *Envelope, s *Server) {

	//1.broad-cast messaging

	envlope.Pid = s.serverID
	ms, _ := json.Marshal(*envlope)

	for j1 := range s.peers {
		if (j1 + 1) != s.serverID {
			fmt.Println("Sending")
			_, err1 := s.peersock[j1].Send(string(ms), 0)
			if err1 != nil {
				fmt.Println(err1)
			}
			_, err := s.peersock[j1].Recv(0)
			if err == nil {
				fmt.Println("Ack rec", j1)
			}
		}
	}

}
