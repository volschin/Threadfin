package src

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

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
	if !strings.Contains(navigation, "enableGroupSelection(\".bulk\")") {
		t.Fatal("openLegacyMenu does not preserve legacy Mapping bulk selection setup")
	}

	network := readUITypeScript(t, "network_ts.ts")
	if !strings.Contains(network, `openLegacyMenu(response["openMenu"])`) {
		t.Fatal("numeric openMenu response does not use openLegacyMenu")
	}
	if strings.Contains(network, `document.getElementById(response["openMenu"])`) {
		t.Fatal("numeric openMenu response still selects a navigation element directly")
	}
}

func TestGroupedNavigationInitializesLegacyMenuMetadata(t *testing.T) {
	navigation := readUITypeScript(t, "navigation_ts.ts")
	if !strings.Contains(navigation, "initializeLegacyMenuItems()") {
		t.Fatal("grouped navigation does not initialize legacy menu metadata before ShowContent can use it")
	}

	menu := readUITypeScript(t, "menu_ts.ts")
	if !strings.Contains(menu, "initializeTableHeader(): void") {
		t.Fatal("legacy table metadata is still coupled to legacy menu DOM rendering")
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

func readUITypeScript(t *testing.T, name string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("..", "ts", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
