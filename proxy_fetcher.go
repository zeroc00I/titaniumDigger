// go run proxy_fetcher.go -u https://u2x5r29q2fla6joilf8nwy8hx83zrpfe.oastify.com -w geo_blocking_probe/5zu47n4rxm.up.proxies -t 100 -m POST -d 'teste=1'
package main

import (
    "bufio"
    "flag"
    "fmt"
    "net"
    "net/url"
    "os"
    "strings"
    "sync"
    "time"
    
    "github.com/valyala/fasthttp"
    "github.com/valyala/fasthttp/fasthttpproxy"
)

var (
    successCount int
    failureCount int
    mu           sync.Mutex
    wg           sync.WaitGroup
)

func main() {
    method := flag.String("m", "GET", "HTTP method")
    data := flag.String("d", "", "Request body")
    proxylist := flag.String("x", "", "Comma-separated list of HTTP proxies")
    concurrency := flag.Int("t", 5, "Number of concurrent workers")
    targetURL := flag.String("u", "", "Target URL")
    proxyFile := flag.String("w", "", "File with proxy list (one per line, IP:PORT or http://IP:PORT format)")

    flag.Parse()

    var targets []string
    if *proxyFile != "" {
        targets = loadProxies(*proxyFile)
    } else if *proxylist != "" {
        targets = strings.Split(*proxylist, ",")
    } else {
        targets = []string{""} // Empty string represents no-proxy
    }

    workChan := make(chan string)

    // Start workers
    for i := 0; i < *concurrency; i++ {
        go worker(workChan, *method, *data, *targetURL)
    }

    // Feed work channel
    wg.Add(len(targets))
    for _, addr := range targets {
        workChan <- addr
    }
    
    wg.Wait()
    close(workChan)

    fmt.Printf("\nResults:\nSuccess:%d\nFailure:%d\n", successCount, failureCount)
}

func loadProxies(filename string) []string {
    var proxies []string
    file, err := os.Open(filename)
    if err != nil {
        fmt.Printf("Error opening proxy file: %v\n", err)
        return proxies
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        proxy := strings.TrimSpace(scanner.Text())
        if proxy != "" {
            parsedProxy, err := parseProxy(proxy)
            if err != nil {
                fmt.Printf("Invalid proxy format: %s\n", proxy)
                continue
            }
            proxies = append(proxies, parsedProxy)
        }
    }

    if err := scanner.Err(); err != nil {
        fmt.Printf("Error reading proxy file: %v\n", err)
    }

    return proxies
}

func parseProxy(proxy string) (string, error) {
    if strings.HasPrefix(proxy, "http://") || strings.HasPrefix(proxy, "https://") {
        u, err := url.Parse(proxy)
        if err != nil {
            return "", err
        }
        return u.Host, nil
    }
    
    host, port, err := net.SplitHostPort(proxy)
    if err != nil {
        return "", err
    }
    return net.JoinHostPort(host, port), nil
}

func worker(ch <-chan string, meth, d, target string) {
    for addr := range ch {
        client := &fasthttp.Client{}
        
        // Configure client timeout
        client.ReadTimeout = 10 * time.Second
        
        // Set up dialer if needed
        if addr != "" {
            client.Dial = fasthttpproxy.FasthttpHTTPDialer(addr)
        }
        
        req := fasthttp.AcquireRequest()
        resp := fasthttp.AcquireResponse()
        
        defer func() {
            fasthttp.ReleaseRequest(req)
            fasthttp.ReleaseResponse(resp)
            wg.Done()
        }()

        req.SetRequestURI(target)
        req.Header.SetMethod(meth)
        
        // Set request body if applicable
        if meth == "POST" || meth == "PUT" || meth == "PATCH" { 
            req.SetBodyString(d)
            req.Header.SetContentType("application/x-www-form-urlencoded")
        }

        err := client.Do(req, resp)
        
        mu.Lock()
        if err != nil || resp.StatusCode() >= 400 {
            failureCount++
            fmt.Printf("[FAIL] %s %s\n", addr, target)
        } else {
            successCount++
            fmt.Printf("[OK] %s %s (%d)\n", addr, target, resp.StatusCode())
        }
        mu.Unlock()
    }
}
