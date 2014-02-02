package main

import(
	"encoding/json"
	"os"
	"fmt"
        "io/ioutil"
	"strconv"
          zmq "github.com/pebbe/zmq4"
        "time"
)
const (BROADCAST = -1)
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
        serverID int
        serverAddress string
        peers []int
        peerAddress map[string]string
        outbox chan *Envelope
        inbox chan *Envelope
	c int
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

type cc struct{
Ser map[string]string
Send int
Receive int
Peer []int
}

func New(id int,path string)Server{

 
//1.reading the configuration file
file, e := ioutil.ReadFile(path)
 if e != nil {
        fmt.Printf("File error: %v\n", e)
        os.Exit(1)
    }
    
       var c cc
       json.Unmarshal(file,&c)

        id1:=strconv.Itoa(id)
	//2.initialising
		
		o:= make(chan *Envelope, 1000)
		i:= make(chan *Envelope, 1000)

s:=Server{id,c.Ser[id1],c.Peer,c.Ser,o,i,0}

         //3.handling server sockets
        go handleServerrep(&s)
        go handleServerreq(&s)   

fmt.Println(s.serverID)
fmt.Println(s.peers)
for i,_:=range s.peerAddress{
fmt.Println(i+":"+s.peerAddress[i])
}         
           return s
}


//SUBROUTINE RECEIVE

func handleServerrep(s *Server){

        //1.point-point messaging

       
	socketrep, _ := zmq.NewSocket(zmq.REP)
        /*
        err := socket.Connect(PROTOCOL + addr)

    */
        socketrep.Bind("tcp://"+s.serverAddress)


        

       //socket.Bind("tcp://127.0.0.1:6000")
        s.c=0
        for {
                
		msg, _ := socketrep.Recv(0)
                var msg1 Envelope
                json.Unmarshal([]byte(msg), &msg1)
		s.inbox<-&msg1//&Envelope{1,1,msg}
		s.c=s.c+1                
		count:=strconv.Itoa(s.c)
		println("Got", string(msg)+"--msg count:"+count)
                socketrep.Send("Ack ok", 0)
        }

}


//SUBROUTINE SEND

func handleServerreq(s *Server/*,pid int*/){

         //send msgs 
        for{
         select{
         case envelope :=<-s.Outbox(): 
           fmt.Printf("Msg send request to-- %d: '%s'\n", envelope.Pid, envelope.Msg)
	   if envelope.Pid==-1{
                Sendbroadcastmessage(envelope,s)
           }else{
        	Sendmessage(envelope,s)
           }
         case <-time.After(10 * time.Second): 
           fmt.Println("Waited and waited. Ab thak gaya\n")
         }  
         }      
}

func Sendmessage(envlope *Envelope,s *Server){
         
        peerid:=(envlope.Pid)
           //1.point-point messaging
        socketreq, _ := zmq.NewSocket(zmq.REQ)
	t:=strconv.Itoa(peerid)
        socketreq.Connect("tcp://"+s.peerAddress[t])
        fmt.Println("s--msg from 1 to"+t+":"+s.peerAddress[t])
	envlope.Pid=s.serverID       
	ms,_:=json.Marshal(*envlope)

        socketreq.Send(string(ms), 0)
	
        fmt.Println("Sending")
        _,err:=socketreq.Recv(0)
	if err == nil {
		fmt.Println("Ack")
        }
        //s.inbox<-&Envelope{int(s.serverAddress),2,m}
        
        
}



func Sendbroadcastmessage(envlope *Envelope,s *Server){
         
                   //1.broad-cast messaging
        socketreq, _ := zmq.NewSocket(zmq.REQ)
        envlope.Pid=s.serverID       
	ms,_:=json.Marshal(*envlope)
        for j1:=range s.peers{
         j:=strconv.Itoa(j1+1)
         fmt.Println(j)
         
	 
         if (j1+1)!=s.serverID{
	 fmt.Println(j+":"+s.peerAddress[j])
         socketreq.Connect("tcp://"+s.peerAddress[j])
	 
         }
        
        
        }
         for j1:=range s.peers{
          if (j1+1)!=s.serverID{
          socketreq.Send(string(ms), 0)
          _,err:=socketreq.Recv(0)
	 if err == nil {
		fmt.Println("Ack rec",j1)
       	
	  }}
}
        //s.inbox<-&Envelope{int(s.serverAddress),2,m}
        
        
}



/*
func utilhandlefile() 
{
os.Create("output.txt")
f,_:= os.OpenFile("output.txt", os.O_APPEND|os.O_WRONLY, 0600)
return f
}*/

func main(){
s:=New(1,"./d.json")
s1:=New(2,"./d.json")
s2:=New(3,"./d.json")

//-------------------------tests----------------

//1.basic test send and receive
s.Outbox() <-&Envelope{Pid:2 , Msg: "hello there---2"}
s.Outbox() <-&Envelope{Pid:3 , Msg: "hello there1---3"}
s.Outbox() <-&Envelope{Pid:2 , Msg: "hello there2---2"}

//2.broadcast from every server to everyone else
s.Outbox() <-&Envelope{Pid:-1 /*cluster.BROADCAST*/, Msg: "broadcast from 1................................................."}
s1.Outbox() <-&Envelope{Pid:-1 /*cluster.BROADCAST*/, Msg: "broadcast from 2...................................................."}
s2.Outbox() <-&Envelope{Pid:-1 /*cluster.BROADCAST*/, Msg: "broadcast from 3.........................................................."}

//3.round--robin test
for i:=0;i<10;i++{
l:=strconv.Itoa(i)
msg:="round-robin msg no:"+l
s.Outbox() <-&Envelope{Pid:2, Msg: msg+" from 1"}
s1.Outbox() <-&Envelope{Pid:3, Msg: msg+" from 2"}
s2.Outbox() <-&Envelope{Pid:1, Msg: msg+" from 3"}
}
//4.send arbitrary long no of messages
for i:=0;i<200;i++{
Msg1:=strconv.Itoa(i)
s2.Outbox() <-&Envelope{Pid:1 /*cluster.BROADCAST*/, Msg: "hey 1--msg:"+Msg1}
//s2.Outbox() <-&Envelope{Pid:1 /*cluster.BROADCAST*/, Msg: "hello there1--msg:"+Msg1}
}


//UTIL CODE
os.Create("output.txt")
f,_:= os.OpenFile("output.txt", os.O_APPEND|os.O_WRONLY, 0600)
c:=0
for{
select{
       case envelope :=<-s.Inbox():
           c=c+1
           ss:=fmt.Sprintf("s---Received msg from %d: '%s'\n", envelope.Pid, envelope.Msg)
           f.WriteString(ss)
       case <-time.After(10 * time.Second): 
           fmt.Println("Waited and waited. Ab thak gaya\n")
}


select{
       case envelope :=<-s1.Inbox(): 
           ss1:=fmt.Sprintf("s1---Received msg from %d: '%s'\n", envelope.Pid, envelope.Msg)
           f.WriteString(ss1)
       case <-time.After(10 * time.Second): 
           fmt.Println("Waited and waited. Ab thak gaya\n")
}

select{
       case envelope :=<-s2.Inbox(): 
           ss2:=fmt.Sprintf("s2--Received msg from %d: '%s'\n", envelope.Pid, envelope.Msg)
           f.WriteString(ss2)
       case <-time.After(10 * time.Second): 
           fmt.Println("Waited and waited. Ab thak gaya\n")
}
}




}
/*
func (serv *Server) sendMsgto() int{
handleServerreq
*/
