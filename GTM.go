package main

import (
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
	tlds     []string
	id       string
	proxyURL string
	client   *http.Client
}

// NewURLExtractor initializes a new URLExtractor
func NewURLExtractor(tlds []string, id, proxyURL string) *URLExtractor {
	tr := &http.Transport{
		Proxy: http.ProxyURL(parseProxyURL(proxyURL)),
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: tr}
	return &URLExtractor{tlds: tlds, id: id, proxyURL: proxyURL, client: client}
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

	urlRegex := BuildURLRegex(ue.tlds)

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

// CountTLDs counts occurrences of each TLD found in the URLs.
func CountTLDs(urlSet map[string]struct{}, tlds []string) map[string]int {
	tldCount := make(map[string]int)
	for url := range urlSet {
		for _, tld := range tlds {
			if strings.HasSuffix(url, "."+strings.ToLower(tld)) || strings.HasSuffix(url, "."+strings.ToUpper(tld)) {
				tldCount[tld]++
			}
		}
	}
	return tldCount
}

func main() {
	if len(os.Args) < 2 { // Check for at least 2 arguments (id)
		fmt.Println("Usage: go run extract_urls.go [id] [-tld tld1 tld2 ...] [-debug]")
		return
	}

	id := os.Args[1]
	proxyURL := "http://127.0.0.1:8080" // Replace with your proxy URL if needed

	defaultTLDs := []string{"COM", "BR", "NET", "ORG"}
	var customTLDs []string
	debugMode := false

	for i := 2; i < len(os.Args); i++ {
		if os.Args[i] == "-tld" && i+1 < len(os.Args) {
			customTLDs = append(customTLDs, os.Args[i+1])
			i++ // Skip the next argument as it's a TLD value
		} else if os.Args[i] == "-debug" {
			debugMode = true // Set debug mode if -debug flag is present
		}
	}

	var allTLDs []string
	if len(customTLDs) > 0 {
        allTLDs = append(defaultTLDs, customTLDs...) // Combine default and custom TLDs if provided
    } else {
        allTLDs = defaultTLDs // Use only default TLDs if no custom TLDs are provided
    }

	urlExtractor := NewURLExtractor(allTLDs, id, proxyURL)

	urlSet, err := urlExtractor.ExtractURLs()
	if err != nil {
        fmt.Printf("Error extracting URLs: %s\n", err)
        return
    }

	unwantedSubstrings := []string{"browsingTopics", "this.", "browsingTopics", ".brand", "compatMod", "google", "adservices", "facebook", "licdn", "gstatic", "jquery", "doubleclick", "hotjar", "linkedin", "recaptcha.net", "bing", "pinimg", "yimg"}
	filteredSet := FilterURLs(urlSet, unwantedSubstrings)

	if len(filteredSet) == 0 {
        fmt.Println("No valid URLs found.")
        return
    }

	var sortedURLs []string
	for url := range filteredSet {
        url = strings.Replace(url, "\\/", "/", -1) // Replace all occurrences of "\/"
        url = strings.Replace(url, "\\", "/", -1)  // Replace all occurrences of "\"

        if strings.HasSuffix(url, "/") {
            url = strings.TrimSuffix(url, "/") // Remove the trailing slash
        }
        sortedURLs = append(sortedURLs, url) // Collect URLs for sorting
    }

	sort.Strings(sortedURLs) // Sort the collected URLs

	for _, url := range sortedURLs {
        fmt.Println(url) // Print each unique URL that is not unwanted and sorted
    }

	if debugMode { // If debug mode is enabled, print the TLD report table.
        tldCounts := CountTLDs(filteredSet, allTLDs)
        
        fmt.Println("\nTLD | Occurrences")
        fmt.Println("--- | ---")
        for tld, count := range tldCounts {
            fmt.Printf("%-3s | %d\n", tld, count)
        }
    }
}
