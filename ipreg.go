package main

import(
	"net"
	"os/signal"
	"os"
	"log"
	"github.com/nealjc/ipreg/web"
	"github.com/nealjc/ipreg/scanner"
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

func readSubnets() (subnets []*scanner.Subnet, e error) {
	// TODO: read from config file
	fromConfig := "192.168.1.0/24"
	_, ipNet, e := net.ParseCIDR(fromConfig)
	if e != nil {
		return nil, e
	}
	sub := scanner.NewSubnet("First Subnet", ipNet.IP.String(),
		ipNet.Mask.String())
	subnets = append(subnets, sub)
	return
}

func main() {
	subnets, e := readSubnets()
	if e != nil {
		log.Fatal(e.Error())
		return
	}
	log.Printf("%s %s %s", subnets[0].Name, subnets[0].Network, subnets[0].Netmask)
	// TODO: replace generateAddresses/nodeStatus with readSubnets results
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
