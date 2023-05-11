package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/abiiranathan/sqlserver/sqlserver"
	_ "github.com/mattn/go-sqlite3"
)

var (
	host   = "0.0.0.0"
	port   = "9999"
	dbName = "db.sqlite3"
)

func init() {
	flag.StringVar(&host, "host", host, "Remote Hostname or IP address")
	flag.StringVar(&port, "port", port, "Remote port")
	flag.StringVar(&dbName, "db", dbName, "The database to open and serve")
}

func main() {
	flag.Parse()

	// Open database connection
	db, err := sqlserver.OpenDatabase(dbName)
	if err != nil {
		log.Fatalf("[SQLSERVER]: Error opening database: %v\n", err)
	}
	defer db.Close()

	// Create the TCP listener
	address := fmt.Sprintf("%s:%s", host, port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[SQLSERVER]: error creating listener: %v\n", err)
		os.Exit(1)
	}
	defer ln.Close()

	fmt.Println("[SQLSERVER]: Listening on", ln.Addr())

	// Handle incoming connections
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error accepting connection: %v\n", err)
			continue
		}
		go sqlserver.HandleConnection(conn, db)
	}
}
