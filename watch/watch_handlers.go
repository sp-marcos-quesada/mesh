package watch

import (
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/peer"
	"github.com/marcosQuesada/mesh/router/handler"
	"log"
)

func (w *defaultWatcher) Handlers() map[message.MsgType]handler.Handler{
	return map[message.MsgType]handler.Handler{
		message.PING:   w.HandlePing,
		message.PONG:   w.HandlePong,
	}
}


func (w *defaultWatcher) HandlePing(c peer.NodePeer, msg message.Message) (message.Message, error) {
	ping := msg.(*message.Ping)
	log.Println("Handle Ping", ping.Id, "from: ", ping.From.String())

	return &message.Pong{Id: ping.Id, From: ping.To, To: ping.From}, nil
}

func (w *defaultWatcher) HandlePong(c peer.NodePeer, msg message.Message) (message.Message, error) {
	pong := msg.(*message.Pong)
	log.Println("Handle Pong ", pong.Id, c.Node(), "from: ", pong.From.String())
	requestID := w.requestListener.Id(c.Node(), pong.Id)
	go w.requestListener.Notify(msg, requestID)

	return nil, nil
}
