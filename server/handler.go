package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func Register(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	switch req.Method {
	case http.MethodPost:

		var (
			userPayload RegisterUser
			user        User
			err         error
		)

		err = json.NewDecoder(req.Body).Decode(&userPayload)
		if err != nil {
			InvalidJsonResp(w, err)
			return
		}

		err, aboveOneField := ValidateInput(userPayload)
		if aboveOneField {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error" : "Invalid Payload"}`)
			return
		}
		if err != nil {
			InvalidJsonResp(w, err)
			return
		}

		user.Email = userPayload.Email
		user.DateJoined = time.Now()
		user.LastLogin = time.Now()
		user.HashedPassword, err = HashPassword(userPayload.Password)
		if err != nil {
			InternalIssues(w)
			return
		}

		err = AddUser(Client, user)
		if err != nil {
			log.Printf("Error occured in adding details to db, %v", err)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error" : "User already exists, please login"}`)
			return
		}

		accessToken, refreshToken, err := GenerateToken(user.Email)
		if err != nil {
			InternalIssues(w)
			return
		}

		logRes := loginResponse{
			ID:           user.ID,
			Email:        user.Email,
			FirstName:    user.FirstName,
			IsActive:     user.IsActive,
			DateJoined:   user.DateJoined,
			LastLogin:    user.LastLogin,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}
		successResp := SuccessResponse{
			Message: "success",
			Data:    logRes,
		}
		jsonResp, err := json.Marshal(successResp)
		if err != nil {
			InternalIssues(w)
		}

		fmt.Fprint(w, string(jsonResp))
		return

	default:
		MethodNotAllowedResponse(w)
		return
	}
}

func AddUser(client *mongo.Client, userDetails User) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := client.Database("account").Collection("user")

	// check if user exists
	var duplicateUser User
	filter := bson.M{"email": userDetails.Email}
	err := collection.FindOne(ctx, filter).Decode(&duplicateUser)
	if err != nil {
		_, err = collection.InsertOne(ctx, userDetails)
		if err != nil {
			log.Println("Error in inserting item with error, ", err)
			return err
		}
		return nil
	}

	if duplicateUser.Email == userDetails.Email {
		return errors.New("User already exists")
	}
	return nil
}