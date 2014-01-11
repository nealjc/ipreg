package main

import(
	"net"
	"os/signal"
	"os"
	"log"
	"github.com/nealjc/ipreg/web"
	"github.com/nealjc/ipreg/scanner"
	"code.google.com/p/gcfg"
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
	inputFile := "./config.txt"
	type Config struct {
		Parameters struct {
			TimeBetweenScansInMinutes int
		}
		Subnet map[string]*struct {
			Network string
		}
	}
	config := Config{}
	e = gcfg.ReadFileInto(&config, inputFile)
	if e != nil {
		return nil, e
	}

	for subnetName, network := range(config.Subnet) {
		log.Printf("Adding subnet %s %s", subnetName, network.Network)
		_, ipNet, e := net.ParseCIDR(network.Network)
		if e != nil {
			return nil, e
		}
		sub := scanner.NewSubnet(subnetName, ipNet.String())
		generateAllInSubnet(ipNet, sub)
		subnets = append(subnets, sub)
	}
	return
}

func main() {
	// TODO: require config file input
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
