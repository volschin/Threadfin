package src

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const (
	bootstrapCSSURL       = "https://cdn.jsdelivr.net/npm/bootstrap@5.3.8/dist/css/bootstrap.min.css"
	bootstrapCSSIntegrity = "sha384-sRIl4kxILFvY47J16cr9ZwB07vP4J8+LH7qKQnuqkuIAvNWLzeN8tE5YBujZqJLB"
	bootstrapBundleURL    = "https://cdn.jsdelivr.net/npm/bootstrap@5.3.8/dist/js/bootstrap.bundle.min.js"
	bootstrapJSIntegrity  = "sha384-FKyoEForCGlyvwx9Hj09JcYn3nv7wiPVlz7YYwJrWVcXK/BmnVDxM+D2scQbITxI"
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

func TestHTMLPagesLoadBootstrap538Stylesheet(t *testing.T) {
	for _, name := range []string{"index.html", "configuration.html", "login.html", "create-first-user.html"} {
		t.Run(name, func(t *testing.T) {
			doc := parseHTMLFile(t, name)
			link := findElement(t, doc, "link", "href", bootstrapCSSURL)
			if got := attribute(link, "integrity"); got != bootstrapCSSIntegrity {
				t.Fatalf("Bootstrap stylesheet integrity = %q, want %q", got, bootstrapCSSIntegrity)
			}
		})
	}
}

func TestInteractivePagesLoadBootstrap538Bundle(t *testing.T) {
	for _, name := range []string{"index.html", "configuration.html"} {
		t.Run(name, func(t *testing.T) {
			doc := parseHTMLFile(t, name)
			script := findElement(t, doc, "script", "src", bootstrapBundleURL)
			if got := attribute(script, "integrity"); got != bootstrapJSIntegrity {
				t.Fatalf("Bootstrap bundle integrity = %q, want %q", got, bootstrapJSIntegrity)
			}
		})
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

func TestNavbarsUseBootstrap53DarkTheme(t *testing.T) {
	for _, name := range []string{"index.html", "login.html"} {
		t.Run(name, func(t *testing.T) {
			doc := parseHTMLFile(t, name)
			navbar := findElementWithClass(t, doc, "nav", "navbar")
			if got := attribute(navbar, "data-bs-theme"); got != "dark" {
				t.Errorf("navbar data-bs-theme = %q, want dark", got)
			}
			if hasClass(navbar, "navbar-dark") {
				t.Error("navbar still uses deprecated navbar-dark class")
			}
		})
	}
}

type htmlElement struct {
	name       string
	attributes map[string]string
}

var (
	elementPattern   = regexp.MustCompile(`(?is)<(link|nav|script)\b([^>]*)>`)
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
