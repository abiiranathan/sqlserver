package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/abiiranathan/sqlserver/sqlclient"
	"github.com/chzyer/readline"
)

var (
	host = "localhost"
	port = "9999"
)

func init() {
	flag.StringVar(&host, "host", host, "Remote Hostname or IP address")
	flag.StringVar(&port, "port", port, "Remote port")
}

func main() {
	flag.Parse()

	// Compose the address
	addr := fmt.Sprintf("%s:%s", host, port)
	fmt.Printf("Connecting to sqlserver on [%s]", addr)

	// Connect to remote sqlserver
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Printf("unable to sqlserver: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("\nWelcome to the sqlclient!\nEnter your SQL queries below. Type \"exit\" or \"q\" to quit.")

	// History file for the prompt
	HOME, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
	}

	// Put history file in home directory as a hidden file
	historyFile := filepath.Join(HOME, ".sqlclient")

	// Read a query from the user
	rl, err := readline.NewEx(&readline.Config{
		Prompt:                 "> ",
		HistoryFile:            historyFile,
		DisableAutoSaveHistory: true,
	})

	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		query := sqlclient.ReadQueryFromPrompt(rl)
		if query == "exit" || query == "q" {
			break
		}

		// Send the query to the server
		err = sqlclient.SendQuery(conn, query)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// Read the result from the server
		result, err := sqlclient.ReadResult(conn)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// Print the error and continue
		if result.Error != "" {
			fmt.Println(result.Error)
			continue
		}

		// Print the result and continue
		fmt.Println(sqlclient.FormatTable(result.Columns, result.Data))
	}
}
