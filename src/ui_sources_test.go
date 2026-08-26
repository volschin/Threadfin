package src

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLegacySourceCommandsAndPayloadShapeRemainExplicit(t *testing.T) {
	menu := readUITypeScript(t, "menu_ts.ts")
	savePopup := menu[strings.Index(menu, "function savePopupData("):strings.Index(menu, "function donePopupData(")]

	for _, contract := range []struct {
		kind          string
		saveCommand   string
		updateCommand string
	}{
		{kind: "m3u", saveCommand: "saveFilesM3U", updateCommand: "updateFileM3U"},
		{kind: "hdhr", saveCommand: "saveFilesHDHR", updateCommand: "updateFileHDHR"},
		{kind: "xmltv", saveCommand: "saveFilesXMLTV", updateCommand: "updateFileXMLTV"},
	} {
		t.Run(contract.kind, func(t *testing.T) {
			sectionStart := strings.Index(savePopup, `case "`+contract.kind+`":`)
			if sectionStart < 0 {
				t.Fatalf("missing %s source command branch", contract.kind)
			}
			section := savePopup[sectionStart:]
			if next := strings.Index(section[1:], "\n    case \""); next >= 0 {
				section = section[:next+1]
			}
			for _, command := range []string{contract.saveCommand, contract.updateCommand} {
				if !strings.Contains(section, `cmd = "`+command+`"`) {
					t.Errorf("%s branch no longer emits %s", contract.kind, command)
				}
			}
			for _, payloadLine := range []string{
				`data["files"] = new Object`,
				`data["files"][dataType] = new Object`,
				`data["files"][dataType][id] = input`,
			} {
				if !strings.Contains(section, payloadLine) {
					t.Errorf("%s payload no longer contains %q", contract.kind, payloadLine)
				}
			}
		})
	}

	if !strings.Contains(savePopup, `if (remove == true) {`) ||
		!strings.Contains(savePopup, `input["delete"] = true`) {
		t.Fatal("source delete no longer adds only the established delete marker to form values")
	}
}

func TestSourceAssetsArePresent(t *testing.T) {
	for _, name := range []string{
		filepath.Join("..", "ts", "sources_ts.ts"),
		filepath.Join("..", "html", "js", "sources_ts.js"),
	} {
		if _, err := os.Stat(name); err != nil {
			t.Errorf("source UI asset %q: %v", name, err)
		}
	}
}

func TestGeneratedSourceCommandsPreserveExactPayloads(t *testing.T) {
	temp := t.TempDir()
	scriptPath := filepath.Join(temp, "source-payloads.js")
	if err := os.WriteFile(scriptPath, []byte(sourcePayloadNodeScript), 0o600); err != nil {
		t.Fatal(err)
	}

	command := exec.Command("node", scriptPath, filepath.Join("..", "html", "js", "menu_ts.js"))
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("execute generated source payload characterization: %v\n%s", err, output)
	}

	var got []map[string]any
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode generated source payloads: %v\n%s", err, output)
	}
	want := []map[string]any{
		{
			"cmd": "saveFilesM3U",
			"data": map[string]any{"files": map[string]any{"m3u": map[string]any{"-": map[string]any{
				"name": "Local channels", "description": "Primary", "file.source": "/srv/threadfin/channels.m3u",
				"buffer": "ffmpeg", "tuner": float64(3), "http_proxy.ip": "", "http_proxy.port": "",
				"http_headers.origin": "https://origin.example", "http_headers.referer": "https://referer.example",
			}}}},
		},
		{
			"cmd": "updateFileM3U",
			"data": map[string]any{"files": map[string]any{"m3u": map[string]any{"M1": map[string]any{
				"name": "Remote channels", "description": "Edited", "file.source": "https://example.test/channels.m3u",
				"buffer": "-", "tuner": float64(2), "http_proxy.ip": "192.0.2.2", "http_proxy.port": "8080",
				"http_headers.origin": "", "http_headers.referer": "",
			}}}},
		},
		{
			"cmd": "saveFilesHDHR",
			"data": map[string]any{"files": map[string]any{"hdhr": map[string]any{"H1": map[string]any{
				"name": "Living room tuner", "description": "HDHR", "file.source": "192.0.2.20:5004",
				"buffer": "vlc", "tuner": float64(4), "http_proxy.ip": "", "http_proxy.port": "", "delete": true,
			}}}},
		},
		{
			"cmd": "saveFilesXMLTV",
			"data": map[string]any{"files": map[string]any{"xmltv": map[string]any{"X1": map[string]any{
				"name": "Guide", "description": "Local guide", "file.source": "C:\\guides\\guide.xml",
				"http_proxy.ip": "", "http_proxy.port": "", "delete": true,
			}}}},
		},
	}
	if !reflect.DeepEqual(got, want) {
		gotJSON, _ := json.MarshalIndent(got, "", "  ")
		wantJSON, _ := json.MarshalIndent(want, "", "  ")
		t.Fatalf("generated source payloads changed\ngot: %s\nwant: %s", gotJSON, wantJSON)
	}
}

func TestConfigurationWizardRequiredOnlyForMissingSettings(t *testing.T) {
	settingsPath := filepath.Join(t.TempDir(), "settings.json")
	required, err := configurationWizardRequired(settingsPath)
	if err != nil {
		t.Fatal(err)
	}
	if !required {
		t.Fatal("fresh configuration does not enable the existing first-run wizard")
	}

	if err := os.WriteFile(settingsPath, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	required, err = configurationWizardRequired(settingsPath)
	if err != nil {
		t.Fatal(err)
	}
	if required {
		t.Fatal("existing settings unexpectedly re-enable the first-run wizard")
	}
}

func TestInitDetectsWizardBeforeCreatingSettingsFile(t *testing.T) {
	content, err := os.ReadFile("config.go")
	if err != nil {
		t.Fatal(err)
	}
	config := string(content)
	detect := strings.Index(config, `configurationWizardRequired(getPlatformFile(System.Folder.Config + "settings.json"))`)
	create := strings.Index(config, "err = createSystemFiles()")
	if detect < 0 || create < 0 || detect > create {
		t.Fatal("startup does not detect a genuinely missing settings file before createSystemFiles makes it indistinguishable from an existing configuration")
	}
}

func TestGeneratedSourceStateUsesExistingResponseAndConfirmedResults(t *testing.T) {
	temp := t.TempDir()
	scriptPath := filepath.Join(temp, "source-state.js")
	if err := os.WriteFile(scriptPath, []byte(sourceStateNodeScript), 0o600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("node", scriptPath, filepath.Join("..", "html", "js", "sources_ts.js"))
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("execute generated source state: %v\n%s", err, output)
	}

	var got struct {
		Playlist []struct {
			ID                string `json:"id"`
			ProviderType      string `json:"providerType"`
			Name              string `json:"name"`
			TypeLabel         string `json:"typeLabel"`
			LastUpdate        string `json:"lastUpdate"`
			Availability      int    `json:"availability"`
			AvailabilityKnown bool   `json:"availabilityKnown"`
			Counts            []struct {
				Value string `json:"value"`
			} `json:"counts"`
		} `json:"playlist"`
		XMLTV []struct {
			ID           string `json:"id"`
			ProviderType string `json:"providerType"`
			Counts       []struct {
				Value string `json:"value"`
			} `json:"counts"`
		} `json:"xmltv"`
		Locations map[string]bool `json:"locations"`
		Progress  map[string]any  `json:"progress"`
		Error     map[string]any  `json:"error"`
		Success   map[string]any  `json:"success"`
		Next      map[string]any  `json:"next"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode generated source state: %v\n%s", err, output)
	}
	if len(got.Playlist) != 2 || got.Playlist[0].ID != "H1" || got.Playlist[1].ID != "M1" {
		t.Fatalf("playlist source order/IDs = %+v, want H1 then M1 by visible name", got.Playlist)
	}
	if item := got.Playlist[1]; item.ProviderType != "m3u" || item.TypeLabel != "M3U" || item.LastUpdate != "2026-08-26 10:00:00" || item.Availability != 75 || !item.AvailabilityKnown || len(item.Counts) != 5 || item.Counts[0].Value != "12" || item.Counts[1].Value != "2" || item.Counts[2].Value != "3%" || item.Counts[3].Value != "80%" || item.Counts[4].Value != "90%" {
		t.Fatalf("derived M3U row = %+v", item)
	}
	if len(got.XMLTV) != 1 || got.XMLTV[0].ID != "X1" || got.XMLTV[0].ProviderType != "xmltv" || len(got.XMLTV[0].Counts) != 2 || got.XMLTV[0].Counts[0].Value != "5" || got.XMLTV[0].Counts[1].Value != "20" {
		t.Fatalf("derived XMLTV row = %+v", got.XMLTV)
	}
	for location, accepted := range map[string]bool{
		"https": true, "http": true, "unix": true, "relative": true, "windows": true,
		"empty": false, "ftp": false,
	} {
		if got.Locations[location] != accepted {
			t.Errorf("location %s accepted=%t, want %t", location, got.Locations[location], accepted)
		}
	}
	if got.Progress["state"] != "progress" || got.Error["state"] != "error" || got.Error["message"] != "provider rejected fixture" {
		t.Fatalf("request feedback progress=%v error=%v", got.Progress, got.Error)
	}
	if got.Success["state"] != "success" || !strings.Contains(got.Success["message"].(string), "2026-08-26 11:00:00") {
		t.Fatalf("confirmed update feedback = %v", got.Success)
	}
	if got.Next["state"] != "success" || got.Next["nextDestination"] != "mapping" {
		t.Fatalf("new XMLTV next-step feedback = %v", got.Next)
	}
}

func TestSourceUIIntegrationAndWizardMarkup(t *testing.T) {
	menu := readUITypeScript(t, "menu_ts.ts")
	if !strings.Contains(menu, `renderSourceManagementPage(menuKey, doc)`) {
		t.Fatal("Playlist/XMLTV destinations do not use the shared source renderer")
	}
	if !strings.Contains(menu, `enhanceSourcePopup(dataType)`) || !strings.Contains(menu, `validateSourcePopup(dataType)`) {
		t.Fatal("existing source popups are missing shared guidance/field validation")
	}
	network := readUITypeScript(t, "network_ts.ts")
	if !strings.Contains(network, `completeSourceRequest(data["cmd"], data, response)`) {
		t.Fatal("source feedback is not tied to the actual WebSocket response")
	}
	if !strings.Contains(network, `completeConfigurationWizard()`) {
		t.Fatal("successful wizard completion does not use the Overview completion route")
	}

	index, err := os.ReadFile(filepath.Join("..", "html", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	configuration, err := os.ReadFile(filepath.Join("..", "html", "configuration.html"))
	if err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{"index": string(index), "configuration": string(configuration)} {
		if !strings.Contains(content, `src="js/sources_ts.js"`) {
			t.Errorf("%s page does not load the shared source module", name)
		}
	}
	for _, fragment := range []string{`class="tf-setup"`, `id="wizard-progress"`, `id="wizard-request-status"`, `href="css/app-shell.css"`} {
		if !strings.Contains(string(configuration), fragment) {
			t.Errorf("configuration wizard is missing %s", fragment)
		}
	}

	styles, err := os.ReadFile(filepath.Join("..", "html", "css", "app-shell.css"))
	if err != nil {
		t.Fatal(err)
	}
	for _, selector := range []string{".tf-app .tf-sources", ".tf-app .tf-source-row", ".tf-app .tf-source-actions", ".tf-setup .tf-wizard-progress"} {
		if !strings.Contains(string(styles), selector) {
			t.Errorf("source/wizard stylesheet is missing %s", selector)
		}
	}
}

func TestWizardCompletionUsesOverviewHash(t *testing.T) {
	temp := t.TempDir()
	scriptPath := filepath.Join(temp, "wizard-route.js")
	if err := os.WriteFile(scriptPath, []byte(wizardRouteNodeScript), 0o600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(
		"node", scriptPath,
		filepath.Join("..", "html", "js", "sources_ts.js"),
		filepath.Join("..", "html", "js", "configuration_ts.js"),
	)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("execute generated wizard route: %v\n%s", err, output)
	}
	if strings.TrimSpace(string(output)) != "/web/#overview" {
		t.Fatalf("wizard completion route = %q, want /web/#overview", output)
	}
}

func TestSourceAndWizardAssetsAreEmbeddedAndMatchSource(t *testing.T) {
	webUI = make(map[string]interface{})
	loadHTMLMap()
	for _, name := range []string{
		"html/configuration.html",
		"html/css/app-shell.css",
		"html/js/configuration_ts.js",
		"html/js/menu_ts.js",
		"html/js/network_ts.js",
		"html/js/sources_ts.js",
		"html/lang/en.json",
	} {
		t.Run(filepath.Base(name), func(t *testing.T) {
			encoded, ok := webUI[name].(string)
			if !ok {
				t.Fatalf("embedded asset %q is missing or is not a string", name)
			}
			got, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(filepath.Join("..", name))
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != string(want) {
				t.Errorf("embedded asset %q does not match source", name)
			}
		})
	}
}

const sourceStateNodeScript = `
const fs = require("fs");
const vm = require("vm");
const source = fs.readFileSync(process.argv[2], "utf8");
const status = {hidden: true, attributes: {}, setAttribute(k, v) { this.attributes[k] = v; }};
const context = {
  console: {log() {}, warn() {}},
  document: {getElementById(id) { return id === "source-form-status" ? status : null; }},
  SERVER: {},
};
vm.createContext(context);
vm.runInContext(source, context);
const fixture = {
  settings: {files: {
    m3u: {M1: {name: "Zulu", "file.source": "/srv/channels.m3u", "last.update": "2026-08-26 10:00:00", "provider.availability": 75, tuner: 2, compatibility: {streams: 12, "group.title": 3, "tvg.id": 80, "stream.id": 90}}},
    hdhr: {H1: {name: "Alpha", "file.source": "192.0.2.20:5004", compatibility: {streams: 4, "group.title": 1}}},
    xmltv: {X1: {name: "Guide", "file.source": "https://example.test/guide.xml", "provider.availability": 100, compatibility: {"xmltv.channels": 5, "xmltv.programs": 20}}},
  }},
};
const playlist = context.selectSourceList(fixture, "playlist");
const xmltv = context.selectSourceList(fixture, "xmltv");
const locations = {
  https: context.sourceLocationAccepted("https://example.test/file"),
  http: context.sourceLocationAccepted("http://example.test/file"),
  unix: context.sourceLocationAccepted("/srv/threadfin/file.xml"),
  relative: context.sourceLocationAccepted("guides/file.xml"),
  windows: context.sourceLocationAccepted("C:\\guides\\file.xml"),
  empty: context.sourceLocationAccepted("  "),
  ftp: context.sourceLocationAccepted("ftp://example.test/file"),
};
context.beginSourceRequest("m3u", "M1", false, 1);
const progress = context.sourceFeedbackByKey["m3u:M1"];
const request = {files: {m3u: {M1: {name: "Zulu", "file.source": "/srv/channels.m3u"}}}};
context.completeSourceRequest("updateFileM3U", request, {status: false, err: "provider rejected fixture"});
const error = context.sourceFeedbackByKey["m3u:M1"];
context.completeSourceRequest("updateFileM3U", request, {status: true, settings: {files: {m3u: {M1: {name: "Zulu", "file.source": "/srv/channels.m3u", "last.update": "2026-08-26 11:00:00"}}}}});
const success = context.sourceFeedbackByKey["m3u:M1"];
context.completeSourceRequest("saveFilesXMLTV", {files: {xmltv: {"-": {name: "New guide", "file.source": "/srv/new.xml"}}}}, {status: true, settings: {files: {xmltv: {X2: {name: "New guide", "file.source": "/srv/new.xml", "last.update": "2026-08-26 11:05:00"}}}}});
const next = context.sourcePageFeedback.xmltv;
process.stdout.write(JSON.stringify({playlist, xmltv, locations, progress, error, success, next}));
`

const wizardRouteNodeScript = `
const fs = require("fs");
const vm = require("vm");
let assigned = "";
const context = {
  console: {log() {}, warn() {}},
  window: {location: {assign(value) { assigned = value; }}},
  document: {getElementById() { return null; }, querySelectorAll() { return []; }},
  PopupContent: class {},
  Server: class {request() {}},
  showElement() {},
};
vm.createContext(context);
vm.runInContext(fs.readFileSync(process.argv[2], "utf8"), context);
vm.runInContext(fs.readFileSync(process.argv[3], "utf8"), context);
context.completeConfigurationWizard();
process.stdout.write(assigned);
`

const sourcePayloadNodeScript = `
const fs = require("fs");
const vm = require("vm");
const source = fs.readFileSync(process.argv[2], "utf8");
let controls = {inputs: [], selects: []};
const sent = [];
const popup = {
  getElementsByTagName(name) {
    if (name === "TABLE") return [{
      getElementsByTagName(child) {
        return child === "INPUT" ? controls.inputs : controls.selects;
      }
    }];
    return [];
  }
};
class CaptureServer {
  constructor(cmd) { this.cmd = cmd; }
  request(data) { sent.push({cmd: this.cmd, data: JSON.parse(JSON.stringify(data))}); }
}
const context = {
  console: {log() {}, warn() {}},
  document: {
    getElementById(id) { return id === "popup-custom" ? popup : null; },
    addEventListener() {},
  },
  Server: CaptureServer,
  showElement() {},
  confirm() { return true; },
  validateSourcePopup() { return true; },
  beginSourceRequest() {},
  UNDO: {}, SERVER: {}, BULK_EDIT: false,
};
vm.createContext(context);
vm.runInContext(source, context);
function input(type, name, value, checked) { return {type, name, value, checked: !!checked}; }
function select(name, value) { return {name, value}; }
function run(kind, id, remove, option, inputs, selects) {
  controls = {inputs, selects};
  context.savePopupData(kind, id, remove, option);
}
run("m3u", "-", false, 0, [
  input("text", "name", "Local channels"), input("text", "description", "Primary"),
  input("text", "file.source", "/srv/threadfin/channels.m3u"), input("text", "http_proxy.ip", ""),
  input("text", "http_proxy.port", ""), input("text", "http_headers.origin", "https://origin.example"),
  input("text", "http_headers.referer", "https://referer.example")
], [select("buffer", "ffmpeg"), select("tuner", "3")]);
run("m3u", "M1", false, 1, [
  input("text", "name", "Remote channels"), input("text", "description", "Edited"),
  input("text", "file.source", "https://example.test/channels.m3u"), input("text", "http_proxy.ip", "192.0.2.2"),
  input("text", "http_proxy.port", "8080"), input("text", "http_headers.origin", ""),
  input("text", "http_headers.referer", "")
], [select("buffer", "-"), select("tuner", "2")]);
run("hdhr", "H1", true, 0, [
  input("text", "name", "Living room tuner"), input("text", "description", "HDHR"),
  input("text", "file.source", "192.0.2.20:5004"), input("text", "http_proxy.ip", ""), input("text", "http_proxy.port", "")
], [select("buffer", "vlc"), select("tuner", "4")]);
run("xmltv", "X1", true, 0, [
  input("text", "name", "Guide"), input("text", "description", "Local guide"),
  input("text", "file.source", "C:\\guides\\guide.xml"), input("text", "http_proxy.ip", ""), input("text", "http_proxy.port", "")
], []);
process.stdout.write(JSON.stringify(sent));
`
