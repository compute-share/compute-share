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

func main() {
	http.HandleFunc("/api/add-job", handlers.JobHandler)
	port := 8080
	addr := fmt.Sprintf(":%d", port)

	fmt.Printf("Server starting on port %d... \n", port)
	// log.Fatal(http.ListenAndServe(addr, nil))
	go func() {
        if err := http.ListenAndServe(addr, nil); err != nil {
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
			log.Printf("error unmarshalling job request: %v", err)
		} else {
			go jobs.WatchJobCompletion(jobRequest.Namespace, jobRequest.JobName)
		}
        
		
    }

    jobQueue.ConsumeMessages(messageHandler)

	select {}
}