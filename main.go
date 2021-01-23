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
	logDir := filepath.Join(wd, "logging", "superDBAppLog.txt")
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

	//Mongo No-SQL Stuff
	//Serve our static files
	log.Fatal(http.ListenAndServe(":80", myRouter))
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano()) //Randomly Seed

	//Mongo Connect
	mongoClient = connectDB()
	/* Do below so our map dosen't go crazy... */

	//Handle Requests
	handleRequests()
	defer mongoClient.Disconnect(theContext) //Disconnect in 10 seconds if you can't connect
}
