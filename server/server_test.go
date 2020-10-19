package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHashPassword(t *testing.T) {
	_, err := HashPassword("myStrongPassword")
	if err != nil {
		t.Fatalf("Failed to hash password with error %s", err)
	}

	_, err = HashPassword("")
	if err == nil {
		t.Fatalf("Hashing an invalid password")
	}
}

func TestCheckPasswordHash(t *testing.T) {

	hashedPassword, err := HashPassword("myStrongPassword")
	if err != nil {
		t.Errorf("Strange, unable to hash password with error %s", err)
	}

	if err := CheckPasswordHash("myStrongPassword", hashedPassword); err != nil {
		t.Fatalf("Checkpassword is not retrieving current password")
	}
}

func TestGenerateToken(t *testing.T) {

	_, _, err := GenerateToken("uche@gmail.com")
	if err != nil {
		t.Fatalf("Could not generate tokens with error, %s", err)
	}

	_, _, err = GenerateToken("")
	if err == nil {
		t.Fatalf("Generating token for an empty string as email")
	}
}

func TestInvalidRegisterRequest(t *testing.T) {
	getReq, err := http.NewRequest("GET", "/api/register", nil)
	if err != nil {
		t.Fatal("Could not create request with error ", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Register)
	handler.ServeHTTP(rr, getReq)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Fatalf("Hanlder returned invalid response for invalid request")
	}
}

func TestRegister(t *testing.T) {
	postReq, err := http.NewRequest("POST", "/api/register", bytes.NewBuffer([]byte("")))
	if err != nil {
		t.Fatal("Could not create request with error ", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Register)
	handler.ServeHTTP(rr, postReq)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Fatalf("Hanlder returned invalid response for invalid payload")
	}
}

// func TestBasicToken(t *testing.T) {
// 	// postReq, err := http.NewRequest("POST", "/api/register", bytes.NewBuffer([]byte("")))
// 	// if err != nil {
// 	// 	t.Fatal("Could not create request with error ", err)
// 	// }

// 	rr := httptest.NewRecorder()
// 	http.Handle("/api/register", BasicToken(http.HandlerFunc(Register)))

// 	if status := rr.Code; status != http.StatusForbidden {
// 		log.Println(rr.Code, rr.Body.String(), "nothing")
// 		t.Fatalf("Handler returned invalid response for request header")
// 	}

// 	if rr.Body.String() != `{"error" : "Token not passed"}` {
// 		t.Fatalf("Invalid response body")
// 	}
// }

func TestValidate(t *testing.T) {
	payload := RegisterUser{
		Email:           "Not-AN-Email",
		Password:        "myStrongPassword",
		ConfirmPassword: "myStrongPassword",
	}
	err, aboveOne := validateInput(payload)
	if err == nil {
		t.Fatal("Could not invalidate the error email")
	}
	if err.Error() != "email should be a valid email address" {
		t.Fatal("Error message is inconsistent with error, ", err)
	}
	if aboveOne {
		t.Fatal("Recorded more errors than expected")
	}

	payload = RegisterUser{
		Email:           "Not-AN-Email",
		Password:        "myStrongPasswords",
		ConfirmPassword: "myStrongPassword",
	}
	err, aboveOne = validateInput(payload)
	if err == nil {
		t.Fatal("Could not invalidate the two field errors")
	}
	if !aboveOne {
		t.Fatal("Expected to record two field errors")
	}

	payload = RegisterUser{
		Password:        "myStrongPassword",
		ConfirmPassword: "myStrongPassword",
	}
	err, aboveOne = validateInput(payload)
	if err == nil {
		t.Fatal("Could not invalidate the email required field error")
	}
	if err.Error() != "email is required" {
		t.Fatal("Inconsistent error message")
	}
	if aboveOne {
		t.Fatal("Expected to record one field error")
	}

	payload = RegisterUser{
		Email:           "alozyuche@gmail.com",
		Password:        "myStrongPasswords",
		ConfirmPassword: "myStrongPassword",
	}
	err, aboveOne = validateInput(payload)
	if err == nil {
		t.Fatal("Could not invalidate the two field errors")
	}
	if err.Error() != "confirm_password should be the same as Password" {
		t.Fatal("Inconsistent error message, ", err)
	}
	if aboveOne {
		t.Fatal("Expected to record one field error")
	}

	payload = RegisterUser{
		Email:           "alozyuche@gmail.com",
		Password:        "myStrongPassword",
		ConfirmPassword: "myStrongPassword",
	}
	err, aboveOne = validateInput(payload)
	if err != nil {
		t.Fatal("Catching non existent errors")
	}
	if aboveOne {
		t.Fatal("Expected to record zero field error")
	}
}
