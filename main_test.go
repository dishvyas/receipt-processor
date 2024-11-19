package main

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func createMockRequest(method, url string, payload interface{}) (*http.Request, *httptest.ResponseRecorder) {
    var body []byte
    if payload != nil {
        body, _ = json.Marshal(payload)
    }
    req := httptest.NewRequest(method, url, bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()
    return req, rec
}

func TestProcessReceiptHandler(t *testing.T) {
    tests := []struct {
        name           string
        receipt        Receipt
        expectedStatus int
        shouldHaveID   bool
    }{
        {
            name: "Valid receipt",
            receipt: Receipt{
                Retailer:     "Target",
                PurchaseDate: "2022-01-01",
                PurchaseTime: "13:01",
                Items: []Item{
                    {ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
                },
                Total: "35.35",
            },
            expectedStatus: http.StatusOK,
            shouldHaveID:   true,
        },
        {
            name: "Missing retailer",
            receipt: Receipt{
                PurchaseDate: "2022-01-01",
                PurchaseTime: "13:01",
                Items: []Item{
                    {ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
                },
                Total: "35.35",
            },
            expectedStatus: http.StatusBadRequest,
            shouldHaveID:   false,
        },
        {
            name: "Empty items list",
            receipt: Receipt{
                Retailer:     "Target",
                PurchaseDate: "2022-01-01",
                PurchaseTime: "13:01",
                Items:        []Item{},
                Total:        "35.35",
            },
            expectedStatus: http.StatusBadRequest,
            shouldHaveID:   false,
        },
        {
            name: "Invalid purchase date format",
            receipt: Receipt{
                Retailer:     "Target",
                PurchaseDate: "01-01-2022",
                PurchaseTime: "13:01",
                Items: []Item{
                    {ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
                },
                Total: "35.35",
            },
            expectedStatus: http.StatusBadRequest,
            shouldHaveID:   false,
        },
    }

    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            req, rec := createMockRequest(http.MethodPost, "/receipts/process", test.receipt)
            processReceiptHandler(rec, req)

            if rec.Code != test.expectedStatus {
                t.Errorf("[%s] Expected status code %d, got %d", test.name, test.expectedStatus, rec.Code)
            }

            if test.shouldHaveID {
                var resp map[string]string
                if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
                    t.Fatalf("[%s] Failed to parse response JSON: %v", test.name, err)
                }
                if _, exists := resp["id"]; !exists {
                    t.Errorf("[%s] Response does not contain 'id'", test.name)
                }
            }
        })
    }
}

func TestGetPointsHandler(t *testing.T) {
    // Mock data
    mockID := "test-id"
    receipts[mockID] = 42

    tests := []struct {
        name           string
        url            string
        expectedStatus int
        expectedPoints int
    }{
        {
            name:           "Valid receipt ID",
            url:            "/receipts/test-id/points",
            expectedStatus: http.StatusOK,
            expectedPoints: 42,
        },
        {
            name:           "Invalid receipt ID",
            url:            "/receipts/invalid-id/points",
            expectedStatus: http.StatusNotFound,
            expectedPoints: 0,
        },
    }

    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            req, rec := createMockRequest(http.MethodGet, test.url, nil)
            handleReceipts(rec, req)

            if rec.Code != test.expectedStatus {
                t.Errorf("[%s] Expected status code %d, got %d", test.name, test.expectedStatus, rec.Code)
            }

            if test.expectedStatus == http.StatusOK {
                var resp map[string]int
                if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
                    t.Fatalf("[%s] Failed to parse response JSON: %v", test.name, err)
                }
                if resp["points"] != test.expectedPoints {
                    t.Errorf("[%s] Expected points %d, got %d", test.name, test.expectedPoints, resp["points"])
                }
            }
        })
    }
}

func TestCalculatePoints(t *testing.T) {
    tests := []struct {
        name           string
        receipt        Receipt
        expectedPoints int
    }{
        {
            name: "Simple receipt",
            receipt: Receipt{
                Retailer:     "Target",
                PurchaseDate: "2022-01-01",
                PurchaseTime: "13:01",
                Items: []Item{
                    {ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
                },
                Total: "35.35",
            },
            expectedPoints: 12, // 6 (Retailer) + 6 (Odd day)
        },
        {
            name: "Round total and multiple of 0.25",
            receipt: Receipt{
                Retailer:     "Walmart",
                PurchaseDate: "2023-12-25",
                PurchaseTime: "15:30",
                Items: []Item{
                    {ShortDescription: "Pepsi", Price: "2.00"},
                },
                Total: "2.00",
            },
            expectedPoints: 98, // 7 (Retailer) + 6 (Odd day) + 10 (Time) + 50 (Round total) + 25 (Multiple of 0.25)
        },
    }

    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            points := calculatePoints(test.receipt)
            if points != test.expectedPoints {
                t.Errorf("[%s] Expected points %d, got %d", test.name, test.expectedPoints, points)
            }
        })
    }
}