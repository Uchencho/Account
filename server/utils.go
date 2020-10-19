package server

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator"
)

var (
	validate = validator.New()
)

type loginResponse struct {
	Email        string    `json:"email"`
	FirstName    string    `json:"first_name"`
	PhoneNumber  string    `json:"phone_number"`
	UserAddress  string    `json:"user_address"`
	IsActive     bool      `json:"is_active"`
	DateJoined   time.Time `json:"date_joined"`
	LastLogin    time.Time `json:"last_login"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func InvalidJsonResp(w http.ResponseWriter, err error) {
	if err.Error() == "EOF" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error" : "Invalid Payload"}`)
		return
	}
	log.Printf("error in decoding json: %v", err)
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, `{"error" : "%s"}`, err.Error())
}

func MethodNotAllowedResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	fmt.Fprint(w, `{"message" : "Method Not allowed"}`)
}

func InternalIssues(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprint(w, `{"error" : "Something went wrong"}`)
}

// Validates a struct
func validateInput(object interface{}) (error, bool) {

	err := validate.Struct(object)
	if err != nil {

		//Validation syntax is invalid
		if _, ok := err.(*validator.InvalidValidationError); ok {
			log.Println(err)
			return err, false
		}

		if len(err.(validator.ValidationErrors)) > 1 {
			log.Println("Error is more than one")
			return err, true
		}

		for _, err := range err.(validator.ValidationErrors) {

			// Retrieve json field
			reflectedValue := reflect.ValueOf(object)
			field, _ := reflectedValue.Type().FieldByName(err.StructField())

			var name string
			if name = field.Tag.Get("json"); name == "" {
				name = strings.ToLower(err.StructField())
			}

			switch err.Tag() {
			case "required":
				return fmt.Errorf("%s is required", name), false
			case "email":
				return fmt.Errorf("%s should be a valid email address", name), false
			case "eqfield":
				return fmt.Errorf("%s should be the same as %s", name, err.Param()), false
			default:
				log.Println(err.Tag())
				return fmt.Errorf("%s is Invalid", name), false
			}
		}
		return err, false
	}
	return nil, false
}
