package pkg

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	pb "github.com/berdebotond/bankable_technical_test/protos"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type mockInvestorStream struct {
	grpc.ServerStream
	Responses []*pb.Investor
}

func (x *mockInvestorStream) Send(m *pb.Investor) error {
	x.Responses = append(x.Responses, m)
	return nil
}
func TestGetIssuer(t *testing.T) {
	// Setup
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	s := &server{db: db}

	// Mock database
	rows := sqlmock.NewRows([]string{"id", "name", "balance"}).AddRow("1", "Issuer Name", 100.0)
	mock.ExpectQuery("SELECT id, name, balance FROM issuer WHERE id = \\$1").WithArgs("1").WillReturnRows(rows)

	// Test
	issuer, err := s.GetIssuer(context.Background(), &pb.Issuer{Id: "1"})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, issuer)
	assert.Equal(t, "1", issuer.Id)
	assert.Equal(t, "Issuer Name", issuer.Name)
	assert.Equal(t, float32(100.0), issuer.Balance)
}

// Similarly, you can write tests for other functions such as CreateInvoice, GetIssuer, GetInvestors, and GetInvoice.

func TestGetIssuerNotFound(t *testing.T) {
	// Setup
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	s := &server{db: db}

	// Mock database
	mock.ExpectQuery("SELECT id, name, balance FROM issuer WHERE id = \\$1").WithArgs("nonexistent").WillReturnError(sql.ErrNoRows)

	// Test
	issuer, err := s.GetIssuer(context.Background(), &pb.Issuer{Id: "nonexistent"})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, issuer)
	assert.Equal(t, "issuer not found", err.Error())
}

func TestGetInvoiceNotFound(t *testing.T) {
	// Setup
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	s := &server{db: db}

	// Mock database
	mock.ExpectQuery("SELECT id, issuer_id, status, investor_id FROM invoice WHERE id = \\$1").WithArgs("nonexistent").WillReturnError(sql.ErrNoRows)

	// Test
	invoice, err := s.GetInvoice(context.Background(), &pb.Invoice{Id: "nonexistent"})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, invoice)
	assert.Equal(t, "invoice not found", err.Error())
}

func TestGetInvestors(t *testing.T) {
	// Create a new mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer db.Close()

	// Define the expected result
	expectedInvestors := []*pb.Investor{
		{Id: "1", Balance: 1000.0, Name: "Investor 1"},
		{Id: "2", Balance: 2000.0, Name: "Investor 2"},
	}

	// Set up the mock database to return the expected result
	rows := sqlmock.NewRows([]string{"id", "name", "balance"})
	for _, investor := range expectedInvestors {
		rows.AddRow(investor.Id, investor.Name, investor.Balance)
	}
	mock.ExpectQuery("SELECT id, name, balance FROM Investor").WillReturnRows(rows)

	// Create a new server with the mock database
	s := &server{db: db}

	// Create a mock stream
	stream := &mockInvestorStream{}

	// Call the function and check the result
	err = s.GetInvestors(nil, stream)
	assert.NoError(t, err)
	assert.Equal(t, expectedInvestors, stream.Responses)
}
