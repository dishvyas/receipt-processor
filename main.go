package main

import (
    "crypto/rand"
    "encoding/hex"
	"encoding/json"
    "log"
    "net/http"
	"math"
    "strconv"
    "strings"
)

// Generate a unique ID
func generateID() string {
    bytes := make([]byte, 16)
    if _, err := rand.Read(bytes); err != nil {
        log.Fatalf("Failed to generate unique ID: %v", err)
    }
    return hex.EncodeToString(bytes)
}

type Item struct {
    ShortDescription string `json:"shortDescription"`
    Price            string `json:"price"`
}

type Receipt struct {
    Retailer     string  `json:"retailer"`
    PurchaseDate string  `json:"purchaseDate"`
    PurchaseTime string  `json:"purchaseTime"`
    Items        []Item  `json:"items"`
    Total        string  `json:"total"`
}

var receipts = make(map[string]int) // map for ID -> Points

func main() {
    http.HandleFunc("/receipts/process", processReceiptHandler)
    http.HandleFunc("/receipts/", handleReceipts) 

    log.Println("Starting server on :8080...")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func isValidDate(date string) bool {
	// Check if the date is in the format "YYYY-MM-DD"
	if len(date) != 10 || date[4] != '-' || date[7] != '-' {
        return false
    }
    year, err1 := strconv.Atoi(date[:4])
    month, err2 := strconv.Atoi(date[5:7])
    day, err3 := strconv.Atoi(date[8:10])
    if err1 != nil || err2 != nil || err3 != nil {
        return false
    }
    if year < 1 || month < 1 || month > 12 || day < 1 || day > 31 {
        return false
    }
    return true
}

func isValidTime(time string) bool {
	// Check if the time is in the format "HH:MM"
    if len(time) != 5 || time[2] != ':' {
        return false
    }
    hour, err1 := strconv.Atoi(time[:2])
    minute, err2 := strconv.Atoi(time[3:])
    if err1 != nil || err2 != nil {
        return false
    }
    if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
        return false
    }
    return true
}
// Process receipt
func processReceiptHandler(w http.ResponseWriter, r *http.Request) {
    log.Printf("Received %s request for %s", r.Method, r.URL.Path)

    if r.Method != http.MethodPost {
        http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
        return
    }

    if r.Header.Get("Content-Type") != "application/json" {
        http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
        return
    }

    var receipt Receipt
    if err := json.NewDecoder(r.Body).Decode(&receipt); err != nil {
        http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
        return
    }

    // Validate retailer
    if strings.TrimSpace(receipt.Retailer) == "" {
        log.Printf("Invalid receipt: missing or empty retailer field")
        http.Error(w, "Invalid receipt: retailer field is required", http.StatusBadRequest)
        return
    }

    // Validate purchaseDate
    if !isValidDate(receipt.PurchaseDate) {
        log.Printf("Invalid purchaseDate: %s", receipt.PurchaseDate)
        http.Error(w, "Invalid purchaseDate. Expected format: YYYY-MM-DD", http.StatusBadRequest)
        return
    }

    // Validate purchaseTime
    if !isValidTime(receipt.PurchaseTime) {
        log.Printf("Invalid purchaseTime: %s", receipt.PurchaseTime)
        http.Error(w, "Invalid purchaseTime. Expected format: HH:MM", http.StatusBadRequest)
        return
    }

    // Validate items
    if len(receipt.Items) < 1 {
        log.Printf("Invalid receipt: items list must contain at least one item")
        http.Error(w, "Invalid receipt: items list must contain at least one item", http.StatusBadRequest)
        return
    }

    for i, item := range receipt.Items {
        if strings.TrimSpace(item.ShortDescription) == "" {
            log.Printf("Invalid item[%d]: missing shortDescription", i)
            http.Error(w, "Invalid receipt: all items must have a valid shortDescription", http.StatusBadRequest)
            return
        }
        if _, err := strconv.ParseFloat(item.Price, 64); err != nil {
            log.Printf("Invalid item[%d]: invalid price: %s", i, item.Price)
            http.Error(w, "Invalid receipt: all items must have a valid price in the format '0.00'", http.StatusBadRequest)
            return
        }
    }

    // Generate a unique ID for the receipt
    id := generateID()
    log.Printf("Generated unique ID: %s for new receipt", id)

    points := calculatePoints(receipt)
    log.Printf("Successfully processed receipt with ID: %s, Points: %d", id, points)
    receipts[id] = points

    response := map[string]string{"id": id}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func handleReceipts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
        http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
        return
    }
	
	path := strings.TrimPrefix(r.URL.Path, "/receipts/")
    
	if strings.HasSuffix(path, "/points") {
        id := strings.TrimSuffix(path, "/points")
        id = strings.TrimSuffix(id, "/") // Ensures no trailing slashes
		log.Printf("Raw path: %s", r.URL.Path)
		log.Printf("Cleaned path: %s", path)
		log.Printf("Extracted ID: %s", id)
        log.Printf("Received GET request for /receipts/%s", id)
        getPointsHandler(w, r, id)
    } else {
        http.Error(w, "Endpoint not found", http.StatusNotFound)
    }
}

// Get points
func getPointsHandler(w http.ResponseWriter, r *http.Request, id string) {
    log.Printf("Received %s request for %s", r.Method, r.URL.Path)

    if r.Method != http.MethodGet {
        http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
        return
    }

    if id == "" {
        log.Printf("Empty receipt ID in GET request")
        http.Error(w, "Invalid receipt ID", http.StatusBadRequest)
        return
    }

    points, exists := receipts[id]
    if !exists {
        log.Printf("Receipt not found for ID: %s", id)
        http.Error(w, "Receipt not found", http.StatusNotFound)
        return
    }

    log.Printf("Successfully retrieved points for ID: %s, Points: %d", id, points)
    response := map[string]int{"points": points}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// Calculate points for the receipt based on the rules.
func calculatePoints(receipt Receipt) int {
    points := 0

    // Rule 1: One point for every alphanumeric character in the retailer name
    for _, char := range receipt.Retailer {
        if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
            points++
        }
    }

    // Rule 2: 50 points if the total is a round dollar amount with no cents
    if total, err := strconv.ParseFloat(receipt.Total, 64); err == nil && math.Mod(total, 1) == 0 {
        points += 50
    }

    // Rule 3: 25 points if the total is a multiple of 0.25
    if total, err := strconv.ParseFloat(receipt.Total, 64); err == nil && math.Mod(total, 0.25) == 0 {
        points += 25
    }

    // Rule 4: 5 points for every two items on the receipt
    points += (len(receipt.Items) / 2) * 5

    // Rule 5: If the trimmed length of the item description is a multiple of 3,
    // multiply the price by 0.2 and round up to the nearest integer
    for _, item := range receipt.Items {
        descriptionLength := len(strings.TrimSpace(item.ShortDescription))
        if descriptionLength%3 == 0 {
            if price, err := strconv.ParseFloat(item.Price, 64); err == nil {
                points += int(math.Ceil(price * 0.2))
            }
        }
    }

    // Rule 6: 6 points if the day in the purchase date is odd
    if len(receipt.PurchaseDate) >= 10 { // Ensure the date is in the correct format
        day := receipt.PurchaseDate[len(receipt.PurchaseDate)-2:]
        if dayInt, err := strconv.Atoi(day); err == nil && dayInt%2 != 0 {
            points += 6
        }
    }

    // Rule 7: 10 points if the time of purchase is after 2:00 PM and before 4:00 PM
    if len(receipt.PurchaseTime) >= 5 { // Ensure the time is in HH:MM format
        hour, errHour := strconv.Atoi(receipt.PurchaseTime[:2])
        minute, errMinute := strconv.Atoi(receipt.PurchaseTime[3:])
        if errHour == nil && errMinute == nil && hour == 14 && minute >= 0 || (hour == 15 && minute < 60) {
            points += 10
        }
    }

    return points
}