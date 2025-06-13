package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthIdentity struct {
	Subject string
	Issuer  string
}

const DiscordOAuthBackend = "https://discord.com/api/users/@me"

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrTokenExpired = errors.New("token expired")
	ErrReplay       = errors.New("token replay")
	ErrMissmatch    = errors.New("token issuer missmatch")
	ErrInvalidToken = errors.New("invalid token")
	ErrInvalidSub   = errors.New("invalid subject line")
	ErrIDNotFound   = errors.New("id is not found")
	ErrMissingOAuth = errors.New("no OAuth Token was provided")
	ErrMalformed    = errors.New("malformed token")
)

var (
	usedJTI   = make(map[string]int64)
	jtiMutex  sync.Mutex
	jtiWindow = int64(300) // seconds, match token lifetime
)

func loadBotPublicKey() []byte {
	key, err := os.ReadFile("bot-public.pem")
	if err != nil {
		log.Fatal(err)
	}
	return key
}

func loadSingleEndPointPublicKey() []byte {
	key, err := os.ReadFile("single-endpoint.pem")
	if err != nil {
		log.Fatal(err)
	}
	return key
}

func CleanupJTI() {
	for {
		time.Sleep(time.Minute)
		now := time.Now().Unix()
		jtiMutex.Lock()
		for jti, exp := range usedJTI {
			if exp < now {
				delete(usedJTI, jti)
			}
		}
		jtiMutex.Unlock()
	}
}

var issuerPublicKeys = map[string][]byte{
	"gov-bot":         loadBotPublicKey(),            // []byte loaded from file
	"single-endpoint": loadSingleEndPointPublicKey(), // []byte loaded from file
}

func verifyTokenMultiIssuer(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, ErrMalformed
		}
		issuer, ok := claims["iss"].(string)
		if !ok {
			fmt.Println("Missing iss claim")
			return nil, ErrMalformed
		}
		pubKeyPEM, ok := issuerPublicKeys[issuer]
		if !ok {
			return nil, fmt.Errorf("unknown issuer: %s", issuer)
		}
		return jwt.ParseRSAPublicKeyFromPEM(pubKeyPEM)
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func Authenticate(r *http.Request) (*AuthIdentity, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		fmt.Println("Missing Bearer Token")
		return nil, ErrUnauthorized
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	claims, err := verifyTokenMultiIssuer(tokenString)
	if err != nil {
		fmt.Println("MultiIssuer verification failed, %w", err)
		return nil, ErrUnauthorized
	}

	if exp, ok := claims["exp"].(float64); ok {
		if int64(exp) < time.Now().Unix() {
			return nil, ErrTokenExpired
		}
	}

	jti, ok := claims["jti"].(string)
	if !ok || jti == "" {
		return nil, ErrMalformed
	}
	exp := int64(claims["exp"].(float64))

	jtiMutex.Lock()
	if usedExp, exists := usedJTI[jti]; exists && usedExp >= time.Now().Unix() {
		jtiMutex.Unlock()
		return nil, ErrReplay
	}
	usedJTI[jti] = exp
	jtiMutex.Unlock()

	sub, ok := claims["sub"].(string)
	if !ok {
		return nil, ErrInvalidSub
	}
	if claims["iss"] == "gov-bot" {
		return &AuthIdentity{
			Subject: sub,
			Issuer:  "gov-bot",
		}, nil
	}
	err = doOauth(claims, sub)
	if err != nil {
		return nil, err
	}

	return &AuthIdentity{Subject: sub, Issuer: "Foreign"}, nil
}

func doOauth(claims jwt.MapClaims, subject string) error {
	oauthToken, ok := claims["oauth_token"].(string)
	if !ok || oauthToken == "" {
		return ErrMissingOAuth
	}
	return verifyDiscordToken(oauthToken, subject)
}

func verifyDiscordToken(token string, subject string) error {
	req, err := http.NewRequest("GET", DiscordOAuthBackend, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ErrUnauthorized
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ErrUnauthorized
	}

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return err
	}

	discordID, ok := userInfo["id"].(string)
	if !ok {
		return ErrIDNotFound
	}

	if discordID != subject {
		return ErrMissmatch
	}

	return nil
}
