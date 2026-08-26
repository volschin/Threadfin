package src

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
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

	var result struct {
		Sent                      []map[string]any `json:"sent"`
		CancelledNoRequest        bool             `json:"cancelledNoRequest"`
		LogsSafe                  bool             `json:"logsSafe"`
		SensitivePayloadPreserved bool             `json:"sensitivePayloadPreserved"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
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
				"name": "Local channels", "description": "Primary", "file.source": "/srv/threadfin/channels.m3u",
				"buffer": "ffmpeg", "tuner": float64(3), "http_proxy.ip": "", "http_proxy.port": "",
				"http_headers.origin": "https://origin.example", "http_headers.referer": "https://referer.example",
			}}}},
		},
		{
			"cmd": "saveFilesM3U",
			"data": map[string]any{"files": map[string]any{"m3u": map[string]any{"M1": map[string]any{
				"name": "Local channels", "description": "Primary", "file.source": "/srv/threadfin/channels.m3u",
				"buffer": "ffmpeg", "tuner": float64(3), "http_proxy.ip": "", "http_proxy.port": "",
				"http_headers.origin": "https://origin.example", "http_headers.referer": "https://referer.example", "delete": true,
			}}}},
		},
		{
			"cmd": "saveFilesHDHR",
			"data": map[string]any{"files": map[string]any{"hdhr": map[string]any{"-": map[string]any{
				"name": "Living room tuner", "description": "HDHR", "file.source": "192.0.2.20:5004",
				"buffer": "vlc", "tuner": float64(4), "http_proxy.ip": "", "http_proxy.port": "",
			}}}},
		},
		{
			"cmd": "updateFileHDHR",
			"data": map[string]any{"files": map[string]any{"hdhr": map[string]any{"H1": map[string]any{
				"name": "Living room tuner", "description": "HDHR", "file.source": "192.0.2.20:5004",
				"buffer": "vlc", "tuner": float64(4), "http_proxy.ip": "", "http_proxy.port": "",
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
			"data": map[string]any{"files": map[string]any{"xmltv": map[string]any{"-": map[string]any{
				"name": "Guide", "description": "Local guide", "file.source": "C:\\guides\\guide.xml",
				"http_proxy.ip": "", "http_proxy.port": "",
			}}}},
		},
		{
			"cmd": "updateFileXMLTV",
			"data": map[string]any{"files": map[string]any{"xmltv": map[string]any{"X1": map[string]any{
				"name": "Guide", "description": "Local guide", "file.source": "C:\\guides\\guide.xml",
				"http_proxy.ip": "", "http_proxy.port": "",
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
	if !reflect.DeepEqual(result.Sent, want) {
		t.Fatal("generated source create/update/delete command and payload matrix changed")
	}
	if !result.CancelledNoRequest {
		t.Fatal("cancelled source deletion sent a request")
	}
	if !result.LogsSafe {
		t.Fatal("source save diagnostics included form values")
	}
	if !result.SensitivePayloadPreserved {
		t.Fatal("display-only sanitization changed the editable/sent source value")
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

func TestConfigurationWizardRequiredPropagatesUnexpectedStatError(t *testing.T) {
	required, err := configurationWizardRequired("invalid\x00settings-path")
	if err == nil {
		t.Fatal("invalid settings path did not propagate its stat error")
	}
	if errors.Is(err, os.ErrNotExist) {
		t.Fatal("test fixture unexpectedly produced an os.ErrNotExist error")
	}
	if required {
		t.Fatal("unexpected stat error incorrectly enabled the wizard")
	}
}

func TestProviderSourceLogSanitizerPreservesOnlySafeContext(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{name: "URL userinfo query and fragment", source: "https://reader:password@provider.example/path/list.m3u?token=value&other=value#fragment", want: "https://provider.example/path/list.m3u"},
		{name: "URL query", source: "http://provider.example/guide.xml?token=value", want: "http://provider.example/guide.xml"},
		{name: "Unix local path", source: "/srv/threadfin/local.m3u?literal-filename", want: "/srv/threadfin/local.m3u?literal-filename"},
		{name: "Windows local path", source: `C:\guides\local.xml`, want: `C:\guides\local.xml`},
		{name: "malformed remote URL", source: "https://%", want: "[redacted source]"},
		{name: "single slash remote URL", source: "HtTpS:/reader:password@provider.example/list.m3u?token=value", want: "[redacted source]"},
		{name: "opaque remote URL", source: "HTTP:reader:password@provider.example?token=value", want: "[redacted source]"},
		{name: "invalid escape remote URL", source: "https://provider.example/%zz?token=value", want: "[redacted source]"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := sanitizeProviderSourceForLog(test.source); got != test.want {
				t.Fatal("provider source log sanitization did not produce the expected safe representation")
			}
		})
	}
}

func TestProviderClassificationKeepsMalformedHTTPSourceOutOfLocalFileErrors(t *testing.T) {
	restorePersistentState(t)
	previousWebScreenLog := WebScreenLog
	previousLogOutput := log.Writer()
	log.SetOutput(io.Discard)
	t.Cleanup(func() {
		WebScreenLog = previousWebScreenLog
		log.SetOutput(previousLogOutput)
	})

	tests := []struct {
		name   string
		source string
	}{
		{name: "mixed-case URL", source: "HtTpS://reader:credential-marker@provider.example/%zz?token=query-marker"},
		{name: "single-slash URL", source: "https:/reader:credential-marker@provider.example/%zz?token=query-marker"},
		{name: "opaque URL", source: "HTTP:reader:credential-marker@provider.example/%zz?token=query-marker"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			System = SystemStruct{}
			System.Name = "Threadfin"
			Settings = SettingsStruct{}
			Settings.LogEntriesRAM = 100
			Settings.Files.M3U = map[string]interface{}{
				"M1": map[string]interface{}{
					"file.source": test.source,
					"new":         true,
				},
			}
			WebScreenLog = WebScreenLogStruct{}

			err := getProviderData("m3u", "M1")
			if err == nil {
				t.Fatal("malformed remote provider unexpectedly succeeded")
			}
			var pathError *os.PathError
			if errors.As(err, &pathError) || !strings.Contains(err.Error(), "[redacted source]") {
				t.Fatal("URL-like provider did not return the normalized remote-source error")
			}
			logs := strings.Join(WebScreenLog.Log, "\n")
			for _, marker := range []string{"credential-marker", "query-marker"} {
				if strings.Contains(err.Error(), marker) || strings.Contains(logs, marker) {
					t.Fatal("remote provider failure exposed source credentials or query values")
				}
			}
			if !strings.Contains(logs, "Download:") || !strings.Contains(logs, "[redacted source]") || strings.Contains(logs, "Open:") {
				t.Fatal("URL-like provider did not use the remote download branch")
			}
		})
	}
}

func TestProviderClassificationPreservesSupportedLocalPathBranches(t *testing.T) {
	restorePersistentState(t)
	previousWebScreenLog := WebScreenLog
	previousLogOutput := log.Writer()
	log.SetOutput(io.Discard)
	t.Cleanup(func() {
		WebScreenLog = previousWebScreenLog
		log.SetOutput(previousLogOutput)
	})

	tests := []struct {
		name   string
		source string
	}{
		{name: "Unix absolute", source: "/task5-provider-fixture/missing.m3u"},
		{name: "relative", source: "task5-provider-fixture/missing.m3u"},
		{name: "UNC", source: `\\server\share\missing.m3u`},
		{name: "Windows drive", source: `C:\guides\missing.m3u`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			System = SystemStruct{}
			System.Name = "Threadfin"
			Settings = SettingsStruct{}
			Settings.LogEntriesRAM = 100
			Settings.Files.M3U = map[string]interface{}{
				"M1": map[string]interface{}{
					"file.source": test.source,
					"new":         true,
				},
			}
			WebScreenLog = WebScreenLogStruct{}

			err := getProviderData("m3u", "M1")
			if err == nil {
				t.Fatal("missing local provider fixture unexpectedly succeeded")
			}
			var pathError *os.PathError
			if !errors.As(err, &pathError) {
				t.Fatal("supported local provider did not use the file branch")
			}
			logs := strings.Join(WebScreenLog.Log, "\n")
			if !strings.Contains(logs, "Open:") || strings.Contains(logs, "Download:") {
				t.Fatal("supported local provider did not retain the local file branch")
			}
		})
	}
}

func TestProviderParseErrorDoesNotExposeRawURLLikeSource(t *testing.T) {
	marker := "credential-marker"
	_, _, err := downloadFileFromServer("https://provider.example/%zz?token="+marker, "")
	if err == nil {
		t.Fatal("invalid provider URL fixture did not return its expected parse error")
	}
	if strings.Contains(err.Error(), marker) {
		t.Fatal("provider URL parse error exposed a query value")
	}
}

func TestProviderHTTPStatusErrorDoesNotExposeQueryValues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	marker := "credential-marker"
	_, _, err := downloadFileFromServer(server.URL+"/guide.xml?token="+marker, "")
	if err == nil {
		t.Fatal("provider fixture did not return its expected HTTP status error")
	}
	if strings.Contains(err.Error(), marker) {
		t.Fatal("provider HTTP status error exposed a query value")
	}
}

func TestProviderTransportErrorDoesNotExposeQueryValues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	providerURL := server.URL
	server.Close()

	marker := "credential-marker"
	_, _, err := downloadFileFromServer(providerURL+"/guide.xml?token="+marker, "")
	if err == nil {
		t.Fatal("closed provider fixture did not return its expected transport error")
	}
	if strings.Contains(err.Error(), marker) {
		t.Fatal("provider transport error exposed a query value")
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
			Source       string `json:"source"`
			Counts       []struct {
				Value string `json:"value"`
			} `json:"counts"`
		} `json:"xmltv"`
		Locations map[string]bool   `json:"locations"`
		Displays  map[string]string `json:"displays"`
		Progress  map[string]any    `json:"progress"`
		Error     map[string]any    `json:"error"`
		Success   map[string]any    `json:"success"`
		Next      map[string]any    `json:"next"`
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
	if len(got.XMLTV) != 1 || got.XMLTV[0].ID != "X1" || got.XMLTV[0].ProviderType != "xmltv" || got.XMLTV[0].Source != "https://provider.example/guide.xml" || len(got.XMLTV[0].Counts) != 2 || got.XMLTV[0].Counts[0].Value != "5" || got.XMLTV[0].Counts[1].Value != "20" {
		t.Fatal("derived XMLTV row did not retain safe source context and existing counts")
	}
	for name, want := range map[string]string{
		"url":       "https://provider.example/path/list.m3u",
		"local":     "/srv/threadfin/local.m3u?literal-filename",
		"windows":   `C:\guides\local.xml`,
		"malformed": "{{.sources.locationInvalidDisplay}}",
	} {
		if got.Displays[name] != want {
			t.Errorf("display sanitization case %s did not produce the expected safe representation", name)
		}
	}
	for _, name := range []string{"singleSlash", "opaque", "invalidEscape"} {
		if got.Displays[name] == "" || strings.Contains(got.Displays[name], "credential-marker") || strings.Contains(got.Displays[name], "query-marker") {
			t.Errorf("display sanitization case %s exposed URL userinfo or query content", name)
		}
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

func TestGeneratedTask5RequestsSettleBusyAndTransportFailures(t *testing.T) {
	temp := t.TempDir()
	scriptPath := filepath.Join(temp, "source-request-failures.js")
	if err := os.WriteFile(scriptPath, []byte(sourceRequestFailureNodeScript), 0o600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(
		"node", scriptPath,
		filepath.Join("..", "html", "js", "sources_ts.js"),
		filepath.Join("..", "html", "js", "configuration_ts.js"),
		filepath.Join("..", "html", "js", "network_ts.js"),
	)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("execute generated Task 5 failure lifecycle: %v\n%s", err, output)
	}
	var got struct {
		BusySource               map[string]any `json:"busySource"`
		BusyWizard               map[string]any `json:"busyWizard"`
		TransportSource          map[string]any `json:"transportSource"`
		TransportWizard          map[string]any `json:"transportWizard"`
		LogsSafe                 bool           `json:"logsSafe"`
		MalformedSourcesRejected bool           `json:"malformedSourcesRejected"`
		MalformedWizardsRejected bool           `json:"malformedWizardsRejected"`
		TransportOneShot         bool           `json:"transportOneShot"`
		StatusFalsePreserved     bool           `json:"statusFalsePreserved"`
		NormalWizardRouted       bool           `json:"normalWizardRouted"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode generated Task 5 failure lifecycle: %v\n%s", err, output)
	}
	for name, result := range map[string]map[string]any{
		"busy source":      got.BusySource,
		"transport source": got.TransportSource,
	} {
		if result["state"] != "error" || result["message"] == "" {
			t.Errorf("%s request did not replace progress with an error", name)
		}
	}
	for name, result := range map[string]map[string]any{
		"busy wizard":      got.BusyWizard,
		"transport wizard": got.TransportWizard,
	} {
		if result["disabled"] != false || result["status"] == "" || result["routed"] != false {
			t.Errorf("%s request did not re-enable Next with an error while suppressing completion", name)
		}
	}
	if !got.LogsSafe {
		t.Fatal("Task 5 WebSocket diagnostics included request or response content")
	}
	if !got.MalformedSourcesRejected || !got.MalformedWizardsRejected {
		t.Fatal("parse-valid malformed WebSocket response was accepted as a source or wizard success")
	}
	if !got.TransportOneShot {
		t.Fatal("WebSocket onerror followed by onclose settled the request more than once")
	}
	if !got.StatusFalsePreserved {
		t.Fatal("explicit status:false response no longer preserves its server error")
	}
	if !got.NormalWizardRouted {
		t.Fatal("explicit status:true wizard completion no longer follows the established route")
	}
}

func TestGeneratedSourcePopupValidationLabelsAndFocusReturn(t *testing.T) {
	temp := t.TempDir()
	scriptPath := filepath.Join(temp, "source-popup-accessibility.js")
	if err := os.WriteFile(scriptPath, []byte(sourcePopupAccessibilityNodeScript), 0o600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("node", scriptPath, filepath.Join("..", "html", "js", "sources_ts.js"))
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("execute generated source popup accessibility: %v\n%s", err, output)
	}
	var got struct {
		Valid                 bool              `json:"valid"`
		Values                map[string]string `json:"values"`
		Labels                map[string]string `json:"labels"`
		SourceInvalid         bool              `json:"sourceInvalid"`
		SourceFocusCount      int               `json:"sourceFocusCount"`
		ErrorVisible          bool              `json:"errorVisible"`
		PopupKind             string            `json:"popupKind"`
		InvokerFocusCount     int               `json:"invokerFocusCount"`
		ReplacementFocusCount int               `json:"replacementFocusCount"`
		SourceClassReset      bool              `json:"sourceClassReset"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode generated source popup accessibility: %v\n%s", err, output)
	}
	if got.Valid {
		t.Fatal("unsupported source scheme passed popup validation")
	}
	if got.Values["name"] != "Preserved name" || got.Values["source"] != "ftp://provider.example/list.m3u" {
		t.Fatal("source validation changed entered values")
	}
	if !got.SourceInvalid || got.SourceFocusCount != 1 || !got.ErrorVisible {
		t.Fatal("source validation did not annotate, focus, and explain the invalid field")
	}
	for field, want := range map[string]string{"name": "Source name", "source": "Source location", "buffer": "Buffer"} {
		if got.Labels[field] != want {
			t.Errorf("source popup field %s does not have its visible programmatic label", field)
		}
	}
	if got.PopupKind != "m3u" || got.InvokerFocusCount != 1 || got.ReplacementFocusCount != 1 {
		t.Fatal("source popup did not return focus to its invoking control after modal close")
	}
	if !got.SourceClassReset {
		t.Fatal("source popup styling leaked into a subsequently opened legacy popup")
	}
}

func TestGeneratedWizardLabelsAndModeAwareProgress(t *testing.T) {
	temp := t.TempDir()
	scriptPath := filepath.Join(temp, "wizard-mode-accessibility.js")
	if err := os.WriteFile(scriptPath, []byte(wizardModeAccessibilityNodeScript), 0o600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("node", scriptPath, filepath.Join("..", "html", "js", "configuration_ts.js"))
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("execute generated wizard mode/accessibility state: %v\n%s", err, output)
	}
	var got struct {
		PMS struct {
			Step        string `json:"step"`
			Action      string `json:"action"`
			XMLTVHidden bool   `json:"xmltvHidden"`
			LabelledBy  string `json:"labelledBy"`
		} `json:"pms"`
		XEPG struct {
			Step        string `json:"step"`
			Action      string `json:"action"`
			XMLTVHidden bool   `json:"xmltvHidden"`
			LabelledBy  string `json:"labelledBy"`
		} `json:"xepg"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode generated wizard mode/accessibility state: %v\n%s", err, output)
	}
	if got.PMS.Step != "Step 3 of 3" || got.PMS.Action != "Finish setup" || !got.PMS.XMLTVHidden || got.PMS.LabelledBy != "wizard-field-label" {
		t.Fatal("PMS M3U step is not a labelled final step with truthful 3-of-3 progress")
	}
	if got.XEPG.Step != "Step 4 of 4" || got.XEPG.Action != "Finish setup" || got.XEPG.XMLTVHidden || got.XEPG.LabelledBy != "wizard-field-label" {
		t.Fatal("XEPG XMLTV step is not a labelled final step with truthful 4-of-4 progress")
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
	for _, contract := range []string{
		".tf-source-popup :focus-visible,\n.tf-setup :focus-visible",
		".tf-source-popup :is(button, input, select, [role=\"button\"]),\n  .tf-setup :is(button, input, select, [role=\"button\"])",
		"min-height: 44px",
		"min-inline-size: 44px",
	} {
		if !strings.Contains(string(styles), contract) {
			t.Errorf("source/wizard accessibility stylesheet is missing %q", contract)
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
  URL,
  document: {getElementById(id) { return id === "source-form-status" ? status : null; }},
  SERVER: {},
};
vm.createContext(context);
vm.runInContext(source, context);
const fixture = {
  settings: {files: {
    m3u: {M1: {name: "Zulu", "file.source": "/srv/channels.m3u", "last.update": "2026-08-26 10:00:00", "provider.availability": 75, tuner: 2, compatibility: {streams: 12, "group.title": 3, "tvg.id": 80, "stream.id": 90}}},
    hdhr: {H1: {name: "Alpha", "file.source": "192.0.2.20:5004", compatibility: {streams: 4, "group.title": 1}}},
    xmltv: {X1: {name: "Guide", "file.source": "https://reader:password@provider.example/guide.xml?token=value#fragment", "provider.availability": 100, compatibility: {"xmltv.channels": 5, "xmltv.programs": 20}}},
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
const displays = {
  url: context.sourceLocationForDisplay("https://reader:password@provider.example/path/list.m3u?token=value&other=value#fragment"),
  local: context.sourceLocationForDisplay("/srv/threadfin/local.m3u?literal-filename"),
  windows: context.sourceLocationForDisplay("C:\\guides\\local.xml"),
  malformed: context.sourceLocationForDisplay("https://%"),
  singleSlash: context.sourceLocationForDisplay("HtTpS:/reader:credential-marker@provider.example/list.m3u?token=query-marker"),
  opaque: context.sourceLocationForDisplay("HTTP:reader:credential-marker@provider.example?token=query-marker"),
  invalidEscape: context.sourceLocationForDisplay("https://provider.example/%zz?token=query-marker"),
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
process.stdout.write(JSON.stringify({playlist, xmltv, locations, displays, progress, error, success, next}));
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

const sourceRequestFailureNodeScript = `
const fs = require("fs");
const vm = require("vm");
const marker = "credential-marker";
let logsSafe = true;
let routed = false;
const sockets = [];
const sourceStatus = {hidden: true, textContent: "", setAttribute() {}};
const wizardStatus = {textContent: ""};
const next = {disabled: false};
class FakeWebSocket {
  constructor() { sockets.push(this); }
  send() {}
}
const context = {
  URL,
  console: {
    log(...values) { if (values.some(value => JSON.stringify(value).includes(marker))) logsSafe = false; },
    warn(...values) { if (values.some(value => JSON.stringify(value).includes(marker))) logsSafe = false; },
  },
  document: {
    cookie: "Token=test-token",
    getElementById(id) {
      if (id === "source-form-status") return sourceStatus;
      if (id === "wizard-request-status") return wizardStatus;
      if (id === "next") return next;
      return null;
    },
    querySelectorAll() { return []; },
  },
  window: {location: {protocol: "http:", hostname: "127.0.0.1", port: "34400", assign() { routed = true; }}},
  location: {reload() { routed = true; }},
  WebSocket: FakeWebSocket,
  PopupContent: class {},
  showElement() {},
  createLayout() {},
  alert() {},
  SERVER: {}, UNDO: {}, SERVER_CONNECTION: false, WS_AVAILABLE: false,
};
vm.createContext(context);
vm.runInContext(fs.readFileSync(process.argv[2], "utf8").replaceAll("{{.sources.responseInvalid}}", "Invalid response"), context);
vm.runInContext(fs.readFileSync(process.argv[3], "utf8").replaceAll("{{.sources.responseInvalid}}", "Invalid response"), context);
let sourceCompletionCount = 0;
const completeSourceRequest = context.completeSourceRequest;
context.completeSourceRequest = function(...args) {
  sourceCompletionCount += 1;
  return completeSourceRequest(...args);
};
let network = fs.readFileSync(process.argv[4], "utf8")
  .replaceAll("{{.sources.requestBusy}}", "Another request is already in progress. Try again.")
  .replaceAll("{{.sources.transportError}}", "Connection lost before Threadfin confirmed the request. Try again.")
  .replaceAll("{{.sources.responseInvalid}}", "Invalid response");
vm.runInContext(network, context);
const sourceData = {files: {m3u: {M1: {name: "Fixture", "file.source": "https://provider.example/list.m3u?token=" + marker}}}};
context.sourceData = sourceData;

context.beginSourceRequest("m3u", "M1", false, 1);
context.SERVER_CONNECTION = true;
vm.runInContext('new Server("updateFileM3U").request(sourceData)', context);
const busySource = JSON.parse(JSON.stringify(context.sourceFeedbackByKey["m3u:M1"]));

next.disabled = true;
wizardStatus.textContent = "Saving";
routed = false;
context.SERVER_CONNECTION = true;
vm.runInContext('new Server("saveWizard").request({wizard: {tuner: 1}})', context);
const busyWizard = {disabled: next.disabled, status: wizardStatus.textContent, routed};

context.beginSourceRequest("m3u", "M1", false, 1);
context.SERVER_CONNECTION = false;
vm.runInContext('new Server("updateFileM3U").request(sourceData)', context);
const completionCountBeforeTransport = sourceCompletionCount;
const transportSocket = sockets[sockets.length - 1];
transportSocket.onerror({});
transportSocket.onclose({code: 1006});
const transportSource = JSON.parse(JSON.stringify(context.sourceFeedbackByKey["m3u:M1"]));
const transportOneShot = sourceCompletionCount - completionCountBeforeTransport === 1;

next.disabled = true;
wizardStatus.textContent = "Saving";
routed = false;
context.SERVER_CONNECTION = false;
vm.runInContext('new Server("saveWizard").request({wizard: {tuner: 1}})', context);
sockets[sockets.length - 1].onclose({code: 1006});
const transportWizard = {disabled: next.disabled, status: wizardStatus.textContent, routed};

const malformedPayloads = ["{}", "[]", '"text"', '{"status":"true","reload":true}'];
const malformedSourceResults = malformedPayloads.map(payload => {
  context.beginSourceRequest("m3u", "M1", false, 1);
  context.SERVER_CONNECTION = false;
  routed = false;
  vm.runInContext('new Server("updateFileM3U").request(sourceData)', context);
  sockets[sockets.length - 1].onmessage({data: payload});
  const feedback = context.sourceFeedbackByKey["m3u:M1"];
  return feedback && feedback.state === "error" && feedback.message === "Invalid response" && routed === false;
});
const malformedSourcesRejected = malformedSourceResults.every(Boolean);

const malformedWizardResults = malformedPayloads.map(payload => {
  next.disabled = true;
  wizardStatus.textContent = "Saving";
  context.SERVER_CONNECTION = false;
  routed = false;
  vm.runInContext('new Server("saveWizard").request({wizard: {tuner: 1}})', context);
  sockets[sockets.length - 1].onmessage({data: payload});
  return next.disabled === false && wizardStatus.textContent === "Invalid response" && routed === false;
});
const malformedWizardsRejected = malformedWizardResults.every(Boolean);

context.beginSourceRequest("m3u", "M1", false, 1);
context.SERVER_CONNECTION = false;
routed = false;
vm.runInContext('new Server("updateFileM3U").request(sourceData)', context);
sockets[sockets.length - 1].onmessage({data: '{"status":false,"err":"server rejected fixture"}'});
const statusFalseFeedback = context.sourceFeedbackByKey["m3u:M1"];
const statusFalsePreserved = statusFalseFeedback.state === "error" && statusFalseFeedback.message === "server rejected fixture" && routed === false;

next.disabled = true;
wizardStatus.textContent = "Saving";
context.SERVER_CONNECTION = false;
routed = false;
vm.runInContext('new Server("saveWizard").request({wizard: {tuner: 1}})', context);
sockets[sockets.length - 1].onmessage({data: '{"status":true,"reload":true}'});
const normalWizardRouted = routed === true && next.disabled === false;

process.stdout.write(JSON.stringify({busySource, busyWizard, transportSource, transportWizard, logsSafe, malformedSourcesRejected, malformedWizardsRejected, transportOneShot, statusFalsePreserved, normalWizardRouted}));
`

const sourcePopupAccessibilityNodeScript = `
const fs = require("fs");
const vm = require("vm");
class Element {
  constructor(tag = "div") {
    this.tagName = tag.toUpperCase();
    this.attributes = {};
    this.children = [];
    this.parentElement = null;
    this.className = "";
    this.hidden = false;
    this.textContent = "";
    this.value = "";
    this.focusCount = 0;
    this.listeners = {};
    this.classList = {
      add: value => { if (!this.className.split(/\s+/).includes(value)) this.className = (this.className + " " + value).trim(); },
      remove: value => { this.className = this.className.split(/\s+/).filter(name => name && name !== value).join(" "); },
      contains: value => this.className.split(/\s+/).includes(value),
    };
  }
  setAttribute(name, value) { this.attributes[name] = String(value); }
  getAttribute(name) { return this.attributes[name]; }
  removeAttribute(name) { delete this.attributes[name]; }
  appendChild(child) { child.parentElement = this; this.children.push(child); return child; }
  focus() { this.focusCount += 1; }
  addEventListener(name, listener) { this.listeners[name] = listener; }
  closest(selector) { return selector === "tr" ? this.row : null; }
}
function field(tag, name, label, value) {
  const control = new Element(tag);
  control.setAttribute("name", name);
  control.value = value;
  const labelCell = new Element("td");
  labelCell.textContent = label + ":";
  control.row = {querySelector(selector) { return selector === "td:first-child" ? labelCell : null; }};
  const parent = new Element("td");
  parent.appendChild(control);
  return control;
}
const name = field("input", "name", "Source name", "Preserved name");
const source = field("input", "file.source", "Source location", "ftp://provider.example/list.m3u");
const buffer = field("select", "buffer", "Buffer", "-");
const controls = [name, source, buffer];
const popup = new Element("div");
popup.querySelector = function(selector) {
  const match = selector.match(/\[name="([^"]+)"\]/);
  return match ? controls.find(control => control.getAttribute("name") === match[1]) || null : null;
};
function descendants(node) {
  return node.children.reduce((all, child) => all.concat(child, descendants(child)), []);
}
popup.querySelectorAll = function(selector) {
  if (selector === "input, select") return controls;
  if (selector === '[aria-invalid="true"]') return controls.filter(control => control.getAttribute("aria-invalid") === "true");
  if (selector === ".tf-source-field-error") return controls.flatMap(control => descendants(control.parentElement)).filter(element => element.className === "tf-source-field-error");
  return [];
};
const modal = new Element("div");
const byID = id => [popup].concat(descendants(popup), controls.flatMap(control => [control.parentElement].concat(descendants(control.parentElement)))).find(element => element.id === id) || null;
let popupKind = "";
let invokerConnected = true;
const replacement = new Element("button");
replacement.setAttribute("data-source-focus-key", "m3u:M1:edit");
const context = {
  URL,
  console: {log() {}, warn() {}},
  document: {
    createElement(tag) { return new Element(tag); },
    getElementById(id) { if (id === "popup-custom") return popup; if (id === "popup") return modal; return byID(id); },
    contains(element) { return element !== null && (element !== invoker || invokerConnected); },
    querySelectorAll(selector) { return selector === "[data-source-focus-key]" ? [replacement] : []; },
  },
  openPopUp(kind) { popupKind = kind; },
};
vm.createContext(context);
vm.runInContext(fs.readFileSync(process.argv[2], "utf8"), context);
context.enhanceSourcePopup("m3u");
const valid = context.validateSourcePopup("m3u");
const error = byID("source-file-source-error");
const sourceClassApplied = popup.classList.contains("tf-source-popup");
context.enhanceSourcePopup("filter");
const sourceClassReset = sourceClassApplied && !popup.classList.contains("tf-source-popup");
const invoker = new Element("button");
invoker.setAttribute("data-source-focus-key", "m3u:M1:edit");
context.openSourcePopup("m3u", undefined, invoker);
modal.listeners["hidden.bs.modal"]();
invokerConnected = false;
context.openSourcePopup("m3u", undefined, invoker);
modal.listeners["hidden.bs.modal"]();
process.stdout.write(JSON.stringify({
  valid,
  values: {name: name.value, source: source.value},
  labels: {name: name.getAttribute("aria-label"), source: source.getAttribute("aria-label"), buffer: buffer.getAttribute("aria-label")},
  sourceInvalid: source.getAttribute("aria-invalid") === "true",
  sourceFocusCount: source.focusCount,
  errorVisible: !!error && error.hidden === false && error.textContent.length > 0,
  popupKind,
  invokerFocusCount: invoker.focusCount,
  replacementFocusCount: replacement.focusCount,
  sourceClassReset,
}));
`

const wizardModeAccessibilityNodeScript = `
const fs = require("fs");
const vm = require("vm");
class Element {
  constructor(tag = "div") {
    this.tagName = tag.toUpperCase();
    this.children = [];
    this.attributes = {};
    this.textContent = "";
    this.value = "";
    this.name = "";
    this.hidden = false;
    this.disabled = false;
    this.className = "";
  }
  set innerHTML(_value) { this.children = []; }
  get innerHTML() { return ""; }
  appendChild(child) { this.children.push(child); return child; }
  setAttribute(name, value) {
    this.attributes[name] = String(value);
    if (name === "class") this.className = String(value);
    if (name === "name") this.name = String(value);
  }
  getAttribute(name) { return this.attributes[name]; }
  removeAttribute(name) { delete this.attributes[name]; }
}
class PopupContent {
  createInput(type, name, value) {
    const input = new Element("input");
    input.type = type;
    input.name = name;
    input.value = value || "";
    return input;
  }
  createSelect(_text, values, selected, name) {
    const select = new Element("select");
    select.name = name;
    select.value = selected === undefined ? values[0] : selected;
    return select;
  }
}
const content = new Element("div");
const progress = [new Element("li"), new Element("li"), new Element("li"), new Element("li")];
const step = new Element("p");
const next = new Element("input");
const status = new Element("p");
const context = {
  console: {log() {}, warn() {}},
  PopupContent,
  Server: class {request() {}},
  showElement() {},
  sourceLocationAccepted() { return true; },
  document: {
    createElement(tag) { return new Element(tag); },
    getElementById(id) {
      if (id === "content") return content;
      if (id === "wizard-step-status") return step;
      if (id === "next") return next;
      if (id === "wizard-request-status") return status;
      return null;
    },
    querySelectorAll(selector) { return selector === "#wizard-progress li" ? progress : []; },
  },
  SERVER: {settings: {epgSource: "PMS"}},
};
vm.createContext(context);
let source = fs.readFileSync(process.argv[2], "utf8")
  .replaceAll("{{.wizard.progress}}", "Step {current} of {total}")
  .replaceAll("{{.wizard.finish}}", "Finish setup")
  .replaceAll("{{.button.next}}", "Next");
vm.runInContext(source, context);
context.showConfigurationWizard(2);
const pmsField = content.children.find(child => child.className === "wizard");
const pms = {step: step.textContent, action: next.value, xmltvHidden: progress[3].hidden, labelledBy: pmsField.getAttribute("aria-labelledby") || ""};
context.SERVER.settings.epgSource = "XEPG";
context.showConfigurationWizard(3);
const xepgField = content.children.find(child => child.className === "wizard");
const xepg = {step: step.textContent, action: next.value, xmltvHidden: progress[3].hidden, labelledBy: xepgField.getAttribute("aria-labelledby") || ""};
process.stdout.write(JSON.stringify({pms, xepg}));
`

const sourcePayloadNodeScript = `
const fs = require("fs");
const vm = require("vm");
const source = fs.readFileSync(process.argv[2], "utf8");
let controls = {inputs: [], selects: []};
const sent = [];
const marker = "credential-marker";
let leaked = false;
let confirmResult = true;
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
  console: {
    log(...values) { if (values.some(value => JSON.stringify(value).includes(marker))) leaked = true; },
    warn(...values) { if (values.some(value => JSON.stringify(value).includes(marker))) leaked = true; },
  },
  document: {
    getElementById(id) { return id === "popup-custom" ? popup : null; },
    addEventListener() {},
  },
  Server: CaptureServer,
  showElement() {},
  confirm() { return confirmResult; },
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
const m3uInputs = [
  input("text", "name", "Local channels"), input("text", "description", "Primary"),
  input("text", "file.source", "/srv/threadfin/channels.m3u"), input("text", "http_proxy.ip", ""),
  input("text", "http_proxy.port", ""), input("text", "http_headers.origin", "https://origin.example"),
  input("text", "http_headers.referer", "https://referer.example")
];
const m3uSelects = [select("buffer", "ffmpeg"), select("tuner", "3")];
const hdhrInputs = [
  input("text", "name", "Living room tuner"), input("text", "description", "HDHR"),
  input("text", "file.source", "192.0.2.20:5004"), input("text", "http_proxy.ip", ""), input("text", "http_proxy.port", "")
];
const hdhrSelects = [select("buffer", "vlc"), select("tuner", "4")];
const xmltvInputs = [
  input("text", "name", "Guide"), input("text", "description", "Local guide"),
  input("text", "file.source", "C:\\guides\\guide.xml"), input("text", "http_proxy.ip", ""), input("text", "http_proxy.port", "")
];
run("m3u", "-", false, 0, m3uInputs, m3uSelects);
run("m3u", "M1", false, 1, m3uInputs, m3uSelects);
run("m3u", "M1", true, 0, m3uInputs, m3uSelects);
run("hdhr", "-", false, 0, hdhrInputs, hdhrSelects);
run("hdhr", "H1", false, 1, hdhrInputs, hdhrSelects);
run("hdhr", "H1", true, 0, hdhrInputs, hdhrSelects);
run("xmltv", "-", false, 0, xmltvInputs, []);
run("xmltv", "X1", false, 1, xmltvInputs, []);
run("xmltv", "X1", true, 0, xmltvInputs, []);
const sentBeforeCancel = sent.length;
confirmResult = false;
run("xmltv", "X1", true, 0, xmltvInputs, []);
const cancelledNoRequest = sent.length === sentBeforeCancel;
confirmResult = true;
run("m3u", "M2", false, 0, [
  input("text", "name", "Log fixture"), input("text", "description", ""),
  input("text", "file.source", "https://provider.example/list.m3u?token=" + marker),
  input("text", "http_proxy.ip", ""), input("text", "http_proxy.port", ""),
  input("text", "http_headers.origin", ""), input("text", "http_headers.referer", "")
], m3uSelects);
const sensitivePayloadPreserved = sent[sent.length - 1].data.files.m3u.M2["file.source"] === "https://provider.example/list.m3u?token=" + marker;
sent.pop();
process.stdout.write(JSON.stringify({sent, cancelledNoRequest, logsSafe: !leaked, sensitivePayloadPreserved}));
`
