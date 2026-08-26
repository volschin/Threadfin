package src

import "testing"

var benchmarkFilterResult bool

func BenchmarkFilterConditions(b *testing.B) {
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
		b.Run(tc.name, func(b *testing.B) {
			if got := checkConditions(tc.streamValues, tc.conditions, tc.conditionType); got != tc.want {
				b.Fatalf("fixture result = %t, want %t", got, tc.want)
			}
			b.ReportAllocs()
			for b.Loop() {
				benchmarkFilterResult = checkConditions(tc.streamValues, tc.conditions, tc.conditionType)
			}
		})
	}
}
