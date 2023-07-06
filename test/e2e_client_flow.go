package main

import (
	"context"
	"io"
	"log"
	"time"

	cfg "github.com/berdebotond/bankable_technical_test/config"
	"github.com/berdebotond/bankable_technical_test/pkg"
	pb "github.com/berdebotond/bankable_technical_test/protos"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {

	// Set up a connection to the server.
	config, err := cfg.LoadConfig()
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewInvoiceServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Get first investor
	db := pkg.SetupDatabase(config.DatabaseHost, config.DatabasePort, config.DatabaseUser, config.DatabasePassword, config.DatabaseName)
	var investorId string
	err = db.QueryRow("SELECT id FROM investor LIMIT 1").Scan(&investorId)
	if err != nil {
		log.Fatalf("could not get investors: %v", err)
	}

	// Get first issuer
	var issuerId string
	err = db.QueryRow("SELECT id FROM issuer LIMIT 1").Scan(&issuerId)
	if err != nil {
		log.Fatalf("could not get issuer: %v", err)
	}

	// Call CreateInvoice
	invoice, err := c.CreateInvoice(ctx, &pb.Invoice{IssuerId: issuerId, Status: "open", InvestorId: investorId, Price: 10})
	if err != nil {
		log.Fatalf("could not create invoice: %v", err)
	}
	// check if investor balance was updated
	row, err := db.Query("SELECT balance FROM investor WHERE id = $1", investorId)
	if err != nil {
		log.Fatalf("could not get investor: %v", err)
	}
	defer row.Close()
	var original_balance float32
	for row.Next() {
		err = row.Scan(&original_balance)
		if err != nil {
			log.Fatalf("could not get investor: %v", err)
		}
		log.Printf("Investor balance: %v", original_balance)
	}
	// Call GetInvoice
	log.Printf("Invoice: %v", invoice)
	_, err = c.GetInvoice(ctx, &pb.Invoice{Id: invoice.GetId()})
	if err != nil {
		log.Fatalf("could not get invoice: %v", err)
	}

	// Call GetIssuer
	_, err = c.GetIssuer(ctx, &pb.Issuer{Id: issuerId})
	if err != nil {
		log.Fatalf("could not get issuer: %v", err)
	}

	// Call GetInvestors
	stream, err := c.GetInvestors(ctx, &emptypb.Empty{})
	if err != nil {
		log.Fatalf("could not get investors: %v", err)
	}
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to receive: %v", err)
		}
	}

	// Call PlaceBid
	bid, err := c.PlaceBid(ctx, &pb.Bid{InvestorId: investorId, InvoiceId: invoice.Id, Amount: 100.0})
	if err != nil {
		log.Fatalf("could not place bid: %v", err)
	}
	log.Printf("Bid created with id: %v", bid.GetId())

	// Call PlaceBid 2
	bid2, err := c.PlaceBid(ctx, &pb.Bid{InvestorId: investorId, InvoiceId: invoice.Id, Amount: 120.0})
	if err != nil {
		log.Fatalf("could not place bid: %v", err)
	}
	log.Printf("Bid created with id: %v", bid2.GetId())
	// check if investor balance was updated
	row, err = db.Query("SELECT balance FROM investor WHERE id = $1", investorId)
	if err != nil {
		log.Fatalf("could not get investor: %v", err)
	}
	defer row.Close()
	for row.Next() {
		var balance float32
		err = row.Scan(&balance)
		if err != nil {
			log.Fatalf("could not get investor: %v", err)
		}
		log.Printf("Investor balance: %v", original_balance)
		if balance != original_balance-120.0 {
			log.Fatalf("investor balance should %v, got: %v", original_balance-120, balance)
		}
	}
	// Call ApproveTrade
	_, err = c.ApproveTrade(ctx, bid2)
	if err != nil {
		log.Fatalf("could not approve trade: %v", err)
	}
	// Check if trade was accepted and invoice was closed
	invoice, err = c.GetInvoice(ctx, &pb.Invoice{Id: invoice.GetId()})
	if err != nil {
		log.Fatalf("could not get invoice: %v", err)
	}

	// list all invoices
	row, err = db.Query("SELECT * FROM invoice")
	if err != nil {
		log.Fatalf("could not get invoices: %v", err)
	}
	defer row.Close()
	for row.Next() {
		var id string
		var issuerId string
		var investorId string
		var status string
		var price float32

		err = row.Scan(&id, &issuerId, &investorId, &status, &price)
		if err != nil {
			log.Fatalf("could not get invoices: %v", err)
		}
		log.Printf("Invoice: %v, %v, %v, %v, %v", id, issuerId, investorId, status, price)
	}
	if invoice.Status != "closed" {
		log.Fatalf("invoice status should be closed, got: %v", invoice.Status)
	}

	log.Println("Trade approved")
	log.Println("All tests passed")
}
