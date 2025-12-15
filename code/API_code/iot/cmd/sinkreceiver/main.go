/*
Copyright (C) 2022-2025 Contributors | TIM S.p.A. to CAMARA a Series of LF Projects, LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// Mock AM data response structure
type AccessAndMobilitySubscriptionData struct {
	SubsRegTimer int `json:"subsRegTimer"`
	ActiveTime   int `json:"activeTime"`
}

// Mock PP data update request structure
type PpDataUpdate struct {
	PpData *PpDataPayload `json:"ppData"`
}

type PpDataPayload struct {
	CommunicationCharacteristics *CommunicationCharacteristics `json:"communicationCharacteristics"`
}

type CommunicationCharacteristics struct {
	PpMaximumLatency      *string `json:"ppMaximumLatency,omitempty"`
	PpMaximumResponseTime *string `json:"ppMaximumResponseTime,omitempty"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleCallback)
	mux.HandleFunc("/nudm-sdm/v2/", handleGetAMData)
	mux.HandleFunc("/nudm-pp/v1/", handleSetPPData)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Sink receiver started on port %s\n", port)
		log.Printf("Mock endpoints available:")
		log.Printf("  GET  /nudm-sdm/v2/{supi}/am-data")
		log.Printf("  PATCH /nudm-pp/v1/{ueId}/pp-data")
		log.Printf("  POST / (callback receiver)")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v\n", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := server.Close(); err != nil {
		log.Printf("Error closing server: %v\n", err)
	}
}

// handleGetAMData mocks GET /nudm-sdm/v2/{supi}/am-data
func handleGetAMData(w http.ResponseWriter, r *http.Request) {
	log.Printf("=== Mock AM Data Request ===")
	log.Printf("Method: %s, URL: %s", r.Method, r.URL.String())

	// Extract supi from URL path
	// URL format: /nudm-sdm/v2/{supi}/am-data
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		log.Printf("Invalid URL format")
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	supi := pathParts[2]

	log.Printf("SUPI: %s", supi)

	// Check if supi contains @fail.com - return 404
	if strings.Contains(supi, "@fail.com") {
		log.Printf("SUPI contains @fail.com - returning 404")
		http.Error(w, "Subscriber not found", http.StatusNotFound)
		return
	}

	// Return mock AM data
	response := AccessAndMobilitySubscriptionData{
		SubsRegTimer: 3600,
		ActiveTime:   10,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	log.Printf("Returned mock AM data: subsRegTimer=%d, activeTime=%d", response.SubsRegTimer, response.ActiveTime)
	log.Println("============================")
}

// handleSetPPData mocks PATCH /nudm-pp/v1/{ueId}/pp-data
func handleSetPPData(w http.ResponseWriter, r *http.Request) {
	log.Printf("=== Mock PP Data Update Request ===")
	log.Printf("Method: %s, URL: %s", r.Method, r.URL.String())

	// Extract ueId from URL path
	// URL format: /nudm-pp/v1/{ueId}/pp-data
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		log.Printf("Invalid URL format")
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	ueId := pathParts[2]

	log.Printf("UE ID: %s", ueId)

	// Check if ueId contains @fail.com - return 404
	if strings.Contains(ueId, "@fail.com") {
		log.Printf("UE ID contains @fail.com - returning 404")
		http.Error(w, "UE not found", http.StatusNotFound)
		return
	}

	// Read and log request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	log.Printf("Request body: %s", string(body))

	// Try to parse and log the PP data
	var ppUpdate PpDataUpdate
	if err := json.Unmarshal(body, &ppUpdate); err == nil {
		if ppUpdate.PpData != nil && ppUpdate.PpData.CommunicationCharacteristics != nil {
			cc := ppUpdate.PpData.CommunicationCharacteristics
			log.Printf("PP Data Update:")
			if cc.PpMaximumLatency != nil {
				log.Printf("  ppMaximumLatency: %s", *cc.PpMaximumLatency)
			}
			if cc.PpMaximumResponseTime != nil {
				log.Printf("  ppMaximumResponseTime: %s", *cc.PpMaximumResponseTime)
			}
		}
	}

	// Return 204 No Content on success
	w.WriteHeader(http.StatusNoContent)
	log.Printf("PP Data updated successfully - returning 204")
	log.Println("===================================")
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	log.Printf("=== Received callback ===")
	log.Printf("Method: %s", r.Method)
	log.Printf("URL: %s", r.URL.String())
	log.Printf("Remote Address: %s", r.RemoteAddr)

	log.Println("Headers:")
	for name, values := range r.Header {
		for _, value := range values {
			log.Printf("  %s: %s", name, value)
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	log.Printf("Body (raw): %s", string(body))

	// Try to pretty print JSON if valid
	if len(body) > 0 {
		var prettyJSON map[string]interface{}
		if err := json.Unmarshal(body, &prettyJSON); err == nil {
			formatted, _ := json.MarshalIndent(prettyJSON, "", "  ")
			log.Printf("Body (formatted):\n%s", string(formatted))

			// Check for error event type
			if eventType, ok := prettyJSON["type"].(string); ok {
				if strings.Contains(eventType, ".error") {
					log.Printf("⚠️  ERROR EVENT DETECTED: %s", eventType)
					if data, ok := prettyJSON["data"].(map[string]interface{}); ok {
						log.Printf("   Transaction ID: %v", data["transactionId"])
						log.Printf("   Error Code: %v", data["code"])
						log.Printf("   Error Message: %v", data["message"])
						log.Printf("   Status: %v", data["status"])
					}
				}
			}
		}
	}

	log.Println("========================")

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}
