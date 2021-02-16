package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

//Data sent to index page
type MessageViewData struct {
	TestString  string    `json:"TestString"`
	TheMessages []Message `json:"TheMessages"`
	WhatPage    int       `json:"WhatPage"`
}

//Message displayed on the board
type Message struct {
	MessageID       int       `json:"MessageID"`       //ID of this Message
	UserID          int       `json:"UserID"`          //ID of the owner of this message
	PosterName      string    `json:"PosterName"`      //Username of the poster of this message
	Messages        []Message `json:"Messages"`        //Array of Messages under this one
	IsChild         bool      `json:"IsChild"`         //Is this message childed to another message
	HasChildren     bool      `json:"HasChildren"`     //Whether this message has children to list
	ParentMessageID int       `json:"ParentMessageID"` //The ID of this parent
	UberParentID    int       `json:"UberParentID"`    //The final parent of this parent, IF EQUAL PARENT
	Order           int       `json:"Order"`           //Order the commnet is in with it's reply tree
	RepliesAmount   int       `json:"RepliesAmount"`   //Amount of replies this message has
	TheMessage      string    `json:"TheMessage"`      //The Message in the post
	WhatBoard       string    `json:"WhatBoard"`       //The board this message is apart of
	DateCreated     string    `json:"DateCreated"`     //When the message was created
	LastUpdated     string    `json:"LastUpdated"`     //When the message was last updated
}

//All the Messages on the board
type MessageBoard struct {
	MessageBoardID         int             `json:"MessageBoardID"`
	BoardName              string          `json:"BoardName"`              //The Name of the board
	AllMessages            []Message       `json:"AllMessages"`            //All the IDs listed
	AllMessagesMap         map[int]Message `json:"AllMessagesMap"`         //A map of ALL messages
	AllOriginalMessages    []Message       `json:"AllOriginalMessages"`    //All the messages that AREN'T replies
	AllOriginalMessagesMap map[int]Message `json:"AllOriginalMessagesMap"` //Map of original Messages
	LastUpdated            string          `json:"LastUpdated"`            //Last time this messageboard was updated
	DateCreated            string          `json:"DateCreated"`            //Date this board was created
}

var loadedMessagesMap map[int]Message

/* This is the current amount of results our User is looking at
it changes as the User clicks forwards or backwards for more results */
var currentPageNumber int = 1

//Inserts one message into our 'messages' collection
func insertOneNewMessage(w http.ResponseWriter, r *http.Request) {
	//Unwrap from JSON
	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		theErr := "Error reading the request from insertOneNewMessage: " + err.Error() + "\n" + string(bs)
		logWriter(theErr)
		fmt.Println(theErr)
	}

	//Marshal it into our type
	var messageIn Message
	json.Unmarshal(bs, &messageIn)

	//Send this to the 'message' collection for safekeeping
	messageCollection := mongoClient.Database("microservice").Collection("messages") //Here's our collection
	collectedStuff := []interface{}{messageIn}
	//Insert Our Data
	_, err = messageCollection.InsertMany(context.TODO(), collectedStuff)
	//Declare data to return
	type ReturnMessage struct {
		TheErr     string `json:"TheErr"`
		ResultMsg  string `json:"ResultMsg"`
		SuccOrFail int    `json:"SuccOrFail"`
	}
	theReturnMessage := ReturnMessage{}
	if err != nil {
		theErr := "Error writing insert message in insertOneNewMessage in crudoperations: " + err.Error()
		logWriter(theErr)
		theReturnMessage.TheErr = theErr
		theReturnMessage.ResultMsg = theErr
		theReturnMessage.SuccOrFail = 1
		theJSONMessage, err := json.Marshal(theReturnMessage)
		//Send the response back
		if err != nil {
			errIs := "Error formatting JSON for return in insertOneNewMessage: " + err.Error()
			logWriter(errIs)
		}
		fmt.Fprint(w, string(theJSONMessage))
	} else {
		theErr := "User successfully inserted message in insertOneNewMessage in crudoperations: " + string(bs)
		logWriter(theErr)
		theReturnMessage.TheErr = ""
		theReturnMessage.ResultMsg = theErr
		theReturnMessage.SuccOrFail = 0
		theJSONMessage, err := json.Marshal(theReturnMessage)
		//Send the response back
		if err != nil {
			errIs := "Error formatting JSON for return in insertOneNewMessage: " + err.Error()
			logWriter(errIs)
		}
		fmt.Fprint(w, string(theJSONMessage))
	}
}

//Deletes one message in our 'messages' collection
func deleteOneMessage(w http.ResponseWriter, r *http.Request) {
	//Get the byte slice from the request body ajax
	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		theErr := "Error reading the request from deleteUser: " + err.Error() + "\n" + string(bs)
		logWriter(theErr)
		fmt.Println(theErr)
	}
	//Declare JSON we're looking for
	type MessageDelete struct {
		MessageID int `json:"MessageID"`
	}
	//Marshal it into our type
	var postedMessageID MessageDelete
	json.Unmarshal(bs, &postedMessageID)

	//Search for Message and delete
	messageCollection := mongoClient.Database("microservice").Collection("messages") //Here's our collection
	deletes := []bson.M{
		{"messageid": postedMessageID.MessageID},
	} //Here's our filter to look for
	deletes = append(deletes, bson.M{"messageid": bson.M{
		"$eq": postedMessageID.MessageID,
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
	_, err = messageCollection.BulkWrite(theContext, writes)
	if err != nil {
		theErr := "Error writing delete Message in deleteOneMessage in crudoperations: " + err.Error()
		logWriter(theErr)
		theReturnMessage.TheErr = theErr
		theReturnMessage.ResultMsg = theErr
		theReturnMessage.SuccOrFail = 1
		theJSONMessage, err := json.Marshal(theReturnMessage)
		//Send the response back
		if err != nil {
			errIs := "Error formatting JSON for return in deleteOneMessage: " + err.Error()
			logWriter(errIs)
		}
		fmt.Fprint(w, string(theJSONMessage))
	} else {
		theErr := "Message successfully deleted in deleteOneMessage in crudoperations: " + string(bs)
		logWriter(theErr)
		theReturnMessage.TheErr = ""
		theReturnMessage.ResultMsg = theErr
		theReturnMessage.SuccOrFail = 0
		theJSONMessage, err := json.Marshal(theReturnMessage)
		//Send the response back
		if err != nil {
			errIs := "Error formatting JSON for return in deleteOneMessage: " + err.Error()
			logWriter(errIs)
		}
		fmt.Fprint(w, string(theJSONMessage))
	}
}

//Update one message in our 'messages' collection
func updateOneMessage(w http.ResponseWriter, r *http.Request) {
	//Unwrap from JSON
	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		theErr := "Error reading the request from updateOneMessage: " + err.Error() + "\n" + string(bs)
		logWriter(theErr)
		fmt.Println(theErr)
	}

	//Marshal it into our type
	var theMessageUpdate Message
	json.Unmarshal(bs, &theMessageUpdate)

	//Update User
	theTimeNow := time.Now()
	messageCollection := mongoClient.Database("microservice").Collection("messages") //Here's our collection
	theFilter := bson.M{
		"messageid": bson.M{
			"$eq": theMessageUpdate.MessageID,
		},
	}
	updatedDocument := bson.M{
		"$set": bson.M{
			"messageid":       theMessageUpdate.MessageID,
			"userid":          theMessageUpdate.UserID,
			"postername":      theMessageUpdate.PosterName,
			"messages":        theMessageUpdate.Messages,
			"ischild":         theMessageUpdate.IsChild,
			"haschildren":     theMessageUpdate.HasChildren,
			"parentmessageid": theMessageUpdate.ParentMessageID,
			"uberparentid":    theMessageUpdate.UberParentID,
			"order":           theMessageUpdate.Order,
			"repliesamount":   theMessageUpdate.RepliesAmount,
			"themessage":      theMessageUpdate.TheMessage,
			"whatboard":       theMessageUpdate.WhatBoard,
			"datecreated":     theMessageUpdate.DateCreated,
			"lastupdated":     theTimeNow.Format("2006-01-02 15:04:05"),
		},
	}
	_, err = messageCollection.UpdateOne(theContext, theFilter, updatedDocument)

	//Declare data to return
	type ReturnMessage struct {
		TheErr     string `json:"TheErr"`
		ResultMsg  string `json:"ResultMsg"`
		SuccOrFail int    `json:"SuccOrFail"`
	}
	theReturnMessage := ReturnMessage{}
	if err != nil {
		theErr := "Error writing update Message in updateOneMessage in crudoperations: " + err.Error()
		logWriter(theErr)
		theReturnMessage.TheErr = theErr
		theReturnMessage.ResultMsg = theErr
		theReturnMessage.SuccOrFail = 1
		theJSONMessage, err := json.Marshal(theReturnMessage)
		//Send the response back
		if err != nil {
			errIs := "Error formatting JSON for return in updateOneMessage: " + err.Error()
			logWriter(errIs)
		}
		fmt.Fprint(w, string(theJSONMessage))
	} else {
		theErr := "Message successfully updated in updateOneMessage in crudoperations: " + string(bs)
		logWriter(theErr)
		theReturnMessage.TheErr = ""
		theReturnMessage.ResultMsg = theErr
		theReturnMessage.SuccOrFail = 0
		theJSONMessage, err := json.Marshal(theReturnMessage)
		//Send the response back
		if err != nil {
			errIs := "Error formatting JSON for return in updateOneMessage: " + err.Error()
			logWriter(errIs)
		}
		fmt.Fprint(w, string(theJSONMessage))
	}
}

//Returns true if our test message board is already created
func isMessageBoardCreated(w http.ResponseWriter, r *http.Request) {
	//Return message
	type ReturnMessage struct {
		TheErr      []string     `json:"TheErr"`
		ResultMsg   []string     `json:"ResultMsg"`
		SuccOrFail  int          `json:"SuccOrFail"`
		GivenHDogMB MessageBoard `json:"GivenHDogMB"`
		GivenHamMB  MessageBoard `json:"GivenHamMB"`
	}
	theReturnMessage := ReturnMessage{}
	theReturnMessage.SuccOrFail = 0 //Declare this first to check against failure when returning response

	/* Run a mongo query to see if the messageboard is created;
	if it isn't, create it and return the new created board.
	If it is, just return the board */
	theTimeNow := time.Now() //needed for later
	//Find the hotdog messageboard
	messageCollectionHD := mongoClient.Database("microservice").Collection("messageboard") //Here's our collection
	//Query Mongo for all Messages
	theFilterHD := bson.M{
		"boardname": bson.M{
			"$eq": "hotdog", // check if bool field has value of 'false'
		},
	}
	findOptionsHD := options.Find()
	messageBoardHD, err := messageCollectionHD.Find(theContext, theFilterHD, findOptionsHD)
	if err != nil {
		if strings.Contains(err.Error(), "no documents in result") {
			themessage := "No document returned; creating hotdog messageboard"
			logWriter(themessage)

			theReturnMessage.GivenHDogMB = MessageBoard{
				MessageBoardID: randomIDCreationAPISimple(),
				BoardName:      "hotdog",
				LastUpdated:    theTimeNow.Format("2006-01-02 15:04:05"),
				DateCreated:    theTimeNow.Format("2006-01-02 15:04:05"),
			}
			theReturnMessage.ResultMsg = append(theReturnMessage.ResultMsg, themessage)
		} else {
			themessage := "Error getting the hotdog database: " + err.Error()
			logWriter(themessage)
			theReturnMessage.ResultMsg = append(theReturnMessage.ResultMsg, themessage)
			theReturnMessage.TheErr = append(theReturnMessage.TheErr, themessage)
			theReturnMessage.SuccOrFail = 1
		}
	} else {
		// create a value into which the single document can be decoded
		var aMessageBoard MessageBoard
		//Loop over query results and fill hotdogs array
		for messageBoardHD.Next(theContext) {
			err := messageBoardHD.Decode(&aMessageBoard)
			if err != nil {
				errmsg := "Error decoding messageboard in MongoDB for all Users: " + err.Error()
				fmt.Println(errmsg)
				logWriter(errmsg)
			}
			//Assign our message board to the 'theMessageBoard' to work with
			theReturnMessage.GivenHDogMB = aMessageBoard
		}
		// Close the cursor once finished
		messageBoardHD.Close(theContext)
	}

	//Find the hamburger messageboard
	messageCollectionHam := mongoClient.Database("microservice").Collection("messageboard") //Here's our collection
	//Query Mongo for all Messages
	theFilterHam := bson.M{
		"boardname": bson.M{
			"$eq": "hamburger", // check if bool field has value of 'false'
		},
	}
	findOptionsHam := options.Find()
	messageBoardHam, err := messageCollectionHam.Find(theContext, theFilterHam, findOptionsHam)
	if err != nil {
		if strings.Contains(err.Error(), "no documents in result") {
			themessage := "No document returned; creating hamburger messageboard"
			logWriter(themessage)

			theReturnMessage.GivenHamMB = MessageBoard{
				MessageBoardID: randomIDCreationAPISimple(),
				BoardName:      "hamburger",
				LastUpdated:    theTimeNow.Format("2006-01-02 15:04:05"),
				DateCreated:    theTimeNow.Format("2006-01-02 15:04:05"),
			}
			theReturnMessage.ResultMsg = append(theReturnMessage.ResultMsg, themessage)
		} else {
			themessage := "Error getting the hamburger database: " + err.Error()
			logWriter(themessage)
			theReturnMessage.ResultMsg = append(theReturnMessage.ResultMsg, themessage)
			theReturnMessage.TheErr = append(theReturnMessage.TheErr, themessage)
			theReturnMessage.SuccOrFail = 1
		}
	} else {
		// create a value into which the single document can be decoded
		var aMessageBoard MessageBoard
		//Loop over query results and fill hotdogs array
		for messageBoardHam.Next(theContext) {
			err := messageBoardHam.Decode(&aMessageBoard)
			if err != nil {
				errmsg := "Error decoding messageboard in MongoDB for all Users: " + err.Error()
				fmt.Println(errmsg)
				logWriter(errmsg)
			}
			//Assign our message board to the 'theMessageBoard' to work with
			theReturnMessage.GivenHamMB = aMessageBoard
		}
		// Close the cursor once finished
		messageBoardHam.Close(theContext)
	}

	//Determine if we have an error yet and what to return
	if theReturnMessage.SuccOrFail != 0 {
		//Log a failure and return the failure
		aMessage := "Failure returning hotdog or hamburger messageboards"
		theReturnMessage.ResultMsg = append(theReturnMessage.ResultMsg, aMessage)
		theReturnMessage.TheErr = append(theReturnMessage.TheErr, aMessage)
	} else {
		aMessage := "Success getting hotdog and hamburger messageboards"
		theReturnMessage.SuccOrFail = 0
		theReturnMessage.ResultMsg = append(theReturnMessage.ResultMsg, aMessage)
	}

	//Write the Reponse back
	theJSONMessage, err := json.Marshal(theReturnMessage)
	//Send the response back
	if err != nil {
		errIs := "Error formatting JSON for return in isMessageBoardCreated: " + err.Error()
		logWriter(errIs)
	}
	fmt.Fprint(w, string(theJSONMessage))
}
