package main

import(
	"net"
	"os/signal"
	"os"
	"log"
	"github.com/nealjc/ipreg/web"
	"github.com/nealjc/ipreg/scanner"
)


func waitForSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<- c
	log.Println("Got signal, exiting")
}

func generateAllInSubnet(ipNet *net.IPNet, subnet *scanner.Subnet) {
	inc := func (ip net.IP) {
		for j:= len(ip)-1; j >=0; j-- {
			ip[j]++
			if ip[j] > 0 {
				break
			}
		}
	}
	for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); inc(ip) {
		subnet.Nodes[ip.String()] = &scanner.NodeStatus{scanner.UNKNOWN}
		subnet.OrderedAddresses = append(subnet.OrderedAddresses, ip.String())
	}
}

func readSubnets() (subnets []*scanner.Subnet, e error) {
	// TODO: read from config file
	fromConfig := "192.168.1.0/24"
	_, ipNet, e := net.ParseCIDR(fromConfig)
	if e != nil {
		return nil, e
	}
	sub := scanner.NewSubnet("First Subnet", ipNet.String())
	generateAllInSubnet(ipNet, sub)
	subnets = append(subnets, sub)
	return
}

func main() {
	subnets, e := readSubnets()
	if e != nil {
		log.Fatal(e.Error())
		return
	}
	go scanner.StartScanner(subnets)
	go web.StartServer(subnets)
	waitForSignal()
	web.StopServer()
	scanner.StopScanner()
}
