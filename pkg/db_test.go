package pkg

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	pb "github.com/berdebotond/bankable_technical_test/protos"
)

func TestCloseInvoice(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open stub database connection: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	bid := &pb.Bid{
		Id:         "bid-id",
		InvestorId: "investor-id",
		InvoiceId:  "invoice-id",
		Amount:     100,
		Status:     "approved",
	}

	// Update the regular expression to match the actual SQL query
	mock.ExpectExec("UPDATE invoice SET status = 'closed', investor_id = \\$1 WHERE id = \\$2").
		WithArgs(bid.GetInvestorId(), bid.GetInvoiceId()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = CloseInvoice(ctx, db, bid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestInsertBid(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open stub database connection: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	bid := &pb.Bid{
		Id:         "bid-id",
		InvestorId: "investor-id",
		InvoiceId:  "invoice-id",
		Amount:     100,
		Status:     "approved",
	}

	mock.ExpectExec("INSERT INTO bid \\(investor_id, invoice_id, amount, status\\) VALUES \\(\\$1, \\$2, \\$3, 'pending'\\)").
		WithArgs(bid.InvestorId, bid.InvoiceId, bid.Amount).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = InsertBid(ctx, db, bid, "pending")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestDetermineBidStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open stub database connection: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	bid := &pb.Bid{
		Id:         "bid-id",
		InvestorId: "investor-id",
		InvoiceId:  "invoice-id",
		Amount:     100,
		Status:     "approved",
	}

	mock.ExpectQuery("SELECT price FROM invoice WHERE id = \\$1").
		WithArgs(bid.InvoiceId).
		WillReturnRows(sqlmock.NewRows([]string{"price"}).AddRow(100))

	status, err := DetermineBidStatus(ctx, db, bid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if status != "approved" {
		t.Errorf("unexpected status: got %v want %v", status, "approved")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestUpdateInvestorInInvoice(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open stub database connection: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	bid := &pb.Bid{
		Id:         "bid-id",
		InvestorId: "investor-id",
		InvoiceId:  "invoice-id",
		Amount:     100,
		Status:     "approved",
	}

	mock.ExpectExec("UPDATE invoice SET investor_id = \\$1 WHERE id = \\$2").
		WithArgs(bid.InvestorId, bid.InvoiceId).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = UpdateInvestorInInvoice(ctx, db, bid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestCloseBids(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open stub database connection: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	bid := &pb.Bid{
		Id:         "bid-id",
		InvestorId: "investor-id",
		InvoiceId:  "invoice-id",
		Amount:     100,
		Status:     "approved",
	}

	mock.ExpectExec("UPDATE bid SET status = 'closed' WHERE invoice_id = \\$1").
		WithArgs(bid.InvoiceId).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE investor SET balance = balance + bid.amount FROM bid WHERE bid.invoice_id = $1 AND investor.id = bid.investor_id")).
		WithArgs(bid.InvoiceId).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = CloseBids(ctx, db, bid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestIncreasePreviousInvestorsBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open stub database connection: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	bid := &pb.Bid{
		Id:         "bid-id",
		InvestorId: "investor-id",
		InvoiceId:  "invoice-id",
		Amount:     100,
		Status:     "approved",
	}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE investor SET balance = balance + bid.amount FROM bid WHERE bid.invoice_id = $1 AND investor.id = bid.investor_id")).
		WithArgs(bid.InvoiceId).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = IncreasePreviousInvestorsBalance(ctx, db, bid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}
