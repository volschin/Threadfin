package src

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"threadfin/src/internal/authentication"
)

func TestUIAdminBrowserUsersExcludeCredentialMaterial(t *testing.T) {
	if err := authentication.Init(t.TempDir()+"/config", 60); err != nil {
		t.Fatal(err)
	}
	userID, err := authentication.CreateNewUser("browser-admin", "plain-text-marker")
	if err != nil {
		t.Fatal(err)
	}
	permissions := map[string]interface{}{
		"username":              "browser-admin",
		"defaultUser":           true,
		"authentication.web":    true,
		"authentication.pms":    true,
		"authentication.m3u":    false,
		"authentication.xml":    true,
		"authentication.api":    false,
		"authentication.config": false,
	}
	if err := authentication.WriteUserData(userID, permissions); err != nil {
		t.Fatal(err)
	}

	raw, err := authentication.GetAllUserData()
	if err != nil {
		t.Fatal(err)
	}
	rawRecord := raw[userID].(map[string]interface{})
	passwordHash := rawRecord["_password"].(string)
	if !strings.HasPrefix(passwordHash, "$argon2id$") || strings.Contains(passwordHash, "plain-text-marker") {
		t.Fatalf("stored password = %q, want Argon2id verifier without plaintext", passwordHash)
	}

	browser, err := browserUserData()
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(browser)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"_password", "_salt", "_username", "plain-text-marker", passwordHash} {
		if strings.Contains(string(encoded), forbidden) {
			t.Errorf("browser users response contains forbidden credential material %q: %s", forbidden, encoded)
		}
	}
	wantData, ok := browser[userID].(map[string]interface{})["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("browser user %q does not preserve users[id].data: %#v", userID, browser[userID])
	}
	for key, want := range permissions {
		if got := wantData[key]; got != want {
			t.Errorf("browser users[%q].data[%q] = %#v, want %#v", userID, key, got, want)
		}
	}
}

func TestUIAdminSettingsRequestAcceptsPersistedHTTPSKeysAndTypes(t *testing.T) {
	var request RequestStruct
	if err := json.Unmarshal([]byte(`{"settings":{"excludeStreamHttps":true,"httpsPort":8443}}`), &request); err != nil {
		t.Fatal(err)
	}
	if request.Settings.ExcludeStreamsHttps == nil || !*request.Settings.ExcludeStreamsHttps {
		t.Fatal("excludeStreamHttps did not reach the boolean request field")
	}
	if request.Settings.HttpsPort == nil || *request.Settings.HttpsPort != 8443 {
		t.Fatalf("httpsPort = %#v, want integer 8443", request.Settings.HttpsPort)
	}
}

func TestUIAdminSettingsRequestNormalizesLegacyExcludeStreamsHTTPSAlias(t *testing.T) {
	for _, test := range []struct {
		name    string
		payload string
		want    bool
	}{
		{name: "canonical singular", payload: `{"settings":{"excludeStreamHttps":true}}`, want: true},
		{name: "legacy plural", payload: `{"settings":{"excludeStreamsHttps":true}}`, want: true},
		{name: "canonical wins", payload: `{"settings":{"excludeStreamHttps":false,"excludeStreamsHttps":true}}`, want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			var request RequestStruct
			if err := json.Unmarshal([]byte(test.payload), &request); err != nil {
				t.Fatal(err)
			}
			normalizeSettingsRequestAliases(&request)
			if request.Settings.ExcludeStreamsHttps == nil || *request.Settings.ExcludeStreamsHttps != test.want {
				t.Fatalf("canonical excludeStreamHttps = %#v, want %t", request.Settings.ExcludeStreamsHttps, test.want)
			}
			encoded, err := json.Marshal(request.Settings)
			if err != nil {
				t.Fatal(err)
			}
			var settings map[string]any
			if err := json.Unmarshal(encoded, &settings); err != nil {
				t.Fatal(err)
			}
			if settings["excludeStreamHttps"] != test.want {
				t.Fatalf("normalized settings = %s, canonical value was lost", encoded)
			}
			if _, exists := settings["excludeStreamsHttps"]; exists {
				t.Fatalf("normalized settings = %s, legacy spelling must not be emitted", encoded)
			}
		})
	}
}

func TestUIAdminSettingsAuthenticationEnableRequiresReload(t *testing.T) {
	for _, test := range []struct {
		name                    string
		previous, current, want bool
	}{
		{name: "enable", previous: false, current: true, want: true},
		{name: "remain enabled", previous: true, current: true, want: false},
		{name: "disable", previous: true, current: false, want: false},
		{name: "remain disabled", previous: false, current: false, want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := authenticationSettingsRequireReload(test.previous, test.current); got != test.want {
				t.Fatalf("authenticationSettingsRequireReload(%t, %t) = %t, want %t", test.previous, test.current, got, test.want)
			}
		})
	}
}

func TestUIAdminSettingsGeneratedCategoriesTypesAndPreview(t *testing.T) {
	output := runUIAdminNodeFixture(t, "settings-contracts", settingsAdminNodeScript, "settings_page_ts.js")
	var got struct {
		Sections         []map[string]any  `json:"sections"`
		Serialized       map[string]any    `json:"serialized"`
		Types            map[string]string `json:"types"`
		PresentationOnly bool              `json:"presentationOnly"`
		PMSPreview       string            `json:"pmsPreview"`
		XEPGPreview      string            `json:"xepgPreview"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode generated Settings fixture: %v\n%s", err, output)
	}
	wantSections := []map[string]any{
		{"key": "general", "label": "General", "settings": []any{"epgSource", "ThreadfinAutoUpdate", "ssdp", "tuner", "epgCategories", "epgCategoriesColors", "dummy", "dummyChannel", "ignoreFilters", "api"}},
		{"key": "files", "label": "Files", "settings": []any{"update", "files.update", "temp.path", "cache.images", "bindIpAddress", "httpThreadfinDomain", "forceHttps", "excludeStreamHttps", "httpsPort", "httpsThreadfinDomain", "xepg.replace.missing.images", "xepg.replace.channel.title", "enableNonAscii"}},
		{"key": "streaming", "label": "Streaming", "settings": []any{"udpxy", "buffer.size.kb", "buffer.timeout", "user.agent", "ffmpeg.path", "ffmpeg.options", "ffmpeg.forceHttp", "vlc.path", "vlc.options"}},
		{"key": "backup", "label": "Backup", "settings": []any{"backup.path", "backup.keep"}},
		{"key": "authentication", "label": "Authentication", "settings": []any{"authentication.web", "authentication.pms", "authentication.m3u", "authentication.xml", "authentication.api"}},
	}
	if !reflect.DeepEqual(got.Sections, wantSections) {
		t.Fatalf("Settings sections = %#v, want %#v", got.Sections, wantSections)
	}
	if _, exists := got.Serialized["unchanged.marker"]; exists {
		t.Fatal("partial save included an unchanged control")
	}
	if !reflect.DeepEqual(got.Serialized["update"], []any{"0000", "1200"}) {
		t.Errorf("update serialization = %#v, want string array", got.Serialized["update"])
	}
	for key, want := range map[string]any{
		"tuner": float64(4), "backup.keep": float64(20), "buffer.size.kb": float64(2048),
		"buffer.timeout": 1.5, "httpsPort": float64(8443), "epgSource": "XEPG",
		"authentication.web": false, "forceHttps": true, "dummyChannel": "30_Minutes",
	} {
		if got.Serialized[key] != want {
			t.Errorf("serialized %s = %#v, want %#v", key, got.Serialized[key], want)
		}
	}
	booleanKeys := map[string]bool{
		"ThreadfinAutoUpdate": true, "ssdp": true, "dummy": true, "ignoreFilters": true, "api": true,
		"files.update": true, "cache.images": true, "forceHttps": true, "excludeStreamHttps": true,
		"xepg.replace.missing.images": true, "xepg.replace.channel.title": true, "enableNonAscii": true,
		"ffmpeg.forceHttp": true, "authentication.web": true, "authentication.pms": true,
		"authentication.m3u": true, "authentication.xml": true, "authentication.api": true,
	}
	numericKeys := map[string]bool{
		"tuner": true, "backup.keep": true, "buffer.size.kb": true, "buffer.timeout": true, "httpsPort": true,
	}
	wantTypes := make(map[string]string)
	for _, section := range wantSections {
		for _, rawKey := range section["settings"].([]any) {
			key := rawKey.(string)
			wantTypes[key] = "string"
			if booleanKeys[key] {
				wantTypes[key] = "boolean"
			} else if numericKeys[key] {
				wantTypes[key] = "number"
			} else if key == "update" {
				wantTypes[key] = "array"
			}
		}
	}
	if len(got.Types) != len(wantTypes) {
		t.Fatalf("serialized key count = %d, want every legacy key (%d)", len(got.Types), len(wantTypes))
	}
	for key, wantType := range wantTypes {
		if got.Types[key] != wantType {
			t.Errorf("serialized type %s = %q, want %q", key, got.Types[key], wantType)
		}
	}
	if !got.PresentationOnly || !strings.Contains(got.PMSPreview, "PMS") || !strings.Contains(got.PMSPreview, "client") ||
		!strings.Contains(got.XEPGPreview, "XEPG") || !strings.Contains(got.XEPGPreview, "Mapping") || !strings.Contains(got.XEPGPreview, "M3U") {
		t.Fatalf("EPG preview contract = presentationOnly %t PMS %q XEPG %q", got.PresentationOnly, got.PMSPreview, got.XEPGPreview)
	}
}

func TestUIAdminSettingsGeneratedTabsSupportKeyboardActivationAndFocus(t *testing.T) {
	output := runUIAdminNodeFixture(t, "settings-tabs", settingsTabsNodeScript, "settings_page_ts.js")
	var got []struct {
		Start     int    `json:"start"`
		Key       string `json:"key"`
		Active    int    `json:"active"`
		Focused   int    `json:"focused"`
		Prevented bool   `json:"prevented"`
		Selected  int    `json:"selected"`
		Visible   int    `json:"visible"`
		TabStops  int    `json:"tabStops"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode generated Settings tab fixture: %v\n%s", err, output)
	}
	if len(got) != 14 {
		t.Fatalf("Settings keyboard cases = %d, want 14", len(got))
	}
	for _, result := range got {
		want := result.Start
		switch result.Key {
		case "ArrowRight", "ArrowDown":
			want = (result.Start + 1) % 5
		case "ArrowLeft", "ArrowUp":
			want = (result.Start + 4) % 5
		case "Home":
			want = 0
		case "End":
			want = 4
		}
		if result.Active != want || result.Focused != want || !result.Prevented ||
			result.Selected != 1 || result.Visible != 1 || result.TabStops != 1 {
			t.Errorf("Settings tab start=%d key=%s => active=%d focus=%d prevented=%t selected=%d visible=%d tabStops=%d, want index %d and one consistent active tab",
				result.Start, result.Key, result.Active, result.Focused, result.Prevented, result.Selected, result.Visible, result.TabStops, want)
		}
	}
}

func TestUIAdminSettingsGeneratedControlsHaveVisibleProgrammaticLabels(t *testing.T) {
	output := runUIAdminNodeFixture(t, "settings-labels", settingsLabelsNodeScript, "settings_page_ts.js")
	var got struct {
		Controls []map[string]any `json:"controls"`
		Hidden   map[string]any   `json:"hidden"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode generated Settings labels: %v\n%s", err, output)
	}
	if len(got.Controls) != 39 {
		t.Fatalf("labelled Settings controls = %d, want all 39 legacy settings", len(got.Controls))
	}
	ids := make(map[string]bool)
	for _, control := range got.Controls {
		name, _ := control["name"].(string)
		id, _ := control["id"].(string)
		labelledBy, _ := control["labelledBy"].(string)
		labelText, _ := control["labelText"].(string)
		if id == "" || ids[id] {
			t.Errorf("Settings control %q does not have a unique id: %q", name, id)
		}
		ids[id] = true
		if labelledBy == "" || strings.TrimSpace(labelText) == "" {
			t.Errorf("Settings control %q lacks a visible programmatic label: labelledBy=%q text=%q", name, labelledBy, labelText)
		}
	}
	if got.Hidden["id"] != "" || got.Hidden["labelledBy"] != "" {
		t.Fatalf("hidden Settings field must be excluded from visual labelling: %#v", got.Hidden)
	}
}

func TestUIAdminBackupAndRestoreGeneratedCommandsAreSeparatedAndSecretSafe(t *testing.T) {
	output := runUIAdminNodeFixture(t, "backup-restore", backupRestoreNodeScript, "base_ts.js")
	var got struct {
		Sent          []map[string]any `json:"sent"`
		CancelCount   int              `json:"cancelCount"`
		ConfirmPrompt string           `json:"confirmPrompt"`
		LogsSafe      bool             `json:"logsSafe"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode backup/restore fixture: %v\n%s", err, output)
	}
	want := []map[string]any{
		{"cmd": "ThreadfinBackup", "data": map[string]any{}},
		{"cmd": "ThreadfinRestore", "data": map[string]any{"base64": "data:application/zip;base64,c2VjcmV0LWJhY2t1cA=="}},
	}
	if !reflect.DeepEqual(got.Sent, want) {
		t.Fatalf("backup/restore envelopes = %#v, want %#v", got.Sent, want)
	}
	if got.CancelCount != 1 {
		t.Fatalf("cancelled restore request count = %d, want only backup", got.CancelCount)
	}
	lowerPrompt := strings.ToLower(got.ConfirmPrompt)
	if !strings.Contains(got.ConfirmPrompt, "threadfin-backup.zip") ||
		(!strings.Contains(lowerPrompt, "replace") && !strings.Contains(lowerPrompt, "overwrite")) {
		t.Fatalf("restore confirmation is not specific and destructive: %q", got.ConfirmPrompt)
	}
	if !got.LogsSafe {
		t.Fatal("restore base64 was written to the browser console")
	}
}

func TestUIAdminUsersGeneratedEnvelopesAndSafeRows(t *testing.T) {
	output := runUIAdminNodeFixture(t, "users-contracts", usersAdminNodeScript, "users_page_ts.js")
	var got struct {
		Permissions      []map[string]any `json:"permissions"`
		Create           map[string]any   `json:"create"`
		Edit             map[string]any   `json:"edit"`
		Delete           map[string]any   `json:"delete"`
		Row              map[string]any   `json:"row"`
		DefaultCanDelete bool             `json:"defaultCanDelete"`
		DefaultWEBLocked bool             `json:"defaultWEBLocked"`
		RegularCanDelete bool             `json:"regularCanDelete"`
		PasswordInRow    bool             `json:"passwordInRow"`
		VerifierInRow    bool             `json:"verifierInRow"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode generated Users fixture: %v\n%s", err, output)
	}
	if len(got.Permissions) != 6 {
		t.Fatalf("Users permissions = %#v, want WEB/PMS/M3U/XML/API/CONFIG", got.Permissions)
	}
	for _, permission := range got.Permissions {
		if permission["key"] == "" || permission["label"] == "" || strings.TrimSpace(permission["description"].(string)) == "" {
			t.Errorf("permission lacks protocol label or explanation: %#v", permission)
		}
	}
	newData := map[string]any{
		"username": "new-user", "password": "new-secret", "confirm": "new-secret", "defaultUser": false,
		"authentication.web": true, "authentication.pms": true, "authentication.m3u": false,
		"authentication.xml": true, "authentication.api": false, "authentication.config": false,
	}
	if got.Create["cmd"] != "saveNewUser" || !reflect.DeepEqual(got.Create["data"], map[string]any{"userData": newData}) {
		t.Fatalf("new user envelope = %#v", got.Create)
	}
	editData := map[string]any{
		"username": "renamed", "password": "", "confirm": "", "defaultUser": false,
		"authentication.web": true, "authentication.pms": false, "authentication.m3u": true,
		"authentication.xml": false, "authentication.api": true, "authentication.config": true,
	}
	if got.Edit["cmd"] != "saveUserData" || !reflect.DeepEqual(got.Edit["data"], map[string]any{"userData": map[string]any{"id-regular": editData}}) {
		t.Fatalf("edit user envelope = %#v", got.Edit)
	}
	deleteData := make(map[string]any, len(editData)+1)
	for key, value := range editData {
		deleteData[key] = value
	}
	deleteData["delete"] = true
	if got.Delete["cmd"] != "saveUserData" || !reflect.DeepEqual(got.Delete["data"], map[string]any{"userData": map[string]any{"id-regular": deleteData}}) {
		t.Fatalf("delete user envelope = %#v", got.Delete)
	}
	if got.Row["username"] != "visible-user" || got.PasswordInRow || got.VerifierInRow {
		t.Fatalf("safe Users row = %#v password=%t verifier=%t", got.Row, got.PasswordInRow, got.VerifierInRow)
	}
	if got.DefaultCanDelete || !got.DefaultWEBLocked || !got.RegularCanDelete {
		t.Fatalf("default-user restrictions = default delete %t web locked %t regular delete %t", got.DefaultCanDelete, got.DefaultWEBLocked, got.RegularCanDelete)
	}
}

func TestUIAdminGeneratedUsersPopupControlsHaveVisibleProgrammaticLabels(t *testing.T) {
	output := runUIAdminNodeFixture(t, "users-popup-labels", usersPopupLabelsNodeScript, "menu_ts.js", "users_page_ts.js")
	var got struct {
		Create []map[string]any `json:"create"`
		Edit   []map[string]any `json:"edit"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode generated Users popup labels: %v\n%s", err, output)
	}
	wantNames := []string{"username", "password", "confirm", "authentication.web", "authentication.pms", "authentication.m3u", "authentication.xml", "authentication.api", "authentication.config"}
	for mode, controls := range map[string][]map[string]any{"create": got.Create, "edit": got.Edit} {
		if len(controls) != len(wantNames) {
			t.Fatalf("%s Users controls = %#v, want %d labelled controls", mode, controls, len(wantNames))
		}
		ids := make(map[string]bool)
		for index, control := range controls {
			if control["name"] != wantNames[index] {
				t.Errorf("%s Users control %d name = %#v, want %q", mode, index, control["name"], wantNames[index])
			}
			id, _ := control["id"].(string)
			labelledBy, _ := control["labelledBy"].(string)
			labelText, _ := control["labelText"].(string)
			if id == "" || ids[id] {
				t.Errorf("%s Users control %q does not have a unique non-empty id: %q", mode, control["name"], id)
			}
			ids[id] = true
			if labelledBy == "" || strings.TrimSpace(strings.TrimSuffix(labelText, ":")) == "" {
				t.Errorf("%s Users control %q lacks a visible programmatic label: labelledBy=%q text=%q", mode, control["name"], labelledBy, labelText)
			}
			if control["hidden"] == true {
				t.Errorf("%s Users visible field %q was treated as hidden", mode, control["name"])
			}
		}
	}
}

func TestUIAdminUsersGeneratedResponsiveIdentityDetails(t *testing.T) {
	output := runUIAdminNodeFixture(t, "users-responsive", usersResponsiveNodeScript, "users_page_ts.js")
	var got struct {
		IdentityClass          string   `json:"identityClass"`
		ActionClass            string   `json:"actionClass"`
		DesktopPermissionCells int      `json:"desktopPermissionCells"`
		DetailsTag             string   `json:"detailsTag"`
		Summary                string   `json:"summary"`
		Terms                  []string `json:"terms"`
		Descriptions           []string `json:"descriptions"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode generated responsive Users row: %v\n%s", err, output)
	}
	if got.IdentityClass != "tf-user-identity" || got.ActionClass != "tf-user-actions" || got.DesktopPermissionCells != 6 {
		t.Fatalf("Users identity/table vocabulary = identity %q action %q permission cells %d", got.IdentityClass, got.ActionClass, got.DesktopPermissionCells)
	}
	if got.DetailsTag != "DETAILS" || got.Summary != "Permissions" || len(got.Terms) != 6 || len(got.Descriptions) != 6 {
		t.Fatalf("narrow Users permission disclosure = tag %q summary %q terms %#v descriptions %#v", got.DetailsTag, got.Summary, got.Terms, got.Descriptions)
	}
	for index, permission := range []string{"WEB", "PMS", "M3U", "XML", "API", "CONFIG"} {
		if got.Terms[index] != permission ||
			(!strings.Contains(got.Descriptions[index], "Allowed") && !strings.Contains(got.Descriptions[index], "Denied")) ||
			!strings.Contains(got.Descriptions[index], userPermissionDescriptionForTest(permission)) {
			t.Errorf("responsive permission %s = term %q description %q", permission, got.Terms[index], got.Descriptions[index])
		}
	}

	css, err := os.ReadFile(filepath.Join("..", "html", "css", "app-shell.css"))
	if err != nil {
		t.Fatal(err)
	}
	stylesheet := string(css)
	for _, contract := range []string{
		".tf-app .tf-user-permission-details {\n  display: none;",
		".tf-app .tf-users-table {\n    min-width: 0;",
		".tf-app .tf-users-table thead,\n  .tf-app .tf-user-permission-cell {\n    display: none;",
		".tf-app .tf-user-permission-details {\n    display: block;",
	} {
		if !strings.Contains(stylesheet, contract) {
			t.Errorf("responsive Users stylesheet missing contract %q", contract)
		}
	}
}

func userPermissionDescriptionForTest(label string) string {
	switch label {
	case "WEB":
		return "web interface"
	case "PMS":
		return "DVR discovery"
	case "M3U":
		return "generated M3U"
	case "XML":
		return "generated XMLTV"
	case "API":
		return "API commands"
	case "CONFIG":
		return "private-LAN"
	default:
		return ""
	}
}

func TestUIAdminLogGeneratedRenderingFilteringRefreshAndReset(t *testing.T) {
	output := runUIAdminNodeFixture(t, "log-contracts", logAdminNodeScript, "log_page_ts.js", "logs_ts.js")
	var got struct {
		LiteralText     string            `json:"literalText"`
		InnerHTMLWrites int               `json:"innerHTMLWrites"`
		Severities      map[string]string `json:"severities"`
		KnownLayer      string            `json:"knownLayer"`
		UnknownLayer    string            `json:"unknownLayer"`
		FirstVisible    []string          `json:"firstVisible"`
		RefreshVisible  []string          `json:"refreshVisible"`
		FocusPreserved  bool              `json:"focusPreserved"`
		ResetSent       []map[string]any  `json:"resetSent"`
		ResetPrompt     string            `json:"resetPrompt"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode generated Log fixture: %v\n%s", err, output)
	}
	if got.LiteralText != `<img src=x onerror="credential-marker">` || got.InnerHTMLWrites != 0 {
		t.Fatalf("malicious-looking log was not rendered literally: text=%q innerHTML writes=%d", got.LiteralText, got.InnerHTMLWrites)
	}
	wantSeverities := map[string]string{"debug": "DEBUG", "warning": "WARNING", "error": "ERROR", "plain": ""}
	if !reflect.DeepEqual(got.Severities, wantSeverities) {
		t.Fatalf("response-backed log severities = %#v, want %#v", got.Severities, wantSeverities)
	}
	if got.KnownLayer != "XEPG" || got.UnknownLayer != "" {
		t.Fatalf("log layer derivation = known %q unknown %q", got.KnownLayer, got.UnknownLayer)
	}
	if !reflect.DeepEqual(got.FirstVisible, []string{"[WARNING] first needle"}) ||
		!reflect.DeepEqual(got.RefreshVisible, []string{"[WARNING] first needle", "[WARNING] second needle"}) || !got.FocusPreserved {
		t.Fatalf("stable log filter/refresh = first %v refresh %v focus %t", got.FirstVisible, got.RefreshVisible, got.FocusPreserved)
	}
	if !reflect.DeepEqual(got.ResetSent, []map[string]any{{"cmd": "resetLogs", "data": map[string]any{}}}) {
		t.Fatalf("reset Logs envelope = %#v", got.ResetSent)
	}
	lowerPrompt := strings.ToLower(got.ResetPrompt)
	if !strings.Contains(lowerPrompt, "reset") || (!strings.Contains(lowerPrompt, "cannot") && !strings.Contains(lowerPrompt, "permanent")) {
		t.Fatalf("reset Logs confirmation is not destructive: %q", got.ResetPrompt)
	}
}

func TestUIAdminAuthenticationGeneratedValidationUsesNativePOSTWithoutLoggingCredentials(t *testing.T) {
	output := runUIAdminNodeFixture(t, "authentication-contracts", authenticationAdminNodeScript, "authentication_ts.js")
	var got struct {
		ValidResult       bool   `json:"validResult"`
		ValidSubmitCalls  int    `json:"validSubmitCalls"`
		EmptyResult       bool   `json:"emptyResult"`
		EmptyFocus        string `json:"emptyFocus"`
		MismatchResult    bool   `json:"mismatchResult"`
		MismatchFocus     string `json:"mismatchFocus"`
		MismatchMessage   string `json:"mismatchMessage"`
		CredentialsLogged bool   `json:"credentialsLogged"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode generated authentication fixture: %v\n%s", err, output)
	}
	if !got.ValidResult || got.ValidSubmitCalls != 0 {
		t.Fatalf("valid authentication = result %t imperative submits %d, want native form POST", got.ValidResult, got.ValidSubmitCalls)
	}
	if got.EmptyResult || got.EmptyFocus != "username" {
		t.Fatalf("empty authentication = result %t focus %q", got.EmptyResult, got.EmptyFocus)
	}
	if got.MismatchResult || got.MismatchFocus != "password" || strings.TrimSpace(got.MismatchMessage) == "" {
		t.Fatalf("password mismatch = result %t focus %q message %q", got.MismatchResult, got.MismatchFocus, got.MismatchMessage)
	}
	if got.CredentialsLogged {
		t.Fatal("authentication JavaScript logged raw credentials")
	}
}

func TestUIAdminAuthenticationPagesExposeLabelsAutocompleteAndNativeSubmit(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		wantExtra []string
	}{
		{name: "login.html", password: "current-password"},
		{name: "create-first-user.html", password: "new-password", wantExtra: []string{
			`<label for="confirm">`, `id="confirm" type="password" name="confirm" autocomplete="new-password"`,
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			content, err := os.ReadFile(filepath.Join("..", "html", test.name))
			if err != nil {
				t.Fatal(err)
			}
			html := string(content)
			for _, contract := range append([]string{
				`class="tf-auth-page"`, `id="authentication" action="" method="post" onsubmit="return login(event);"`,
				`<label for="username">`, `id="username" type="text" name="username" autocomplete="username"`,
				`autofocus required`, `<label for="password">`, `id="password" type="password" name="password" autocomplete="` + test.password + `"`,
				`id="submit" type="submit"`, `id="err"`, `aria-live="polite"`, `href="css/app-shell.css"`,
			}, test.wantExtra...) {
				if !strings.Contains(html, contract) {
					t.Errorf("authentication page is missing behavior contract %q", contract)
				}
			}
		})
	}
}

func TestUIAdminAuthenticationPOSTPreservesTokenRedirectContracts(t *testing.T) {
	restorePersistentState(t)
	System.ConfigurationWizard = false
	System.ScanInProgress = 0
	Settings = SettingsStruct{AuthenticationWEB: true, Language: "en", Port: "34400"}

	tests := []struct {
		name      string
		values    url.Values
		seedUser  bool
		wantUsers int
	}{
		{name: "login", values: url.Values{"username": {"existing-user"}, "password": {"existing-password"}}, seedUser: true, wantUsers: 1},
		{name: "first user", values: url.Values{"username": {"first-user"}, "password": {"first-password"}, "confirm": {"first-password"}}, wantUsers: 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := authentication.Init(filepath.Join(t.TempDir(), "config"), 60); err != nil {
				t.Fatal(err)
			}
			if test.seedUser {
				userID, err := authentication.CreateNewUser("existing-user", "existing-password")
				if err != nil {
					t.Fatal(err)
				}
				if err := authentication.WriteUserData(userID, map[string]interface{}{"username": "existing-user", "authentication.web": true}); err != nil {
					t.Fatal(err)
				}
			}
			request := httptest.NewRequest(http.MethodPost, "http://threadfin.example/web/", strings.NewReader(test.values.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			response := httptest.NewRecorder()
			Web(response, request)
			if response.Code != http.StatusMovedPermanently || response.Header().Get("Location") != "/web" {
				t.Fatalf("authentication POST = status %d location %q, want 301 /web", response.Code, response.Header().Get("Location"))
			}
			cookie := response.Header().Get("Set-Cookie")
			if !strings.Contains(cookie, "Token=") || strings.Contains(cookie, test.values.Get("password")) || strings.Contains(response.Body.String(), test.values.Get("password")) {
				t.Fatalf("authentication POST token/secret contract = cookie %q body %q", cookie, response.Body.String())
			}
			users, err := authentication.GetAllUserData()
			if err != nil {
				t.Fatal(err)
			}
			if len(users) != test.wantUsers {
				t.Fatalf("authentication users = %d, want %d", len(users), test.wantUsers)
			}
		})
	}
}

func TestUIAdminScopedAssetsLanguageAndEmbeddingParity(t *testing.T) {
	styles, err := os.ReadFile(filepath.Join("..", "html", "css", "app-shell.css"))
	if err != nil {
		t.Fatal(err)
	}
	for _, selector := range []string{
		".tf-app .tf-settings", ".tf-app .tf-users", ".tf-app .tf-log", ".tf-auth-page .tf-auth-card",
		".tf-app .tf-admin-table-scroll", ".tf-app .tf-destructive-action", ".tf-auth-page :focus-visible",
	} {
		if !strings.Contains(string(styles), selector) {
			t.Errorf("admin/auth stylesheet is missing scoped selector %q", selector)
		}
	}

	languageBytes, err := os.ReadFile(filepath.Join("..", "html", "lang", "en.json"))
	if err != nil {
		t.Fatal(err)
	}
	var language map[string]any
	if err := json.Unmarshal(languageBytes, &language); err != nil {
		t.Fatal(err)
	}
	admin, ok := language["admin"].(map[string]any)
	if !ok {
		t.Fatal("English language catalog has no admin copy")
	}
	for _, key := range []string{"settings", "users", "log", "authentication"} {
		if _, exists := admin[key]; !exists {
			t.Errorf("English admin language catalog is missing %q", key)
		}
	}
	usersLanguage, ok := language["users"].(map[string]any)
	if !ok {
		t.Fatal("English language catalog has no Users copy")
	}
	configLanguage, ok := usersLanguage["config"].(map[string]any)
	if !ok || configLanguage["title"] != "CONFIG Access" || !strings.Contains(configLanguage["description"].(string), "private-LAN") {
		t.Fatalf("English CONFIG permission copy = %#v, want explicit private-LAN source-fetch authority", configLanguage)
	}

	indexBytes, err := os.ReadFile(filepath.Join("..", "html", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	for _, script := range []string{"settings_page_ts.js", "users_page_ts.js", "log_page_ts.js"} {
		if !strings.Contains(string(indexBytes), `src="js/`+script+`"`) {
			t.Errorf("index does not load %s", script)
		}
	}
	for _, generated := range []string{"authentication_ts.js", "base_ts.js"} {
		content, err := os.ReadFile(filepath.Join("..", "html", "js", generated))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(content), "console.log(reader.result)") || (generated == "authentication_ts.js" && strings.Contains(string(content), "console.log")) {
			t.Errorf("generated %s retains raw credential/base64 console logging", generated)
		}
	}

	webUI = make(map[string]interface{})
	loadHTMLMap()
	for _, name := range []string{
		"html/index.html", "html/login.html", "html/create-first-user.html", "html/css/app-shell.css", "html/lang/en.json",
		"html/js/authentication_ts.js", "html/js/base_ts.js", "html/js/logs_ts.js", "html/js/log_page_ts.js",
		"html/js/menu_ts.js", "html/js/settings_ts.js", "html/js/settings_page_ts.js", "html/js/users_page_ts.js",
	} {
		t.Run(filepath.Base(name), func(t *testing.T) {
			encoded, ok := webUI[name].(string)
			if !ok {
				t.Fatalf("embedded asset %q is missing", name)
			}
			got, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(filepath.Join("..", name))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("embedded asset %q does not match source", name)
			}
		})
	}
}

func runUIAdminNodeFixture(t *testing.T, name, source string, assets ...string) []byte {
	t.Helper()
	script := filepath.Join(t.TempDir(), name+".js")
	if err := os.WriteFile(script, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	args := []string{script}
	for _, asset := range assets {
		args = append(args, filepath.Join("..", "html", "js", asset))
	}
	output, err := exec.Command("node", args...).CombinedOutput()
	if err != nil {
		t.Fatalf("execute generated admin fixture %s: %v\n%s", name, err, output)
	}
	return output
}

const settingsAdminNodeScript = `
const fs = require("fs");
const vm = require("vm");
const context = {console: {log() {}, warn() {}}};
vm.createContext(context);
vm.runInContext(fs.readFileSync(process.argv[2], "utf8"), context);
const booleanKeys = new Set([
  "ThreadfinAutoUpdate", "ssdp", "dummy", "ignoreFilters", "api", "files.update", "cache.images",
  "forceHttps", "excludeStreamHttps", "xepg.replace.missing.images", "xepg.replace.channel.title",
  "enableNonAscii", "ffmpeg.forceHttp", "authentication.web", "authentication.pms", "authentication.m3u",
  "authentication.xml", "authentication.api"
]);
const numericSelectKeys = new Set(["tuner", "backup.keep", "buffer.size.kb"]);
const values = {
  update: "0000,,1200", tuner: "4", "backup.keep": "20", "buffer.size.kb": "2048",
  "buffer.timeout": "1.5", httpsPort: "8443", epgSource: "XEPG", "authentication.web": false,
  forceHttps: true, dummyChannel: "30_Minutes", "backup.path": "/var/lib/threadfin/backup"
};
const controls = [];
for (const section of context.SETTINGS_SECTIONS) {
  for (const name of section.settings) {
    if (booleanKeys.has(name)) {
      controls.push({tagName: "INPUT", type: "checkbox", name, checked: values[name] === undefined ? true : values[name], className: "changed"});
    } else if (numericSelectKeys.has(name) || name === "epgSource" || name === "dummyChannel") {
      controls.push({tagName: "SELECT", name, value: values[name] === undefined ? "value" : values[name], className: "changed"});
    } else {
      controls.push({tagName: "INPUT", type: "text", name, value: values[name] === undefined ? "value" : values[name], className: "changed"});
    }
  }
}
controls.push({tagName: "INPUT", type: "text", name: "unchanged.marker", value: "do-not-send", className: ""});
const serialized = context.serializeSettingsChanges(controls);
const types = {};
for (const [key, value] of Object.entries(serialized)) types[key] = Array.isArray(value) ? "array" : typeof value;
const server = {settings: {epgSource: "PMS"}};
const pmsTarget = {textContent: ""};
const xepgTarget = {textContent: ""};
context.renderEPGSourcePreview("PMS", pmsTarget);
context.renderEPGSourcePreview("XEPG", xepgTarget);
process.stdout.write(JSON.stringify({
  sections: context.SETTINGS_SECTIONS,
  serialized,
  types,
  presentationOnly: server.settings.epgSource === "PMS",
  pmsPreview: pmsTarget.textContent,
  xepgPreview: xepgTarget.textContent
}));
`

const settingsTabsNodeScript = `
const fs = require("fs");
const vm = require("vm");
let focused = -1;
class Element {
  constructor(index, attribute, value) { this.index = index; this.attributes = {[attribute]: value}; this.hidden = false; }
  getAttribute(name) { return this.attributes[name]; }
  setAttribute(name, value) { this.attributes[name] = String(value); }
  focus() { focused = this.index; }
}
const tabs = ["general", "files", "streaming", "backup", "authentication"].map((key, index) => new Element(index, "data-settings-tab", key));
const panels = ["general", "files", "streaming", "backup", "authentication"].map((key, index) => new Element(index, "data-settings-panel", key));
const root = {querySelectorAll(selector) { return selector === "[data-settings-panel]" ? panels : tabs; }};
const context = {console: {log() {}, warn() {}}, document: {querySelector(selector) { return selector === ".tf-settings" ? root : null; }}};
vm.createContext(context);
vm.runInContext(fs.readFileSync(process.argv[2], "utf8"), context);
function run(start, key) {
  context.activateSettingsSection(context.SETTINGS_SECTIONS[start].key);
  focused = -1;
  let prevented = false;
  context.handleSettingsTabKeydown({key, preventDefault() { prevented = true; }}, start);
  return {
    start, key, prevented, focused,
    active: tabs.findIndex(tab => tab.getAttribute("aria-selected") === "true"),
    selected: tabs.filter(tab => tab.getAttribute("aria-selected") === "true").length,
    visible: panels.filter(panel => !panel.hidden).length,
    tabStops: tabs.filter(tab => tab.getAttribute("tabindex") === "0").length,
  };
}
const results = [];
for (let start = 0; start < 5; start++) results.push(run(start, "ArrowRight"));
for (let start = 0; start < 5; start++) results.push(run(start, "ArrowLeft"));
results.push(run(2, "ArrowDown"), run(2, "ArrowUp"), run(3, "Home"), run(1, "End"));
process.stdout.write(JSON.stringify(results));
`

const settingsLabelsNodeScript = `
const fs = require("fs");
const vm = require("vm");
class Label { constructor(text) { this.id = ""; this.textContent = text; } }
class Row { constructor(label) { this.label = label; } querySelector(selector) { return selector === "td:first-child" ? this.label : null; } }
class Control {
  constructor(name, type, label) { this.name = name; this.type = type; this.id = ""; this.attributes = {}; this.row = new Row(label); }
  closest(selector) { return selector === "tr" ? this.row : null; }
  setAttribute(name, value) { this.attributes[name] = String(value); }
  getAttribute(name) { return this.attributes[name] || ""; }
}
const context = {console: {log() {}, warn() {}}, document: {querySelector() { return null; }}};
vm.createContext(context);
vm.runInContext(fs.readFileSync(process.argv[2], "utf8"), context);
const controls = [];
for (const section of context.SETTINGS_SECTIONS) {
  for (const name of section.settings) controls.push(new Control(name, "text", new Label("Label for " + name)));
}
const hidden = new Control("internal", "hidden", new Label("Internal"));
const root = {querySelectorAll(selector) { return selector === "input, select, textarea" ? controls.concat(hidden) : []; }};
context.labelSettingsControls(root);
function result(control) {
  const labelledBy = control.getAttribute("aria-labelledby");
  return {name: control.name, id: control.id, labelledBy, labelText: control.row.label.id === labelledBy ? control.row.label.textContent : ""};
}
process.stdout.write(JSON.stringify({controls: controls.map(result), hidden: result(hidden)}));
`

const backupRestoreNodeScript = `
const fs = require("fs");
const vm = require("vm");
const marker = "data:application/zip;base64,c2VjcmV0LWJhY2t1cA==";
const sent = [];
let leaked = false;
let confirmResult = false;
let confirmPrompt = "";
let fileInput = null;
let lastReader = null;
class FakeElement {
  constructor(tag) { this.tagName = tag.toUpperCase(); this.attributes = {}; this.style = {}; this.files = []; this.parentNode = null; }
  setAttribute(key, value) { this.attributes[key] = String(value); if (key === "id") this.id = String(value); }
  remove() { if (fileInput === this) fileInput = null; }
  click() {}
}
class FakeFileReader {
  constructor() { this.result = null; lastReader = this; }
  readAsDataURL() { this.result = marker; }
}
class CaptureServer {
  constructor(cmd) { this.cmd = cmd; }
  request(data) { sent.push({cmd: this.cmd, data: JSON.parse(JSON.stringify(data))}); }
}
function MainMenuItem() {}
function SettingsCategoryItem() {}
function ClipboardJS() { this.on = function() {}; }
const document = {
  body: {appendChild(node) { fileInput = node; node.parentNode = this; }},
  querySelectorAll() { return []; },
  querySelector(selector) { return selector === "input[type=file]" ? fileInput : null; },
  createElement(tag) { return new FakeElement(tag); },
  getElementById(id) {
    if (id === "popup" || id === "loading") return new FakeElement("div");
    if (id === "upload") return fileInput;
    return null;
  }
};
const context = {
  document,
  window: {},
  console: {
    log(...values) { if (values.some(value => String(value).includes(marker))) leaked = true; },
    warn(...values) { if (values.some(value => String(value).includes(marker))) leaked = true; }
  },
  bootstrap: {Modal: function() { this.show = function() {}; this.hide = function() {}; }, Tooltip: function() {}},
  ClipboardJS,
  MainMenuItem,
  SettingsCategoryItem,
  Server: CaptureServer,
  FileReader: FakeFileReader,
  confirm(message) { confirmPrompt = message; return confirmResult; },
  alert() {},
  setTimeout,
  clearTimeout
};
context.bootstrap.Tooltip.getInstance = function() { return {setContent() {}}; };
vm.createContext(context);
vm.runInContext(fs.readFileSync(process.argv[2], "utf8"), context);
context.backup();
context.restore();
fileInput.files = [{name: "threadfin-backup.zip"}];
fileInput.onchange();
const cancelCount = sent.length;
confirmResult = true;
context.restore();
fileInput.files = [{name: "threadfin-backup.zip"}];
fileInput.onchange();
lastReader.onload();
process.stdout.write(JSON.stringify({sent, cancelCount, confirmPrompt, logsSafe: !leaked}));
`

const usersAdminNodeScript = `
const fs = require("fs");
const vm = require("vm");
const context = {console: {log() {}, warn() {}}};
vm.createContext(context);
vm.runInContext(fs.readFileSync(process.argv[2], "utf8"), context);
const createInput = {
  username: "new-user", password: "new-secret", confirm: "new-secret", defaultUser: false,
  "authentication.web": true, "authentication.pms": true, "authentication.m3u": false,
  "authentication.xml": true, "authentication.api": false, "authentication.config": false
};
const editInput = {
  username: "renamed", password: "", confirm: "", defaultUser: false,
  "authentication.web": true, "authentication.pms": false, "authentication.m3u": true,
  "authentication.xml": false, "authentication.api": true, "authentication.config": true
};
const record = {data: {
  username: "visible-user", defaultUser: true,
  "authentication.web": true, "authentication.pms": true, "authentication.m3u": false,
  "authentication.xml": true, "authentication.api": false, "authentication.config": false,
  password: "must-not-render", _password: "$argon2id$must-not-render", _salt: "salt", _username: "digest"
}};
const row = context.userRowValues(record);
const passwordInRow = JSON.stringify(row).includes("must-not-render") || Object.prototype.hasOwnProperty.call(row, "password");
const verifierInRow = Object.keys(row).some(key => key.startsWith("_"));
const defaultActions = context.userActions(record.data);
const regularActions = context.userActions({...record.data, defaultUser: false});
process.stdout.write(JSON.stringify({
  permissions: context.USER_PERMISSION_DEFINITIONS,
  create: context.buildUserRequest("-", createInput, false),
  edit: context.buildUserRequest("id-regular", editInput, false),
  delete: context.buildUserRequest("id-regular", editInput, true),
  row,
  defaultCanDelete: defaultActions.canDelete,
  defaultWEBLocked: defaultActions.webLocked,
  regularCanDelete: regularActions.canDelete,
  passwordInRow,
  verifierInRow
}));
`

const usersPopupLabelsNodeScript = `
const fs = require("fs");
const vm = require("vm");
class Element {
  constructor(tag = "div") {
    this.tagName = tag.toUpperCase(); this.attributes = {}; this.children = []; this.parentElement = null;
    this._innerHTML = ""; this._textContent = ""; this.checked = false; this.disabled = false; this.listeners = {};
    this.classList = {add() {}, remove() {}, contains() { return false; }};
  }
  set id(value) { this.attributes.id = String(value); }
  get id() { return this.attributes.id || ""; }
  set className(value) { this.attributes.class = String(value); }
  get className() { return this.attributes.class || ""; }
  set type(value) { this.attributes.type = String(value); }
  get type() { return this.attributes.type || ""; }
  set name(value) { this.attributes.name = String(value); }
  get name() { return this.attributes.name || ""; }
  set value(value) { this.attributes.value = String(value === undefined ? "" : value); }
  get value() { return this.attributes.value || ""; }
  set innerHTML(value) { this._innerHTML = String(value); this.children = []; }
  get innerHTML() { return this._innerHTML; }
  set textContent(value) { this._textContent = String(value); this._innerHTML = ""; this.children = []; }
  get textContent() { return this._textContent || this._innerHTML.replace(/<[^>]+>/g, "") || this.children.map(child => child.textContent).join(""); }
  setAttribute(name, value) { this.attributes[name] = String(value); }
  getAttribute(name) { return this.attributes[name]; }
  appendChild(child) { child.parentElement = this; this.children.push(child); return child; }
  addEventListener(name, listener) { this.listeners[name] = listener; }
  getElementsByTagName(tag) { const wanted = tag.toUpperCase(); return descendants(this).filter(item => item.tagName === wanted); }
  querySelectorAll(selector) {
    const items = descendants(this);
    if (selector === "input, select" || selector === "input, select, textarea") return items.filter(item => item.tagName === "INPUT" || item.tagName === "SELECT" || item.tagName === "TEXTAREA");
    const named = selector.match(/^\[name="([^"]+)"\]$/); if (named) return items.filter(item => item.name === named[1]);
    return [];
  }
  querySelector(selector) { return this.querySelectorAll(selector)[0] || null; }
  closest(selector) { let item = this.parentElement; while (item) { if (selector === "tr" && item.tagName === "TR") return item; item = item.parentElement; } return null; }
  focus() {}
}
function descendants(node) { return node.children.reduce((all, child) => all.concat(child, descendants(child)), []); }
const popup = new Element("div"); popup.id = "popup-custom";
const roots = [popup];
const document = {
  createElement(tag) { return new Element(tag); },
  getElementById(id) { return roots.concat(roots.flatMap(descendants)).find(item => item.id === id) || null; },
  addEventListener() {}, querySelectorAll() { return []; }, contains() { return true; }
};
const user = {username: "editor", defaultUser: false, "authentication.web": true, "authentication.pms": false, "authentication.m3u": true, "authentication.xml": false, "authentication.api": true, "authentication.config": true};
const context = {
  document, window: {}, console: {log() {}, warn() {}}, SERVER: {}, UNDO: {}, BULK_EDIT: false,
  getLocalData(type, id) { return type === "users" && id === "regular" ? {...user} : {}; },
  enhanceSourcePopup() {}, enhanceFilterPopup() {}, showPopUpElement() {}, showElement() {},
  bootstrap: {Modal: function() {}, Tooltip: function() {}}, ClipboardJS: function() {}, confirm() { return true; }
};
vm.createContext(context);
vm.runInContext(fs.readFileSync(process.argv[2], "utf8"), context);
vm.runInContext(fs.readFileSync(process.argv[3], "utf8"), context);
function snapshot() {
  const wanted = ["username", "password", "confirm", "authentication.web", "authentication.pms", "authentication.m3u", "authentication.xml", "authentication.api", "authentication.config"];
  return wanted.map(name => {
    const control = popup.querySelector('[name="' + name + '"]');
    const labelledBy = control ? control.getAttribute("aria-labelledby") || "" : "";
    const label = labelledBy ? document.getElementById(labelledBy) : null;
    return {name, id: control ? control.id : "", labelledBy, labelText: label ? label.textContent : "", hidden: control ? control.type === "hidden" : false};
  });
}
context.openPopUp("users", undefined); const create = snapshot();
context.openPopUp("users", {id: "regular"}); const edit = snapshot();
process.stdout.write(JSON.stringify({create, edit}));
`

const usersResponsiveNodeScript = `
const fs = require("fs");
const vm = require("vm");
class Element {
  constructor(tag) { this.tagName = tag.toUpperCase(); this.children = []; this.attributes = {}; this.parentElement = null; this._text = ""; this.listeners = {}; }
  set id(value) { this.attributes.id = String(value); } get id() { return this.attributes.id || ""; }
  set className(value) { this.attributes.class = String(value); } get className() { return this.attributes.class || ""; }
  set textContent(value) { this._text = String(value); this.children = []; } get textContent() { return this._text + this.children.map(child => child.textContent).join(""); }
  set innerHTML(_value) { this.children = []; this._text = ""; }
  appendChild(child) { child.parentElement = this; this.children.push(child); return child; }
  setAttribute(name, value) { this.attributes[name] = String(value); }
  getAttribute(name) { return this.attributes[name] || ""; }
  addEventListener(name, listener) { this.listeners[name] = listener; }
}
function descendants(node) { return node.children.reduce((all, child) => all.concat(child, descendants(child)), []); }
const document = {createElement(tag) { return new Element(tag); }, createTextNode(text) { const node = new Element("span"); node.textContent = text; return node; }};
const context = {document, console: {log() {}, warn() {}}, openPopUp() {}, SERVER: {users: {user1: {data: {
  username: "alice", defaultUser: false, "authentication.web": true, "authentication.pms": false,
  "authentication.m3u": true, "authentication.xml": false, "authentication.api": true, "authentication.config": true
}}}}};
vm.createContext(context);
vm.runInContext(fs.readFileSync(process.argv[2], "utf8"), context);
const host = new Element("div");
context.renderUsersPage(host);
const items = descendants(host);
const identity = items.find(item => item.className === "tf-user-identity");
const action = items.find(item => item.className === "tf-user-actions");
const desktopPermissionCells = items.filter(item => item.className === "tf-user-permission-cell").length;
const details = items.find(item => item.className === "tf-user-permission-details");
const summary = details ? descendants(details).find(item => item.tagName === "SUMMARY") : null;
const terms = details ? descendants(details).filter(item => item.tagName === "DT").map(item => item.textContent) : [];
const descriptions = details ? descendants(details).filter(item => item.tagName === "DD").map(item => item.textContent) : [];
process.stdout.write(JSON.stringify({identityClass: identity ? identity.className : "", actionClass: action ? action.className : "", desktopPermissionCells, detailsTag: details ? details.tagName : "", summary: summary ? summary.textContent : "", terms, descriptions}));
`

const logAdminNodeScript = `
const fs = require("fs");
const vm = require("vm");
let innerHTMLWrites = 0;
class Element {
  constructor(tag) { this.tagName = tag.toUpperCase(); this.children = []; this.attributes = {}; this.dataset = {}; this.hidden = false; this._text = ""; }
  set textContent(value) { this._text = String(value); this.children = []; }
  get textContent() { return this._text + this.children.map(child => child.textContent).join(""); }
  set innerHTML(_value) { innerHTMLWrites++; }
  appendChild(child) { this.children.push(child); return child; }
  append(child) { return this.appendChild(child); }
  replaceChildren(...children) { this.children = children; }
  setAttribute(key, value) { this.attributes[key] = String(value); }
}
const search = new Element("input");
const list = new Element("div");
const wrapper = {scrollTop: 0, scrollHeight: 100};
const document = {
  activeElement: search,
  createElement(tag) { return new Element(tag); },
  getElementById(id) { if (id === "content_log_entries") return list; if (id === "box-wrapper") return wrapper; return null; }
};
const sent = [];
let confirmResult = false;
let resetPrompt = "";
class CaptureServer { constructor(cmd) { this.cmd = cmd; } request(data) { sent.push({cmd: this.cmd, data: JSON.parse(JSON.stringify(data))}); } }
const context = {
  console: {log() {}, warn() {}}, document, Server: CaptureServer,
  SERVER: {log: {log: []}},
  confirm(message) { resetPrompt = message; return confirmResult; },
  setTimeout(listener) { listener(); }
};
vm.createContext(context);
vm.runInContext(fs.readFileSync(process.argv[2], "utf8"), context);
vm.runInContext(fs.readFileSync(process.argv[3], "utf8"), context);
const malicious = '<img src=x onerror="credential-marker">';
const maliciousElement = context.createLogEntryElement(malicious);
const severities = {
  debug: context.logEntrySeverity("2026 [DEBUG] detail"),
  warning: context.logEntrySeverity("2026 [WARNING] detail"),
  error: context.logEntrySeverity("2026 [ERROR] detail"),
  plain: context.logEntrySeverity("2026 plain detail")
};
const knownLayer = context.logEntryLayer("2026-08-26 12:00:00 [Threadfin] XEPG:&nbsp;Ready");
const unknownLayer = context.logEntryLayer("2026-08-26 12:00:00 [Threadfin] message without a layer");
context.LOG_VIEW_STATE.search = "needle";
context.LOG_VIEW_STATE.severities.WARNING = true;
context.renderLogEntryList(list, ["[WARNING] first needle", "[ERROR] needle", "[WARNING] unrelated"]);
const firstVisible = list.children.map(child => child.queryText);
context.renderLogEntryList(list, ["[WARNING] first needle", "[ERROR] needle", "[WARNING] unrelated", "[WARNING] second needle"]);
const refreshVisible = list.children.map(child => child.queryText);
const focusPreserved = document.activeElement === search;
context.resetLogs();
confirmResult = true;
context.resetLogs();
process.stdout.write(JSON.stringify({
  literalText: maliciousElement.queryText,
  innerHTMLWrites,
  severities,
  knownLayer,
  unknownLayer,
  firstVisible,
  refreshVisible,
  focusPreserved,
  resetSent: sent,
  resetPrompt
}));
`

const authenticationAdminNodeScript = `
const fs = require("fs");
const vm = require("vm");
const usernameMarker = "username-secret-marker";
const passwordMarker = "password-secret-marker";
let credentialsLogged = false;
let focused = "";
let submitCalls = 0;
const fields = {};
function field(id, name, value) {
  return fields[id] = {
    id, name, value, style: {},
    setAttribute() {},
    focus() { focused = id; }
  };
}
const form = {submit() { submitCalls++; }};
const content = {getElementsByTagName() { return Object.values(fields); }};
const err = {textContent: "", innerHTML: ""};
const document = {getElementById(id) { if (id === "content") return content; if (id === "authentication") return form; if (id === "err") return err; return fields[id] || null; }};
const context = {
  document,
  console: {
    log(...values) { const text = values.map(value => JSON.stringify(value)).join(" "); if (text.includes(usernameMarker) || text.includes(passwordMarker)) credentialsLogged = true; },
    warn(...values) { const text = values.map(value => JSON.stringify(value)).join(" "); if (text.includes(usernameMarker) || text.includes(passwordMarker)) credentialsLogged = true; }
  },
  alert() {}
};
vm.createContext(context);
vm.runInContext(fs.readFileSync(process.argv[2], "utf8"), context);
function setFields(username, password, confirm) {
  for (const key of Object.keys(fields)) delete fields[key];
  field("username", "username", username);
  field("password", "password", password);
  if (confirm !== undefined) field("confirm", "confirm", confirm);
  focused = "";
  err.textContent = "";
  err.innerHTML = "";
}
setFields(usernameMarker, passwordMarker);
const validResult = context.login();
const validSubmitCalls = submitCalls;
setFields("", passwordMarker);
const emptyResult = context.login();
const emptyFocus = focused;
setFields(usernameMarker, passwordMarker, "different-password");
const mismatchResult = context.login();
const mismatchFocus = focused;
const mismatchMessage = err.textContent || err.innerHTML;
process.stdout.write(JSON.stringify({validResult, validSubmitCalls, emptyResult, emptyFocus, mismatchResult, mismatchFocus, mismatchMessage, credentialsLogged}));
`
