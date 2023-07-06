package pkg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	pb "github.com/berdebotond/bankable_technical_test/protos"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

// server is used to implement InvoiceServiceServer.
func SetupServer(db *sql.DB) *grpc.Server {
	s := grpc.NewServer()
	pb.RegisterInvoiceServiceServer(s, &server{db: db})
	return s
}

func (s *server) PlaceBid(ctx context.Context, in *pb.Bid) (*pb.Bid, error) {

	// Check if the investor exists and has enough balance
	err := CheckInvestorBalance(ctx, s.db, in)
	if err != nil {
		return nil, err
	}

	// Reduce the investor's balance
	err = RededuceInvestorBalance(ctx, s.db, in)
	if err != nil {
		return nil, err
	}

	// Close previous bids
	err = CloseBids(ctx, s.db, in)
	if err != nil {
		return nil, err
	}

	// Determine the status of the bid
	status, err := DetermineBidStatus(ctx, s.db, in)
	if err != nil {
		return nil, err
	}

	// Insert the new bid
	err = InsertBid(ctx, s.db, in, status)
	if err != nil {
		return nil, err
	}

	if status == "approved" {
		// Update invoice status and investor id
		err = CloseInvoice(ctx, s.db, in)
		if err != nil {
			return nil, err
		}
	}
	// Update the invoice
	err = UpdateInvestorInInvoice(ctx, s.db, in)
	if err != nil {
		return nil, err
	}
	// Commit the transaction
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	bids, err := ListAllBids(ctx, s.db)

	if err != nil {
		return nil, err
	}

	log.Printf("Bids: %v", bids)

	return in, nil
}

// ApproveTrade approves a trade and set invoice status to closed
func (s *server) ApproveTrade(ctx context.Context, in *pb.Bid) (*pb.Bid, error) {
	log.Printf("Approving trade: %v", in)
	// Update invoice status and investor id
	log.Printf("Updating invoice: %v", in.GetInvoiceId())
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	// Rollback the transaction if anything goes wrong
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	err = CloseInvoice(ctx, s.db, in)

	if err != nil {
		log.Printf("Error updating invoice: %v", err)
		return nil, err
	}

	err = CloseBids(ctx, s.db, in)

	if err != nil {
		log.Printf("Error updating invoice: %v", err)
		return nil, err
	}

	err = CloseInvoice(ctx, s.db, in)
	if err != nil {
		log.Printf("Error updating invoice: %v", err)
		return nil, err
	}

	// Update issuer balance
	log.Printf("Updating Issuer")

	err = UpadeIssuerBalanceByBid(ctx, s.db, in)
	if err != nil {
		log.Printf("Error updating issuer balance: %v", err)
		return nil, err
	}
	bids, err := ListAllBids(ctx, s.db)

	if err != nil {
		return nil, err
	}

	log.Printf("Bids: %v", bids)
	log.Printf("Trade approved: %v", in)
	return in, nil
}

// CreateInvoice creates a new invoice with an existing issuer
func (s *server) CreateInvoice(ctx context.Context, in *pb.Invoice) (*pb.Invoice, error) {
	log.Printf("Received: %v", in.GetInvestorId())

	log.Printf("Issuer ID: %v, Status: %v, Investor ID: %v", in.GetIssuerId(), in.GetStatus(), in.GetInvestorId())

	if in.GetPrice() <= 0 {
		return nil, errors.New("price must be greater than 0")
	}
	var id string
	err := s.db.QueryRow("INSERT INTO invoice (issuer_id, status, investor_id, price) VALUES ($1, $2, $3, $4) RETURNING id", in.GetIssuerId(), in.GetStatus(), in.GetInvestorId(), in.GetPrice()).Scan(&id)
	if err != nil {
		return nil, err
	}
	log.Println("Inserted invoice into database")

	in.Id = id
	return in, nil
}

// GetIssuer returns an issuer by id
func (s *server) GetIssuer(ctx context.Context, in *pb.Issuer) (*pb.Issuer, error) {
	log.Printf("Received: %v", in.GetId())

	row := s.db.QueryRow("SELECT id, name, balance FROM issuer WHERE id = $1", in.GetId())

	issuer := &pb.Issuer{}
	err := row.Scan(&issuer.Id, &issuer.Name, &issuer.Balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("issuer not found")
		}
		return nil, err
	}

	return issuer, nil

}

// GetInvestors returns all investors in stream since it could be a large number of investors
func (s *server) GetInvestors(in *empty.Empty, stream pb.InvoiceService_GetInvestorsServer) error {
	rows, err := s.db.Query("SELECT id, name, balance FROM Investor")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		investor := &pb.Investor{}
		err := rows.Scan(&investor.Id, &investor.Name, &investor.Balance)
		if err != nil {
			return err
		}
		if err := stream.Send(investor); err != nil {
			return err
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

// GetInvoice returns an invoice by id
func (s *server) GetInvoice(ctx context.Context, in *pb.Invoice) (*pb.Invoice, error) {
	log.Printf("Received: %v", in.GetInvestorId())

	row := s.db.QueryRow("SELECT id, issuer_id, status, investor_id FROM invoice WHERE id = $1", in.GetId())

	invoice := &pb.Invoice{}
	err := row.Scan(&invoice.Id, &invoice.IssuerId, &invoice.Status, &invoice.InvestorId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("invoice not found")
		}
		return nil, err
	}

	return invoice, nil

}
