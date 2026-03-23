package contextextractor

import (
	"strings"
	"testing"
)

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
