package pkg

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	pb "github.com/berdebotond/bankable_technical_test/protos"
	"github.com/stretchr/testify/assert"
)

func TestPlaceBid(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer db.Close()

	server := &server{db: db}

	bid := &pb.Bid{
		InvestorId: "investor1",
		InvoiceId:  "invoice1",
		Amount:     100.0,
	}

	expectedQuery := "INSERT INTO bid (investor_id, invoice_id, amount) VALUES (?, ?, ?) RETURNING id"
	mock.ExpectQuery(expectedQuery).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))

	result, err := server.PlaceBid(context.Background(), bid)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "1", result.GetId())

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestApproveTrade(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer db.Close()

	server := &server{db: db}

	invoiceID := "invoice1"

	expectedQuery := "UPDATE invoice SET status = 'closed' WHERE id = ?"
	mock.ExpectExec(expectedQuery).
		WithArgs(invoiceID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	bid := &pb.Bid{InvoiceId: invoiceID}
	result, err := server.ApproveTrade(context.Background(), bid)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// Similarly, you can write tests for other functions such as CreateInvoice, GetIssuer, GetInvestors, and GetInvoice.
