package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	asnmap "github.com/projectdiscovery/asnmap/libs"
)

const (
	ToolName     = "ASN Explorer"
	Version      = "4.0.0"
	DefaultUA    = "Mozilla/5.0 (compatible; ASNExplorer/4.0; +https://github.com/you/asn-explorer)"
	APIEndpoint  = "https://api.asrank.caida.org/v2/graphql"
	DefaultProxy = "http://127.0.0.1:8080"
)

type Config struct {
	Threads        int
	File           string
	ASN            string
	Org            string
	ExtractCIDR    bool
	ExtractDomains bool
	Proxy          string
	Verbose        bool
}

type ASNResponse struct {
	Data struct {
		ASNs struct {
			TotalCount int `json:"totalCount"`
			Edges      []struct {
				Node struct {
					ASN          string `json:"asn"`
					ASNName      string `json:"asnName"`
					Organization struct {
						OrgName string `json:"orgName"`
					} `json:"organization"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"asns"`
	} `json:"data"`
}

type CRTEntry struct {
	CommonName string `json:"common_name"`
}

type loggingTransport struct {
	transport http.RoundTripper
	config    *Config
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.config.Verbose {
		log.Printf("[REQUEST] %s %s", req.Method, req.URL)
		log.Printf("[HEADERS] %+v", req.Header)
		if req.Body != nil {
			body, _ := io.ReadAll(req.Body)
			req.Body = io.NopCloser(bytes.NewBuffer(body))
			log.Printf("[BODY] %s", body)
		}
	}

	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		if t.config.Verbose {
			log.Printf("[ERROR] %v", err)
		}
		return nil, err
	}

	if t.config.Verbose {
		log.Printf("[RESPONSE] %s", resp.Status)
		log.Printf("[RESPONSE HEADERS] %+v", resp.Header)
		body, _ := io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewBuffer(body))
		log.Printf("[RESPONSE BODY] %s", body)
	}

	return resp, nil
}

func createClient(cfg *Config) (*http.Client, error) {
	baseTransport := &http.Transport{
		DisableKeepAlives: true,
	}

	if cfg.Proxy != "" {
		proxyURL, err := url.Parse(cfg.Proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %v", err)
		}
		baseTransport.Proxy = http.ProxyURL(proxyURL)
	}

	var transport http.RoundTripper = baseTransport

	if cfg.Verbose {
		transport = &loggingTransport{
			transport: transport,
			config:    cfg,
		}
	}

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}, nil
}

func parseFlags() Config {
	var cfg Config
	flag.IntVar(&cfg.Threads, "t", 4, "Number of threads")
	flag.StringVar(&cfg.File, "f", "", "Input file")
	flag.StringVar(&cfg.ASN, "a", "", "Specific ASN(s) (comma-separated)")
	flag.StringVar(&cfg.Org, "o", "", "Organization to search")
	flag.BoolVar(&cfg.ExtractCIDR, "ec", false, "Extract CIDRs")
	flag.BoolVar(&cfg.ExtractDomains, "ed", false, "Extract domains")
	flag.StringVar(&cfg.Proxy, "p", DefaultProxy, "Proxy server")
	flag.BoolVar(&cfg.Verbose, "v", false, "Verbose output")
	flag.Parse()

	proxySet := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "p" {
			proxySet = true
		}
	})
	if !proxySet {
		cfg.Proxy = ""
	}

	return cfg
}

func main() {
	log.Printf("%s v%s", ToolName, Version)
	cfg := parseFlags()
	ctx := context.Background()

	httpClient, err := createClient(&cfg)
	if err != nil {
		log.Fatalf("Error creating HTTP client: %v", err)
	}

	asnClient, err := asnmap.NewClient()
	if err != nil {
		log.Fatalf("Error creating ASN client: %v", err)
	}

	var wg sync.WaitGroup

	if cfg.ASN != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			processASNs(ctx, cfg, asnClient)
		}()
	}

	if cfg.Org != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			processOrganization(ctx, cfg, httpClient, asnClient)
		}()
	}

	wg.Wait()
}

func processASNs(ctx context.Context, cfg Config, client *asnmap.Client) {
	log.Printf("Processing ASN(s): %s", cfg.ASN)
	asns := strings.Split(cfg.ASN, ",")

	var wg sync.WaitGroup
	asnChan := make(chan string, len(asns))
	resultChan := make(chan []string, len(asns))
	seen := sync.Map{}

	// Launch workers
	for i := 0; i < cfg.Threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for asn := range asnChan {
				cleanASN := strings.TrimPrefix(strings.TrimSpace(asn), "AS")
				cidrs, err := getASNPrefixes(ctx, client, cleanASN, cfg)
				if err != nil && cfg.Verbose {
					log.Printf("Error processing ASN %s: %v", cleanASN, err)
					continue
				}
				resultChan <- cidrs
			}
		}()
	}

	// Feed ASNs to workers
	go func() {
		for _, asn := range asns {
			asnChan <- asn
		}
		close(asnChan)
	}()

	// Process results
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect unique results
	for cidrs := range resultChan {
		for _, cidr := range cidrs {
			if _, loaded := seen.LoadOrStore(cidr, struct{}{}); !loaded {
				fmt.Println(cidr)
			}
		}
	}
}

func processOrganization(ctx context.Context, cfg Config, httpClient *http.Client, asnClient *asnmap.Client) {
	log.Printf("Processing organization: %s", cfg.Org)
	result, err := getASNsForOrg(ctx, httpClient, cfg.Org, cfg)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	log.Printf("Found %d ASNs", result.Data.ASNs.TotalCount)

	if cfg.ExtractCIDR {
		processOrgCIDRs(ctx, cfg, asnClient, result)
	}

	if cfg.ExtractDomains {
		processOrgDomains(ctx, cfg, httpClient, result)
	} else {
		printASNInfo(result)
	}
}

func getASNsForOrg(ctx context.Context, client *http.Client, org string, cfg Config) (*ASNResponse, error) {
	escapedOrg := strings.ReplaceAll(org, `"`, `\"`)
	query := fmt.Sprintf(`query { asns(name: "%s", first:10000, offset:0, sort:"rank") { 
		totalCount 
		edges { 
			node { 
				asn 
				asnName
				organization { 
					orgName 
				} 
			} 
		} 
	} }`, escapedOrg)

	reqBody := struct {
		Query string `json:"query"`
	}{Query: query}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(reqBody); err != nil {
		return nil, fmt.Errorf("encoding request failed: %v", err)
	}

	resp, err := client.Post(APIEndpoint, "application/json", &buf)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s\nResponse: %s", resp.Status, body)
	}

	var result ASNResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding failed: %v", err)
	}

	return &result, nil
}

func getASNPrefixes(ctx context.Context, client *asnmap.Client, asn string, cfg Config) ([]string, error) {
	responses, err := client.GetData(asn)
	if err != nil {
		return nil, fmt.Errorf("asnmap query failed: %v", err)
	}

	var cidrs []string
	for _, response := range responses {
		ipnets, err := asnmap.GetCIDR([]*asnmap.Response{response})
		if err != nil {
			continue // Skip invalid entries
		}
		for _, ipnet := range ipnets {
			cidrs = append(cidrs, ipnet.String())
		}
	}

	return cidrs, nil
}

func getDomainsForOrg(ctx context.Context, client *http.Client, org string, cfg Config) ([]string, error) {
	url := fmt.Sprintf("https://crt.sh/?q=%s&output=json", url.QueryEscape(org))
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", DefaultUA)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var entries []CRTEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	var domains []string
	for _, entry := range entries {
		domain := strings.ToLower(strings.TrimSpace(entry.CommonName))
		if domain != "" && !strings.Contains(domain, "*") {
			if _, exists := seen[domain]; !exists {
				seen[domain] = struct{}{}
				domains = append(domains, domain)
			}
		}
	}

	return domains, nil
}

func processOrgCIDRs(ctx context.Context, cfg Config, client *asnmap.Client, result *ASNResponse) {
	var wg sync.WaitGroup
	asnChan := make(chan string, len(result.Data.ASNs.Edges))
	resultChan := make(chan []string, len(result.Data.ASNs.Edges))
	seen := sync.Map{}

	for i := 0; i < cfg.Threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for asn := range asnChan {
				cleanASN := strings.TrimPrefix(asn, "AS")
				cidrs, err := getASNPrefixes(ctx, client, cleanASN, cfg)
				if err != nil && cfg.Verbose {
					log.Printf("Error processing ASN %s: %v", cleanASN, err)
					continue
				}
				resultChan <- cidrs
			}
		}()
	}

	go func() {
		for _, edge := range result.Data.ASNs.Edges {
			asnChan <- edge.Node.ASN
		}
		close(asnChan)
	}()

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for cidrs := range resultChan {
		for _, cidr := range cidrs {
			if _, loaded := seen.LoadOrStore(cidr, struct{}{}); !loaded {
				fmt.Println(cidr)
			}
		}
	}
}

func processOrgDomains(ctx context.Context, cfg Config, client *http.Client, result *ASNResponse) {
	orgs := make(map[string]struct{})
	for _, edge := range result.Data.ASNs.Edges {
		orgs[strings.ToLower(edge.Node.Organization.OrgName)] = struct{}{}
	}

	var wg sync.WaitGroup
	orgChan := make(chan string, len(orgs))
	resultChan := make(chan []string, len(orgs))
	seen := sync.Map{}

	for i := 0; i < cfg.Threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for org := range orgChan {
				domains, err := getDomainsForOrg(ctx, client, org, cfg)
				if err != nil && cfg.Verbose {
					log.Printf("Error processing %s: %v", org, err)
					continue
				}
				resultChan <- domains
			}
		}()
	}

	go func() {
		for org := range orgs {
			orgChan <- org
		}
		close(orgChan)
	}()

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for domains := range resultChan {
		for _, domain := range domains {
			if _, loaded := seen.LoadOrStore(domain, struct{}{}); !loaded {
				fmt.Println(domain)
			}
		}
	}
}

func printASNInfo(result *ASNResponse) {
	for _, edge := range result.Data.ASNs.Edges {
		node := edge.Node
		fmt.Printf("[ASN %s] %s (%s)\n", node.ASN, node.ASNName, node.Organization.OrgName)
	}
}
