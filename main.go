package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/go-playground/validator/v10"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Advertisement struct {
	ID 			primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name 		string `json:"name,omitempty" bson:"name,omitempty" validate:"name"`
	Price 		float64 `json:"price,omitempty" bson:"price,omitempty" validate:"price"`
	Description string `json:"description,omitempty" bson:"description,omitempty" validate:"description"`
	Links 		[]string `json:"links,omitempty" bson:"links,omitempty" validate:"links"`
	CreatedAt   int64 `json:"created_at,omitempty" bson:"created_at,omitempty" `

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

func validatePrice(fl validator.FieldLevel) bool {
	if fl.Field().Float() > 0 {
		return true
	} else {
		return false
	}
}

func validateAdvertisement(advertisement *Advertisement) error {
	validate := validator.New()
	validate.RegisterValidation("links", validateLinks)
	validate.RegisterValidation("description", validateDescription)
	validate.RegisterValidation("name", validateName)
	validate.RegisterValidation("price", validatePrice)
	return validate.Struct(advertisement)
}




func addAdvertisement(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var advertisement Advertisement
	json.NewDecoder(request.Body).Decode(&advertisement)
	err := validateAdvertisement(&advertisement)
	advertisement.CreatedAt = int64(time.Now().Unix())
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
	addDescription := false  // Flag to Add description to a response
	addLinks := false  // Flag to Add all links to a response

	// Sets type of a response to JSON:
	response.Header().Add("content-type", "application/json")
	params := mux.Vars(request)  // gets Params from URL
	id, _ := primitive.ObjectIDFromHex(params["id"]) // ID of an Ad
	values := request.URL.Query() // gets all Values from the URL
	fields, ok := values["fields"]  // Extracts fields from the URL
	if ok { // Fields parameter was specified
		for _, field := range fields {
			if field == "description" {
				addDescription = true  // We need to add description
			}
			if field == "links" {
				addLinks = true  // We need to add all links
			}

		}
	}

	var advertisement Advertisement
	collection := client.Database("advertisements").Collection("advertisements")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := collection.FindOne(ctx, Advertisement{ID: id}).Decode(&advertisement)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	advertisement.CreatedAt = 0  // Removes CreatedAt field from
	if addLinks && addDescription {
		// Sends full info about Advertisement:
		json.NewEncoder(response).Encode(advertisement)
		return
	} else if addLinks {
		advertisement.Description = ""
		json.NewEncoder(response).Encode(advertisement)
		return
	} else if addDescription {
		advertisement.Links = []string{advertisement.Links[0]}
		json.NewEncoder(response).Encode(advertisement)
		return
	} else {
		advertisement.Description = ""
		advertisement.Links = []string{advertisement.Links[0]}
		json.NewEncoder(response).Encode(advertisement)
		return
	}
}



func getPage(response http.ResponseWriter, request *http.Request) {
	PAGE_SIZE := 10  // Number of Ads on one page
	response.Header().Add("content-type", "application/json")
	var advertisements []Advertisement
	params := mux.Vars(request)
	pageNum, err := strconv.Atoi(params["pageNum"])
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	values := request.URL.Query()
	sortParameter, ok := values["sort"]  // Gets Param for Sort
	sortDirection, directionOk := values["direction"]  // Gets Direction of Sort

	if !directionOk {
		sortDirection = []string{"desc"}  // Default value for Sort Direction
	}

	if ok {
		collection := client.Database("advertisements").Collection("advertisements")
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		findOptions := options.Find()
		findOptions.SetLimit(int64(PAGE_SIZE))
		findOptions.SetSkip(int64(PAGE_SIZE * (pageNum - 1)))

		if sortParameter[0] == "price" { // Sort by Price
			if sortDirection[0] == "asc" {
				findOptions.SetSort(bson.D{{"price", 1}})
			} else {
				findOptions.SetSort(bson.D{{"price", -1}})
			}

		}

		if sortParameter[0] == "time" {  // Sort by TimeStamps
			if sortDirection[0] == "asc" {
				findOptions.SetSort(bson.D{{"created_at", 1}})
			} else {
				findOptions.SetSort(bson.D{{"created_at", -1}})
			}
		}

		cursor, err := collection.Find(ctx, bson.D{}, findOptions)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
			return
		}
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var advertisement Advertisement
			cursor.Decode(&advertisement)
			advertisement.CreatedAt = 0
			advertisements = append(advertisements, advertisement)

		}
		if err := cursor.Err(); err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
			return
		}
		json.NewEncoder(response).Encode(advertisements)

	}else {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": Sort parameter is not specified.\nChose added or price"`))
		return
	}


}


func handleRequests()  {

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homePage)
	router.HandleFunc("/add_advertisement", addAdvertisement).Methods("POST")
	router.HandleFunc("/advertisement/{id}", getAdvertisement).Methods("GET")
	router.HandleFunc("/ads/{pageNum}", getPage)
	log.Fatal(http.ListenAndServe(os.Getenv("PORT"), router))
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