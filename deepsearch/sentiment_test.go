package deepsearch

import (
	"math"
	"testing"
)

func TestCalculateSentimentScore(t *testing.T) {
	tests := []struct {
		name  string
		input string
		// wantSign: +1 positive, -1 negative, 0 neutral (±0.1 tolerance)
		wantSign int
	}{
		{
			name:     "clearly positive text",
			input:    "strong growth opportunity innovation profit success benefit increase resilient",
			wantSign: 1,
		},
		{
			name:     "clearly negative text",
			input:    "risk threat decline loss failure crisis pressure recession collapse default",
			wantSign: -1,
		},
		{
			name:     "empty string is neutral",
			input:    "",
			wantSign: 0,
		},
		{
			name:     "unrelated words are neutral",
			input:    "the cat sat on the mat yesterday morning",
			wantSign: 0,
		},
		{
			name:     "balanced text approaches zero",
			input:    "growth risk opportunity threat profit loss",
			wantSign: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			score := CalculateSentimentScore(tc.input)

			if score < -1.0 || score > 1.0 {
				t.Errorf("score %f out of range [-1, 1]", score)
			}

			switch tc.wantSign {
			case 1:
				if score <= 0 {
					t.Errorf("expected positive score, got %f", score)
				}
			case -1:
				if score >= 0 {
					t.Errorf("expected negative score, got %f", score)
				}
			case 0:
				if math.Abs(score) > 0.3 {
					t.Errorf("expected near-zero score, got %f", score)
				}
			}
		})
	}
}

func TestCalculateSentimentScoreDeterministic(t *testing.T) {
	input := "market growth risk opportunity disruption"
	first := CalculateSentimentScore(input)
	for i := 0; i < 5; i++ {
		if got := CalculateSentimentScore(input); got != first {
			t.Errorf("non-deterministic: call %d got %f, want %f", i+1, got, first)
		}
	}
}

func TestAdjustStabilityBySentiment(t *testing.T) {
	tests := []struct {
		name      string
		base      float64
		sentiment float64
		wantMin   float64
		wantMax   float64
	}{
		{
			name:      "neutral sentiment — output equals base (clamped to [0.05,0.50])",
			base:      0.30,
			sentiment: 0.0,
			wantMin:   0.30,
			wantMax:   0.30,
		},
		{
			name:      "positive sentiment — reduces modifier (stability improves)",
			base:      0.30,
			sentiment: 1.0,
			// adjusted = 0.30 * (1.0 - 1.0*0.3) = 0.30 * 0.7 = 0.21
			wantMin: 0.20,
			wantMax: 0.22,
		},
		{
			name:      "negative sentiment — increases modifier (stability decreases)",
			base:      0.30,
			sentiment: -1.0,
			// adjusted = 0.30 * (1.0 - (-1.0)*0.3) = 0.30 * 1.3 = 0.39
			wantMin: 0.38,
			wantMax: 0.40,
		},
		{
			name:      "clamp lower bound — result never below 0.05",
			base:      0.01,
			sentiment: -1.0,
			wantMin:   0.05,
			wantMax:   0.05,
		},
		{
			name:      "clamp upper bound — result never above 0.50",
			base:      0.99,
			sentiment: -1.0,
			wantMin:   0.50,
			wantMax:   0.50,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := AdjustStabilityBySentiment(tc.base, tc.sentiment)
			if got < tc.wantMin-1e-9 || got > tc.wantMax+1e-9 {
				t.Errorf("AdjustStabilityBySentiment(%f, %f) = %f, want [%f, %f]",
					tc.base, tc.sentiment, got, tc.wantMin, tc.wantMax)
			}
		})
	}
}
