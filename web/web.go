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
{"Name":"Neal", "Email":"ncharbonneau@gmail.com", "Note":"Using for VM x"}

DELETE
/status/<address>

*/
package web

import(
	"net"
	"net/http"
	"fmt"
	"io/ioutil"
	"github.com/nealjc/ipreg/scanner"
	"encoding/json"
	"log"
	"strings"
)

var serverControl chan bool
var database *DbConnection

func StartServer(status []*scanner.Subnet) {
	db, err := InitializeDB()
	if err != nil {
		log.Fatal("Failed to load database")
	}
	database = db
	serverControl = make(chan bool)
	statusPage := StatusPage{status, http.FileServer(http.Dir("."))}
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
	database.Close()
}

type StatusPage struct {
	Subnets []*scanner.Subnet
	fileHandler http.Handler
}

type AddressStatus struct {
	Address string
	Status string
	Name string
	Email string
	Note string
}

func (s *StatusPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	log.Printf("Request to %q", path[1])
	switch (path[1]) {
	case "subnets":
		s.handleSubnets(w, r)
	case "addresses":
		s.handleAddresses(w, r)
	case "status":
		log.Print("Handling status...")
		s.handleStatus(w, r)
		log.Print("Done status...")
	default:
		s.fileHandler.ServeHTTP(w, r)
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
	
	subnet := make([]AddressStatus, len(s.Subnets[subnetIndex].OrderedAddresses),
		len(s.Subnets[subnetIndex].OrderedAddresses))
	for i, address := range(s.Subnets[subnetIndex].OrderedAddresses) {
		reg, err := database.GetRegistration(address)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		subnet[i] = AddressStatus{address,
			formatStatus(s.Subnets[subnetIndex].Nodes[address]),
			reg.Name, reg.Email, reg.Note}
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
	// TODO: index
	address := strings.Split(r.URL.Path, "/")[2]
	if !s.validAddress(address) {
		log.Print("Invalid address")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	
	if r.Method == "GET" {
		log.Print("GET address")
		s.handleGetStatus(w, address)
	} else if r.Method == "PUT" {
		log.Print("PUT address")
		s.handlePutStatus(w, r, address)
	} else if r.Method == "DELETE" {
		log.Print("DELETE address")
		s.handleDeleteStatus(w, address)
	} else {
		log.Print("Bad method for address")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (s *StatusPage) validAddress(address string) bool {
	for _, subnet := range(s.Subnets) {
		if _, ok := subnet.Nodes[address]; ok {
			return true
		}
	}
	return false
}

func (s *StatusPage) handleGetStatus(w http.ResponseWriter, address string) {
	reg, err := database.GetRegistration(address)
	if err != nil {
		log.Printf("Failed to lookup registration %s", address)
		w.WriteHeader(http.StatusInternalServerError);
		return
	}
	for _, subnet := range(s.Subnets) {
		if status, ok := subnet.Nodes[address]; ok {
			withStatus := make(map[string]string)
			withStatus["Name"] = reg.Name
			withStatus["Email"] = reg.Email
			withStatus["Note"] = reg.Note
			withStatus["Status"] = formatStatus(status)
			b, err := json.Marshal(withStatus)
			if err != nil {
				log.Printf("Failed to marshal registration %s", address)
			}
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%s", b)
			return
		}
	}
}

func (s *StatusPage) handlePutStatus(w http.ResponseWriter, r *http.Request,
	address string) {
	log.Printf("User is updating reg for %s", address)
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body for %s", address)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var registration RegistrationRecord
	err = json.Unmarshal(data, &registration)
	if err != nil {
		log.Printf("Error unmarshaling request body for %s", address)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("Updating reg info %+v for  %s", registration, address)
	if database.SetRegistration(address, registration) {
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		log.Print("Address updated successfuly")
		return
	}
	log.Print("Address update failed")
	w.WriteHeader(http.StatusInternalServerError);
	
}

func (s *StatusPage) handleDeleteStatus(w http.ResponseWriter, address string) {
	log.Printf("User is deleting reg for %s", address)
	if database.DeleteRegistration(address) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
		return
	}
	w.WriteHeader(http.StatusInternalServerError);
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



