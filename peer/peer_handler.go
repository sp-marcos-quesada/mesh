package peer

import (
	"fmt"
	"github.com/marcosQuesada/mesh/message"
	n "github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/watch"
	"log"
	"sync"
	"time"
)

//PeerHandler is in charge of Handle Client Lifecycle
//Sends pings on ticker to check remote state
type PeerHandler interface {
	BootClient(n.Node) //message.Status
	AcceptClient(NodePeer) message.Status
	Route(message.Message)

	Notify(n.Node, error) //Used to get notifications of Client conn failures
	Peers() map[string]NodePeer
	Len() int
}

type defaultPeerHandler struct {
	watcher watch.Watcher
	peers   map[string]NodePeer
	mutex   sync.Mutex
	from    n.Node
}

func DefaultPeerHandler(node n.Node) *defaultPeerHandler {
	return &defaultPeerHandler{
		watcher: watch.New(),
		peers:   make(map[string]NodePeer),
		from:    node,
	}
}

func (n *defaultPeerHandler) BootClient(node n.Node) {
	log.Println("Orchestrartor boot client", node)

	var c *Peer
	log.Println("Starting Dial Client on Node ", n.from.String(), "destination: ", node.String())
	//Blocking call, wait until connection success
	c = NewDialer(n.from, node)
	c.Run()

	//Say Hello and wait response
	c.SayHello()

	select {
	case <-time.NewTimer(time.Second).C:
		log.Println("Client has not receive response, Timeout")
		return
	case rsp := <-c.ReceiveChan():
		switch rsp.(type) {
		case *message.Welcome:
			log.Println("Client has received Welcome from", node.String(), rsp.(*message.Welcome))
			err := n.accept(c)
			if err != nil {
				log.Println("Error Accepting Peer, Peer dies! ", err)
				/*				n.inChan <- memberUpdate{
								node:  node,
								event: PeerStatusError,
							}*/
				return
			} else {
				//o.clients[node.String()] = c
				/*				o.inChan <- memberUpdate{
								node:  node,
								event: PeerStatusConnected,
							}*/

				//aggregate receiveChan to mainChan
				//o.aggregate(c.ReceiveChan())
				log.Println("Client Achieved: ", node)
			}
		case *message.Abort:
			log.Println("Response Abort ", rsp.(*message.Abort), " remote node:", node.String())
			/*			o.inChan <- memberUpdate{
						node:  node,
						event: PeerStatusError,
					}*/
		default:
			log.Println("Unexpected type On response ")
		}
	}

}

func (n *defaultPeerHandler) AcceptClient(p NodePeer) (response message.Status) {
	select {
	case msg := <-p.ReceiveChan():
		log.Println("Msg Received ", msg)

		switch msg.(type) {
		case *message.Hello:
			p.Identify(msg.(*message.Hello).From)
			err := n.accept(p)
			if err != nil {
				p.Send(&message.Abort{Id: msg.(*message.Hello).Id, From: n.from, Details: map[string]interface{}{"foo_bar": 1231}})
				p.Exit()

				response = PeerStatusAbort
			}
			p.Send(&message.Welcome{Id: msg.(*message.Hello).Id, From: n.from, Details: map[string]interface{}{"foo_bar": 1231}})

			response = PeerStatusConnected
		case *message.Ping:
			log.Println("Router Ping: ", msg.(*message.Ping))
			p.Send(&message.Pong{Id: msg.(*message.Ping).Id, From: n.from, Details: map[string]interface{}{}})
		}
	}

	return
}

func (h *defaultPeerHandler) Notify(n n.Node, err error) {

}

func (h *defaultPeerHandler) Peers() map[string]NodePeer {
	return h.peers
}

func (h *defaultPeerHandler) Route(m message.Message) {

}

func (h *defaultPeerHandler) Len() int {
	return len(h.peers)
}
func (h *defaultPeerHandler) accept(p NodePeer) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	node := p.Node()
	if _, ok := h.peers[node.String()]; ok {
		return fmt.Errorf("Peer: %s Already registered", node.String())
	}
	h.peers[node.String()] = p
	fmt.Println("XX Accepted Peer type:", p.Mode(), " from: ", p.Node())
	return nil
}

func (h *defaultPeerHandler) remove(p NodePeer) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	node := p.Node()
	if _, ok := h.peers[node.String()]; !ok {
		return fmt.Errorf("Peer Not found")
	}

	delete(h.peers, node.String())

	return nil
}
