package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"time"
)

type Advertisement struct {
	ID primitive.ObjectID `json:"_id,omitempty" bson: "_id,omitempty"`
	Name string `json:"name,omitempty" bson:"name,omitempty"`
	Description string `json: "description,omitempty" bson:"description,omitempty"`
	Links []string `json: "links,omitempty" bson:"links,omitempty"`

}

type Advertisements []Advertisement

func allAdvertisements(response http.ResponseWriter, request *http.Request){
	advertisements := Advertisements{
		Advertisement{Name: "Test ad", Description: "Test desc", Links: []string{"First links", "Second Link"}},
	}

	fmt.Println("Endpoint: All advertisements")
	json.NewEncoder(response).Encode(advertisements)

}

func homePage(w http.ResponseWriter, r *http.Request)  {
	fmt.Fprint(w, "Homepage Endpoint Hit")
	
}

func addAdvertisement(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var advertisement Advertisement
	json.NewDecoder(request.Body).Decode(&advertisement)
	collection := client.Database("advertisements").Collection("advertisements")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, _ := collection.InsertOne(ctx, advertisement)
	json.NewEncoder(response).Encode(result)

}

func handleRequest()  {

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homePage)
	router.HandleFunc("/advertisements", allAdvertisements)
	router.HandleFunc("/add_advertisement", addAdvertisement)
	log.Fatal(http.ListenAndServe(":8001", router))
}

var client *mongo.Client



func main() {

	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(
		"mongodb+srv://DimDimi4:stepler32@cluster0.xwf9a.mongodb.net/advertisements?retryWrites=true&w=majority",
	))
	if (err == nil) {
		print("Database is connected")
	}
	handleRequest()





}