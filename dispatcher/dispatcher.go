package dispatcher

import (
	"fmt"
	"log"
	"sync"
)

type Listener func(Event)

type Dispatcher interface {
	RegisterListener(Event, Listener)
	Dispatch(Event)
	Aggregate(chan Event)
	Run()
	Exit()
}

type defaultDispatcher struct {
	listeners map[EventType][]Listener
	mutex     sync.Mutex
	EventChan chan Event
}

func New() *defaultDispatcher {
	return &defaultDispatcher{
		listeners: make(map[EventType][]Listener, 0),
		EventChan: make(chan Event, 0),
	}
}

func (d *defaultDispatcher) RegisterListener(e Event, l Listener) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if _, ok := d.listeners[e.GetEventType()]; !ok {
		d.listeners[e.GetEventType()] = make([]Listener, 0)
	}

	d.listeners[e.GetEventType()] = append(d.listeners[e.GetEventType()], l)
}

func (d *defaultDispatcher) Run() {
	go func() {
		for {
			select {
			case e, open := <-d.EventChan:
				if !open {
					log.Println("Exiting Dispatcher Run loop")
					return
				}

				d.Dispatch(e)
			}
		}
	}()
}

func (d *defaultDispatcher) Dispatch(e Event) {
	fmt.Println("Dispatch event ", e)
	if _, ok := d.listeners[e.GetEventType()]; !ok {
		return
	}

	for _, v := range d.listeners[e.GetEventType()] {
		v(e)
	}
}

//Enable event channel aggregation
func (d *defaultDispatcher) Aggregate(e chan Event) {
	fmt.Println("Aggregate ")
	go func() {
		for {
			select {
			case m, open := <-e:
				fmt.Println("Aggregate rcv ", m, open)
				if !open {
					return
				}
				d.EventChan <- m
			}
		}
	}()
}

func (d *defaultDispatcher) Exit() {
	close(d.EventChan)
}
