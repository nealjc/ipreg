package scanner

import(
	"net"
	"log"
	"time"
	"syscall"
	"sync/atomic"
)

type ScanResult struct {
	Address string
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

func checkHost(address string) ScanResult {
	con, err := net.DialTimeout("tcp", net.JoinHostPort(address, "22"),
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

func scanNetwork(addresses []string) <- chan ScanResult {
	results := make(chan ScanResult)

	for _, address := range(addresses) {
		go func(addr string) {
			results <- checkHost(addr)
		}(address)
	}
	return results
}

var scannerControl chan bool

func runScan(toScan map[string]*NodeStatus) {
	addresses := make([]string, len(toScan))
	i := 0
	for k := range(toScan) {
		addresses[i] = k
		i++
	}
	resultChannel := scanNetwork(addresses)
	timeout := time.After(time.Duration(1) * time.Minute)
	for i := 0 ; i < len(addresses); i++ {
		select {
		case <- timeout:
			// TODO: should fill in UNKNOWN or something for remaining
			log.Print("Scan timed out")
			break
		case r := <- resultChannel:
			status := toScan[r.Address]
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

func StartScanner(toScan map[string]*NodeStatus) {
	scannerControl = make(chan bool)
	for {
		log.Println("Running network scanner");
		runScan(toScan)
		timeout := time.After(time.Duration(10)*time.Minute)
		select {
		case <- timeout:
			continue
		case <- scannerControl:
			break
		}
	}
	scannerControl <- true
}

func StopScanner() {
	scannerControl <- true
	<- scannerControl
	log.Print("Scanner stopped")
}








