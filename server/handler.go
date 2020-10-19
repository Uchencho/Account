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

type loginInfo struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func Register(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:

		var (
			userPayload RegisterUser
			err         error
		)

		err = json.NewDecoder(req.Body).Decode(&userPayload)
		if err != nil {
			InvalidJsonResp(w, err)
			return
		}

		err, aboveOneField := validateInput(userPayload)
		if aboveOneField {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error" : "Invalid Payload"}`)
			return
		}
		if err != nil {
			InvalidJsonResp(w, err)
			return
		}

		user := User{
			Email:      userPayload.Email,
			DateJoined: time.Now(),
			LastLogin:  time.Now(),
			IsActive:   true,
			FirstName:  userPayload.FirstName,
		}
		user.HashedPassword, err = HashPassword(userPayload.Password)
		if err != nil {
			InternalIssues(w)
			return
		}

		err = addUser(Client, user)
		if err != nil {
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

func Login(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case http.MethodPost:

		var loginDetails loginInfo

		err := json.NewDecoder(req.Body).Decode(&loginDetails)
		if err != nil {
			InvalidJsonResp(w, err)
			return
		}
		err, aboveOneField := validateInput(loginDetails)
		if aboveOneField {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error" : "Invalid Payload"}`)
			return
		}
		if err != nil {
			InvalidJsonResp(w, err)
			return
		}

		user, err := getUser(Client, loginDetails.Email)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error" : "User does not exist"}`)
			return
		}

		err = CheckPasswordHash(loginDetails.Password, user.HashedPassword)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error" : "Email/Password is incorrect"}`)
			return
		}

		accessToken, refreshToken, err := GenerateToken(user.Email)
		if err != nil {
			InternalIssues(w)
			return
		}

		logRes := loginResponse{
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
	}
}

func addUser(client *mongo.Client, userDetails User) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := client.Database("account").Collection("user")

	// check if user exists
	var duplicateUser User
	filter := bson.M{"email": userDetails.Email}
	err := collection.FindOne(ctx, filter).Decode(&duplicateUser)
	if err != nil {
		// User does not exist
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

func getUser(client *mongo.Client, email string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := client.Database("account").Collection("user")

	// check if user exists
	var userDetails User
	filter := bson.M{"email": email}
	err := collection.FindOne(ctx, filter).Decode(&userDetails)

	if err != nil {
		return User{}, err
	}
	return userDetails, nil
}
