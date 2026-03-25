package deepsearch

import "strings"

// positiveWords and negativeWords are deterministic word lists for sentiment scoring.
var positiveWords = []string{
	"growth", "opportunity", "innovation", "expansion", "profit", "success",
	"advantage", "recovery", "surge", "gain", "promising", "upside", "optimistic",
	"partnership", "breakthrough", "accelerate", "demand", "bullish", "benefit",
	"positive", "improve", "increase", "strong", "resilient", "outperform",
}

var negativeWords = []string{
	"risk", "threat", "decline", "disruption", "loss", "failure", "crisis",
	"pressure", "downturn", "recession", "instability", "collapse", "default",
	"vulnerability", "uncertainty", "volatile", "bearish", "negative", "danger",
	"decrease", "weak", "underperform", "headwind", "challenge", "concern",
}

// CalculateSentimentScore returns a sentiment score in [-1.0, 1.0] by counting
// positive and negative word occurrences in the text. The score is:
//
//	score = (positive_count - negative_count) / (positive_count + negative_count + 1)
//
// A score > 0 is net positive; < 0 is net negative; 0 is neutral.
func CalculateSentimentScore(text string) float64 {
	lower := strings.ToLower(text)
	var pos, neg int
	for _, w := range positiveWords {
		pos += strings.Count(lower, w)
	}
	for _, w := range negativeWords {
		neg += strings.Count(lower, w)
	}
	total := pos + neg + 1 // +1 prevents division by zero
	return float64(pos-neg) / float64(total)
}

// AdjustStabilityBySentiment applies the approved formula:
//
//	adjusted = base * (1.0 - sentiment * 0.3)
//
// Clamped to [0.05, 0.50]. A negative sentiment score increases the modifier
// (reducing stability); a positive score decreases it (raising stability).
func AdjustStabilityBySentiment(base, sentiment float64) float64 {
	adjusted := base * (1.0 - sentiment*0.3)
	if adjusted < 0.05 {
		return 0.05
	}
	if adjusted > 0.50 {
		return 0.50
	}
	return adjusted
}
