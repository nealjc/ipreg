/*
Package web is responsible for serving a REST API for retreiving IP status and reserving
IPs.

The available URLs are:

GETs
/subnets
/addresses/<subnet name>
/status/<address>

PUTs
/status/<address>

DELETE
/status/<address>

*/
package web

import(
	"net"
	"net/http"
	"fmt"
	"github.com/nealjc/ipreg/scanner"
	"encoding/json"
	"log"
	"strings"
)

func StartServer(status []*scanner.Subnet) {
	serverControl = make(chan bool)
	statusPage := StatusPage{status}
	server := http.Server{
		Addr: ":8080",
		Handler: nil,
	}
	http.Handle("/", &statusPage)
	l, e := net.Listen("tcp", ":8080")
	if e != nil {
		log.Fatal("Failed to start server")
		return
	}
	go server.Serve(l)
	log.Print("HTTP Server started")
	select {
	case <- serverControl:
		l.Close()
	}
	serverControl <- true
	log.Print("HTTP Server stopped")

}

func StopServer() {
	serverControl <- true
	<- serverControl
}

type StatusPage struct {
	Subnets []*scanner.Subnet
}

var serverControl chan bool

func (s *StatusPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	log.Printf("Request to %q", path[1])
	switch (path[1]) {
	case "subnets":
		s.handleSubnets(w, r)
	case "addresses":
		s.handleAddresses(w, r)
	case "status":
		s.handleStatus(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (s *StatusPage) handleSubnets(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// TODO: index out of bounds error
	subnets := make([][]string, len(s.Subnets), len(s.Subnets))
	for i, subnet := range(s.Subnets) {
		subnets[i] = make([]string, 2, 2)
		subnets[i][0] = subnet.Name
		subnets[i][1] = subnet.Network
	}
	b, err := json.Marshal(subnets)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", b)
}

func (s *StatusPage) handleAddresses(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// TODO: index out of bounds error
	subnetIndex := s.lookupSubnet(strings.Split(r.URL.Path, "/")[2])
	if subnetIndex == -1 {
		w.WriteHeader(http.StatusNotFound)
		return 
	}
	
	subnet := make([]string, len(s.Subnets[subnetIndex].OrderedAddresses),
		len(s.Subnets[subnetIndex].OrderedAddresses))
	for i, address := range(s.Subnets[subnetIndex].OrderedAddresses) {
		subnet[i] = address
	}
	b, err := json.Marshal(subnet)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", b)
}

func (s *StatusPage) lookupSubnet(name string) int {
	for i, subnet := range(s.Subnets) {
		if subnet.Name == name {
			return i
		}
	}
	return -1
}

func (s *StatusPage) handleStatus(w http.ResponseWriter, r *http.Request) {

	address := strings.Split(r.URL.Path, "/")[2]
	for _, subnet := range(s.Subnets) {
		if status, ok := subnet.Nodes[address]; ok {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%s", formatStatus(status))
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func formatStatus(status *scanner.NodeStatus) string {
	switch status.Status {
	case scanner.DOWN:
		return "Down"
	case scanner.UP:
		return "Up"
	case scanner.UNKNOWN:
		return "Unknown"
	default:
		return "Invalid state"
	}
}
