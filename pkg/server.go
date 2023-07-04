package pkg

import (
	"context"
	"database/sql"
	"errors"
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

// PlaceBid places a bid on an invoice
func (s *server) PlaceBid(ctx context.Context, in *pb.Bid) (*pb.Bid, error) {
	var bidId string
	err := s.db.QueryRow("INSERT INTO bid (investor_id, invoice_id, amount) VALUES ($1, $2, $3) RETURNING id", in.GetInvestorId(), in.GetInvoiceId(), in.GetAmount()).Scan(&bidId)
	if err != nil {
		return nil, err
	}
	in.Id = bidId
	log.Printf("Placed bid: %v", in)
	return in, nil
}

// ApproveTrade approves a trade and set invoice status to closed
func (s *server) ApproveTrade(ctx context.Context, in *pb.Bid) (*pb.Bid, error) {
	_, err := s.db.Exec("UPDATE invoice SET status = 'closed' WHERE id = $1", in.GetInvoiceId())
	if err != nil {
		return nil, err
	}
	log.Printf("Trade approved: %v", in)
	return in, nil
}

// CreateInvoice creates a new invoice with an existing issuer
func (s *server) CreateInvoice(ctx context.Context, in *pb.Invoice) (*pb.Invoice, error) {
	log.Printf("Received: %v", in.GetInvestorId())

	log.Printf("Issuer ID: %v, Status: %v, Investor ID: %v", in.GetIssuerId(), in.GetStatus(), in.GetInvestorId())

	var id string
	err := s.db.QueryRow("INSERT INTO invoice (issuer_id, status, investor_id) VALUES ($1, $2, $3) RETURNING id", in.GetIssuerId(), in.GetStatus(), in.GetInvestorId()).Scan(&id)
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

	row := s.db.QueryRow("SELECT id, balance FROM issuer WHERE id = $1", in.GetId())

	issuer := &pb.Issuer{}
	err := row.Scan(&issuer.Id, &issuer.Balance)
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
	rows, err := s.db.Query("SELECT id, balance FROM Investor")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		investor := &pb.Investor{}
		err := rows.Scan(&investor.Id, &investor.Balance)
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
