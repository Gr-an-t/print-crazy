package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func mCreate(client *mongo.Client, ctx context.Context, data *bson.D) {
	collection := client.Database("your_database").Collection("your_collection")

	_, err := collection.InsertOne(ctx, data)

	if err != nil {

		log.Fatal(err)

	}

	fmt.Println("Data inserted successfully!")

}

func mRead(client *mongo.Client, ctx context.Context, collectionName string) {
	collection := client.Database("your_database").Collection(collectionName)

	cursor, err := collection.Find(ctx, bson.D{})

	if err != nil {

		log.Fatal(err)

	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {

		var result bson.M

		err := cursor.Decode(&result)

		if err != nil {

			log.Fatal(err)

		}

		fmt.Println(result)

	}

	if err := cursor.Err(); err != nil {

		log.Fatal(err)

	}

}

func mUpdate(client *mongo.Client, ctx context.Context, collectionName string, filter *bson.D, update *bson.D) error {
	collection := client.Database("your_database").Collection(collectionName)

	// Perform the update operation
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	// Check if a document was matched
	if result.MatchedCount == 0 {
		return fmt.Errorf("no documents matched the filter")
	}

	fmt.Printf("Matched %d document(s) and modified %d document(s)\n", result.MatchedCount, result.ModifiedCount)
	return nil
}

func mDelete(client *mongo.Client, ctx context.Context, collectionName string, filter *bson.D) {
	collection := client.Database("your_database").Collection(collectionName)

	_, err := collection.DeleteOne(ctx, filter)

	if err != nil {

		log.Fatal(err)

	}

	fmt.Println("Data deleted successfully!")

}

func connectToMongo() (*mongo.Client, context.Context) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://your_username:your_password@localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	fmt.Println("Successfully connected to MongoDB!")
	return client, ctx
}