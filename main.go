package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

//Here is our waitgroup
var wg sync.WaitGroup

//Parse our templates
func init() {
	//AmazonCredentialRead
	getCreds()
}

//Writes to the log; called from most anywhere in this program!
func logWriter(logMessage string) {
	//Logging info

	wd, _ := os.Getwd()
	logDir := filepath.Join(wd, "logging", "logging.txt")
	logFile, err := os.OpenFile(logDir, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)

	defer logFile.Close()

	if err != nil {
		fmt.Println("Failed opening log file")
	}

	log.SetOutput(logFile)

	log.Println(logMessage)
}

//Handles all requests coming in
func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)

	debugMessage := "\n\nWe are now handling requests"

	fmt.Println(debugMessage)
	logWriter(debugMessage)
	//Mongo No-SQL Stuff
	myRouter.HandleFunc("/testPing", testPing).Methods("POST")                               //Test a post to this server
	myRouter.HandleFunc("/addUser", addUser).Methods("POST")                                 //add a User
	myRouter.HandleFunc("/deleteUser", deleteUser).Methods("POST")                           //Delete a User
	myRouter.HandleFunc("/updateUser", updateUser).Methods("POST")                           //update a User
	myRouter.HandleFunc("/insertOneNewMessage", insertOneNewMessage).Methods("POST")         //insert a Message
	myRouter.HandleFunc("/deleteOneMessage", deleteOneMessage).Methods("POST")               //Delete a Message
	myRouter.HandleFunc("/updateOneMessage", updateOneMessage).Methods("POST")               //update a Message
	myRouter.HandleFunc("/uberUpdate", uberUpdate).Methods("POST")                           //update messages after a reply
	myRouter.HandleFunc("/updateMongoMessageBoard", updateMongoMessageBoard).Methods("POST") //update a Messageboard
	myRouter.HandleFunc("/giveAllUsernames", giveAllUsernames).Methods("GET")                //Return allUsernames
	myRouter.HandleFunc("/isMessageBoardCreated", isMessageBoardCreated).Methods("GET")      //Return messageboards
	//Field Validation Stuff
	myRouter.HandleFunc("/randomIDCreationAPI", randomIDCreationAPI).Methods("GET") //update a Message
	myRouter.HandleFunc("/userLogin", userLogin).Methods("POST")                    //update a Message
	//Serve our static files
	log.Fatal(http.ListenAndServe(":8080", myRouter))
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano()) //Randomly Seed

	//Mongo Connect
	mongoClient = connectDB()

	/* Test JSON Creation */
	type UpdatedMongoBoard struct {
		UpdatedMessageBoard MessageBoard `json:"UpdatedMessageBoard"`
	}
	theUpdatedMongoBoard := UpdatedMongoBoard{}
	theMBTest := MessageBoard{}
	theMessageTest := Message{
		MessageID:       334545,
		UserID:          445653,
		PosterName:      "JimUsername",
		Messages:        []Message{},
		IsChild:         false,
		HasChildren:     false,
		ParentMessageID: 0,
		UberParentID:    0,
		Order:           0,
		RepliesAmount:   0,
		TheMessage:      "Test message one",
		DateCreated:     "Uhhh",
		LastUpdated:     "eaadf",
	}

	theMBTest.AllMessages = append(theMBTest.AllMessages, theMessageTest)
	theMBTest.BoardName = "hotdog"
	theMBTest.MessageBoardID = 652774270816

	theUpdatedMongoBoard.UpdatedMessageBoard = theMBTest

	yee, _ := json.Marshal(theUpdatedMongoBoard)

	fmt.Printf("DEBUG: \n\n Here is yee: %v\n\n", string(yee))

	//Handle Requests
	handleRequests()
	defer mongoClient.Disconnect(theContext) //Disconnect in 10 seconds if you can't connect
}
