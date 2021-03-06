package server

import (
	"log"
	"net"

	"github.com/marcosQuesada/mesh/cli"
	"github.com/marcosQuesada/mesh/cluster"
	"github.com/marcosQuesada/mesh/config"
	"github.com/marcosQuesada/mesh/dispatcher"
	//"github.com/marcosQuesada/mesh/message"
	n "github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
	"github.com/marcosQuesada/mesh/router"
)

type Server struct {
	config *config.Config
	node   n.Node
	exit   chan bool
	router router.Router
}

func New(c *config.Config) *Server {
	return &Server{
		config: c,
		exit:   make(chan bool),
		node:   c.Addr,
		router: router.New(c.Addr),
	}
}

func (s *Server) Start() {
	c := cluster.Start(s.node, s.config.Cluster)

	disp := dispatcher.New()
	disp.RegisterListener(&peer.OnPeerConnectedEvent{}, c.OnPeerConnectedEvent)
	disp.RegisterListener(&peer.OnPeerDisconnectedEvent{}, c.OnPeerDisconnected)
	disp.RegisterListener(&peer.OnPeerAbortedEvent{}, c.OnPeerAborted)
	disp.RegisterListener(&peer.OnPeerErroredEvent{}, c.OnPeerErrored)
	go disp.ConsumeEventChan()
	go disp.AggregateChan(s.router.Events())

	s.router.RegisterHandlersFromInstance(c)
	//aggregate coordinator snd chan
	go s.router.AggregateChan(c.SndChan())

	s.startDialPeers()
	s.startServer()
	s.run()

	//TODO: aggregate watcher snd chan
	//s.router.AggregateChan(c.SndChan())

}

func (s *Server) Close() {
	//TODO: fix dispatcher Shutdown
	//d.Exit()

	close(s.exit)
}

func (s *Server) run() {
	for {
		select {
		case <-s.exit:
			//Notify Exit to remote Peer
			return
		}
	}
}

func (s *Server) startServer() {
	listener, err := net.Listen("tcp", string(s.config.Addr.String()))
	if err != nil {
		log.Println("Error starting Socket Server: ", err)
		return
	}

	go s.startAcceptorPeers(listener)
}

func (s *Server) startAcceptorPeers(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error starting socket client to: ", s.node.String(), "err: ", err)
			return
		}

		c := peer.NewAcceptor(conn, s.node)
		go c.Run()

		s.router.Accept(c)
	}
}

//Start Dial Peers
func (s *Server) startDialPeers() {
	for _, node := range s.config.Cluster {
		//avoid local connexion
		if node.String() == s.node.String() {
			continue
		}

		go s.router.InitDialClient(node)
	}
}

// Cli Socket server
func (s *Server) startCliServer() error {
	listener, err := net.Listen("tcp", s.config.Addr.String())
	if err != nil {
		log.Println("Error Listening Cli Server")
		return err
	}
	go func() error {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("Error Accepting")
				return err
			}
			defer conn.Close()
			go s.handleCliConnection(conn)
		}
	}()

	return nil
}

// Socket Client access
func (s *Server) handleCliConnection(conn net.Conn) {
	defer conn.Close()

	c := &cli.CliSession{
		Conn: conn,
		//server: s,
	}
	c.Handle()
}
