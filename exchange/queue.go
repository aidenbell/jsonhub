/*
The Queue package.
Queues are bound to an exchange and the exchange routes messages in to the queue. Queues can
distribute messages to and define clients and their semantics.
*/
package exchange

import (
	"log"
	"encoding/json"
	"fmt"
	"reflect"

	// Some built-in matchers
	"github.com/aidenbell/jsonhub/match_modules/ext_ci_match"
	"github.com/aidenbell/jsonhub/match_modules/ext_geojson"
)


type Exchanger interface {
	Send(Messager)
	AddQueue( *Queue )
	RemoveQueue( *Queue )
}


type DistType int

/*
WIP Distribution types. Clients on the same
queue get given messages based on the distribution
type
*/
const (
	DistBroadcast DistType = iota
	DistRandom
	DistRoundRobin
)

/*
A basic queue that accepts a list of clients and matches messages against a match specification.
The queue has various configuration options.
*/
type Queue struct {
	Exchange Exchanger
	In chan Messager
	Clients []chan Messager
	MatchSpec string
	newClients chan chan Messager
	deadClients chan chan Messager
	exitChan chan int
	
	// Queue options
	pingOnly bool			// You get empty messages, not the src message. Useful for counting
	distMethod DistType		// What distribution of messages to clients?
}

/*
 * We define sensible defaults on new queues
 * then allow configuration via methods
 */
func NewQueue(e Exchanger, spec string) *Queue {
	// Create a queue for the client
	q := Queue{
		e,
		make(chan Messager),
		make([]chan Messager,0),
		spec,
		make(chan chan Messager),
		make(chan chan Messager),
		make(chan int),
		false,
		DistBroadcast} // TODO: Make configurable
		
	return &q
}

// If true, clients get an empty message or "ping" when a message matches rather
// than the complete source message.
func (q *Queue) SetPingOnly(p bool) {
	q.pingOnly = p
}

// Getter for ping only setting
func(q *Queue) PingOnly() bool {
	return q.pingOnly;
}

// Set the distribution method of the queue, such as broadcast, round-robin or random
func (q *Queue) SetDistMethod(d DistType) {
	q.distMethod = d
}

// Getter for the distribution method
func (q *Queue) DistMethod() DistType {
	return q.distMethod;
}

/*
A goroutine for managing the client list.
reads clients joining and leaving the queue from channels
and modifies the list.

It also reads messages from the In chan and sends those
messages to clients based on the DistMethod of the queue.
 */
func(q *Queue) clientMgr() {
	for {
		select {
		case c := <- q.newClients:
			log.Printf("Client (o_o) %p\n",c)
			q.Clients = append(q.Clients, c)
			
		case d := <- q.deadClients:
			for i,v := range q.Clients {
				if v == d {
					log.Printf("Client (x_x) %p\n",d)
					q.Clients = append(q.Clients[:i], q.Clients[i+1:]...)
				}
			}
			if len(q.Clients) == 0 {
				log.Printf("Queue (x_x) %p\n",q)
				q.Exchange.RemoveQueue(q)			
				return;
			}
		
		case m := <- q.In:
			// Basic "send to all"
			for _,c := range q.Clients {
				select {
				case c <- m:
				default:
				}
			}
		}
	}
}

/*
Add a client to the queue in the form of a channel being read by some client
handling code. A "client" to the queue is just a channel accepting messages.
*/
func(q *Queue) AddClient(c chan Messager) {
	q.newClients <- c
}


/*
Remove a client from the queue. The client will not get any more messages.
It is up to the client handling code to cleanup any resources.
*/
func(q *Queue) RemoveClient(c chan Messager) {
	q.deadClients <- c
}

/*
Start the asynchronous process that starts the queue consuming messages
and distributing them to clients.
*/
func(q *Queue) Run() {
	go q.clientMgr()
}


func(q *Queue) MessageMatches( m Messager ) bool {
	// parse the message JSON
	parsed := map[string]interface{}{}
	err := json.Unmarshal([]byte(m.Raw()), &parsed)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", m.Raw(), err)
		return false
	}

	// Parse the spec JSON
	// TODO: Move to queue creation
	parsed_spec := map[string]interface{}{}
	err = json.Unmarshal([]byte(q.MatchSpec), &parsed_spec)
	
	if err != nil {
		fmt.Println("Error parsing spec JSON:", err);
		return false
	}

	return matchObject(parsed, parsed_spec)
}

func matchObject(parsed map[string]interface{}, parsed_spec map[string]interface{}) bool {

	// Run the comparitors
	matches := true		// Does this message attribute match the spec?

	for k,sv := range parsed_spec {
		v := parsed[k]

		// If there is an attribute in the spec that isn't in
		// the message, that counts as a fail. The spec definition implies "attribute exists"
		if v == nil {
			return false
		}

		svtype := reflect.TypeOf(sv)
		vtype := reflect.TypeOf(v)


		// Switch on spec value type, defaulting to basic Go comparison
		switch svcast := sv.(type) {
			// Spec value is an object, either extension
			// or plain object comparison
			case map[string]interface{}:
				if ext, ok := svcast["__match__"]; ok {
					// Spec is an extension, so delegate to that
					// extension for match result
					extstr := ext.(string)
					switch extstr {
					case "case-insensitive-match":
						matches = ext_ci_match.ExtCaseInsensitiveMatch(v,svcast)
					case "geojson-within":
						matches = ext_geojson.ExtGeoJSONWithin(v,svcast)	
					default:
						return false
					}
				} else {
					// Spec is a plain object, ensure type match
					// and compare objects
					if svtype != vtype {
						return false
					}
					vcast := sv.(map[string]interface{})
					matches = matchObject(vcast, svcast)
				}
				break

			// Spec val isn't an object, just do type comparison
			// and value comparison for matching types.
			default:
				// Check that the types match firstly
				if svtype != vtype {
					return false
				}
				// Types match, compare values
				matches = (v == sv)
		}

		if matches == false {
			break; // Don't keep processing message attrs if we get a false
		}
	}

	return matches
}
