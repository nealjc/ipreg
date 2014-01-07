/*
Package web is responsible for serving a REST API for retreiving IP status and reserving
IPs.
*/
package web

import(
	"net"
	"net/http"
	"fmt"
	"github.com/nealjc/netmonitor/scanner"
	"encoding/json"
	"log"
)

type StatusPage struct {
	StatusMap map[string]*scanner.NodeStatus
}

var serverControl chan bool

func (s *StatusPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(s.StatusMap)
	if err != nil {
		fmt.Fprintf(w, "Error")
	}
	fmt.Fprintf(w, "%s", b)
}

func StartServer(status map[string]*scanner.NodeStatus) {
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














