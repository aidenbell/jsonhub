package exchange

import "fmt"

/*
 * Exchange
 */
type Exchange struct {
	Name       string
	Queues     []*Queue
	Exit       chan int
	In         chan Messager
	newQueues  chan *Queue
	deadQueues chan *Queue
}

/*
 * Exchange constructor
 */
func NewExchange() *Exchange {
	return &Exchange{
		"TestExchange",
		make([]*Queue, 0),
		make(chan int),
		make(chan Messager),
		make(chan *Queue),
		make(chan *Queue)}
}

/*
 * The QueueMgr basically listens for data coming in on various
 * channels that signals management operations for the queues on
 * the exchange.
 */
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
