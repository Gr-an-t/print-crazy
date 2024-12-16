package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/phin1x/go-ipp"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var apiKeys map[string]bool


func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKeys = make(map[string]bool)
	apiKeys[os.Getenv("API_KEY_1")] = true
	apiKeys[os.Getenv("API_KEY_2")] = true
}



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
		log.Println("Inserting data into leaderboard...")

		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var requestData struct {
			Name string `json:"name"`
		}

		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			http.Error(w, "Invalid JSON data", http.StatusBadRequest)
			return
		}

		collection := client.Database("leaderboardDB").Collection("leaderboard")

		// Check if the name already exists
		var existingEntry struct {
			Name  string `bson:"name"`
			Score int    `bson:"score"`
			Cost  int    `bson:"cost"`
		}

		err = collection.FindOne(ctx, bson.M{"name": requestData.Name}).Decode(&existingEntry)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				// Name doesn't exist, insert a new entry
				newEntry := bson.D{
					{Key: "name", Value: requestData.Name},
					{Key: "score", Value: 0},
					{Key: "cost", Value: 0},
				}

				_, err := collection.InsertOne(ctx, newEntry)
				if err != nil {
					log.Println(err)
					http.Error(w, "Failed to insert data into database", http.StatusInternalServerError)
					return
				}
			} else {
				// Error finding the document
				log.Println(err)
				http.Error(w, "Failed to query database", http.StatusInternalServerError)
				return
			}
		} else {
			// Name exists, update the score and cost
			update := bson.D{
				{Key: "$inc", Value: bson.D{
					{Key: "score", Value: 1},
					{Key: "cost", Value: 1},
				}},
			}

			_, err := collection.UpdateOne(ctx, bson.M{"name": requestData.Name}, update)
			if err != nil {
				log.Println(err)
				http.Error(w, "Failed to update data in database", http.StatusInternalServerError)
				return
			}
		}

		// Recalculate ranks
		err = recalculateRanks(collection, ctx)
		if err != nil {
			http.Error(w, "Failed to recalculate ranks", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Data processed and ranks updated successfully"))
	}
}


func recalculateRanks(collection *mongo.Collection, ctx context.Context) error {
	cursor, err := collection.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{Key: "score", Value: -1}}))
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	rank := 1
	for cursor.Next(ctx) {
		var entry bson.M
		if err := cursor.Decode(&entry); err != nil {
			return err
		}

		_, err := collection.UpdateOne(
			ctx,
			bson.D{{Key: "_id", Value: entry["_id"]}},
			bson.D{{Key: "$set", Value: bson.D{{Key: "rank", Value: rank}}}},
		)
		if err != nil {
			return err
		}

		rank++
	}

	return nil
}

func leaderboardReadHandler(client *mongo.Client, ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		collection := client.Database("leaderboardDB").Collection("leaderboard")

		cursor, err := collection.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{Key: "rank", Value: 1}}))
		if err != nil {
			http.Error(w, "Failed to fetch leaderboard data", http.StatusInternalServerError)
			return
		}
		defer cursor.Close(ctx)

		var leaderboard []bson.M
		if err := cursor.All(ctx, &leaderboard); err != nil {
			http.Error(w, "Failed to decode leaderboard data", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(leaderboard)
		if err != nil {
			http.Error(w, "Failed to encode leaderboard data", http.StatusInternalServerError)
			return
		}
	}
}

func leaderboardUpdateHandler(client *mongo.Client, ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var requestData struct {
			Filter map[string]interface{} `json:"filter"` 
			Update map[string]interface{} `json:"update"` 
		}

		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			http.Error(w, "Invalid JSON data", http.StatusBadRequest)
			return
		}
		if len(requestData.Filter) == 0 || len(requestData.Update) == 0 {
			http.Error(w, "Filter and update fields cannot be empty", http.StatusBadRequest)
			return
		}

		filter := bson.D{}
		for key, value := range requestData.Filter {
			filter = append(filter, bson.E{Key: key, Value: value})
		}
		update := bson.D{{Key: "$set", Value: requestData.Update}}

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

func sendPrintHandler(client *mongo.Client, ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	var requestData struct {
		Message string `json:"message"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	log.Printf("Print job initiated: %s", requestData.Message)


	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Print job initiated successfully"))
	}
}

func main() {
	mongoClient, ctx := connectToMongo()
	if mongoClient == nil || ctx == nil {
		log.Println("Failed to establish connection to MongoDB")
		return
	}
	log.Println("Successfully connected to MongoDB!")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		protocol := "http"
		if r.TLS != nil {
			protocol = "https"
		}
		fullAddress := fmt.Sprintf("%s://%s%s", protocol, r.Host, r.URL.Path)
		log.Printf("Received request: %s", fullAddress)
	})

	http.HandleFunc("/leaderboardInsert", validateAPIKey(leaderboardInsertHandler(mongoClient, ctx)))
	http.HandleFunc("/leaderboardUpdate", validateAPIKey(leaderboardUpdateHandler(mongoClient, ctx)))
	http.HandleFunc("/leaderboardRead", validateAPIKey(leaderboardReadHandler(mongoClient, ctx)))
	http.HandleFunc("/sendPrint", validateAPIKey(sendPrintHandler(mongoClient, ctx)))

	serverAddress := ":8676"
	log.Printf("Server is starting and will listen on %s", serverAddress)

	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"}, 
		AllowedHeaders: []string{"Content-Type", "X-API-Key"}, 
		AllowCredentials: true, 
	})

	handler := corsHandler.Handler(http.DefaultServeMux)

	if err := http.ListenAndServe(serverAddress, handler); err != nil {
		log.Fatalf("Failed to start server on %s: %v", serverAddress, err)
	}

	printClient := ipp.NewIPPClient("printserver", 631, "user", "password", true)
	printClient.PrintFile("/path/to/file", "my-printer", map[string]interface{}{})

	select {} 
}