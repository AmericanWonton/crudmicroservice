package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"gopkg.in/mgo.v2/bson"
)

//Mongo DB Declarations
var mongoClient *mongo.Client

var theContext context.Context
var mongoURI string //Connection string loaded

type AUser struct { //Using this for Mongo
	UserName    string `json:"UserName"`
	Password    string `json:"Password"`
	UserID      int    `json:"UserID"`
	Email       string `json:"Email"`
	PhoneACode  int    `json:"PhoneACode"`
	PhoneNumber int    `json:"PhoneNumber"`
	PostsMade   int    `json:"PostsMade"`
	RepliesMade int    `json:"RepliesMade"`
	DateCreated string `json:"DateCreated"`
	DateUpdated string `json:"DateUpdated"`
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

//This deletes a User to our database; called from anywhere
func deleteUser(w http.ResponseWriter, req *http.Request) {
	//Get the byte slice from the request body ajax
	bs, err := ioutil.ReadAll(req.Body)
	if err != nil {
		theErr := "Error reading the request from deleteUser: " + err.Error() + "\n" + string(bs)
		logWriter(theErr)
		fmt.Println(theErr)
	}
	//Declare JSON we're looking for
	type UserDelete struct {
		UserID int `json:"UserID"`
	}
	//Marshal it into our type
	var postedUserID UserDelete
	json.Unmarshal(bs, &postedUserID)

	fmt.Printf("DEBUG: Here is our UserID: %v\n", postedUserID.UserID)

	//Search for User and delete
	userCollection := mongoClient.Database("microservice").Collection("users") //Here's our collection
	deletes := []bson.M{
		{"userid": postedUserID.UserID},
	} //Here's our filter to look for
	deletes = append(deletes, bson.M{"userid": bson.M{
		"$eq": postedUserID.UserID,
	}}, bson.M{"userid": bson.M{
		"$eq": postedUserID.UserID,
	}},
	)

	// create the slice of write models
	var writes []mongo.WriteModel

	for _, del := range deletes {
		model := mongo.NewDeleteManyModel().SetFilter(del)
		writes = append(writes, model)
	}

	//Declare data to return
	type ReturnMessage struct {
		TheErr     string `json:"TheErr"`
		ResultMsg  string `json:"ResultMsg"`
		SuccOrFail int    `json:"SuccOrFail"`
	}
	theReturnMessage := ReturnMessage{}

	// run bulk write
	bulkWrite, err := userCollection.BulkWrite(theContext, writes)
	if err != nil {
		theErr := "Error writing delete User in deleteUser in crudoperations: " + err.Error()
		logWriter(theErr)
		theReturnMessage.TheErr = theErr
		theReturnMessage.ResultMsg = theErr
		theReturnMessage.SuccOrFail = 1
		theJSONMessage, err := json.Marshal(theReturnMessage)
		//Send the response back
		if err != nil {
			errIs := "Error formatting JSON for return in deleteUser: " + err.Error()
			logWriter(errIs)
		}
		fmt.Fprint(w, string(theJSONMessage))
	} else {
		theErr := "User successfully deleted in deleteUser in crudoperations: " + string(bs)
		fmt.Printf("DEBUG: %v . Here is the amount deleted:%v\n", theErr, bulkWrite.DeletedCount)
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

//This updates a User to our database; called from anywhere
func updateUser(w http.ResponseWriter, req *http.Request) {
	//Unwrap from JSON
	bs, err := ioutil.ReadAll(req.Body)
	if err != nil {
		theErr := "Error reading the request from updateUser: " + err.Error() + "\n" + string(bs)
		logWriter(theErr)
		fmt.Println(theErr)
	}

	//Marshal it into our type
	var theUserUpdate AUser
	json.Unmarshal(bs, &theUserUpdate)

	fmt.Printf("DEBUG: The UserID sent to us is: %v\n\n", theUserUpdate.UserID)

	//Update User
	theTimeNow := time.Now()
	userCollection := mongoClient.Database("microservice").Collection("users") //Here's our collection
	theFilter := bson.M{
		"userid": bson.M{
			"$eq": theUserUpdate.UserID, // check if bool field has value of 'false'
		},
	}
	updatedDocument := bson.M{
		"$set": bson.M{
			"username":    theUserUpdate.UserName,
			"password":    theUserUpdate.Password,
			"userid":      theUserUpdate.UserID,
			"email":       theUserUpdate.Email,
			"phoneacode":  theUserUpdate.PhoneACode,
			"phonenumber": theUserUpdate.PhoneNumber,
			"postsmade":   theUserUpdate.PostsMade,
			"repliesmade": theUserUpdate.RepliesMade,
			"datecreated": theUserUpdate.DateCreated,
			"dateupdated": theTimeNow.Format("2006-01-02 15:04:05"),
		},
	}
	updatedInfo, err := userCollection.UpdateOne(theContext, theFilter, updatedDocument)

	//Declare data to return
	type ReturnMessage struct {
		TheErr     string `json:"TheErr"`
		ResultMsg  string `json:"ResultMsg"`
		SuccOrFail int    `json:"SuccOrFail"`
	}
	theReturnMessage := ReturnMessage{}
	if err != nil {
		theErr := "Error writing update User in updateUser in crudoperations: " + err.Error()
		logWriter(theErr)
		theReturnMessage.TheErr = theErr
		theReturnMessage.ResultMsg = theErr
		theReturnMessage.SuccOrFail = 1
		theJSONMessage, err := json.Marshal(theReturnMessage)
		//Send the response back
		if err != nil {
			errIs := "Error formatting JSON for return in updateUser: " + err.Error()
			logWriter(errIs)
		}
		fmt.Fprint(w, string(theJSONMessage))
	} else {
		theErr := "User successfully updated in updateUser in crudoperations: " + string(bs)
		fmt.Printf("DEBUG: %v. Here is the update results: %v\n", theErr, updatedInfo.ModifiedCount)
		logWriter(theErr)
		theReturnMessage.TheErr = ""
		theReturnMessage.ResultMsg = theErr
		theReturnMessage.SuccOrFail = 0
		theJSONMessage, err := json.Marshal(theReturnMessage)
		//Send the response back
		if err != nil {
			errIs := "Error formatting JSON for return in updateUser: " + err.Error()
			logWriter(errIs)
		}
		fmt.Fprint(w, string(theJSONMessage))
	}
}

/* This function returns a map of ALL Usernames entered in our database
when called, (should be on the index page ) */
func giveAllUsernames(w http.ResponseWriter, req *http.Request) {
	//Declare empty map to fill and return
	usernameMap := make(map[string]bool) //Clear Map for future use on page load
	//Declare data to return
	type ReturnMessage struct {
		TheErr          []string        `json:"TheErr"`
		ResultMsg       []string        `json:"ResultMsg"`
		SuccOrFail      int             `json:"SuccOrFail"`
		ReturnedUserMap map[string]bool `json:"ReturnedUserMap"`
	}
	theReturnMessage := ReturnMessage{}
	theReturnMessage.SuccOrFail = 0 //Initially set to success

	userCollection := mongoClient.Database("microservice").Collection("users") //Here's our collection

	//Query Mongo for all Users
	theFilter := bson.M{}
	findOptions := options.Find()
	currUser, err := userCollection.Find(theContext, theFilter, findOptions)
	if err != nil {
		if strings.Contains(err.Error(), "no documents in result") {
			theErr := "No documents were returned for hotdogs in MongoDB: " + err.Error()
			fmt.Printf("DEBUG: %v\n", theErr)
			theReturnMessage.ResultMsg = append(theReturnMessage.ResultMsg, theErr)
			theReturnMessage.TheErr = append(theReturnMessage.TheErr, theErr)
			theReturnMessage.SuccOrFail = 1
			logWriter(theErr)
		} else {
			theErr := "There was an error returning results for this Users, :" + err.Error()
			fmt.Printf("DEBUG: %v\n", theErr)
			theReturnMessage.ResultMsg = append(theReturnMessage.ResultMsg, theErr)
			theReturnMessage.TheErr = append(theReturnMessage.TheErr, theErr)
			theReturnMessage.SuccOrFail = 1
			logWriter(theErr)
		}
	}
	//Loop over query results and fill hotdogs array
	for currUser.Next(theContext) {
		// create a value into which the single document can be decoded
		var aUser AUser
		err := currUser.Decode(&aUser)
		if err != nil {
			theErr := "Error decoding Users in MongoDB in giveAllUsernames: " + err.Error()
			theReturnMessage.ResultMsg = append(theReturnMessage.ResultMsg, theErr)
			theReturnMessage.TheErr = append(theReturnMessage.TheErr, theErr)
			theReturnMessage.SuccOrFail = 0
			logWriter(theErr)
		}
		//Fill Username map with the found Username
		usernameMap[aUser.UserName] = true
	}
	// Close the cursor once finished
	currUser.Close(theContext)

	//Check to see if anyusernames were returned or we have errors
	if theReturnMessage.SuccOrFail >= 1 {
		theErr := "There are a number of errors for returning these Usernames..."
		theReturnMessage.ResultMsg = append(theReturnMessage.ResultMsg, theErr)
		theReturnMessage.TheErr = append(theReturnMessage.TheErr, theErr)
	} else if len(usernameMap) <= 0 {
		theErr := "No usernames returned...this could be the site's first deployment with no users!"
		theReturnMessage.ResultMsg = append(theReturnMessage.ResultMsg, theErr)
		theReturnMessage.TheErr = append(theReturnMessage.TheErr, theErr)
		theReturnMessage.SuccOrFail = 2
	} else {
		theErr := "No issues returning Usernames"
		theReturnMessage.ResultMsg = append(theReturnMessage.ResultMsg, theErr)
		theReturnMessage.TheErr = append(theReturnMessage.TheErr, theErr)
		theReturnMessage.SuccOrFail = 0
	}
	theReturnMessage.ReturnedUserMap = usernameMap
	//Format the JSON map for returning our results
	theJSONMessage, err := json.Marshal(theReturnMessage)
	//Send the response back
	if err != nil {
		errIs := "Error formatting JSON for return in giveAllUsernames: " + err.Error()
		logWriter(errIs)
	}
	fmt.Fprint(w, string(theJSONMessage))
}

/* This function searches with a Username and password to return a yes or no response
if the User is found; is so, we return the User, with a successful response.
If not, we return a failed response and an empty User profile */
func userLogin(w http.ResponseWriter, req *http.Request) {
	//Declare type to be returned later through JSON Response
	type ReturnMessage struct {
		TheErr     string `json:"TheErr"`
		ResultMsg  string `json:"ResultMsg"`
		SuccOrFail int    `json:"SuccOrFail"`
		TheUser    AUser  `json:"TheUser"`
	}
	theResponseMessage := ReturnMessage{}
	//Collect JSON from Postman or wherever
	//Get the byte slice from the request body ajax
	bs, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Println(err)
		logWriter(err.Error())
	}

	type LoginData struct {
		Username string `json:"Username"`
		Password string `json:"Password"`
	}

	//Marshal the user data into our type
	var dataForLogin LoginData
	json.Unmarshal(bs, &dataForLogin)

	var theUserReturned AUser //Initialize User to be returned after Mongo query

	//Query for the User, given the userID for the User
	user_collection := mongoClient.Database("microservice").Collection("users") //Here's our collection
	theFilter := bson.M{
		"username": bson.M{
			"$eq": dataForLogin.Username, // check if bool field has value of 'false'
		},
		"password": bson.M{
			"$eq": dataForLogin.Password,
		},
	}
	findOptions := options.FindOne()
	findUser := user_collection.FindOne(theContext, theFilter, findOptions)
	if findUser.Err() != nil {
		if strings.Contains(err.Error(), "no documents in result") {
			returnedErr := "For " + dataForLogin.Username + ", no User was returned: " + err.Error()
			fmt.Println(returnedErr)
			logWriter(returnedErr)
			theResponseMessage.SuccOrFail = 1
			theResponseMessage.ResultMsg = returnedErr
			theResponseMessage.TheErr = returnedErr
			theResponseMessage.TheUser = AUser{}
		} else {
			returnedErr := "For " + dataForLogin.Username + ", there was a Mongo Error: " + err.Error()
			fmt.Println(returnedErr)
			logWriter(returnedErr)
			theResponseMessage.SuccOrFail = 1
			theResponseMessage.ResultMsg = returnedErr
			theResponseMessage.TheErr = returnedErr
			theResponseMessage.TheUser = AUser{}
		}
	} else {
		//Found User, decode to return
		err := findUser.Decode(&theUserReturned)
		if err != nil {
			returnedErr := "For " + dataForLogin.Username +
				", there was an error decoding document from Mongo: " + err.Error()
			fmt.Println(returnedErr)
			logWriter(returnedErr)
			theResponseMessage.SuccOrFail = 2
			theResponseMessage.ResultMsg = returnedErr
			theResponseMessage.TheErr = returnedErr
			theResponseMessage.TheUser = AUser{}
		} else {
			returnedErr := "For " + dataForLogin.Username +
				", User should be successfully decoded."
			fmt.Println(returnedErr)
			logWriter(returnedErr)
			theResponseMessage.SuccOrFail = 0
			theResponseMessage.ResultMsg = returnedErr
			theResponseMessage.TheErr = ""
			theResponseMessage.TheUser = theUserReturned
		}
	}
	//Errors/Success are recorded, User given, send JSON back
	theJSONMessage, err := json.Marshal(theResponseMessage)
	//Send the response back
	if err != nil {
		errIs := "Error formatting JSON for return in userLogin: " + err.Error()
		logWriter(errIs)
	}
	fmt.Fprint(w, string(theJSONMessage))
}

//This should give a random id value to both food groups
func randomIDCreationAPI(w http.ResponseWriter, req *http.Request) {
	type ReturnMessage struct {
		TheErr     []string `json:"TheErr"`
		ResultMsg  []string `json:"ResultMsg"`
		SuccOrFail int      `json:"SuccOrFail"`
		RandomID   int      `json:"RandomID"`
	}
	theReturnMessage := ReturnMessage{}
	finalID := 0        //The final, unique ID to return to the food/user
	randInt := 0        //The random integer added onto ID
	randIntString := "" //The integer built through a string...
	min, max := 0, 9    //The min and Max value for our randInt
	foundID := false
	for foundID == false {
		randInt = 0
		randIntString = ""
		//Create the random number, convert it to string
		for i := 0; i < 12; i++ {
			randInt = rand.Intn(max-min) + min
			randIntString = randIntString + strconv.Itoa(randInt)
		}
		//Once we have a string of numbers, we can convert it back to an integer
		theID, err := strconv.Atoi(randIntString)
		if err != nil {
			fmt.Printf("We got an error converting a string back to a number, %v\n", err)
			fmt.Printf("Here is randInt: %v\n and randIntString: %v\n", randInt, randIntString)
			fmt.Println(err)
			log.Fatal(err)
		}
		//Search all our collections to see if this UserID is unique
		canExit := []bool{true, true, true}
		//User collection
		userCollection := mongoClient.Database("microservice").Collection("users") //Here's our collection
		var testAUser AUser
		theErr := userCollection.FindOne(theContext, bson.M{"userid": theID}).Decode(&testAUser)
		if theErr != nil {
			if strings.Contains(theErr.Error(), "no documents in result") {
				fmt.Printf("It's all good, this document wasn't found for User and our ID is clean.\n")
				canExit[0] = true
			} else {
				fmt.Printf("DEBUG: We have another error for finding a unique UserID: \n%v\n", theErr)
				canExit[0] = false
				log.Fatal(theErr)
			}
		}
		//Check Messageboard collection
		messageBoardCollection := mongoClient.Database("microservice").Collection("messageboard") //Here's our collection
		var testMessageBoard MessageBoard
		//Give 0 values to determine if these IDs are found
		theFilter := bson.M{
			"$or": []interface{}{
				bson.M{"messageboardid": theID},
			},
		}
		theErr = messageBoardCollection.FindOne(theContext, theFilter).Decode(&testMessageBoard)
		if theErr != nil {
			if strings.Contains(theErr.Error(), "no documents in result") {
				fmt.Printf("It's all good, this document wasn't found for User/Hotdog and our ID is clean.\n")
				canExit[1] = true
			} else {
				fmt.Printf("DEBUG: We have another error for finding a unique UserID: \n%v\n", theErr)
				canExit[1] = false
			}
		}
		//Check Message collection
		messagesCollection := mongoClient.Database("microservice").Collection("messages") //Here's our collection
		var testMessage Message
		//Give 0 values to determine if these IDs are found
		theFilter2 := bson.M{
			"$or": []interface{}{
				bson.M{"messageid": theID},
				bson.M{"userid": theID},
				bson.M{"parentmessageid": theID},
				bson.M{"uberparentid": theID},
			},
		}
		theErr = messagesCollection.FindOne(theContext, theFilter2).Decode(&testMessage)
		if theErr != nil {
			if strings.Contains(theErr.Error(), "no documents in result") {
				canExit[2] = true
				fmt.Printf("It's all good, this document wasn't found for User/hamburger and our ID is clean.\n")
			} else {
				fmt.Printf("DEBUG: We have another error for finding a unique UserID: \n%v\n", theErr)
				canExit[2] = false
			}
		}
		//Final check to see if we can exit this loop
		if canExit[0] == true && canExit[1] == true && canExit[2] == true {
			finalID = theID
			foundID = true
			theReturnMessage.RandomID = finalID
			theReturnMessage.SuccOrFail = 0
		} else {
			foundID = false
		}
	}

	/* Return the marshaled response */
	//Send the response back
	theJSONMessage, err := json.Marshal(theReturnMessage)
	//Send the response back
	if err != nil {
		errIs := "Error formatting JSON for return in randomIDCreationAPI: " + err.Error()
		logWriter(errIs)
	}
	fmt.Fprint(w, string(theJSONMessage))
}

//This creates a random API with no writer call,(called from this Microservice)
func randomIDCreationAPISimple() int {
	finalID := 0        //The final, unique ID to return to the food/user
	randInt := 0        //The random integer added onto ID
	randIntString := "" //The integer built through a string...
	min, max := 0, 9    //The min and Max value for our randInt
	foundID := false
	for foundID == false {
		randInt = 0
		randIntString = ""
		//Create the random number, convert it to string
		for i := 0; i < 12; i++ {
			randInt = rand.Intn(max-min) + min
			randIntString = randIntString + strconv.Itoa(randInt)
		}
		//Once we have a string of numbers, we can convert it back to an integer
		theID, err := strconv.Atoi(randIntString)
		if err != nil {
			fmt.Printf("We got an error converting a string back to a number, %v\n", err)
			fmt.Printf("Here is randInt: %v\n and randIntString: %v\n", randInt, randIntString)
			fmt.Println(err)
			log.Fatal(err)
		}
		//Search all our collections to see if this UserID is unique
		canExit := []bool{true, true, true}
		//User collection
		userCollection := mongoClient.Database("microservice").Collection("users") //Here's our collection
		var testAUser AUser
		theErr := userCollection.FindOne(theContext, bson.M{"userid": theID}).Decode(&testAUser)
		if theErr != nil {
			if strings.Contains(theErr.Error(), "no documents in result") {
				fmt.Printf("It's all good, this document wasn't found for User and our ID is clean.\n")
				canExit[0] = true
			} else {
				fmt.Printf("DEBUG: We have another error for finding a unique UserID: \n%v\n", theErr)
				canExit[0] = false
				log.Fatal(theErr)
			}
		}
		//Check Messageboard collection
		messageBoardCollection := mongoClient.Database("microservice").Collection("messageboard") //Here's our collection
		var testMessageBoard MessageBoard
		//Give 0 values to determine if these IDs are found
		theFilter := bson.M{
			"$or": []interface{}{
				bson.M{"messageboardid": theID},
			},
		}
		theErr = messageBoardCollection.FindOne(theContext, theFilter).Decode(&testMessageBoard)
		if theErr != nil {
			if strings.Contains(theErr.Error(), "no documents in result") {
				fmt.Printf("It's all good, this document wasn't found for User/Hotdog and our ID is clean.\n")
				canExit[1] = true
			} else {
				fmt.Printf("DEBUG: We have another error for finding a unique UserID: \n%v\n", theErr)
				canExit[1] = false
			}
		}
		//Check Message collection
		messagesCollection := mongoClient.Database("microservice").Collection("messages") //Here's our collection
		var testMessage Message
		//Give 0 values to determine if these IDs are found
		theFilter2 := bson.M{
			"$or": []interface{}{
				bson.M{"messageid": theID},
				bson.M{"userid": theID},
				bson.M{"parentmessageid": theID},
				bson.M{"uberparentid": theID},
			},
		}
		theErr = messagesCollection.FindOne(theContext, theFilter2).Decode(&testMessage)
		if theErr != nil {
			if strings.Contains(theErr.Error(), "no documents in result") {
				canExit[2] = true
				fmt.Printf("It's all good, this document wasn't found for User/hamburger and our ID is clean.\n")
			} else {
				fmt.Printf("DEBUG: We have another error for finding a unique UserID: \n%v\n", theErr)
				canExit[2] = false
			}
		}
		//Final check to see if we can exit this loop
		if canExit[0] == true && canExit[1] == true && canExit[2] == true {
			finalID = theID
			foundID = true
		} else {
			foundID = false
		}
	}

	return finalID
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
