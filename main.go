package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
)

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
	myRouter.HandleFunc("/testPing", testPing).Methods("POST")     //Test a post to this server
	myRouter.HandleFunc("/addUser", addUser).Methods("POST")       //add a User
	myRouter.HandleFunc("/deleteUser", deleteUser).Methods("POST") //Delete a User
	myRouter.HandleFunc("/updateUser", updateUser).Methods("POST") //update a User
	//Serve our static files
	log.Fatal(http.ListenAndServe(":8080", myRouter))
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano()) //Randomly Seed

	//Mongo Connect
	mongoClient = connectDB()

	//Handle Requests
	handleRequests()
	defer mongoClient.Disconnect(theContext) //Disconnect in 10 seconds if you can't connect
}
