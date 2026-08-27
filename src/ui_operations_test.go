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

type operationsInteractionResult struct {
	ReadyEndpointKeys         []string `json:"readyEndpointKeys"`
	ReadyCopyLabels           []string `json:"readyCopyLabels"`
	CopiedValue               string   `json:"copiedValue"`
	CopyStatus                string   `json:"copyStatus"`
	PMSUnavailableKeys        []string `json:"pmsUnavailableKeys"`
	PMSUnavailableReasons     []string `json:"pmsUnavailableReasons"`
	ClientGuideSummaries      []string `json:"clientGuideSummaries"`
	ClientGuidesCollapsed     bool     `json:"clientGuidesCollapsed"`
	ConnectionsText           string   `json:"connectionsText"`
	PlaylistCapacityLabel     string   `json:"playlistCapacityLabel"`
	ClientCapacityLabel       string   `json:"clientCapacityLabel"`
	PlaylistCapacityAfter     string   `json:"playlistCapacityAfter"`
	ClientCapacityAfter       string   `json:"clientCapacityAfter"`
	ActiveTableIdentity       bool     `json:"activeTableIdentity"`
	InactiveTableIdentity     bool     `json:"inactiveTableIdentity"`
	ActiveTableFocusPreserved bool     `json:"activeTableFocusPreserved"`
	ActiveRows                []string `json:"activeRows"`
	InactiveRows              []string `json:"inactiveRows"`
}

type connectionsDVRResult struct {
	SelectorAvailable bool   `json:"selectorAvailable"`
	CardAvailable     bool   `json:"cardAvailable"`
	RenderedValue     string `json:"renderedValue"`
	CopyAction        bool   `json:"copyAction"`
	UnavailableText   string `json:"unavailableText"`
}

func TestConnectionsAcceptsBackendDVRAddressForms(t *testing.T) {
	tests := []struct {
		name      string
		address   string
		available bool
	}{
		{name: "bare hostname", address: "threadfin.example.com", available: true},
		{name: "bare IPv4", address: "192.0.2.10", available: true},
		{name: "hostname and port", address: "threadfin.example.com:34400", available: true},
		{name: "authentication prefix", address: "user:password@threadfin.example.com", available: true},
		{name: "bracketed IPv6", address: "[2001:db8::10]", available: true},
		{name: "whitespace", address: "threadfin example.com", available: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := evaluateConnectionsDVRFixture(t, test.address)
			if result.SelectorAvailable != test.available || result.CardAvailable != test.available {
				t.Fatalf("DVR %q availability = selector %t card %t, want %t", test.address, result.SelectorAvailable, result.CardAvailable, test.available)
			}
			if test.available {
				if result.RenderedValue != test.address || !result.CopyAction || result.UnavailableText != "" {
					t.Fatalf("available DVR %q rendered value %q copy=%t unavailable=%q", test.address, result.RenderedValue, result.CopyAction, result.UnavailableText)
				}
			} else if result.RenderedValue != "" || result.CopyAction || !strings.HasPrefix(result.UnavailableText, "Unavailable — ") {
				t.Fatalf("invalid DVR %q rendered value %q copy=%t unavailable=%q", test.address, result.RenderedValue, result.CopyAction, result.UnavailableText)
			}
		})
	}
}

func TestConnectionsAndActivityExecuteCurrentOperationsBehavior(t *testing.T) {
	result := evaluateOperationsInteraction(t)

	if strings.Join(result.ReadyEndpointKeys, ",") != "dvr,m3u,xmltv" {
		t.Fatalf("ready Connections endpoint keys = %v, want DVR, M3U, XMLTV only", result.ReadyEndpointKeys)
	}
	if strings.Join(result.ReadyCopyLabels, ",") != "Copy DVR address,Copy M3U URL,Copy XMLTV URL" {
		t.Fatalf("ready copy actions = %v", result.ReadyCopyLabels)
	}
	if result.CopiedValue != "127.0.0.1:34400" || result.CopyStatus != "DVR address copied." {
		t.Fatalf("Connections copy result = value %q status %q", result.CopiedValue, result.CopyStatus)
	}
	if strings.Join(result.PMSUnavailableKeys, ",") != "m3u,xmltv" {
		t.Fatalf("PMS unavailable endpoints = %v, want M3U and XMLTV", result.PMSUnavailableKeys)
	}
	for index, reason := range result.PMSUnavailableReasons {
		if !strings.HasPrefix(reason, "Unavailable — ") || strings.TrimSpace(strings.TrimPrefix(reason, "Unavailable — ")) == "" {
			t.Errorf("PMS unavailable reason %d = %q, want a visible unavailable status and explanatory reason", index, reason)
		}
	}
	if strings.Join(result.ClientGuideSummaries, ",") != "Plex,Jellyfin,Emby" || !result.ClientGuidesCollapsed {
		t.Fatalf("client notes = %v collapsed=%t", result.ClientGuideSummaries, result.ClientGuidesCollapsed)
	}
	lowerConnections := strings.ToLower(result.ConnectionsText)
	for _, claim := range []string{"configured successfully", "successfully configured", "connected successfully"} {
		if strings.Contains(lowerConnections, claim) {
			t.Errorf("Connections makes an unobservable downstream claim %q", claim)
		}
	}
	if result.PlaylistCapacityLabel != "Playlist source connections: 1 / 2" || result.ClientCapacityLabel != "Client connections: 1 / 4" {
		t.Fatalf("Activity capacity labels = playlist %q client %q", result.PlaylistCapacityLabel, result.ClientCapacityLabel)
	}
	if result.PlaylistCapacityAfter != "Playlist source connections: 2 / 3" || result.ClientCapacityAfter != "Client connections: 2 / 4" {
		t.Fatalf("Activity live capacity = playlist %q client %q", result.PlaylistCapacityAfter, result.ClientCapacityAfter)
	}
	if !result.ActiveTableIdentity || !result.InactiveTableIdentity || !result.ActiveTableFocusPreserved {
		t.Fatalf("Activity refresh replaced table state: active=%t inactive=%t focus=%t", result.ActiveTableIdentity, result.InactiveTableIdentity, result.ActiveTableFocusPreserved)
	}
	if strings.Join(result.ActiveRows, ",") != "News,Movies" || strings.Join(result.InactiveRows, ",") != "Weather" {
		t.Fatalf("Activity stream rows = active %v inactive %v", result.ActiveRows, result.InactiveRows)
	}
}

func TestOperationsMarkupKeepsLegacyIDsAndResponsiveOverflow(t *testing.T) {
	indexBytes, err := os.ReadFile(filepath.Join("..", "html", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	index := string(indexBytes)
	activityStart := strings.Index(index, `id="activity-content"`)
	activityEnd := strings.Index(index[activityStart:], `</section>`)
	if activityStart < 0 || activityEnd < 0 {
		t.Fatal("Activity host is missing")
	}
	activity := index[activityStart : activityStart+activityEnd]
	for _, id := range []string{"playlist-connection-information", "client-connection-information", "myStreamsBox", "activeStreams", "inactiveStreams"} {
		if !strings.Contains(activity, `id="`+id+`"`) {
			t.Errorf("Activity does not retain compatibility ID %q", id)
		}
	}
	for _, script := range []string{`src="js/connections_ts.js"`, `src="js/activity_ts.js"`} {
		if !strings.Contains(index, script) {
			t.Errorf("operations module is not loaded: %s", script)
		}
	}
	if strings.Contains(strings.ToLower(index), "gzip") {
		t.Fatal("operations markup invents an XMLTV GZIP response field")
	}
	styles, err := os.ReadFile(filepath.Join("..", "html", "css", "app-shell.css"))
	if err != nil {
		t.Fatal(err)
	}
	css := string(styles)
	for _, contract := range []string{
		".tf-app .tf-connections",
		".tf-app .tf-activity",
		".tf-app .tf-activity-table-scroll",
		"overflow-x: auto",
		"overflow-wrap: anywhere",
	} {
		if !strings.Contains(css, contract) {
			t.Errorf("responsive operations CSS is missing %q", contract)
		}
	}
}

func TestOperationsMarkupUsesConnectionsAsTheOnlyServerInformationEntry(t *testing.T) {
	indexBytes, err := os.ReadFile(filepath.Join("..", "html", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	index := string(indexBytes)
	for _, obsolete := range []string{`id="server_information"`, `id="server-information-link"`} {
		if strings.Contains(index, obsolete) {
			t.Errorf("obsolete Server Information entry remains in the application shell: %s", obsolete)
		}
	}
}

func TestOperationsGeneratedAssetsAreEmbeddedAndMatchSource(t *testing.T) {
	webUI = make(map[string]interface{})
	loadHTMLMap()

	for _, name := range []string{
		"html/index.html",
		"html/css/app-shell.css",
		"html/js/connections_ts.js",
		"html/js/activity_ts.js",
		"html/js/navigation_ts.js",
		"html/js/network_ts.js",
		"html/js/overview_ts.js",
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

func evaluateOperationsInteraction(t *testing.T) operationsInteractionResult {
	t.Helper()

	fixture := readyXEPGFixture()
	fixture["clientInfo"].(map[string]any)["activeClients"] = 1
	fixture["clientInfo"].(map[string]any)["activePlaylist"] = 1
	fixture["data"] = map[string]any{
		"StreamPreviewUI": map[string]any{
			"activeStreams":   []string{"News", "Movies"},
			"inactiveStreams": []string{"Weather"},
		},
	}
	fixtureJSON, err := json.Marshal(fixture)
	if err != nil {
		t.Fatal(err)
	}
	temp := t.TempDir()
	fixturePath := filepath.Join(temp, "fixture.json")
	if err := os.WriteFile(fixturePath, fixtureJSON, 0o600); err != nil {
		t.Fatal(err)
	}
	scriptPath := filepath.Join(temp, "operations-interaction.js")
	if err := os.WriteFile(scriptPath, []byte(operationsNodeScript), 0o600); err != nil {
		t.Fatal(err)
	}

	command := exec.Command(
		"node", scriptPath,
		filepath.Join("..", "html", "js", "app_state_ts.js"),
		filepath.Join("..", "html", "js", "overview_ts.js"),
		filepath.Join("..", "html", "js", "connections_ts.js"),
		filepath.Join("..", "html", "js", "activity_ts.js"),
		filepath.Join("..", "html", "js", "menu_ts.js"),
		filepath.Join("..", "html", "js", "navigation_ts.js"),
		filepath.Join("..", "html", "js", "network_ts.js"),
		fixturePath,
	)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("execute generated Connections/Activity sequence: %v\n%s", err, output)
	}

	var result operationsInteractionResult
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("decode Connections/Activity result: %v\n%s", err, output)
	}
	return result
}

func evaluateConnectionsDVRFixture(t *testing.T, address string) connectionsDVRResult {
	t.Helper()

	fixture := readyXEPGFixture()
	fixture["clientInfo"].(map[string]any)["DVR"] = address
	fixtureJSON, err := json.Marshal(fixture)
	if err != nil {
		t.Fatal(err)
	}
	temp := t.TempDir()
	fixturePath := filepath.Join(temp, "fixture.json")
	if err := os.WriteFile(fixturePath, fixtureJSON, 0o600); err != nil {
		t.Fatal(err)
	}
	scriptPath := filepath.Join(temp, "connections-dvr.js")
	if err := os.WriteFile(scriptPath, []byte(connectionsDVRNodeScript), 0o600); err != nil {
		t.Fatal(err)
	}

	command := exec.Command(
		"node", scriptPath,
		filepath.Join("..", "html", "js", "app_state_ts.js"),
		filepath.Join("..", "html", "js", "overview_ts.js"),
		filepath.Join("..", "html", "js", "connections_ts.js"),
		fixturePath,
	)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("execute generated DVR selector/Connections fixture: %v\n%s", err, output)
	}

	var result connectionsDVRResult
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("decode DVR selector/Connections result: %v\n%s", err, output)
	}
	return result
}

const connectionsDVRNodeScript = `
const fs = require("fs");
const vm = require("vm");

class Element {
  constructor(tagName, document) {
    this.tagName = tagName.toUpperCase();
    this.ownerDocument = document;
    this.children = [];
    this.attributes = {};
    this.listeners = {};
    this.className = "";
    this.textContent = "";
    this._id = "";
    this.classList = { contains: value => this.className.split(/\s+/).includes(value) };
  }
  get id() { return this._id; }
  set id(value) { this._id = String(value); if (this._id) this.ownerDocument.byID[this._id] = this; }
  get firstChild() { return this.children.length ? this.children[0] : null; }
  appendChild(child) { child.parentNode = this; this.children.push(child); return child; }
  removeChild(child) { const index = this.children.indexOf(child); if (index >= 0) this.children.splice(index, 1); return child; }
  setAttribute(name, value) { this.attributes[name] = String(value); if (name == "id") this.id = value; }
  getAttribute(name) { return this.attributes[name] === undefined ? null : this.attributes[name]; }
  addEventListener(name, listener) { this.listeners[name] = listener; }
}

const document = {
  byID: {},
  body: null,
  createElement(tagName) { return new Element(tagName, this); },
  getElementById(id) { return this.byID[id] || null; },
  execCommand() { return false; },
};
document.body = new Element("body", document);
const host = document.createElement("section");
host.id = "connections-content";
document.body.appendChild(host);

function walk(root, predicate, matches = []) {
  if (predicate(root)) matches.push(root);
  for (const child of root.children) walk(child, predicate, matches);
  return matches;
}

const fixture = JSON.parse(fs.readFileSync(process.argv[5], "utf8"));
const context = {
  document,
  navigator: { clipboard: { writeText() { return Promise.resolve(); } } },
  ClipboardJS: undefined,
};
vm.createContext(context);
for (let index = 2; index <= 4; index++) vm.runInContext(fs.readFileSync(process.argv[index], "utf8"), context);

const state = vm.runInContext("selectOverviewState", context)(fixture);
vm.runInContext("renderConnections", context)(fixture);
const endpoint = state.outputs.endpoints.find(endpoint => endpoint.key == "dvr");
const card = walk(host, element => element.getAttribute("data-endpoint") == "dvr")[0];
const value = walk(card, element => element.tagName == "CODE")[0];
const copy = walk(card, element => element.tagName == "BUTTON")[0];
const unavailable = walk(card, element => element.classList.contains("tf-endpoint-unavailable"))[0];
process.stdout.write(JSON.stringify({
  selectorAvailable: endpoint.available,
  cardAvailable: card.getAttribute("data-available") == "true",
  renderedValue: value ? value.textContent : "",
  copyAction: Boolean(copy),
  unavailableText: unavailable ? unavailable.textContent : "",
}));
`

const operationsNodeScript = `
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
    this.open = false;
    this.style = {};
    this.className = "";
    this.textContent = "";
    this.type = "";
    this.value = "";
    this._id = "";
    this._innerHTML = "";
    this.classList = { contains: value => this.className.split(/\s+/).includes(value) };
  }
  get id() { return this._id; }
  set id(value) {
    this._id = String(value);
    if (this._id) this.ownerDocument.byID[this._id] = this;
  }
  get innerHTML() { return this._innerHTML; }
  set innerHTML(value) {
    this._innerHTML = String(value);
    this.textContent = String(value).replace(/<[^>]*>/g, "");
    this.children = [];
  }
  get innerText() { return this.textContent; }
  set innerText(value) { this.textContent = String(value); }
  get firstChild() { return this.children.length ? this.children[0] : null; }
  appendChild(child) {
    if (child.parentNode) {
      const previousIndex = child.parentNode.children.indexOf(child);
      if (previousIndex >= 0) child.parentNode.children.splice(previousIndex, 1);
    }
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
    if (name == "class") this.className = String(value);
  }
  getAttribute(name) { return this.attributes[name] === undefined ? null : this.attributes[name]; }
  removeAttribute(name) { delete this.attributes[name]; }
  addEventListener(name, listener) {
    if (!this.listeners[name]) this.listeners[name] = [];
    this.listeners[name].push(listener);
  }
  click() {
    const event = { preventDefault() {}, clearSelection() {} };
    for (const listener of this.listeners.click || []) listener(event);
  }
  focus() { this.ownerDocument.activeElement = this; }
  select() {}
}

class TestDocument {
  constructor() {
    this.byID = {};
    this.cookie = "";
    this.activeElement = null;
    this.body = new TestElement("body", this);
    this.listeners = {};
  }
  createElement(tagName) { return new TestElement(tagName, this); }
  getElementById(id) { return this.byID[id] || null; }
  querySelector(selector) {
    if (selector == ".tf-nav-toggle") return null;
    return null;
  }
  querySelectorAll() { return []; }
  addEventListener(name, listener) { this.listeners[name] = listener; }
  execCommand(command) { return command == "copy"; }
}

function append(document, parent, tagName, id) {
  const element = document.createElement(tagName);
  if (id) element.id = id;
  parent.appendChild(element);
  return element;
}

function walk(root, predicate, matches = []) {
  if (predicate(root)) matches.push(root);
  for (const child of root.children) walk(child, predicate, matches);
  return matches;
}

function visibleText(root) {
  return walk(root, () => true).map(element => element.textContent).filter(Boolean).join(" ");
}

function endpointKeys(host, available) {
  return walk(host, element => element.getAttribute("data-endpoint") && element.getAttribute("data-available") == String(available))
    .map(element => element.getAttribute("data-endpoint"));
}

function tableValues(table) {
  return walk(table, element => element.className == "tdVal").map(element => element.textContent);
}

const document = new TestDocument();
append(document, document.body, "main", "main-content");
append(document, document.body, "section", "overview-content");
const connectionsHost = append(document, document.body, "section", "connections-content");
const contentHost = append(document, document.body, "section", "content");
const activityHost = append(document, document.body, "section", "activity-content");
append(document, activityHost, "p", "playlist-connection-information");
append(document, activityHost, "p", "client-connection-information");
const streamsBox = append(document, activityHost, "div", "myStreamsBox");
const activeTable = append(document, streamsBox, "table", "activeStreams");
const inactiveTable = append(document, streamsBox, "table", "inactiveStreams");
append(document, document.body, "nav", "main-menu");

let copiedValue = "";
let lastSocket = null;
class TestWebSocket {
  constructor(url) { this.url = url; lastSocket = this; }
  send(value) { this.sent = value; }
}

const context = {
  console: { log() {}, warn() {} },
  document,
  navigator: { clipboard: { writeText(value) { copiedValue = value; return Promise.resolve(); } } },
  ClipboardJS: undefined,
  WebSocket: TestWebSocket,
  SERVER: JSON.parse(fs.readFileSync(process.argv[9], "utf8")),
  SERVER_CONNECTION: false,
  WS_AVAILABLE: false,
  UNDO: {},
  menuItems: [],
  showElement() {},
  showLogs() {},
  enableGroupSelection() {},
  getObjKeys(value) { return Object.keys(value || {}); },
  alert() {},
  setInterval() { return 0; },
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
context.window.addEventListener = function() {};
context.window.scrollTo = function() {};
context.location = context.window.location;
vm.createContext(context);
for (let index = 2; index <= 8; index++) {
  vm.runInContext(fs.readFileSync(process.argv[index], "utf8"), context);
}

(async function() {
  vm.runInContext("renderConnections(SERVER)", context);
  const readyEndpointKeys = endpointKeys(connectionsHost, true);
  const readyCopyButtons = walk(connectionsHost, element => element.tagName == "BUTTON" && element.classList.contains("tf-copy-action"));
  const readyCopyLabels = readyCopyButtons.map(button => button.textContent);
  readyCopyButtons[0].click();
  await Promise.resolve();
  await Promise.resolve();
  const copyStatus = document.getElementById("connections-copy-status").textContent;

  context.SERVER.settings.epgSource = "PMS";
  context.SERVER.clientInfo.epgSource = "PMS";
  vm.runInContext("renderConnections(SERVER)", context);
  const pmsUnavailableCards = walk(connectionsHost, element => element.getAttribute("data-endpoint") && element.getAttribute("data-available") == "false");
  const pmsUnavailableKeys = pmsUnavailableCards.map(card => card.getAttribute("data-endpoint"));
  const pmsUnavailableReasons = pmsUnavailableCards.map(card => {
    const reason = walk(card, element => element.classList.contains("tf-endpoint-unavailable"))[0];
    return reason ? reason.textContent : "";
  });
  const details = walk(connectionsHost, element => element.tagName == "DETAILS");
  const clientGuideSummaries = details.map(detail => walk(detail, element => element.tagName == "SUMMARY")[0].textContent);
  const clientGuidesCollapsed = details.every(detail => !detail.open);
  const connectionsText = visibleText(connectionsHost);

  context.SERVER.settings.epgSource = "XEPG";
  context.SERVER.clientInfo.epgSource = "XEPG";
  vm.runInContext('openDestination("activity", false)', context);
  const playlistCapacityLabel = document.getElementById("playlist-connection-information").textContent;
  const clientCapacityLabel = document.getElementById("client-connection-information").textContent;
  const activeReference = document.getElementById("activeStreams");
  const inactiveReference = document.getElementById("inactiveStreams");
  activeReference.focus();

  vm.runInContext('new Server("updateLog").request({})', context);
  lastSocket.onopen.call(lastSocket);
  lastSocket.onmessage.call(lastSocket, { data: JSON.stringify({
    status: true,
    clientInfo: { activeClients: 2, activePlaylist: 2, totalPlaylist: 3 },
    log: { warnings: 1 },
  }) });

  const result = {
    readyEndpointKeys,
    readyCopyLabels,
    copiedValue,
    copyStatus,
    pmsUnavailableKeys,
    pmsUnavailableReasons,
    clientGuideSummaries,
    clientGuidesCollapsed,
    connectionsText,
    playlistCapacityLabel,
    clientCapacityLabel,
    playlistCapacityAfter: document.getElementById("playlist-connection-information").textContent,
    clientCapacityAfter: document.getElementById("client-connection-information").textContent,
    activeTableIdentity: document.getElementById("activeStreams") === activeReference,
    inactiveTableIdentity: document.getElementById("inactiveStreams") === inactiveReference,
    activeTableFocusPreserved: document.activeElement === activeReference,
    activeRows: tableValues(activeReference),
    inactiveRows: tableValues(inactiveReference),
  };
  process.stdout.write(JSON.stringify(result));
})().catch(error => {
  process.stderr.write(String(error && error.stack || error));
  process.exit(1);
});
`
