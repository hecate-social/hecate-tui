package llmtools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// RegisterWebSearchTools adds web search and fetch tools to the registry.
func RegisterWebSearchTools(r *Registry) {
	r.Register(webSearchTool(), webSearchHandler)
	r.Register(webFetchTool(), webFetchHandler)
}

// --- web_search ---

func webSearchTool() Tool {
	params := NewObjectParameters()
	params.AddProperty("query", String("Search query"))
	params.AddProperty("num_results", Integer("Number of results to return (default: 5, max: 10)"))
	params.AddRequired("query")

	return Tool{
		Name:             "web_search",
		Description:      "Search the web using DuckDuckGo. Returns titles, URLs, and snippets for matching pages.",
		Parameters:       params,
		Category:         CategoryWeb,
		RequiresApproval: false,
	}
}

type webSearchArgs struct {
	Query      string `json:"query"`
	NumResults int    `json:"num_results"`
}

// SearchResult represents a single search result.
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

func webSearchHandler(ctx context.Context, args json.RawMessage) (string, error) {
	var a webSearchArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if a.Query == "" {
		return "", fmt.Errorf("query is required")
	}

	numResults := a.NumResults
	if numResults <= 0 {
		numResults = 5
	}
	if numResults > 10 {
		numResults = 10
	}

	results, err := duckDuckGoSearch(ctx, a.Query, numResults)
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return fmt.Sprintf("No results found for: %s", a.Query), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Search results for '%s':\n\n", a.Query))

	for i, r := range results {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, r.Title))
		sb.WriteString(fmt.Sprintf("   URL: %s\n", r.URL))
		if r.Snippet != "" {
			sb.WriteString(fmt.Sprintf("   %s\n", r.Snippet))
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// duckDuckGoSearch performs a search using DuckDuckGo HTML interface.
// DuckDuckGo's API is limited, so we scrape the HTML lite version.
func duckDuckGoSearch(ctx context.Context, query string, numResults int) ([]SearchResult, error) {
	// Use the HTML lite version which is easier to parse
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set a browser-like User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return parseDuckDuckGoResults(string(body), numResults), nil
}

// parseDuckDuckGoResults extracts search results from DuckDuckGo HTML.
func parseDuckDuckGoResults(html string, maxResults int) []SearchResult {
	var results []SearchResult

	// Match result blocks: <a class="result__a" href="...">Title</a>
	// and <a class="result__snippet">Snippet</a>
	resultPattern := regexp.MustCompile(`<a[^>]*class="result__a"[^>]*href="([^"]*)"[^>]*>([^<]*)</a>`)
	snippetPattern := regexp.MustCompile(`<a[^>]*class="result__snippet"[^>]*>([^<]*)</a>`)

	resultMatches := resultPattern.FindAllStringSubmatch(html, maxResults*2)
	snippetMatches := snippetPattern.FindAllStringSubmatch(html, maxResults*2)

	for i, match := range resultMatches {
		if len(results) >= maxResults {
			break
		}

		if len(match) < 3 {
			continue
		}

		rawURL := match[1]
		title := cleanHTMLText(match[2])

		// DuckDuckGo wraps URLs in a redirect - extract the actual URL
		actualURL := extractDDGURL(rawURL)
		if actualURL == "" {
			continue
		}

		snippet := ""
		if i < len(snippetMatches) && len(snippetMatches[i]) > 1 {
			snippet = cleanHTMLText(snippetMatches[i][1])
		}

		results = append(results, SearchResult{
			Title:   title,
			URL:     actualURL,
			Snippet: snippet,
		})
	}

	return results
}

// extractDDGURL extracts the actual URL from DuckDuckGo's redirect URL.
func extractDDGURL(ddgURL string) string {
	// DDG URLs look like: //duckduckgo.com/l/?uddg=https%3A%2F%2Fexample.com&...
	if strings.Contains(ddgURL, "uddg=") {
		parts := strings.Split(ddgURL, "uddg=")
		if len(parts) > 1 {
			encodedURL := strings.Split(parts[1], "&")[0]
			decodedURL, err := url.QueryUnescape(encodedURL)
			if err == nil {
				return decodedURL
			}
		}
	}

	// Some results have direct URLs
	if strings.HasPrefix(ddgURL, "http") {
		return ddgURL
	}

	return ""
}

// --- web_fetch ---

func webFetchTool() Tool {
	params := NewObjectParameters()
	params.AddProperty("url", String("URL to fetch"))
	params.AddProperty("max_length", Integer("Maximum content length in characters (default: 5000)"))
	params.AddRequired("url")

	return Tool{
		Name:             "web_fetch",
		Description:      "Fetch a web page and extract its text content. HTML is converted to readable text.",
		Parameters:       params,
		Category:         CategoryWeb,
		RequiresApproval: false,
	}
}

type webFetchArgs struct {
	URL       string `json:"url"`
	MaxLength int    `json:"max_length"`
}

func webFetchHandler(ctx context.Context, args json.RawMessage) (string, error) {
	var a webFetchArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if a.URL == "" {
		return "", fmt.Errorf("url is required")
	}

	// Validate URL
	parsedURL, err := url.Parse(a.URL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", fmt.Errorf("only http and https URLs are supported")
	}

	maxLength := a.MaxLength
	if maxLength <= 0 {
		maxLength = 5000
	}
	if maxLength > 50000 {
		maxLength = 50000
	}

	content, err := fetchAndExtractText(ctx, a.URL, maxLength)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Content from: %s\n\n", a.URL))
	sb.WriteString(content)

	return sb.String(), nil
}

// fetchAndExtractText fetches a URL and extracts readable text from HTML.
func fetchAndExtractText(ctx context.Context, targetURL string, maxLength int) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch returned status %d", resp.StatusCode)
	}

	// Limit read size
	limitedReader := io.LimitReader(resp.Body, 1024*1024) // 1MB max
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/plain") {
		// Plain text - return as-is
		text := string(body)
		if len(text) > maxLength {
			text = text[:maxLength] + "\n\n... (truncated)"
		}
		return text, nil
	}

	// Convert HTML to text
	text := htmlToText(string(body))

	if len(text) > maxLength {
		text = text[:maxLength] + "\n\n... (truncated)"
	}

	return text, nil
}

// htmlToText converts HTML to readable plain text.
// This is a simple implementation - could be enhanced with a proper HTML parser.
func htmlToText(html string) string {
	// Remove script and style elements
	scriptPattern := regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	stylePattern := regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	html = scriptPattern.ReplaceAllString(html, "")
	html = stylePattern.ReplaceAllString(html, "")

	// Remove HTML comments
	commentPattern := regexp.MustCompile(`(?s)<!--.*?-->`)
	html = commentPattern.ReplaceAllString(html, "")

	// Convert block elements to newlines
	blockTags := []string{"p", "div", "br", "h1", "h2", "h3", "h4", "h5", "h6", "li", "tr", "article", "section"}
	for _, tag := range blockTags {
		openPattern := regexp.MustCompile(fmt.Sprintf(`(?i)</?%s[^>]*>`, tag))
		html = openPattern.ReplaceAllString(html, "\n")
	}

	// Handle special elements
	html = regexp.MustCompile(`(?i)<br\s*/?>|<hr[^>]*>`).ReplaceAllString(html, "\n")

	// Extract link text with URL
	linkPattern := regexp.MustCompile(`(?i)<a[^>]*href="([^"]*)"[^>]*>([^<]*)</a>`)
	html = linkPattern.ReplaceAllString(html, "$2 ($1)")

	// Remove all remaining HTML tags
	tagPattern := regexp.MustCompile(`<[^>]*>`)
	html = tagPattern.ReplaceAllString(html, "")

	// Decode common HTML entities
	html = decodeHTMLEntities(html)

	// Clean up whitespace
	html = regexp.MustCompile(`[ \t]+`).ReplaceAllString(html, " ")
	html = regexp.MustCompile(`\n\s*\n\s*`).ReplaceAllString(html, "\n\n")
	html = strings.TrimSpace(html)

	return html
}

// cleanHTMLText removes HTML tags and decodes entities from text.
func cleanHTMLText(text string) string {
	// Remove tags
	tagPattern := regexp.MustCompile(`<[^>]*>`)
	text = tagPattern.ReplaceAllString(text, "")

	// Decode entities
	text = decodeHTMLEntities(text)

	return strings.TrimSpace(text)
}

// decodeHTMLEntities decodes common HTML entities.
func decodeHTMLEntities(text string) string {
	entities := map[string]string{
		"&amp;":   "&",
		"&lt;":    "<",
		"&gt;":    ">",
		"&quot;":  "\"",
		"&apos;":  "'",
		"&#39;":   "'",
		"&nbsp;":  " ",
		"&ndash;": "\u2013",
		"&mdash;": "\u2014",
		"&lsquo;": "\u2018",
		"&rsquo;": "\u2019",
		"&ldquo;": "\u201C",
		"&rdquo;": "\u201D",
		"&bull;":  "\u2022",
		"&hellip;": "\u2026",
		"&copy;":  "\u00A9",
		"&reg;":   "\u00AE",
		"&trade;": "\u2122",
	}

	for entity, char := range entities {
		text = strings.ReplaceAll(text, entity, char)
	}

	// Decode numeric entities
	numericPattern := regexp.MustCompile(`&#(\d+);`)
	text = numericPattern.ReplaceAllStringFunc(text, func(match string) string {
		var num int
		_, _ = fmt.Sscanf(match, "&#%d;", &num)
		if num > 0 && num < 0x10FFFF {
			return string(rune(num))
		}
		return match
	})

	return text
}
