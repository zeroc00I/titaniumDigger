package main

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func handleClient(conn net.Conn, db *sql.DB) {
	defer conn.Close()

	clientAddress := conn.RemoteAddr()
	fmt.Printf("Connected to client from IP address: %s\n", clientAddress)

	response := "HTTP/2 200\r\n\r\n"
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Printf("Error sending response to client: %s\n", err)
		return
	}

	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("Error reading from client: %s\n", err)
			return
		}

		data := buffer[:n]
		fmt.Printf("Received data from client: %s\n", string(data))

		// Store data in the local database
		storeDataInDatabase(clientAddress.String(), string(data), db)
	}
}

func storeDataInDatabase(ip, content string, db *sql.DB) {
	stmt, err := db.Prepare("INSERT INTO data (ip, content, datetime) values (?, ?, ?)")
	if err != nil {
		fmt.Printf("Error preparing SQL statement: %s\n", err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(ip, content, time.Now())
	if err != nil {
		fmt.Printf("Error inserting data into the database: %s\n", err)
	}
}

func main() {
	host := "0.0.0.0"
	port := 65535

	// Open the SQLite database
	db, err := sql.Open("sqlite3", "local.db")
	if err != nil {
		fmt.Printf("Error opening the database: %s\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create the data table if it doesn't exist
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS data (id INTEGER PRIMARY KEY, ip TEXT, content TEXT, datetime DATETIME)")
	if err != nil {
		fmt.Printf("Error creating the data table: %s\n", err)
		os.Exit(1)
	}

	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		fmt.Printf("Error binding to port %d: %s\n", port, err)
		os.Exit(1)
	}

	fmt.Printf("Listening on port %d...\n", port)

	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %s\n", err)
			continue
		}

		go handleClient(conn, db)
	}
}
