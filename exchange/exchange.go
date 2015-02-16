// Package exchange defines the core mechanics of jsonhub. It defined what
// an exchange is, how Queues have messages delivered to them and how we push
// messages to an Exchange.
//
// All of the functionality available through the jsonhubd command should be
// available by easily composing types from this package (or types implementing
// their interfaces).
package exchange

import (
	"fmt"
	"log"
)

// DistType is the data type for specifying a method of distributing
// a message to clients in a ClientPool
type DistType int

// WIP Distribution types. Clients on the same
// queue get given messages based on the distribution
// type
const (
	DistBroadcast  DistType = iota // Send messages to all clients on the queue
	DistRandom                     // Send to a random client
	DistRoundRobin                 // Balance across clients
)

// A subscriber is something that can Receive messages
type Subscriber interface {
	Receive(Messager)
}

// A publisher is something that accepts Subscribers and you can publish
// message through
type Publisher interface {
	Subscribe(Subscriber)
	Unsubscribe(Subscriber)
	Publish(Messager)
}

// Something that can act as both a Publisher and Subscriber
type PubSuber interface {
	Publisher
	Subscriber
}

type Exchange struct {
	subscribers []Subscriber
	exit        chan int
	in          chan Messager
	pushSub     chan Subscriber
	popSub      chan Subscriber
	distMethod  DistType // What distribution of messages to clients?

}

// NewExchange creates a basic broadcasting exchange
func NewExchange() *Exchange {
	e := &Exchange{
		make([]Subscriber, 0),
		make(chan int),
		make(chan Messager),   // Our input queue
		make(chan Subscriber), // New subscribers
		make(chan Subscriber),
		DistBroadcast} // Unsubscribers
	go e.run()
	return e
}

// The QueueMgr basically listens for data coming in on various
// channels that signals management operations for the queues on
// the exchange.
func (e *Exchange) run() {
	log.Printf("%p(%T) running", e, e)
	for {
		select {
		case sub := <-e.pushSub:
			e.subscribers = append(e.subscribers, sub)

		case unsub := <-e.popSub:
			for i, sub := range e.subscribers {
				if sub == unsub {
					e.subscribers = append(e.subscribers[:i], e.subscribers[i+1:]...)
				}
			}

		case m := <-e.in:
			for _, sub := range e.subscribers {
				log.Printf("%p loop sending to Subscriber %T", e, sub)
				sub.Receive(m)
			}
		}
	}
}

// Subscribe adds a Subscriber to the Exchange
func (e *Exchange) Subscribe(sub Subscriber) {
	// This guard isn't really needed?
	e.pushSub <- sub
	log.Printf("%p got subscriber %T(%p)\n", e, sub, sub)
}

// Unsubscribe removes a Subscriber from the Exchange
func (e *Exchange) Unsubscribe(sub Subscriber) {
	e.popSub <- sub
}

// Publish sends a message to the Exchange. In the case of Exchange
// this simply broadcasts the message to all connected Subscribers
// in a broadcast style.
func (e *Exchange) Publish(m Messager) {
	select {
	case e.in <- m:
		fmt.Printf("Exchange.Publish() got %T -> %s\n", m, m.Raw())
	default:
	}
}

// SetDistMethod allows configuring the way messages are distributed to clients
// in this pool
func (q *Exchange) SetDistMethod(d DistType) {
	q.distMethod = d
}

// DistMethod is the getter for the current distribution method
func (q *Exchange) DistMethod() DistType {
	return q.distMethod
}
