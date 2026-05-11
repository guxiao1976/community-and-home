package engine

// ModerationResult is the output of a content review check
type ModerationResult struct {
	Pass       bool           `json:"pass"`
	RiskLevel  string         `json:"risk_level"`
	Reason     string         `json:"reason"`
	NeedReview bool           `json:"need_review"`
	Details    []MatchDetail  `json:"details"`
}

// MatchDetail describes a single match found during review
type MatchDetail struct {
	Layer       string  `json:"layer"`
	MatchedText string  `json:"matched_text,omitempty"`
	Category    string  `json:"category,omitempty"`
	Severity    int     `json:"severity,omitempty"`
	Confidence  float64 `json:"confidence,omitempty"`
}
