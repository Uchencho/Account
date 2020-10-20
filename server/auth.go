package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type RegisterUser struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required"`
	ConfirmPassword string `json:"confirm_password" validate:"eqfield=Password"`
	DeviceID        string `json:"device_id,omitempty"`
	FirstName       string `json:"first_name"`
}

type User struct {
	Email          string    `json:"email"`
	HashedPassword string    `json:"password,omitempty"`
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

const (
	frontEndOrigin string = "*"
)

var (
	signingKey        = []byte(os.Getenv("SIGNING_KEY"))
	refreshSigningKey = []byte(os.Getenv("REFRESH_SIGNING_KEY"))
)

type Key string

type tokenDetails struct {
	RefreshToken string `json:"refresh_token,omitempty" validate:"required"`
	AccessToken  string `json:"access_token,omitempty"`
}

// Hashes a password
func HashPassword(password string) (string, error) {
	if len(password) < 1 {
		return "", errors.New("Cant hash an empty string")
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 4)
	return string(bytes), err
}

// Checks the password and the hash, returns a non nil error if not the same
func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
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
	claims["exp"] = time.Now().Add(time.Hour * 15).Unix()

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

// Middleware that checks if a token was passed
func BasicToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Header["Authorization"] != nil {
			if len(strings.Split(r.Header["Authorization"][0], " ")) < 2 {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprint(w, `{"error" : "Invalid token format"}`)
				return
			}

			accessToken := strings.Split(r.Header["Authorization"][0], " ")[1]
			basic_token := os.Getenv("BASIC_TOKEN")
			if basic_token == accessToken {

				//Allow CORS here
				w.Header().Set("Access-Control-Allow-Origin", frontEndOrigin)
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				next.ServeHTTP(w, r)
				return
			} else {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprint(w, `{"error" : "Invalid token passed"}`)
				return
			}
		}
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"error" : "Token not passed"}`)
	})
}

// Checks if the accesstoken passed is correct
func checkAccessToken(accessToken string) (interface{}, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("An error occurred")
		}
		return signingKey, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["client"], nil
	}
	return "", errors.New("Credentials not provided")
}

// Checks if the refresh token passed is correct
func checkRefreshToken(refreshToken string) (interface{}, error) {
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("An error occurred")
		}
		return refreshSigningKey, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["client"], nil
	}
	return "", errors.New("Credentials not provided")
}

// Creates a new access token only
func newAccessToken(email string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["client"] = email
	claims["exp"] = time.Now().Add(time.Hour * 2).Unix()

	accessToken, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}
	return accessToken, nil
}

// Middleware that returns the details of the user
func TheUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Header["Authorization"] != nil {
			if len(strings.Split(r.Header["Authorization"][0], " ")) < 2 {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprint(w, `{"error" : "Invalid token format"}`)
				return
			}

			accessToken := strings.Split(r.Header["Authorization"][0], " ")[1]
			email, err := checkAccessToken(accessToken)
			if err != nil && "Token is expired" == err.Error() {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, `{"error" : "Token has expired, please login"}`)
				return
			} else if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, `{"error" : "Invalid Token"}`)
				return
			}

			user, err := getUser(Client, fmt.Sprint(email))
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, `{"error" : "User does not exist"}`)
				return
			}

			const userKey Key = "user"
			ctx := context.WithValue(r.Context(), userKey, user)

			//Allow CORS here
			w.Header().Set("Access-Control-Allow-Origin", frontEndOrigin)
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"error" : "Token not passed"}`)
	})
}
