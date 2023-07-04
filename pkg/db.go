package pkg

import (
	"database/sql"
	"fmt"
	"log"

	pb "github.com/berdebotond/bankable_technical_test/protos"
	"github.com/brianvoe/gofakeit"
	_ "github.com/lib/pq"
)

// SetupDatabase sets up the database connection and returns the db object
func SetupDatabase(host string, port string, user string, password string, dbname string) *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to database")
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// Create the messages table if it doesn't already exist
	_, err = db.Exec(`
	CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

	CREATE TABLE IF NOT EXISTS invoice (
		id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
		issuer_id UUID,
		status VARCHAR(255),
		investor_id UUID
	);
	
	CREATE TABLE IF NOT EXISTS issuer (
		id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
		balance FLOAT NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS investor (
		id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
		balance FLOAT NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS bid (
		id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
		investor_id UUID,
		invoice_id UUID,
		amount FLOAT NOT NULL
		);

	DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_invoice_issuer') THEN
			ALTER TABLE invoice ADD CONSTRAINT fk_invoice_issuer FOREIGN KEY (issuer_id) REFERENCES issuer(id);
		END IF;
	
		IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_invoice_investor') THEN
			ALTER TABLE invoice ADD CONSTRAINT fk_invoice_investor FOREIGN KEY (investor_id) REFERENCES investor(id);
		END IF;
	
		IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_bid_investor') THEN
			ALTER TABLE bid ADD CONSTRAINT fk_bid_investor FOREIGN KEY (investor_id) REFERENCES investor(id);
		END IF;
	
		IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_bid_invoice') THEN
			ALTER TABLE bid ADD CONSTRAINT fk_bid_invoice FOREIGN KEY (invoice_id) REFERENCES invoice(id);
		END IF;
	END $$;
	`)

	log.Println("Created messages table")
	if err != nil {
		log.Fatal(err)
	}
	err = InitializeMockData(db)
	if err != nil {
		log.Fatal(err)
	}
	// log all inestor and invocer
	investors, err := GetAllInvestors(db)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("All investors: %v", investors)
	issuers, err := GetAllIssuers(db)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("All issuers: %v", issuers)
	return db
}

// Get all investors from the database only use locally
func GetAllInvestors(db *sql.DB) ([]*pb.Investor, error) {
	rows, err := db.Query("SELECT * FROM investor")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	investors := make([]*pb.Investor, 0)
	for rows.Next() {
		var id string
		var balance float32
		if err := rows.Scan(&id, &balance); err != nil {
			return nil, err
		}
		investors = append(investors, &pb.Investor{Id: id, Balance: balance})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return investors, nil
}

// Get all issuers from the database only use locally
func GetAllIssuers(db *sql.DB) ([]*pb.Issuer, error) {
	rows, err := db.Query("SELECT * FROM issuer")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	issuers := make([]*pb.Issuer, 0)
	for rows.Next() {
		var id string
		var balance float32
		if err := rows.Scan(&id, &balance); err != nil {
			return nil, err
		}
		issuers = append(issuers, &pb.Issuer{Id: id, Balance: balance})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return issuers, nil
}

// InitializeMockData inserts mock data into the database
func InitializeMockData(db *sql.DB) error {
	// Seed the random number generator
	gofakeit.Seed(0)

	for i := 0; i < 15; i++ {
		issuerBalance := gofakeit.Float64Range(1000.0, 5000.0)
		_, err := db.Exec("INSERT INTO issuer (balance) VALUES ($1)", issuerBalance)
		if err != nil {
			return err
		}

		investorBalance := gofakeit.Float64Range(5000.0, 10000.0)
		_, err = db.Exec("INSERT INTO investor (balance) VALUES ($1)", investorBalance)
		if err != nil {
			return err
		}
	}

	return nil
}
