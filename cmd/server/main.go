package main

import (
	"fmt"
	"log"
	"net"
	"os"

	cfg "github.com/berdebotond/bankable_technical_test/config"
	"github.com/berdebotond/bankable_technical_test/pkg"
)

func main() {
	config, err := cfg.LoadConfig()
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}
	db := pkg.SetupDatabase(config.DatabaseHost, config.DatabasePort, config.DatabaseUser, config.DatabasePassword, config.DatabaseName)
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
