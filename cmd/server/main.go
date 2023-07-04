package main

import (
	"log"
	"net"

	"github.com/berdebotond/bankable_technical_test/pkg"
)

func main() {
	db := pkg.SetupDatabase()
	defer db.Close()

	s := pkg.SetupServer(db)

	log.Printf("Server started on port 50051")
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
