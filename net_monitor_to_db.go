package main

import (
    "database/sql"
    "fmt"
    "html/template"
    "net"
    "net/http"
    "os"
    "sync"
    "time"

    "github.com/gorilla/mux"
    _ "github.com/mattn/go-sqlite3"
    "github.com/microcosm-cc/bluemonday"
    "encoding/json"
)

var (
    db           *sql.DB
    templateText = `
    <!DOCTYPE html>
    <html>
    <head>
        <title>SQLite Data Table</title>
    <style>
            body {
                font-family: Arial, sans-serif;
            }
    
            h1 {
                color: #333;
            }
    
            table {
                width: 100%;
table-layout: fixed;
                max-width: 100%;
                border-collapse: collapse;
                margin: 20px 0;
            }
    
            table, th, td {
                border: 1px solid #333;
            }
    
            th, td {
                padding: 8px;
                text-align: left;
                overflow: hidden;
                white-space: nowrap;
            }
    
            th {
                background-color: #333;
                color: #fff;
            }
            td:nth-child(3) {
        white-space: normal; /* Allow text to wrap to the next line */
    }
            tr:nth-child(even) {
                background-color: #f2f2f2;
            }
        </style>
	</head>
    <body>
        <h1>Junk VPS Traffic</h1>
        <table id="data-table">
            <thead>
                <tr>
                    <th>ID</th>
                    <th>IP</th>
                    <th>Content</th>
                    <th>Datetime</th>
                </tr>
            </thead>
            <tbody id="table-body">
                {{range .}}
                <tr>
                    <td>{{.ID}}</td>
                    <td>{{.IP}}</td>
                    <td>{{.Content | html}}</td>
                    <td>{{.Datetime}}</td>
                </tr>
                {{end}}
            </tbody>
        </table>
        <script>
            function updateTable() {
                fetch("/data")
                    .then(response => response.json())
                    .then(data => {
                        const table = document.getElementById("data-table");
                        const tbody = document.getElementById("table-body");
    
                        // Clear the existing rows
                        tbody.innerHTML = "";
    
                        // Iterate through the data in reverse order (latest first)
                        for (let i = 0; i <= data.length; i++) {
                            const row = data[i];
                            const tr = document.createElement("tr");
                            tr.innerHTML = "<td>" + row.ID + "</td><td>" + row.IP + "</td><td>" + row.Content + "</td><td>" + row.Datetime + "</td>";
                            tbody.appendChild(tr);
                        }
                    });
            }
            updateTable(); // Initial update
            setInterval(updateTable, 3000); // Update every 3 seconds
        </script>
    </body>
    </html>
    `
)

type Data struct {
    ID       int
    IP       string
    Content  string
    Datetime string
}

func getData(w http.ResponseWriter, r *http.Request) {
    var data []Data
    rows, err := db.Query("SELECT * FROM data ORDER BY ID DESC LIMIT 100")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    for rows.Next() {
        var d Data
        err = rows.Scan(&d.ID, &d.IP, &d.Content, &d.Datetime)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        data = append(data, d)
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    if err := json.NewEncoder(w).Encode(data); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func main() {
    // Open the SQLite database
    dbPath := "local.db"
    var err error
    db, err = sql.Open("sqlite3", dbPath)
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

    // Set up the web server
    r := mux.NewRouter()
    r.HandleFunc("/data", getData)
    r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        tmpl, err := template.New("webpage").Parse(templateText)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        if err := tmpl.Execute(w, nil); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    })

    srv := &http.Server{
        Handler:      r,
	Addr:         "0.0.0.0:12",
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
    }

    go func() {
        if err := srv.ListenAndServe(); err != nil {
            fmt.Printf("HTTP server stopped: %s\n", err)
        }
    }()
    // ...
    fmt.Println("Web server is running on http://localhost:8080")

    // Start the server listening on port 65535
    listener, err := net.Listen("tcp", "0.0.0.0:65535")
    if err != nil {
        fmt.Printf("Error starting server: %s\n", err)
        os.Exit(1)
    }
    defer listener.Close()

    var wg sync.WaitGroup // WaitGroup to manage Goroutines

    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Printf("Error accepting connection: %s\n", err)
            continue
        }
        wg.Add(1)
        go handleConnection(conn, db, &wg) // Handle each connection concurrently
    }

    select {}
}

func handleConnection(conn net.Conn, db *sql.DB, wg *sync.WaitGroup) {
    defer wg.Done()
    defer conn.Close()

    // Read data from the connection
    buffer := make([]byte, 1024)
    n, err := conn.Read(buffer)
    if err != nil {
        fmt.Printf("Error reading data: %s\n", err)
        return
    }
    data := string(buffer[:n])

    // Extract the client's IP address and port number
    remoteAddr := conn.RemoteAddr().String()

    // Sanitize the HTML content using bluemonday
    sanitizer := bluemonday.UGCPolicy()
    sanitizedData := sanitizer.Sanitize(data)

    // Insert data, IP, sanitized content, and port into the database
    _, err = db.Exec("INSERT INTO data (ip, content, datetime) VALUES (?, ?, datetime('now'))", remoteAddr, sanitizedData)
    if err != nil {
        fmt.Printf("Error inserting data into the database: %s\n", err)
    }
}
