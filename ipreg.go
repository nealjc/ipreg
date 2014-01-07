package main

import(
	"net"
	"os/signal"
	"os"
	"log"
	"github.com/nealjc/netmonitor/web"
	"github.com/nealjc/netmonitor/scanner"
)


func generateAddresses(rangeStart, rangeEnd string) (results []string) {
	startIP := net.ParseIP(rangeStart).To4()
	endIP := net.ParseIP(rangeEnd).To4()
	for i := startIP[3]; i <= endIP[3]; i++ {
		addr := net.IPv4(startIP[0],
			startIP[1], startIP[2], i).String()
		results = append(results, addr)
	}
	return
}

func waitForSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<- c
	log.Println("Got signal, exiting")
}

func main() {
	addresses := generateAddresses("192.168.1.1", "192.168.1.254")
	nodeStatus := make(map[string]*scanner.NodeStatus)
	for _, addr := range(addresses) {
		nodeStatus[addr] = &scanner.NodeStatus{scanner.UNKNOWN}
	}
	go scanner.StartScanner(nodeStatus)
	go web.StartServer(nodeStatus)
	waitForSignal()
	web.StopServer()
	scanner.StopScanner()
}







