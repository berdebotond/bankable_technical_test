package pkg

import (
	"database/sql"

	pb "github.com/berdebotond/bankable_technical_test/protos"
)

type server struct {
	db *sql.DB
	pb.UnimplementedInvoiceServiceServer
}
