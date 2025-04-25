package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "html/template"
    "log"
    "net"
    "net/http"
    "os"
    "sync"
    "time"

    "github.com/gorilla/mux"
    _ "github.com/mattn/go-sqlite3"
    "github.com/microcosm-cc/bluemonday"
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
    log.Println("Fetching data from database...")
    rows, err := db.Query("SELECT * FROM data ORDER BY ID DESC LIMIT 100")
    if err != nil {
        log.Printf("Database query error: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var data []Data
    for rows.Next() {
        var d Data
        err = rows.Scan(&d.ID, &d.IP, &d.Content, &d.Datetime)
        if err != nil {
            log.Printf("Row scan error: %v", err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        data = append(data, d)
    }

    log.Printf("Returning %d records as JSON", len(data))
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(data); err != nil {
        log.Printf("JSON encoding error: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func main() {
    // Database initialization
    dbPath := "local.db"
    log.Printf("Opening database at %s", dbPath)
    var err error
    db, err = sql.Open("sqlite3", dbPath)
    if err != nil {
        log.Fatalf("Database open error: %v", err)
    }
    defer db.Close()

    // Table creation
    log.Println("Creating data table if not exists")
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS data (
        id INTEGER PRIMARY KEY, 
        ip TEXT, 
        content TEXT, 
        datetime DATETIME
    )`)
    if err != nil {
        log.Fatalf("Table creation error: %v", err)
    }

    // Web server setup
    r := mux.NewRouter()
    r.HandleFunc("/data", getData)
    r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        log.Println("Handling root request")
        
        // Fetch actual data for the template
        rows, err := db.Query("SELECT * FROM data ORDER BY ID DESC LIMIT 100")
        if err != nil {
            log.Printf("Template data query error: %v", err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        defer rows.Close()

        var templateData []Data
        for rows.Next() {
            var d Data
            err = rows.Scan(&d.ID, &d.IP, &d.Content, &d.Datetime)
            if err != nil {
                log.Printf("Template row scan error: %v", err)
                continue
            }
            templateData = append(templateData, d)
        }

        tmpl, err := template.New("webpage").Parse(templateText)
        if err != nil {
            log.Printf("Template parse error: %v", err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        log.Printf("Rendering template with %d items", len(templateData))
        if err := tmpl.Execute(w, templateData); err != nil {
            log.Printf("Template execute error: %v", err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    })

    // HTTP Server Configuration
    port := "12" // Original port from user's code
    srv := &http.Server{
        Handler:      r,
        Addr:         "0.0.0.0:" + port,
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
    }

    log.Printf("Starting web server on port %s", port)
    go func() {
        if err := srv.ListenAndServe(); err != nil {
            log.Printf("HTTP server error: %v", err)
        }
    }()

    // TCP Server Setup
    tcpPort := "65535"
    log.Printf("Starting TCP server on port %s", tcpPort)
    listener, err := net.Listen("tcp", "0.0.0.0:"+tcpPort)
    if err != nil {
        log.Fatalf("TCP listen error: %v", err)
    }
    defer listener.Close()

    var wg sync.WaitGroup

    for {
        log.Println("Waiting for TCP connections...")
        conn, err := listener.Accept()
        if err != nil {
            log.Printf("TCP accept error: %v", err)
            continue
        }
        
        log.Printf("New connection from %s", conn.RemoteAddr())
        wg.Add(1)
        go handleConnection(conn, db, &wg)
    }
}

func handleConnection(conn net.Conn, db *sql.DB, wg *sync.WaitGroup) {
    defer wg.Done()
    defer conn.Close()

    buffer := make([]byte, 1024)
    n, err := conn.Read(buffer)
    if err != nil {
        log.Printf("Read error: %v", err)
        return
    }
    data := string(buffer[:n])
    log.Printf("Received %d bytes: %q", n, data)

    sanitizer := bluemonday.UGCPolicy()
    sanitizedData := sanitizer.Sanitize(data)
    log.Printf("Sanitized data: %q", sanitizedData)

    _, err = db.Exec(
        "INSERT INTO data (ip, content, datetime) VALUES (?, ?, datetime('now'))",
        conn.RemoteAddr().String(),
        sanitizedData,
    )
    if err != nil {
        log.Printf("DB insert error: %v", err)
    } else {
        log.Println("Data inserted successfully")
    }
}
