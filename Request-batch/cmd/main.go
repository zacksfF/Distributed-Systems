package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// UserData represents data to be sent in a request
type UserData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func main() {
	// Define the API endpoint
	apiUrl := "http://localhost:8080/users"

	// Sample user data
	users := []UserData{
		{"1", "Zakaria"},
		{"2", "Saif"},
		{"3", "Memmmmm"},
	}

	// Batch size (adjust as needed)
	batchSize := 2

	for i := 0; i < len(users); i += batchSize {
		// Create a batch of user data
		var batchData []UserData
		for j := i; j < i+batchSize && j < len(users); j++ {
			batchData = append(batchData, users[j])
		}

		// Marshal batch data into JSON
		jsonData, err := json.Marshal(batchData)
		if err != nil {
			fmt.Println("Error marshalling data:", err)
			continue
		}

		// Create a new HTTP request
		req, err := http.NewRequest(http.MethodPost, apiUrl, bytes.NewReader(jsonData))
		if err != nil {
			fmt.Println("Error creating request:", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		// Send the request and handle response
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request:", err)
			continue
		}
		defer resp.Body.Close()

		// Process response (replace with your logic to handle the response)
		fmt.Println("Batch response status:", resp.StatusCode)
	}
}
