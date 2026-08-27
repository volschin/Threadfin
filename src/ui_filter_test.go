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

func TestUIFilterAssetsAndLegacySaveContract(t *testing.T) {
	for _, name := range []string{
		filepath.Join("..", "ts", "filter_ts.ts"),
		filepath.Join("..", "html", "js", "filter_ts.js"),
	} {
		if _, err := os.Stat(name); err != nil {
			t.Errorf("filter UI asset %q: %v", name, err)
		}
	}

	menu := readUITypeScript(t, "menu_ts.ts")
	savePopup := menu[strings.Index(menu, "function savePopupData("):strings.Index(menu, "function donePopupData(")]
	for _, contract := range []string{
		`cmd = "saveFilter"`,
		`data["filter"] = new Object`,
		`data["filter"][id] = input`,
		`input["delete"] = true`,
		`case "tuner":`,
		`input[name] = (inputs[i] as HTMLInputElement).value`,
	} {
		if !strings.Contains(savePopup, contract) {
			t.Errorf("legacy Filter save contract no longer contains %q", contract)
		}
	}
}

func TestUIFilterGeneratedStateAndPayloads(t *testing.T) {
	temp := t.TempDir()
	scriptPath := filepath.Join(temp, "filter-ui.js")
	if err := os.WriteFile(scriptPath, []byte(filterUINodeScript), 0o600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("node", scriptPath, filepath.Join("..", "html", "js", "filter_ts.js"), filepath.Join("..", "html", "js", "menu_ts.js"))
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("execute generated Filter state and payloads: %v\n%s", err, output)
	}
	var got struct {
		Counts struct {
			Imported int `json:"imported"`
			Selected int `json:"selected"`
			Excluded int `json:"excluded"`
		} `json:"counts"`
		Order            []string         `json:"order"`
		Group            string           `json:"group"`
		Custom           string           `json:"custom"`
		CustomMalformed  string           `json:"customMalformed"`
		CustomRejected   []string         `json:"customRejected"`
		Payloads         []map[string]any `json:"payloads"`
		NoPreview        bool             `json:"noPreview"`
		InitialFeedback  bool             `json:"initialFeedback"`
		ServerError      string           `json:"serverError"`
		ConfirmedSuccess string           `json:"confirmedSuccess"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode generated Filter state and payloads: %v\n%s", err, output)
	}
	if got.Counts.Imported != 5 || got.Counts.Selected != 3 || got.Counts.Excluded != 2 {
		t.Fatalf("Filter saved-stream counts = %+v, want imported 5 selected 3 excluded 2", got.Counts)
	}
	if !reflect.DeepEqual(got.Order, []string{"G1", "C1"}) {
		t.Fatalf("Filter ordering = %v, want Group before Custom", got.Order)
	}
	for _, want := range []string{"exactly", "case-sensitive", "any of \"FHD\" or \"UHD\"", "any of \"ES\" or \"IT\""} {
		if !strings.Contains(got.Group, want) {
			t.Errorf("Group summary %q is missing %q", got.Group, want)
		}
	}
	for _, want := range []string{"complete stream data", "case-insensitive", "Includes complete stream data containing any of \"HD\" or \"UHD\"", "Excludes complete stream data containing any of \"ES\" or \"IT\""} {
		if !strings.Contains(got.Custom, want) {
			t.Errorf("Custom summary %q is missing %q", got.Custom, want)
		}
	}
	if !got.NoPreview {
		t.Fatal("Filter summary claims an unsaved preview")
	}
	if !strings.Contains(got.CustomMalformed, "evaluated as entered") || strings.Contains(got.CustomMalformed, "containing \"bad") {
		t.Fatalf("malformed Custom rule summary invents matching behavior: %q", got.CustomMalformed)
	}
	for _, summary := range got.CustomRejected {
		if !strings.Contains(summary, "evaluated as entered") || strings.Contains(summary, "Includes complete stream data") || strings.Contains(summary, "Excludes complete stream data") {
			t.Fatalf("noncanonical Custom rule summary invents clause semantics: %q", summary)
		}
	}
	if !got.InitialFeedback || got.ServerError != "server rejected fixture" || got.ConfirmedSuccess == "" {
		t.Fatalf("Filter feedback is not response-confirmed: initial=%t error=%q success=%q", got.InitialFeedback, got.ServerError, got.ConfirmedSuccess)
	}

	wantPayloads := []map[string]any{
		{"cmd": "saveFilter", "data": map[string]any{"filter": map[string]any{"-1": map[string]any{
			"name": "Sport", "description": "Sports channels", "type": "group-title", "filter": "Sports", "liveEvent": true, "caseSensitive": true, "include": "FHD,UHD", "exclude": "ES,IT", "startingNumber": "1000", "x-category": "sports",
		}}}},
		{"cmd": "saveFilter", "data": map[string]any{"filter": map[string]any{"C1": map[string]any{
			"name": "Catch all", "description": "", "type": "custom-filter", "filter": "tvg-id=abc {HD,UHD} !{ES,IT}", "caseSensitive": false, "startingNumber": "not-a-number", "x-category": "",
		}}}},
		{"cmd": "saveFilter", "data": map[string]any{"filter": map[string]any{"G1": map[string]any{"name": "Sport", "type": "group-title", "filter": "Sports", "delete": true}}}},
	}
	if !reflect.DeepEqual(got.Payloads, wantPayloads) {
		t.Fatalf("generated Filter create/edit/delete payloads changed\n got: %#v\nwant: %#v", got.Payloads, wantPayloads)
	}
}

func TestUIFilterProductionPopupSaveContract(t *testing.T) {
	temp := t.TempDir()
	scriptPath := filepath.Join(temp, "filter-production-popup.js")
	if err := os.WriteFile(scriptPath, []byte(filterProductionPopupNodeScript), 0o600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("node", scriptPath, filepath.Join("..", "html", "js", "filter_ts.js"), filepath.Join("..", "html", "js", "menu_ts.js"))
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("execute production Filter popup/save contract: %v\n%s", err, output)
	}
	var got struct {
		Payloads      []map[string]any `json:"payloads"`
		Starting      string           `json:"starting"`
		CancelLoading bool             `json:"cancelLoading"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode production Filter popup/save contract: %v\n%s", err, output)
	}
	if got.Starting != "not-a-number" {
		t.Fatalf("production Custom popup startingNumber = %q, want an unvalidated string", got.Starting)
	}
	if got.CancelLoading {
		t.Fatal("cancelled source or Filter delete showed loading")
	}
	if len(got.Payloads) != 5 {
		t.Fatalf("production Filter popup/save requests = %d, want group create/edit/delete and custom create/edit", len(got.Payloads))
	}
	wantPayloads := []map[string]any{
		{"cmd": "saveFilter", "data": map[string]any{"filter": map[string]any{"-1": map[string]any{"name": "group-title name", "description": "", "type": "group-title", "filter": "Sports", "liveEvent": false, "caseSensitive": false, "include": "", "exclude": "", "startingNumber": "1000", "x-category": ""}}}},
		{"cmd": "saveFilter", "data": map[string]any{"filter": map[string]any{"G1": map[string]any{"name": "group-title name", "description": "", "type": "group-title", "filter": "Sports", "liveEvent": false, "caseSensitive": false, "include": "", "exclude": "", "startingNumber": "1000", "x-category": ""}}}},
		{"cmd": "saveFilter", "data": map[string]any{"filter": map[string]any{"G1": map[string]any{"name": "group-title name", "description": "", "type": "group-title", "filter": "Sports", "liveEvent": false, "caseSensitive": false, "include": "", "exclude": "", "startingNumber": "1000", "x-category": "", "delete": true}}}},
		{"cmd": "saveFilter", "data": map[string]any{"filter": map[string]any{"-1": map[string]any{"name": "custom-filter name", "description": "", "type": "custom-filter", "caseSensitive": false, "filter": "tvg-id=abc {HD,UHD} !{ES,IT}", "startingNumber": "not-a-number", "x-category": ""}}}},
		{"cmd": "saveFilter", "data": map[string]any{"filter": map[string]any{"C1": map[string]any{"name": "custom-filter name", "description": "", "type": "custom-filter", "caseSensitive": false, "filter": "tvg-id=abc {HD,UHD} !{ES,IT}", "startingNumber": "not-a-number", "x-category": ""}}}},
	}
	if !reflect.DeepEqual(got.Payloads, wantPayloads) {
		t.Fatalf("production Filter popup/save payloads changed\n got: %#v\nwant: %#v", got.Payloads, wantPayloads)
	}
	for _, payload := range got.Payloads {
		if payload["cmd"] != "saveFilter" {
			t.Fatalf("production popup used command %#v, want saveFilter", payload["cmd"])
		}
		filter, ok := payload["data"].(map[string]any)["filter"].(map[string]any)
		if !ok || len(filter) != 1 {
			t.Fatalf("production Filter popup changed save envelope: %#v", payload)
		}
	}
}

func TestUIFilterGeneratedDOMRendersSavedCountsAndStates(t *testing.T) {
	temp := t.TempDir()
	scriptPath := filepath.Join(temp, "filter-dom.js")
	if err := os.WriteFile(scriptPath, []byte(filterDOMNodeScript), 0o600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("node", scriptPath, filepath.Join("..", "html", "js", "filter_ts.js"))
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("execute generated Filter DOM: %v\n%s", err, output)
	}
	var got struct {
		Populated        string `json:"populated"`
		Empty            string `json:"empty"`
		DeleteFocus      int    `json:"deleteFocus"`
		MainFocus        int    `json:"mainFocus"`
		InvokerFocus     int    `json:"invokerFocus"`
		ReplacementFocus int    `json:"replacementFocus"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode generated Filter DOM: %v\n%s", err, output)
	}
	for _, want := range []string{
		"Filter", "Filters select which imported streams become channels available for Mapping.",
		"{{.filter.counts.imported}}", "5", "{{.filter.counts.selected}}", "3", "{{.filter.counts.excluded}}", "2", "{{.filter.group}}", "{{.filter.custom}}", "Sport", "Catch all",
	} {
		if !strings.Contains(got.Populated, want) {
			t.Errorf("rendered populated Filter page is missing %q: %q", want, got.Populated)
		}
	}
	if strings.Index(got.Populated, "{{.filter.group}}") > strings.Index(got.Populated, "{{.filter.custom}}") {
		t.Fatal("rendered Filter page does not put Group Filter before Custom Filter")
	}
	if strings.Contains(strings.ToLower(got.Populated), "preview") {
		t.Fatal("rendered Filter page claims an unsaved preview")
	}
	if got.MainFocus != 1 || got.DeleteFocus != 1 {
		t.Fatalf("Filter deletion did not restore fallback after main focus and modal hide: %+v", got)
	}
	if got.InvokerFocus != 1 || got.ReplacementFocus != 1 {
		t.Fatalf("ordinary Filter popup focus restoration regressed: %+v", got)
	}
	for _, want := range []string{"{{.filter.empty.title}}", "{{.filter.addGroup}}"} {
		if !strings.Contains(got.Empty, want) {
			t.Errorf("rendered empty Filter state is missing %q: %q", want, got.Empty)
		}
	}
}

func TestUIFilterGeneratedPopupValidationRetainsValuesAndFocus(t *testing.T) {
	temp := t.TempDir()
	scriptPath := filepath.Join(temp, "filter-popup.js")
	if err := os.WriteFile(scriptPath, []byte(filterPopupNodeScript), 0o600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("node", scriptPath, filepath.Join("..", "html", "js", "filter_ts.js"))
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("execute generated Filter popup validation: %v\n%s", err, output)
	}
	var got struct {
		Valid                bool   `json:"valid"`
		Name                 string `json:"name"`
		Starting             string `json:"starting"`
		Focused              int    `json:"focused"`
		Invalid              bool   `json:"invalid"`
		Error                bool   `json:"error"`
		Summary              string `json:"summary"`
		PopupClass           bool   `json:"popupClass"`
		NameErrorAssociation string `json:"nameErrorAssociation"`
		RuleErrorAssociation string `json:"ruleErrorAssociation"`
		ErrorRetainedName    string `json:"errorRetainedName"`
		ErrorRetainedFocus   int    `json:"errorRetainedFocus"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode generated Filter popup validation: %v\n%s", err, output)
	}
	if got.Valid || got.Name != "" || got.Starting != "not-a-number" || got.Focused != 1 || !got.Invalid || !got.Error || !got.PopupClass {
		t.Fatalf("Filter popup validation did not retain/focus the invalid form: %+v", got)
	}
	if got.NameErrorAssociation != "filter-name-error" || got.RuleErrorAssociation != "filter-filter-error" {
		t.Fatalf("Filter popup validation errors are not associated with their fields: %+v", got)
	}
	if got.ErrorRetainedName != "Retained" || got.ErrorRetainedFocus != 1 {
		t.Fatalf("Filter error response did not retain modal values/focus: %+v", got)
	}
	for _, want := range []string{"exactly \"Sports\"", "case-sensitive", "any of \"FHD\" or \"UHD\""} {
		if !strings.Contains(got.Summary, want) {
			t.Errorf("Filter popup summary %q is missing %q", got.Summary, want)
		}
	}
}

func TestUIFilterGeneratedTransportRetainsFormState(t *testing.T) {
	temp := t.TempDir()
	scriptPath := filepath.Join(temp, "filter-transport.js")
	if err := os.WriteFile(scriptPath, []byte(filterTransportNodeScript), 0o600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("node", scriptPath, filepath.Join("..", "html", "js", "filter_ts.js"), filepath.Join("..", "html", "js", "network_ts.js"))
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("execute generated Filter transport contract: %v\n%s", err, output)
	}
	var got struct {
		BusyStatus      string `json:"busyStatus"`
		TransportStatus string `json:"transportStatus"`
		Name            string `json:"name"`
		Focus           int    `json:"focus"`
		Settled         bool   `json:"settled"`
		LoadingHides    int    `json:"loadingHides"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode generated Filter transport contract: %v\n%s", err, output)
	}
	if got.BusyStatus != "busy" || got.TransportStatus != "transport" || got.Name != "Retained" || got.Focus != 0 || !got.Settled || got.LoadingHides != 1 {
		t.Fatalf("Filter busy/transport handling did not retain and settle the form: %+v", got)
	}
}

func TestUIEmbeddedAssets(t *testing.T) {
	index, err := os.ReadFile(filepath.Join("..", "html", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(index), `src="js/filter_ts.js"`) {
		t.Fatal("index does not load the Filter module")
	}
	styles, err := os.ReadFile(filepath.Join("..", "html", "css", "app-shell.css"))
	if err != nil {
		t.Fatal(err)
	}
	for _, selector := range []string{".tf-app .tf-filters", ".tf-app .tf-filter-row", ".tf-filter-popup :is(button, input, select, [role=\"button\"])"} {
		if !strings.Contains(string(styles), selector) {
			t.Errorf("Filter stylesheet is missing %s", selector)
		}
	}

	webUI = make(map[string]interface{})
	loadHTMLMap()
	for _, name := range []string{
		"html/index.html",
		"html/css/app-shell.css",
		"html/js/filter_ts.js",
		"html/lang/en.json",
	} {
		t.Run(filepath.Base(name), func(t *testing.T) {
			encoded, ok := webUI[name].(string)
			if !ok {
				t.Fatalf("embedded Filter asset %q is missing", name)
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
				t.Errorf("embedded Filter asset %q does not match source", name)
			}
		})
	}
}

const filterUINodeScript = `
const fs = require("fs");
const vm = require("vm");
const filterSource = fs.readFileSync(process.argv[2], "utf8");
const menuSource = fs.readFileSync(process.argv[3], "utf8");
let controls = {inputs: [], selects: []};
const payloads = [];
const popup = {getElementsByTagName(name) { return name === "TABLE" ? [{getElementsByTagName(child) { return child === "INPUT" ? controls.inputs : controls.selects; }}] : []; }};
class CaptureServer { constructor(cmd) { this.cmd = cmd; } request(data) { payloads.push({cmd: this.cmd, data: JSON.parse(JSON.stringify(data))}); } }
const context = {
  console: {log() {}, warn() {}},
  document: {getElementById(id) { return id === "popup-custom" ? popup : null; }, addEventListener() {}},
  Server: CaptureServer, showElement() {}, confirm() { return true; }, validateSourcePopup() { return true; }, validateFilterPopup() { return true; }, beginSourceRequest() {}, beginFilterRequest() {},
  UNDO: {}, SERVER: {}, BULK_EDIT: false,
};
vm.createContext(context);
vm.runInContext(filterSource, context);
vm.runInContext(menuSource, context);
const fixture = {settings: {filter: {
  C1: {name: "Catch all", type: "custom-filter", filter: "tvg-id=abc {HD,UHD} !{ES,IT}", startingNumber: "not-a-number", caseSensitive: false},
  G1: {name: "Sport", type: "group-title", filter: "Sports", startingNumber: "1000", caseSensitive: true, include: "FHD,UHD", exclude: "ES,IT"}
}}, data: {StreamPreviewUI: {activeStreams: ["one", "two", "three"], inactiveStreams: ["four", "five"]}}};
const counts = context.filterStreamCounts(fixture);
const filters = context.selectFilterList(fixture);
function input(type, name, value, checked) { return {type, name, value, checked: !!checked}; }
function select(name, value) { return {name, value}; }
function save(id, inputs, selects, remove) { controls = {inputs, selects}; context.savePopupData("filter", id, !!remove, 0); }
save("-1", [input("text", "name", "Sport"), input("text", "description", "Sports channels"), input("hidden", "type", "group-title"), input("checkbox", "liveEvent", "", true), input("checkbox", "caseSensitive", "", true), input("text", "include", "FHD,UHD"), input("text", "exclude", "ES,IT"), input("text", "startingNumber", "1000")], [select("filter", "Sports"), select("x-category", "sports")]);
save("C1", [input("text", "name", "Catch all"), input("text", "description", ""), input("hidden", "type", "custom-filter"), input("checkbox", "caseSensitive", "", false), input("text", "filter", "tvg-id=abc {HD,UHD} !{ES,IT}"), input("text", "startingNumber", "not-a-number")], [select("x-category", "")]);
save("G1", [input("text", "name", "Sport"), input("hidden", "type", "group-title"), input("text", "filter", "Sports")], [], true);
const initialFeedback = context.filterPageFeedback.filter === undefined;
context.completeFilterRequest("saveFilter", {filter: {"G1": {name: "Sport"}}}, {status: false, err: "server rejected fixture"});
const serverError = context.filterPageFeedback.filter.message;
context.completeFilterRequest("saveFilter", {filter: {"G1": {name: "Sport"}}}, {status: true});
const confirmedSuccess = context.filterPageFeedback.filter.message;
const rejectedRules = ["Sport !{,}", "Sport {HD.one}", " Sport {HD}", "Sport  {HD}", "Sport {HD", "Sport {HD} {UHD}"];
process.stdout.write(JSON.stringify({counts, order: filters.map(filter => filter.id), group: context.filterRuleSummary(fixture.settings.filter.G1), custom: context.filterRuleSummary(fixture.settings.filter.C1), customMalformed: context.filterRuleSummary({type: "custom-filter", filter: "bad {include", caseSensitive: false}), customRejected: rejectedRules.map(rule => context.filterRuleSummary({type: "custom-filter", filter: rule, caseSensitive: false})), payloads, noPreview: !/preview|result count|will select/i.test(context.filterRuleSummary(fixture.settings.filter.G1)), initialFeedback, serverError, confirmedSuccess}));
`

const filterDOMNodeScript = `
const fs = require("fs");
const vm = require("vm");
class Element {
  constructor(tag) { this.tagName = tag.toUpperCase(); this.children = []; this.attributes = {}; this.className = ""; this.textContent = ""; this.hidden = false; this.listeners = {}; this.focusCount = 0; }
  set innerHTML(value) { this.children = []; this.textContent = value; }
  get innerHTML() { return this.textContent; }
  appendChild(child) { this.children.push(child); return child; }
  setAttribute(name, value) { this.attributes[name] = String(value); }
  getAttribute(name) { return this.attributes[name]; }
  addEventListener(name, listener) { this.listeners[name] = listener; }
  focus() { this.focusCount += 1; }
}
function text(node) { return (node.textContent || "") + node.children.map(text).join(" "); }
function descendants(node) { return node.children.reduce((all, child) => all.concat(child, descendants(child)), []); }
const context = {
  console: {log() {}, warn() {}},
  document: {createElement(tag) { return new Element(tag); }, getElementById(id) { return id === "popup" ? modal : id === "main-content" ? main : null; }, contains(element) { return element !== deletedInvoker && element !== disconnectedInvoker; }, querySelectorAll() { return [replacement]; }},
  openPopUp() {},
  SERVER: {settings: {filter: {
  C1: {name: "Catch all", type: "custom-filter", filter: "tvg-id=abc {HD,UHD} !{ES,IT}", startingNumber: "not-a-number", caseSensitive: false},
    G1: {name: "Sport", type: "group-title", filter: "Sports", startingNumber: "1000", caseSensitive: true, include: "FHD,UHD", exclude: "ES,IT"}
  }}, data: {StreamPreviewUI: {activeStreams: ["one", "two", "three"], inactiveStreams: ["four", "five"]}}},
};
vm.createContext(context);
vm.runInContext(fs.readFileSync(process.argv[2], "utf8"), context);
const modal = new Element("div");
const main = new Element("main");
const replacement = new Element("button");
replacement.setAttribute("data-filter-focus-key", "G1:edit");
const connectedInvoker = new Element("button");
connectedInvoker.setAttribute("data-filter-focus-key", "G1:edit");
const disconnectedInvoker = new Element("button");
disconnectedInvoker.setAttribute("data-filter-focus-key", "G1:edit");
const deletedInvoker = new Element("button");
deletedInvoker.setAttribute("data-filter-focus-key", "G1:delete");
context.openFilterPopup("group-title", {id: "G1"}, connectedInvoker);
modal.listeners["hidden.bs.modal"]();
context.openFilterPopup("group-title", {id: "G1"}, disconnectedInvoker);
modal.listeners["hidden.bs.modal"]();
const populated = new Element("section");
context.renderFilterManagementPage(populated);
context.SERVER = {settings: {filter: {}}, data: {StreamPreviewUI: {activeStreams: [], inactiveStreams: []}}};
const empty = new Element("section");
context.renderFilterManagementPage(empty);
context.SERVER = {settings: {filter: {G1: {name: "Sport", type: "group-title", filter: "Sports", startingNumber: "1000"}}}, data: {StreamPreviewUI: {activeStreams: [], inactiveStreams: []}}};
context.openFilterPopup("group-title", {id: "G1"}, deletedInvoker);
context.completeFilterRequest("saveFilter", {filter: {G1: {delete: true}}}, {status: true});
const afterDelete = new Element("section");
context.renderFilterManagementPage(afterDelete);
main.focus();
modal.listeners["hidden.bs.modal"]();
const addGroup = descendants(afterDelete).find(element => element.getAttribute("data-filter-focus-key") === "add-group");
process.stdout.write(JSON.stringify({populated: text(populated), empty: text(empty), deleteFocus: addGroup ? addGroup.focusCount : 0, mainFocus: main.focusCount, invokerFocus: connectedInvoker.focusCount, replacementFocus: replacement.focusCount}));
`

const filterProductionPopupNodeScript = `
const fs = require("fs");
const vm = require("vm");
const payloads = [];
const displayCalls = [];
let acceptConfirm = true;
let controls = [];
function control(tag, type, name, value) { const item = {tagName: tag, type: type || "", name: name || "", value: value === undefined ? "" : value, setAttribute(key, value) { this[key] = String(value); if (key === "type" || key === "name" || key === "value") this[key] = String(value); }}; let checked = false; Object.defineProperty(item, "checked", {get() { return checked; }, set(value) { checked = !!value; }}); return item; }
const table = {getElementsByTagName(tag) { return controls.filter(item => item.tagName === tag); }};
const popup = {classList: {add() {}, remove() {}}, getElementsByTagName(tag) { return tag === "TABLE" ? [table] : []; }, querySelectorAll() { return []; }, querySelector() { return null; }, appendChild() {}};
class CaptureServer { constructor(cmd) { this.cmd = cmd; } request(data) { payloads.push({cmd: this.cmd, data: JSON.parse(JSON.stringify(data))}); } }
const context = {console: {log() {}, warn() {}}, makeControl: control, collect(item) { controls.push(item); }, document: {getElementById(id) { return id === "popup-custom" ? popup : null; }, createElement() { return control("INPUT"); }, addEventListener() {}}, Server: CaptureServer, showElement(id, value) { displayCalls.push([id, value]); }, showPopUpElement() {}, confirm() { return acceptConfirm; }, validateSourcePopup() { return true; }, validateFilterPopup() { return true; }, beginSourceRequest() {}, beginFilterRequest() {}, enhanceSourcePopup() {}, enhanceFilterPopup() {}, getLocalData(type, id) { if (type === "m3uGroups") return {text: ["Sports"], value: ["Sports"]}; return id === -1 ? {} : context.SERVER.settings.filter[id]; }, UNDO: {}, BULK_EDIT: false, SERVER: {settings: {epgCategories: "", filter: {G1: {name: "Old group", description: "", type: "group-title", filter: "Sports", liveEvent: false, caseSensitive: false, include: "", exclude: "", startingNumber: "1000", "x-category": ""}, C1: {name: "Old custom", description: "", type: "custom-filter", filter: "tvg-id=old", caseSensitive: false, startingNumber: "1000", "x-category": ""}}}}};
vm.createContext(context);
vm.runInContext(fs.readFileSync(process.argv[2], "utf8"), context);
vm.runInContext(fs.readFileSync(process.argv[3], "utf8"), context);
vm.runInContext('PopupContent = class { createHeadline() {} createInteraction() {} addInteraction() {} description() {} createInput(type, name, value) { const item = makeControl("INPUT", type, name, value); collect(item); return item; } createCheckbox(name) { const item = makeControl("INPUT", "checkbox", name, ""); collect(item); return item; } createSelect(text, values, set, name) { const item = makeControl("SELECT", "", name, set === undefined || set === "" ? values[0] : set); collect(item); return item; } appendRow() {} };', context);
function set(name, value, checked) { const item = controls.find(item => item.name === name); if (!item) throw new Error("missing generated field " + name); if (checked !== undefined) item.checked = checked; else item.value = value; }
function save(kind, id, update, remove) { controls = []; context.openPopUp(kind, update ? {id} : undefined); set("name", kind + " name"); set("filter", kind === "group-title" ? "Sports" : "tvg-id=abc {HD,UHD} !{ES,IT}"); const start = controls.find(item => item.name === "startingNumber"); if (start) start.value = kind === "custom-filter" ? "not-a-number" : "1000"; context.savePopupData("filter", id, !!remove, 0); return start && start.value; }
save("group-title", "-1", false, false);
save("group-title", "G1", true, false);
save("group-title", "G1", true, true);
const starting = save("custom-filter", "-1", false, false);
save("custom-filter", "C1", true, false);
acceptConfirm = false;
displayCalls.length = 0;
controls = [];
context.savePopupData("filter", "G1", true, 0);
context.savePopupData("m3u", "M1", true, 0);
process.stdout.write(JSON.stringify({payloads, starting, cancelLoading: displayCalls.some(call => call[0] === "loading" && call[1] === true)}));
`

const filterTransportNodeScript = `
const fs = require("fs");
const vm = require("vm");
const sockets = [];
const displayCalls = [];
class Element {
  constructor(id) { this.id = id || ""; this.value = ""; this.focusCount = 0; this.hidden = true; this.attributes = {}; }
  setAttribute(name, value) { this.attributes[name] = String(value); }
  getAttribute(name) { return this.attributes[name]; }
  focus() { this.focusCount += 1; }
}
class FakeWebSocket {
  constructor() { sockets.push(this); }
  send() {}
}
const name = new Element();
name.value = "Retained";
const popup = {querySelectorAll() { return [name]; }};
const status = new Element("filter-form-status");
const context = {
  console: {log() {}, warn() {}},
  document: {getElementById(id) { return id === "popup-custom" ? popup : id === "filter-form-status" ? status : null; }},
  window: {location: {protocol: "http:", hostname: "127.0.0.1", port: "34400"}},
  WebSocket: FakeWebSocket,
  getCookie() { return ""; },
  showElement(id, value) { displayCalls.push([id, value]); },
  alert() {},
  SERVER_CONNECTION: false,
  WS_AVAILABLE: true,
  UNDO: {},
};
vm.createContext(context);
vm.runInContext(fs.readFileSync(process.argv[2], "utf8").replaceAll("{{.sources.responseInvalid}}", "invalid"), context);
vm.runInContext(fs.readFileSync(process.argv[3], "utf8").replaceAll("{{.sources.requestBusy}}", "busy").replaceAll("{{.sources.transportError}}", "transport").replaceAll("{{.sources.responseInvalid}}", "invalid"), context);
const data = {filter: {G1: {name: "Retained"}}};
context.data = data;
context.SERVER_CONNECTION = true;
vm.runInContext('new Server("saveFilter").request(data)', context);
const busyStatus = status.textContent;
context.SERVER_CONNECTION = false;
status.textContent = "";
displayCalls.length = 0;
vm.runInContext('new Server("saveFilter").request(data)', context);
const socket = sockets[sockets.length - 1];
socket.onerror({});
socket.onclose({code: 1006});
process.stdout.write(JSON.stringify({busyStatus, transportStatus: status.textContent, name: name.value, focus: name.focusCount, settled: context.SERVER_CONNECTION === false, loadingHides: displayCalls.filter(call => call[0] === "loading" && call[1] === false).length}));
`

const filterPopupNodeScript = `
const fs = require("fs");
const vm = require("vm");
class Element {
  constructor(tag) { this.tagName = tag.toUpperCase(); this.attributes = {}; this.children = []; this.parentElement = null; this.className = ""; this.hidden = false; this.textContent = ""; this.value = ""; this.focusCount = 0; this.listeners = {}; this.classList = {add: value => { if (!this.className.split(/\s+/).includes(value)) this.className = (this.className + " " + value).trim(); }, remove: value => { this.className = this.className.split(/\s+/).filter(name => name && name !== value).join(" "); }}; }
  setAttribute(name, value) { this.attributes[name] = String(value); }
  getAttribute(name) { return this.attributes[name]; }
  removeAttribute(name) { delete this.attributes[name]; }
  appendChild(child) { child.parentElement = this; this.children.push(child); return child; }
  focus() { this.focusCount += 1; }
  addEventListener(name, listener) { this.listeners[name] = listener; }
  closest(selector) { return selector === "tr" ? this.row : null; }
}
function field(tag, name, label, value, type, checked) { const control = new Element(tag); control.type = type || ""; control.checked = !!checked; control.name = name; control.setAttribute("name", name); control.value = value; const labelCell = new Element("td"); labelCell.textContent = label + ":"; control.row = {querySelector(selector) { return selector === "td:first-child" ? labelCell : null; }}; const parent = new Element("td"); parent.appendChild(control); return control; }
const name = field("input", "name", "Filter Name", "", "text");
const filter = field("select", "filter", "Group Title", "Sports");
const include = field("input", "include", "Include", "FHD,UHD", "text");
const exclude = field("input", "exclude", "Exclude", "ES,IT", "text");
const sensitive = field("input", "caseSensitive", "Case Sensitive", "", "checkbox", true);
const type = field("input", "type", "", "group-title", "hidden");
const starting = field("input", "startingNumber", "Filter Starting Number", "not-a-number", "text");
const controls = [name, filter, include, exclude, sensitive, type, starting];
const popup = new Element("div");
function descendants(node) { return node.children.reduce((all, child) => all.concat(child, descendants(child)), []); }
popup.querySelector = function(selector) { const match = selector.match(/\[name="([^"]+)"\]/); return match ? controls.find(control => control.getAttribute("name") === match[1]) || null : null; };
popup.querySelectorAll = function(selector) { if (selector === "input, select") return controls; if (selector === '[aria-invalid="true"]') return controls.filter(control => control.getAttribute("aria-invalid") === "true"); if (selector === ".tf-filter-field-error") return descendants(popup).filter(element => element.className === "tf-filter-field-error"); return []; };
const byID = id => [popup].concat(descendants(popup), controls.flatMap(control => [control.parentElement].concat(descendants(control.parentElement)))).find(element => element.id === id) || null;
const context = {console: {log() {}, warn() {}}, document: {createElement(tag) { return new Element(tag); }, getElementById(id) { return id === "popup-custom" ? popup : byID(id); }}};
vm.createContext(context);
vm.runInContext(fs.readFileSync(process.argv[2], "utf8"), context);
context.enhanceFilterPopup("group-title");
const valid = context.validateFilterPopup("group-title");
const error = byID("filter-name-error");
const summary = byID("filter-rule-summary");
const invalidName = name.value;
name.value = "Retained";
context.completeFilterRequest("saveFilter", {filter: {G1: {name: "Retained"}}}, {status: false, err: "transport fixture"});
process.stdout.write(JSON.stringify({valid, name: invalidName, starting: starting.value, focused: name.focusCount, invalid: name.getAttribute("aria-invalid") === "true", error: !!error && !error.hidden && error.textContent.length > 0, summary: summary && summary.textContent, popupClass: popup.className.split(/\s+/).includes("tf-filter-popup"), nameErrorAssociation: name.getAttribute("aria-errormessage"), ruleErrorAssociation: filter.getAttribute("aria-errormessage"), errorRetainedName: name.value, errorRetainedFocus: name.focusCount}));
`
