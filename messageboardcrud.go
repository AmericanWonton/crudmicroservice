package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
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

//A simple insert that is called within the CRUD Mircorservice
func insertOneNewMessageSimple(newMessage Message) {
	//Send this to the 'message' collection for safekeeping
	messageCollection := mongoClient.Database("microservice").Collection("messages") //Here's our collection
	collectedStuff := []interface{}{newMessage}
	//Insert Our Data
	_, err := messageCollection.InsertMany(context.TODO(), collectedStuff)
	if err != nil {
		theErr := "Error writing insert message in insertOneNewMessageSimple in crudoperations: " + err.Error()
		logWriter(theErr)
	} else {
		theErr := "User successfully inserted message in insertOneNewMessageSimple in crudoperations: "
		logWriter(theErr)
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

	//Update Message
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

//Simple update to one message,(called from this Microservice)
func updateOneMessageSimple(updatedMessage Message) {
	//Update Message
	theTimeNow := time.Now()
	messageCollection := mongoClient.Database("microservice").Collection("messages") //Here's our collection
	theFilter := bson.M{
		"messageid": bson.M{
			"$eq": updatedMessage.MessageID,
		},
	}
	updatedDocument := bson.M{
		"$set": bson.M{
			"messageid":       updatedMessage.MessageID,
			"userid":          updatedMessage.UserID,
			"postername":      updatedMessage.PosterName,
			"messages":        updatedMessage.Messages,
			"ischild":         updatedMessage.IsChild,
			"haschildren":     updatedMessage.HasChildren,
			"parentmessageid": updatedMessage.ParentMessageID,
			"uberparentid":    updatedMessage.UberParentID,
			"order":           updatedMessage.Order,
			"repliesamount":   updatedMessage.RepliesAmount,
			"themessage":      updatedMessage.TheMessage,
			"whatboard":       updatedMessage.WhatBoard,
			"datecreated":     updatedMessage.DateCreated,
			"lastupdated":     theTimeNow.Format("2006-01-02 15:04:05"),
		},
	}
	_, err2 := messageCollection.UpdateOne(theContext, theFilter, updatedDocument)
	if err2 != nil {
		message := "Got an error updating a message in updateOneMessageSimple: " + err2.Error()
		fmt.Println(message)
		logWriter(message)
	} else {
		message := "We updated the document successfully"
		fmt.Println(message)
		logWriter(message)
	}
}

//This func updates the Mongo MessageBoard collection to what we have now
func updateMongoMessageBoard(w http.ResponseWriter, r *http.Request) {
	//This is the return message
	type ReturnMessage struct {
		TheErr     []string `json:"TheErr"`
		ResultMsg  []string `json:"ResultMsg"`
		SuccOrFail int      `json:"SuccOrFail"`
	}
	theReturnMessage := ReturnMessage{}
	theReturnMessage.SuccOrFail = 0 //Declare this first to check against failure when returning response

	type UpdatedMongoBoard struct {
		UpdatedMessageBoard MessageBoard `json:"UpdatedMessageBoard"`
	}
	//Unwrap from JSON
	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		theErr := "Error reading the request from updateMongoMessageBoard: " + err.Error() + "\n" + string(bs)
		logWriter(theErr)
		fmt.Println(theErr)
	}

	//Marshal it into our type
	var theboardUpdate UpdatedMongoBoard
	json.Unmarshal(bs, &theboardUpdate)

	message_collection := mongoClient.Database("microservice").Collection("messageboard") //Here's our collection
	theFilter := bson.M{
		"messageboardid": bson.M{
			"$eq": theboardUpdate.UpdatedMessageBoard.MessageBoardID, // check if test value is present for reply Message
		},
		"boardname": bson.M{
			"$eq": theboardUpdate.UpdatedMessageBoard.BoardName, // check if test value is present for reply Message
		},
	}

	updatedDocument := bson.M{}
	updatedDocument = bson.M{
		"$set": bson.M{
			"messageboardid":         theboardUpdate.UpdatedMessageBoard.MessageBoardID,
			"boardname":              theboardUpdate.UpdatedMessageBoard.BoardName,
			"allmessages":            theboardUpdate.UpdatedMessageBoard.AllMessages,
			"allmessagesmap":         theboardUpdate.UpdatedMessageBoard.AllMessagesMap,
			"alloriginalmessages":    theboardUpdate.UpdatedMessageBoard.AllOriginalMessages,
			"alloriginalmessagesmap": theboardUpdate.UpdatedMessageBoard.AllOriginalMessagesMap,
			"lastupdated":            theboardUpdate.UpdatedMessageBoard.LastUpdated,
			"datecreated":            theboardUpdate.UpdatedMessageBoard.DateCreated,
		},
	}

	stuffUpdated, err2 := message_collection.UpdateOne(theContext, theFilter, updatedDocument)
	if err2 != nil {
		message := "We got an error updating this document: " + err2.Error()
		fmt.Println(message)
		logWriter(message)
		theReturnMessage.TheErr = append(theReturnMessage.TheErr, message)
		theReturnMessage.ResultMsg = append(theReturnMessage.ResultMsg, message)
		theReturnMessage.SuccOrFail = 1
	} else {
		theInt := strconv.Itoa(int(stuffUpdated.MatchedCount))
		message := "Here is the update for our messageboard: " + theInt
		fmt.Println(message)
		logWriter(message)
		theReturnMessage.ResultMsg = append(theReturnMessage.ResultMsg, message)
	}

	//Determine if we have an error yet and what to return
	if theReturnMessage.SuccOrFail != 0 {
		//Log a failure and return the failure
		aMessage := "Failure updating hotdog or hamburger messageboards"
		theReturnMessage.ResultMsg = append(theReturnMessage.ResultMsg, aMessage)
		theReturnMessage.TheErr = append(theReturnMessage.TheErr, aMessage)
	} else {
		aMessage := "Success updating hotdog and hamburger messageboards"
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

/*This func is a simple update of the messageboards, no API needed
(Returns a bool for a failure and a string for failure/success messages)
*/
func updateMongoMessageBoardSimple(updatedMessageBoard MessageBoard) (bool, string) {
	//Declare variables for logging to return
	goodUpdate, updateRecord := true, ""

	message_collection := mongoClient.Database("microservice").Collection("messageboard") //Here's our collection
	theFilter := bson.M{
		"messageboardid": bson.M{
			"$eq": updatedMessageBoard.MessageBoardID, // check if test value is present for reply Message
		},
		"boardname": bson.M{
			"$eq": updatedMessageBoard.BoardName, // check if test value is present for reply Message
		},
	}

	updatedDocument := bson.M{}
	updatedDocument = bson.M{
		"$set": bson.M{
			"messageboardid":         updatedMessageBoard.MessageBoardID,
			"boardname":              updatedMessageBoard.BoardName,
			"allmessages":            updatedMessageBoard.AllMessages,
			"allmessagesmap":         updatedMessageBoard.AllMessagesMap,
			"alloriginalmessages":    updatedMessageBoard.AllOriginalMessages,
			"alloriginalmessagesmap": updatedMessageBoard.AllOriginalMessagesMap,
			"lastupdated":            updatedMessageBoard.LastUpdated,
			"datecreated":            updatedMessageBoard.DateCreated,
		},
	}

	stuffUpdated, err2 := message_collection.UpdateOne(theContext, theFilter, updatedDocument)
	if err2 != nil {
		message := "We got an error updating this document: " + err2.Error()
		fmt.Println(message)
		updateRecord = updateRecord + message
		goodUpdate = false
	} else {
		theInt := strconv.Itoa(int(stuffUpdated.MatchedCount))
		message := "Here is the update for our messageboard: " + theInt
		fmt.Println(message)
		updateRecord = updateRecord + message
	}

	return goodUpdate, updateRecord
}

//This updates the entirety of the messageboard and returns it to the mainpage Microservice
func uberUpdate(w http.ResponseWriter, r *http.Request) {
	theTimeNow := time.Now() //Used for time updates
	//Return message
	type ReturnMessage struct {
		TheErr             []string        `json:"TheErr"`
		ResultMsg          []string        `json:"ResultMsg"`
		SuccOrFail         int             `json:"SuccOrFail"`
		GivenHDogMB        MessageBoard    `json:"GivenHDogMB"`
		GivenHamMB         MessageBoard    `json:"GivenHamMB"`
		GivenLoadedMapHDog map[int]Message `json:"GivenLoadedMapHDog"`
		GivenLoadedMapHam  map[int]Message `json:"GivenLoadedMapHam"`
	}
	theReturnMessage := ReturnMessage{}
	theReturnMessage.SuccOrFail = 0 //Declare this first to check against failure when returning response
	//Declare type we want JSON marshaled into
	type UberUpdateMessages struct {
		TheNewestMessage Message         `json:"TheNewestMessage"`
		TheParentMessage Message         `json:"TheParentMessage"`
		WhatBoard        string          `json:"WhatBoard"`
		HotdogMB         MessageBoard    `json:"HotdogMB"`
		HamburgerMB      MessageBoard    `json:"HamburgerMB"`
		LoadedMapHDog    map[int]Message `json:"LoadedMapHDog"`
		LoadedMapHam     map[int]Message `json:"LoadedMapHam"`
	}
	//Unwrap from JSON
	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		theErr := "Error reading the request from uberUpdate: " + err.Error() + "\n" + string(bs)
		logWriter(theErr)
		fmt.Println(theErr)
	}

	//Marshal it into our type
	var uberUpdateMessages UberUpdateMessages
	json.Unmarshal(bs, &uberUpdateMessages)

	/* Begin  the uber update */
	if uberUpdateMessages.TheParentMessage.IsChild == false {
		//This is the uberParent; simply add this to the []Message list
		uberUpdateMessages.TheParentMessage.Messages = append(uberUpdateMessages.TheParentMessage.Messages, uberUpdateMessages.TheNewestMessage)
		//parentMessage.Messages = append([]Message{newestMessage}, parentMessage.Messages...)
		uberUpdateMessages.TheParentMessage.RepliesAmount = uberUpdateMessages.TheParentMessage.RepliesAmount + 1
		uberUpdateMessages.TheParentMessage.HasChildren = true
		uberUpdateMessages.TheParentMessage.LastUpdated = theTimeNow.Format("2006-01-02 15:04:05")
		//Update Message in 'messages' table
		updateOneMessageSimple(uberUpdateMessages.TheParentMessage)
		//Update the messages tables
		insertOneNewMessageSimple(uberUpdateMessages.TheNewestMessage)
		//Update MessageBoard properties, (based on which Messageboard)
		switch uberUpdateMessages.WhatBoard {
		case "hotdog":
			//Initial Clear for no map goofs
			theReturnMessage.GivenLoadedMapHDog = make(map[int]Message)
			theReturnMessage.GivenLoadedMapHam = make(map[int]Message)

			uberUpdateMessages.HotdogMB.AllMessagesMap[uberUpdateMessages.TheParentMessage.MessageID] = uberUpdateMessages.TheParentMessage         //Update UberParent
			uberUpdateMessages.HotdogMB.AllOriginalMessagesMap[uberUpdateMessages.TheParentMessage.MessageID] = uberUpdateMessages.TheParentMessage //updateUberParent
			uberUpdateMessages.HotdogMB.AllMessages = append(uberUpdateMessages.HotdogMB.AllMessages, uberUpdateMessages.TheNewestMessage)          //Add newest message
			uberUpdateMessages.HotdogMB.AllMessagesMap[uberUpdateMessages.TheNewestMessage.MessageID] = uberUpdateMessages.TheNewestMessage         //add newest message
			uberUpdateMessages.HotdogMB.LastUpdated = theTimeNow.Format("2006-01-02 15:04:05")
			//update uberParent
			for j := 0; j < len(uberUpdateMessages.HotdogMB.AllOriginalMessages); j++ {
				if uberUpdateMessages.HotdogMB.AllOriginalMessages[j].MessageID == uberUpdateMessages.TheParentMessage.MessageID {
					uberUpdateMessages.HotdogMB.AllOriginalMessages[j] = uberUpdateMessages.TheParentMessage
					//Update the loadedMessageMap
					uberUpdateMessages.LoadedMapHDog[j+1] = uberUpdateMessages.TheParentMessage
					break
				}
			}
			//Update Mongo Collections
			dbUpdateResult, theErr := updateMongoMessageBoardSimple(uberUpdateMessages.HotdogMB)
			fmt.Printf("DEBUG: Here is how the update Messageboard went: %v and %v\n", dbUpdateResult, theErr)

			//Update the return message with for all of our maps and messageboards
			theReturnMessage.GivenHDogMB = uberUpdateMessages.HotdogMB
			theReturnMessage.GivenLoadedMapHDog = uberUpdateMessages.LoadedMapHDog
			theReturnMessage.ResultMsg = append(theReturnMessage.ResultMsg, "Updated the hotdog messageboard in uberUpdate")
			break
		case "hamburger":
			//Initial Clear for no map goofs
			theReturnMessage.GivenLoadedMapHDog = make(map[int]Message)
			theReturnMessage.GivenLoadedMapHam = make(map[int]Message)

			uberUpdateMessages.HamburgerMB.AllMessagesMap[uberUpdateMessages.TheParentMessage.MessageID] = uberUpdateMessages.TheParentMessage         //Update UberParent
			uberUpdateMessages.HamburgerMB.AllOriginalMessagesMap[uberUpdateMessages.TheParentMessage.MessageID] = uberUpdateMessages.TheParentMessage //updateUberParent
			uberUpdateMessages.HamburgerMB.AllMessages = append(uberUpdateMessages.HamburgerMB.AllMessages, uberUpdateMessages.TheNewestMessage)       //Add newest message
			uberUpdateMessages.HamburgerMB.AllMessagesMap[uberUpdateMessages.TheNewestMessage.MessageID] = uberUpdateMessages.TheNewestMessage         //add newest message
			uberUpdateMessages.HamburgerMB.LastUpdated = theTimeNow.Format("2006-01-02 15:04:05")
			//update uberParent
			for j := 0; j < len(uberUpdateMessages.HamburgerMB.AllOriginalMessages); j++ {
				if uberUpdateMessages.HamburgerMB.AllOriginalMessages[j].MessageID == uberUpdateMessages.TheParentMessage.MessageID {
					uberUpdateMessages.HamburgerMB.AllOriginalMessages[j] = uberUpdateMessages.TheParentMessage
					//Update the loadedMessageMap
					uberUpdateMessages.LoadedMapHam[j+1] = uberUpdateMessages.TheParentMessage
					break
				}
			}
			//Update Mongo Collections
			dbUpdateResult, theErr := updateMongoMessageBoardSimple(uberUpdateMessages.HamburgerMB)
			fmt.Printf("DEBUG: Here is how the update Messageboard went: %v and %v\n", dbUpdateResult, theErr)
			//Update the return message with for all of our maps and messageboards
			theReturnMessage.GivenHamMB = uberUpdateMessages.HamburgerMB
			theReturnMessage.GivenLoadedMapHam = uberUpdateMessages.LoadedMapHam
			theReturnMessage.ResultMsg = append(theReturnMessage.ResultMsg, "Updated the hamburger messageboard in uberUpdate")
			break
		default:
			//Wrong board entered...write the error
			message := "Wrong board entered in uberUpdate: " + uberUpdateMessages.WhatBoard
			theReturnMessage.SuccOrFail = 1
			theReturnMessage.ResultMsg = append(theReturnMessage.ResultMsg, message)
			theReturnMessage.TheErr = append(theReturnMessage.TheErr, message)
			break
		}
	} else {
		//Add newest message to parent message to update it
		uberUpdateMessages.TheParentMessage.HasChildren = true //Had it or before, now this has children
		uberUpdateMessages.TheParentMessage.RepliesAmount = uberUpdateMessages.TheParentMessage.RepliesAmount + 1
		uberUpdateMessages.TheParentMessage.Messages = append(uberUpdateMessages.TheParentMessage.Messages, uberUpdateMessages.TheNewestMessage)
		uberUpdateMessages.TheParentMessage.LastUpdated = theTimeNow.Format("2006-01-02 15:04:05")
		//Update Message in 'messages' table
		updateOneMessageSimple(uberUpdateMessages.TheParentMessage)
		//Insert new Message
		insertOneNewMessageSimple(uberUpdateMessages.TheNewestMessage)
		//Update MessageBoard properties, (based on which Messageboard)
		switch uberUpdateMessages.WhatBoard {
		case "hotdog":
			//Search our message board for the Uber Parent
			uberParentMessage := uberUpdateMessages.HotdogMB.AllOriginalMessagesMap[uberUpdateMessages.TheParentMessage.UberParentID]
			//Update all parent messages recursivley until we finally update the uberParentMessage
			uberParentMessage, uberUpdateMessages.HotdogMB = updateToUber(uberParentMessage, uberUpdateMessages.TheParentMessage, uberUpdateMessages.HotdogMB)
			fmt.Printf("DEBUG: Here is our UberParent Update: %v\n", uberParentMessage)
			fmt.Println()
			fmt.Println()
			//Update MessageBoard properties
			uberUpdateMessages.HotdogMB.AllMessagesMap[uberParentMessage.MessageID] = uberParentMessage                                                  //Update UberParent
			uberUpdateMessages.HotdogMB.AllOriginalMessagesMap[uberParentMessage.MessageID] = uberParentMessage                                          //updateUberParent
			uberUpdateMessages.HotdogMB.AllMessages = append([]Message{uberUpdateMessages.TheNewestMessage}, uberUpdateMessages.HotdogMB.AllMessages...) //Add newest message
			uberUpdateMessages.HotdogMB.AllMessagesMap[uberUpdateMessages.TheNewestMessage.MessageID] = uberUpdateMessages.TheNewestMessage              //add newest message
			uberUpdateMessages.HotdogMB.LastUpdated = theTimeNow.Format("2006-01-02 15:04:05")
			//update uberParent
			for j := 0; j < len(uberUpdateMessages.HotdogMB.AllOriginalMessages); j++ {
				if uberUpdateMessages.HotdogMB.AllOriginalMessages[j].MessageID == uberParentMessage.MessageID {
					uberUpdateMessages.HotdogMB.AllOriginalMessages[j] = uberParentMessage
					//Update the loadedMessageMap
					uberUpdateMessages.LoadedMapHDog[j+1] = uberParentMessage
				}
			}

			//Update Mongo Collections
			updateMongoMessageBoardSimple(uberUpdateMessages.HotdogMB)
			//Update Message in 'messages' table
			updateOneMessageSimple(uberParentMessage)

			//Update the return message with for all of our maps and messageboards
			theReturnMessage.GivenHDogMB = uberUpdateMessages.HotdogMB
			theReturnMessage.GivenLoadedMapHDog = uberUpdateMessages.LoadedMapHDog
			theReturnMessage.ResultMsg = append(theReturnMessage.ResultMsg, "Updated the hotdog messageboard in uberUpdate")
			break
		case "hamburger":
			//Search our message board for the Uber Parent
			uberParentMessage := uberUpdateMessages.HamburgerMB.AllOriginalMessagesMap[uberUpdateMessages.TheParentMessage.UberParentID]
			//Update all parent messages recursivley until we finally update the uberParentMessage
			uberParentMessage, uberUpdateMessages.HamburgerMB = updateToUber(uberParentMessage, uberUpdateMessages.TheParentMessage, uberUpdateMessages.HamburgerMB)
			//Update MessageBoard properties
			uberUpdateMessages.HamburgerMB.AllMessagesMap[uberParentMessage.MessageID] = uberParentMessage                                                     //Update UberParent
			uberUpdateMessages.HamburgerMB.AllOriginalMessagesMap[uberParentMessage.MessageID] = uberParentMessage                                             //updateUberParent
			uberUpdateMessages.HamburgerMB.AllMessages = append([]Message{uberUpdateMessages.TheNewestMessage}, uberUpdateMessages.HamburgerMB.AllMessages...) //Add newest message
			uberUpdateMessages.HamburgerMB.AllMessagesMap[uberUpdateMessages.TheNewestMessage.MessageID] = uberUpdateMessages.TheNewestMessage                 //add newest message
			uberUpdateMessages.HamburgerMB.LastUpdated = theTimeNow.Format("2006-01-02 15:04:05")
			//update uberParent
			for j := 0; j < len(uberUpdateMessages.HamburgerMB.AllOriginalMessages); j++ {
				if uberUpdateMessages.HamburgerMB.AllOriginalMessages[j].MessageID == uberParentMessage.MessageID {
					uberUpdateMessages.HamburgerMB.AllOriginalMessages[j] = uberParentMessage
					//Update the loadedMessageMap
					uberUpdateMessages.LoadedMapHam[j+1] = uberParentMessage
				}
			}

			//Update Mongo Collections
			updateMongoMessageBoardSimple(uberUpdateMessages.HamburgerMB)
			//Update Message in 'messages' table
			updateOneMessageSimple(uberParentMessage)

			//Update the return message with for all of our maps and messageboards
			theReturnMessage.GivenHamMB = uberUpdateMessages.HamburgerMB
			theReturnMessage.GivenLoadedMapHam = uberUpdateMessages.LoadedMapHam
			theReturnMessage.ResultMsg = append(theReturnMessage.ResultMsg, "Updated the Hamburger messageboard in uberUpdate")
			break
		default:
			message := "Wrong board entered in uberUpdate: " + uberUpdateMessages.WhatBoard
			theReturnMessage.SuccOrFail = 1
			theReturnMessage.ResultMsg = append(theReturnMessage.ResultMsg, message)
			theReturnMessage.TheErr = append(theReturnMessage.TheErr, message)
			break
		}
	}
	//Write the Reponse back
	theJSONMessage, err := json.Marshal(theReturnMessage)
	//Send the response back
	if err != nil {
		errIs := "Error formatting JSON for return in uberUpdate: " + err.Error()
		logWriter(errIs)
	}
	fmt.Fprint(w, string(theJSONMessage))
}

//This func finds the parent FROM THE UBERPARENT to update; called from within this service
func updateToUber(uberParentMessage Message, parentMessage Message, theboard MessageBoard) (Message, MessageBoard) {
	theTimeNow := time.Now()      //Used for updating time properties in our parent
	uberParentUpdated := false    //Are we currently on the UberParent Message, matching their ID?
	finalUberMessage := Message{} //The final message with the updated UberParent
	//Loop and update until we find the UberParent and update it into the 'finalUberMessage'
	pastParent := parentMessage
	currentMessage := theboard.AllMessagesMap[parentMessage.ParentMessageID] //First set the searcher parent to it's OWN parent
	for {
		if uberParentUpdated == true {
			break //UberParent is found and updated, ready to be returned. End this updating search
		} else {
			/* Determine if the current parent is an uberParent */
			if currentMessage.MessageID == uberParentMessage.MessageID {
				//This is the parent update UberParent for us to return and update in Mongo, then break!
				uberParentUpdated = true //Set break value
				/* Step 1: Update the parentMessage in the UberParent */
				for v := 0; v < len(currentMessage.Messages); v++ {
					//Search fo rparent in UberParent Messages then update it
					if currentMessage.Messages[v].MessageID == pastParent.MessageID {
						currentMessage.Messages[v] = pastParent
						break
					}
				}
				/* Step 2: Update the Past Parent in the the messageboard table */
				pastParent.LastUpdated = theTimeNow.Format("2006-01-02 15:04:05")
				theboard.AllMessagesMap[pastParent.MessageID] = pastParent
				/* DEBUG: We should add a goroutine to update other sections of the MessageBoard
				that ARENT' maps and quickly updated */
				/* Step 2: Update finalUberMessage to return for final updating */
				finalUberMessage = currentMessage
			} else {
				/*
					This is not the UberParent. We will update the currentMessage in all appropriate spots,
					THEN we will assign the currentMessage as the ParentID, then move this currentMessage to
					be the 'pastParent'
				*/
				//Step 1: Update the currentMessage's Message Array with the updated parentMessage
				for x := 0; x < len(currentMessage.Messages); x++ {
					if currentMessage.Messages[x].MessageID == pastParent.MessageID {
						//We found the pastParent in the currentMessage Messages array. Update it
						pastParent.LastUpdated = theTimeNow.Format("2006-01-02 15:04:05")
						currentMessage.Messages[x] = pastParent
						break
					}
				}
				//Step 2: Update the messageBoard so we won't have infinite loops
				currentMessage.LastUpdated = theTimeNow.Format("2006-01-02 15:04:05")
				pastParent.LastUpdated = theTimeNow.Format("2006-01-02 15:04:05")
				theboard.AllMessagesMap[currentMessage.MessageID] = currentMessage
				theboard.AllMessagesMap[pastParent.MessageID] = pastParent
				/* DEBUG: We should add a goroutine to update other sections of the MessageBoard
				that ARENT' maps and quickly updated */
				/* Step 3: Update the search criteria until we hit the uberParent */
				pastParent = currentMessage
				currentMessage = theboard.AllMessagesMap[currentMessage.ParentMessageID] //First set the searcher parent to it's OWN parent
			}
		}
	}

	return finalUberMessage, theboard //This should be the completed UberMessage
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
