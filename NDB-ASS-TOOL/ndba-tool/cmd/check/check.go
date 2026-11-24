package check

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type resultEntry struct {
	Target   string `json:"target"`
	Category string `json:"category"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
}

// RunCmd executes the `check` command.
func RunCmd(args []string) error {
	cmd := flag.NewFlagSet("check", flag.ExitOnError)
	filePath := cmd.String("file", "c2.txt", "input file containing targets (one per line)")
	outputPath := cmd.String("out", "result.json", "output JSON file path")
	timeoutSec := cmd.Int("timeout", 10, "timeout per target in seconds")
	cmd.Parse(args)

	f, err := os.Open(*filePath)
	if err != nil {
		return fmt.Errorf("unable to open input file: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var results []resultEntry

	client := &http.Client{Timeout: time.Duration(*timeoutSec) * time.Second}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") { // skip blanks/comments
			continue
		}

		entry := resultEntry{Target: line}
		category, isURL, isIP := categorize(line)
		entry.Category = category

		connected, errStr := checkTarget(line, isURL, isIP, client, time.Duration(*timeoutSec)*time.Second)
		if connected {
			entry.Status = "connected"
		} else {
			entry.Status = "not_connected"
		}
		entry.Error = errStr

		results = append(results, entry)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed reading input: %v", err)
	}

	// Write output
	outF, err := os.Create(*outputPath)
	if err != nil {
		return fmt.Errorf("unable to create output file: %v", err)
	}
	defer outF.Close()

	enc := json.NewEncoder(outF)
	enc.SetIndent("", "  ")
	if err := enc.Encode(results); err != nil {
		return fmt.Errorf("failed to encode JSON: %v", err)
	}

	fmt.Printf("Wrote %d results to %s\n", len(results), *outputPath)
	return nil
}

// categorize returns a category label and booleans for URL/IP.
func categorize(s string) (string, bool, bool) {
	if looksLikeURL(s) {
		return "url", true, false
	}
	if net.ParseIP(s) != nil {
		return "ip", false, true
	}
	return "domain", false, false
}

func looksLikeURL(s string) bool {
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		_, err := url.ParseRequestURI(s)
		return err == nil
	}
	return false
}

func checkTarget(target string, isURL, isIP bool, httpClient *http.Client, timeout time.Duration) (bool, string) {
	if isURL {
		// HTTP GET
		resp, err := httpClient.Get(target)
		if err != nil {
			return false, err.Error()
		}
		defer resp.Body.Close()
		// consider any HTTP response as reachable
		return true, ""
	}

	if isIP {
		// TCP dial common ports to test reachability. Try 80 then 443.
		ports := []string{"80", "443"}
		for _, p := range ports {
			conn, err := net.DialTimeout("tcp", net.JoinHostPort(target, p), timeout)
			if err == nil {
				conn.Close()
				return true, ""
			}
		}
		return false, "tcp connect timeout or refused"
	}

	// Domain: try HTTP HEAD on https, then http
	urls := []string{"https://" + target, "http://" + target}
	for _, u := range urls {
		req, _ := http.NewRequest(http.MethodHead, u, nil)
		resp, err := httpClient.Do(req)
		if err == nil {
			resp.Body.Close()
			return true, ""
		}
	}
	return false, "unreachable via http/https"
}
