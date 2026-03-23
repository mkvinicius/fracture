// Package contextextractor fetches and summarizes public context from URLs
// (company website, LinkedIn, Instagram, Twitter/X, Facebook, YouTube, etc.)
// to enrich FRACTURE simulations with real-world data.
//
// Robustness improvements over v1.4.x:
//   - Per-URL context with configurable timeout (default 12s)
//   - Retry with exponential backoff (up to 2 retries) for transient errors
//   - Structured HTML parsing via golang.org/x/net/html (no more regex-only stripping)
//   - Graceful degradation: errors are surfaced in ExtractedContext.Error, never panic
//   - Blocked/dynamic pages (LinkedIn, Instagram, X) are flagged as "limited" with
//     a clear message so the LLM knows the context may be incomplete
//   - HTML entity decoding via html.UnescapeString (stdlib)
package contextextractor

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	xhtml "golang.org/x/net/html"
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

// socialNetworks lists platforms that typically block server-side scraping.
// For these, we still attempt extraction but mark partial results clearly.
var socialNetworks = map[SourceType]bool{
	SourceLinkedIn:  true,
	SourceInstagram: true,
	SourceTwitter:   true,
	SourceFacebook:  true,
}

// ExtractedContext holds the scraped content from a single URL.
type ExtractedContext struct {
	URL         string
	SourceType  SourceType
	Title       string
	Description string
	Content     string // cleaned text content, max 2000 chars
	Error       string
	// Limited is true when the source is a social network that may have
	// returned a login wall or minimal content instead of the real page.
	Limited bool
}

// CompanyContext aggregates context from multiple URLs.
type CompanyContext struct {
	Sources   []*ExtractedContext
	Summary   string // combined summary for LLM injection
	HasErrors bool
}

// perURLTimeout is the maximum time allowed for a single URL fetch (including retries).
const perURLTimeout = 12 * time.Second

// maxBodyBytes is the maximum response body size we read (512 KB).
const maxBodyBytes = 512 * 1024

// maxContentRunes is the maximum content length injected into the LLM prompt.
const maxContentRunes = 2000

// httpTransport is a shared transport with reasonable timeouts.
var httpTransport = &http.Transport{
	ResponseHeaderTimeout: 8 * time.Second,
	DisableKeepAlives:     false,
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
// It retries up to 2 times on transient network errors with exponential backoff.
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

	ctx, cancel := context.WithTimeout(context.Background(), perURLTimeout)
	defer cancel()

	var (
		body []byte
		err  error
	)

	// Retry loop: up to 3 attempts (1 initial + 2 retries) for transient errors.
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 500ms, 1s
			select {
			case <-ctx.Done():
				result.Error = "timeout waiting for retry"
				return result
			case <-time.After(time.Duration(attempt*500) * time.Millisecond):
			}
		}

		body, err = fetchURL(ctx, rawURL)
		if err == nil {
			break
		}
		// Only retry on network/timeout errors, not on HTTP 4xx/5xx
		if isHTTPError(err) {
			break
		}
	}

	if err != nil {
		result.Error = err.Error()
		// For social networks, surface a clear "limited" message instead of just an error
		if socialNetworks[result.SourceType] {
			result.Limited = true
			result.Error = fmt.Sprintf("blocked or login-required (%s) — partial context only", result.SourceType)
		}
		return result
	}

	// Parse HTML with golang.org/x/net/html for structured extraction
	doc, parseErr := xhtml.Parse(strings.NewReader(string(body)))
	if parseErr != nil {
		// Fallback to raw text extraction if HTML parse fails
		result.Content = truncateRunes(cleanRawText(string(body)), maxContentRunes)
		return result
	}

	result.Title = extractTitleNode(doc)
	result.Description = extractMetaNode(doc)
	result.Content = extractBodyText(doc, result.SourceType)

	// Flag social networks that returned suspiciously little content
	if socialNetworks[result.SourceType] && utf8.RuneCountInString(result.Content) < 100 {
		result.Limited = true
		if result.Content == "" {
			result.Content = fmt.Sprintf("[%s page returned minimal content — likely a login wall or bot block. Use the URL as context only.]", result.SourceType)
		}
	}

	return result
}

// ExtractFromURLs fetches context from multiple URLs concurrently.
// Each URL gets its own per-URL timeout; the overall call is bounded by the
// longest individual fetch (they run in parallel).
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
		if s == nil {
			continue
		}
		label := strings.ToUpper(string(s.SourceType))
		if s.Limited {
			label += " [LIMITED]"
		}
		sb.WriteString(fmt.Sprintf("[%s] %s\n", label, s.URL))
		if s.Error != "" && !s.Limited {
			sb.WriteString(fmt.Sprintf("Error: %s\n\n", s.Error))
			continue
		}
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

// ─── Internal helpers ─────────────────────────────────────────────────────────

// httpError wraps an HTTP status error so we can distinguish it from network errors.
type httpError struct{ code int }

func (e *httpError) Error() string { return fmt.Sprintf("HTTP %d", e.code) }

func isHTTPError(err error) bool {
	_, ok := err.(*httpError)
	return ok
}

// fetchURL performs a single HTTP GET with browser-like headers.
func fetchURL(ctx context.Context, rawURL string) ([]byte, error) {
	client := &http.Client{
		Transport: httpTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Rotate User-Agent strings to reduce bot detection
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, &httpError{resp.StatusCode}
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}
	return body, nil
}

// extractTitleNode extracts the <title> text from a parsed HTML tree.
func extractTitleNode(doc *xhtml.Node) string {
	var title string
	var walk func(*xhtml.Node)
	walk = func(n *xhtml.Node) {
		if n.Type == xhtml.ElementNode && n.Data == "title" && n.FirstChild != nil {
			title = html.UnescapeString(strings.TrimSpace(n.FirstChild.Data))
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return title
}

// extractMetaNode extracts description from <meta name="description"> or og:description.
func extractMetaNode(doc *xhtml.Node) string {
	var desc string
	var walk func(*xhtml.Node)
	walk = func(n *xhtml.Node) {
		if n.Type == xhtml.ElementNode && n.Data == "meta" {
			var name, property, content string
			for _, a := range n.Attr {
				switch strings.ToLower(a.Key) {
				case "name":
					name = strings.ToLower(a.Val)
				case "property":
					property = strings.ToLower(a.Val)
				case "content":
					content = a.Val
				}
			}
			if (name == "description" || property == "og:description") && content != "" {
				if desc == "" {
					desc = html.UnescapeString(strings.TrimSpace(content))
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return desc
}

// skipTags are HTML elements whose content should be ignored during text extraction.
var skipTags = map[string]bool{
	"script": true, "style": true, "nav": true, "footer": true,
	"header": true, "noscript": true, "iframe": true, "svg": true,
	"form": true, "button": true, "input": true, "select": true,
}

// contentTags are elements that are likely to contain meaningful body text.
var contentTags = map[string]bool{
	"p": true, "h1": true, "h2": true, "h3": true, "h4": true, "h5": true,
	"li": true, "td": true, "th": true, "blockquote": true, "article": true,
	"section": true, "main": true, "div": true, "span": true,
}

// extractBodyText walks the HTML tree and collects visible text from content nodes.
// For social networks it focuses on bio/about sections.
func extractBodyText(doc *xhtml.Node, srcType SourceType) string {
	var parts []string

	var walk func(*xhtml.Node, bool)
	walk = func(n *xhtml.Node, collect bool) {
		if n.Type == xhtml.ElementNode {
			tag := strings.ToLower(n.Data)
			if skipTags[tag] {
				return // skip entire subtree
			}
			// For LinkedIn/social, prioritise sections that mention "about" or "description"
			if socialNetworks[srcType] {
				for _, a := range n.Attr {
					v := strings.ToLower(a.Val)
					if strings.Contains(v, "about") || strings.Contains(v, "description") ||
						strings.Contains(v, "summary") || strings.Contains(v, "bio") {
						collect = true
					}
				}
			}
			if contentTags[tag] {
				collect = true
			}
		}

		if n.Type == xhtml.TextNode && collect {
			t := strings.TrimSpace(html.UnescapeString(n.Data))
			if len(t) > 20 { // skip noise tokens
				parts = append(parts, t)
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, collect)
		}
	}

	walk(doc, false)

	text := strings.Join(parts, " ")
	// Collapse repeated whitespace
	text = strings.Join(strings.Fields(text), " ")
	return truncateRunes(text, maxContentRunes)
}

// cleanRawText is a fallback for when HTML parsing fails — strips obvious tags.
func cleanRawText(raw string) string {
	// Remove common noise tags inline
	for _, tag := range []string{"script", "style", "nav", "footer", "header"} {
		for {
			open := strings.Index(strings.ToLower(raw), "<"+tag)
			close := strings.Index(strings.ToLower(raw), "</"+tag+">")
			if open < 0 || close < 0 || close < open {
				break
			}
			raw = raw[:open] + " " + raw[close+len("</"+tag+">"):]
		}
	}
	// Strip remaining tags
	var sb strings.Builder
	inTag := false
	for _, r := range raw {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
			sb.WriteRune(' ')
		case !inTag:
			sb.WriteRune(r)
		}
	}
	return html.UnescapeString(strings.Join(strings.Fields(sb.String()), " "))
}

// truncateRunes truncates a string to at most maxLen Unicode code points.
func truncateRunes(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxLen]) + "..."
}
