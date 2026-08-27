package src

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const runtimeHTMLInjectionPayload = `</span><img src=x onerror="globalThis.threadfinInjected=true"><script>globalThis.threadfinInjected=true</script>`

type runtimeValueSecurityResult struct {
	ProviderText  string   `json:"providerText"`
	MappingText   string   `json:"mappingText"`
	GroupText     string   `json:"groupText"`
	ProbeRows     []string `json:"probeRows"`
	PlaylistText  string   `json:"playlistText"`
	ClientText    string   `json:"clientText"`
	HeadlineText  string   `json:"headlineText"`
	TitleText     string   `json:"titleText"`
	Description   string   `json:"description"`
	DescriptionBR string   `json:"descriptionBR"`
	DangerousTags []string `json:"dangerousTags"`
	EventHandlers int      `json:"eventHandlers"`
	Injected      bool     `json:"injected"`
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
	})

	t.Run("TypeScript source contract", func(t *testing.T) {
		assignment := regexp.MustCompile(`\.innerHTML\s*=\s*(.+?)\s*;?\s*$`)
		emptyLiteral := regexp.MustCompile(`^(?:""|'')$`)
		for _, name := range []string{"network_ts.ts", "menu_ts.ts"} {
			content, err := os.ReadFile(filepath.Join("..", "ts", name))
			if err != nil {
				t.Fatal(err)
			}
			for index, line := range strings.Split(string(content), "\n") {
				match := assignment.FindStringSubmatch(strings.TrimSpace(line))
				if len(match) == 0 {
					continue
				}
				if !emptyLiteral.MatchString(strings.TrimSpace(match[1])) {
					t.Errorf("%s:%d assigns runtime or interpolated markup through innerHTML: %s", name, index+1, strings.TrimSpace(line))
				}
			}
		}
	})
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
    this.value = "";
    this.name = "";
    this.type = "";
    this.checked = false;
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
  addEventListener(name, listener) { this.listeners[name] = listener; }
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
  body: null,
  activeElement: null,
  cookie: "",
  createElement(tagName) { return new Element(tagName, this); },
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
changedName.value = payload;
changedGroup.value = payload;

runtimeContext = {
  console: {log() {}, warn() {}},
  document,
  window: {location: {protocol: "http:", hostname: "127.0.0.1", port: "34400"}},
  location: {reload() {}},
  WebSocket: class {},
  SERVER: {
    clientInfo: {activePlaylist: payload, totalPlaylist: payload, activeClients: payload, totalClients: payload},
    settings: {files: {xmltv: {X1: {name: payload}}}, epgCategoriesColors: ""},
    xepg: {epgMapping: {channel: {"x-active": true, "tvg-logo": "https://example.invalid/logo.png"}}},
  },
  UNDO: {},
  WS_AVAILABLE: false,
  BULK_EDIT: false,
  threadfinInjected: false,
  createSearchObj() {},
  searchInMapping() {},
  getObjKeys(value) { return Object.keys(value || {}); },
  showElement() {},
  showPreview() {},
  showPopUpElement() {},
  renderNavigation() {},
  restoreInitialDestinationFromHistory() {},
  alert() {},
  setInterval() {},
  setTimeout() { return 1; },
  clearTimeout() {},
};
runtimeContext.WebSocket.OPEN = 1;
vm.createContext(runtimeContext);
vm.runInContext(fs.readFileSync(process.argv[2], "utf8"), runtimeContext);
vm.runInContext(fs.readFileSync(process.argv[3], "utf8"), runtimeContext);

vm.runInContext('donePopupData("mapping", "channel")', runtimeContext);
const providerText = displays[6].textContent;
const mappingText = displays[3].textContent;
const groupText = displays[5].textContent;

runtimeContext.SERVER.clientInfo = {activePlaylist: payload, totalPlaylist: payload, activeClients: payload, totalClients: payload};
vm.runInContext('applyThreadfinResponse("probeChannel", {}, {status: true, probeInfo: {resolution: globalThis.payload, frameRate: globalThis.payload, audioChannel: globalThis.payload}, clientInfo: SERVER.clientInfo})', Object.assign(runtimeContext, {payload}));

vm.runInContext('globalThis.popupFixture = new PopupContent(); popupFixture.createHeadline(globalThis.payload); popupFixture.description(globalThis.payload); popupFixture.description(globalThis.payload + "<br>tail"); globalThis.titleFixture = popupFixture.createTitle(globalThis.payload)', runtimeContext);
const headline = popup.children.find(child => child.tagName == "H3");
const descriptions = walk(popup).filter(child => child.tagName == "PRE");
const description = descriptions[0];
const descriptionBR = descriptions[1];
const title = runtimeContext.titleFixture;
const roots = [document.body, title];
const nodes = roots.flatMap(root => walk(root, []));
const dangerousTags = nodes.filter(node => node.tagName == "SCRIPT" || (node.tagName == "IMG" && (node.getAttribute("src") == "x" || node.hasAttribute("onerror")))).map(node => node.tagName);
const eventHandlers = nodes.reduce((count, node) => count + Object.keys(node.attributes || {}).filter(name => name.startsWith("on")).length, 0);
const probeRows = document.getElementById("probeDetails").children.filter(child => child.tagName == "P").map(child => child.textContent);

process.stdout.write(JSON.stringify({
  providerText,
  mappingText,
  groupText,
  probeRows,
  playlistText: document.getElementById("playlist-connection-information").textContent,
  clientText: document.getElementById("client-connection-information").textContent,
  headlineText: headline ? headline.textContent : "",
  titleText: title ? title.textContent : "",
  description: description ? description.textContent : "",
  descriptionBR: descriptionBR ? descriptionBR.textContent : "",
  dangerousTags,
  eventHandlers,
  injected: parsedInjection || runtimeContext.threadfinInjected,
}));
`
