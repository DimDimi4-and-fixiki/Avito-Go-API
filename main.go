package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/go-playground/validator.v10"
	"log"
	"net/http"
	"time"
)

type Advertisement struct {
	ID 			primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name 		string `json:"name,omitempty" bson:"name,omitempty" validate:"name"`
	Price 		float64 `json:"price,omitempty" bson:"price, omitempty" validate:"gte=0"`
	Description string `json:"description,omitempty" bson:"description,omitempty" validate:"description"`
	Links 		[]string `json:"links,omitempty" bson:"links,omitempty" validate:"links"`

}

type Advertisements []Advertisement


func homePage(w http.ResponseWriter, r *http.Request)  {
	fmt.Fprint(w, "Homepage Endpoint Hit")
	
}

func validateLinks(fl validator.FieldLevel) bool {
	l := fl.Field().Len()
	if l > 3 {
		return false
	}
	return true
}

func validateDescription(fl validator.FieldLevel) bool {
	MAX_LEN := 1000  // Max len of a description
	l := len(fl.Field().String())
	if l > MAX_LEN {
		return false
	}
	return true
}

func validateName(fl validator.FieldLevel) bool {
	MAX_LEN := 200 // Max len for a Name
	l := len(fl.Field().String())
	if l > MAX_LEN {
		return false
	}
	return true
}

func validateAdvertisement(advertisement *Advertisement) error {
	validate := validator.New()
	validate.RegisterValidation("links", validateLinks)
	validate.RegisterValidation("description", validateDescription)
	validate.RegisterValidation("name", validateName)
	return validate.Struct(advertisement)
}




func addAdvertisement(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var advertisement Advertisement
	json.NewDecoder(request.Body).Decode(&advertisement)
	err := validateAdvertisement(&advertisement)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "Error": "` + err.Error() + `"}`))
		return
	}
	collection := client.Database("advertisements").Collection("advertisements")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, _ := collection.InsertOne(ctx, advertisement)
	json.NewEncoder(response).Encode(result)

}

func getAdvertisement (response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var advertisement Advertisement
	collection := client.Database("advertisements").Collection("advertisements")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := collection.FindOne(ctx, Advertisement{ID: id}).Decode(&advertisement)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(response).Encode(advertisement)
}

func getAdvertisements (response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var advertisements []Advertisement
	collection := client.Database("advertisements").Collection("advertisements")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var advertisement Advertisement
		cursor.Decode(&advertisement)
		advertisements = append(advertisements, advertisement)

	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(response).Encode(advertisements)
}


func handleRequests()  {

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homePage)
	router.HandleFunc("/add_advertisement", addAdvertisement).Methods("POST")
	router.HandleFunc("/advertisements", getAdvertisements).Methods("GET")
	router.HandleFunc("/advertisement/{id}", getAdvertisement).Methods("GET")
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
	handleRequests()





}