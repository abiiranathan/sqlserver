package sqlclient

import (
	"bytes"
	"encoding/gob"
	"net"
	"strings"

	"github.com/abiiranathan/sqlserver/sqlserver"
	"github.com/chzyer/readline"
	"github.com/jedib0t/go-pretty/v6/table"
)

// Reads a query from the prompt. Blocks until a semicolon.
// If query is exit or q, this function exits, returning the current line
// The upstream caller must check for these strings and exit the program.
func ReadQueryFromPrompt(rl *readline.Instance) string {
	var query string
	var cmds []string

	rl.CaptureExitSignal()
	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// Handle exit commands
		if line == "exit" || line == "q" {
			return line
		}

		cmds = append(cmds, line)
		if !strings.HasSuffix(line, ";") {
			rl.SetPrompt("  ")
			continue
		}

		query = strings.Join(cmds, " ")
		//lint:ignore SA4006 cmds is actually used.
		cmds = cmds[:0]

		rl.SetPrompt("> ")
		rl.SaveHistory(query)
		break
	}
	return query
}

// Write query string into the connection.
func SendQuery(conn net.Conn, query string) error {
	_, err := conn.Write([]byte(query))
	return err
}

// Read server reply from the connection into SqlResult
func ReadResult(conn net.Conn) (*sqlserver.SqlResult, error) {
	decoder := gob.NewDecoder(conn)
	var result sqlserver.SqlResult
	err := decoder.Decode(&result)

	if err != nil {
		return nil, err
	}
	// Return the result
	return &result, nil
}

func FormatTable(columnNames []string, rows [][]string) string {
	// Create a new table object
	tbl := table.NewWriter()
	tbl.SetStyle(table.StyleLight)

	// Set the column headers
	header := make(table.Row, len(columnNames))
	for i, name := range columnNames {
		header[i] = name
	}
	tbl.AppendHeader(header)

	// Add the rows of data
	for _, row := range rows {
		tblRow := table.Row{}
		for _, v := range row {
			tblRow = append(tblRow, v)
		}
		tbl.AppendRow(tblRow)
	}

	// Render the table to a string
	var buffer bytes.Buffer
	tbl.SetOutputMirror(&buffer)
	tbl.Render()
	return buffer.String()
}
