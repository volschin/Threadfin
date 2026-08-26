package src

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCreateFilterRulesSortsIDsAndResetsPerRuleState(t *testing.T) {
	restorePersistentState(t)

	Settings = SettingsStruct{}
	Data = DataStruct{}
	Settings.Filter = map[int64]interface{}{
		30: FilterStruct{Type: "custom-filter", Filter: "third", CaseSensitive: true},
		10: FilterStruct{Type: "group-title", Filter: "News", LiveEvent: true},
		20: FilterStruct{Type: "custom-filter", Filter: "second"},
	}

	want := []Filter{
		{LiveEvent: true, Rule: "News", Type: "group-title"},
		{Rule: "second", Type: "custom-filter"},
		{CaseSensitive: true, Rule: "third", Type: "custom-filter"},
	}

	for build := 0; build < 32; build++ {
		if err := createFilterRules(); err != nil {
			t.Fatalf("createFilterRules() build %d error = %v", build, err)
		}
		if !reflect.DeepEqual(Data.Filter, want) {
			t.Fatalf("createFilterRules() build %d filters = %#v, want %#v", build, Data.Filter, want)
		}
	}
}

func filterTestStream(group, name, values string) map[string]string {
	return map[string]string{
		"group-title": group,
		"name":        name,
		"_values":     values,
	}
}

func TestFilterThisStreamUsesFirstAcceptingFilter(t *testing.T) {
	restorePersistentState(t)

	tests := []struct {
		name          string
		filters       []Filter
		stream        map[string]string
		wantStatus    bool
		wantLiveEvent bool
	}{
		{
			name: "include condition failure falls through",
			filters: []Filter{
				{Type: "custom-filter", Rule: "{ota}", LiveEvent: true},
				{Type: "group-title", Rule: "ErsatzTV"},
			},
			stream:     filterTestStream("ErsatzTV", "Ersatz News", `tvg-id="ersatz-1"`),
			wantStatus: true,
		},
		{
			name: "exclude condition failure falls through",
			filters: []Filter{
				{Type: "custom-filter", Rule: "!{blocked}", LiveEvent: true},
				{Type: "group-title", Rule: "ErsatzTV"},
			},
			stream:     filterTestStream("ErsatzTV", "Ersatz News", `tvg-id="ersatz-1" blocked`),
			wantStatus: true,
		},
		{
			name: "declined filter cannot leak live event",
			filters: []Filter{
				{Type: "custom-filter", Rule: "{missing}", LiveEvent: true},
			},
			stream: filterTestStream("ErsatzTV", "Ersatz News", `tvg-id="ersatz-1"`),
		},
		{
			name: "first passing filter wins",
			filters: []Filter{
				{Type: "group-title", Rule: "ErsatzTV", LiveEvent: true},
				{Type: "group-title", Rule: "ErsatzTV"},
			},
			stream:        filterTestStream("ErsatzTV", "Ersatz News", `tvg-id="ersatz-1"`),
			wantStatus:    true,
			wantLiveEvent: true,
		},
		{
			name: "later accepting filter supplies live event",
			filters: []Filter{
				{Type: "custom-filter", Rule: "{missing}"},
				{Type: "group-title", Rule: "ErsatzTV", LiveEvent: true},
			},
			stream:        filterTestStream("ErsatzTV", "Ersatz News", `tvg-id="ersatz-1"`),
			wantStatus:    true,
			wantLiveEvent: true,
		},
		{
			name:          "conditions-only include accepts matching stream",
			filters:       []Filter{{Type: "custom-filter", Rule: "{ersatz}", LiveEvent: true}},
			stream:        filterTestStream("ErsatzTV", "Ersatz News", `tvg-id="ersatz-1"`),
			wantStatus:    true,
			wantLiveEvent: true,
		},
		{
			name:    "conditions-only include declines nonmatching stream",
			filters: []Filter{{Type: "custom-filter", Rule: "{ota}"}},
			stream:  filterTestStream("ErsatzTV", "Ersatz News", `tvg-id="ersatz-1"`),
		},
		{
			name:          "conditions-only exclude accepts stream without excluded value",
			filters:       []Filter{{Type: "custom-filter", Rule: "!{blocked}", LiveEvent: true}},
			stream:        filterTestStream("ErsatzTV", "Ersatz News", `tvg-id="ersatz-1"`),
			wantStatus:    true,
			wantLiveEvent: true,
		},
		{
			name:    "genuinely empty custom filter remains ignored",
			filters: []Filter{{Type: "custom-filter", Rule: "", LiveEvent: true}},
			stream:  filterTestStream("ErsatzTV", "Ersatz News", `tvg-id="ersatz-1"`),
		},
		{
			name:       "case-insensitive custom rule still matches",
			filters:    []Filter{{Type: "custom-filter", Rule: "SPORTS"}},
			stream:     filterTestStream("Sports", "Sports News", `tvg-id="sports-1"`),
			wantStatus: true,
		},
		{
			name:    "case-sensitive custom rule still declines different case",
			filters: []Filter{{Type: "custom-filter", Rule: "SPORTS", CaseSensitive: true}},
			stream:  filterTestStream("Sports", "Sports News", `tvg-id="sports-1"`),
		},
		{
			name:       "case-insensitive group rule still matches",
			filters:    []Filter{{Type: "group-title", Rule: "SPORTS"}},
			stream:     filterTestStream("Sports", "Sports News", `tvg-id="sports-1"`),
			wantStatus: true,
		},
		{
			name:    "case-sensitive group rule still declines different case",
			filters: []Filter{{Type: "group-title", Rule: "SPORTS", CaseSensitive: true}},
			stream:  filterTestStream("Sports", "Sports News", `tvg-id="sports-1"`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			Data.Filter = tc.filters
			gotStatus, gotLiveEvent := filterThisStream(tc.stream)
			if gotStatus != tc.wantStatus || gotLiveEvent != tc.wantLiveEvent {
				t.Fatalf("filterThisStream() = (%t, %t), want (%t, %t)", gotStatus, gotLiveEvent, tc.wantStatus, tc.wantLiveEvent)
			}
		})
	}
}

func TestBuildDatabaseDVRProducesStableFilterSets(t *testing.T) {
	restorePersistentState(t)

	root := t.TempDir()
	System = SystemStruct{}
	System.Flag.Info = true
	System.Folder.Data = root + string(os.PathSeparator)
	System.File.Settings = filepath.Join(root, "settings.json")
	System.UnfilteredChannelLimit = 100
	System.PlexChannelLimit = 100

	Settings = SettingsStruct{}
	Settings.Files.M3U = map[string]interface{}{
		"provider": map[string]interface{}{"name": "Filter Fixture"},
	}
	Settings.Filter = map[int64]interface{}{
		10: FilterStruct{Type: "custom-filter", Filter: "{ota-}"},
		20: FilterStruct{Type: "group-title", Filter: "ErsatzTV"},
	}

	playlist := `#EXTM3U
#EXTINF:-1 tvg-id="ota-16" tvg-name="OTA 16" group-title="OTA",OTA 16
https://example.test/ota
#EXTINF:-1 tvg-id="ersatz-1" tvg-name="Ersatz News" group-title="ErsatzTV",Ersatz News
https://example.test/ersatz
`
	if err := os.WriteFile(filepath.Join(root, "provider.m3u"), []byte(playlist), 0600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	wantActive := []string{"Ersatz News [ErsatzTV]", "OTA 16 [OTA]"}
	for build := 0; build < 32; build++ {
		if err := buildDatabaseDVR(); err != nil {
			t.Fatalf("buildDatabaseDVR() build %d error = %v", build, err)
		}
		if !reflect.DeepEqual(Data.StreamPreviewUI.Active, wantActive) {
			t.Fatalf("buildDatabaseDVR() build %d active = %#v, want %#v", build, Data.StreamPreviewUI.Active, wantActive)
		}
		if len(Data.StreamPreviewUI.Inactive) != 0 {
			t.Fatalf("buildDatabaseDVR() build %d inactive = %#v, want empty", build, Data.StreamPreviewUI.Inactive)
		}
	}
}
