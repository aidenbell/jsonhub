// The exchange package defines the core mechanics of jsonhub. It defined what
// an exchange is, how Queues have messages delivered to them and how we push
// messages to an Exchange.
//
// All of the functionality available through the jsonhubd command should be
// available by easily composing types from this package (or types implementing
// their interfaces).
package exchange

import "fmt"

type Exchange struct {
	Name       string
	Queues     []*Queue
	Exit       chan int
	In         chan Messager
	newQueues  chan *Queue
	deadQueues chan *Queue
}

func NewExchange() *Exchange {
	return &Exchange{
		"TestExchange",
		make([]*Queue, 0),
		make(chan int),
		make(chan Messager),
		make(chan *Queue),
		make(chan *Queue)}
}

// The QueueMgr basically listens for data coming in on various
// channels that signals management operations for the queues on
// the exchange.
func (e *Exchange) QueueMgr() {
	for {
		select {
		case q := <-e.newQueues:
			e.Queues = append(e.Queues, q)

		case d := <-e.deadQueues:
			for i, v := range e.Queues {
				if v == d {
					e.Queues = append(e.Queues[:i], e.Queues[i+1:]...)
				}
			}

		case m := <-e.In:
			for _, q := range e.Queues {
				if q.MessageMatches(m) == true {
					// TODO: Make q.Send() and move to there.
					if q.PingOnly() {
						pingMsg := NewJsonMessage("{}") // Empty ping message.
						select {
						case q.In <- pingMsg:
						default:
						}
					} else {
						select {
						case q.In <- m:
						default:
						}
					}
				}
			}
		}
	}
}

func (e *Exchange) AddQueue(q *Queue) {
	// This guard isn't really needed?
	select {
	case e.newQueues <- q:
	default:
	}
}

func (e *Exchange) RemoveQueue(q *Queue) {
	e.deadQueues <- q
}

func (e *Exchange) Run() {
	go e.QueueMgr()
}

func (e *Exchange) Send(m Messager) {

	select {
	case e.In <- m:
		fmt.Printf("Exchange.Send() got %T -> %s\n", m, m.Raw())
	default:
	}
}
