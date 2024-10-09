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
func (ue *URLExtractor) ExtractURLs() (string, error) { // Change return type to string
	targetURL := fmt.Sprintf("https://googletagmanager.com/gtm.js?id=%s", ue.id)

	resp, err := ue.client.Get(targetURL)
	if err != nil {
		return "", fmt.Errorf("error making GET request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received status code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	return string(body), nil // Return the raw body as a string
}

// NormalizePath normalizes the input path by replacing \\/ with / and removing unnecessary characters.
func NormalizePath(path string) string {
	path = strings.ReplaceAll(path, "\\/", "/") // Normalize \/
	path = strings.ReplaceAll(path, "\\", "/")  // Replace all occurrences of \
	path = strings.ReplaceAll(path, "//", "/")  // Replace multiple slashes with a single slash
	return strings.TrimSpace(path)               // Trim any leading or trailing whitespace
}

// ExtractPaths extracts paths from arg1 values that start with \/
func ExtractPaths(body string) []string {
	var paths []string
	pathRegex := regexp.MustCompile(`"arg1":"(\\/[^"]+)"`) // Regex to match arg1 values starting with \/

	matches := pathRegex.FindAllStringSubmatch(body, -1)
	for _, match := range matches {
		if len(match) > 1 {
			path := NormalizePath(match[1]) // Normalize the extracted path
			if path != "" {                  // Only add non-empty paths
				paths = append(paths, path)
			}
		}
	}
	return paths
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

// CountDynamicTLDs counts unique TLDs found in the URLs.
func CountDynamicTLDs(urlSet map[string]struct{}) map[string]int {
	tldCount := make(map[string]int)
	tldRegex := regexp.MustCompile(`\.(\w+)$`) // Matches TLD at the end of a URL

	for url := range urlSet {
		if matches := tldRegex.FindStringSubmatch(url); len(matches) > 1 {
			tldCount[matches[1]]++ // Increment count for this TLD
		}
	}
	return tldCount
}

func main() {
	if len(os.Args) < 2 { // Check for at least 2 arguments (id)
		fmt.Println("Usage: go run extract_urls.go [id] [-tld tld1 tld2 ...] [-debug] [-path]")
		return
	}

	id := os.Args[1]
	proxyURL := "http://127.0.0.1:8080" // Replace with your proxy URL if needed

	defaultTLDs := []string{"COM", "BR", "NET", "ORG"}
	var customTLDs []string
	debugMode := false
	pathFlag := false

	for i := 2; i < len(os.Args); i++ {
        switch os.Args[i] {
        case "-tld":
            if i+1 < len(os.Args) {
                customTLDs = append(customTLDs, os.Args[i+1])
                i++ // Skip the next argument as it's a TLD value
            }
        case "-debug":
            debugMode = true // Set debug mode if -debug flag is present
        case "-path":
            pathFlag = true // Set path flag if -path is present
        }
    }

	var allTLDs []string
	if len(customTLDs) > 0 {
        allTLDs = append(defaultTLDs, customTLDs...) // Combine default and custom TLDs if provided
    } else {
        allTLDs = defaultTLDs // Use only default TLDs if no custom TLDs are provided
    }

	urlExtractor := NewURLExtractor(allTLDs, id, proxyURL)

	body, err := urlExtractor.ExtractURLs()
	if err != nil {
        fmt.Printf("Error extracting URLs: %s\n", err)
        return
    }

	var paths []string
	if pathFlag { // If path flag is enabled, extract paths from the raw body.
	    paths = ExtractPaths(body)
	    fmt.Println("\n --- Extracted Paths ---")
	    for _, path := range paths {
	        fmt.Println(NormalizePath(path)) // Normalize each extracted path before printing
	    }
    }

	urlRegex := BuildURLRegex(allTLDs)

	re := regexp.MustCompile(`[{}()\[\],]`)
	splitContent := re.Split(body, -1)

	urlSet := make(map[string]struct{}) // Use a map to track unique URLs

	for _, content := range splitContent {
		for _, match := range urlRegex.FindAllString(content, -1) {
			urlSet[NormalizePath(match)] = struct{}{} // Normalize each match before adding to the set
		}
	}

	unwantedSubstrings := []string{".command","youtube","computeGtmParameter","browsingTopics", "this.", "browsingTopics", ".brand", "compatMod", "google", "adservices", "facebook", "licdn", "gstatic", "jquery", "doubleclick", "hotjar", "linkedin", "recaptcha.net", "bing", "pinimg", "yimg"}
	filteredSet := FilterURLs(urlSet, unwantedSubstrings)

	if len(filteredSet) == 0 {
        fmt.Println("No valid URLs found.")
        return
    }

	var sortedURLs []string
	for url := range filteredSet {
        url = NormalizePath(url) // Normalize each URL before sorting and printing

        if strings.HasSuffix(url, "/") {
            url = strings.TrimSuffix(url, "/") // Remove the trailing slash if necessary
        }
        sortedURLs = append(sortedURLs, url) // Collect URLs for sorting
    }

	sort.Strings(sortedURLs) // Sort the collected URLs
	fmt.Println("\n --- URLs ---")
	for _, url := range sortedURLs {
        fmt.Println(url) // Print each unique URL that is not unwanted and sorted
    }

	if debugMode { // If debug mode is enabled, print the dynamic TLD report table.
        tldCounts := CountDynamicTLDs(filteredSet)
        
        fmt.Println("\nDynamic TLD | Occurrences")
        fmt.Println("--- | ---")
        for tld, count := range tldCounts {
            fmt.Printf("%-3s | %d\n", tld, count)
        }
    }
}
