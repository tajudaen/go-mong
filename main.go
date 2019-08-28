package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Person Model(Schema)
type Person struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Firstname string             `json:"firstname,omitempty" bson:"firstname,omitempty"`
	Lastname  string             `json:"lastname,omitempty" bson:"lastname,omitempty"`
}

var client *mongo.Client

func createPersonEndpoint(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "application/json")
	var person Person
	json.NewDecoder(req.Body).Decode(&person)
	collection := client.Database("tajgo").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, _ := collection.InsertOne(ctx, person)
	json.NewEncoder(res).Encode(result)
}

func getPeopleEndpoint(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "application/json")
	var people []Person
	collection := client.Database("tajgo").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(`{"message": "` + err.Error() + `"}`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var person Person
		cursor.Decode(&person)
		people = append(people, person)
	}
	if err := cursor.Err(); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(`{"message": "` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(res).Encode(people)
}

func getPersonEndpoint(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "application/json")
	params := mux.Vars(req)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var person Person
	collection := client.Database("tajgo").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := collection.FindOne(ctx, Person{ID: id}).Decode(&person)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(`{"message": "` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(res).Encode(person)
}

func updatePersonEndpoint(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "application/json")
	params := mux.Vars(req)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var person Person
	collection := client.Database("tajgo").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	_ = json.NewDecoder(req.Body).Decode(&person)
	person.ID = id
	err := collection.FindOneAndUpdate(
		ctx,
		bson.D{
			bson.E{"_id", id},
		},
		bson.D{bson.E{"$set", person}}).Decode(&person)

	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(`{"message": "` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(res).Encode(person)
}

func main() {
	fmt.Println("starting application")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)
	router := mux.NewRouter()
	router.HandleFunc("/person", createPersonEndpoint).Methods("POST")
	router.HandleFunc("/person", getPeopleEndpoint).Methods("GET")
	router.HandleFunc("/person/{id}", getPersonEndpoint).Methods("GET")
	router.HandleFunc("/person/{id}", updatePersonEndpoint).Methods("PUT")
	http.ListenAndServe(":9090", router)
}
