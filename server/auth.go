package server

import (
	"errors"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type RegisterUser struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required"`
	ConfirmPassword string `json:"confirm_password" validate:"eqfield=Password"`
	DeviceID        string `json:"device_id,omitempty"`
}

type User struct {
	ID             uint      `json:"id"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"password"`
	FirstName      string    `json:"first_name"`
	PhoneNumber    string    `json:"phone_number"`
	UserAddress    string    `json:"user_address"`
	IsActive       bool      `json:"is_active"`
	DateJoined     time.Time `json:"date_joined"`
	LastLogin      time.Time `json:"last_login"`
	Longitude      string    `json:"longitude"`
	Latitude       string    `json:"latitude"`
	DeviceID       string    `json:"device_id"`
}

var (
	signingKey        = []byte(os.Getenv("SIGNING_KEY"))
	refreshSigningKey = []byte(os.Getenv("REFRESH_SIGNING_KEY"))
)

// Hashes a password
func HashPassword(password string) (string, error) {
	if len(password) < 1 {
		return "", errors.New("Cant hash an empty string")
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 4)
	return string(bytes), err
}

// Generates an acess and refresh token on authentication
func GenerateToken(email string) (string, string, error) {

	if len(email) == 0 {
		return "", "", errors.New("Can't generate token for an invalid email")
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["client"] = email
	claims["exp"] = time.Now().Add(time.Minute * 15).Unix()

	accessToken, err := token.SignedString(signingKey)
	if err != nil {
		return "", "", err
	}

	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshClaims := refreshToken.Claims.(jwt.MapClaims)

	refreshClaims["authorized"] = true
	refreshClaims["client"] = email
	refreshClaims["exp"] = time.Now().Add(time.Hour * 8).Unix()

	refreshString, err := refreshToken.SignedString(refreshSigningKey)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshString, nil
}
