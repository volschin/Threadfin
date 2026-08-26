package src

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"threadfin/src/internal/authentication"
)

type configRoundTripFunc func(*http.Request) (*http.Response, error)

func (roundTrip configRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return roundTrip(request)
}

type configRepeatingReader struct{}

func (configRepeatingReader) Read(buffer []byte) (int, error) {
	for index := range buffer {
		buffer[index] = 'x'
	}
	return len(buffer), nil
}

type configCountingReader struct {
	remaining int
	read      int
}

func (reader *configCountingReader) Read(buffer []byte) (int, error) {
	if reader.remaining == 0 {
		return 0, io.EOF
	}
	count := len(buffer)
	if count > reader.remaining {
		count = reader.remaining
	}
	for index := 0; index < count; index++ {
		buffer[index] = ' '
	}
	reader.remaining -= count
	reader.read += count
	return count, nil
}

type configDataErrorReader struct {
	data []byte
	err  error
}

func (reader *configDataErrorReader) Read(buffer []byte) (int, error) {
	if len(reader.data) == 0 {
		return 0, reader.err
	}
	count := copy(buffer, reader.data)
	reader.data = reader.data[count:]
	if len(reader.data) == 0 {
		return count, reader.err
	}
	return count, nil
}

func TestConfigHTTPIngressStopsAtLimitBeforeClassificationOrAuthorization(t *testing.T) {
	restorePersistentState(t)
	Settings.API = false
	prefix := []byte(`{"cmd":"config.describe"}`)
	reader := &configCountingReader{remaining: configAPIRequestLimit + 4096}
	request := httptest.NewRequest(http.MethodPost, "/api/", io.MultiReader(bytes.NewReader(prefix), reader))
	response := httptest.NewRecorder()
	API(response, request)
	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversize unauthenticated status = %d body = %q, want 413", response.Code, response.Body.String())
	}
	if reader.read > configAPIRequestLimit+1-len(prefix) {
		t.Fatalf("HTTP ingress read %d tail bytes beyond the 64 KiB classification boundary", reader.read)
	}
}

func TestConfigHTTPIngressRejectsReadErrorBeforeConfigDispatch(t *testing.T) {
	restorePersistentState(t)
	Settings = SettingsStruct{API: true, AuthenticationAPI: false}
	readErr := errors.New("injected request read failure")
	request := httptest.NewRequest(http.MethodPost, "/api/", nil)
	request.Body = io.NopCloser(&configDataErrorReader{data: []byte(`{"cmd":"config.describe"}`), err: readErr})
	response := httptest.NewRecorder()
	API(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("read-error status = %d body = %q, want 400 before config dispatch", response.Code, response.Body.String())
	}
	if strings.Contains(response.Body.String(), "configuration") || strings.Contains(response.Body.String(), readErr.Error()) {
		t.Fatalf("read-error response dispatched or leaked cause: %q", response.Body.String())
	}
}

func TestSamePortRestoreCompletesThroughRealRestartPath(t *testing.T) {
	restorePersistentState(t)
	tempRoot := t.TempDir()
	t.Setenv("TMPDIR", tempRoot)
	System = SystemStruct{}
	System.Name = "ThreadfinRestoreTest"
	System.DBVersion = "0.5.0"
	System.Version = "test"
	System.Folder.Config = filepath.Join(t.TempDir(), "config") + string(os.PathSeparator)
	if err := os.MkdirAll(System.Folder.Config, 0o755); err != nil {
		t.Fatalf("create isolated config folder: %v", err)
	}
	tempPath := filepath.Join(tempRoot, "threadfinrestoretest") + string(os.PathSeparator)
	if err := os.MkdirAll(filepath.Join(tempPath, "restore-test"), 0o755); err != nil {
		t.Fatalf("create isolated runtime temp folder: %v", err)
	}
	initialSettings := map[string]interface{}{
		"version":      "0.5.0",
		"port":         "34400",
		"uuid":         "restore-test",
		"temp.path":    tempPath,
		"ssdp":         false,
		"files.update": false,
		"epgSource":    "PMS",
		"files": map[string]interface{}{
			"m3u": map[string]interface{}{}, "hdhr": map[string]interface{}{}, "xmltv": map[string]interface{}{},
		},
		"filter": map[string]interface{}{},
	}
	if err := saveMapToJSONFile(filepath.Join(System.Folder.Config, "settings.json"), initialSettings); err != nil {
		t.Fatalf("seed isolated restore settings: %v", err)
	}
	if err := Init(); err != nil {
		t.Fatalf("initialize isolated restore state: %v", err)
	}
	Settings.Port = "34400"
	Settings.SSDP = false
	Settings.FilesUpdate = false
	Settings.EpgSource = "PMS"
	Settings.Files.M3U = map[string]interface{}{}
	Settings.Files.HDHR = map[string]interface{}{}
	Settings.Files.XMLTV = map[string]interface{}{}
	Settings.Filter = map[int64]interface{}{}
	if err := saveSettings(Settings); err != nil {
		t.Fatalf("persist isolated restore settings: %v", err)
	}
	archiveName, err := ThreadfinBackup()
	if err != nil {
		t.Fatalf("create isolated restore archive: %v", err)
	}
	archive, err := os.ReadFile(filepath.Join(System.Folder.Temp, archiveName))
	if err != nil {
		t.Fatalf("read isolated restore archive: %v", err)
	}
	dataURL := "data:application/zip;base64," + base64.StdEncoding.EncodeToString(archive)

	done := make(chan error, 1)
	go func() {
		_, restoreErr := ThreadfinRestoreFromWeb(dataURL)
		done <- restoreErr
	}()
	select {
	case restoreErr := <-done:
		if restoreErr != nil {
			t.Fatalf("same-port restore failed: %v", restoreErr)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("same-port restore deadlocked while restarting and rebuilding under the configuration mutation lock")
	}
}

func TestConfigRequestDecoderIsStrictAndBounded(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "empty", body: ""},
		{name: "malformed", body: `{"cmd":"config.get"`},
		{name: "trailing", body: `{"cmd":"config.get","token":"x"} {}`},
		{name: "duplicate", body: `{"cmd":"config.get","cmd":"config.describe","token":"x"}`},
		{name: "wrong case", body: `{"Cmd":"config.get","token":"x"}`},
		{name: "unknown", body: `{"cmd":"config.get","token":"x","settings":{}}`},
		{name: "null", body: `null`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := decodeConfigAPIRequest([]byte(test.body)); err == nil {
				t.Fatalf("decode accepted %s input", test.name)
			}
		})
	}
	oversize := []byte(`{"cmd":"config.get","token":"` + strings.Repeat("x", configAPIRequestLimit) + `"}`)
	if _, err := decodeConfigAPIRequest(oversize); !errors.Is(err, errConfigRequestTooLarge) {
		t.Fatalf("oversize decode error = %v, want request too large", err)
	}
}

func TestConfigRequestDecoderEnforcesCommandSpecificPresenceAndRejectsNull(t *testing.T) {
	validAction := `{"type":"onboarding.update","onboarding":{"tuner":2}}`
	tests := []struct {
		name string
		body string
	}{
		{name: "null cmd", body: `{"cmd":null,"token":"t"}`},
		{name: "describe null token", body: `{"cmd":"config.describe","token":null}`},
		{name: "get null token", body: `{"cmd":"config.get","token":null}`},
		{name: "plan null token", body: `{"cmd":"config.plan","token":null,"action":` + validAction + `}`},
		{name: "apply null token", body: `{"cmd":"config.apply","token":null,"action":` + validAction + `,"config_revision":"r","plan_digest":"p"}`},
		{name: "plan null action", body: `{"cmd":"config.plan","token":"t","action":null}`},
		{name: "apply null action", body: `{"cmd":"config.apply","token":"t","action":null,"config_revision":"r","plan_digest":"p"}`},
		{name: "apply null revision", body: `{"cmd":"config.apply","token":"t","action":` + validAction + `,"config_revision":null,"plan_digest":"p"}`},
		{name: "apply null digest", body: `{"cmd":"config.apply","token":"t","action":` + validAction + `,"config_revision":"r","plan_digest":null}`},
		{name: "describe forbidden null action", body: `{"cmd":"config.describe","token":"t","action":null}`},
		{name: "describe forbidden null revision", body: `{"cmd":"config.describe","token":"t","config_revision":null}`},
		{name: "describe forbidden null digest", body: `{"cmd":"config.describe","token":"t","plan_digest":null}`},
		{name: "get forbidden null action", body: `{"cmd":"config.get","token":"t","action":null}`},
		{name: "get forbidden null revision", body: `{"cmd":"config.get","token":"t","config_revision":null}`},
		{name: "get forbidden null digest", body: `{"cmd":"config.get","token":"t","plan_digest":null}`},
		{name: "plan forbidden null revision", body: `{"cmd":"config.plan","token":"t","action":` + validAction + `,"config_revision":null}`},
		{name: "plan forbidden null digest", body: `{"cmd":"config.plan","token":"t","action":` + validAction + `,"plan_digest":null}`},
		{name: "apply forbidden null field", body: `{"cmd":"config.apply","token":"t","action":` + validAction + `,"config_revision":"r","plan_digest":"p","settings":null}`},
		{name: "describe missing token", body: `{"cmd":"config.describe"}`},
		{name: "get missing token", body: `{"cmd":"config.get"}`},
		{name: "plan missing action", body: `{"cmd":"config.plan","token":"t"}`},
		{name: "apply missing action", body: `{"cmd":"config.apply","token":"t","config_revision":"r","plan_digest":"p"}`},
		{name: "apply missing revision", body: `{"cmd":"config.apply","token":"t","action":` + validAction + `,"plan_digest":"p"}`},
		{name: "apply missing digest", body: `{"cmd":"config.apply","token":"t","action":` + validAction + `,"config_revision":"r"}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := decodeConfigAPIRequest([]byte(test.body)); !errors.Is(err, errConfigInvalidRequest) {
				t.Fatalf("decode error = %v, want invalid request for %s", err, test.body)
			}
		})
	}

	valid := []string{
		`{"cmd":"config.describe","token":"t"}`,
		`{"cmd":"config.get","token":"t"}`,
		`{"cmd":"config.plan","token":"t","action":` + validAction + `}`,
		`{"cmd":"config.apply","token":"t","action":` + validAction + `,"config_revision":"r","plan_digest":"p"}`,
	}
	for _, body := range valid {
		if _, err := decodeConfigAPIRequest([]byte(body)); err != nil {
			t.Fatalf("valid command-specific request rejected: %v body=%s", err, body)
		}
	}
}

func initConfigAuthenticationTest(t *testing.T) string {
	t.Helper()
	configDir := t.TempDir()
	System.Folder.Config = filepath.Join(configDir, "config")
	if err := authentication.Init(System.Folder.Config, 60); err != nil {
		t.Fatal(err)
	}
	return configDir
}

type configAuthTestOptions struct {
	apiEnabled     bool
	authAPIEnabled bool
	setupComplete  bool
	apiPermission  bool
	configPermit   bool
}

func setupConfigAPITest(t *testing.T, options configAuthTestOptions) string {
	t.Helper()
	restorePersistentState(t)
	initConfigAuthenticationTest(t)
	Settings = SettingsStruct{API: options.apiEnabled, AuthenticationAPI: options.authAPIEnabled, Tuner: 2, EpgSource: "XEPG"}
	Settings.Files.M3U = map[string]interface{}{}
	Settings.Files.HDHR = map[string]interface{}{}
	Settings.Files.XMLTV = map[string]interface{}{}
	Settings.Filter = map[int64]interface{}{}
	Data.XEPG.Channels = map[string]interface{}{}
	System.ConfigurationWizard = !options.setupComplete
	userID, err := authentication.CreateNewUser("config-agent", "password")
	if err != nil {
		t.Fatal(err)
	}
	if err := authentication.WriteUserData(userID, map[string]interface{}{
		"username":              "config-agent",
		"authentication.api":    options.apiPermission,
		"authentication.config": options.configPermit,
	}); err != nil {
		t.Fatal(err)
	}
	token, err := authentication.UserAuthentication("config-agent", "password")
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func callConfigAPI(t *testing.T, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	API(response, request)
	return response
}

func TestConfigPermissionMissingIsDeniedWithoutMutatingAuthenticationStorage(t *testing.T) {
	configDir := initConfigAuthenticationTest(t)
	userID, err := authentication.CreateNewUser("legacy-user", "password")
	if err != nil {
		t.Fatal(err)
	}
	if err := authentication.WriteUserData(userID, map[string]interface{}{
		"username":           "legacy-user",
		"authentication.api": true,
	}); err != nil {
		t.Fatal(err)
	}
	databasePath := filepath.Join(configDir, "authentication.json")
	before, err := os.ReadFile(databasePath)
	if err != nil {
		t.Fatal(err)
	}

	granted, err := authentication.PermissionGranted(userID, "authentication.config")
	if err != nil {
		t.Fatal(err)
	}
	if granted {
		t.Fatal("missing CONFIG permission granted access")
	}
	after, err := os.ReadFile(databasePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("permission read mutated authentication storage")
	}
}

func TestConfigPermissionDefaultsFalseForExistingAndFirstUsers(t *testing.T) {
	initConfigAuthenticationTest(t)
	legacyID, err := authentication.CreateNewUser("legacy-user", "password")
	if err != nil {
		t.Fatal(err)
	}
	if err := authentication.WriteUserData(legacyID, map[string]interface{}{"username": "legacy-user"}); err != nil {
		t.Fatal(err)
	}

	if err := activatedSystemAuthentication(); err != nil {
		t.Fatal(err)
	}
	legacy, err := authentication.ReadUserData(legacyID)
	if err != nil {
		t.Fatal(err)
	}
	if value, ok := legacy["authentication.config"].(bool); !ok || value {
		t.Fatalf("legacy CONFIG permission = %#v, want false", legacy["authentication.config"])
	}
	newID, err := authentication.CreateNewUser("new-user", "password")
	if err != nil {
		t.Fatal(err)
	}
	newUser, err := authentication.ReadUserData(newID)
	if err != nil {
		t.Fatal(err)
	}
	if value, ok := newUser["authentication.config"].(bool); !ok || value {
		t.Fatalf("new-user CONFIG permission = %#v, want false", newUser["authentication.config"])
	}

	initConfigAuthenticationTest(t)
	System.ConfigurationWizard = true
	if _, err := createFirstUserForAuthentication("first-user", "password"); err != nil {
		t.Fatal(err)
	}
	users, err := authentication.GetAllUserData()
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range users {
		data := value.(map[string]interface{})["data"].(map[string]interface{})
		if config, ok := data["authentication.config"].(bool); !ok || config {
			t.Fatalf("first-user CONFIG permission = %#v, want false", data["authentication.config"])
		}
	}
}

func TestConfigDescribeAuthorizationIsFailClosedAndGeneric(t *testing.T) {
	allowed := configAuthTestOptions{apiEnabled: true, authAPIEnabled: true, setupComplete: true, apiPermission: true, configPermit: true}
	tests := []struct {
		name    string
		options configAuthTestOptions
		token   string
	}{
		{name: "API disabled", options: configAuthTestOptions{authAPIEnabled: true, setupComplete: true, apiPermission: true, configPermit: true}},
		{name: "API authentication disabled", options: configAuthTestOptions{apiEnabled: true, setupComplete: true, apiPermission: true, configPermit: true}},
		{name: "setup incomplete", options: configAuthTestOptions{apiEnabled: true, authAPIEnabled: true, apiPermission: true, configPermit: true}},
		{name: "API permission denied", options: configAuthTestOptions{apiEnabled: true, authAPIEnabled: true, setupComplete: true, configPermit: true}},
		{name: "CONFIG permission denied", options: configAuthTestOptions{apiEnabled: true, authAPIEnabled: true, setupComplete: true, apiPermission: true}},
		{name: "missing token", options: allowed, token: "-"},
		{name: "invalid token", options: allowed, token: "invalid"},
	}
	var denialBody string
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			token := setupConfigAPITest(t, test.options)
			if test.token != "" {
				token = test.token
			}
			body := fmt.Sprintf(`{"cmd":"config.describe","token":"%s"}`, token)
			response := callConfigAPI(t, http.MethodPost, "/api/", body)
			if response.Code != http.StatusForbidden {
				t.Fatalf("status = %d body = %q, want 403", response.Code, response.Body.String())
			}
			if denialBody == "" {
				denialBody = response.Body.String()
			}
			if response.Body.String() != denialBody {
				t.Fatalf("authorization denial differs: got %q want %q", response.Body.String(), denialBody)
			}
			if strings.Contains(response.Body.String(), "token") || strings.Contains(response.Body.String(), "permission") || strings.Contains(response.Body.String(), "setup") {
				t.Fatalf("authorization denial leaks cause: %q", response.Body.String())
			}
		})
	}
}

func TestConfigDescribeReturnsBoundedContractAndRotatesToken(t *testing.T) {
	token := setupConfigAPITest(t, configAuthTestOptions{apiEnabled: true, authAPIEnabled: true, setupComplete: true, apiPermission: true, configPermit: true})
	response := callConfigAPI(t, http.MethodPost, "/api/", fmt.Sprintf(`{"cmd":"config.describe","token":"%s"}`, token))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d body = %q", response.Code, response.Body.String())
	}
	var decoded struct {
		Status      bool   `json:"status"`
		Token       string `json:"token"`
		Description struct {
			SchemaVersion string         `json:"schema_version"`
			Workflow      []string       `json:"workflow"`
			Authority     string         `json:"authority"`
			TokenPolicy   string         `json:"token_policy"`
			Actions       []string       `json:"actions"`
			Bounds        map[string]int `json:"bounds"`
			Exclusions    []string       `json:"exclusions"`
		} `json:"description"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Status || decoded.Token == "" || decoded.Token == token {
		t.Fatalf("describe status/token = %t/%q", decoded.Status, decoded.Token)
	}
	wantWorkflow := []string{"config.describe", "config.get", "config.plan", "config.apply"}
	if fmt.Sprint(decoded.Description.Workflow) != fmt.Sprint(wantWorkflow) || decoded.Description.SchemaVersion == "" {
		t.Fatalf("describe workflow/version = %#v/%q", decoded.Description.Workflow, decoded.Description.SchemaVersion)
	}
	if !strings.Contains(decoded.Description.Authority, "private-LAN") || !strings.Contains(decoded.Description.TokenPolicy, "rotates") {
		t.Fatalf("describe authority/token policy = %q / %q", decoded.Description.Authority, decoded.Description.TokenPolicy)
	}
	if len(decoded.Description.Actions) != 10 || decoded.Description.Bounds["request_bytes"] != 64*1024 || decoded.Description.Bounds["mapping_targets"] != 480 {
		t.Fatalf("describe actions/bounds = %d / %#v", len(decoded.Description.Actions), decoded.Description.Bounds)
	}
	joined := strings.Join(decoded.Description.Exclusions, " ")
	for _, exclusion := range []string{"users", "authentication", "paths", "network", "backup", "updates", "raw maps"} {
		if !strings.Contains(joined, exclusion) {
			t.Errorf("describe exclusions missing %q: %q", exclusion, joined)
		}
	}
	if strings.Contains(response.Body.String(), "config-agent") || strings.Contains(response.Body.String(), "password") {
		t.Fatalf("describe leaked principal or credential: %q", response.Body.String())
	}
	oldTokenResponse := callConfigAPI(t, http.MethodPost, "/api/", fmt.Sprintf(`{"cmd":"config.describe","token":"%s"}`, token))
	if oldTokenResponse.Code != http.StatusForbidden {
		t.Fatalf("old token status = %d, want 403", oldTokenResponse.Code)
	}
}

func TestConfigDescribeLetsAgentDeriveEveryAcceptedActionKeyAndType(t *testing.T) {
	description := describeConfigAPI()
	type expectedAction struct {
		payload  string
		required []string
		optional []string
		fields   map[string]string
	}
	fieldTypes := func(entries ...string) map[string]string {
		fields := make(map[string]string, len(entries)/2)
		for index := 0; index < len(entries); index += 2 {
			fields[entries[index]] = entries[index+1]
		}
		return fields
	}
	expected := map[string]expectedAction{
		"onboarding.update":        {payload: "onboarding", optional: []string{"tuner", "epg_source"}, fields: fieldTypes("tuner", "integer", "epg_source", "string")},
		"source.create":            {payload: "source", required: []string{"source_type", "name", "location"}, optional: []string{"description", "tuner", "buffer"}, fields: fieldTypes("source_type", "string", "name", "string", "location", "string", "description", "string", "tuner", "integer", "buffer", "string")},
		"source.update":            {payload: "source", required: []string{"source_type", "id"}, optional: []string{"name", "description", "location", "tuner", "buffer"}, fields: fieldTypes("source_type", "string", "id", "string", "name", "string", "description", "string", "location", "string", "tuner", "integer", "buffer", "string")},
		"source.delete":            {payload: "source", required: []string{"source_type", "id"}, fields: fieldTypes("source_type", "string", "id", "string")},
		"source.refresh":           {payload: "source", required: []string{"source_type", "id"}, fields: fieldTypes("source_type", "string", "id", "string")},
		"filter.create":            {payload: "filter", required: []string{"type", "name", "filter"}, optional: []string{"description", "case_sensitive", "live_event", "include", "exclude", "starting_number", "category", "active"}, fields: fieldTypes("type", "string", "name", "string", "filter", "string", "description", "string", "case_sensitive", "boolean", "live_event", "boolean", "include", "string", "exclude", "string", "starting_number", "string", "category", "string", "active", "boolean")},
		"filter.update":            {payload: "filter", required: []string{"id"}, optional: []string{"type", "name", "description", "filter", "case_sensitive", "live_event", "include", "exclude", "starting_number", "category", "active"}, fields: fieldTypes("id", "string", "type", "string", "name", "string", "description", "string", "filter", "string", "case_sensitive", "boolean", "live_event", "boolean", "include", "string", "exclude", "string", "starting_number", "string", "category", "string", "active", "boolean")},
		"filter.delete":            {payload: "filter", required: []string{"id"}, fields: fieldTypes("id", "string")},
		"mapping.patch":            {payload: "mapping_patch", required: []string{"id"}, optional: []string{"channel_id", "active", "name", "description", "update_channel_name", "update_channel_icon", "category", "group_title", "xmltv_file", "mapping", "ppv_extra", "backup_channel_1", "backup_channel_2", "backup_channel_3"}, fields: fieldTypes("id", "string", "channel_id", "string", "active", "boolean", "name", "string", "description", "string", "update_channel_name", "boolean", "update_channel_icon", "boolean", "category", "string", "group_title", "string", "xmltv_file", "string", "mapping", "string", "ppv_extra", "string", "backup_channel_1", "string", "backup_channel_2", "string", "backup_channel_3", "string")},
		"mapping.bulk_channel_ids": {payload: "mapping_bulk_channel_ids", required: []string{"items"}, fields: fieldTypes("items", "array")},
	}
	if len(description.ActionSchemas) != len(expected) || description.ActionEnvelope.AdditionalProperties || len(description.ActionEnvelope.ExactlyOne) != 5 {
		t.Fatalf("action schema count/envelope strictness/exclusive payloads = %d/%t/%#v", len(description.ActionSchemas), description.ActionEnvelope.AdditionalProperties, description.ActionEnvelope.ExactlyOne)
	}
	for actionName, want := range expected {
		schema, ok := description.ActionSchemas[actionName]
		if !ok {
			t.Errorf("describe omitted action %q", actionName)
			continue
		}
		if schema.PayloadKey != want.payload || schema.AdditionalProperties {
			t.Errorf("action %q payload/strictness = %q/%t", actionName, schema.PayloadKey, schema.AdditionalProperties)
		}
		gotRequired, gotOptional := append([]string(nil), schema.Required...), append([]string(nil), schema.Optional...)
		wantRequired, wantOptional := append([]string(nil), want.required...), append([]string(nil), want.optional...)
		sort.Strings(gotRequired)
		sort.Strings(gotOptional)
		sort.Strings(wantRequired)
		sort.Strings(wantOptional)
		if fmt.Sprint(gotRequired) != fmt.Sprint(wantRequired) || fmt.Sprint(gotOptional) != fmt.Sprint(wantOptional) {
			t.Errorf("action %q required/optional = %#v/%#v, want %#v/%#v", actionName, gotRequired, gotOptional, wantRequired, wantOptional)
		}
		if len(schema.Fields) != len(want.fields) {
			t.Errorf("action %q field count = %d, want %d", actionName, len(schema.Fields), len(want.fields))
		}
		for field, wantType := range want.fields {
			if got := schema.Fields[field].Type; got != wantType {
				t.Errorf("action %q field %q type = %q, want %q", actionName, field, got, wantType)
			}
		}
	}
	for _, command := range description.Workflow {
		request, requestOK := description.Requests[command]
		response, responseOK := description.Responses[command]
		if !requestOK || !responseOK || request.AdditionalProperties || response.AdditionalProperties {
			t.Errorf("workflow command %q lacks strict request/response schema", command)
		}
	}
	if !description.Authentication.FailClosed || !description.Authentication.SetupComplete || len(description.ResponseModels) == 0 {
		t.Fatalf("auth/response schema incomplete: %#v / %d models", description.Authentication, len(description.ResponseModels))
	}
}

func TestConfigCommandsRequireExactPostAPIPath(t *testing.T) {
	for _, test := range []struct {
		method string
		path   string
		want   int
	}{
		{method: http.MethodGet, path: "/api/", want: http.StatusMethodNotAllowed},
		{method: http.MethodPost, path: "/api/config", want: http.StatusNotFound},
	} {
		t.Run(test.method+test.path, func(t *testing.T) {
			token := setupConfigAPITest(t, configAuthTestOptions{apiEnabled: true, authAPIEnabled: true, setupComplete: true, apiPermission: true, configPermit: true})
			response := callConfigAPI(t, test.method, test.path, fmt.Sprintf(`{"cmd":"config.describe","token":"%s"}`, token))
			if response.Code != test.want {
				t.Fatalf("status = %d body = %q, want %d", response.Code, response.Body.String(), test.want)
			}
		})
	}
}

func TestConfigCanonicalRevisionAndDigestAreDeterministicAndBound(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	left := configHiddenState{
		Settings: map[string]interface{}{
			"files": map[string]interface{}{"m3u": map[string]interface{}{"M1": map[string]interface{}{"name": "One", "secret": "alpha"}}},
			"tuner": float64(2),
		},
		Mapping: map[string]interface{}{"channel-b": map[string]interface{}{"x-name": "B"}, "channel-a": map[string]interface{}{"x-name": "A"}},
	}
	right := configHiddenState{
		Settings: map[string]interface{}{
			"tuner": float64(2),
			"files": map[string]interface{}{"m3u": map[string]interface{}{"M1": map[string]interface{}{"secret": "alpha", "name": "One"}}},
		},
		Mapping: map[string]interface{}{"channel-a": map[string]interface{}{"x-name": "A"}, "channel-b": map[string]interface{}{"x-name": "B"}},
	}
	leftRevision, err := configRevisionWithKey(key, left)
	if err != nil {
		t.Fatal(err)
	}
	rightRevision, err := configRevisionWithKey(key, right)
	if err != nil {
		t.Fatal(err)
	}
	if leftRevision != rightRevision {
		t.Fatalf("field order changed revision: %q != %q", leftRevision, rightRevision)
	}
	right.Settings["files"].(map[string]interface{})["m3u"].(map[string]interface{})["M1"].(map[string]interface{})["secret"] = "beta"
	changedRevision, err := configRevisionWithKey(key, right)
	if err != nil {
		t.Fatal(err)
	}
	if changedRevision == leftRevision {
		t.Fatal("hidden configuration change did not change revision")
	}

	action := json.RawMessage(`{"type":"onboarding.update","onboarding":{"tuner":2}}`)
	consequences := []string{"persist onboarding settings"}
	warnings := []string{"rebuild output data"}
	digest, err := configPlanDigestWithKey(key, "principal-a", leftRevision, action, consequences, warnings)
	if err != nil {
		t.Fatal(err)
	}
	again, err := configPlanDigestWithKey(key, "principal-a", leftRevision, action, consequences, warnings)
	if err != nil || digest != again {
		t.Fatalf("digest is not deterministic: %q / %q / %v", digest, again, err)
	}
	for name, candidate := range map[string]struct {
		principal string
		revision  string
		action    json.RawMessage
	}{
		"principal": {principal: "principal-b", revision: leftRevision, action: action},
		"revision":  {principal: "principal-a", revision: changedRevision, action: action},
		"action":    {principal: "principal-a", revision: leftRevision, action: json.RawMessage(`{"type":"onboarding.update","onboarding":{"tuner":3}}`)},
	} {
		candidateDigest, err := configPlanDigestWithKey(key, candidate.principal, candidate.revision, candidate.action, consequences, warnings)
		if err != nil {
			t.Fatal(err)
		}
		if candidateDigest == digest {
			t.Fatalf("%s difference did not change digest", name)
		}
	}
}

func TestConfigGetReturnsSanitizedTypedStateAndHiddenStateRevision(t *testing.T) {
	token := setupConfigAPITest(t, configAuthTestOptions{apiEnabled: true, authAPIEnabled: true, setupComplete: true, apiPermission: true, configPermit: true})
	Settings.Files.M3U["M1"] = map[string]interface{}{
		"name": "Public list", "description": "Channels", "type": "m3u", "file.source": "https://agent:password@example.test/list.m3u?token=secret#fragment",
		"http_proxy.ip": "proxy-secret", "http_headers.origin": "https://secret-origin.test", "file.threadfin": "/private/data/M1.m3u", "id.provider": "M1",
	}
	Settings.Files.HDHR["H1"] = map[string]interface{}{"name": "Living room", "description": "Tuner", "type": "hdhr", "file.source": "192.168.1.20", "id.provider": "H1"}
	Settings.Files.XMLTV["X1"] = map[string]interface{}{"name": "Guide", "description": "Local", "type": "xmltv", "file.source": "/private/guide.xml", "id.provider": "X1"}
	Settings.Filter[7] = map[string]interface{}{"name": "News", "description": "Keep", "type": "group-title", "filter": "News", "include": "", "exclude": "", "startingNumber": "100", "x-category": "News", "active": true}
	Data.XEPG.Channels["channel-1"] = map[string]interface{}{
		"x-channelID": "101", "x-active": true, "x-name": "One", "x-description": "Description", "x-mapping": "guide-1", "x-xmltv-file": "X1.xml",
		"_uuid.key": "uuid-secret", "_values": "derived-secret", "url": "https://stream.test/live?token=secret", "tvg-logo": "https://logo.test/private.png",
	}

	response := callConfigAPI(t, http.MethodPost, "/api/", fmt.Sprintf(`{"cmd":"config.get","token":"%s"}`, token))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d body = %q", response.Code, response.Body.String())
	}
	var decoded struct {
		Status bool   `json:"status"`
		Token  string `json:"token"`
		State  struct {
			ConfigRevision string `json:"config_revision"`
			Onboarding     struct {
				Tuner     int    `json:"tuner"`
				EpgSource string `json:"epg_source"`
			} `json:"onboarding"`
			Sources []struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Location string `json:"location"`
			} `json:"sources"`
			Filters  []map[string]interface{} `json:"filters"`
			Mappings []map[string]interface{} `json:"mappings"`
		} `json:"state"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Status || decoded.Token == "" || decoded.State.ConfigRevision == "" || decoded.State.Onboarding.Tuner != 2 || decoded.State.Onboarding.EpgSource != "XEPG" {
		t.Fatalf("typed state header = %#v", decoded)
	}
	locations := map[string]string{}
	for _, source := range decoded.State.Sources {
		locations[source.ID] = source.Location
	}
	if locations["M1"] != "https://example.test/list.m3u" || locations["H1"] != "[private-lan source]" || locations["X1"] != "[local source]" {
		t.Fatalf("sanitized locations = %#v", locations)
	}
	if len(decoded.State.Filters) != 1 || len(decoded.State.Mappings) != 1 {
		t.Fatalf("typed filters/mappings = %#v / %#v", decoded.State.Filters, decoded.State.Mappings)
	}
	serialized := response.Body.String()
	for _, secret := range []string{"config-agent", "password", "token=secret", "fragment", "/private/guide.xml", "/private/data", "proxy-secret", "secret-origin", "uuid-secret", "derived-secret", "stream.test", "logo.test", "_values", "_uuid"} {
		if strings.Contains(serialized, secret) {
			t.Errorf("config.get leaked %q in %q", secret, serialized)
		}
	}

	beforeRevision := decoded.State.ConfigRevision
	Settings.Files.M3U["M1"].(map[string]interface{})["http_proxy.ip"] = "changed-hidden-proxy"
	next := callConfigAPI(t, http.MethodPost, "/api/", fmt.Sprintf(`{"cmd":"config.get","token":"%s"}`, decoded.Token))
	var nextDecoded struct {
		State struct {
			ConfigRevision string `json:"config_revision"`
		} `json:"state"`
	}
	if err := json.Unmarshal(next.Body.Bytes(), &nextDecoded); err != nil {
		t.Fatal(err)
	}
	if nextDecoded.State.ConfigRevision == beforeRevision {
		t.Fatal("hidden source configuration change did not invalidate revision")
	}
}

func TestConfigActionValidationRejectsMultipleForbiddenAndOutOfBoundsInput(t *testing.T) {
	setupConfigAPITest(t, configAuthTestOptions{apiEnabled: true, authAPIEnabled: true, setupComplete: true, apiPermission: true, configPermit: true})
	Settings.Files.M3U["M1"] = map[string]interface{}{"id.provider": "M1", "file.source": "https://example.test/list.m3u", "name": "List"}
	Settings.Filter[1] = map[string]interface{}{"type": "group-title", "filter": "News"}
	Data.XEPG.Channels["channel-1"] = map[string]interface{}{"x-channelID": "1"}
	tests := []struct {
		name   string
		action string
	}{
		{name: "null", action: `null`},
		{name: "unknown field", action: `{"type":"onboarding.update","onboarding":{"tuner":2,"password":"secret"}}`},
		{name: "wrong case", action: `{"Type":"onboarding.update","onboarding":{"tuner":2}}`},
		{name: "multiple payloads", action: `{"type":"onboarding.update","onboarding":{"tuner":2},"source":{"source_type":"m3u","name":"x","location":"https://example.test/x"}}`},
		{name: "no update", action: `{"type":"onboarding.update","onboarding":{}}`},
		{name: "bad epg source", action: `{"type":"onboarding.update","onboarding":{"epg_source":"xepg"}}`},
		{name: "tuner low", action: `{"type":"onboarding.update","onboarding":{"tuner":0}}`},
		{name: "local source", action: `{"type":"source.create","source":{"source_type":"m3u","name":"List","location":"/private/list.m3u"}}`},
		{name: "source proxy excluded", action: `{"type":"source.update","source":{"source_type":"m3u","id":"M1","proxy":"http://proxy.test"}}`},
		{name: "mapping logo excluded", action: `{"type":"mapping.patch","mapping_patch":{"id":"channel-1","logo":"https://logo.test/private.png"}}`},
		{name: "bulk item wrong case", action: `{"type":"mapping.bulk_channel_ids","mapping_bulk_channel_ids":{"items":[{"ID":"channel-1","channel_id":"2"}]}}`},
		{name: "unknown source id", action: `{"type":"source.delete","source":{"source_type":"m3u","id":"missing"}}`},
		{name: "filter expression too long", action: `{"type":"filter.update","filter":{"id":"1","type":"group-title","filter":"` + strings.Repeat("a", 2049) + `"}}`},
		{name: "unknown mapping id", action: `{"type":"mapping.patch","mapping_patch":{"id":"missing","name":"x"}}`},
		{name: "too many mappings", action: func() string {
			items := make([]string, 481)
			for index := range items {
				items[index] = fmt.Sprintf(`{"id":"channel-1","channel_id":"%d"}`, index)
			}
			return `{"type":"mapping.bulk_channel_ids","mapping_bulk_channel_ids":{"items":[` + strings.Join(items, ",") + `]}}`
		}()},
		{name: "control character", action: `{"type":"filter.create","filter":{"type":"custom-filter","name":"bad\u0000name","filter":"news"}}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := normalizeConfigAction(json.RawMessage(test.action)); err == nil {
				t.Fatalf("normalization accepted %s action", test.name)
			}
		})
	}
}

func TestConfigActionRejectsExplicitNullForEveryTypedField(t *testing.T) {
	setupConfigActionAdapterState(t)
	tests := []struct {
		name       string
		actionType string
		payloadKey string
		base       map[string]interface{}
		fields     []string
	}{
		{name: "onboarding", actionType: "onboarding.update", payloadKey: "onboarding", base: map[string]interface{}{"tuner": 2}, fields: []string{"tuner", "epg_source"}},
		{name: "source", actionType: "source.update", payloadKey: "source", base: map[string]interface{}{"source_type": "m3u", "id": "M1", "name": "Name"}, fields: []string{"source_type", "id", "name", "description", "location", "tuner", "buffer"}},
		{name: "filter", actionType: "filter.update", payloadKey: "filter", base: map[string]interface{}{"id": "1", "name": "Name"}, fields: []string{"id", "type", "name", "description", "filter", "case_sensitive", "live_event", "include", "exclude", "starting_number", "category", "active"}},
		{name: "mapping patch", actionType: "mapping.patch", payloadKey: "mapping_patch", base: map[string]interface{}{"id": "channel-1", "name": "Name"}, fields: []string{"id", "channel_id", "active", "name", "description", "update_channel_name", "update_channel_icon", "category", "group_title", "xmltv_file", "mapping", "ppv_extra", "backup_channel_1", "backup_channel_2", "backup_channel_3"}},
	}
	for _, test := range tests {
		for _, field := range test.fields {
			t.Run(test.name+"/"+field, func(t *testing.T) {
				payload := make(map[string]interface{}, len(test.base)+1)
				for key, value := range test.base {
					payload[key] = value
				}
				payload[field] = nil
				encoded, err := json.Marshal(map[string]interface{}{"type": test.actionType, test.payloadKey: payload})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := normalizeConfigAction(encoded); err == nil {
					t.Fatalf("normalization accepted explicit null for %s.%s", test.payloadKey, field)
				}
			})
		}
	}
	for name, action := range map[string]string{
		"type":                 `{"type":null,"onboarding":{"tuner":2}}`,
		"payload":              `{"type":"onboarding.update","onboarding":null}`,
		"bulk items":           `{"type":"mapping.bulk_channel_ids","mapping_bulk_channel_ids":{"items":null}}`,
		"bulk item id":         `{"type":"mapping.bulk_channel_ids","mapping_bulk_channel_ids":{"items":[{"id":null,"channel_id":"2"}]}}`,
		"bulk item channel id": `{"type":"mapping.bulk_channel_ids","mapping_bulk_channel_ids":{"items":[{"id":"channel-1","channel_id":null}]}}`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := normalizeConfigAction(json.RawMessage(action)); err == nil {
				t.Fatalf("normalization accepted explicit null in %s", name)
			}
		})
	}
}

func TestConfigMappingPatchValidatesCurrentGuideAndBackupReferences(t *testing.T) {
	setupConfigActionAdapterState(t)
	Data.XEPG.Channels["channel-1"] = map[string]interface{}{
		"x-channelID": "1", "x-name": "Primary", "tvg-name": "Primary", "x-xmltv-file": "X1.xml", "x-mapping": "guide-1",
		"x-backup-channel-1": "-", "x-backup-channel-2": "-", "x-backup-channel-3": "-",
	}
	Data.XEPG.Channels["channel-2"] = map[string]interface{}{"x-channelID": "2", "tvg-name": "Backup"}
	Data.XMLTV.Mapping = map[string]interface{}{
		"X1.xml":          map[string]interface{}{"guide-1": map[string]interface{}{}},
		"Threadfin Dummy": map[string]interface{}{"PPV": map[string]interface{}{}, "60_Minutes": map[string]interface{}{}},
	}
	accepted := []string{
		`{"type":"mapping.patch","mapping_patch":{"id":"channel-1","xmltv_file":"X1.xml","mapping":"guide-1","backup_channel_1":"Backup"}}`,
		`{"type":"mapping.patch","mapping_patch":{"id":"channel-1","xmltv_file":"-","mapping":"-"}}`,
		`{"type":"mapping.patch","mapping_patch":{"id":"channel-1","xmltv_file":"Threadfin Dummy","mapping":"60_Minutes"}}`,
	}
	for _, action := range accepted {
		if _, err := normalizeConfigAction(json.RawMessage(action)); err != nil {
			t.Fatalf("valid current mapping reference rejected: %s: %v", action, err)
		}
	}
	rejected := []string{
		`{"type":"mapping.patch","mapping_patch":{"id":"channel-1","xmltv_file":"missing.xml","mapping":"guide-1"}}`,
		`{"type":"mapping.patch","mapping_patch":{"id":"channel-1","mapping":"missing-guide"}}`,
		`{"type":"mapping.patch","mapping_patch":{"id":"channel-1","backup_channel_1":"Missing Backup"}}`,
		`{"type":"mapping.patch","mapping_patch":{"id":"channel-1","xmltv_file":"-","mapping":"guide-1"}}`,
		`{"type":"mapping.patch","mapping_patch":{"id":"channel-1","xmltv_file":"Threadfin Dummy","mapping":"invalid"}}`,
	}
	for _, action := range rejected {
		if _, err := normalizeConfigAction(json.RawMessage(action)); err == nil {
			t.Fatalf("invalid mapping reference accepted: %s", action)
		}
	}
}

type configPlanTestResponse struct {
	Status bool   `json:"status"`
	Token  string `json:"token"`
	Error  *struct {
		Code string `json:"code"`
	} `json:"error"`
	Plan *struct {
		Action         json.RawMessage `json:"action"`
		Consequences   []string        `json:"consequences"`
		Warnings       []string        `json:"warnings"`
		ConfigRevision string          `json:"config_revision"`
		PlanDigest     string          `json:"plan_digest"`
	} `json:"plan"`
	Apply *struct {
		Persisted      bool   `json:"persisted"`
		Postprocessing string `json:"postprocessing"`
		ConfigRevision string `json:"config_revision"`
	} `json:"apply"`
}

func decodeConfigPlanTestResponse(t *testing.T, response *httptest.ResponseRecorder) configPlanTestResponse {
	t.Helper()
	var decoded configPlanTestResponse
	if err := json.Unmarshal(response.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("decode response %q: %v", response.Body.String(), err)
	}
	return decoded
}

func configPlanBody(token, action string) string {
	return fmt.Sprintf(`{"cmd":"config.plan","token":%q,"action":%s}`, token, action)
}

func configApplyBody(token string, plan configPlanTestResponse, action string) string {
	return fmt.Sprintf(`{"cmd":"config.apply","token":%q,"config_revision":%q,"plan_digest":%q,"action":%s}`,
		token, plan.Plan.ConfigRevision, plan.Plan.PlanDigest, action)
}

func TestConfigPlanIsNonMutatingAndApplyRejectsTamperAndStaleState(t *testing.T) {
	token := setupConfigAPITest(t, configAuthTestOptions{apiEnabled: true, authAPIEnabled: true, setupComplete: true, apiPermission: true, configPermit: true})
	action := `{"type":"onboarding.update","onboarding":{"tuner":3,"epg_source":"PMS"}}`
	settingsBefore, err := json.Marshal(Settings)
	if err != nil {
		t.Fatal(err)
	}
	logsBefore := append([]string(nil), WebScreenLog.Log...)
	var calls atomic.Int32
	previousApply := configDomainApply
	configDomainApply = func(_ context.Context, action configNormalizedAction) (configApplyResult, error) {
		calls.Add(1)
		Settings.Tuner = *action.Onboarding.Tuner
		Settings.EpgSource = *action.Onboarding.EpgSource
		return configApplyResult{Persisted: true, Postprocessing: "rebuilt"}, nil
	}
	t.Cleanup(func() { configDomainApply = previousApply })

	planResponse := callConfigAPI(t, http.MethodPost, "/api/", configPlanBody(token, action))
	if planResponse.Code != http.StatusOK {
		t.Fatalf("plan status = %d body = %q", planResponse.Code, planResponse.Body.String())
	}
	plan := decodeConfigPlanTestResponse(t, planResponse)
	if plan.Plan == nil || plan.Plan.ConfigRevision == "" || plan.Plan.PlanDigest == "" || len(plan.Plan.Consequences) == 0 || len(plan.Plan.Warnings) == 0 {
		t.Fatalf("incomplete plan = %#v", plan)
	}
	settingsAfter, err := json.Marshal(Settings)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(settingsBefore, settingsAfter) || fmt.Sprint(logsBefore) != fmt.Sprint(WebScreenLog.Log) || calls.Load() != 0 {
		t.Fatalf("plan mutated state/log/domain: settings %s -> %s logs %#v -> %#v calls %d", settingsBefore, settingsAfter, logsBefore, WebScreenLog.Log, calls.Load())
	}

	tamperedAction := `{"type":"onboarding.update","onboarding":{"tuner":4,"epg_source":"PMS"}}`
	tamperedResponse := callConfigAPI(t, http.MethodPost, "/api/", configApplyBody(plan.Token, plan, tamperedAction))
	tampered := decodeConfigPlanTestResponse(t, tamperedResponse)
	if tamperedResponse.Code != http.StatusConflict || tampered.Error == nil || tampered.Error.Code != "plan_conflict" || calls.Load() != 0 {
		t.Fatalf("tampered apply = status %d body %q calls %d", tamperedResponse.Code, tamperedResponse.Body.String(), calls.Load())
	}

	planResponse = callConfigAPI(t, http.MethodPost, "/api/", configPlanBody(tampered.Token, action))
	plan = decodeConfigPlanTestResponse(t, planResponse)
	configMutationMutex.Lock()
	Settings.Files.M3U["hidden-browser-change"] = map[string]interface{}{"file.source": "https://example.test/changed.m3u"}
	configMutationMutex.Unlock()
	staleResponse := callConfigAPI(t, http.MethodPost, "/api/", configApplyBody(plan.Token, plan, action))
	stale := decodeConfigPlanTestResponse(t, staleResponse)
	if staleResponse.Code != http.StatusConflict || stale.Error == nil || stale.Error.Code != "plan_conflict" || calls.Load() != 0 {
		t.Fatalf("stale apply = status %d body %q calls %d", staleResponse.Code, staleResponse.Body.String(), calls.Load())
	}
}

func TestConfigApplyCredentialFailuresUseOneFixedSizeGenericConflictContract(t *testing.T) {
	token := setupConfigAPITest(t, configAuthTestOptions{apiEnabled: true, authAPIEnabled: true, setupComplete: true, apiPermission: true, configPermit: true})
	action := `{"type":"onboarding.update","onboarding":{"tuner":3}}`
	plan := decodeConfigPlanTestResponse(t, callConfigAPI(t, http.MethodPost, "/api/", configPlanBody(token, action)))
	if plan.Plan == nil {
		t.Fatal("plan was not returned")
	}
	validRevision, validDigest := plan.Plan.ConfigRevision, plan.Plan.PlanDigest
	tamperedRevision := validRevision[:len(validRevision)-1] + "0"
	if tamperedRevision == validRevision {
		tamperedRevision = validRevision[:len(validRevision)-1] + "1"
	}
	tamperedDigest := validDigest[:len(validDigest)-1] + "0"
	if tamperedDigest == validDigest {
		tamperedDigest = validDigest[:len(validDigest)-1] + "1"
	}
	tests := []struct {
		name     string
		revision string
		digest   string
	}{
		{name: "short revision and malformed digest", revision: "r1:00", digest: "p1:" + strings.Repeat("g", 64)},
		{name: "long revision and short digest", revision: validRevision + "00", digest: "p1:00"},
		{name: "malformed revision and long digest", revision: "r1:" + strings.Repeat("z", 64), digest: validDigest + "00"},
		{name: "valid-size tamper both", revision: tamperedRevision, digest: tamperedDigest},
	}
	var applyCalls atomic.Int32
	previousApply := configDomainApply
	configDomainApply = func(_ context.Context, _ configNormalizedAction) (configApplyResult, error) {
		applyCalls.Add(1)
		return configApplyResult{}, nil
	}
	t.Cleanup(func() { configDomainApply = previousApply })
	currentToken := plan.Token
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := fmt.Sprintf(`{"cmd":"config.apply","token":%q,"config_revision":%q,"plan_digest":%q,"action":%s}`, currentToken, test.revision, test.digest, action)
			response := callConfigAPI(t, http.MethodPost, "/api/", body)
			decoded := decodeConfigPlanTestResponse(t, response)
			if response.Code != http.StatusConflict || decoded.Error == nil || decoded.Error.Code != "plan_conflict" {
				t.Fatalf("credential failure = status %d body %q, want generic 409 plan_conflict", response.Code, response.Body.String())
			}
			if decoded.Token == "" {
				t.Fatal("authorized conflict did not rotate the token")
			}
			currentToken = decoded.Token
		})
	}
	if applyCalls.Load() != 0 {
		t.Fatalf("credential failures executed domain %d times", applyCalls.Load())
	}
	if configPlanCredentialsMatch(validRevision, "r1:00", validDigest, validDigest) || configPlanCredentialsMatch(validRevision, validRevision, validDigest, "p1:00") {
		t.Fatal("fixed-size credential matcher accepted a malformed revision or digest")
	}
}

func TestConfigPlanRedactsSourceCredentialsFromReturnedAction(t *testing.T) {
	token := setupConfigAPITest(t, configAuthTestOptions{apiEnabled: true, authAPIEnabled: true, setupComplete: true, apiPermission: true, configPermit: true})
	action := `{"type":"source.create","source":{"source_type":"m3u","name":"Private","location":"https://agent:password@source.example/list.m3u?token=secret#fragment"}}`
	response := callConfigAPI(t, http.MethodPost, "/api/", fmt.Sprintf(`{"cmd":"config.plan","token":"%s","action":%s}`, token, action))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d body = %q", response.Code, response.Body.String())
	}
	for _, secret := range []string{"agent", "password", "token=secret", "fragment"} {
		if strings.Contains(response.Body.String(), secret) {
			t.Fatalf("plan response leaked %q: %s", secret, response.Body.String())
		}
	}
	if !strings.Contains(response.Body.String(), `https://source.example/list.m3u`) {
		t.Fatalf("plan response did not retain a safe source marker: %s", response.Body.String())
	}
}

func TestConfigSensitiveSourceLocationIsWriteOnlyAndRequiresOriginalActionReplay(t *testing.T) {
	token := setupConfigAPITest(t, configAuthTestOptions{apiEnabled: true, authAPIEnabled: true, setupComplete: true, apiPermission: true, configPermit: true})
	secrets := []string{"s3cr3t-user", "s3cr3t-pass", "access_token=top-secret", "private-fragment"}
	action := `{"type":"source.create","source":{"source_type":"m3u","name":"Private","location":"https://s3cr3t-user:s3cr3t-pass@source.example/list.m3u?access_token=top-secret#private-fragment"}}`

	describeResponse := callConfigAPI(t, http.MethodPost, "/api/", fmt.Sprintf(`{"cmd":"config.describe","token":%q}`, token))
	if describeResponse.Code != http.StatusOK {
		t.Fatalf("describe status = %d body = %q", describeResponse.Code, describeResponse.Body.String())
	}
	var describe struct {
		Token       string `json:"token"`
		Description struct {
			SensitiveInputPolicy struct {
				Fields                      []string `json:"fields"`
				AcceptedURLComponents       []string `json:"accepted_url_components"`
				Returned                    bool     `json:"returned"`
				Logged                      bool     `json:"logged"`
				BoundTo                     []string `json:"bound_to"`
				PlanActionReplayable        bool     `json:"plan_action_replayable"`
				ApplyRequiresOriginalAction bool     `json:"apply_requires_original_action"`
			} `json:"sensitive_input_policy"`
		} `json:"description"`
	}
	if err := json.Unmarshal(describeResponse.Body.Bytes(), &describe); err != nil {
		t.Fatal(err)
	}
	policy := describe.Description.SensitiveInputPolicy
	if fmt.Sprint(policy.Fields) != fmt.Sprint([]string{"action.source.location"}) || fmt.Sprint(policy.AcceptedURLComponents) != fmt.Sprint([]string{"userinfo", "query", "fragment"}) || policy.Returned || policy.Logged || policy.PlanActionReplayable || !policy.ApplyRequiresOriginalAction || fmt.Sprint(policy.BoundTo) != fmt.Sprint([]string{"plan_digest", "config_revision"}) {
		t.Fatalf("sensitive input policy is incomplete: %#v", policy)
	}
	for _, forbiddenClaim := range []string{"URL credentials are forbidden", "userinfo is forbidden", "query is forbidden", "fragment is forbidden"} {
		if strings.Contains(describeResponse.Body.String(), forbiddenClaim) {
			t.Fatalf("describe contains incompatible IPTV URL claim %q", forbiddenClaim)
		}
	}

	getResponse := callConfigAPI(t, http.MethodPost, "/api/", fmt.Sprintf(`{"cmd":"config.get","token":%q}`, describe.Token))
	get := decodeConfigPlanTestResponse(t, getResponse)
	planResponse := callConfigAPI(t, http.MethodPost, "/api/", configPlanBody(get.Token, action))
	plan := decodeConfigPlanTestResponse(t, planResponse)
	if planResponse.Code != http.StatusOK || plan.Plan == nil {
		t.Fatalf("secret source plan = status %d body %q", planResponse.Code, planResponse.Body.String())
	}
	redactedAction := string(plan.Plan.Action)
	for _, secret := range secrets {
		for label, value := range map[string]string{"describe": describeResponse.Body.String(), "get": getResponse.Body.String(), "plan": planResponse.Body.String(), "plan action": redactedAction} {
			if strings.Contains(value, secret) {
				t.Fatalf("%s leaked %q: %s", label, secret, value)
			}
		}
	}

	var applyCalls atomic.Int32
	var appliedLocation string
	previousApply := configDomainApply
	configDomainApply = func(_ context.Context, normalized configNormalizedAction) (configApplyResult, error) {
		applyCalls.Add(1)
		appliedLocation = *normalized.Source.Location
		return configApplyResult{Persisted: true, Postprocessing: "rebuilt"}, nil
	}
	t.Cleanup(func() { configDomainApply = previousApply })

	redactedResponse := callConfigAPI(t, http.MethodPost, "/api/", configApplyBody(plan.Token, plan, redactedAction))
	redacted := decodeConfigPlanTestResponse(t, redactedResponse)
	if redactedResponse.Code != http.StatusConflict || redacted.Error == nil || redacted.Error.Code != "plan_conflict" {
		t.Fatalf("redacted display action replay = status %d body %q", redactedResponse.Code, redactedResponse.Body.String())
	}
	alteredAction := strings.Replace(action, "access_token=top-secret", "access_token=changed-secret", 1)
	alteredResponse := callConfigAPI(t, http.MethodPost, "/api/", configApplyBody(redacted.Token, plan, alteredAction))
	altered := decodeConfigPlanTestResponse(t, alteredResponse)
	if alteredResponse.Code != http.StatusConflict || altered.Error == nil || altered.Error.Code != "plan_conflict" {
		t.Fatalf("altered secret replay = status %d body %q", alteredResponse.Code, alteredResponse.Body.String())
	}
	exactResponse := callConfigAPI(t, http.MethodPost, "/api/", configApplyBody(altered.Token, plan, action))
	exact := decodeConfigPlanTestResponse(t, exactResponse)
	if exactResponse.Code != http.StatusOK || exact.Apply == nil || applyCalls.Load() != 1 || appliedLocation != "https://s3cr3t-user:s3cr3t-pass@source.example/list.m3u?access_token=top-secret#private-fragment" {
		t.Fatalf("exact original replay = status %d body %q calls/location %d/%q", exactResponse.Code, exactResponse.Body.String(), applyCalls.Load(), appliedLocation)
	}
	for _, response := range []*httptest.ResponseRecorder{redactedResponse, alteredResponse, exactResponse} {
		for _, secret := range secrets {
			if strings.Contains(response.Body.String(), secret) {
				t.Fatalf("apply response leaked %q: %s", secret, response.Body.String())
			}
		}
	}
	logs := strings.Join(WebScreenLog.Log, "\n")
	for _, secret := range secrets {
		if strings.Contains(logs, secret) {
			t.Fatalf("configuration workflow log leaked %q: %s", secret, logs)
		}
	}
}

func TestConfigConcurrentIdenticalAppliesExecuteExactlyOnce(t *testing.T) {
	token := setupConfigAPITest(t, configAuthTestOptions{apiEnabled: true, authAPIEnabled: true, setupComplete: true, apiPermission: true, configPermit: true})
	action := `{"type":"onboarding.update","onboarding":{"tuner":9}}`
	plan := decodeConfigPlanTestResponse(t, callConfigAPI(t, http.MethodPost, "/api/", configPlanBody(token, action)))
	secondToken, err := authentication.UserAuthentication("config-agent", "password")
	if err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	previousApply := configDomainApply
	configDomainApply = func(_ context.Context, action configNormalizedAction) (configApplyResult, error) {
		calls.Add(1)
		Settings.Tuner = *action.Onboarding.Tuner
		return configApplyResult{Persisted: true, Postprocessing: "rebuilt"}, nil
	}
	t.Cleanup(func() { configDomainApply = previousApply })

	tokens := []string{plan.Token, secondToken}
	responses := make([]*httptest.ResponseRecorder, 2)
	var wait sync.WaitGroup
	wait.Add(2)
	for index := range responses {
		go func(index int) {
			defer wait.Done()
			responses[index] = callConfigAPI(t, http.MethodPost, "/api/", configApplyBody(tokens[index], plan, action))
		}(index)
	}
	wait.Wait()
	statuses := []int{responses[0].Code, responses[1].Code}
	sort.Ints(statuses)
	if fmt.Sprint(statuses) != fmt.Sprint([]int{http.StatusOK, http.StatusConflict}) || calls.Load() != 1 || Settings.Tuner != 9 {
		t.Fatalf("concurrent applies statuses/calls/tuner = %#v/%d/%d", statuses, calls.Load(), Settings.Tuner)
	}
	for _, response := range responses {
		decoded := decodeConfigPlanTestResponse(t, response)
		if response.Code == http.StatusOK && (decoded.Apply == nil || !decoded.Apply.Persisted || decoded.Apply.ConfigRevision == "") {
			t.Fatalf("successful apply result = %q", response.Body.String())
		}
	}
}

func TestConfigRevocationSerializedBeforeWaitingApplyPreventsExecution(t *testing.T) {
	token := setupConfigAPITest(t, configAuthTestOptions{apiEnabled: true, authAPIEnabled: true, setupComplete: true, apiPermission: true, configPermit: true})
	action := `{"type":"onboarding.update","onboarding":{"tuner":9}}`
	plan := decodeConfigPlanTestResponse(t, callConfigAPI(t, http.MethodPost, "/api/", configPlanBody(token, action)))
	if plan.Plan == nil {
		t.Fatal("plan was not returned")
	}
	userID, err := authentication.GetUserID(plan.Token)
	if err != nil {
		t.Fatal(err)
	}
	var applyCalls atomic.Int32
	previousApply := configDomainApply
	configDomainApply = func(_ context.Context, _ configNormalizedAction) (configApplyResult, error) {
		applyCalls.Add(1)
		return configApplyResult{Persisted: true, Postprocessing: "rebuilt"}, nil
	}
	t.Cleanup(func() { configDomainApply = previousApply })

	configMutationMutex.Lock()
	requestStarted := make(chan struct{})
	responseDone := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		close(requestStarted)
		request := httptest.NewRequest(http.MethodPost, "/api/", strings.NewReader(configApplyBody(plan.Token, plan, action)))
		response := httptest.NewRecorder()
		API(response, request)
		responseDone <- response
	}()
	<-requestStarted
	revoke := RequestStruct{UserData: map[string]interface{}{
		userID: map[string]interface{}{"authentication.api": true, "authentication.config": false},
	}}
	if err := saveUserDataConfigLocked(revoke); err != nil {
		configMutationMutex.Unlock()
		t.Fatalf("serialize CONFIG permission revocation: %v", err)
	}
	configMutationMutex.Unlock()

	select {
	case response := <-responseDone:
		decoded := decodeConfigPlanTestResponse(t, response)
		if response.Code != http.StatusForbidden || decoded.Error == nil || decoded.Error.Code != "forbidden" {
			t.Fatalf("apply after serialized revocation = status %d body %q", response.Code, response.Body.String())
		}
		if applyCalls.Load() != 0 {
			t.Fatalf("revoked CONFIG apply executed domain %d times", applyCalls.Load())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("apply did not resume after serialized permission revocation")
	}
}

func setupConfigActionAdapterState(t *testing.T) {
	t.Helper()
	setupConfigAPITest(t, configAuthTestOptions{apiEnabled: true, authAPIEnabled: true, setupComplete: true, apiPermission: true, configPermit: true})
	Settings.Files.M3U["M1"] = map[string]interface{}{"id.provider": "M1", "type": "m3u", "name": "M3U", "file.source": "https://example.test/old.m3u", "unknown": "keep"}
	Settings.Files.HDHR["H1"] = map[string]interface{}{"id.provider": "H1", "type": "hdhr", "name": "HDHR", "file.source": "192.168.1.20", "unknown": "keep"}
	Settings.Files.XMLTV["X1"] = map[string]interface{}{"id.provider": "X1", "type": "xmltv", "name": "XMLTV", "file.source": "https://example.test/old.xml", "unknown": "keep"}
	Settings.Filter[1] = map[string]interface{}{"type": "group-title", "name": "Group", "filter": "News", "unknown": "keep"}
	Settings.Filter[2] = map[string]interface{}{"type": "custom-filter", "name": "Custom", "filter": "tvg-id", "unknown": "keep"}
	Data.XEPG.Channels["channel-1"] = map[string]interface{}{"x-channelID": "1", "x-name": "Original", "unknown": map[string]interface{}{"deep": "preserved"}}
	Data.XEPG.Channels["channel-2"] = map[string]interface{}{"x-channelID": "2", "x-name": "Other", "unknown": "preserved"}
}

func setupIsolatedConfigDomainState(t *testing.T) {
	t.Helper()
	restorePersistentState(t)
	root := t.TempDir()
	System = SystemStruct{}
	System.Name = "Threadfin"
	System.AppName = "threadfin"
	System.PlexChannelLimit = 480
	System.UnfilteredChannelLimit = 480
	System.ServerProtocol.WEB = "http"
	System.Folder.Config = root + string(os.PathSeparator)
	System.Folder.Data = filepath.Join(root, "data") + string(os.PathSeparator)
	System.Folder.ImagesCache = filepath.Join(root, "cache", "images") + string(os.PathSeparator)
	System.File.Settings = filepath.Join(root, "settings.json")
	System.File.URLS = filepath.Join(root, "urls.json")
	System.File.XEPG = filepath.Join(root, "xepg.json")
	System.File.M3U = filepath.Join(root, "threadfin.m3u")
	System.File.XML = filepath.Join(root, "threadfin.xml")
	System.Compressed.GZxml = filepath.Join(root, "threadfin.xml.gz")
	System.Domain = "127.0.0.1:34400"
	for _, folder := range []string{System.Folder.Data, System.Folder.ImagesCache} {
		if err := os.MkdirAll(folder, 0o755); err != nil {
			t.Fatalf("create isolated domain folder: %v", err)
		}
	}
	Settings = SettingsStruct{}
	Settings.UUID = "config-domain-test"
	Settings.TempPath = root + string(os.PathSeparator)
	Settings.BackupPath = filepath.Join(root, "backup") + string(os.PathSeparator)
	Settings.EpgSource = "PMS"
	Settings.Tuner = 1
	Settings.Files.M3U = map[string]interface{}{}
	Settings.Files.HDHR = map[string]interface{}{}
	Settings.Files.XMLTV = map[string]interface{}{}
	Settings.Filter = map[int64]interface{}{}
	Data = DataStruct{}
	Data.Cache.StreamingURLS = map[string]StreamInfo{}
	Data.XEPG.Channels = map[string]interface{}{}
	if err := saveSettings(Settings); err != nil {
		t.Fatalf("persist isolated domain settings: %v", err)
	}
	if err := saveMapToJSONFile(System.File.URLS, map[string]interface{}{}); err != nil {
		t.Fatalf("persist isolated URL cache: %v", err)
	}
}

func TestConfigExistingSourceRefreshReturnsRealFetchFailure(t *testing.T) {
	setupIsolatedConfigDomainState(t)
	Settings.Files.M3U["M1"] = map[string]interface{}{
		"id.provider": "M1", "type": "m3u", "name": "Existing", "description": "", "file.source": "http://127.0.0.1/rejected.m3u",
		"file.threadfin": "M1.m3u", "tuner": float64(1), "http_proxy.ip": "", "http_proxy.port": "", "counter.error": float64(0), "counter.download": float64(0), "provider.availability": 100,
	}
	if err := saveSettings(Settings); err != nil {
		t.Fatal(err)
	}
	normalized, err := normalizeConfigAction(json.RawMessage(`{"type":"source.refresh","source":{"source_type":"m3u","id":"M1"}}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := applyConfigActionDomain(context.TODO(), normalized)
	if err == nil {
		t.Fatal("existing-provider refresh reported success after the real CONFIG fetch policy rejected its target")
	}
	if result.Persisted || result.Postprocessing != "failed" {
		t.Fatalf("failed refresh result = %#v, want unpersisted/failed", result)
	}
	stored := Settings.Files.M3U["M1"].(map[string]interface{})
	if stored["counter.error"] != float64(1) || stored["counter.download"] != float64(1) {
		t.Fatalf("failed refresh counters were not persisted by the real provider path: %#v", stored)
	}
	if legacyErr := getProviderDataWithOptions("m3u", "M1", browserProviderFetchOptions()); legacyErr != nil {
		t.Fatalf("legacy existing-provider refresh no longer retains its historical cached-file success behavior: %v", legacyErr)
	}
}

func TestConfigReportsPersistedMutationWhenRealPostprocessingFails(t *testing.T) {
	setupIsolatedConfigDomainState(t)
	System.File.URLS = filepath.Join(t.TempDir(), "missing", "urls.json")
	normalized, err := normalizeConfigAction(json.RawMessage(`{"type":"onboarding.update","onboarding":{"tuner":2,"epg_source":"PMS"}}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := applyConfigActionDomain(context.TODO(), normalized)
	if err == nil {
		t.Fatal("real postprocessing path reported success with an unwritable URL cache destination")
	}
	if !result.Persisted || result.Postprocessing != "failed" || Settings.Tuner != 2 {
		t.Fatalf("postprocessing failure result/settings = %#v / tuner %d, want persisted=true failed and tuner=2", result, Settings.Tuner)
	}
}

func TestConfigSuccessfulRefreshReportsQueuedOutputWork(t *testing.T) {
	domain := configDomainRequest{Kind: "source.refresh"}
	if got := configDomainPostprocessing(domain, ""); got != "queued" {
		t.Fatalf("successful refresh postprocessing = %q, want queued", got)
	}
	if configDomainRequestPersisted(domain) {
		t.Fatal("refresh persistence must be derived from successful execution, not stale existing state")
	}
}

func TestConfigRealDomainAdaptersPersistRepresentativeMutationsAndUnknownMappingFields(t *testing.T) {
	setupIsolatedConfigDomainState(t)
	previousOptions := configProviderOptions
	configProviderOptions = func(ctx context.Context) providerFetchOptions {
		return providerFetchOptions{
			Context: ctx, MaxTimeout: time.Second, RedactSource: true,
			Resolve: func(context.Context, string) ([]net.IP, error) { return []net.IP{net.ParseIP("203.0.113.20")}, nil },
			Transport: configRoundTripFunc(func(request *http.Request) (*http.Response, error) {
				body := "#EXTM3U\n#EXTINF:-1 tvg-id=\"news\" tvg-name=\"News\",News\nhttp://stream.example/news\n"
				return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body)), Request: request}, nil
			}),
		}
	}
	t.Cleanup(func() { configProviderOptions = previousOptions })

	apply := func(actionJSON string) configApplyResult {
		t.Helper()
		normalized, err := normalizeConfigAction(json.RawMessage(actionJSON))
		if err != nil {
			t.Fatalf("normalize representative action %s: %v", actionJSON, err)
		}
		result, err := applyConfigActionDomain(context.TODO(), normalized)
		if err != nil {
			t.Fatalf("apply representative action %s: %v (result %#v)", actionJSON, err, result)
		}
		if !result.Persisted || result.Postprocessing == "failed" {
			t.Fatalf("representative action result = %#v", result)
		}
		return result
	}

	onboarding := apply(`{"type":"onboarding.update","onboarding":{"tuner":2}}`)
	if onboarding.Postprocessing != "not_required" || Settings.Tuner != 2 {
		t.Fatalf("real onboarding result/settings = %#v / %d", onboarding, Settings.Tuner)
	}
	source := apply(`{"type":"source.create","source":{"source_type":"m3u","name":"Injected source","location":"https://source.example/list.m3u","tuner":1,"buffer":"-"}}`)
	if source.Postprocessing != "rebuilt" || len(Settings.Files.M3U) != 1 {
		t.Fatalf("real source result/settings = %#v / %#v", source, Settings.Files.M3U)
	}
	for sourceID := range Settings.Files.M3U {
		if _, err := os.Stat(filepath.Join(System.Folder.Data, sourceID+".m3u")); err != nil {
			t.Fatalf("real provider file was not persisted: %v", err)
		}
	}

	Settings.Filter[1] = map[string]interface{}{"type": "group-title", "name": "Old", "filter": "News", "unknown": "filter-preserved"}
	if err := saveSettings(Settings); err != nil {
		t.Fatal(err)
	}
	filter := apply(`{"type":"filter.update","filter":{"id":"1","name":"Renamed"}}`)
	storedFilter := Settings.Filter[1].(map[string]interface{})
	if filter.Postprocessing != "rebuilt" || storedFilter["name"] != "Renamed" || storedFilter["unknown"] != "filter-preserved" {
		t.Fatalf("real filter result/state = %#v / %#v", filter, storedFilter)
	}

	Data.XEPG.Channels = map[string]interface{}{
		"channel-1": map[string]interface{}{
			"x-channelID": "1", "x-active": true, "x-name": "Original", "tvg-name": "News", "name": "News", "url": "http://stream.example/news",
			"x-xmltv-file": "-", "x-mapping": "-", "unknown": map[string]interface{}{"deep": "preserved"},
		},
	}
	Data.Streams.Active = []interface{}{map[string]interface{}{"_file.m3u.id": "M1", "name": "News", "tvg-name": "News", "url": "http://stream.example/news"}}
	mapping := apply(`{"type":"mapping.patch","mapping_patch":{"id":"channel-1","name":"Mapped"}}`)
	storedMapping := Data.XEPG.Channels["channel-1"].(map[string]interface{})
	unknown, ok := storedMapping["unknown"].(map[string]interface{})
	if mapping.Postprocessing != "rebuilt" || storedMapping["x-name"] != "Mapped" || !ok || unknown["deep"] != "preserved" {
		t.Fatalf("real mapping result/state = %#v / %#v", mapping, storedMapping)
	}
	persistedMapping, err := loadJSONFileToMap(System.File.XEPG)
	if err != nil {
		t.Fatal(err)
	}
	persistedUnknown := persistedMapping["channel-1"].(map[string]interface{})["unknown"].(map[string]interface{})
	if persistedUnknown["deep"] != "preserved" {
		t.Fatalf("persisted full-map unknown field was lost: %#v", persistedMapping["channel-1"])
	}
}

func TestScheduledAndPPVBrowserMutationsExecuteBehindSharedConfigLock(t *testing.T) {
	for _, test := range []struct {
		name   string
		setup  func(t *testing.T)
		mutate func()
	}{
		{name: "scheduled", setup: setupIsolatedConfigDomainState, mutate: runScheduledConfigMutation},
		{name: "ppv", setup: func(t *testing.T) {
			setupIsolatedConfigDomainState(t)
			mapping := map[string]interface{}{"channel": map[string]interface{}{"x-mapping": "PPV", "x-active": false}}
			if err := saveMapToJSONFile(System.File.XEPG, mapping); err != nil {
				t.Fatal(err)
			}
		}, mutate: func() {
			response := httptest.NewRecorder()
			enablePPV(response, httptest.NewRequest(http.MethodPost, "/ppv/enable", nil))
			if response.Code != http.StatusOK {
				panic(fmt.Sprintf("PPV response status %d", response.Code))
			}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			test.setup(t)
			configMutationMutex.Lock()
			done := make(chan struct{})
			go func() {
				defer close(done)
				test.mutate()
			}()
			select {
			case <-done:
				configMutationMutex.Unlock()
				t.Fatalf("%s mutation bypassed the shared configuration lock", test.name)
			case <-time.After(20 * time.Millisecond):
			}
			configMutationMutex.Unlock()
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Fatalf("%s mutation did not complete after shared lock release", test.name)
			}
		})
	}
}

func TestConfigAllowlistedActionsBuildOneNarrowDomainRequest(t *testing.T) {
	setupConfigActionAdapterState(t)
	tests := []struct {
		name       string
		action     string
		wantKind   string
		wantType   string
		wantID     string
		wantDelete bool
	}{
		{name: "onboarding", action: `{"type":"onboarding.update","onboarding":{"tuner":4,"epg_source":"PMS"}}`, wantKind: "onboarding"},
		{name: "m3u create", action: `{"type":"source.create","source":{"source_type":"m3u","name":"New M3U","description":"Description","location":"https://example.test/new.m3u","tuner":2,"buffer":"ffmpeg"}}`, wantKind: "source.save", wantType: "m3u", wantID: "-"},
		{name: "m3u update", action: `{"type":"source.update","source":{"source_type":"m3u","id":"M1","name":"Renamed"}}`, wantKind: "source.save", wantType: "m3u", wantID: "M1"},
		{name: "m3u delete", action: `{"type":"source.delete","source":{"source_type":"m3u","id":"M1"}}`, wantKind: "source.save", wantType: "m3u", wantID: "M1", wantDelete: true},
		{name: "m3u refresh", action: `{"type":"source.refresh","source":{"source_type":"m3u","id":"M1"}}`, wantKind: "source.refresh", wantType: "m3u", wantID: "M1"},
		{name: "hdhr create", action: `{"type":"source.create","source":{"source_type":"hdhr","name":"New HDHR","location":"192.168.1.30","tuner":2,"buffer":"-"}}`, wantKind: "source.save", wantType: "hdhr", wantID: "-"},
		{name: "hdhr update", action: `{"type":"source.update","source":{"source_type":"hdhr","id":"H1","description":"Renamed"}}`, wantKind: "source.save", wantType: "hdhr", wantID: "H1"},
		{name: "hdhr delete", action: `{"type":"source.delete","source":{"source_type":"hdhr","id":"H1"}}`, wantKind: "source.save", wantType: "hdhr", wantID: "H1", wantDelete: true},
		{name: "hdhr refresh", action: `{"type":"source.refresh","source":{"source_type":"hdhr","id":"H1"}}`, wantKind: "source.refresh", wantType: "hdhr", wantID: "H1"},
		{name: "xmltv create", action: `{"type":"source.create","source":{"source_type":"xmltv","name":"New XMLTV","location":"https://example.test/new.xml"}}`, wantKind: "source.save", wantType: "xmltv", wantID: "-"},
		{name: "xmltv update", action: `{"type":"source.update","source":{"source_type":"xmltv","id":"X1","description":"Renamed"}}`, wantKind: "source.save", wantType: "xmltv", wantID: "X1"},
		{name: "xmltv delete", action: `{"type":"source.delete","source":{"source_type":"xmltv","id":"X1"}}`, wantKind: "source.save", wantType: "xmltv", wantID: "X1", wantDelete: true},
		{name: "xmltv refresh", action: `{"type":"source.refresh","source":{"source_type":"xmltv","id":"X1"}}`, wantKind: "source.refresh", wantType: "xmltv", wantID: "X1"},
		{name: "group create", action: `{"type":"filter.create","filter":{"type":"group-title","name":"Sports","filter":"Sports","include":"HD","starting_number":"100"}}`, wantKind: "filter", wantID: "-1"},
		{name: "group update", action: `{"type":"filter.update","filter":{"id":"1","name":"News HD"}}`, wantKind: "filter", wantID: "1"},
		{name: "group delete", action: `{"type":"filter.delete","filter":{"id":"1"}}`, wantKind: "filter", wantID: "1", wantDelete: true},
		{name: "custom create", action: `{"type":"filter.create","filter":{"type":"custom-filter","name":"Custom 2","filter":"tvg-id"}}`, wantKind: "filter", wantID: "-1"},
		{name: "custom update", action: `{"type":"filter.update","filter":{"id":"2","description":"Changed"}}`, wantKind: "filter", wantID: "2"},
		{name: "custom delete", action: `{"type":"filter.delete","filter":{"id":"2"}}`, wantKind: "filter", wantID: "2", wantDelete: true},
		{name: "mapping patch", action: `{"type":"mapping.patch","mapping_patch":{"id":"channel-1","channel_id":"11","name":"Changed","active":true}}`, wantKind: "mapping"},
		{name: "mapping bulk", action: `{"type":"mapping.bulk_channel_ids","mapping_bulk_channel_ids":{"items":[{"id":"channel-1","channel_id":"21"},{"id":"channel-2","channel_id":"22"}]}}`, wantKind: "mapping"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			normalized, err := normalizeConfigAction(json.RawMessage(test.action))
			if err != nil {
				t.Fatal(err)
			}
			domain, err := buildConfigDomainRequest(normalized)
			if err != nil {
				t.Fatal(err)
			}
			if domain.Kind != test.wantKind || domain.FileType != test.wantType {
				t.Fatalf("domain kind/type = %q/%q, want %q/%q", domain.Kind, domain.FileType, test.wantKind, test.wantType)
			}
			if test.wantType != "" {
				var values map[string]interface{}
				switch test.wantType {
				case "m3u":
					values = domain.Request.Files.M3U[test.wantID].(map[string]interface{})
				case "hdhr":
					values = domain.Request.Files.HDHR[test.wantID].(map[string]interface{})
				case "xmltv":
					values = domain.Request.Files.XMLTV[test.wantID].(map[string]interface{})
				}
				if (values["delete"] == true) != test.wantDelete {
					t.Fatalf("source delete = %#v, want %t", values["delete"], test.wantDelete)
				}
				for _, forbidden := range []string{"http_proxy.ip", "http_proxy.port", "http_headers.origin", "http_headers.referer", "path", "raw"} {
					if _, exists := values[forbidden]; exists {
						t.Fatalf("source request contains forbidden key %q: %#v", forbidden, values)
					}
				}
			}
			if test.wantKind == "filter" {
				id, _ := strconv.ParseInt(test.wantID, 10, 64)
				values := domain.Request.Filter[id].(map[string]interface{})
				if (values["delete"] == true) != test.wantDelete {
					t.Fatalf("filter delete = %#v, want %t", values["delete"], test.wantDelete)
				}
			}
			if test.wantKind == "mapping" {
				one := domain.Request.EpgMapping["channel-1"].(map[string]interface{})
				if one["unknown"].(map[string]interface{})["deep"] != "preserved" || domain.Request.EpgMapping["channel-2"].(map[string]interface{})["unknown"] != "preserved" {
					t.Fatalf("mapping request lost unknown fields: %#v", domain.Request.EpgMapping)
				}
			}
		})
	}
}

func TestConfigApplyCallsExactlyOneDomainSeamAndReportsPersistedFailure(t *testing.T) {
	setupConfigActionAdapterState(t)
	normalized, err := normalizeConfigAction(json.RawMessage(`{"type":"mapping.patch","mapping_patch":{"id":"channel-1","name":"Changed"}}`))
	if err != nil {
		t.Fatal(err)
	}
	previous := configExecuteDomainRequest
	var calls int
	configExecuteDomainRequest = func(request configDomainRequest) (configApplyResult, error) {
		calls++
		if request.Kind != "mapping" {
			t.Fatalf("domain kind = %q", request.Kind)
		}
		return configApplyResult{Persisted: true, Postprocessing: "failed"}, errors.New("postprocessing failed")
	}
	t.Cleanup(func() { configExecuteDomainRequest = previous })
	result, err := applyConfigActionDomain(context.TODO(), normalized)
	if err == nil || calls != 1 || !result.Persisted || result.Postprocessing != "failed" {
		t.Fatalf("domain seam result/calls/error = %#v/%d/%v", result, calls, err)
	}
}

func TestProviderDestinationPolicyBlocksRestrictedTargetsAndKeepsPrivateLAN(t *testing.T) {
	tests := []struct {
		address string
		want    bool
	}{
		{address: "127.0.0.1", want: false},
		{address: "::1", want: false},
		{address: "0.0.0.0", want: false},
		{address: "::", want: false},
		{address: "169.254.1.2", want: false},
		{address: "fe80::1", want: false},
		{address: "224.0.0.1", want: false},
		{address: "ff02::1", want: false},
		{address: "169.254.169.254", want: false},
		{address: "100.100.100.200", want: false},
		{address: "168.63.129.16", want: false},
		{address: "fd00:ec2::254", want: false},
		{address: "10.0.0.2", want: true},
		{address: "172.16.0.2", want: true},
		{address: "192.168.1.20", want: true},
		{address: "fd12:3456::1", want: true},
		{address: "203.0.113.10", want: true},
	}
	for _, test := range tests {
		t.Run(test.address, func(t *testing.T) {
			if got := providerDestinationAllowed(net.ParseIP(test.address)); got != test.want {
				t.Fatalf("provider destination %s allowed = %t, want %t", test.address, got, test.want)
			}
		})
	}
}

func TestProviderSafeDialUsesValidatedResolvedAddressAndRejectsRebinding(t *testing.T) {
	dialed := ""
	dialErr := errors.New("dial stopped")
	dial := func(_ context.Context, _, address string) (net.Conn, error) {
		dialed = address
		return nil, dialErr
	}
	resolvePrivate := func(context.Context, string) ([]net.IP, error) { return []net.IP{net.ParseIP("192.168.1.30")}, nil }
	_, err := providerSafeDialContext(resolvePrivate, dial)(context.Background(), "tcp", "tuner.test:80")
	if !errors.Is(err, dialErr) || dialed != "192.168.1.30:80" {
		t.Fatalf("safe private dial = address %q error %v", dialed, err)
	}

	dialed = ""
	resolveRebound := func(context.Context, string) ([]net.IP, error) { return []net.IP{net.ParseIP("127.0.0.1")}, nil }
	_, err = providerSafeDialContext(resolveRebound, dial)(context.Background(), "tcp", "public-looking.test:80")
	if err == nil || dialed != "" || strings.Contains(err.Error(), "127.0.0.1") || strings.Contains(err.Error(), "public-looking.test") {
		t.Fatalf("rebound dial leaked or dialed restricted target: address %q error %v", dialed, err)
	}
}

func TestProviderRedirectPolicyRevalidatesEveryDestination(t *testing.T) {
	resolve := func(_ context.Context, host string) ([]net.IP, error) {
		if host == "private.test" {
			return []net.IP{net.ParseIP("192.168.1.40")}, nil
		}
		return []net.IP{net.ParseIP("127.0.0.1")}, nil
	}
	policy := providerRedirectPolicy(resolve)
	privateRequest, _ := http.NewRequest(http.MethodGet, "http://private.test/lineup.json", nil)
	if err := policy(privateRequest, nil); err != nil {
		t.Fatalf("private-LAN redirect rejected: %v", err)
	}
	restrictedRequest, _ := http.NewRequest(http.MethodGet, "http://blocked.test/metadata", nil)
	err := policy(restrictedRequest, nil)
	if err == nil || strings.Contains(err.Error(), "blocked.test") || strings.Contains(err.Error(), "metadata") {
		t.Fatalf("restricted redirect error = %v", err)
	}
	if err := policy(privateRequest, make([]*http.Request, 10)); err == nil {
		t.Fatal("redirect limit was not enforced")
	}
}

func TestProviderDownloadBoundsBodiesTimeoutsAndSanitizesErrors(t *testing.T) {
	resolve := func(context.Context, string) ([]net.IP, error) { return []net.IP{net.ParseIP("203.0.113.10")}, nil }
	oversizeTransport := configRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(io.LimitReader(configRepeatingReader{}, providerResponseBodyLimit+1)), Request: request}, nil
	})
	_, _, err := downloadFileFromServerWithOptions("https://user:password@example.test/list.m3u?token=secret#fragment", "", providerFetchOptions{Resolve: resolve, Transport: oversizeTransport})
	if !errors.Is(err, errProviderResponseTooLarge) || strings.Contains(err.Error(), "password") || strings.Contains(err.Error(), "token=secret") || strings.Contains(err.Error(), "fragment") {
		t.Fatalf("oversize error = %v", err)
	}

	timeoutTransport := configRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		<-request.Context().Done()
		return nil, request.Context().Err()
	})
	_, _, err = downloadFileFromServerWithOptions("https://example.test/list.m3u?secret=value", "", providerFetchOptions{Resolve: resolve, Transport: timeoutTransport, Timeout: time.Millisecond})
	if err == nil || strings.Contains(err.Error(), "secret=value") {
		t.Fatalf("timeout error = %v", err)
	}

	failureTransport := configRoundTripFunc(func(*http.Request) (*http.Response, error) { return nil, errors.New("transport failed") })
	_, _, err = downloadFileFromServerWithOptions("https://user:password@example.test/list.m3u?token=secret", "", providerFetchOptions{Resolve: resolve, Transport: failureTransport})
	if err == nil || strings.Contains(err.Error(), "user") || strings.Contains(err.Error(), "password") || strings.Contains(err.Error(), "token=secret") || strings.Contains(err.Error(), "transport failed") {
		t.Fatalf("sanitized transport error = %v", err)
	}
}

func TestProviderGZIPExpansionIsBoundedWithSecretFreeError(t *testing.T) {
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	if _, err := writer.Write(bytes.Repeat([]byte("A"), 4096)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	body, err := extractGZIPBounded(compressed.Bytes(), "https://user:password@source.example/list.m3u?token=secret", 1024)
	if !errors.Is(err, errProviderResponseTooLarge) || len(body) != 0 {
		t.Fatalf("bounded gzip result = %d bytes / %v, want empty size-limit error", len(body), err)
	}
	for _, secret := range []string{"user", "password", "source.example", "token=secret"} {
		if strings.Contains(err.Error(), secret) {
			t.Fatalf("gzip error leaked %q: %v", secret, err)
		}
	}
}

func TestConfigProviderFetchForbidsProxyWhileBrowserPolicyKeepsIt(t *testing.T) {
	configOptions := configProviderFetchOptions(context.TODO())
	if configOptions.AllowProxy || configOptions.AllowLocal || !configOptions.RedactSource {
		t.Fatalf("CONFIG fetch options allow proxy/local sources: %#v", configOptions)
	}
	browserOptions := browserProviderFetchOptions()
	if !browserOptions.AllowProxy || !browserOptions.AllowLocal || browserOptions.RedactSource {
		t.Fatalf("browser fetch options lost proxy/local support: %#v", browserOptions)
	}
	resolve := func(context.Context, string) ([]net.IP, error) { return []net.IP{net.ParseIP("203.0.113.10")}, nil }
	_, _, err := downloadFileFromServerWithOptions("https://example.test/list.m3u", "http://proxy.test:8080", providerFetchOptions{Resolve: resolve, Transport: configRoundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("CONFIG proxy request reached transport")
		return nil, nil
	})})
	if !errors.Is(err, errProviderProxyForbidden) || strings.Contains(err.Error(), "proxy.test") {
		t.Fatalf("CONFIG proxy error = %v", err)
	}
	configOptions.Resolve = resolve
	configOptions.Transport = configRoundTripFunc(func(*http.Request) (*http.Response, error) { return nil, errors.New("private transport detail") })
	_, _, err = downloadFileFromServerWithOptions("https://user:password@source.example/private/list.m3u?token=secret", "", configOptions)
	for _, leak := range []string{"user", "password", "source.example", "private/list", "token=secret", "transport detail"} {
		if err == nil || strings.Contains(err.Error(), leak) {
			t.Fatalf("CONFIG fetch error leaked %q: %v", leak, err)
		}
	}
}

func TestConfigProviderTimeoutCoversResolutionAndParentCancellation(t *testing.T) {
	restorePersistentState(t)
	Settings.BufferTimeout = 1e30
	resolverStarted := make(chan struct{}, 1)
	stuckResolver := func(ctx context.Context, _ string) ([]net.IP, error) {
		resolverStarted <- struct{}{}
		<-ctx.Done()
		return nil, ctx.Err()
	}
	options := configProviderFetchOptions(context.TODO())
	options.Resolve = stuckResolver
	options.MaxTimeout = 20 * time.Millisecond
	started := time.Now()
	_, _, err := downloadFileFromServerWithOptions("https://stuck.example/list.m3u", "", options)
	if err == nil || time.Since(started) > time.Second {
		t.Fatalf("stuck CONFIG resolver result = %v after %s, want bounded cancellation", err, time.Since(started))
	}
	select {
	case <-resolverStarted:
	default:
		t.Fatal("CONFIG timeout did not cover initial DNS resolution")
	}

	parent, cancel := context.WithCancel(context.TODO())
	cancel()
	options = configProviderFetchOptions(parent)
	options.Resolve = stuckResolver
	started = time.Now()
	_, _, err = downloadFileFromServerWithOptions("https://cancelled.example/list.m3u", "", options)
	if err == nil || time.Since(started) > time.Second {
		t.Fatalf("cancelled CONFIG fetch result = %v after %s", err, time.Since(started))
	}
}

func TestBrowserProviderProxyExecutesWithoutResolvingOriginOrApplyingOriginPolicyToProxy(t *testing.T) {
	var requests atomic.Int32
	proxy := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requests.Add(1)
		response.Header().Set("Content-Type", "application/x-mpegURL")
		_, _ = response.Write([]byte("#EXTM3U\n"))
	}))
	defer proxy.Close()

	filename, body, err := downloadFileFromServerWithOptions("http://origin.invalid/private/list.m3u", proxy.URL, browserProviderFetchOptions())
	if err != nil {
		t.Fatalf("browser proxy fetch failed: %v", err)
	}
	if requests.Load() != 1 || filename != "list.m3u" || string(body) != "#EXTM3U\n" {
		t.Fatalf("browser proxy result = requests %d filename %q body %q", requests.Load(), filename, body)
	}
}

func TestBrowserAndScheduledMutationsInvalidatePlansThroughSharedLock(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func()
	}{
		{name: "browser", mutate: func() {
			unlock := lockConfigMutationForCommand("saveSettings")
			Settings.Tuner++
			unlock()
		}},
		{name: "scheduled", mutate: func() {
			withScheduledConfigMutation(func() {
				Settings.Files.M3U["scheduled"] = map[string]interface{}{"file.source": "https://example.test/scheduled.m3u"}
			})
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			token := setupConfigAPITest(t, configAuthTestOptions{apiEnabled: true, authAPIEnabled: true, setupComplete: true, apiPermission: true, configPermit: true})
			action := `{"type":"onboarding.update","onboarding":{"tuner":9}}`
			plan := decodeConfigPlanTestResponse(t, callConfigAPI(t, http.MethodPost, "/api/", configPlanBody(token, action)))
			test.mutate()
			response := callConfigAPI(t, http.MethodPost, "/api/", configApplyBody(plan.Token, plan, action))
			decoded := decodeConfigPlanTestResponse(t, response)
			if response.Code != http.StatusConflict || decoded.Error == nil || decoded.Error.Code != "plan_conflict" {
				t.Fatalf("apply after %s mutation = status %d body %q", test.name, response.Code, response.Body.String())
			}
		})
	}
}

func TestGenericRevisionParticipatingBrowserCommandsTakeConfigLock(t *testing.T) {
	for _, command := range []string{"saveSettings", "saveFilesM3U", "updateFileM3U", "saveFilesHDHR", "updateFileHDHR", "saveFilesXMLTV", "updateFileXMLTV", "saveFilter", "saveEpgMapping", "saveWizard"} {
		if !configMutationCommand(command) {
			t.Errorf("mutation command %q does not participate in config lock", command)
		}
	}
	for _, command := range []string{"getServerConfig", "updateLog", "saveUserData", "saveNewUser", "resetLogs", "ThreadfinBackup", "ThreadfinRestore", "probeChannel"} {
		if configMutationCommand(command) {
			t.Errorf("unrelated command %q unexpectedly takes config lock", command)
		}
	}
}
