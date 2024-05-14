package handlers

import (
	// "compute-share/backend/internal/db"
	"compute-share/backend/internal/routes"
	"compute-share/backend/internal/models"
	"encoding/json"
	// "fmt"
	"net/http"
)

func JobHandler(w http.ResponseWriter, r *http.Request) {
	if (r.Method != "POST") {
		http.Error(w, "Invalid method", http.StatusBadRequest)
		return
	}
	// Authorize request with id_token - removed for easy server testing

	// ctx := r.Context()
	// authClient, err := config.App.Auth(ctx)
	// if err != nil {
	// 	http.Error(w, "Authentication error", http.StatusInternalServerError)
	// 	return
	// }
	// idToken := r.Header.Get("Authorization")
	// if idToken == "" {
	// 	http.Error(w, "No token provided", http.StatusBadRequest)
	// 	return
	// }
	// token, err := authClient.VerifyIDToken(ctx, idToken)
	// if err != nil {
	// 	http.Error(w, "Invalid token", http.StatusUnauthorized)
	// 	return
	// }

	// uid := token.UID;
	uid := "ISU9Srb06iXE2P2DYk15K5UrZnh2"
	var job models.Job
	err := json.NewDecoder(r.Body).Decode(&job)
    if err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
	job.BuyerId = uid

    if job.JobId == "" {
        http.Error(w, "Invalid or missing job_id or disk_space", http.StatusBadRequest)
        return
    }
	routes.AddJob(job)

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Job added successfully"))
}
