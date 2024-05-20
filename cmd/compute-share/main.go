package main

import (
	"fmt"
	"log"
	"net/http"
	"encoding/json"

	"compute-share/backend/pkg/rabbit_mq"
	"compute-share/backend/internal/handlers"
	"compute-share/backend/internal/kubernetes"
	"compute-share/backend/internal/models"
)

var (
	jobQueue *mq.RabbitMQ
)

func init() {
	jobQueue = mq.GetJobQueueSingleton()
}

// CORS Middleware
func enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}


func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/add-job", handlers.JobHandler)

	port := 8080
	addr := fmt.Sprintf(":%d", port)

	fmt.Printf("Server starting on port %d... \n", port)

	handler := enableCors(mux)

	go func() {
		if err := http.ListenAndServe(addr, handler); err != nil {
			log.Fatalf("Failed to start HTTP server: %s", err)
		}
	}()

	defer jobQueue.Close()

    messageHandler := func(msg []byte) {
        fmt.Printf("Processing message: %s\n", msg)
		var jobRequest models.Job
		json.Unmarshal(msg, &jobRequest)
		_, err := jobs.CreateKubernetesJob(jobRequest)
		if err != nil {
			log.Printf("Error creating job: %v", err)
		} else {
			go jobs.WatchJobStatus(jobRequest.Namespace, jobRequest.JobName)
		}
    }

    jobQueue.ConsumeMessages(messageHandler)

	select {}
}