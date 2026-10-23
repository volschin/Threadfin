package src

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestFileToBase64ReturnsOpenError(t *testing.T) {
	_, err := fileToBase64(filepath.Join(t.TempDir(), "missing.html"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("fileToBase64() error = %v, want os.ErrNotExist", err)
	}
}

func TestCreateMapFromFilesOrdersAssetsDeterministically(t *testing.T) {
	oldBlankMap, oldMapName := blankMap, mapName
	defer func() {
		blankMap = oldBlankMap
		mapName = oldMapName
	}()

	dir := t.TempDir()
	names := []string{"alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel"}
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	blankMap = make(map[string]interface{})
	mapName = "assets"
	got, err := createMapFromFiles(dir)
	if err != nil {
		t.Fatal(err)
	}

	var want strings.Builder
	for _, name := range names {
		path := filepath.Join(dir, name)
		fmt.Fprintf(&want, "\tassets[%q] = %q\n", path, "eA==")
	}
	if got != want.String() {
		t.Fatalf("createMapFromFiles() emitted unstable ordering:\n%s", got)
	}
}

func TestHTMLPagesLoadLocalBootstrap6Stylesheet(t *testing.T) {
	for _, name := range []string{"index.html", "configuration.html", "login.html", "create-first-user.html"} {
		t.Run(name, func(t *testing.T) {
			findElement(t, parseHTMLFile(t, name), "link", "href", "css/vendor/bootstrap.min.css")
			findElement(t, parseHTMLFile(t, name), "link", "href", "css/app.css")
		})
	}
	css, err := os.ReadFile(filepath.Join("..", "html", "css", "vendor", "bootstrap.min.css"))
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := os.ReadFile(filepath.Join("..", "package.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(manifest), `"bootstrap": "6.0.0-alpha.1"`) {
		t.Fatal("Bootstrap dependency is not pinned to the qualified alpha")
	}
	for _, feature := range []string{".combobox-toggle", ".chip", ".drawer", "@layer"} {
		if !strings.Contains(string(css), feature) {
			t.Errorf("local Bootstrap stylesheet lacks native v6 feature %s", feature)
		}
	}
}

func TestPagesLoadNativeESMBundles(t *testing.T) {
	for page, bundle := range map[string]string{"index.html": "app", "configuration.html": "setup", "login.html": "auth", "create-first-user.html": "auth"} {
		script := findElement(t, parseHTMLFile(t, page), "script", "src", "js/"+bundle+".js")
		if attribute(script, "type") != "module" {
			t.Errorf("%s must load native ESM", page)
		}
	}
}

func TestAuthenticationPagesDoNotLoadUnusedThirdPartyScripts(t *testing.T) {
	for _, name := range []string{"login.html", "create-first-user.html"} {
		t.Run(name, func(t *testing.T) {
			doc := parseHTMLFile(t, name)
			for _, node := range doc {
				if node.name != "script" {
					continue
				}
				src := attribute(node, "src")
				if strings.Contains(src, "bootstrap") || strings.Contains(src, "clipboard.js") {
					t.Errorf("authentication page loads unused third-party script %q", src)
				}
			}
		})
	}
}

func TestNavbarsInheritApplicationColorTheme(t *testing.T) {
	for _, name := range []string{"index.html", "login.html"} {
		navbar := findElementWithClass(t, parseHTMLFile(t, name), "nav", "navbar")
		if attribute(navbar, "data-bs-theme") != "" || hasClass(navbar, "navbar-dark") {
			t.Errorf("%s navbar overrides the user's application theme", name)
		}
	}
	theme, err := os.ReadFile(filepath.Join("..", "ts", "theme.js"))
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{"prefers-color-scheme: dark", "document.documentElement.dataset.bsTheme", "threadfin-theme"} {
		if !strings.Contains(string(theme), fragment) {
			t.Errorf("theme preference contract missing %s", fragment)
		}
	}
}

func TestIndexHTMLProvidesSemanticApplicationShell(t *testing.T) {
	doc := parseHTMLFile(t, "index.html")

	viewport := findElement(t, doc, "meta", "name", "viewport")
	if got := attribute(viewport, "content"); got != "width=device-width, initial-scale=1.0" {
		t.Errorf("viewport content = %q, want responsive device viewport", got)
	}

	findElement(t, doc, "link", "href", "css/app.css")
	findElementWithClass(t, doc, "div", "tf-app")
	findElementWithClass(t, doc, "aside", "tf-sidebar")

	navigation := findElement(t, doc, "nav", "id", "main-menu")
	if got := attribute(navigation, "aria-label"); got != "Primary" {
		t.Errorf("primary navigation aria-label = %q, want Primary", got)
	}

	main := findElement(t, doc, "main", "id", "main-content")
	if got := attribute(main, "tabindex"); got != "-1" {
		t.Errorf("main content tabindex = %q, want -1", got)
	}

	for _, id := range []string{"content", "overview-content", "connections-content", "activity-content"} {
		findElement(t, doc, "section", "id", id)
	}
}

func TestIndexHTMLEmbeddedAssetsMatchSource(t *testing.T) {
	webUI = make(map[string]interface{})
	loadHTMLMap()

	for _, name := range []string{"html/index.html", "html/configuration.html", "html/login.html", "html/create-first-user.html", "html/css/app-shell.css", "html/css/app.css", "html/css/vendor/bootstrap.min.css", "html/css/vendor/fontawesome.min.css", "html/js/app.js", "html/js/setup.js", "html/js/auth.js"} {
		t.Run(name, func(t *testing.T) {
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
				t.Errorf("embedded asset %q does not match its source file", name)
			}
		})
	}
}

func TestIndexHTMLMobileTargetsHaveMinimumInlineSize(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "html", "css", "app-shell.css"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "min-inline-size: 44px;") {
		t.Error("mobile interactive targets lack a 44px minimum inline size")
	}
}

type htmlElement struct {
	name       string
	attributes map[string]string
}

var (
	elementPattern   = regexp.MustCompile(`(?is)<(aside|dialog|div|link|main|meta|nav|script|section)\b([^>]*)>`)
	attributePattern = regexp.MustCompile(`([[:alnum:]_-]+)\s*=\s*"([^"]*)"`)
)

func parseHTMLFile(t *testing.T, name string) []htmlElement {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("..", "html", name))
	if err != nil {
		t.Fatal(err)
	}

	var elements []htmlElement
	for _, match := range elementPattern.FindAllSubmatch(content, -1) {
		element := htmlElement{name: strings.ToLower(string(match[1])), attributes: make(map[string]string)}
		for _, attr := range attributePattern.FindAllSubmatch(match[2], -1) {
			element.attributes[strings.ToLower(string(attr[1]))] = string(attr[2])
		}
		elements = append(elements, element)
	}
	return elements
}

func findElement(t *testing.T, elements []htmlElement, name, attributeName, value string) htmlElement {
	t.Helper()
	for _, element := range elements {
		if element.name == name && attribute(element, attributeName) == value {
			return element
		}
	}
	t.Fatalf("missing <%s> with %s=%q", name, attributeName, value)
	return htmlElement{}
}

func findElementWithClass(t *testing.T, elements []htmlElement, name, class string) htmlElement {
	t.Helper()
	for _, element := range elements {
		if element.name == name && hasClass(element, class) {
			return element
		}
	}
	t.Fatalf("missing <%s> with class %q", name, class)
	return htmlElement{}
}

func attribute(element htmlElement, name string) string {
	return element.attributes[name]
}

func hasClass(element htmlElement, class string) bool {
	for _, candidate := range strings.Fields(attribute(element, "class")) {
		if candidate == class {
			return true
		}
	}
	return false
}

// Controllers share one lexical ESM scope. Verify the build input order and a
// declared controller symbol in the actual shipped bundle rather than expecting
// classic script tags that would initialize the application twice.
func assertUIBundleControllers(t *testing.T, page, bundle string, controllers ...string) {
	t.Helper()
	script := findElement(t, parseHTMLFile(t, page), "script", "src", "js/"+bundle+".js")
	if attribute(script, "type") != "module" {
		t.Fatalf("%s does not load its ESM bundle", page)
	}
	build, err := os.ReadFile(filepath.Join("..", "scripts", "build-ui.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	list := regexp.MustCompile(`(?m)const ` + bundle + ` = \[(.*?)\]`).FindStringSubmatch(string(build))
	if len(list) != 2 {
		t.Fatalf("bundle controller list %s missing", bundle)
	}
	shipped, err := os.ReadFile(filepath.Join("..", "html", "js", bundle+".js"))
	if err != nil {
		t.Fatal(err)
	}
	previous := -1
	for _, controller := range controllers {
		at := strings.Index(list[1], "'"+controller+"'")
		if at <= previous {
			t.Errorf("%s controller absent or out of order in %s", controller, bundle)
		}
		previous = at
		generated, err := os.ReadFile(filepath.Join("..", "html", "js", controller+"_ts.js"))
		if err != nil {
			t.Fatal(err)
		}
		symbol := regexp.MustCompile(`(?:function|class) ([A-Za-z_][A-Za-z_0-9]*)`).FindStringSubmatch(string(generated))
		if len(symbol) != 2 || !strings.Contains(string(shipped), symbol[1]) {
			t.Errorf("%s controller is absent from shipped %s bundle", controller, bundle)
		}
	}
}
