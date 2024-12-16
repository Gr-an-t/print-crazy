package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/phin1x/go-ipp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)


func validateAPIKey(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if _, valid := apiKeys[apiKey]; !valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func leaderboardInsertHandler(client *mongo.Client, ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var requestData map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			http.Error(w, "Invalid JSON data", http.StatusBadRequest)
			return
		}

		data := bson.D{}
		for key, value := range requestData {
			data = append(data, bson.E{Key: key, Value: value})
		}

		mCreate(client, ctx, &data)

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Data inserted successfully"))
	}
}

func leaderboardUpdateHandler(client *mongo.Client, ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var requestData struct {
			Filter map[string]interface{} `json:"filter"` // MongoDB filter criteria
			Update map[string]interface{} `json:"update"` // Update data
		}

		// Parse and validate JSON input
		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			http.Error(w, "Invalid JSON data", http.StatusBadRequest)
			return
		}
		if len(requestData.Filter) == 0 || len(requestData.Update) == 0 {
			http.Error(w, "Filter and update fields cannot be empty", http.StatusBadRequest)
			return
		}

		// Convert input to BSON
		filter := bson.D{}
		for key, value := range requestData.Filter {
			filter = append(filter, bson.E{Key: key, Value: value})
		}
		update := bson.D{{Key: "$set", Value: requestData.Update}}

		// Call mUpdate to perform the update
		err = mUpdate(client, ctx, "your_collection", &filter, &update)
		if err != nil {
			if err.Error() == "no documents matched the filter" {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, "Error updating document", http.StatusInternalServerError)
				log.Println("Update error:", err)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Document updated successfully!"))
	}
}

func main() {
	mongoClient, ctx := connectToMongo()
	if mongoClient == nil || ctx == nil {
		fmt.Println("Failed to establish connection to MongoDB")
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello!"))
	})
	http.ListenAndServe(":8676", nil)

	http.HandleFunc("/leaderboardInsert", leaderboardInsertHandler(mongoClient, ctx))

	http.HandleFunc("/leaderboardUpdate", validateAPIKey(leaderboardUpdateHandler(mongoClient, ctx)))
	

	printClient := ipp.NewIPPClient("printserver", 631, "user", "password", true)
	// print file
	printClient.PrintFile("/path/to/file", "my-printer", map[string]interface{}{})
}
