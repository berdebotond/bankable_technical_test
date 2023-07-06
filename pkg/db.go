package pkg

import (
	"context"
	"database/sql"
	"errors"
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
		investor_id UUID,
		price FLOAT
	);
	
	CREATE TABLE IF NOT EXISTS issuer (
		id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
		balance FLOAT NOT NULL,
		name VARCHAR(255)
	);
	
	CREATE TABLE IF NOT EXISTS investor (
		id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
		balance FLOAT NOT NULL,
		name VARCHAR(255)
	);
	
	CREATE TABLE IF NOT EXISTS bid (
		id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
		investor_id UUID,
		invoice_id UUID,
		amount FLOAT NOT NULL,
		status VARCHAR(255)
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
	rows, err := db.Query("SELECT balance, name  FROM investor")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	investors := make([]*pb.Investor, 0)
	for rows.Next() {
		var balance float32
		var name string
		if err := rows.Scan(&balance, &name); err != nil {
			return nil, err
		}
		investors = append(investors, &pb.Investor{Name: name, Balance: balance})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return investors, nil
}

// Get all issuers from the database only use locally
func GetAllIssuers(db *sql.DB) ([]*pb.Issuer, error) {
	rows, err := db.Query("SELECT balance, name FROM issuer")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	issuers := make([]*pb.Issuer, 0)
	for rows.Next() {
		var balance float32
		var name string
		if err := rows.Scan(&balance, &name); err != nil {
			return nil, err
		}
		issuers = append(issuers, &pb.Issuer{Balance: balance, Name: name})
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
	log.Println("Seeded random number generator")
	for i := 0; i < 15; i++ {
		issuerBalance := gofakeit.Float64Range(1000.0, 5000.0)
		issuerName := gofakeit.Name()
		_, err := db.Exec("INSERT INTO issuer (balance, name) VALUES ($1, $2)", issuerBalance, issuerName)
		//_, err := db.Exec("INSERT INTO issuer (balance, name) VALUES ($1, $2)", issuerBalance, issuerName)
		if err != nil {
			return err
		}

		investorBalance := gofakeit.Float64Range(5000.0, 10000.0)
		investorName := gofakeit.Name()
		_, err = db.Exec("INSERT INTO investor (balance, name) VALUES ($1, $2)", investorBalance, investorName)
		if err != nil {
			return err
		}
	}
	log.Println("Inserted mock data")
	return nil
}

func CheckInvestorBalance(ctx context.Context, db *sql.DB, in *pb.Bid) error {
	log.Println("Checking investor's balance")
	var balance float32
	err := db.QueryRowContext(ctx, "SELECT balance FROM investor WHERE id = $1", in.InvestorId).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("investor not found: %w", err)
		}
		return fmt.Errorf("failed to get investor's balance: %w", err)
	}
	if balance < in.Amount {
		return errors.New("investor doesn't have enough balance")
	}
	//commit changes
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	return nil
}

func RededuceInvestorBalance(ctx context.Context, db *sql.DB, in *pb.Bid) error {
	log.Println("Reducing investor's balance")
	_, err := db.ExecContext(ctx, "UPDATE investor SET balance = balance - $1 WHERE id = $2", in.Amount, in.InvestorId)
	if err != nil {
		return fmt.Errorf("failed to reduce investor's balance: %w", err)
	}
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}
	return nil
}

func CloseBids(ctx context.Context, db *sql.DB, in *pb.Bid) error {
	log.Println("Closing bid")

	_, err := db.ExecContext(ctx, "UPDATE bid SET status = 'closed' WHERE invoice_id = $1", in.InvoiceId)
	if err != nil {
		return fmt.Errorf("failed to close bid: %w", err)
	}
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	err = IncreasePreviousInvestorsBalance(ctx, db, in)
	if err != nil {
		return fmt.Errorf("failed to increase previous investors' balance: %w", err)
	}

	return nil
}

func IncreasePreviousInvestorsBalance(ctx context.Context, db *sql.DB, in *pb.Bid) error {
	log.Printf("Increasing previous investors' balance with id: %v", in.InvoiceId)
	_, err := db.ExecContext(ctx, "UPDATE investor SET balance = balance + bid.amount FROM bid WHERE bid.invoice_id = $1 AND investor.id = bid.investor_id", in.InvoiceId)
	if err != nil {
		return fmt.Errorf("failed to increase previous investors' balance: %w", err)
	}
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}
	return nil
}

func UpdateInvestorInInvoice(ctx context.Context, db *sql.DB, in *pb.Bid) error {
	log.Println("Updating investor in invoice")
	_, err := db.ExecContext(ctx, "UPDATE invoice SET investor_id = $1 WHERE id = $2", in.InvestorId, in.InvoiceId)
	if err != nil {
		return fmt.Errorf("failed to update investor in invoice: %w", err)
	}
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}
	return nil
}

func DetermineBidStatus(ctx context.Context, db *sql.DB, in *pb.Bid) (string, error) {
	log.Println("Determining bid status")
	var price float32
	err := db.QueryRowContext(ctx, "SELECT price FROM invoice WHERE id = $1", in.InvoiceId).Scan(&price)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("invoice not found: %w", err)
		}
		return "", fmt.Errorf("failed to get invoice price: %w", err)
	}
	if in.Amount == price {

		return "approved", nil
	}
	return "pending", nil
}

func InsertBid(ctx context.Context, db *sql.DB, in *pb.Bid, status string) error {
	log.Println("Inserting bid")
	_, err := db.ExecContext(ctx, "INSERT INTO bid (investor_id, invoice_id, amount, status) VALUES ($1, $2, $3, 'pending')", in.InvestorId, in.InvoiceId, in.Amount)
	if err != nil {
		return fmt.Errorf("failed to insert bid: %w", err)
	}
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}
	return nil
}

func CloseInvoice(ctx context.Context, db *sql.DB, in *pb.Bid) error {
	log.Printf("Approving invoice with invoice id %s", in.GetInvoiceId())
	_, err := db.ExecContext(ctx, "UPDATE invoice SET status = 'closed', investor_id = $1 WHERE id = $2", in.GetInvestorId(), in.GetInvoiceId())

	if err != nil {
		log.Printf("Error updating invoice: %v", err)
		return err
	}

	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}
	return nil
}

func UpadeIssuerBalanceByBid(ctx context.Context, db *sql.DB, in *pb.Bid) error {
	log.Printf("Updating issuer's balance ")
	_, err := db.ExecContext(ctx, "UPDATE issuer SET balance = balance + $1 WHERE id = (SELECT issuer_id FROM invoice WHERE id = $2)", in.GetAmount(), in.GetInvoiceId())
	if err != nil {
		return fmt.Errorf("failed to update issuer's balance: %w", err)
	}
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}
	return nil
}

func ListAllBids(ctx context.Context, db *sql.DB) ([]*pb.Bid, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, investor_id, invoice_id, amount, status FROM bid")
	if err != nil {
		return nil, fmt.Errorf("failed to query bids: %w", err)
	}
	defer rows.Close()

	var bids []*pb.Bid
	for rows.Next() {
		var bid pb.Bid
		if err := rows.Scan(&bid.Id, &bid.InvestorId, &bid.InvoiceId, &bid.Amount, &bid.Status); err != nil {
			return nil, fmt.Errorf("failed to scan bid: %w", err)
		}
		bids = append(bids, &bid)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read bids: %w", err)
	}
	return bids, nil
}
