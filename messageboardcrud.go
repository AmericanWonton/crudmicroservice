package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
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
	PosterName      string    `json:"PosterName"`      //Username of the poster
	Messages        []Message `json:"Messages"`        //Array of Messages under this one
	IsChild         bool      `json:"IsChild"`         //Is this message childed to another message
	HasChildren     bool      `json:"HasChildren"`     //Whether this message has children to list
	ParentMessageID int       `json:"ParentMessageID"` //The ID of this parent
	UberParentID    int       `json:"UberParentID"`    //The final parent of this parent, IF EQUAL PARENT
	Order           int       `json:"Order"`           //Order the comment is in with it's reply tree
	RepliesAmount   int       `json:"RepliesAmount"`   //Amount of replies this message has
	TheMessage      string    `json:"TheMessage"`      //The MEssage in the post
	DateCreated     string    `json:"DateCreated"`     //When the message was created
	LastUpdated     string    `json:"LastUpdated"`     //When the message was last updated
}

//All the Messages on the board
type MessageBoard struct {
	MessageBoardID         int             `json:"MessageBoardID"`         //The Random ID of this Messageboard
	BoardName              string          `json:"BoardName"`              //Name of this messageboard
	AllMessages            []Message       `json:"AllMessages"`            //All the IDs listed
	AllMessagesMap         map[int]Message `json:"AllMessagesMap"`         //A map of ALL messages
	AllOriginalMessages    []Message       `json:"AllOriginalMessages"`    //All the messages that AREN'T replies
	AllOriginalMessagesMap map[int]Message `json:"AllOriginalMessagesMap"` //Map of original Messages
	LastUpdated            string          `json:"LastUpdated"`            //Last time this messageboard was updated
}

var loadedMessagesMap map[int]Message
var theMessageBoard MessageBoard //The board containing all our messages
/* This is the current amount of results our User is looking at
it changes as the User clicks forwards or backwards for more results */
var currentPageNumber int = 1

//Inserts one message into our 'messages' collection
func insertOneNewMessage(w http.ResponseWriter, r *http.Request) {

	type MessageInsert struct {
		MessageInserted Message `json:"MessageInserted"`
	}
	//Unwrap from JSON
	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		theErr := "Error reading the request from insertOneNewMessage: " + err.Error() + "\n" + string(bs)
		logWriter(theErr)
		fmt.Println(theErr)
	}

	//Marshal it into our type
	var messageIn MessageInsert
	json.Unmarshal(bs, &messageIn)

	//Send this to the 'message' collection for safekeeping
	messageCollection := mongoClient.Database("microservice").Collection("messages") //Here's our collection
	collectedStuff := []interface{}{messageIn}
	//Insert Our Data
	insertManyResult, err := messageCollection.InsertMany(context.TODO(), collectedStuff)
	//Declare data to return
	type ReturnMessage struct {
		TheErr     string `json:"TheErr"`
		ResultMsg  string `json:"ResultMsg"`
		SuccOrFail int    `json:"SuccOrFail"`
	}
	fmt.Printf("DEBUG: Need to print this, IDE won't shut up: %v\n", insertManyResult)
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
		{"MessageID": postedMessageID.MessageID},
	} //Here's our filter to look for
	deletes = append(deletes, bson.M{"MessageID": bson.M{
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
	type MessageUpdate struct {
		UpdatedMessage Message `json:"UpdatedMessage"`
	}
	//Unwrap from JSON
	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		theErr := "Error reading the request from updateOneMessage: " + err.Error() + "\n" + string(bs)
		logWriter(theErr)
		fmt.Println(theErr)
	}

	//Marshal it into our type
	var theMessageUpdate MessageUpdate
	json.Unmarshal(bs, &theMessageUpdate)

	//Update User
	theTimeNow := time.Now()
	messageCollection := mongoClient.Database("microservice").Collection("messages") //Here's our collection
	theFilter := bson.M{
		"messageid": bson.M{
			"$eq": theMessageUpdate.UpdatedMessage.MessageID,
		},
	}
	updatedDocument := bson.M{
		"$set": bson.M{
			"messageid":       theMessageUpdate.UpdatedMessage.MessageID,
			"userid":          theMessageUpdate.UpdatedMessage.UserID,
			"postername":      theMessageUpdate.UpdatedMessage.PosterName,
			"messages":        theMessageUpdate.UpdatedMessage.Messages,
			"ischild":         theMessageUpdate.UpdatedMessage.IsChild,
			"haschildren":     theMessageUpdate.UpdatedMessage.HasChildren,
			"parentmessageid": theMessageUpdate.UpdatedMessage.ParentMessageID,
			"uberparentid":    theMessageUpdate.UpdatedMessage.UberParentID,
			"order":           theMessageUpdate.UpdatedMessage.Order,
			"repliesamount":   theMessageUpdate.UpdatedMessage.RepliesAmount,
			"themessage":      theMessageUpdate.UpdatedMessage.TheMessage,
			"datecreated":     theMessageUpdate.UpdatedMessage.DateCreated,
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
