package main

// Made with love by zeroc00i

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
)

var needle =".git/config"

func main() {
	var scanner *bufio.Scanner
	fmt.Printf("%s",len(os.Args))

	if len(os.Args) == 3 {
		needle = os.Args[2]
	}
	if len(os.Args) > 1 {
		wordlistFile, err := os.Open(os.Args[1])
		if err != nil {
			panic(err)
		}
		defer wordlistFile.Close()
		scanner = bufio.NewScanner(wordlistFile)
	} else {
		scanner = bufio.NewScanner(os.Stdin)
	}
	


	for scanner.Scan() {
		u := scanner.Text()
		parsedUrl, err := url.Parse(u)
		if err != nil {
			fmt.Printf("Error parsing URL: %v\n", err)
			continue
		}
		if len(u) > 120 {
			//fmt.Printf("URL too %v long: Skipping\n",u)
			continue
		}
		pathParts := strings.Split(parsedUrl.Path, "/")
		for i := range pathParts {
			newPath := path.Join(pathParts[0:i+1]...)
			fmt.Printf("%s://%s/%s/%s\n", parsedUrl.Scheme, parsedUrl.Host, newPath, needle)
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	fmt.Println("the wordlist was fully completed")
	os.Exit(1)

}
