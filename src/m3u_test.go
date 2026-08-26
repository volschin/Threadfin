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

func TestCompareChannelNumbers(t *testing.T) {
	tests := []struct {
		name  string
		left  string
		right string
		want  int
	}{
		{name: "numeric order", left: "2", right: "10", want: -1},
		{name: "numeric equality", left: "2", right: "2.0", want: 0},
		{name: "number before text", left: "2", right: "alpha", want: -1},
		{name: "text after number", left: "alpha", right: "2", want: 1},
		{name: "lexical order", left: "alpha", right: "beta", want: -1},
		{name: "lexical equality", left: "alpha", right: "alpha", want: 0},
		{name: "NaN retains numeric equivalence", left: "NaN", right: "2", want: 0},
		{name: "NaN is parsed before text", left: "NaN", right: "alpha", want: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareChannelNumbers(
				XEPGChannelStruct{TvgChno: tt.left},
				XEPGChannelStruct{TvgChno: tt.right},
			)
			if got != tt.want {
				t.Fatalf("compareChannelNumbers(%q, %q) = %d, want %d", tt.left, tt.right, got, tt.want)
			}

			reverse := compareChannelNumbers(
				XEPGChannelStruct{TvgChno: tt.right},
				XEPGChannelStruct{TvgChno: tt.left},
			)
			if reverse != -tt.want {
				t.Fatalf("compareChannelNumbers(%q, %q) = %d, want %d", tt.right, tt.left, reverse, -tt.want)
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
