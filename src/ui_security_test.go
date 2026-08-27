package src

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

const runtimeHTMLInjectionPayload = `</span><img src=x onerror="globalThis.threadfinInjected=true"><script>globalThis.threadfinInjected=true</script>`
const runtimeDoubleQuoteHandlerPayload = `");globalThis.threadfinInjected=true;//`
const runtimeSingleQuoteHandlerPayload = `');globalThis.threadfinInjected=true;//`
const runtimeHandlerFixtureID = "handler-channel"

type runtimeValueSecurityResult struct {
	ProviderText  string     `json:"providerText"`
	MappingText   string     `json:"mappingText"`
	GroupText     string     `json:"groupText"`
	ProbeRows     []string   `json:"probeRows"`
	PlaylistText  string     `json:"playlistText"`
	ClientText    string     `json:"clientText"`
	HeadlineText  string     `json:"headlineText"`
	TitleText     string     `json:"titleText"`
	Description   string     `json:"description"`
	DescriptionBR string     `json:"descriptionBR"`
	ChannelStart  string     `json:"channelStart"`
	MappingValue  string     `json:"mappingValue"`
	BackupValue   string     `json:"backupValue"`
	HideValue     string     `json:"hideValue"`
	MappingDetail string     `json:"mappingDetail"`
	PPVTitle      string     `json:"ppvTitle"`
	PreviewTitles []string   `json:"previewTitles"`
	PreviewValues []string   `json:"previewValues"`
	HandlerAttrs  []string   `json:"handlerAttrs"`
	ProbeValues   []string   `json:"probeValues"`
	XMLTVCalls    [][]string `json:"xmltvCalls"`
	ToggleIDs     []string   `json:"toggleIds"`
	LogoIDs       []string   `json:"logoIds"`
	DoneCalls     [][]string `json:"doneCalls"`
	ChangedStates []bool     `json:"changedStates"`
	ProviderAttrs []string   `json:"providerAttrs"`
	ProviderCalls [][]string `json:"providerCalls"`
	ProviderData  [][]string `json:"providerData"`
	DangerousTags []string   `json:"dangerousTags"`
	EventHandlers int        `json:"eventHandlers"`
	Injected      bool       `json:"injected"`
}

func TestGeneratedRuntimeValuesRemainText(t *testing.T) {
	t.Run("compiled renderers", func(t *testing.T) {
		result := evaluateRuntimeValueSecurity(t)
		payload := runtimeHTMLInjectionPayload

		expected := map[string]string{
			"provider display":  result.ProviderText,
			"mapping display":   result.MappingText,
			"group display":     result.GroupText,
			"playlist capacity": result.PlaylistText,
			"client capacity":   result.ClientText,
			"popup headline":    result.HeadlineText,
			"popup title":       result.TitleText,
			"popup description": result.Description,
		}
		want := map[string]string{
			"provider display":  payload,
			"mapping display":   payload,
			"group display":     payload,
			"playlist capacity": "Playlist Connections: " + payload + " / " + payload,
			"client capacity":   "Client Connections: " + payload + " / " + payload,
			"popup headline":    `</SPAN><IMG SRC=X ONERROR="GLOBALTHIS.THREADFININJECTED=TRUE"><SCRIPT>GLOBALTHIS.THREADFININJECTED=TRUE</SCRIPT>`,
			"popup title":       payload + ":",
			"popup description": payload,
		}
		for name, got := range expected {
			if got != want[name] {
				t.Errorf("%s text = %q, want exact runtime value %q", name, got, want[name])
			}
		}

		wantProbeRows := []string{
			"Resolution: " + payload,
			"Frame Rate: " + payload + " FPS",
			"Audio: " + payload,
		}
		if strings.Join(result.ProbeRows, "\n") != strings.Join(wantProbeRows, "\n") {
			t.Errorf("probe rows = %#v, want %#v", result.ProbeRows, wantProbeRows)
		}
		if len(result.DangerousTags) != 0 {
			t.Errorf("runtime display values created executable elements: %v", result.DangerousTags)
		}
		if result.EventHandlers != 0 {
			t.Errorf("runtime display values created %d event-handler attributes", result.EventHandlers)
		}
		if result.Injected {
			t.Error("runtime display value executed its injection marker")
		}
		if want := payload + "<br>tail"; result.DescriptionBR != want {
			t.Errorf("runtime description with markup-like text = %q, want %q", result.DescriptionBR, want)
		}
		for name, got := range map[string]string{
			"channel-start display": result.ChannelStart,
			"mapping-value display": result.MappingValue,
			"backup display":        result.BackupValue,
			"hide display":          result.HideValue,
			"PPV title":             result.PPVTitle,
		} {
			want := payload
			if name == "PPV title" {
				want += ":"
			}
			if got != want {
				t.Errorf("%s = %q, want exact runtime value %q", name, got, want)
			}
		}
		if want := payload + " (" + payload + ")"; result.MappingDetail != want {
			t.Errorf("mapping TVG/EPG detail = %q, want %q", result.MappingDetail, want)
		}
		if strings.Join(result.PreviewTitles, ",") != "Active Streams,Inactive Streams" {
			t.Errorf("preview captions = %v, want exact repository captions", result.PreviewTitles)
		}
		if strings.Join(result.PreviewValues, "\n") != payload+"\n"+payload {
			t.Errorf("preview values = %v, want exact runtime text", result.PreviewValues)
		}
		if len(result.HandlerAttrs) != 0 {
			t.Errorf("mapping controls retain executable inline handlers: %v", result.HandlerAttrs)
		}
		if strings.Join(result.ProbeValues, "\n") != runtimeDoubleQuoteHandlerPayload {
			t.Errorf("probe listener values = %q, want inert exact provider URL", result.ProbeValues)
		}
		if len(result.XMLTVCalls) < 6 {
			t.Fatalf("XMLTV/mapping listener calls = %v, want initial mapping, three backups, file change, and replacement mapping", result.XMLTVCalls)
		}
		for _, call := range result.XMLTVCalls {
			if len(call) != 3 || call[0] != runtimeHandlerFixtureID || call[2] != runtimeSingleQuoteHandlerPayload {
				t.Errorf("XMLTV/mapping listener call = %v, want inert id/file payloads", call)
			}
		}
		if strings.Join(result.ToggleIDs, "\n") != runtimeHandlerFixtureID || strings.Join(result.LogoIDs, "\n") != runtimeHandlerFixtureID {
			t.Errorf("mapping ID listeners = toggle %v logo %v, want inert exact IDs", result.ToggleIDs, result.LogoIDs)
		}
		if len(result.DoneCalls) != 1 || len(result.DoneCalls[0]) != 2 || result.DoneCalls[0][0] != "mapping" || result.DoneCalls[0][1] != runtimeHandlerFixtureID {
			t.Errorf("Done listener calls = %v, want mapping and inert exact ID", result.DoneCalls)
		}
		for index, changed := range result.ChangedStates {
			if !changed {
				t.Errorf("mapping listener %d did not preserve its changed-state side effect", index)
			}
		}
		if len(result.ProviderAttrs) != 0 {
			t.Errorf("provider controls retain executable inline handlers: %v", result.ProviderAttrs)
		}
		wantProviderCalls := [][]string{
			{"m3u", runtimeDoubleQuoteHandlerPayload, "true", "0"},
			{"m3u", runtimeDoubleQuoteHandlerPayload, "false", "1"},
			{"m3u", runtimeDoubleQuoteHandlerPayload, "false", "0"},
			{"hdhr", runtimeDoubleQuoteHandlerPayload, "true", "0"},
			{"hdhr", runtimeDoubleQuoteHandlerPayload, "false", "1"},
			{"hdhr", runtimeDoubleQuoteHandlerPayload, "false", "0"},
			{"xmltv", runtimeDoubleQuoteHandlerPayload, "true", "0"},
			{"xmltv", runtimeDoubleQuoteHandlerPayload, "false", "1"},
			{"xmltv", runtimeDoubleQuoteHandlerPayload, "false", "0"},
			{"filter", runtimeDoubleQuoteHandlerPayload, "true", "0"},
			{"filter", runtimeDoubleQuoteHandlerPayload, "false", "0"},
			{"users", runtimeDoubleQuoteHandlerPayload, "true", "0"},
			{"users", runtimeDoubleQuoteHandlerPayload, "false", "undefined"},
		}
		if !reflect.DeepEqual(result.ProviderCalls, wantProviderCalls) {
			t.Errorf("provider listener calls = %#v, want exact inert arguments %#v", result.ProviderCalls, wantProviderCalls)
		}
		wantProviderData := [][]string{
			{"m3u", runtimeHTMLInjectionPayload, runtimeSingleQuoteHandlerPayload},
			{"hdhr", runtimeHTMLInjectionPayload, runtimeSingleQuoteHandlerPayload},
			{"xmltv", runtimeHTMLInjectionPayload, runtimeSingleQuoteHandlerPayload},
			{"custom-filter", runtimeHTMLInjectionPayload, ""},
			{"users", runtimeHTMLInjectionPayload, ""},
		}
		if !reflect.DeepEqual(result.ProviderData, wantProviderData) {
			t.Errorf("provider popup runtime input values = %#v, want exact inert data %#v", result.ProviderData, wantProviderData)
		}
	})

	t.Run("TypeScript source contract", func(t *testing.T) {
		for _, name := range []string{"network_ts.ts", "menu_ts.ts"} {
			content, err := os.ReadFile(filepath.Join("..", "ts", name))
			if err != nil {
				t.Fatal(err)
			}
			for _, violation := range runtimeDOMSinkViolations(string(content)) {
				t.Errorf("%s: %s", name, violation)
			}
		}
	})

	t.Run("source contract rejects multiline and indirect handler source", func(t *testing.T) {
		mutations := map[string]struct {
			source string
			want   string
		}{
			"multiline setAttribute concatenation": {source: `button.setAttribute(
  "onclick",
  "javascript: probeChannel('" + providerURL + "')"
)`, want: "supplies computed source"},
			"indirect setAttribute source": {source: `const handlerSource = "javascript: probeChannel('" + providerURL + "')"
button.setAttribute("onclick", handlerSource)`, want: "supplies computed source"},
			"multiline event property source": {source: `button.onclick =
  "javascript: probeChannel('" + providerURL + "')"`, want: "assigns executable source text"},
			"indirect event property source": {source: `const handlerSource = "javascript: probeChannel('" + providerURL + "')"
button.onclick = handlerSource`, want: "assigns indirect string source"},
		}
		for name, mutation := range mutations {
			t.Run(name, func(t *testing.T) {
				violations := strings.Join(runtimeDOMSinkViolations(mutation.source), "\n")
				if !strings.Contains(violations, mutation.want) {
					t.Fatalf("source contract violations = %q, want %q", violations, mutation.want)
				}
			})
		}
	})
}

func runtimeDOMSinkViolations(source string) []string {
	var violations []string
	lineNumber := func(offset int) int {
		return strings.Count(source[:offset], "\n") + 1
	}

	forbiddenHTMLSink := regexp.MustCompile(`(?:\.outerHTML\s*=|\.insertAdjacentHTML\s*\(|document\.(?:write|writeln)\s*\()`)
	for _, location := range forbiddenHTMLSink.FindAllStringIndex(source, -1) {
		violations = append(violations, fmt.Sprintf("line %d uses an executable HTML parsing sink", lineNumber(location[0])))
	}

	assignment := regexp.MustCompile(`\.innerHTML\s*=\s*(.+?)\s*;?\s*$`)
	emptyLiteral := regexp.MustCompile(`^(?:""|'')$`)
	for index, line := range strings.Split(source, "\n") {
		trimmed := strings.TrimSpace(line)
		match := assignment.FindStringSubmatch(trimmed)
		if len(match) == 0 {
			if strings.Contains(trimmed, ".innerHTML") && strings.Contains(trimmed, "=") {
				violations = append(violations, fmt.Sprintf("line %d has an unrecognized innerHTML assignment", index+1))
			}
			continue
		}
		if !emptyLiteral.MatchString(strings.TrimSpace(match[1])) {
			violations = append(violations, fmt.Sprintf("line %d assigns runtime or interpolated markup through innerHTML", index+1))
		}
	}

	handlerStart := regexp.MustCompile(`setAttribute\(\s*["']on[A-Za-z]+["']\s*,`)
	handlerCall := regexp.MustCompile(`(?s)setAttribute\(\s*["']on[A-Za-z]+["']\s*,\s*(.*?)\)\s*;?\s*(?://[^\n]*)?(?:\n|$)`)
	literalHandlerSource := regexp.MustCompile(`(?s)^(?:"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*')$`)
	starts := handlerStart.FindAllStringIndex(source, -1)
	calls := handlerCall.FindAllStringSubmatchIndex(source, -1)
	if len(calls) != len(starts) {
		violations = append(violations, "has an unrecognized or unterminated inline event-handler construction")
	}
	for _, location := range calls {
		expression := strings.TrimSpace(source[location[2]:location[3]])
		if !literalHandlerSource.MatchString(expression) {
			violations = append(violations, fmt.Sprintf("line %d supplies computed source to an inline event handler", lineNumber(location[0])))
		}
	}

	stringEventProperty := regexp.MustCompile("(?s)\\.on[A-Za-z]+\\s*=\\s*(?:\\\"|'|`)")
	for _, location := range stringEventProperty.FindAllStringIndex(source, -1) {
		violations = append(violations, fmt.Sprintf("line %d assigns executable source text to an event property", lineNumber(location[0])))
	}

	stringAssignment := regexp.MustCompile("(?m)([A-Za-z_$][A-Za-z0-9_$]*)\\s*=\\s*(?:\\\"|'|`)")
	for _, match := range stringAssignment.FindAllStringSubmatch(source, -1) {
		indirectEventProperty := regexp.MustCompile(`\.on[A-Za-z]+\s*=\s*` + regexp.QuoteMeta(match[1]) + `\b`)
		if location := indirectEventProperty.FindStringIndex(source); location != nil {
			violations = append(violations, fmt.Sprintf("line %d assigns indirect string source to an event property", lineNumber(location[0])))
		}
	}

	return violations
}

func evaluateRuntimeValueSecurity(t *testing.T) runtimeValueSecurityResult {
	t.Helper()
	temp := t.TempDir()
	scriptPath := filepath.Join(temp, "runtime-value-security.js")
	if err := os.WriteFile(scriptPath, []byte(runtimeValueSecurityNodeScript), 0o600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(
		"node", scriptPath,
		filepath.Join("..", "html", "js", "menu_ts.js"),
		filepath.Join("..", "html", "js", "network_ts.js"),
		runtimeHTMLInjectionPayload,
		runtimeDoubleQuoteHandlerPayload,
		runtimeSingleQuoteHandlerPayload,
	)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("execute compiled runtime-value renderers: %v\n%s", err, output)
	}
	var result runtimeValueSecurityResult
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("decode runtime-value security result: %v\n%s", err, output)
	}
	return result
}

const runtimeValueSecurityNodeScript = `
const fs = require("fs");
const vm = require("vm");

let parsedInjection = false;
let runtimeContext = null;

class TextNode {
  constructor(value, document) {
    this.tagName = "#TEXT";
    this.ownerDocument = document;
    this.parentNode = null;
    this.attributes = {};
    this.children = [];
    this._text = String(value);
  }
  get textContent() { return this._text; }
  set textContent(value) { this._text = String(value); }
}

class Element {
  constructor(tagName, document) {
    this.tagName = String(tagName).toUpperCase();
    this.ownerDocument = document;
    this.parentNode = null;
    this.children = [];
    this.attributes = {};
    this.listeners = {};
    this.className = "";
    this.style = {};
    this._id = "";
    this._text = "";
    this._value = "";
    this.name = "";
    this.type = "";
    this.checked = false;
    this.selectedIndex = 0;
    this.classList = {
      add: value => { if (!this.className.split(/\s+/).includes(value)) this.className = (this.className + " " + value).trim(); },
      contains: value => this.className.split(/\s+/).includes(value),
    };
  }
  get id() { return this._id; }
  set id(value) {
    this._id = String(value);
    if (this._id) this.ownerDocument.byID[this._id] = this;
  }
  get childNodes() { return this.children; }
  get firstChild() { return this.children.length ? this.children[0] : null; }
  get parentElement() { return this.parentNode; }
  get options() { return this.children.filter(child => child.tagName == "OPTION"); }
  get value() { return this._value; }
  set value(value) {
    this._value = String(value);
    if (this.tagName == "SELECT") {
      const index = this.options.findIndex(option => option.value == this._value);
      if (index >= 0) this.selectedIndex = index;
    }
  }
  get textContent() { return this._text + this.children.map(child => child.textContent || "").join(""); }
  set textContent(value) { this.children = []; this._text = String(value); }
  get innerText() { return this.textContent; }
  set innerText(value) { this.textContent = value; }
  get innerHTML() { return this.textContent; }
  set innerHTML(value) {
    this.children = [];
    this._text = "";
    const markup = String(value);
    if (markup === "") return;
    parseHTML(this, markup);
  }
  appendChild(child) { child.parentNode = this; this.children.push(child); return child; }
  removeChild(child) { const index = this.children.indexOf(child); if (index >= 0) this.children.splice(index, 1); child.parentNode = null; return child; }
  setAttribute(name, value) {
    const normalized = String(name).toLowerCase();
    const text = String(value);
    this.attributes[normalized] = text;
    if (normalized == "id") this.id = text;
    if (normalized == "class") this.className = text;
    if (normalized == "name") this.name = text;
    if (normalized == "type") this.type = text;
    if (normalized == "value") this.value = text;
  }
  getAttribute(name) { const value = this.attributes[String(name).toLowerCase()]; return value === undefined ? null : value; }
  hasAttribute(name) { return this.getAttribute(name) !== null; }
  addEventListener(name, listener) {
    if (!this.listeners[name]) this.listeners[name] = [];
    this.listeners[name].push(listener);
  }
  dispatchEvent(event) {
    const current = event || {type: ""};
    current.target = this;
    current.currentTarget = this;
    for (const listener of this.listeners[current.type] || []) listener.call(this, current);
    const inline = this.getAttribute("on" + current.type);
    if (inline && runtimeContext) {
      runtimeContext.__eventTarget = this;
      const source = inline.replace(/^\s*javascript:\s*/i, "");
      try {
        vm.runInContext("(function(){" + source + "\n}).call(__eventTarget)", runtimeContext);
      } catch (_error) {
        runtimeContext.inlineHandlerErrors += 1;
      }
    }
    return true;
  }
  getElementsByClassName(name) { return walk(this).filter(item => item !== this && item.className.split(/\s+/).includes(name)); }
  querySelectorAll(selector) {
    const names = selector.split(",").map(value => value.trim().toUpperCase());
    return walk(this).filter(item => item !== this && names.includes(item.tagName));
  }
}

function parseHTML(root, markup) {
  const stack = [root];
  const tokenPattern = /<[^>]*>|[^<]+/g;
  let match;
  while ((match = tokenPattern.exec(markup)) !== null) {
    const token = match[0];
    if (!token.startsWith("<")) {
      stack[stack.length - 1].appendChild(new TextNode(token, root.ownerDocument));
      continue;
    }
    const closing = /^<\s*\/\s*([A-Za-z0-9-]+)/.exec(token);
    if (closing) {
      const name = closing[1].toUpperCase();
      for (let index = stack.length - 1; index > 0; index--) {
        if (stack[index].tagName == name) { stack.length = index; break; }
      }
      continue;
    }
    const opening = /^<\s*([A-Za-z0-9-]+)/.exec(token);
    if (!opening) continue;
    const element = root.ownerDocument.createElement(opening[1]);
    const attributes = token.slice(opening[0].length, token.length - 1);
    const attributePattern = /([^\s=/>]+)(?:\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s>]+)))?/g;
    let attribute;
    while ((attribute = attributePattern.exec(attributes)) !== null) {
      const value = attribute[2] === undefined ? (attribute[3] === undefined ? (attribute[4] === undefined ? "" : attribute[4]) : attribute[3]) : attribute[2];
      element.setAttribute(attribute[1], value);
      if (attribute[1].toLowerCase().startsWith("on")) {
        parsedInjection = true;
        if (runtimeContext) runtimeContext.threadfinInjected = true;
      }
    }
    stack[stack.length - 1].appendChild(element);
    if (element.tagName == "SCRIPT") {
      parsedInjection = true;
      if (runtimeContext) runtimeContext.threadfinInjected = true;
    }
    if (!["IMG", "BR", "INPUT", "META", "LINK"].includes(element.tagName) && !/\/\s*>$/.test(token)) stack.push(element);
  }
}

function walk(root, found = []) {
  found.push(root);
  for (const child of root.children || []) walk(child, found);
  return found;
}

const document = {
  byID: {},
  created: [],
  body: null,
  activeElement: null,
  cookie: "",
  createElement(tagName) { const element = new Element(tagName, this); this.created.push(element); return element; },
  createTextNode(value) { return new TextNode(value, this); },
  getElementById(id) { return this.byID[id] || null; },
  addEventListener() {},
};
document.body = document.createElement("body");

function append(parent, tagName, id) {
  const element = document.createElement(tagName);
  if (id) element.id = id;
  parent.appendChild(element);
  return element;
}

const popup = append(document.body, "div", "popup-custom");
const changedName = append(popup, "input");
changedName.className = "changed";
changedName.type = "text";
changedName.name = "x-name";
const changedGroup = append(popup, "input");
changedGroup.className = "changed";
changedGroup.type = "text";
changedGroup.name = "x-group-title";
const changedProvider = append(popup, "input");
changedProvider.className = "changed";
changedProvider.type = "text";
changedProvider.name = "x-xmltv-file";
changedProvider.value = "X1.xml";

const mappingRow = append(document.body, "tr", "channel");
const displays = [];
for (let index = 0; index < 8; index++) {
  const cell = append(mappingRow, "td");
  const display = append(cell, "span");
  displays.push(display);
  if (index == 2) append(display, "img");
}
append(document.body, "p", "playlist-connection-information");
append(document.body, "p", "client-connection-information");
append(document.body, "div", "probeDetails");

const payload = process.argv[4];
const handlerPayload = process.argv[5];
const singleQuoteHandlerPayload = process.argv[6];
const mappingID = "handler-channel";
const probeValues = [];
const xmltvCalls = [];
const toggleIDs = [];
const logoIDs = [];
const doneCalls = [];
const providerCalls = [];
const providerAttrs = [];
const providerData = [];
changedName.value = payload;
changedGroup.value = payload;

const providerFixtures = {
  m3u: {
    "id.provider": handlerPayload,
    name: payload,
    description: payload,
    "file.source": singleQuoteHandlerPayload,
    buffer: "-",
    tuner: "1",
    "http_proxy.ip": "",
    "http_proxy.port": "",
    "http_headers.origin": "",
    "http_headers.referer": "",
  },
  hdhr: {
    "id.provider": handlerPayload,
    name: payload,
    description: payload,
    "file.source": singleQuoteHandlerPayload,
    buffer: "-",
    tuner: "1",
    "http_proxy.ip": "",
    "http_proxy.port": "",
  },
  xmltv: {
    "id.provider": handlerPayload,
    name: payload,
    description: payload,
    "file.source": singleQuoteHandlerPayload,
    "http_proxy.ip": "",
    "http_proxy.port": "",
  },
  "custom-filter": {
    name: payload,
    description: payload,
    type: "custom-filter",
    caseSensitive: false,
    filter: payload,
    startingNumber: "1000",
    "x-category": "",
  },
  users: {
    username: payload,
    defaultUser: false,
    "authentication.web": true,
    "authentication.pms": true,
    "authentication.m3u": true,
    "authentication.xml": true,
    "authentication.api": true,
    "authentication.config": true,
  },
};

const mappingData = {
  "x-active": true,
  "x-name": payload,
  "tvg-id": payload,
  "x-epg": payload,
  "x-description": payload,
  "_uuid.key": "",
  "tvg-logo": "https://example.invalid/logo.png",
  "x-update-channel-icon": false,
  "x-category": "",
  "x-group-title": payload,
  "group-title": payload,
  "x-xmltv-file": singleQuoteHandlerPayload,
  "x-mapping": singleQuoteHandlerPayload,
  "x-backup-channel-1": singleQuoteHandlerPayload,
  "x-backup-channel-2": singleQuoteHandlerPayload,
  "x-backup-channel-3": singleQuoteHandlerPayload,
  "url": handlerPayload,
};

runtimeContext = {
  console: {log() {}, warn() {}},
  document,
  window: {location: {protocol: "http:", hostname: "127.0.0.1", port: "34400"}},
  location: {reload() {}},
  WebSocket: class {},
  SERVER: {
    clientInfo: {activePlaylist: payload, totalPlaylist: payload, activeClients: payload, totalClients: payload},
    settings: {files: {xmltv: {X1: {name: payload}}}, buffer: "-", epgCategories: "", epgCategoriesColors: ""},
    xepg: {
      epgMapping: {
        channel: {"x-active": true, "tvg-logo": "https://example.invalid/logo.png"},
        [mappingID]: mappingData,
      },
      xmltvMap: {
        [singleQuoteHandlerPayload]: {[singleQuoteHandlerPayload]: {"display-name": payload, icon: "https://example.invalid/guide.png"}},
      },
    },
    data: {StreamPreviewUI: {activeStreams: [payload], inactiveStreams: [payload]}},
  },
  UNDO: {},
  WS_AVAILABLE: false,
  BULK_EDIT: false,
  threadfinInjected: false,
  inlineHandlerErrors: 0,
  createSearchObj() {},
  searchInMapping() {},
  getObjKeys(value) { return Object.keys(value || {}); },
  getOwnObjProps(value) { return Object.keys(value || {}); },
  getAllSelectedChannels() { return []; },
  getLocalData(type, id) {
    if (type == "mapping" && id == mappingID) return mappingData;
    if (id == handlerPayload && providerFixtures[type]) return providerFixtures[type];
    return {};
  },
  enhanceSourcePopup() {},
  enhanceUsersPopup() {},
  showElement() {},
  showPreview() {},
  showPopUpElement() {},
  renderNavigation() {},
  restoreInitialDestinationFromHistory() {},
  alert() {},
  probeChannel(value) { probeValues.push(String(value)); },
  toggleChannelStatus(value) { toggleIDs.push(String(value)); },
  changeChannelNumbers() {},
  setInterval() {},
  setTimeout() { return 1; },
  clearTimeout() {},
};
runtimeContext.WebSocket.OPEN = 1;
vm.createContext(runtimeContext);
vm.runInContext(fs.readFileSync(process.argv[2], "utf8"), runtimeContext);
vm.runInContext(fs.readFileSync(process.argv[3], "utf8"), runtimeContext);
const mappingServer = runtimeContext.SERVER;

const realDonePopupData = runtimeContext.donePopupData;
function updateLegacyDisplay(name, value, displayIndex) {
  popup.innerHTML = "";
  const input = append(popup, "input");
  input.className = "changed";
  input.type = "text";
  input.name = name;
  input.value = value;
  realDonePopupData("mapping", "channel");
  return displays[displayIndex].textContent;
}
const providerText = updateLegacyDisplay("x-xmltv-file", "X1.xml", 6);
const mappingText = updateLegacyDisplay("x-name", payload, 3);
const groupText = updateLegacyDisplay("x-group-title", payload, 5);
const channelStart = updateLegacyDisplay("x-channel-start", payload, 3);
const mappingValue = updateLegacyDisplay("x-mapping", payload, 7);
const backupValue = updateLegacyDisplay("x-backup-channel", payload, 7);
const hideValue = updateLegacyDisplay("x-hide-channel", payload, 7);

runtimeContext.SERVER.clientInfo = {activePlaylist: payload, totalPlaylist: payload, activeClients: payload, totalClients: payload};
vm.runInContext('applyThreadfinResponse("probeChannel", {}, {status: true, probeInfo: {resolution: globalThis.payload, frameRate: globalThis.payload, audioChannel: globalThis.payload}, clientInfo: SERVER.clientInfo})', Object.assign(runtimeContext, {payload}));
const renderedProbeRows = document.getElementById("probeDetails").children.filter(child => child.tagName == "P").map(child => child.textContent);

vm.runInContext('globalThis.popupFixture = new PopupContent(); popupFixture.createHeadline(globalThis.payload); popupFixture.description(globalThis.payload); popupFixture.description(globalThis.payload + "<br>tail"); globalThis.titleFixture = popupFixture.createTitle(globalThis.payload)', runtimeContext);
const headline = popup.children.find(child => child.tagName == "H3");
const descriptions = walk(popup).filter(child => child.tagName == "PRE");
const description = descriptions[0];
const descriptionBR = descriptions[1];
const title = runtimeContext.titleFixture;

runtimeContext.SERVER = mappingServer;
const mappingInvoker = append(document.body, "button", mappingID);
runtimeContext.mappingInvoker = mappingInvoker;
runtimeContext.payload = payload;
vm.runInContext('openPopUp("mapping", mappingInvoker)', runtimeContext);
const mappingDetailElement = walk(popup).find(child => child.tagName == "PRE" && walk(child, []).some(item => item.className == "text-danger"));
const mappingDetail = mappingDetailElement ? mappingDetailElement.textContent : "";
const activeControl = document.getElementById("active");
const logoControl = document.getElementById("update-icon");
const xmltvFileControl = document.getElementById("popup-xmltv");
const initialMappingControl = document.getElementById("xmltv-id-picker-input");
const backupControls = [document.getElementById("backup-channel-1"), document.getElementById("backup-channel-2"), document.getElementById("backup-channel-3")];
const popupInputs = walk(popup).filter(item => item.tagName == "INPUT");
const probeControl = popupInputs.find(item => item.value == "{{.button.probeChannel}}");
const doneControl = popupInputs.find(item => item.value == "{{.button.done}}");
const handlerAttrs = [
  [activeControl, "onchange"],
  [logoControl, "onchange"],
  [xmltvFileControl, "onchange"],
  [initialMappingControl, "onchange"],
  ...backupControls.map(control => [control, "onchange"]),
  [probeControl, "onclick"],
  [doneControl, "onclick"],
].map(([control, attribute]) => control && control.getAttribute(attribute)).filter(value => value !== null && value !== undefined);

runtimeContext.checkXmltvChannel = function(id, element, file) {
  const value = element && typeof element == "object" && "value" in element ? element.value : element;
  xmltvCalls.push([String(id), String(value), String(file)]);
};
runtimeContext.changeChannelLogo = function(id) { logoIDs.push(String(id)); };
runtimeContext.donePopupData = function(type, ids) { doneCalls.push([String(type), String(ids)]); };

initialMappingControl.dispatchEvent({type: "change"});
for (const control of backupControls) control.dispatchEvent({type: "change"});
activeControl.dispatchEvent({type: "change"});
logoControl.dispatchEvent({type: "change"});
probeControl.dispatchEvent({type: "click"});
doneControl.dispatchEvent({type: "click"});
xmltvFileControl.value = singleQuoteHandlerPayload;
xmltvFileControl.dispatchEvent({type: "change"});
const replacementMappingControl = document.getElementById("xmltv-id-picker-input");
replacementMappingControl.value = singleQuoteHandlerPayload;
replacementMappingControl.dispatchEvent({type: "change"});

runtimeContext.savePopupData = function(type, id, remove, option) {
  providerCalls.push([String(type), String(id), String(remove), String(option)]);
};
function exerciseProviderPopup(dataType, buttonLabels) {
  const invoker = document.createElement("button");
  invoker.id = handlerPayload;
  document.body.appendChild(invoker);
  runtimeContext.openPopUp(dataType, invoker);

  const inputs = walk(popup).filter(item => item.tagName == "INPUT");
  const nameKey = dataType == "users" ? "username" : "name";
  const nameControl = inputs.find(item => item.name == nameKey);
  const sourceControl = inputs.find(item => item.name == "file.source");
  providerData.push([dataType, nameControl ? nameControl.value : "", sourceControl ? sourceControl.value : ""]);

  for (const label of buttonLabels) {
    const control = inputs.find(item => item.value == label);
    if (!control) throw new Error("missing real " + dataType + " popup control " + label);
    for (const [attribute, value] of Object.entries(control.attributes)) {
      if (attribute.startsWith("on")) providerAttrs.push(dataType + ":" + label + ":" + attribute + "=" + value);
    }
    control.dispatchEvent({type: "click"});
  }
}
exerciseProviderPopup("m3u", ["{{.button.delete}}", "{{.button.update}}", "{{.button.save}}"]);
exerciseProviderPopup("hdhr", ["{{.button.delete}}", "{{.button.update}}", "{{.button.save}}"]);
exerciseProviderPopup("xmltv", ["{{.button.delete}}", "{{.button.update}}", "{{.button.save}}"]);
exerciseProviderPopup("custom-filter", ["{{.button.delete}}", "{{.button.save}}"]);
exerciseProviderPopup("users", ["{{.button.delete}}", "{{.button.save}}"]);

const ppvHostCell = append(document.body, "td");
append(ppvHostCell, "input", "x-ppv-extra");
const ppvChoice = document.createElement("select");
ppvChoice.value = "PPV";
runtimeContext.ppvChoice = ppvChoice;
runtimeContext.ppvOwner = {table: document.createElement("table")};
const createdBeforePPV = document.created.length;
vm.runInContext('checkPPV.call(ppvOwner, globalThis.payload, ppvChoice)', runtimeContext);
const ppvTitleElement = document.created.slice(createdBeforePPV).find(item => item.tagName == "TD" && item.textContent.startsWith(payload + ":"));

const previewBox = append(document.body, "div", "myStreamsBox");
const activeStreams = append(previewBox, "table", "activeStreams");
const inactiveStreams = append(previewBox, "table", "inactiveStreams");
runtimeContext.SERVER.data = {StreamPreviewUI: {activeStreams: [payload], inactiveStreams: [payload]}};
vm.runInContext('showPreview(true)', runtimeContext);
const previewTables = [activeStreams, inactiveStreams];
const previewTitles = previewTables.map(table => table.children.find(child => child.tagName == "CAPTION").textContent);
const previewValues = previewTables.map(table => walk(table).find(child => child.className == "tdVal").textContent);

const roots = [document.body, title];
const nodes = roots.flatMap(root => walk(root, []));
const dangerousTags = nodes.filter(node => node.tagName == "SCRIPT" || (node.tagName == "IMG" && (node.getAttribute("src") == "x" || node.hasAttribute("onerror")))).map(node => node.tagName);
const eventHandlers = nodes.reduce((count, node) => count + Object.keys(node.attributes || {}).filter(name => name == "onerror").length, 0);

process.stdout.write(JSON.stringify({
  providerText,
  mappingText,
  groupText,
  probeRows: renderedProbeRows,
  playlistText: document.getElementById("playlist-connection-information").textContent,
  clientText: document.getElementById("client-connection-information").textContent,
  headlineText: headline ? headline.textContent : "",
  titleText: title ? title.textContent : "",
  description: description ? description.textContent : "",
  descriptionBR: descriptionBR ? descriptionBR.textContent : "",
  channelStart,
  mappingValue,
  backupValue,
  hideValue,
  mappingDetail,
  ppvTitle: ppvTitleElement ? ppvTitleElement.textContent : "",
  previewTitles,
  previewValues,
  handlerAttrs,
  probeValues,
  xmltvCalls,
  toggleIds: toggleIDs,
  logoIds: logoIDs,
  doneCalls,
  changedStates: [logoControl, xmltvFileControl, initialMappingControl, ...backupControls, replacementMappingControl].map(control => control.className.split(/\s+/).includes("changed")),
  providerAttrs,
  providerCalls,
  providerData,
  dangerousTags,
  eventHandlers,
  injected: parsedInjection || runtimeContext.threadfinInjected,
}));
`
