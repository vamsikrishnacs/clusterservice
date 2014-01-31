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

/*
func (serv *Server) Sendmessagetopeer(env *Envelope)
{
serv.connect(s.Aeeraddressenv.pid)
serv.outbox<-&env
}
*/



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

s:=Server{id,c.Ser[id1],c.Peer,c.Ser,o,i}

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

func handleServerrep(s *Server){

        //1.point-point messaging

       
	socketrep, _ := zmq.NewSocket(zmq.REP)
        /*
        err := socket.Connect(PROTOCOL + addr)

    */
        socketrep.Bind("tcp://"+s.serverAddress)


        

       //socket.Bind("tcp://127.0.0.1:6000")
      
        for {
                
		msg, _ := socketrep.Recv(0)
		s.inbox<-&Envelope{1,1,msg}
                println("Got", string(msg))
                socketrep.Send("Ack ok", 0)
        }

}




func handleServerreq(s *Server/*,pid int*/){

         //10 msgs 
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
        //t:=strconv.Itoa(s.peers[peerid])
        t:=strconv.Itoa(peerid)
        socketreq.Connect("tcp://"+s.peerAddress[t])
        fmt.Println("s--msg from 1 to"+t+":"+s.peerAddress[t])
        socketreq.Send(envlope.Msg, 0)
	
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
        
        for j1:=range s.peers{
         j:=strconv.Itoa(j1+1)
         fmt.Println(j)
         
	 
         if (j1+1)!=s.serverID{
	 fmt.Println(j+":"+s.peerAddress[j])
         socketreq.Connect("tcp://"+s.peerAddress[j])
	 }
         }
        socketreq.Send(envlope.Msg, 0)
	
        fmt.Println("Sendingbrr")
        _,err:=socketreq.Recv(0)
	if err == nil {
		fmt.Println("Ack rec")
        }
        //s.inbox<-&Envelope{int(s.serverAddress),2,m}
        
        
}














/*
func (serv *Server) sendMsgto() int{
handleServerreq
*/

func main(){

s:=New(1,"d.json")
s1:=New(2,"d.json")
s2:=New(3,"d.json")
s.Outbox() <-&Envelope{Pid:2 /*cluster.BROADCAST*/, Msg: "hello there"}
s.Outbox() <-&Envelope{Pid:3 /*cluster.BROADCAST*/, Msg: "hello there1"}
s.Outbox() <-&Envelope{Pid:2 /*cluster.BROADCAST*/, Msg: "hello there2"}
s.Outbox() <-&Envelope{Pid:-1 /*cluster.BROADCAST*/, Msg: "hello there3"}



for i:=0;i<20;i++{
Msg1:=strconv.Itoa(i)
s.Outbox() <-&Envelope{Pid:3 /*cluster.BROADCAST*/, Msg: "hey server3--msg:"+Msg1}
s1.Outbox() <-&Envelope{Pid:1 /*cluster.BROADCAST*/, Msg: "hello there1--msg:"+Msg1}
}
for{
select{
       case envelope :=<-s.Inbox(): 
           fmt.Printf("s---Received msg from %d: '%s'\n", envelope.Pid, envelope.Msg)
  
       case <-time.After(10 * time.Second): 
           fmt.Println("Waited and waited. Ab thak gaya\n")
}


select{
       case envelope :=<-s1.Inbox(): 
           fmt.Printf("s1---Received msg from %d: '%s'\n", envelope.Pid, envelope.Msg)
  
       case <-time.After(10 * time.Second): 
           fmt.Println("Waited and waited. Ab thak gaya\n")
}

select{
       case envelope :=<-s2.Inbox(): 
           fmt.Printf("s2--Received msg from %d: '%s'\n", envelope.Pid, envelope.Msg)
  
       case <-time.After(10 * time.Second): 
           fmt.Println("Waited and waited. Ab thak gaya\n")
}
}
/*fmt.Println("s....in")
fmt.Println(s.Inbox())
fmt.Println("s....out")
fmt.Println(s.Outbox())/*
fmt.Println("s1....in")
fmt.Println(s1.Inbox())
fmt.Println(s1.Outbox())
fmt.Println("go")*/
}
