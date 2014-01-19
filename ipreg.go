package main

import(
	"os/signal"
	"os"
	"log"
	"github.com/nealjc/ipreg/web"
	"github.com/nealjc/ipreg/scanner"
	"github.com/nealjc/ipreg/config"
)

func main() {
	subnets, params, e := config.ParseConfig("/etc/ipreg.conf")
	if e != nil {
		log.Fatal(e.Error())
		return
	}
	go scanner.StartScanner(subnets, params.TimeBetweenScans,
		params.MaxJobs)
	if err := web.Initialize(subnets, params.ListenPort); err != nil {
		log.Fatal(err)
	}
	go web.StartServer()
	waitForSignal()
	web.StopServer()
	scanner.StopScanner()
}

func waitForSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<- c
	log.Println("Got signal, exiting")
}
