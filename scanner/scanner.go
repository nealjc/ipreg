package scanner

import(
	"net"
	"log"
	"time"
	"syscall"
	"sync/atomic"
)

type ScanResult struct {
	Address AddrToSubnet
	IsUp bool
}

const (
	UP = iota
	DOWN
	UNKNOWN
)

type NodeStatus struct {
	Status int32
}

func (s *NodeStatus) UpdateStatus(status int32) {
	atomic.StoreInt32(&s.Status, status)
}

type Subnet struct {
	Name string
	Network string
	Nodes map[string]*NodeStatus
	OrderedAddresses []string
}

type AddrToSubnet struct {
	Address string
	Subnet *Subnet
}

func NewSubnet(name, network string) *Subnet {
	subnet := Subnet{name, network,
		make(map[string]*NodeStatus),
		make([]string, 0, 255)} // 255 probably most common
	return &subnet
}

func checkHost(address AddrToSubnet, rateLimiter chan bool) ScanResult {
	<- rateLimiter
	defer func() {
		rateLimiter <- true
	}();
	con, err := net.DialTimeout("tcp", net.JoinHostPort(address.Address, "22"),
		time.Duration(5) * time.Second)
	if err != nil {
		if netErr, ok := err.(*net.OpError); ok {
			if netErr.Err == syscall.ECONNREFUSED {
				return ScanResult{address, true}
			}
		}
		return ScanResult{address, false}
	} else {
		con.Close()
	}
	
	return ScanResult{address, true}
}

func scanNetwork(addresses []AddrToSubnet, maxParallel int) <- chan ScanResult {
	results := make(chan ScanResult)
	rateLimiter := make(chan bool, maxParallel)
	for i := 0; i < maxParallel; i++ {
		rateLimiter <- true
	}

	for _, address := range(addresses) {
		go func(addr AddrToSubnet) {
			results <- checkHost(addr, rateLimiter)
		}(address)
	}
	return results
}

var scannerControl chan bool

func runScan(toScan []*Subnet, maxJobs int) {
	totalNumAddresses := 0
	for _, subnet := range(toScan) {
		totalNumAddresses += len(subnet.Nodes)
	}
	addresses := make([]AddrToSubnet, totalNumAddresses)
	i := 0
	for _, subnet := range(toScan) {
		for addr := range(subnet.Nodes) {
			addresses[i] = AddrToSubnet{addr, subnet}
			i++
		}
	}
	resultChannel := scanNetwork(addresses, maxJobs)
	timeout := time.After(time.Duration(1) * time.Minute)
	for i := 0 ; i < len(addresses); i++ {
		select {
		case <- timeout:
			// TODO: should fill in UNKNOWN or something for remaining
			log.Print("Scan timed out")
			break
		case r := <- resultChannel:
			subnet := r.Address.Subnet
			status := subnet.Nodes[r.Address.Address]
			if r.IsUp {
				status.UpdateStatus(UP)
			} else {
				status.UpdateStatus(DOWN)
			}
		case <- scannerControl:
			log.Print("Scan cancelled for shutdown")
			// propagate the cancel upwards
			scannerControl <- true
			break
		}
	
	}
}

func StartScanner(toScan []*Subnet, timeBetweenScans, maxJobs int) {
	scannerControl = make(chan bool)
	done := false
	for {
		log.Println("Running network scanner");
		runScan(toScan, maxJobs)
		log.Println("Scan finished. Waiting for next scan...");
		timeout := time.After(time.Duration(timeBetweenScans)*time.Minute)
		select {
		case <- timeout:
			continue
		case <- scannerControl:
			done = true
		}
		if done {
			break
		}
	}
	scannerControl <- true
}

func StopScanner() {
	log.Print("Stopping scanner")
	scannerControl <- true
	<- scannerControl
	log.Print("Scanner stopped")
}
