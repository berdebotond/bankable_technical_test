# Bankable Technical Test

This is a Go project that implements a gRPC server for managing bids for invoices with issuers, and investors.

## Project Structure

The project is structured as follows:

```
├── cmd
│   └── server
│       └── main.go
├── config
│   ├── config.go
│   └── config.json
├── Dockerfile
├── go.mod
├── go.sum
├── pkg
│   ├── db.go
│   ├── model.go
│   ├── server.go
│   └── server_test.go
└── protos
    ├── protobuf_grpc.pb.go
    ├── protobuf.pb.go
    └── protobuf.proto
```

## How to Run

The database is initialized with mock data for testing purposes. The mock data includes 15 issuers and 15 investors with random balances and names.

### Docker

1. Ensure you have Go installed on your machine.
2. Clone the repository.
3. Navigate to the root directory of the project.
4. Run `go mod download` to download the necessary dependencies.
5. Run `go run cmd/server/main.go` to start the server.

### Local:

1. Install dependencies: `go mod download`
2. Testing with `go test ./...`
3. Starting the server: `go run cmd/server/main.go`

The server will start and listen on port 50051.

## How to Test

### Unit tests

This project uses Go's built-in testing framework. You can run the tests by navigating to the root directory of the project and running `go test ./...` 

### E2E integartion test

This project has an inplemented end 2 end integartion test, You can run the test by navigating to the root directory of the project and running `go run test/e2e_client_flow.go`
Please ensure you have a PostgreSQL database running and the connection details in `config.json` are correct, as the tests may interact with the database.

## Docker

This project includes a Dockerfile for containerization. To build and run the project in a Docker container, use the following commands:

1. Build the Docker image: `docker build -t bankable_technical_test .`
2. Run the Docker container: `docker run -p 50051:50051 bankable_technical_test`

The server will start and listen on port 50051.


## Endpoint description

1. **PlaceBid**: This endpoint is used to place a bid on an invoice. It first checks if the investor exists and has enough balance. If the investor has enough balance, it reduces the investor's balance, closes previous bids, determines the status of the bid, and inserts the new bid. If the bid status is "approved", it updates the invoice status and investor id.

2. **ApproveTrade**: This endpoint is used to approve a trade and set the invoice status to closed. It updates the invoice status and investor id, closes bids, and updates the issuer balance.

3. **CreateInvoice**: This endpoint is used to create a new invoice with an existing issuer. It inserts a new invoice into the database and returns the created invoice.

4. **GetIssuer**: This endpoint is used to get an issuer by id. It queries the database for the issuer with the given id and returns the issuer.

5. **GetInvestors**: This endpoint is used to get all investors. It queries the database for all investors and returns them in a stream.

6. **GetInvoice**: This endpoint is used to get an invoice by id. It queries the database for the invoice with the given id and returns the invoice.

## Database

The database is a PostgreSQL database, and it is set up with the following tables:

1. **invoice**: This table stores the invoices. Each invoice has an id (UUID), issuer_id (UUID), status (VARCHAR), investor_id (UUID), and price (FLOAT).

2. **issuer**: This table stores the issuers. Each issuer has an id (UUID), balance (FLOAT), and name (VARCHAR).

3. **investor**: This table stores the investors. Each investor has an id (UUID), balance (FLOAT), and name (VARCHAR).

4. **bid**: This table stores the bids. Each bid has an id (UUID), investor_id (UUID), invoice_id (UUID), amount (FLOAT), and status (VARCHAR).

The database also has foreign key constraints to ensure data integrity:

- The invoice table has foreign keys to the issuer and investor tables.
- The bid table has foreign keys to the investor and invoice tables.

The database is initialized with mock data for testing purposes. The mock data includes 15 issuers and 15 investors with random balances and names.

The database also provides several functions for interacting with the data:

- **CheckInvestorBalance**: This function checks if an investor has enough balance to place a bid.
- **RededuceInvestorBalance**: This function reduces an investor's balance when they place a bid.
- **CloseBids**: This function closes a bid.
- **IncreasePreviousInvestorsBalance**: This function increases the balance of previous investors when a bid is closed.
- **UpdateInvestorInInvoice**: This function updates the investor_id in the invoice table when a bid is placed.
- **DetermineBidStatus**: This function determines the status of a bid.
