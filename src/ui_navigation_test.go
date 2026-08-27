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

type navigationShellResult struct {
	CollapsedInitially      bool   `json:"collapsedInitially"`
	ExpandedAfterToggle     bool   `json:"expandedAfterToggle"`
	StoredAfterToggle       string `json:"storedAfterToggle"`
	ToggleExpandedInitially string `json:"toggleExpandedInitially"`
	ToggleLabelAfterToggle  string `json:"toggleLabelAfterToggle"`
	DocumentationLabel      string `json:"documentationLabel"`
	DocumentationHref       string `json:"documentationHref"`
	DocumentationTarget     string `json:"documentationTarget"`
	DocumentationRel        string `json:"documentationRel"`
	DocumentationGroup      string `json:"documentationGroup"`
	ItemsWithIcons          int    `json:"itemsWithIcons"`
	NavigationItems         int    `json:"navigationItems"`
}

func TestNavigationRendersCollapsibleIconRailAndUserGuide(t *testing.T) {
	scriptPath := filepath.Join(t.TempDir(), "navigation-shell.js")
	script := `
const fs = require("fs"), vm = require("vm");

class ClassList {
  constructor(element) { this.element = element; }
  values() { return this.element.className.split(/\s+/).filter(Boolean); }
  contains(value) { return this.values().includes(value); }
  add(value) { if (!this.contains(value)) this.element.className = this.values().concat(value).join(" "); }
  remove(value) { this.element.className = this.values().filter(item => item !== value).join(" "); }
  toggle(value, force) {
    const add = force === undefined ? !this.contains(value) : force;
    if (add) this.add(value); else this.remove(value);
    return add;
  }
}

class Element {
  constructor(tagName, document) {
    this.tagName = tagName.toUpperCase(); this.ownerDocument = document; this.children = [];
    this.parentElement = null; this.attributes = {}; this.listeners = {}; this.className = ""; this.textContent = "";
    this.classList = new ClassList(this);
  }
  appendChild(child) { child.parentElement = this; this.children.push(child); if (child.id) this.ownerDocument.byID[child.id] = child; return child; }
  setAttribute(name, value) { this.attributes[name] = String(value); if (name === "id") { this.id = String(value); this.ownerDocument.byID[this.id] = this; } }
  getAttribute(name) { return this.attributes[name] === undefined ? null : this.attributes[name]; }
  removeAttribute(name) { delete this.attributes[name]; }
  addEventListener(name, listener) { this.listeners[name] = listener; }
  set innerHTML(value) { this.children = []; }
}

function all(root) { return [root].concat(...root.children.map(all)); }
class Document {
  constructor() { this.byID = {}; this.body = new Element("body", this); }
  createElement(tagName) { return new Element(tagName, this); }
  getElementById(id) { return this.byID[id] || null; }
  querySelector(selector) {
    if (selector === ".tf-app") return all(this.body).find(item => item.classList.contains("tf-app")) || null;
    if (selector === ".tf-sidebar-toggle") return all(this.body).find(item => item.classList.contains("tf-sidebar-toggle")) || null;
    if (selector === ".tf-nav-toggle") return all(this.body).find(item => item.classList.contains("tf-nav-toggle")) || null;
    return null;
  }
  querySelectorAll(selector) {
    if (selector === "#main-menu [data-destination]") return all(this.getElementById("main-menu")).filter(item => item.getAttribute("data-destination"));
    return [];
  }
}

const document = new Document();
function append(parent, tagName, id, className) {
  const element = document.createElement(tagName); element.className = className || "";
  if (id) element.setAttribute("id", id); parent.appendChild(element); return element;
}
const app = append(document.body, "div", "", "tf-app");
const sidebar = append(app, "aside", "", "tf-sidebar");
const toggle = append(sidebar, "button", "sidebar-rail-toggle", "tf-sidebar-toggle");
append(sidebar, "nav", "main-menu", "tf-navigation");

const stored = { "threadfin.navigation.collapsed": "true" };
const localStorage = {
  getItem(key) { return stored[key] === undefined ? null : stored[key]; },
  setItem(key, value) { stored[key] = String(value); },
};
const labels = ["Playlist", "XMLTV", "Filter", "Mapping", "Users", "Settings", "Log", "Logout"];
const context = {
  console, document, localStorage,
  SERVER: { settings: { "authentication.web": true } },
  menuItems: labels.map(value => ({ value })),
  setTimeout(callback) { callback(); },
  addEventListener() {},
};
context.window = context;
context.window.history = { state: null, pushState() {}, replaceState() {} };
context.window.location = { hash: "" };
vm.createContext(context);
vm.runInContext(fs.readFileSync(process.argv[2], "utf8"), context);
context.renderNavigation();

const navigation = document.getElementById("main-menu");
const navigationItems = all(navigation).filter(item => item.classList.contains("tf-navigation-item"));
const documentation = navigationItems.find(item => item.getAttribute("href"));
const documentationGroup = documentation.parentElement.parentElement.parentElement;
const groupHeading = documentationGroup.children.find(item => item.tagName === "H2");
const collapsedInitially = app.classList.contains("tf-sidebar-collapsed");
const toggleExpandedInitially = toggle.getAttribute("aria-expanded");
toggle.listeners.click();

process.stdout.write(JSON.stringify({
  collapsedInitially,
  expandedAfterToggle: !app.classList.contains("tf-sidebar-collapsed"),
  storedAfterToggle: stored["threadfin.navigation.collapsed"],
  toggleExpandedInitially,
  toggleLabelAfterToggle: toggle.getAttribute("aria-label"),
  documentationLabel: documentation.textContent || documentation.children.map(item => item.textContent).join(""),
  documentationHref: documentation.getAttribute("href"),
  documentationTarget: documentation.getAttribute("target"),
  documentationRel: documentation.getAttribute("rel"),
  documentationGroup: groupHeading.textContent,
  itemsWithIcons: navigationItems.filter(item => item.children.some(child => child.tagName === "I" && child.getAttribute("aria-hidden") === "true")).length,
  navigationItems: navigationItems.length,
}));
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o600); err != nil {
		t.Fatal(err)
	}

	command := exec.Command("node", scriptPath, filepath.Join("..", "html", "js", "navigation_ts.js"))
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("execute generated navigation shell: %v\n%s", err, output)
	}

	var result navigationShellResult
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("decode navigation shell result: %v\n%s", err, output)
	}
	if !result.CollapsedInitially || result.ToggleExpandedInitially != "false" {
		t.Fatalf("stored collapsed navigation = collapsed %t, aria-expanded %q; want true and false", result.CollapsedInitially, result.ToggleExpandedInitially)
	}
	if !result.ExpandedAfterToggle || result.StoredAfterToggle != "false" || result.ToggleLabelAfterToggle != "Collapse navigation" {
		t.Fatalf("expanded navigation = expanded %t, stored %q, label %q; want true, false, Collapse navigation", result.ExpandedAfterToggle, result.StoredAfterToggle, result.ToggleLabelAfterToggle)
	}
	if result.DocumentationLabel != "User guide" || result.DocumentationGroup != "System" {
		t.Fatalf("documentation item = %q in %q; want User guide in System", result.DocumentationLabel, result.DocumentationGroup)
	}
	if result.DocumentationHref != "https://github.com/volschin/Threadfin/blob/main/docs/user-guide.md" || result.DocumentationTarget != "_blank" || result.DocumentationRel != "noopener noreferrer" {
		t.Fatalf("documentation link = href %q, target %q, rel %q", result.DocumentationHref, result.DocumentationTarget, result.DocumentationRel)
	}
	if result.NavigationItems != 12 || result.ItemsWithIcons != result.NavigationItems {
		t.Fatalf("navigation items = %d, with icons = %d; want all 12 items iconized", result.NavigationItems, result.ItemsWithIcons)
	}
}

func TestLegacyMenuKeysKeepTheirBackendOrder(t *testing.T) {
	base := readUITypeScript(t, "base_ts.ts")
	legacyMenuItem := regexp.MustCompile(`menuItems\.push\(new MainMenuItem\("([^"]+)"`)
	matches := legacyMenuItem.FindAllStringSubmatch(base, -1)

	var got []string
	for _, match := range matches {
		got = append(got, match[1])
	}
	want := []string{"playlist", "xmltv", "filter", "mapping", "users", "settings", "log", "logout"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("legacy menu key order = %v, want %v", got, want)
	}
}

func TestNumericMenuResponsesUseTheExplicitLegacyAdapter(t *testing.T) {
	navigation := readUITypeScript(t, "navigation_ts.ts")
	if !strings.Contains(navigation, "function openLegacyMenu(index: number") {
		t.Fatal("navigation is missing the explicit openLegacyMenu(index) adapter")
	}
	if !strings.Contains(navigation, "new ShowContent(index)") {
		t.Fatal("openLegacyMenu does not route numeric legacy indexes through ShowContent")
	}

	network := readUITypeScript(t, "network_ts.ts")
	if !strings.Contains(network, `openLegacyMenu(response["openMenu"])`) {
		t.Fatal("numeric openMenu response does not use openLegacyMenu")
	}
	if strings.Contains(network, `document.getElementById(response["openMenu"])`) {
		t.Fatal("numeric openMenu response still selects a navigation element directly")
	}
}

func TestGroupedNavigationControlsUseScopedShellStyles(t *testing.T) {
	styles, err := os.ReadFile(filepath.Join("..", "html", "css", "app-shell.css"))
	if err != nil {
		t.Fatal(err)
	}
	css := string(styles)
	for _, selector := range []string{
		".tf-app .tf-navigation-group",
		".tf-app .tf-navigation-group-label",
		".tf-app .tf-navigation-list",
		".tf-app .tf-navigation-item",
		".tf-app .tf-navigation-item:hover",
		".tf-app .tf-navigation-item[aria-current=\"page\"]",
	} {
		if !strings.Contains(css, selector) {
			t.Errorf("grouped navigation stylesheet is missing %s", selector)
		}
	}
	if strings.Contains(css, ".tf-navigation > li") {
		t.Error("grouped navigation stylesheet still depends on direct legacy li children")
	}
}

func TestOpenMenuResponseRoutesOnceWithoutHistoryRestore(t *testing.T) {
	network := readUITypeScript(t, "network_ts.ts")
	if got := strings.Count(network, `openLegacyMenu(response["openMenu"])`); got != 1 {
		t.Fatalf("openMenu adapter calls = %d, want exactly one", got)
	}

	menu := readUITypeScript(t, "menu_ts.ts")
	createLayout := menu[strings.Index(menu, "function createLayout()"):strings.Index(menu, "class PopupWindow")]
	if !strings.Contains(createLayout, "restoreInitialDestinationFromHistory()") {
		t.Fatal("createLayout does not limit history restoration to initial navigation")
	}
	if strings.Contains(createLayout, "restoreDestinationFromHistory()") {
		t.Fatal("createLayout restores history directly after every response")
	}

	navigation := readUITypeScript(t, "navigation_ts.ts")
	if !strings.Contains(navigation, "initialDestinationRestored") || !strings.Contains(navigation, "currentDestination !== undefined") {
		t.Fatal("initial history restoration is not guarded after a destination has already rendered")
	}
}

func TestHistoryRestorationRequiresVisibleDestination(t *testing.T) {
	navigation := readUITypeScript(t, "navigation_ts.ts")
	if !strings.Contains(navigation, "navigationDestinationIsKnown(destination) && navigationDestinationIsVisible(destination)") {
		t.Fatal("history restoration can route a known destination that is not visible to this user")
	}
}

func readUITypeScript(t *testing.T, name string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("..", "ts", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
