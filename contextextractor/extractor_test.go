package contextextractor

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	xhtml "golang.org/x/net/html"
)

// ─── DetectSourceType ─────────────────────────────────────────────────────────

func TestDetectSourceType(t *testing.T) {
	tests := []struct {
		url  string
		want SourceType
	}{
		{"https://linkedin.com/company/fracture", SourceLinkedIn},
		{"https://www.linkedin.com/in/johndoe", SourceLinkedIn},
		{"https://instagram.com/fracture_app", SourceInstagram},
		{"https://www.instagram.com/p/abc123", SourceInstagram},
		{"https://twitter.com/fracture", SourceTwitter},
		{"https://x.com/fracture", SourceTwitter},
		{"https://facebook.com/fracture", SourceFacebook},
		{"https://www.facebook.com/fracture", SourceFacebook},
		{"https://youtube.com/c/fracture", SourceYouTube},
		{"https://www.youtube.com/@fracture", SourceYouTube},
		{"https://fracture.io", SourceWebsite},
		{"https://minha-empresa.com.br", SourceWebsite},
		{"not-a-url", SourceUnknown},
		{"", SourceUnknown},
	}
	for _, tt := range tests {
		got := DetectSourceType(tt.url)
		if got != tt.want {
			t.Errorf("DetectSourceType(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}

// ─── CompanyContext summary ───────────────────────────────────────────────────

func TestCompanyContextSummary(t *testing.T) {
	ctx := &CompanyContext{
		Sources: []*ExtractedContext{
			{
				URL:         "https://example.com",
				SourceType:  SourceWebsite,
				Title:       "Example Company",
				Description: "We build great products",
				Content:     "Example Company builds innovative software solutions for enterprise clients.",
			},
			{
				URL:        "https://linkedin.com/company/example",
				SourceType: SourceLinkedIn,
				Content:    "Example Company — 500 employees — Software & Technology",
			},
		},
	}
	ctx.Summary = buildSummary(ctx.Sources)
	if ctx.Summary == "" {
		t.Error("Summary must not be empty when sources are present")
	}
	if !strings.Contains(ctx.Summary, "Example Company") {
		t.Error("Summary should contain company name from source content")
	}
}

func TestExtractedContextError(t *testing.T) {
	ec := &ExtractedContext{
		URL:   "https://example.com",
		Error: "connection refused",
	}
	if ec.Error == "" {
		t.Error("Error field should be set")
	}
}

// ─── HTML node extraction ─────────────────────────────────────────────────────

func TestExtractTitleNode(t *testing.T) {
	raw := `<html><head><title>ACME Corp — Home</title></head><body></body></html>`
	doc, err := xhtml.Parse(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("parse HTML: %v", err)
	}
	got := extractTitleNode(doc)
	if got != "ACME Corp — Home" {
		t.Errorf("expected 'ACME Corp — Home', got %q", got)
	}
}

func TestExtractMetaDescription(t *testing.T) {
	raw := `<html><head><meta name="description" content="Best gym in town."></head><body></body></html>`
	doc, _ := xhtml.Parse(strings.NewReader(raw))
	got := extractMetaNode(doc)
	if got != "Best gym in town." {
		t.Errorf("expected description, got %q", got)
	}
}

func TestExtractMetaOGDescription(t *testing.T) {
	raw := `<html><head><meta property="og:description" content="Open Graph description."></head><body></body></html>`
	doc, _ := xhtml.Parse(strings.NewReader(raw))
	got := extractMetaNode(doc)
	if got != "Open Graph description." {
		t.Errorf("expected og:description, got %q", got)
	}
}

func TestExtractBodyTextSkipsScripts(t *testing.T) {
	raw := `<html><body>
		<script>var x = 1;</script>
		<p>This is the real content of the page for testing purposes.</p>
		<style>.foo{color:red}</style>
	</body></html>`
	doc, _ := xhtml.Parse(strings.NewReader(raw))
	got := extractBodyText(doc, SourceWebsite)
	if strings.Contains(got, "var x") {
		t.Error("script content should be stripped")
	}
	if !strings.Contains(got, "real content") {
		t.Errorf("expected body text, got %q", got)
	}
}

func TestExtractBodyTextTruncates(t *testing.T) {
	longText := strings.Repeat("word ", 1000) // 5000 chars
	raw := fmt.Sprintf(`<html><body><p>%s</p></body></html>`, longText)
	doc, _ := xhtml.Parse(strings.NewReader(raw))
	got := extractBodyText(doc, SourceWebsite)
	if len([]rune(got)) > maxContentRunes+3 {
		t.Errorf("content should be truncated to %d runes, got %d", maxContentRunes, len([]rune(got)))
	}
	if !strings.HasSuffix(got, "...") {
		t.Error("truncated content should end with '...'")
	}
}

func TestTruncateRunes(t *testing.T) {
	s := "hello world"
	if truncateRunes(s, 100) != s {
		t.Error("short string should not be truncated")
	}
	got := truncateRunes(s, 5)
	if got != "hello..." {
		t.Errorf("expected 'hello...', got %q", got)
	}
}

// ─── ExtractFromURL with test HTTP server ─────────────────────────────────────

func TestExtractFromURL_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<html>
			<head>
				<title>ACME Fitness</title>
				<meta name="description" content="Best gym in town.">
			</head>
			<body>
				<p>We offer personal training and group classes for all fitness levels in the city.</p>
			</body>
		</html>`)
	}))
	defer srv.Close()

	result := ExtractFromURL(srv.URL)
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Title != "ACME Fitness" {
		t.Errorf("expected title 'ACME Fitness', got %q", result.Title)
	}
	if result.Description != "Best gym in town." {
		t.Errorf("expected description, got %q", result.Description)
	}
	if !strings.Contains(result.Content, "personal training") {
		t.Errorf("expected content to contain 'personal training', got %q", result.Content)
	}
}

func TestExtractFromURL_HTTP404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	result := ExtractFromURL(srv.URL)
	if result.Error == "" {
		t.Error("expected error for 404, got none")
	}
	if !strings.Contains(result.Error, "404") {
		t.Errorf("expected 404 in error, got %q", result.Error)
	}
}

func TestExtractFromURL_NormalizeHTTPS(t *testing.T) {
	// URL without scheme should be prefixed with https://
	result := ExtractFromURL("example.com/nonexistent-path-xyz-abc")
	if !strings.HasPrefix(result.URL, "https://") {
		t.Errorf("expected https:// prefix, got %q", result.URL)
	}
}

func TestExtractFromURL_InvalidURL(t *testing.T) {
	result := ExtractFromURL("://invalid-url")
	if result.Error == "" {
		t.Error("expected error for invalid URL")
	}
}

func TestExtractFromURL_SourceTypeSet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<html><body><p>test</p></body></html>`)
	}))
	defer srv.Close()

	result := ExtractFromURL(srv.URL)
	// Test server URLs are plain http, so SourceType should be website
	if result.SourceType != SourceWebsite {
		t.Errorf("expected SourceWebsite, got %q", result.SourceType)
	}
}

// ─── ExtractFromURLs concurrency ─────────────────────────────────────────────

func TestExtractFromURLs_Concurrent(t *testing.T) {
	var servers []*httptest.Server
	var urls []string

	for i := 0; i < 5; i++ {
		i := i
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `<html><head><title>Server %d</title></head><body><p>Content for server number %d here.</p></body></html>`, i, i)
		}))
		servers = append(servers, srv)
		urls = append(urls, srv.URL)
	}
	defer func() {
		for _, s := range servers {
			s.Close()
		}
	}()

	ctx := ExtractFromURLs(urls)
	if len(ctx.Sources) != 5 {
		t.Errorf("expected 5 sources, got %d", len(ctx.Sources))
	}
	for i, s := range ctx.Sources {
		if s == nil {
			t.Errorf("source %d is nil", i)
			continue
		}
		if s.Error != "" {
			t.Errorf("source %d has error: %s", i, s.Error)
		}
	}
	if ctx.Summary == "" {
		t.Error("expected non-empty summary")
	}
}

func TestExtractFromURLs_RaceCondition(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<html><body><p>test content for race detection</p></body></html>`)
	}))
	defer srv.Close()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ExtractFromURLs([]string{srv.URL, srv.URL})
		}()
	}
	wg.Wait()
}

func TestExtractFromURLs_MixedResults(t *testing.T) {
	good := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<html><head><title>Good</title></head><body><p>Good content here for testing.</p></body></html>`)
	}))
	defer good.Close()

	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer bad.Close()

	ctx := ExtractFromURLs([]string{good.URL, bad.URL})
	if !ctx.HasErrors {
		t.Error("expected HasErrors=true when one URL fails")
	}
	if len(ctx.Sources) != 2 {
		t.Errorf("expected 2 sources, got %d", len(ctx.Sources))
	}
}

// ─── buildSummary ─────────────────────────────────────────────────────────────

func TestBuildSummaryWithLimitedSource(t *testing.T) {
	sources := []*ExtractedContext{
		{URL: "https://acme.com", SourceType: SourceWebsite, Title: "ACME", Content: "We build things for our customers."},
		{URL: "https://linkedin.com/acme", SourceType: SourceLinkedIn, Error: "blocked", Limited: true, Content: "[limited]"},
	}
	summary := buildSummary(sources)
	if !strings.Contains(summary, "ACME") {
		t.Error("summary should contain title")
	}
	if !strings.Contains(summary, "[LIMITED]") {
		t.Error("summary should mark limited sources")
	}
	if !strings.Contains(summary, "=== COMPANY CONTEXT") {
		t.Error("summary should have header")
	}
	if !strings.Contains(summary, "=== END COMPANY CONTEXT") {
		t.Error("summary should have footer")
	}
}

func TestBuildSummaryEmpty(t *testing.T) {
	summary := buildSummary(nil)
	if !strings.Contains(summary, "=== COMPANY CONTEXT") {
		t.Error("empty summary should still have header")
	}
}

// ─── cleanRawText fallback ────────────────────────────────────────────────────

func TestCleanRawText(t *testing.T) {
	raw := `<html><head><script>alert(1)</script></head><body><p>Hello &amp; world</p></body></html>`
	got := cleanRawText(raw)
	if strings.Contains(got, "<") || strings.Contains(got, ">") {
		t.Errorf("cleanRawText should strip all tags, got %q", got)
	}
	if !strings.Contains(got, "Hello & world") {
		t.Errorf("expected decoded entities, got %q", got)
	}
	if strings.Contains(got, "alert") {
		t.Error("script content should be stripped")
	}
}
