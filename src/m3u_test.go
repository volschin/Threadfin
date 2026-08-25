package src

import "testing"

func TestParseFrameRateRejectsMalformedRatio(t *testing.T) {
	if got := parseFrameRate([]string{"not-a-number", "also-invalid"}); got != 0 {
		t.Fatalf("parseFrameRate() = %d, want 0 for malformed ratio", got)
	}
}

func TestParseFrameRateRejectsMissingDenominator(t *testing.T) {
	if got := parseFrameRate([]string{"25"}); got != 0 {
		t.Fatalf("parseFrameRate() = %d, want 0 for incomplete ratio", got)
	}
}
