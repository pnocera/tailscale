package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tailscale.com/config"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	components "tailscale.com/pkg/components"
	configurations "tailscale.com/pkg/configurations"
	instances "tailscale.com/pkg/instances"
	kube "tailscale.com/pkg/kube"
	dashboard_log "tailscale.com/pkg/log"
	proxyapi "tailscale.com/pkg/proxyapi"
	"tailscale.com/pkg/version"
)

func main() {
	port := config.New().APIPort()
	RunWebServer(port)
}

var epoch = time.Unix(0, 0).Format(time.RFC1123)

var noCacheHeaders = map[string]string{
	"Expires":         epoch,
	"Cache-Control":   "no-cache, private, max-age=0",
	"Pragma":          "no-cache",
	"X-Accel-Expires": "0",
}

var etagHeaders = []string{
	"ETag",
	"If-Modified-Since",
	"If-Match",
	"If-None-Match",
	"If-Range",
	"If-Unmodified-Since",
}

var upgrader = websocket.Upgrader{} // use default options

// spaHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the given static directory.
type spaHandler struct {
	staticPath string
	indexPath  string
}

var inst instances.Instances
var comps components.Components
var configs configurations.Configurations
var proxys proxyapi.ProxyAPI

// RunWebServer starts the web server that serves the Dapr UI dashboard and the API
func RunWebServer(port int) {
	platform := ""
	kubeClient, daprClient, _ := kube.Clients()
	if kubeClient != nil {
		platform = "kubernetes"
	} else {
		platform = "standalone"
	}

	inst = instances.NewInstances(platform, kubeClient)
	comps = components.NewComponents(platform, daprClient)
	configs = configurations.NewConfigurations(platform, daprClient)
	proxys = proxyapi.NewProxyApi(kubeClient)

	r := mux.NewRouter()
	api := r.PathPrefix("/api/").Subrouter()
	api.HandleFunc("/features", getFeaturesHandler).Methods("GET")
	api.HandleFunc("/instances/{scope}", getInstancesHandler).Methods("GET")
	api.HandleFunc("/instances/{scope}/{id}", deleteInstancesHandler).Methods("DELETE")
	api.HandleFunc("/instances/{scope}/{id}", getInstanceHandler).Methods("GET")
	api.HandleFunc("/instances/{scope}/{id}/containers", getContainersHandler).Methods("GET")
	api.HandleFunc("/instances/{scope}/{id}/logstreams/{container}", getLogStreamsHandler)
	api.HandleFunc("/components/{scope}", getComponentsHandler).Methods("GET")
	api.HandleFunc("/components/{scope}/{name}", getComponentHandler).Methods("GET")
	api.HandleFunc("/deploymentconfiguration/{scope}/{id}", getDeploymentConfigurationHandler).Methods("GET")
	api.HandleFunc("/configurations/{scope}", getConfigurationsHandler).Methods("GET")
	api.HandleFunc("/configurations/{scope}/{name}", getConfigurationHandler).Methods("GET")
	api.HandleFunc("/controlplanestatus", getControlPlaneHandler).Methods("GET")
	api.HandleFunc("/metadata/{scope}/{id}", getMetadataHandler).Methods("GET")
	api.HandleFunc("/platform", getPlatformHandler).Methods("GET")
	api.HandleFunc("/scopes", getScopesHandler).Methods("GET")
	api.HandleFunc("/features", getFeaturesHandler).Methods("GET")
	api.HandleFunc("/version", getVersionHandler).Methods("GET")

	api.HandleFunc("/proxy/{scope}", getProxyHandler).Methods("GET")
	api.HandleFunc("/proxy/{scope}/proxied", getProxiedHandler).Methods("GET")
	api.HandleFunc("/proxy/{scope}/unproxy", unProxyHandler).Methods("GET")
	api.HandleFunc("/proxy/{scope}/{id}", setProxyHandler).Methods("GET")

	spa := spaHandler{staticPath: "/web", indexPath: "index.html"}
	r.PathPrefix("/").Handler(spa)

	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("0.0.0.0:%v", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	fmt.Printf("Proxy Dashboard running on http://localhost:%v\n", port)
	log.Fatal(srv.ListenAndServe())
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// get the volume of the absolute path and remove it
	volume := filepath.VolumeName(path)
	path = strings.Replace(path, volume, "", 1)

	// prepend the path with the path to the static directory
	path = filepath.Join(h.staticPath, path)

	// check whether a file exists at the given path
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	noCache(http.StripPrefix("/", http.FileServer(http.Dir(h.staticPath)))).ServeHTTP(w, r)
}

func getProxyHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scope := vars["scope"]
	resp := proxys.GetProxy(scope)
	respondWithJSON(w, 200, resp)
}

func getProxiedHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scope := vars["scope"]
	resp := proxys.Proxied(scope)
	respondWithJSON(w, 200, resp)
}

func unProxyHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scope := vars["scope"]
	resp := proxys.UnProxy(scope)
	respondWithJSON(w, 200, resp)
}

func setProxyHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appid := vars["id"]
	scope := vars["scope"]
	md := proxys.Proxy(scope, appid)
	respondWithJSON(w, 200, md)
}

func getInstancesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scope := vars["scope"]
	if scope == "All" {
		scope = ""
	}
	resp := inst.GetInstances(scope)
	respondWithJSON(w, 200, resp)
}

func getComponentsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scope := vars["scope"]
	if scope == "All" {
		scope = ""
	}
	resp := comps.GetComponents(scope)
	respondWithJSON(w, 200, resp)
}

func getComponentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scope := vars["scope"]
	if scope == "All" {
		scope = ""
	}
	name := vars["name"]
	resp := comps.GetComponent(scope, name)
	respondWithJSON(w, 200, resp)
}

func getFeaturesHandler(w http.ResponseWriter, r *http.Request) {
	features := []string{}
	if comps.Supported() {
		features = append(features, "components")
	}
	if configs.Supported() {
		features = append(features, "configurations")
	}
	if inst.CheckPlatform() == "kubernetes" {
		features = append(features, "status")
	}
	respondWithJSON(w, 200, features)
}

func getPlatformHandler(w http.ResponseWriter, r *http.Request) {
	resp := inst.CheckPlatform()
	respondWithPlainString(w, 200, resp)
}

func getContainersHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scope := vars["scope"]
	if scope == "All" {
		scope = ""
	}
	id := vars["id"]
	containers := inst.GetContainers(scope, id)
	respondWithJSON(w, 200, containers)
}

func getLogStreamsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scope := vars["scope"]
	if scope == "All" {
		scope = ""
	}
	id := vars["id"]
	container := vars["container"]
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer c.Close()
	reader, err := inst.GetLogStream(scope, id, container)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	lineReader := bufio.NewReader(reader)
	logReader := dashboard_log.NewReader(container, lineReader)
	for {
		logRecord, err := logReader.ReadLog()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if logRecord == nil {
			// Now wait some time before reading more.
			time.Sleep(time.Second)
			continue
		}

		bytes, err := json.Marshal(logRecord)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = c.WriteMessage(websocket.TextMessage, bytes)
		if err != nil {
			log.Println("fail to write log stream, aborting:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func getDeploymentConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scope := vars["scope"]
	if scope == "All" {
		scope = ""
	}
	id := vars["id"]
	details := inst.GetDeploymentConfiguration(scope, id)
	respondWithPlainString(w, 200, details)
}

func getConfigurationsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scope := vars["scope"]
	if scope == "All" {
		scope = ""
	}
	resp := configs.GetConfigurations(scope)
	respondWithJSON(w, 200, resp)
}

func getConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scope := vars["scope"]
	if scope == "All" {
		scope = ""
	}
	name := vars["name"]
	resp := configs.GetConfiguration(scope, name)
	respondWithJSON(w, 200, resp)
}

func getInstanceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scope := vars["scope"]
	if scope == "All" {
		scope = ""
	}
	id := vars["id"]
	resp := inst.GetInstance(scope, id)
	respondWithJSON(w, 200, resp)
}

func getControlPlaneHandler(w http.ResponseWriter, r *http.Request) {
	resp := inst.GetControlPlaneStatus()
	respondWithJSON(w, 200, resp)
}

func getMetadataHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scope := vars["scope"]
	if scope == "All" {
		scope = ""
	}
	id := vars["id"]
	md := inst.GetMetadata(scope, id)
	respondWithJSON(w, 200, md)
}

func deleteInstancesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scope := vars["scope"]
	if scope == "All" {
		scope = ""
	}
	id := vars["id"]
	err := inst.DeleteInstance(scope, id)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	respondWithJSON(w, 200, "")
}

func getScopesHandler(w http.ResponseWriter, r *http.Request) {
	resp := inst.GetScopes()
	respondWithJSON(w, 200, resp)
}

func getVersionHandler(w http.ResponseWriter, r *http.Request) {
	resp := version.GetVersion()
	respondWithPlainString(w, 200, resp)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithPlainString(w http.ResponseWriter, code int, payload string) {
	w.WriteHeader(code)
	_, err := w.Write([]byte(payload))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err := w.Write(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func noCache(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Delete any ETag headers that may have been set
		for _, v := range etagHeaders {
			if r.Header.Get(v) != "" {
				r.Header.Del(v)
			}
		}

		// Set our NoCache headers
		for k, v := range noCacheHeaders {
			w.Header().Set(k, v)
		}

		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
