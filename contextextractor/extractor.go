// Package contextextractor fetches and summarizes public context from URLs
// (company website, LinkedIn, Instagram, Twitter/X, Facebook, YouTube, etc.)
// to enrich FRACTURE simulations with real-world data.
package contextextractor

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// SourceType identifies what kind of URL was provided.
type SourceType string

const (
	SourceWebsite   SourceType = "website"
	SourceLinkedIn  SourceType = "linkedin"
	SourceInstagram SourceType = "instagram"
	SourceTwitter   SourceType = "twitter"
	SourceFacebook  SourceType = "facebook"
	SourceYouTube   SourceType = "youtube"
	SourceUnknown   SourceType = "unknown"
)

// ExtractedContext holds the scraped content from a single URL.
type ExtractedContext struct {
	URL         string
	SourceType  SourceType
	Title       string
	Description string
	Content     string // cleaned text content, max 2000 chars
	Error       string
}

// CompanyContext aggregates context from multiple URLs.
type CompanyContext struct {
	Sources     []*ExtractedContext
	Summary     string // combined summary for LLM injection
	HasErrors   bool
}

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("too many redirects")
		}
		return nil
	},
}

// DetectSourceType classifies a URL by its domain.
func DetectSourceType(rawURL string) SourceType {
	if rawURL == "" {
		return SourceUnknown
	}
	u, err := url.Parse(strings.ToLower(rawURL))
	if err != nil {
		return SourceUnknown
	}
	host := u.Hostname()
	if host == "" {
		return SourceUnknown
	}
	switch {
	case strings.Contains(host, "linkedin.com"):
		return SourceLinkedIn
	case strings.Contains(host, "instagram.com"):
		return SourceInstagram
	case strings.Contains(host, "twitter.com") || strings.Contains(host, "x.com"):
		return SourceTwitter
	case strings.Contains(host, "facebook.com") || strings.Contains(host, "fb.com"):
		return SourceFacebook
	case strings.Contains(host, "youtube.com") || strings.Contains(host, "youtu.be"):
		return SourceYouTube
	default:
		return SourceWebsite
	}
}

// ExtractFromURL fetches and extracts useful text from a single URL.
func ExtractFromURL(rawURL string) *ExtractedContext {
	result := &ExtractedContext{
		URL:        rawURL,
		SourceType: DetectSourceType(rawURL),
	}

	// Normalize URL
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
		result.URL = rawURL
	}

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		result.Error = fmt.Sprintf("invalid URL: %v", err)
		return result
	}

	// Mimic a real browser to avoid bot blocks
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en;q=0.8")

	resp, err := httpClient.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("fetch error: %v", err)
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		return result
	}

	// Read up to 512KB
	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		result.Error = fmt.Sprintf("read error: %v", err)
		return result
	}

	html := string(body)

	result.Title = extractTitle(html)
	result.Description = extractMetaDescription(html)
	result.Content = extractTextContent(html, result.SourceType)

	return result
}

// ExtractFromURLs fetches context from multiple URLs concurrently.
func ExtractFromURLs(urls []string) *CompanyContext {
	ctx := &CompanyContext{}

	type result struct {
		idx int
		ec  *ExtractedContext
	}

	ch := make(chan result, len(urls))

	for i, u := range urls {
		go func(idx int, rawURL string) {
			ch <- result{idx: idx, ec: ExtractFromURL(rawURL)}
		}(i, u)
	}

	sources := make([]*ExtractedContext, len(urls))
	for range urls {
		r := <-ch
		sources[r.idx] = r.ec
		if r.ec.Error != "" {
			ctx.HasErrors = true
		}
	}
	ctx.Sources = sources
	ctx.Summary = buildSummary(sources)
	return ctx
}

// buildSummary creates a structured text block for LLM injection.
func buildSummary(sources []*ExtractedContext) string {
	var sb strings.Builder
	sb.WriteString("=== COMPANY CONTEXT (extracted from public sources) ===\n\n")

	for _, s := range sources {
		if s == nil || s.Error != "" {
			continue
		}
		sb.WriteString(fmt.Sprintf("[%s] %s\n", strings.ToUpper(string(s.SourceType)), s.URL))
		if s.Title != "" {
			sb.WriteString(fmt.Sprintf("Title: %s\n", s.Title))
		}
		if s.Description != "" {
			sb.WriteString(fmt.Sprintf("Description: %s\n", s.Description))
		}
		if s.Content != "" {
			sb.WriteString(fmt.Sprintf("Content:\n%s\n", s.Content))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("=== END COMPANY CONTEXT ===\n")
	return sb.String()
}

// extractTitle pulls the <title> tag from HTML.
func extractTitle(html string) string {
	re := regexp.MustCompile(`(?i)<title[^>]*>(.*?)</title>`)
	m := re.FindStringSubmatch(html)
	if len(m) > 1 {
		return cleanText(m[1])
	}
	return ""
}

// extractMetaDescription pulls the meta description from HTML.
func extractMetaDescription(html string) string {
	re := regexp.MustCompile(`(?i)<meta[^>]+name=["']description["'][^>]+content=["']([^"']+)["']`)
	m := re.FindStringSubmatch(html)
	if len(m) > 1 {
		return cleanText(m[1])
	}
	// Try og:description
	re2 := regexp.MustCompile(`(?i)<meta[^>]+property=["']og:description["'][^>]+content=["']([^"']+)["']`)
	m2 := re2.FindStringSubmatch(html)
	if len(m2) > 1 {
		return cleanText(m2[1])
	}
	return ""
}

// extractTextContent strips HTML tags and returns clean text up to maxLen chars.
func extractTextContent(html string, srcType SourceType) string {
	// Remove scripts, styles, nav, footer
	html = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`).ReplaceAllString(html, " ")
	html = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`).ReplaceAllString(html, " ")
	html = regexp.MustCompile(`(?is)<nav[^>]*>.*?</nav>`).ReplaceAllString(html, " ")
	html = regexp.MustCompile(`(?is)<footer[^>]*>.*?</footer>`).ReplaceAllString(html, " ")
	html = regexp.MustCompile(`(?is)<header[^>]*>.*?</header>`).ReplaceAllString(html, " ")

	// For social networks, focus on bio/about sections
	if srcType == SourceLinkedIn {
		// Try to extract about/description section
		re := regexp.MustCompile(`(?is)about[^<]*</[^>]+>(.*?)(experience|education|skills)`)
		m := re.FindStringSubmatch(html)
		if len(m) > 1 {
			html = m[1]
		}
	}

	// Strip all remaining HTML tags
	html = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(html, " ")

	// Decode common HTML entities
	html = strings.ReplaceAll(html, "&amp;", "&")
	html = strings.ReplaceAll(html, "&lt;", "<")
	html = strings.ReplaceAll(html, "&gt;", ">")
	html = strings.ReplaceAll(html, "&quot;", "\"")
	html = strings.ReplaceAll(html, "&#39;", "'")
	html = strings.ReplaceAll(html, "&nbsp;", " ")

	// Collapse whitespace
	html = regexp.MustCompile(`\s+`).ReplaceAllString(html, " ")
	html = strings.TrimSpace(html)

	// Truncate to 2000 chars (safe for LLM context)
	maxLen := 2000
	if utf8.RuneCountInString(html) > maxLen {
		runes := []rune(html)
		html = string(runes[:maxLen]) + "..."
	}

	return html
}

// cleanText removes extra whitespace and HTML entities from a string.
func cleanText(s string) string {
	s = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&#39;", "'")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	return strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(s, " "))
}
