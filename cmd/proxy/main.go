package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Health struct {
	Status string `json:"status"`
}

type ErrMsg struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Error      string `json:"error"`
}

func main() {

	conf := NewConfig()

	ipaddr := "0.0.0.0"

	log.Output(1, "starting proxy")
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatal("Error error getting interface IP addresses")

	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {

				//ipaddr = ipnet.IP.String()
				log.Output(1, "found IP "+ipaddr)
				break
			}
		}
	}

	origin, err := url.Parse(conf.ForwardHost())
	if err != nil {
		log.Fatal("Error reading configuration, exiting")
	}

	proxy := httputil.NewSingleHostReverseProxy(origin)

	// http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
	// 	health := Health{"ok"}
	// 	js, err := json.Marshal(health)
	// 	if err != nil {
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}
	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.Write(js)
	// })
	/*
		{
		statusCode: 404,
		message: "Cannot GET /dapr/config",
		error: "Not Found"
		}

	*/

	http.HandleFunc("/dapr/config", func(w http.ResponseWriter, r *http.Request) {
		health := ErrMsg{StatusCode: 404, Message: "Cannot GET /dapr/config", Error: "Not Found"}
		js, err := json.Marshal(health)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Output(1, "handling "+r.RequestURI)
		proxy.ServeHTTP(w, r)
	})

	log.Fatal(http.ListenAndServe(ipaddr+conf.ListenHostPort(), nil))
}
