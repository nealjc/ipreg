package config

import (
	"net"
	"log"
	"github.com/nealjc/ipreg/scanner"
	"code.google.com/p/gcfg"
)

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

type Params struct {
	TimeBetweenScansInMinutes int
	MaxParallelJobs int
	ListenPort int
	DatabaseDir string
	HtmlDir string
}

func ParseConfig(inputFile string) (subnets []*scanner.Subnet, params Params, e error) {
	type Config struct {
		Parameters Params
		Subnet map[string]*struct {
			Network string
		}
	}
	config := Config{}
	e = gcfg.ReadFileInto(&config, inputFile)
	if e != nil {
		return nil, Params{}, e
	}

	params = config.Parameters
	for subnetName, network := range(config.Subnet) {
		log.Printf("Adding subnet %s %s", subnetName, network.Network)
		_, ipNet, e := net.ParseCIDR(network.Network)
		if e != nil {
			return nil, Params{}, e
		}
		sub := scanner.NewSubnet(subnetName, ipNet.String())
		generateAllInSubnet(ipNet, sub)
		subnets = append(subnets, sub)
	}
	return
}
