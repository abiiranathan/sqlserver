package sqlserver

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SqlResult struct {
	Columns []string
	Data    [][]string
	Error   string
}

func OpenDatabase(name string) (*sql.DB, error) {
	// Open the SQLite3 database file
	db, err := sql.Open("sqlite3", name)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	return db, err
}

func ExecQuery(db *sql.DB, conn net.Conn, query string) (*SqlResult, error) {
	// Execute the query and send the results back to the client
	ctx := context.Background()
	ctx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(time.Second*5))
	defer cancelFunc()

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// error ignored since rows are not closed
	columns, _ := rows.Columns()

	columnCount := len(columns)
	values := make([]interface{}, columnCount)
	valuePtrs := make([]interface{}, columnCount)
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Send rows to the client
	data := [][]string{}
	for rows.Next() {
		err = rows.Scan(valuePtrs...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[SQLSERVER]: error scanning row: %v\n", err)
			return nil, err
		}

		rowData := []string{}
		for _, value := range values {
			if value == nil {
				rowData = append(rowData, "NULL")
			} else {
				rowData = append(rowData, fmt.Sprintf("%v", value))
			}
		}
		data = append(data, rowData)
	}
	return &SqlResult{Columns: columns, Data: data}, nil
}

// Serialize and send the result as raw bytes
func SendQueryResult(conn net.Conn, data *SqlResult) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	encoder.Encode(data)
	_, err := io.Copy(conn, &buffer)
	if err != nil {
		fmt.Printf("[SQLSERVER]: error sending results: %v\n", err)
		return
	}
}

func HandleConnection(conn net.Conn, db *sql.DB) {
	fmt.Println("Connection from", conn.RemoteAddr())

	// We close each client connection when the communication loop ends.
	defer conn.Close()

	// Create a buffered IO Reader from the net.Conn
	reader := bufio.NewReader(conn)

	// Indifinite communication loop between server and client.
	for {
		var buf bytes.Buffer
		for {
			// We read the query until we encounter a semicolon.
			// if no semi-colon exists, it blocks.
			chunk, err := reader.ReadString(';')
			if err != nil && err != io.EOF {
				// give the client another life-line.
				break
			}

			// The client has disconnected. We return and it's connection will be closed.
			if err == io.EOF {
				return
			}

			// Happy path: Write this chunck to the buffer
			buf.WriteString(chunk)

			// We are at end of query
			if strings.Contains(chunk, ";") {
				break
			}
		}

		// Now we have the query. Hopefully safe!! :)
		query := strings.TrimSpace(buf.String())
		buf.Reset() // reset the buffer

		// Execute the query
		result, err := ExecQuery(db, conn, query)
		if err != nil {
			result = &SqlResult{
				Error: err.Error(),
			}
		}

		// Send the query results onto the connection.
		SendQueryResult(conn, result)
	}
}
