package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

//Mongo DB Declarations
var mongoClient *mongo.Client

var theContext context.Context
var mongoURI string //Connection string loaded

type AUser struct { //Using this for Mongo
	UserName    string `json:"UserName"`
	Password    string `json:"Password"`
	UserID      int    `json:"UserID"`
	DateCreated string `json:"DateCreated"`
	DateUpdated string `json:"DateUpdated"`
	PostsMade   int    `json:"PostsMade"`
	RepliesMade int    `json:"RepliesMade"`
}

//This gets the client to connect to our DB
func connectDB() *mongo.Client {
	//Setup Mongo connection to Atlas Cluster
	theClient, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		fmt.Printf("Errored getting mongo client: %v\n", err)
		log.Fatal(err)
	}
	theContext, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err = theClient.Connect(theContext)
	if err != nil {
		fmt.Printf("Errored getting mongo client context: %v\n", err)
		log.Fatal(err)
	}
	//Double check to see if we've connected to the database
	err = theClient.Ping(theContext, readpref.Primary())
	if err != nil {
		fmt.Printf("Errored pinging MongoDB: %v\n", err)
		log.Fatal(err)
	}
	//List all available databases
	/*
		databases, err := theClient.ListDatabaseNames(theContext, bson.M{})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(databases)
	*/

	return theClient
}

//This gets the crednetials to launch our program
func getCreds() {
	file, err := os.Open("security/mongocreds.txt")

	if err != nil {
		fmt.Printf("Trouble opening Mongo login connections: %v\n", err.Error())
	}

	scanner := bufio.NewScanner(file)

	scanner.Split(bufio.ScanLines)
	var text []string

	for scanner.Scan() {
		text = append(text, scanner.Text())
	}

	file.Close()

	mongoURI = text[0]
}

//This adds a User to our database; called from anywhere
func addUser(w http.ResponseWriter, req *http.Request) {
	//Collect JSON from Postman or wherever
	//Get the byte slice from the request body ajax
	bs, err := ioutil.ReadAll(req.Body)
	if err != nil {
		theErr := "Error reading the request from addUser: " + err.Error() + "\n" + string(bs)
		logWriter(theErr)
		fmt.Println(theErr)
	}
	//Marshal it into our type
	var postedUser AUser
	json.Unmarshal(bs, &postedUser)

	user_collection := mongoClient.Database("microservice").Collection("users") //Here's our collection
	collectedUsers := []interface{}{postedUser}
	//Insert Our Data
	_, err2 := user_collection.InsertMany(context.TODO(), collectedUsers)
	//Declare data to return
	type ReturnMessage struct {
		TheErr     string `json:"TheErr"`
		ResultMsg  string `json:"ResultMsg"`
		SuccOrFail int    `json:"SuccOrFail"`
	}
	theReturnMessage := ReturnMessage{}
	if err2 != nil {
		theErr := "Error adding User in addUser in crudoperations API: " + err2.Error()
		logWriter(theErr)
		theReturnMessage.TheErr = theErr
		theReturnMessage.ResultMsg = theErr
		theReturnMessage.SuccOrFail = 1
		theJSONMessage, err := json.Marshal(theReturnMessage)
		//Send the response back
		if err != nil {
			errIs := "Error formatting JSON for return in addUser: " + err.Error()
			logWriter(errIs)
		}
		fmt.Fprint(w, string(theJSONMessage))
	} else {
		theErr := "User successfully added in addUser in crudoperations: " + string(bs)
		logWriter(theErr)
		theReturnMessage.TheErr = ""
		theReturnMessage.ResultMsg = theErr
		theReturnMessage.SuccOrFail = 0
		theJSONMessage, err := json.Marshal(theReturnMessage)
		//Send the response back
		if err != nil {
			errIs := "Error formatting JSON for return in addUser: " + err.Error()
			logWriter(errIs)
		}
		fmt.Fprint(w, string(theJSONMessage))
	}
}

//This is a test API we can ping on our Amazon server
func testPing(w http.ResponseWriter, r *http.Request) {
	//Initialize struct for taking messages
	type TestCrudPing struct {
		TheCrudPing string `json:"TheCrudPing"`
	}
	//Collect JSON from Postman or wherever
	//Get the byte slice from the request body ajax
	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		logWriter(err.Error())
	}
	//Marshal it into our type
	var postedMessage TestCrudPing
	json.Unmarshal(bs, &postedMessage)

	messageLog := "We've had a ping come in from somewhere: " + postedMessage.TheCrudPing
	logWriter(messageLog)
	fmt.Printf("%v\n", messageLog)

	//Declare data to return
	type ReturnMessage struct {
		TheErr     string `json:"TheErr"`
		ResultMsg  string `json:"ResultMsg"`
		SuccOrFail int    `json:"SuccOrFail"`
	}
	theReturnMessage := ReturnMessage{
		TheErr:     "Here's an error",
		ResultMsg:  "Yo, here's a result",
		SuccOrFail: 0,
	}
	//Send the response back
	theJSONMessage, err := json.Marshal(theReturnMessage)
	//Send the response back
	if err != nil {
		errIs := "Error formatting JSON for return in addUser: " + err.Error()
		logWriter(errIs)
	}
	fmt.Fprint(w, string(theJSONMessage))
}
