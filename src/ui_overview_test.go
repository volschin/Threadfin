package src

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type overviewStateResult struct {
	PlaylistCount       int `json:"playlistCount"`
	SelectedStreamCount int `json:"selectedStreamCount"`
	XMLTV               struct {
		Applicable  bool `json:"applicable"`
		Ready       bool `json:"ready"`
		SourceCount int  `json:"sourceCount"`
	} `json:"xmltv"`
	Mapping struct {
		ActiveCount     int `json:"activeCount"`
		UnresolvedCount int `json:"unresolvedCount"`
	} `json:"mapping"`
	Outputs struct {
		Ready     bool `json:"ready"`
		Endpoints []struct {
			Key       string `json:"key"`
			Available bool   `json:"available"`
		} `json:"endpoints"`
	} `json:"outputs"`
	Activity struct {
		ActiveStreams int `json:"activeStreams"`
		Clients       struct {
			Active int `json:"active"`
			Total  int `json:"total"`
		} `json:"clients"`
		Playlists struct {
			Active int `json:"active"`
			Total  int `json:"total"`
		} `json:"playlists"`
	} `json:"activity"`
	Attention struct {
		Errors   int `json:"errors"`
		Warnings int `json:"warnings"`
	} `json:"attention"`
	Sources []struct {
		Name         string `json:"name"`
		Kind         string `json:"kind"`
		Status       string `json:"status"`
		Availability int    `json:"availability"`
	} `json:"sources"`
	Stages []struct {
		Key    string `json:"key"`
		Status string `json:"status"`
		Action struct {
			Destination string `json:"destination"`
		} `json:"action"`
	} `json:"stages"`
}

func TestOverviewStateClassifiesEmptySetup(t *testing.T) {
	state := evaluateOverviewState(t, map[string]any{
		"settings": map[string]any{
			"epgSource": "XEPG",
			"files": map[string]any{
				"m3u": map[string]any{}, "hdhr": map[string]any{}, "xmltv": map[string]any{},
			},
		},
		"clientInfo": map[string]any{
			"epgSource": "XEPG", "streams": "0 / 0", "DVR": "127.0.0.1:34400",
			"m3u-url":  "http://127.0.0.1:34400/m3u/threadfin.m3u",
			"xepg-url": "http://127.0.0.1:34400/xmltv/threadfin.xml",
		},
	})

	if state.PlaylistCount != 0 || state.SelectedStreamCount != 0 {
		t.Fatalf("empty source counts = playlists %d, selected %d, want 0 and 0", state.PlaylistCount, state.SelectedStreamCount)
	}
	if !state.XMLTV.Applicable || state.XMLTV.Ready || state.XMLTV.SourceCount != 0 {
		t.Fatalf("empty XEPG XMLTV state = %+v, want applicable but not ready", state.XMLTV)
	}
	assertOverviewStage(t, state, "playlist", "empty", "playlist")
	assertOverviewStage(t, state, "filter", "waiting", "filter")
	assertOverviewStage(t, state, "xmltv", "attention", "xmltv")
	assertOverviewStage(t, state, "mapping", "waiting", "mapping")
	assertOverviewStage(t, state, "outputs", "waiting", "connections")
	if state.Outputs.Ready || len(state.Sources) != 0 {
		t.Fatalf("empty output/source state = ready %t, sources %d, want false and 0", state.Outputs.Ready, len(state.Sources))
	}
	assertOverviewEndpoint(t, state, "dvr", false)
	assertOverviewEndpoint(t, state, "m3u", false)
	assertOverviewEndpoint(t, state, "xmltv", false)
}

func TestOverviewStateClassifiesPMSAsClientManaged(t *testing.T) {
	state := evaluateOverviewState(t, map[string]any{
		"settings": map[string]any{
			"epgSource": "PMS",
			"files": map[string]any{
				"m3u": map[string]any{
					"M1": map[string]any{"name": "Local playlist", "last.update": "2026-08-26 10:00:00", "provider.availability": 100},
				},
				"hdhr": map[string]any{}, "xmltv": map[string]any{},
			},
		},
		"clientInfo": map[string]any{
			"epgSource": "PMS", "streams": "2 / 2", "DVR": "127.0.0.1:34400",
			"activeClients": 1, "totalClients": 4, "activePlaylist": 1, "totalPlaylist": 2,
		},
		"log": map[string]any{"errors": 0, "warnings": 0},
	})

	if state.PlaylistCount != 1 || state.SelectedStreamCount != 2 {
		t.Fatalf("PMS source counts = playlists %d, selected %d, want 1 and 2", state.PlaylistCount, state.SelectedStreamCount)
	}
	if state.XMLTV.Applicable || state.XMLTV.Ready {
		t.Fatalf("PMS XMLTV state = %+v, want client-managed and not applicable", state.XMLTV)
	}
	if !state.Outputs.Ready {
		t.Fatal("PMS DVR output with selected channels is not classified ready")
	}
	assertOverviewEndpoint(t, state, "dvr", true)
	assertOverviewEndpoint(t, state, "m3u", false)
	assertOverviewEndpoint(t, state, "xmltv", false)
	if state.Activity.ActiveStreams != 1 || state.Activity.Clients.Active != 1 || state.Activity.Clients.Total != 4 {
		t.Fatalf("PMS activity = %+v, want one active stream and 1 / 4 client capacity", state.Activity)
	}
	assertOverviewStage(t, state, "xmltv", "managed", "settings")
	assertOverviewStage(t, state, "mapping", "managed", "settings")
	assertOverviewStage(t, state, "outputs", "ready", "connections")
}

func TestOverviewStateClassifiesReadyXEPG(t *testing.T) {
	state := evaluateOverviewState(t, readyXEPGFixture())

	if state.PlaylistCount != 1 || state.SelectedStreamCount != 2 {
		t.Fatalf("ready XEPG source counts = playlists %d, selected %d, want 1 and 2", state.PlaylistCount, state.SelectedStreamCount)
	}
	if !state.XMLTV.Applicable || !state.XMLTV.Ready || state.XMLTV.SourceCount != 1 {
		t.Fatalf("ready XEPG XMLTV state = %+v", state.XMLTV)
	}
	if state.Mapping.ActiveCount != 2 || state.Mapping.UnresolvedCount != 0 {
		t.Fatalf("ready XEPG mapping state = %+v, want 2 active and 0 unresolved", state.Mapping)
	}
	if !state.Outputs.Ready {
		t.Fatal("ready XEPG endpoints are not classified ready")
	}
	assertOverviewEndpoint(t, state, "dvr", true)
	assertOverviewEndpoint(t, state, "m3u", true)
	assertOverviewEndpoint(t, state, "xmltv", true)
	if state.Attention.Errors != 0 || state.Attention.Warnings != 0 {
		t.Fatalf("ready XEPG attention = %+v, want clear", state.Attention)
	}
	if len(state.Sources) != 2 || state.Sources[0].Name != "Guide" || state.Sources[0].Kind != "XMLTV" || state.Sources[0].Status != "ready" {
		t.Fatalf("ready XEPG recent source state = %+v", state.Sources)
	}
	for _, key := range []string{"playlist", "filter", "xmltv", "mapping", "outputs"} {
		assertOverviewStage(t, state, key, "ready", map[string]string{
			"playlist": "playlist", "filter": "filter", "xmltv": "xmltv", "mapping": "mapping", "outputs": "connections",
		}[key])
	}
}

func TestOverviewStateClassifiesXEPGMappingAttention(t *testing.T) {
	fixture := readyXEPGFixture()
	fixture["xepg"].(map[string]any)["epgMapping"] = map[string]any{
		"c1": map[string]any{"x-active": true},
		"c2": map[string]any{"x-active": false},
		"c3": map[string]any{"x-active": false},
	}
	fixture["log"] = map[string]any{"errors": 2, "warnings": 3}

	state := evaluateOverviewState(t, fixture)

	if state.Mapping.ActiveCount != 1 || state.Mapping.UnresolvedCount != 2 {
		t.Fatalf("attention mapping state = %+v, want 1 active and 2 unresolved", state.Mapping)
	}
	if state.Attention.Errors != 2 || state.Attention.Warnings != 3 {
		t.Fatalf("attention log counts = %+v, want 2 errors and 3 warnings", state.Attention)
	}
	assertOverviewStage(t, state, "mapping", "attention", "mapping")
	if !state.Outputs.Ready {
		t.Fatal("one generated XEPG channel should keep Threadfin outputs ready while other mappings need attention")
	}
}

func TestOverviewModulesLoadInDependencyOrder(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "html", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	html := string(content)
	stateIndex := strings.Index(html, `src="js/app_state_ts.js"`)
	overviewIndex := strings.Index(html, `src="js/overview_ts.js"`)
	navigationIndex := strings.Index(html, `src="js/navigation_ts.js"`)
	baseIndex := strings.Index(html, `src="js/base_ts.js"`)
	if stateIndex < 0 || overviewIndex < 0 || navigationIndex < 0 || baseIndex < 0 {
		t.Fatalf("Overview dependency scripts are missing: app-state=%d overview=%d navigation=%d base=%d", stateIndex, overviewIndex, navigationIndex, baseIndex)
	}
	if !(stateIndex < overviewIndex && overviewIndex < navigationIndex && navigationIndex < baseIndex) {
		t.Fatalf("Overview script order is app-state=%d overview=%d navigation=%d base=%d", stateIndex, overviewIndex, navigationIndex, baseIndex)
	}
}

func TestOverviewIsInitialDestinationOnlyWithoutValidCurrentOrHistoryState(t *testing.T) {
	navigation := readUITypeScript(t, "navigation_ts.ts")
	start := strings.Index(navigation, "function restoreInitialDestinationFromHistory()")
	if start < 0 {
		t.Fatal("navigation is missing initial destination restoration")
	}
	end := strings.Index(navigation[start:], "function navigationDestinationIsKnown(")
	if end < 0 {
		t.Fatal("navigation is missing the destination validation boundary")
	}
	restore := navigation[start : start+end]
	if !strings.Contains(restore, `openDestination("overview", false)`) {
		t.Fatal("navigation does not open Overview when current state and history provide no valid destination")
	}
	if !strings.Contains(restore, "currentDestination !== undefined") {
		t.Fatal("Overview fallback can replace a destination already selected by a server response")
	}
}

func TestUpdateLogPatchesOnlyOperationalOverviewState(t *testing.T) {
	network := readUITypeScript(t, "network_ts.ts")
	start := strings.Index(network, `case "updateLog":`)
	if start < 0 {
		t.Fatal("network updateLog branch is missing")
	}
	end := strings.Index(network[start:], "default:")
	if end < 0 {
		t.Fatal("network updateLog branch has no default boundary")
	}
	updateLog := network[start : start+end]
	if !strings.Contains(updateLog, "refreshOverviewOperationalState(response)") {
		t.Fatal("updateLog does not patch the live Overview regions")
	}
	if strings.Contains(updateLog, "SERVER = response") {
		t.Fatal("partial updateLog response replaces the complete SERVER snapshot")
	}

	overview := readUITypeScript(t, "overview_ts.ts")
	start = strings.Index(overview, "function refreshOverviewOperationalState(")
	if start < 0 {
		t.Fatal("Overview is missing its operational patch function")
	}
	operational := overview[start:]
	if strings.Contains(operational, ".innerHTML") {
		t.Fatal("operational Overview refresh replaces DOM subtrees and can discard focused controls")
	}
	for _, id := range []string{"overview-active-streams", "overview-client-capacity", "overview-playlist-capacity", "overview-errors", "overview-warnings"} {
		if !strings.Contains(operational, `"`+id+`"`) {
			t.Errorf("operational Overview refresh does not patch %s", id)
		}
	}
}

func TestOverviewCopyUsesAvailableHelperAndPoliteFallbackStatus(t *testing.T) {
	overview := readUITypeScript(t, "overview_ts.ts")
	for _, contract := range []string{"ClipboardJS", "navigator.clipboard", `role", "status"`, `aria-live", "polite"`} {
		if !strings.Contains(overview, contract) {
			t.Errorf("Overview copy/status contract is missing %q", contract)
		}
	}
	if strings.Contains(strings.ToLower(overview), "gzip") {
		t.Fatal("Overview invents an XMLTV GZIP endpoint that the current server response does not provide")
	}
}

func TestOverviewInitialFocusDoesNotScrollBehindMobileNavigation(t *testing.T) {
	navigation := readUITypeScript(t, "navigation_ts.ts")
	start := strings.Index(navigation, "function focusMainContent()")
	if start < 0 {
		t.Fatal("navigation is missing main-content focus management")
	}
	end := strings.Index(navigation[start:], "function restoreDestinationFromHistory(")
	if end < 0 {
		t.Fatal("navigation is missing the focus-management boundary")
	}
	focus := navigation[start : start+end]
	if !strings.Contains(focus, "main.focus({ preventScroll: true })") {
		t.Fatal("initial Overview focus can scroll content behind the sticky mobile navigation")
	}
}

func TestOverviewGeneratedAssetsAreEmbeddedAndMatchSource(t *testing.T) {
	webUI = make(map[string]interface{})
	loadHTMLMap()

	for _, name := range []string{
		"html/js/app_state_ts.js",
		"html/js/overview_ts.js",
		"html/js/navigation_ts.js",
		"html/js/network_ts.js",
	} {
		t.Run(filepath.Base(name), func(t *testing.T) {
			encoded, ok := webUI[name].(string)
			if !ok {
				t.Fatalf("embedded asset %q is missing or is not a string", name)
			}
			got, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				t.Fatalf("decode embedded asset %q: %v", name, err)
			}
			want, err := os.ReadFile(filepath.Join("..", name))
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != string(want) {
				t.Errorf("embedded asset %q does not match its generated source", name)
			}
		})
	}
}

func readyXEPGFixture() map[string]any {
	return map[string]any{
		"settings": map[string]any{
			"epgSource": "XEPG",
			"files": map[string]any{
				"m3u": map[string]any{
					"M1": map[string]any{"name": "Playlist", "last.update": "2026-08-26 09:00:00", "provider.availability": 100},
				},
				"hdhr": map[string]any{},
				"xmltv": map[string]any{
					"X1": map[string]any{
						"name": "Guide", "last.update": "2026-08-26 10:00:00", "provider.availability": 100,
						"compatibility": map[string]any{"xmltv.channels": 2, "xmltv.programs": 4},
					},
				},
			},
		},
		"clientInfo": map[string]any{
			"epgSource": "XEPG", "streams": "2 / 2", "DVR": "127.0.0.1:34400",
			"m3u-url":       "http://127.0.0.1:34400/m3u/threadfin.m3u",
			"xepg-url":      "http://127.0.0.1:34400/xmltv/threadfin.xml",
			"activeClients": 0, "totalClients": 4, "activePlaylist": 0, "totalPlaylist": 2,
		},
		"xepg": map[string]any{
			"epgMapping": map[string]any{
				"c1": map[string]any{"x-active": true},
				"c2": map[string]any{"x-active": true},
			},
		},
		"log": map[string]any{"errors": 0, "warnings": 0},
	}
}

func evaluateOverviewState(t *testing.T, fixture map[string]any) overviewStateResult {
	t.Helper()

	fixtureJSON, err := json.Marshal(fixture)
	if err != nil {
		t.Fatal(err)
	}
	temp := t.TempDir()
	fixturePath := filepath.Join(temp, "fixture.json")
	if err := os.WriteFile(fixturePath, fixtureJSON, 0o600); err != nil {
		t.Fatal(err)
	}
	scriptPath := filepath.Join(temp, "evaluate-overview.js")
	script := `
const fs = require("fs");
const vm = require("vm");
const source = fs.readFileSync(process.argv[2], "utf8");
const fixture = JSON.parse(fs.readFileSync(process.argv[3], "utf8"));
const context = {};
vm.createContext(context);
vm.runInContext(source, context);
const result = vm.runInContext("selectOverviewState", context)(fixture);
process.stdout.write(JSON.stringify(result));
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o600); err != nil {
		t.Fatal(err)
	}

	assetPath := filepath.Join("..", "html", "js", "app_state_ts.js")
	command := exec.Command("node", scriptPath, assetPath, fixturePath)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("evaluate generated app-state selector: %v\n%s", err, output)
	}

	var state overviewStateResult
	if err := json.Unmarshal(output, &state); err != nil {
		t.Fatalf("decode app-state selector result: %v\n%s", err, output)
	}
	return state
}

func assertOverviewStage(t *testing.T, state overviewStateResult, key, status, destination string) {
	t.Helper()
	for _, stage := range state.Stages {
		if stage.Key == key {
			if stage.Status != status || stage.Action.Destination != destination {
				t.Fatalf("stage %q = status %q, destination %q; want %q and %q", key, stage.Status, stage.Action.Destination, status, destination)
			}
			return
		}
	}
	t.Fatalf("missing overview stage %q", key)
}

func assertOverviewEndpoint(t *testing.T, state overviewStateResult, key string, available bool) {
	t.Helper()
	for _, endpoint := range state.Outputs.Endpoints {
		if endpoint.Key == key {
			if endpoint.Available != available {
				t.Fatalf("endpoint %q available = %t, want %t", key, endpoint.Available, available)
			}
			return
		}
	}
	t.Fatalf("missing overview endpoint %q", key)
}
