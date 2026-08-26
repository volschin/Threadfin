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

func TestParseGroupCountLabel(t *testing.T) {
	tests := []struct {
		name      string
		label     string
		wantName  string
		wantCount int
		wantOK    bool
	}{
		{name: "ordinary label", label: "Sports (12)", wantName: "Sports", wantCount: 12, wantOK: true},
		{name: "final opening delimiter", label: "Sports (East) (7)", wantName: "Sports (East)", wantCount: 7, wantOK: true},
		{name: "suffix after closing delimiter", label: "News (3) ignored", wantName: "News", wantCount: 3, wantOK: true},
		{name: "opening delimiter after final close", label: "News (3) trailing (bad", wantOK: false},
		{name: "empty group", label: " (3)", wantOK: false},
		{name: "empty count", label: "News ()", wantOK: false},
		{name: "nonnumeric count", label: "News (many)", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotCount, gotOK := parseGroupCountLabel(tt.label)
			if gotName != tt.wantName || gotCount != tt.wantCount || gotOK != tt.wantOK {
				t.Fatalf(
					"parseGroupCountLabel(%q) = (%q, %d, %v), want (%q, %d, %v)",
					tt.label, gotName, gotCount, gotOK, tt.wantName, tt.wantCount, tt.wantOK,
				)
			}
		})
	}
}

func TestCheckConditionsSeparatorCompatibility(t *testing.T) {
	cases := []struct {
		name          string
		streamValues  string
		conditions    string
		conditionType string
		want          bool
	}{
		{name: "EarlyMatch", streamValues: "sports premium hd", conditions: "sports,news,documentary,movies", conditionType: "include", want: true},
		{name: "LateMatch", streamValues: "sports premium hd", conditions: "news,documentary,movies,sports", conditionType: "include", want: true},
		{name: "NoMatch", streamValues: "sports premium hd", conditions: "news,documentary,movies,kids", conditionType: "include", want: false},
		{name: "EmptyElements", streamValues: "", conditions: "", conditionType: "include", want: true},
		{name: "RepeatedSeparators", streamValues: "documentary", conditions: "sports,,,news", conditionType: "include", want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := checkConditions(tc.streamValues, tc.conditions, tc.conditionType); got != tc.want {
				t.Fatalf("checkConditions() = %t, want %t", got, tc.want)
			}
		})
	}
}
