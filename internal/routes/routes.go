package routes

import (
	"log"
	"encoding/json"
	"compute-share/backend/internal/models"
	"compute-share/backend/pkg/rabbit_mq"
)

// Adds job to message queue
func AddJob(job models.Job) {
	jobJson, err := json.Marshal(job)
	if err != nil {
        log.Fatalf("Error serializing data: %s", err)
    }

	jobQueue := mq.GetJobQueueSingleton()
	// defer jobQueue.Close()

	jobQueue.PublishMessage(string(jobJson))
}