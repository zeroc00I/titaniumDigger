package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
)

// URLExtractor struct to hold necessary fields
type URLExtractor struct {
	tldFile   string
	id        string
	proxyURL  string
	client    *http.Client
}

// NewURLExtractor initializes a new URLExtractor
func NewURLExtractor(tldFile, id, proxyURL string) *URLExtractor {
	tr := &http.Transport{
		Proxy: http.ProxyURL(parseProxyURL(proxyURL)),
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: tr}
	return &URLExtractor{tldFile: tldFile, id: id, proxyURL: proxyURL, client: client}
}

// parseProxyURL parses and returns a proxy URL or nil if there's an error
func parseProxyURL(proxy string) *url.URL {
	if parsedURL, err := url.Parse(proxy); err == nil {
		return parsedURL
	} else {
		fmt.Printf("Error parsing proxy URL: %s\n", err)
		return nil
	}
}

// ReadTLDs reads the TLDs from the specified file and returns them as a slice.
func (ue *URLExtractor) ReadTLDs() ([]string, error) {
	var tlds []string
	file, err := os.Open(ue.tldFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		tld := strings.TrimSpace(scanner.Text())
		if tld != "" {
			tlds = append(tlds, tld)
		}
	}
	return tlds, scanner.Err()
}

// BuildURLRegex constructs a regex pattern to match valid URLs ending with specified TLDs.
func BuildURLRegex(tlds []string) *regexp.Regexp {
	tldPattern := strings.Join(tlds, "|")
	pattern := fmt.Sprintf(`(?i)\b(?:https?://|ftp://|www\.|[a-z0-9.-]+)\.(%s)[^"]*`, tldPattern)
	return regexp.MustCompile(pattern)
}

// ExtractURLs fetches the target URL and extracts unique URLs based on the TLDs.
func (ue *URLExtractor) ExtractURLs() (map[string]struct{}, error) {
	targetURL := fmt.Sprintf("https://googletagmanager.com/gtm.js?id=%s", ue.id)

	resp, err := ue.client.Get(targetURL)
	if err != nil {
		return nil, fmt.Errorf("error making GET request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received status code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	tlds, err := ue.ReadTLDs()
	if err != nil {
		return nil, fmt.Errorf("error reading TLDs: %w", err)
	}

	urlRegex := BuildURLRegex(tlds)

	re := regexp.MustCompile(`[{}()\[\],]`)
	splitContent := re.Split(string(body), -1)

	urlSet := make(map[string]struct{}) // Use a map to track unique URLs

	for _, content := range splitContent {
		for _, match := range urlRegex.FindAllString(content, -1) {
			urlSet[match] = struct{}{} // Add each match to the map
		}
	}

	return urlSet, nil
}

// FilterURLs removes URLs that contain any of the unwanted substrings.
func FilterURLs(urlSet map[string]struct{}, unwanted []string) map[string]struct{} {
	filteredSet := make(map[string]struct{})
	for url := range urlSet {
		isUnwanted := false
		for _, u := range unwanted {
			if strings.Contains(url, u) { // Check if the URL contains any unwanted substring
				isUnwanted = true
				break
			}
		}
		if !isUnwanted {
			filteredSet[url] = struct{}{}
		}
	}
	return filteredSet
}

func main() {
	if len(os.Args) < 4 { // Check for at least 4 arguments
		fmt.Println("Usage: go run extract_urls.go -w tld_list.txt [id]")
		return
	}

	tldFile := os.Args[2]
	id := os.Args[3]
	proxyURL := "http://127.0.0.1:8080" // Replace with your proxy URL if needed

	urlExtractor := NewURLExtractor(tldFile, id, proxyURL)

	urlSet, err := urlExtractor.ExtractURLs()
	if err != nil {
		fmt.Printf("Error extracting URLs: %s\n", err)
		return
	}

	unwantedSubstrings := []string{"browsingTopics","this.","browsingTopics",".brand","compatMod","google", "adservices", "facebook", "licdn", "gstatic", "jquery", "doubleclick", "hotjar", "linkedin", "recaptcha.net", "bing","pinimg","yimg"}
	filteredSet := FilterURLs(urlSet, unwantedSubstrings)

	if len(filteredSet) == 0 {
		fmt.Println("No valid URLs found.")
		return
	}

	//fmt.Println("Extracted URLs:")

	var sortedURLs []string
	for url := range filteredSet {
		url = strings.Replace(url, "\\/", "/", -1) // Replace all occurrences of "\/"
		url = strings.Replace(url, "\\", "/", -1)  // Replace all occurrences of "\"
		
		if strings.HasSuffix(url, "/") {
			url = strings.TrimSuffix(url, "/") // Remove the trailing slash
		}
		sortedURLs = append(sortedURLs, url)       // Collect URLs for sorting
	}

	sort.Strings(sortedURLs) // Sort the collected URLs
	for _, url := range sortedURLs {
		fmt.Println(url) // Print each unique URL that is not unwanted and sorted
	}
}
