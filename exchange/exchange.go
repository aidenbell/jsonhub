// The exchange package defines the core mechanics of jsonhub. It defined what
// an exchange is, how Queues have messages delivered to them and how we push
// messages to an Exchange.
//
// All of the functionality available through the jsonhubd command should be
// available by easily composing types from this package (or types implementing
// their interfaces).
package exchange

import "fmt"

// A subscriber is something that can Receive messages
type Subscriber interface {
	Receive(Messager)
}

// A publisher is something that accepts Subscribers and you can publish
// message through
type Publisher interface {
	Subscribe(Publisher)
	Unsubscribe(Publisher)
}

// Something that can act as both a Publisher and Subscriber
type PubSuber interface {
	Publisher
	Subscriber
}

type Exchange struct {
	subscribers     []Subscriber
	exit       chan int
	in         chan Messager
	pushSub  	 chan Subscriber
	popSub chan Subscriber
}

func NewExchange() *Exchange {
	return &Exchange{
		make([]Subscriber, 0),
		make(chan int),
		make(chan Messager),				// Our input queue
		make(chan Subscriber), 			// New subscribers
		make(chan Subscriber)}			// Unsubscribers
}

// The QueueMgr basically listens for data coming in on various
// channels that signals management operations for the queues on
// the exchange.
func (e *Exchange) QueueMgr() {
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
				sub.Receive(m)
			}
		}
	}
}

func (e *Exchange) Subscribe(sub Subscriber) {
	// This guard isn't really needed?
	select {
	case e.pushSub <- sub:
	default:
	}
}

func (e *Exchange) Unsubscribe(sub Publisher) {
	e.popSub <- sub
}

func (e *Exchange) Run() {
	go e.QueueMgr()
}

func (e *Exchange) Publish(m Messager) {
	select {
	case e.In <- m:
		fmt.Printf("Exchange.Publish() got %T -> %s\n", m, m.Raw())
	default:
	}
}
