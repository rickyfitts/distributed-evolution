package master

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rickyfitts/distributed-evolution/go/api"
	"github.com/rickyfitts/distributed-evolution/go/util"
)

// handler for POST /job requests
// abandons current job and start on the new one
func (m *Master) newJob(w http.ResponseWriter, r *http.Request) {
	log.Printf("##### New Job Request - %v #####", http.MethodOptions)
	// Allow CORS here By * or specific origin
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")

	// ignore preflight request
	if r.Method == http.MethodOptions {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	log.Printf("starting new job")

	// decode request body as job config
	var job api.Job
	err := json.NewDecoder(r.Body).Decode(&job)
	if err != nil {
		log.Printf("error decoding new job request body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// decode and save target image
	m.TargetImageBase64 = job.TargetImage
	img, err := util.DecodeImage(job.TargetImage)
	if err != nil {
		log.Printf("error decoding target image")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	m.setTargetImage(img)

	// save the job with a new ID
	newID := m.Job.ID + 1
	m.Job = job
	m.Job.ID = newID
	m.Job.TargetImage = "" // no need to be passing it around, its saved on m

	m.generateTasks()

	response := State{
		NumWorkers:       m.NumWorkers,
		Tasks:            m.Tasks,
		ThreadsPerWorker: m.ThreadsPerWorker,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handles requests from the ui and websocket communication
func (m *Master) httpServer() {
	r := mux.NewRouter()

	r.HandleFunc("/job", m.newJob).Methods(http.MethodPost, http.MethodOptions)

	r.HandleFunc("/subscribe", m.subscribe)

	port := os.Getenv("HTTP_PORT")

	log.Printf("listening for HTTP on port %v\n", port)

	r.Use(mux.CORSMethodMiddleware(r))

	log.Fatal(http.ListenAndServe(":"+port, r))
}
