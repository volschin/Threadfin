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
	PlaylistReadyCount  int `json:"playlistReadyCount"`
	SelectedStreamCount int `json:"selectedStreamCount"`
	XMLTV               struct {
		Applicable  bool `json:"applicable"`
		Ready       bool `json:"ready"`
		SourceCount int  `json:"sourceCount"`
		ReadyCount  int  `json:"readyCount"`
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
		ActiveClientConnections int `json:"activeClientConnections"`
		Clients                 struct {
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
		Key     string `json:"key"`
		Status  string `json:"status"`
		Summary string `json:"summary"`
		Action  struct {
			Destination string `json:"destination"`
		} `json:"action"`
	} `json:"stages"`
}

type overviewInteractionResult struct {
	ActivityLabel                string `json:"activityLabel"`
	ServerIdentityPreserved      bool   `json:"serverIdentityPreserved"`
	ClientInfoIdentityPreserved  bool   `json:"clientInfoIdentityPreserved"`
	LogIdentityPreserved         bool   `json:"logIdentityPreserved"`
	EndpointIdentityPreserved    bool   `json:"endpointIdentityPreserved"`
	EndpointFocusPreserved       bool   `json:"endpointFocusPreserved"`
	ActiveConnectionsAfterUpdate string `json:"activeConnectionsAfterUpdate"`
	ClientCapacityAfterUpdate    string `json:"clientCapacityAfterUpdate"`
	PlaylistCapacityAfterUpdate  string `json:"playlistCapacityAfterUpdate"`
	ActivityNoteAfterUpdate      string `json:"activityNoteAfterUpdate"`
	AttentionAfterUpdate         string `json:"attentionAfterUpdate"`
	ActiveConnectionsAfterReturn string `json:"activeConnectionsAfterReturn"`
	ClientCapacityAfterReturn    string `json:"clientCapacityAfterReturn"`
	ActivityNoteAfterReturn      string `json:"activityNoteAfterReturn"`
	ServerActiveClients          int    `json:"serverActiveClients"`
	ServerTotalClients           int    `json:"serverTotalClients"`
	ServerErrors                 int    `json:"serverErrors"`
	ServerWarnings               int    `json:"serverWarnings"`
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
	if state.Activity.ActiveClientConnections != 1 || state.Activity.Clients.Active != 1 || state.Activity.Clients.Total != 4 {
		t.Fatalf("PMS activity = %+v, want one active client connection and 1 / 4 client capacity", state.Activity)
	}
	assertOverviewStage(t, state, "xmltv", "managed", "settings")
	assertOverviewStage(t, state, "mapping", "managed", "settings")
	assertOverviewStage(t, state, "outputs", "ready", "connections")
}

func TestOverviewStateClassifiesMixedSourceHealthConservatively(t *testing.T) {
	fixture := readyXEPGFixture()
	files := fixture["settings"].(map[string]any)["files"].(map[string]any)
	files["m3u"].(map[string]any)["M2"] = map[string]any{
		"name": "Unavailable playlist", "provider.availability": 0,
	}
	files["xmltv"].(map[string]any)["X2"] = map[string]any{
		"name": "Empty guide", "provider.availability": 100,
		"compatibility": map[string]any{"xmltv.channels": 0, "xmltv.programs": 0},
	}

	state := evaluateOverviewState(t, fixture)

	if state.PlaylistCount != 2 || state.PlaylistReadyCount != 1 {
		t.Fatalf("mixed playlist health = %d / %d ready, want 1 / 2", state.PlaylistReadyCount, state.PlaylistCount)
	}
	assertOverviewStageSummary(t, state, "playlist", "attention", "1 / 2 ready")
	if state.XMLTV.SourceCount != 2 || state.XMLTV.ReadyCount != 1 || state.XMLTV.Ready {
		t.Fatalf("mixed XMLTV health = %+v, want 1 / 2 ready and conservative false readiness", state.XMLTV)
	}
	assertOverviewStageSummary(t, state, "xmltv", "attention", "1 / 2 ready")
	if state.Outputs.Ready {
		t.Fatal("XEPG output is ready while a configured XMLTV source is not ready")
	}
	assertOverviewEndpoint(t, state, "m3u", false)
	assertOverviewEndpoint(t, state, "xmltv", false)
}

func TestOverviewMappingCountsOnlyValidVisibleMappings(t *testing.T) {
	tests := []struct {
		name           string
		mapping        map[string]any
		wantActive     int
		wantUnresolved int
	}{
		{
			name:       "active valid channel",
			mapping:    map[string]any{"x-active": true, "x-xmltv-file": "X1", "x-mapping": "guide-1"},
			wantActive: 1,
		},
		{
			name:           "active channel missing guide",
			mapping:        map[string]any{"x-active": true, "x-xmltv-file": "-", "x-mapping": "guide-1"},
			wantUnresolved: 1,
		},
		{
			name:           "active channel missing mapping",
			mapping:        map[string]any{"x-active": true, "x-xmltv-file": "X1", "x-mapping": ""},
			wantUnresolved: 1,
		},
		{
			name:    "intentionally inactive valid channel",
			mapping: map[string]any{"x-active": false, "x-xmltv-file": "X1", "x-mapping": "guide-1"},
		},
		{
			name:    "hidden incomplete channel",
			mapping: map[string]any{"x-active": true, "x-hide-channel": true, "x-xmltv-file": "-", "x-mapping": "-"},
		},
		{
			name:       "active Threadfin dummy guide",
			mapping:    map[string]any{"x-active": true, "x-xmltv-file": "Threadfin Dummy", "x-mapping": "PPV"},
			wantActive: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := readyXEPGFixture()
			fixture["xepg"].(map[string]any)["epgMapping"] = map[string]any{"channel": test.mapping}
			state := evaluateOverviewState(t, fixture)
			if state.Mapping.ActiveCount != test.wantActive || state.Mapping.UnresolvedCount != test.wantUnresolved {
				t.Fatalf("mapping state = %+v, want active=%d unresolved=%d", state.Mapping, test.wantActive, test.wantUnresolved)
			}
		})
	}
}

func TestOverviewStateClassifiesReadyXEPG(t *testing.T) {
	state := evaluateOverviewState(t, readyXEPGFixture())

	if state.PlaylistCount != 1 || state.SelectedStreamCount != 2 {
		t.Fatalf("ready XEPG source counts = playlists %d, selected %d, want 1 and 2", state.PlaylistCount, state.SelectedStreamCount)
	}
	if !state.XMLTV.Applicable || !state.XMLTV.Ready || state.XMLTV.SourceCount != 1 || state.XMLTV.ReadyCount != 1 {
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
		"c1": map[string]any{"x-active": true, "x-xmltv-file": "X1", "x-mapping": "guide-1"},
		"c2": map[string]any{"x-active": true, "x-xmltv-file": "-", "x-mapping": "guide-2"},
		"c3": map[string]any{"x-active": false, "x-xmltv-file": "X1", "x-mapping": "-"},
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

func TestUpdateLogMergesOperationalStateAndSurvivesOverviewNavigation(t *testing.T) {
	result := evaluateOverviewInteraction(t)

	if result.ActivityLabel != "Active client connections" {
		t.Fatalf("activity metric label = %q, want honest active client connection label", result.ActivityLabel)
	}
	if !result.ServerIdentityPreserved || !result.ClientInfoIdentityPreserved || !result.LogIdentityPreserved {
		t.Fatalf("updateLog replaced state identity: server=%t clientInfo=%t log=%t", result.ServerIdentityPreserved, result.ClientInfoIdentityPreserved, result.LogIdentityPreserved)
	}
	if !result.EndpointIdentityPreserved || !result.EndpointFocusPreserved {
		t.Fatalf("updateLog replaced or blurred copied endpoint: identity=%t focus=%t", result.EndpointIdentityPreserved, result.EndpointFocusPreserved)
	}
	if result.ActiveConnectionsAfterUpdate != "2" || result.ClientCapacityAfterUpdate != "2 / 4" || result.PlaylistCapacityAfterUpdate != "1 / 2" {
		t.Fatalf("live activity after updateLog = active %q clients %q playlists %q", result.ActiveConnectionsAfterUpdate, result.ClientCapacityAfterUpdate, result.PlaylistCapacityAfterUpdate)
	}
	if result.ActivityNoteAfterUpdate != "Threadfin currently observes 2 active client connections." {
		t.Fatalf("live activity note after updateLog = %q", result.ActivityNoteAfterUpdate)
	}
	if result.AttentionAfterUpdate != "Threadfin reports 1 error and 3 warnings." {
		t.Fatalf("live attention after updateLog = %q", result.AttentionAfterUpdate)
	}
	if result.ActiveConnectionsAfterReturn != "2" || result.ClientCapacityAfterReturn != "2 / 4" || result.ActivityNoteAfterReturn != result.ActivityNoteAfterUpdate {
		t.Fatalf("rerender after navigation lost merged values: active %q clients %q note %q", result.ActiveConnectionsAfterReturn, result.ClientCapacityAfterReturn, result.ActivityNoteAfterReturn)
	}
	if result.ServerActiveClients != 2 || result.ServerTotalClients != 4 || result.ServerErrors != 1 || result.ServerWarnings != 3 {
		t.Fatalf("merged SERVER values = active %d total %d errors %d warnings %d", result.ServerActiveClients, result.ServerTotalClients, result.ServerErrors, result.ServerWarnings)
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
				"c1": map[string]any{"x-active": true, "x-xmltv-file": "X1", "x-mapping": "guide-1"},
				"c2": map[string]any{"x-active": true, "x-xmltv-file": "Threadfin Dummy", "x-mapping": "PPV"},
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

func evaluateOverviewInteraction(t *testing.T) overviewInteractionResult {
	t.Helper()

	fixture := readyXEPGFixture()
	fixture["clientInfo"].(map[string]any)["activePlaylist"] = 1
	fixture["log"].(map[string]any)["errors"] = 1
	fixtureJSON, err := json.Marshal(fixture)
	if err != nil {
		t.Fatal(err)
	}
	temp := t.TempDir()
	fixturePath := filepath.Join(temp, "fixture.json")
	if err := os.WriteFile(fixturePath, fixtureJSON, 0o600); err != nil {
		t.Fatal(err)
	}
	scriptPath := filepath.Join(temp, "overview-interaction.js")
	script := `
const fs = require("fs");
const vm = require("vm");

class TestElement {
  constructor(tagName, document) {
    this.tagName = tagName.toUpperCase();
    this.ownerDocument = document;
    this.children = [];
    this.parentNode = null;
    this.attributes = {};
    this.listeners = {};
    this.hidden = false;
    this.style = {};
    this.className = "";
    this.textContent = "";
    this.type = "";
    this.value = "";
    this._id = "";
    this.classList = { contains: value => this.className.split(/\s+/).includes(value) };
  }
  get id() { return this._id; }
  set id(value) {
    this._id = String(value);
    if (this._id) this.ownerDocument.byID[this._id] = this;
  }
  get firstChild() { return this.children.length ? this.children[0] : null; }
  get lastChild() { return this.children.length ? this.children[this.children.length - 1] : null; }
  appendChild(child) {
    child.parentNode = this;
    this.children.push(child);
    return child;
  }
  removeChild(child) {
    const index = this.children.indexOf(child);
    if (index >= 0) this.children.splice(index, 1);
    child.parentNode = null;
    return child;
  }
  setAttribute(name, value) {
    this.attributes[name] = String(value);
    if (name == "id") this.id = value;
  }
  getAttribute(name) { return this.attributes[name] === undefined ? null : this.attributes[name]; }
  removeAttribute(name) { delete this.attributes[name]; }
  addEventListener(name, listener) { this.listeners[name] = listener; }
  focus() { this.ownerDocument.activeElement = this; }
  select() {}
}

class TestDocument {
  constructor() {
    this.byID = {};
    this.cookie = "";
    this.activeElement = null;
    this.body = new TestElement("body", this);
  }
  createElement(tagName) { return new TestElement(tagName, this); }
  getElementById(id) { return this.byID[id] || null; }
  querySelector() { return null; }
  querySelectorAll() { return []; }
  execCommand(command) { return command == "copy"; }
}

function appendHost(document, tagName, id) {
  const element = document.createElement(tagName);
  element.id = id;
  document.body.appendChild(element);
  return element;
}

function findElement(root, predicate) {
  if (predicate(root)) return root;
  for (const child of root.children) {
    const match = findElement(child, predicate);
    if (match) return match;
  }
  return null;
}

const document = new TestDocument();
const main = appendHost(document, "main", "main-content");
const overviewHost = appendHost(document, "section", "overview-content");
appendHost(document, "section", "connections-content");
appendHost(document, "section", "activity-content");
appendHost(document, "section", "content");
appendHost(document, "nav", "main-menu");

let lastSocket = null;
class TestWebSocket {
  constructor(url) { this.url = url; lastSocket = this; }
  send(value) { this.sent = value; }
}
class TestClipboard {
  constructor(button, options) { this.button = button; this.options = options; }
  on() {}
}

const context = {
  console: { log() {}, warn() {} },
  document,
  navigator: { clipboard: { writeText() { return Promise.resolve(); } } },
  ClipboardJS: TestClipboard,
  WebSocket: TestWebSocket,
  SERVER: JSON.parse(fs.readFileSync(process.argv[6], "utf8")),
  SERVER_CONNECTION: false,
  WS_AVAILABLE: false,
  UNDO: {},
  showElement() {},
  showLogs() {},
  showPreview() {},
  enableGroupSelection() {},
	  alert() {},
	  addEventListener() {},
	  setTimeout(callback) { callback(); return 0; },
  clearTimeout() {},
};
context.window = context;
context.window.location = { protocol: "http:", hostname: "127.0.0.1", port: "34400", hash: "" };
context.window.history = {
  state: null,
  pushState(state, title, hash) { this.state = state; context.window.location.hash = hash; },
  replaceState(state, title, hash) { this.state = state; context.window.location.hash = hash; },
};
context.location = context.window.location;
vm.createContext(context);
for (let index = 2; index <= 5; index++) {
  vm.runInContext(fs.readFileSync(process.argv[index], "utf8"), context);
}

vm.runInContext("renderOverview(SERVER)", context);
const serverReference = context.SERVER;
const clientInfoReference = context.SERVER.clientInfo;
const logReference = context.SERVER.log;
const endpointReference = findElement(overviewHost, element => element.tagName == "BUTTON" && element.classList.contains("tf-copy-action"));
endpointReference.focus();
const activityLabel = findElement(overviewHost, element => element.tagName == "DT").textContent;

vm.runInContext('new Server("updateLog").request({})', context);
lastSocket.onopen.call(lastSocket);
lastSocket.onmessage.call(lastSocket, { data: JSON.stringify({
  status: true,
  clientInfo: { activeClients: 2 },
  log: { warnings: 3 },
}) });

const result = {
  activityLabel,
  serverIdentityPreserved: context.SERVER === serverReference,
  clientInfoIdentityPreserved: context.SERVER.clientInfo === clientInfoReference,
  logIdentityPreserved: context.SERVER.log === logReference,
  endpointIdentityPreserved: findElement(overviewHost, element => element.tagName == "BUTTON" && element.classList.contains("tf-copy-action")) === endpointReference,
  endpointFocusPreserved: document.activeElement === endpointReference,
  activeConnectionsAfterUpdate: document.getElementById("overview-active-client-connections")?.textContent || "",
  clientCapacityAfterUpdate: document.getElementById("overview-client-capacity")?.textContent || "",
  playlistCapacityAfterUpdate: document.getElementById("overview-playlist-capacity")?.textContent || "",
  activityNoteAfterUpdate: document.getElementById("overview-activity-summary")?.textContent || "",
  attentionAfterUpdate: document.getElementById("overview-attention-summary")?.textContent || "",
};

vm.runInContext('openDestination("connections", false)', context);
vm.runInContext('openDestination("overview", false)', context);
result.activeConnectionsAfterReturn = document.getElementById("overview-active-client-connections")?.textContent || "";
result.clientCapacityAfterReturn = document.getElementById("overview-client-capacity")?.textContent || "";
result.activityNoteAfterReturn = document.getElementById("overview-activity-summary")?.textContent || "";
result.serverActiveClients = context.SERVER.clientInfo.activeClients;
result.serverTotalClients = context.SERVER.clientInfo.totalClients;
result.serverErrors = context.SERVER.log.errors;
result.serverWarnings = context.SERVER.log.warnings;
process.stdout.write(JSON.stringify(result));
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o600); err != nil {
		t.Fatal(err)
	}

	command := exec.Command(
		"node", scriptPath,
		filepath.Join("..", "html", "js", "app_state_ts.js"),
		filepath.Join("..", "html", "js", "overview_ts.js"),
		filepath.Join("..", "html", "js", "navigation_ts.js"),
		filepath.Join("..", "html", "js", "network_ts.js"),
		fixturePath,
	)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("execute generated Overview update/navigation sequence: %v\n%s", err, output)
	}

	var result overviewInteractionResult
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("decode Overview update/navigation result: %v\n%s", err, output)
	}
	return result
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

func assertOverviewStageSummary(t *testing.T, state overviewStateResult, key, status, summary string) {
	t.Helper()
	for _, stage := range state.Stages {
		if stage.Key == key {
			if stage.Status != status || stage.Summary != summary {
				t.Fatalf("stage %q = status %q, summary %q; want %q and %q", key, stage.Status, stage.Summary, status, summary)
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
