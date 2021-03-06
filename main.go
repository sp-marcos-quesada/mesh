package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/marcosQuesada/mesh/config"
	"github.com/marcosQuesada/mesh/server"
	"fmt"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	//Parse config
	addr := flag.String("addr", "127.0.0.1:12000", "Port where Mesh is listen on")
	cluster := flag.String("cluster", "127.0.0.1:12000,127.0.0.1:12001,127.0.0.1:12002", "cluster list definition separated by commas")
	flag.Parse()

	//Init logger
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	f, err := os.OpenFile(fmt.Sprintf("%s.log", *addr), os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Panic("error opening file: %v", err)
	}
	defer f.Close()

	//log.SetOutput(f)

	//Create Configuration
	config := config.NewConfig(*addr, *cluster)

	//Define server from config
	s := server.New(config)
	c := make(chan os.Signal, 1)

	signal.Notify(
		c,
		os.Kill,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	//serve until signal
	go func() {
		<-c
		s.Close()
	}()

	//Server Run
	s.Start()
}
