# Receipt Processor API

This is a web service for processing receipts and calculating points based on defined rules. Users can submit receipts to the service, which will calculate and return a unique ID. The ID can then be used to retrieve the points awarded for the receipt based on criteria such as retailer name, purchase total, and item details.

## Endpoints

- **POST /receipts/process**: Submit a receipt for processing and receive a unique ID.
- **GET /receipts/{id}/points**: Retrieve the points awarded for a processed receipt by its unique ID.

## Prerequisites

- [Go](https://golang.org/dl/) must be installed on your machine.  
  Recommended version: `1.19` or later.

## Running Locally

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd receipt-processor
   ```
2. Run:
   ```bash
   go run main.go
   ```
   - The server will start on http://localhost:8080.

3.	Access the API:
	-	Use tools like Postman or curl to test the endpoints.
	-	Example requests:
        - POST :
        ```bash
        curl -X POST -H "Content-Type: application/json" -d '{
        "retailer": "Target",
        "purchaseDate": "2022-01-01",
        "purchaseTime": "13:01",
        "items": [{"shortDescription": "Mountain Dew 12PK", "price": "6.49"}],
        "total": "35.35"
        }' http://localhost:8080/receipts/process
        ``` 

        - GET :
        ```bash
        curl -X GET http://localhost:8080/receipts/<id>/points
        ```
        Note: Replace <id> with the unique receipt ID returned from the POST request.


## Running Tests

To verify the implementation and test against edge cases, run:
    ```bash
    go test ./...
    ```

## What the Tests Cover

	- Validation: Ensures all required fields and formats are correctly handled.
	- Edge Cases: Tests behavior for invalid inputs, missing fields, and other edge scenarios.
	- Point Calculations: Validates correct point calculation based on the defined rules.

    Sample output:
    ```plaintext
    ok  	receipt-processor   0.123s
    ```
