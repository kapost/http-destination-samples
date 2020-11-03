package main

import (
	"log"
	"os"
	"strings"
	"fmt"
	"regexp"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"encoding/hex"
	"crypto/hmac"
	"crypto/sha256"
)

type Capabilities struct {
	Html bool `json:"html"`
	AnyFile bool `json:"any_file"`
}

type Metadata struct {
	ExternalId string `json:"external_id"`
	PublishedUrl string `json:"published_url"`
}

type AuthResponse struct {
	Capabilities Capabilities `json:"capabilities"`
}

type PublishResponse struct {
	Metadata Metadata `json:"metadata"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type PayloadAction struct {
	ExternalId string `json:"external_id"`
}

type Payload struct {
	Action PayloadAction `json:"action"`
}

func VerifySignature(header http.Header, body []byte) bool {
	if header["X-Kapost-Platform"] == nil || header["X-Kapost-Action"] == nil || header["X-Kapost-Signature"] == nil {
		return false
	}

	platform := header["X-Kapost-Platform"][0]
	action := header["X-Kapost-Action"][0]
	signature := header["X-Kapost-Signature"][0]
	secret := []byte(os.Getenv("SIGNATURE_SECRET"))

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(platform + action + string(body)))

	s := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(strings.Join([]string { "sha256", s }, "=")), []byte(signature))
}

func VerifyAction(header http.Header) bool {
	if header["X-Kapost-Action"] == nil {
		return false
	}

	action := header["X-Kapost-Action"][0]
	return action == "auth" || action == "publish" || action == "republish"
}

func VerifyApiKey(header http.Header) bool {
	if header["Authorization"] == nil {
		return false
	}

	re := regexp.MustCompile(`^Bearer\s+`)
	api_key := re.ReplaceAll([]byte(header["Authorization"][0]), []byte(""))

	return hmac.Equal(api_key, []byte(os.Getenv("API_KEY")))
}

func RequestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(&ErrorResponse { "Unsupported HTTP method" })
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&ErrorResponse { "Unexpected error while reading payload" })
		return
	}

	if !VerifySignature(r.Header, body) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(&ErrorResponse { "Invalid signature" })
		return
	}

	if !VerifyApiKey(r.Header) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(&ErrorResponse { "Invalid API Key" })
		return
	}

	if !VerifyAction(r.Header) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(&ErrorResponse { "Action is not supported" })
		return
	}

	action := r.Header["X-Kapost-Action"][0]

	if action == "auth" {
		// NOTE: the api_key could also be validated against some external service on
		// authentication in the app center and an "error" could be returned if it
		// was incorrect, etc.
		resp := &AuthResponse {
			Capabilities: Capabilities{
				Html: true,
				AnyFile: true,
			},
		}
		json.NewEncoder(w).Encode(resp)
	} else if action == "publish" {
		resp := &PublishResponse {
			Metadata {
				ExternalId: "abc33",
				PublishedUrl: "https://localhost/go-test",
			},
		}
		json.NewEncoder(w).Encode(resp)
	} else if action == "republish" {
		payload := Payload {}
		json.Unmarshal(body, &payload)

		if payload.Action.ExternalId != "abc33" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(&ErrorResponse { "Cannot republish because external id not found" })
		} else {
			resp := &PublishResponse {
				Metadata {
					ExternalId: "abc33",
					PublishedUrl: "https://localhost/go-test-republish",
				},
			}
			json.NewEncoder(w).Encode(resp)
		}
	}
}

func main() {
	http.HandleFunc("/", RequestHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), nil))
}
