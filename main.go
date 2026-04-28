package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"eats365cli/api"
	"eats365cli/login"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

const maxRetryAttempts = 5

func main() {
	r := mux.NewRouter()

	// Endpoints
	r.HandleFunc("/merchant-list", handleMerchantList).Methods("POST")
	r.HandleFunc("/register-device", handleRegisterDevice).Methods("POST")

	// Health check endpoint
	r.HandleFunc("/health", healthCheck).Methods("GET")

	// Enable CORS
	corsOptions := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	})

	handler := corsOptions.Handler(r)

	// Start the server
	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// healthCheck handles the health check requests
func healthCheck(w http.ResponseWriter, r *http.Request) {
	// You can add more checks here if needed (e.g., database connection, external API health, etc.)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

// Merchant represents a simplified version of a merchant
type Merchant struct {
	MerchantUID int    `json:"merchantUID"`
	Username    string `json:"username"`
	NickName    string `json:"nickName"`
}

// convertToSimplifiedMerchantList converts the API response to a simplified merchant list
func convertToSimplifiedMerchantList(response api.ObtainMerchantListResponse) []Merchant {
	var simplifiedMerchantList []Merchant
	for _, merchant := range response.MerchantList {
		simplifiedMerchantList = append(simplifiedMerchantList, Merchant{
			MerchantUID: merchant.MerchantUID,
			Username:    merchant.Username,
			NickName:    merchant.NickName,
		})
	}
	return simplifiedMerchantList
}

// handleMerchantList handles the API request to obtain the merchant list
func handleMerchantList(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		DeviceKeyChain string `json:"deviceKeyChain"`
	}

	// Parse JSON request body
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure deviceKeyChain is provided
	if reqBody.DeviceKeyChain == "" {
		log.Println("Missing deviceKeyChain in request")
		http.Error(w, "Missing deviceKeyChain", http.StatusBadRequest)
		return
	}

	// Fetch merchant list using provided deviceKeyChain
	log.Printf("Fetching merchant list for deviceKeyChain: %s", reqBody.DeviceKeyChain)
	response, err := fetchMerchantList(reqBody.DeviceKeyChain)
	if err != nil {
		log.Printf("Error fetching merchant list: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(convertToSimplifiedMerchantList(response)); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// handleRegisterDevice handles device registration
func handleRegisterDevice(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		DeviceKeyChain string `json:"deviceKeyChain"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if reqBody.DeviceKeyChain == "" {
		log.Println("Missing deviceKeyChain in request")
		http.Error(w, "Missing deviceKeyChain", http.StatusBadRequest)
		return
	}

	uuid, err := generateUUID()
	if err != nil {
		log.Printf("Error generating UUID: %v", err)
		http.Error(w, "Failed to generate UUID", http.StatusInternalServerError)
		return
	}

	log.Printf("Registering device with DeviceKeyChain: %s and UUID: %s", reqBody.DeviceKeyChain, uuid)
	err = api.RegisterDevice(reqBody.DeviceKeyChain, uuid)
	if err != nil {
		log.Printf("Error registering device: %v", err)
		http.Error(w, "Failed to register device", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Device registered successfully", "uuid": uuid})
}

// fetchMerchantList fetches the merchant list based on the deviceKeyChain
func fetchMerchantList(deviceKeyChain string) (api.ObtainMerchantListResponse, error) {

	log.Printf("Received deviceKeyChain : %s ", deviceKeyChain)

	// deviceKeyChain = "e4cc4904ba764f56"
	var accessToken, baseURL string
	uuid, err := generateUUID()
	if err != nil {
		log.Printf("Error generating UUID: %v", err)
		return api.ObtainMerchantListResponse{}, err
	}
	// api.RegisterDevice(deviceKeyChain, uuid)
	// Retry logic, max 5 attempts
	for attempts := 0; attempts < maxRetryAttempts; attempts++ {
		log.Printf("Attempt %d to login with deviceKeyChain: %s", attempts+1, deviceKeyChain)
		res, err := login.CheckLoginRequest(deviceKeyChain, uuid)
		if err != nil {
			if attempts == maxRetryAttempts-1 {
				log.Printf("Login failed after %d attempts: %v", maxRetryAttempts, err)
				return api.ObtainMerchantListResponse{}, fmt.Errorf("failed to login after %d attempts: %w", maxRetryAttempts, err)
			}
			log.Printf("Login attempt %d failed: %v. Retrying...", attempts+1, err)
			time.Sleep(5 * time.Second)
			continue
		}

		accessToken = res.AccessToken
		baseURL = res.MPosURL
		log.Printf("Login successful, accessToken: %s, baseURL: %s", accessToken, baseURL)
		break
	}

	merchantListRes, err := api.ObtainMerchantList(deviceKeyChain, uuid, accessToken, baseURL)
	if err != nil {
		log.Printf("Error obtaining merchant list: %v", err)
		return api.ObtainMerchantListResponse{}, fmt.Errorf("failed to obtain merchant list: %w", err)
	}

	log.Printf("Successfully obtained merchant list with %d merchants", len(merchantListRes.MerchantList))
	return *merchantListRes, nil
}

// generateUUID generates a version 4 UUID
func generateUUID() (string, error) {
	uuid := make([]byte, 16)
	if _, err := rand.Read(uuid); err != nil {
		log.Printf("Error generating UUID: %v", err)
		return "", err
	}

	uuid[6] = (uuid[6] & 0x0F) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3F) | 0x80 // Variant is 10

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
			uuid[0:4],    // Time-low
			uuid[4:6],    // Time-mid
			uuid[6:8],    // Time-high and version
			uuid[8:10],   // Clock-seq-and-reserved
			uuid[10:16]), // Node
		nil
}
