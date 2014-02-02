clusterservice
==============

An api for managing servers in a cluster with point-point and broadcast messaging


1.Implemented point-point messaging
2.Broadcast messages
3.dynamic configuration of multiple servers using a json


Testing:

tested with many messages scenario
tested with broadcasting(everyone to everyone else)
tested with round-robin message passing


package usage:

two functions
New(id int,path //(.json config file) string )
Envelope{pid int(-1 for broadcast),message}


//sample
 server := cluster.New(myid, /* config file */"config.json")
   // the returned server object obeys the Server interface above.
  
   wait for keystroke to start.
   // Let each server broadcast a message
   server.Outbox() <- &cluster.Envelope{Pid: cluster.BROADCAST, Msg: "hello there"}
  
   select {
       case envelope := <- case server.Inbox(): 
           fmt.Printf("Received msg from %d: '%s'\n", envelope.Pid, envelope.Msg)
  
       case <- time.After(10 * time.Second): 
           println("Waited and waited. Ab thak gaya\n")
   }


General description:

1. Maintains a struct object for each server
2. Uses Zeromq req/rep sockets for message passing
3. Needs a supply of configuration of list of servers and address in json format


System check:

A count of no of meesages received can be seen in the log in the command line
log pattern:
msg request
send
got msg--(count of last msg)
ack
ok
