package src

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"threadfin/src/internal/authentication"
)

const (
	configAPISchemaVersion = "1"
	configAPIRequestLimit  = 64 * 1024
)

var (
	errConfigRequestTooLarge = errors.New("configuration request is too large")
	errConfigInvalidRequest  = errors.New("invalid configuration request")
	configMutationMutex      sync.Mutex
	configHMACKey            = newConfigHMACKey()
)

func newConfigHMACKey() []byte {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		panic("cannot initialize configuration plan key")
	}
	return key
}

type configAPIRequest struct {
	Cmd            string          `json:"cmd"`
	Token          string          `json:"token"`
	Action         json.RawMessage `json:"action,omitempty"`
	ConfigRevision string          `json:"config_revision,omitempty"`
	PlanDigest     string          `json:"plan_digest,omitempty"`
}

type configDescription struct {
	SchemaVersion        string                        `json:"schema_version"`
	Workflow             []string                      `json:"workflow"`
	Authority            string                        `json:"authority"`
	TokenPolicy          string                        `json:"token_policy"`
	SensitiveInputPolicy configSensitiveInputPolicy    `json:"sensitive_input_policy"`
	Authentication       configAuthenticationSchema    `json:"authentication"`
	Requests             map[string]configObjectSchema `json:"requests"`
	ActionEnvelope       configObjectSchema            `json:"action_envelope"`
	ActionSchemas        map[string]configActionSchema `json:"action_schemas"`
	Responses            map[string]configObjectSchema `json:"responses"`
	ResponseModels       map[string]configObjectSchema `json:"response_models"`
	Actions              []string                      `json:"actions"`
	Bounds               map[string]int                `json:"bounds"`
	Exclusions           []string                      `json:"exclusions"`
}

type configSensitiveInputPolicy struct {
	Fields                      []string `json:"fields"`
	AcceptedURLComponents       []string `json:"accepted_url_components"`
	Returned                    bool     `json:"returned"`
	Logged                      bool     `json:"logged"`
	BoundTo                     []string `json:"bound_to"`
	PlanActionReplayable        bool     `json:"plan_action_replayable"`
	ApplyRequiresOriginalAction bool     `json:"apply_requires_original_action"`
}

type configAuthenticationSchema struct {
	RequiredSettings    []string `json:"required_settings"`
	RequiredPermissions []string `json:"required_permissions"`
	SetupComplete       bool     `json:"setup_complete"`
	FailClosed          bool     `json:"fail_closed"`
}

type configFieldSchema struct {
	Type        string              `json:"type,omitempty"`
	Ref         string              `json:"ref,omitempty"`
	Description string              `json:"description,omitempty"`
	WriteOnly   bool                `json:"write_only,omitempty"`
	Sensitive   bool                `json:"sensitive,omitempty"`
	DisplayOnly bool                `json:"display_only,omitempty"`
	Enum        []string            `json:"enum,omitempty"`
	Minimum     *int                `json:"minimum,omitempty"`
	Maximum     *int                `json:"maximum,omitempty"`
	MinLength   int                 `json:"min_length,omitempty"`
	MaxLength   int                 `json:"max_length,omitempty"`
	Pattern     string              `json:"pattern,omitempty"`
	MinItems    int                 `json:"min_items,omitempty"`
	MaxItems    int                 `json:"max_items,omitempty"`
	Items       *configObjectSchema `json:"items,omitempty"`
	ItemsRef    string              `json:"items_ref,omitempty"`
}

type configObjectSchema struct {
	Type                 string                       `json:"type"`
	Required             []string                     `json:"required"`
	Optional             []string                     `json:"optional"`
	Properties           map[string]configFieldSchema `json:"properties"`
	AdditionalProperties bool                         `json:"additional_properties"`
	MinProperties        int                          `json:"min_properties,omitempty"`
	ExactlyOne           []string                     `json:"exactly_one,omitempty"`
}

type configActionSchema struct {
	PayloadKey           string                       `json:"payload_key"`
	Required             []string                     `json:"required"`
	Optional             []string                     `json:"optional"`
	AtLeastOne           []string                     `json:"at_least_one,omitempty"`
	Constraints          []string                     `json:"constraints,omitempty"`
	Fields               map[string]configFieldSchema `json:"fields"`
	AdditionalProperties bool                         `json:"additional_properties"`
	MinProperties        int                          `json:"min_properties,omitempty"`
}

type configAPIResponse struct {
	Status      bool               `json:"status"`
	Token       string             `json:"token,omitempty"`
	Error       *configAPIError    `json:"error,omitempty"`
	Description *configDescription `json:"description,omitempty"`
	State       *configStateDTO    `json:"state,omitempty"`
	Plan        *configPlanDTO     `json:"plan,omitempty"`
	Apply       *configApplyDTO    `json:"apply,omitempty"`
}

type configAPIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type configHiddenState struct {
	Settings      map[string]interface{} `json:"settings"`
	Mapping       map[string]interface{} `json:"mapping"`
	SetupComplete bool                   `json:"setup_complete"`
}

type configOnboardingDTO struct {
	Tuner     int    `json:"tuner"`
	EpgSource string `json:"epg_source"`
}

type configSourceDTO struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Location    string `json:"location"`
	Tuner       int    `json:"tuner,omitempty"`
	Buffer      string `json:"buffer,omitempty"`
}

type configFilterDTO struct {
	ID             string `json:"id"`
	Type           string `json:"type"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	Filter         string `json:"filter"`
	CaseSensitive  bool   `json:"case_sensitive"`
	LiveEvent      bool   `json:"live_event"`
	Include        string `json:"include,omitempty"`
	Exclude        string `json:"exclude,omitempty"`
	StartingNumber string `json:"starting_number,omitempty"`
	Category       string `json:"category,omitempty"`
	Active         bool   `json:"active"`
}

type configMappingDTO struct {
	ID                string `json:"id"`
	ChannelID         string `json:"channel_id,omitempty"`
	Active            bool   `json:"active"`
	Name              string `json:"name,omitempty"`
	Description       string `json:"description,omitempty"`
	UpdateChannelName bool   `json:"update_channel_name"`
	UpdateChannelIcon bool   `json:"update_channel_icon"`
	Category          string `json:"category,omitempty"`
	GroupTitle        string `json:"group_title,omitempty"`
	XMLTVFile         string `json:"xmltv_file,omitempty"`
	Mapping           string `json:"mapping,omitempty"`
	PPVExtra          string `json:"ppv_extra,omitempty"`
	BackupChannel1    string `json:"backup_channel_1,omitempty"`
	BackupChannel2    string `json:"backup_channel_2,omitempty"`
	BackupChannel3    string `json:"backup_channel_3,omitempty"`
}

type configStateDTO struct {
	ConfigRevision string              `json:"config_revision"`
	Onboarding     configOnboardingDTO `json:"onboarding"`
	Sources        []configSourceDTO   `json:"sources"`
	Filters        []configFilterDTO   `json:"filters"`
	Mappings       []configMappingDTO  `json:"mappings"`
}

type configOnboardingAction struct {
	Tuner     *int    `json:"tuner,omitempty"`
	EpgSource *string `json:"epg_source,omitempty"`
}

type configSourceAction struct {
	SourceType  string  `json:"source_type"`
	ID          string  `json:"id,omitempty"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Location    *string `json:"location,omitempty"`
	Tuner       *int    `json:"tuner,omitempty"`
	Buffer      *string `json:"buffer,omitempty"`
}

type configFilterAction struct {
	ID             string  `json:"id,omitempty"`
	Type           string  `json:"type,omitempty"`
	Name           *string `json:"name,omitempty"`
	Description    *string `json:"description,omitempty"`
	Filter         *string `json:"filter,omitempty"`
	CaseSensitive  *bool   `json:"case_sensitive,omitempty"`
	LiveEvent      *bool   `json:"live_event,omitempty"`
	Include        *string `json:"include,omitempty"`
	Exclude        *string `json:"exclude,omitempty"`
	StartingNumber *string `json:"starting_number,omitempty"`
	Category       *string `json:"category,omitempty"`
	Active         *bool   `json:"active,omitempty"`
}

type configMappingPatchAction struct {
	ID                string  `json:"id"`
	ChannelID         *string `json:"channel_id,omitempty"`
	Active            *bool   `json:"active,omitempty"`
	Name              *string `json:"name,omitempty"`
	Description       *string `json:"description,omitempty"`
	UpdateChannelName *bool   `json:"update_channel_name,omitempty"`
	UpdateChannelIcon *bool   `json:"update_channel_icon,omitempty"`
	Category          *string `json:"category,omitempty"`
	GroupTitle        *string `json:"group_title,omitempty"`
	XMLTVFile         *string `json:"xmltv_file,omitempty"`
	Mapping           *string `json:"mapping,omitempty"`
	PPVExtra          *string `json:"ppv_extra,omitempty"`
	BackupChannel1    *string `json:"backup_channel_1,omitempty"`
	BackupChannel2    *string `json:"backup_channel_2,omitempty"`
	BackupChannel3    *string `json:"backup_channel_3,omitempty"`
}

type configMappingBulkItem struct {
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
}

type configMappingBulkAction struct {
	Items []configMappingBulkItem `json:"items"`
}

type configNormalizedAction struct {
	Type                  string                    `json:"type"`
	Onboarding            *configOnboardingAction   `json:"onboarding,omitempty"`
	Source                *configSourceAction       `json:"source,omitempty"`
	Filter                *configFilterAction       `json:"filter,omitempty"`
	MappingPatch          *configMappingPatchAction `json:"mapping_patch,omitempty"`
	MappingBulkChannelIDs *configMappingBulkAction  `json:"mapping_bulk_channel_ids,omitempty"`
}

type configPlanDTO struct {
	Action         json.RawMessage `json:"action"`
	Consequences   []string        `json:"consequences"`
	Warnings       []string        `json:"warnings"`
	ConfigRevision string          `json:"config_revision"`
	PlanDigest     string          `json:"plan_digest"`
}

type configApplyDTO struct {
	Persisted      bool   `json:"persisted"`
	Postprocessing string `json:"postprocessing"`
	ConfigRevision string `json:"config_revision"`
}

type configApplyResult struct {
	Persisted      bool
	Postprocessing string
}

var configDomainApply = applyConfigActionDomain

func configMutationCommand(command string) bool {
	switch command {
	case "saveSettings", "saveFilesM3U", "updateFileM3U", "saveFilesHDHR", "updateFileHDHR", "saveFilesXMLTV", "updateFileXMLTV", "saveFilter", "saveEpgMapping", "saveWizard", "update.m3u", "update.hdhr", "update.xmltv", "update.xepg":
		return true
	default:
		return false
	}
}

func lockConfigMutationForCommand(command string) func() {
	if !configMutationCommand(command) {
		return func() {}
	}
	configMutationMutex.Lock()
	return configMutationMutex.Unlock
}

func withScheduledConfigMutation(mutate func()) {
	configMutationMutex.Lock()
	defer configMutationMutex.Unlock()
	mutate()
}

func configHMAC(key, content []byte) string {
	digest := hmac.New(sha256.New, key)
	_, _ = digest.Write(content)
	return hex.EncodeToString(digest.Sum(nil))
}

func configDecodeTaggedMAC(value string, prefix [3]byte) ([sha256.Size]byte, int) {
	var raw [3 + sha256.Size*2]byte
	copy(raw[:], value)
	valid := subtle.ConstantTimeEq(int32(len(value)), int32(len(raw)))
	valid &= subtle.ConstantTimeCompare(raw[:3], prefix[:])
	var decoded [sha256.Size]byte
	for index := range decoded {
		high, highValid := configHexNibble(raw[3+index*2])
		low, lowValid := configHexNibble(raw[3+index*2+1])
		decoded[index] = high<<4 | low
		valid &= highValid & lowValid
	}
	return decoded, valid
}

func configHexNibble(value byte) (byte, int) {
	numeric := subtle.ConstantTimeLessOrEq(int('0'), int(value)) & subtle.ConstantTimeLessOrEq(int(value), int('9'))
	lower := value | 0x20
	alpha := subtle.ConstantTimeLessOrEq(int('a'), int(lower)) & subtle.ConstantTimeLessOrEq(int(lower), int('f'))
	numericValue := int(value - '0')
	alphaValue := int(lower-'a') + 10
	decoded := subtle.ConstantTimeSelect(numeric, numericValue, alphaValue)
	return byte(decoded), numeric | alpha
}

func configTaggedMACMatches(expected, supplied string, prefix [3]byte) int {
	expectedMAC, expectedValid := configDecodeTaggedMAC(expected, prefix)
	suppliedMAC, suppliedValid := configDecodeTaggedMAC(supplied, prefix)
	return expectedValid & suppliedValid & subtle.ConstantTimeCompare(expectedMAC[:], suppliedMAC[:])
}

func configPlanCredentialsMatch(expectedRevision, suppliedRevision, expectedDigest, suppliedDigest string) bool {
	revisionMatch := configTaggedMACMatches(expectedRevision, suppliedRevision, [3]byte{'r', '1', ':'})
	digestMatch := configTaggedMACMatches(expectedDigest, suppliedDigest, [3]byte{'p', '1', ':'})
	return revisionMatch&digestMatch == 1
}

func configRevisionWithKey(key []byte, state configHiddenState) (string, error) {
	canonical, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	return "r1:" + configHMAC(key, canonical), nil
}

func configPlanDigestWithKey(key []byte, principal, revision string, action json.RawMessage, consequences, warnings []string) (string, error) {
	var normalizedAction interface{}
	if err := json.Unmarshal(action, &normalizedAction); err != nil {
		return "", err
	}
	canonical, err := json.Marshal(struct {
		Version      string      `json:"version"`
		Principal    string      `json:"principal"`
		Revision     string      `json:"revision"`
		Action       interface{} `json:"action"`
		Consequences []string    `json:"consequences"`
		Warnings     []string    `json:"warnings"`
	}{
		Version: configAPISchemaVersion, Principal: principal, Revision: revision,
		Action: normalizedAction, Consequences: consequences, Warnings: warnings,
	})
	if err != nil {
		return "", err
	}
	return "p1:" + configHMAC(key, canonical), nil
}

func snapshotConfigHiddenState() (configHiddenState, error) {
	settings, err := interfaceToMap(Settings)
	if err != nil {
		return configHiddenState{}, err
	}
	mappingJSON, err := json.Marshal(Data.XEPG.Channels)
	if err != nil {
		return configHiddenState{}, err
	}
	mapping := make(map[string]interface{})
	if len(mappingJSON) > 0 && string(mappingJSON) != "null" {
		if err := json.Unmarshal(mappingJSON, &mapping); err != nil {
			return configHiddenState{}, err
		}
	}
	return configHiddenState{Settings: settings, Mapping: mapping, SetupComplete: !System.ConfigurationWizard}, nil
}

func configString(data map[string]interface{}, key string) string {
	value, _ := data[key].(string)
	return value
}

func configBool(data map[string]interface{}, key string) bool {
	value, _ := data[key].(bool)
	return value
}

func configInt(data map[string]interface{}, key string) int {
	switch value := data[key].(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	default:
		return 0
	}
}

func sanitizedConfigLocation(sourceType, location string) string {
	location = strings.TrimSpace(location)
	if sourceType == "hdhr" {
		return "[private-lan source]"
	}
	if !isRemoteProviderSource(location) {
		return "[local source]"
	}
	parsed, err := url.Parse(location)
	if err != nil || parsed.Host == "" {
		return "[remote source]"
	}
	host := parsed.Hostname()
	if strings.EqualFold(host, "localhost") {
		return "[restricted source]"
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsPrivate() {
			return "[private-lan source]"
		}
		if ip.IsLoopback() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() || ip.IsMulticast() {
			return "[restricted source]"
		}
	}
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.ForceQuery = false
	parsed.Fragment = ""
	return parsed.String()
}

func configSourceState(sourceType string, values map[string]interface{}) configSourceDTO {
	return configSourceDTO{
		ID:          configString(values, "id.provider"),
		Type:        sourceType,
		Name:        configString(values, "name"),
		Description: configString(values, "description"),
		Location:    sanitizedConfigLocation(sourceType, configString(values, "file.source")),
		Tuner:       configInt(values, "tuner"),
		Buffer:      configString(values, "buffer"),
	}
}

func buildConfigState() (configStateDTO, error) {
	hidden, err := snapshotConfigHiddenState()
	if err != nil {
		return configStateDTO{}, err
	}
	revision, err := configRevisionWithKey(configHMACKey, hidden)
	if err != nil {
		return configStateDTO{}, err
	}
	state := configStateDTO{
		ConfigRevision: revision,
		Onboarding:     configOnboardingDTO{Tuner: Settings.Tuner, EpgSource: Settings.EpgSource},
		Sources:        make([]configSourceDTO, 0),
		Filters:        make([]configFilterDTO, 0),
		Mappings:       make([]configMappingDTO, 0),
	}
	appendSources := func(sourceType string, sources map[string]interface{}) {
		ids := make([]string, 0, len(sources))
		for id := range sources {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			values, ok := sources[id].(map[string]interface{})
			if !ok {
				continue
			}
			copy := make(map[string]interface{}, len(values)+1)
			for key, value := range values {
				copy[key] = value
			}
			copy["id.provider"] = id
			state.Sources = append(state.Sources, configSourceState(sourceType, copy))
		}
	}
	appendSources("m3u", Settings.Files.M3U)
	appendSources("hdhr", Settings.Files.HDHR)
	appendSources("xmltv", Settings.Files.XMLTV)

	filterIDs := make([]int64, 0, len(Settings.Filter))
	for id := range Settings.Filter {
		filterIDs = append(filterIDs, id)
	}
	sort.Slice(filterIDs, func(i, j int) bool { return filterIDs[i] < filterIDs[j] })
	for _, id := range filterIDs {
		values, ok := Settings.Filter[id].(map[string]interface{})
		if !ok {
			continue
		}
		state.Filters = append(state.Filters, configFilterDTO{
			ID: strconv.FormatInt(id, 10), Type: configString(values, "type"), Name: configString(values, "name"),
			Description: configString(values, "description"), Filter: configString(values, "filter"),
			CaseSensitive: configBool(values, "caseSensitive"), LiveEvent: configBool(values, "liveEvent"),
			Include: configString(values, "include"), Exclude: configString(values, "exclude"),
			StartingNumber: configString(values, "startingNumber"), Category: configString(values, "x-category"), Active: configBool(values, "active"),
		})
	}

	mappingIDs := make([]string, 0, len(Data.XEPG.Channels))
	for id := range Data.XEPG.Channels {
		mappingIDs = append(mappingIDs, id)
	}
	sort.Strings(mappingIDs)
	for _, id := range mappingIDs {
		values, ok := Data.XEPG.Channels[id].(map[string]interface{})
		if !ok {
			continue
		}
		state.Mappings = append(state.Mappings, configMappingDTO{
			ID: id, ChannelID: configString(values, "x-channelID"), Active: configBool(values, "x-active"),
			Name: configString(values, "x-name"), Description: configString(values, "x-description"),
			UpdateChannelName: configBool(values, "x-update-channel-name"), UpdateChannelIcon: configBool(values, "x-update-channel-icon"),
			Category: configString(values, "x-category"), GroupTitle: configString(values, "x-group-title"),
			XMLTVFile: configString(values, "x-xmltv-file"), Mapping: configString(values, "x-mapping"), PPVExtra: configString(values, "x-ppv-extra"),
			BackupChannel1: configString(values, "x-backup-channel-1"), BackupChannel2: configString(values, "x-backup-channel-2"), BackupChannel3: configString(values, "x-backup-channel-3"),
		})
	}
	return state, nil
}

func decodeExactConfigObject(raw json.RawMessage, target interface{}, allowed ...string) error {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return errConfigInvalidRequest
	}
	if err := rejectDuplicateConfigFields(raw); err != nil {
		return errConfigInvalidRequest
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return errConfigInvalidRequest
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, field := range allowed {
		allowedSet[field] = struct{}{}
	}
	for field, value := range fields {
		if _, ok := allowedSet[field]; !ok {
			return errConfigInvalidRequest
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return errConfigInvalidRequest
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return errConfigInvalidRequest
	}
	return nil
}

func configStringValid(value string, limit int, required bool) bool {
	if required && value == "" {
		return false
	}
	if len(value) > limit {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func normalizeConfigString(value *string) {
	if value != nil {
		trimmed := strings.TrimSpace(*value)
		*value = trimmed
	}
}

func configSourceMap(sourceType string) map[string]interface{} {
	switch sourceType {
	case "m3u":
		return Settings.Files.M3U
	case "hdhr":
		return Settings.Files.HDHR
	case "xmltv":
		return Settings.Files.XMLTV
	default:
		return nil
	}
}

func configLiteralDestinationAllowed(host string) bool {
	host = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
	switch host {
	case "", "localhost", "metadata.google.internal", "metadata.azure.internal", "metadata.azure.com":
		return false
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() || ip.IsMulticast() {
		return false
	}
	for _, blocked := range []string{"100.100.100.200", "168.63.129.16", "fd00:ec2::254"} {
		if ip.Equal(net.ParseIP(blocked)) {
			return false
		}
	}
	return true
}

func configSourceLocationValid(sourceType, location string) bool {
	if !configStringValid(location, 2048, true) {
		return false
	}
	if sourceType == "hdhr" {
		parsed, err := url.Parse("http://" + location)
		return err == nil && parsed.Host != "" && parsed.User == nil && (parsed.Path == "" || parsed.Path == "/") && parsed.RawQuery == "" && parsed.Fragment == "" && configLiteralDestinationAllowed(parsed.Hostname())
	}
	parsed, err := url.Parse(location)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != "" && configLiteralDestinationAllowed(parsed.Hostname())
}

func configFilterByID(id string) (int64, map[string]interface{}, bool) {
	numericID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, nil, false
	}
	value, exists := Settings.Filter[numericID]
	if !exists {
		return 0, nil, false
	}
	filter, ok := value.(map[string]interface{})
	return numericID, filter, ok
}

func configMappingBackupReferenceValid(reference string) bool {
	if reference == "" || reference == "-" {
		return true
	}
	for _, channel := range Data.XEPG.Channels {
		values, ok := channel.(map[string]interface{})
		if ok && configString(values, "tvg-name") == reference {
			return true
		}
	}
	return false
}

func configDummyMappingValid(mapping string) bool {
	if dummy, ok := Data.XMLTV.Mapping["Threadfin Dummy"].(map[string]interface{}); ok {
		_, exists := dummy[mapping]
		return exists
	}
	switch mapping {
	case "PPV", "30_Minutes", "60_Minutes", "90_Minutes", "120_Minutes", "180_Minutes", "240_Minutes", "360_Minutes":
		return true
	default:
		return false
	}
}

func configMappingReferencesValid(existing map[string]interface{}, patch configMappingPatchAction) bool {
	xmltvFile := configString(existing, "x-xmltv-file")
	mapping := configString(existing, "x-mapping")
	backups := [3]string{
		configString(existing, "x-backup-channel-1"),
		configString(existing, "x-backup-channel-2"),
		configString(existing, "x-backup-channel-3"),
	}
	if patch.XMLTVFile != nil {
		xmltvFile = *patch.XMLTVFile
	}
	if patch.Mapping != nil {
		mapping = *patch.Mapping
	}
	for index, value := range []*string{patch.BackupChannel1, patch.BackupChannel2, patch.BackupChannel3} {
		if value != nil {
			backups[index] = *value
		}
	}
	for _, backup := range backups {
		if !configMappingBackupReferenceValid(backup) {
			return false
		}
	}
	if xmltvFile == "-" {
		return mapping == "-"
	}
	if xmltvFile == "Threadfin Dummy" {
		return configDummyMappingValid(mapping)
	}
	channels, ok := Data.XMLTV.Mapping[xmltvFile].(map[string]interface{})
	if !ok {
		return false
	}
	_, exists := channels[mapping]
	return exists
}

func normalizeConfigAction(raw json.RawMessage) (configNormalizedAction, error) {
	var envelope struct {
		Type                  string          `json:"type"`
		Onboarding            json.RawMessage `json:"onboarding,omitempty"`
		Source                json.RawMessage `json:"source,omitempty"`
		Filter                json.RawMessage `json:"filter,omitempty"`
		MappingPatch          json.RawMessage `json:"mapping_patch,omitempty"`
		MappingBulkChannelIDs json.RawMessage `json:"mapping_bulk_channel_ids,omitempty"`
	}
	if err := decodeExactConfigObject(raw, &envelope, "type", "onboarding", "source", "filter", "mapping_patch", "mapping_bulk_channel_ids"); err != nil {
		return configNormalizedAction{}, err
	}
	payloadCount := 0
	for _, payload := range []json.RawMessage{envelope.Onboarding, envelope.Source, envelope.Filter, envelope.MappingPatch, envelope.MappingBulkChannelIDs} {
		if len(payload) != 0 {
			payloadCount++
		}
	}
	if payloadCount != 1 || !configStringValid(envelope.Type, 64, true) {
		return configNormalizedAction{}, errConfigInvalidRequest
	}
	action := configNormalizedAction{Type: envelope.Type}
	switch envelope.Type {
	case "onboarding.update":
		if len(envelope.Onboarding) == 0 {
			return action, errConfigInvalidRequest
		}
		var onboarding configOnboardingAction
		if err := decodeExactConfigObject(envelope.Onboarding, &onboarding, "tuner", "epg_source"); err != nil {
			return action, err
		}
		normalizeConfigString(onboarding.EpgSource)
		if onboarding.Tuner == nil && onboarding.EpgSource == nil {
			return action, errConfigInvalidRequest
		}
		if onboarding.Tuner != nil && (*onboarding.Tuner < 1 || *onboarding.Tuner > 100) {
			return action, errConfigInvalidRequest
		}
		if onboarding.EpgSource != nil && *onboarding.EpgSource != "PMS" && *onboarding.EpgSource != "XEPG" {
			return action, errConfigInvalidRequest
		}
		action.Onboarding = &onboarding
	case "source.create", "source.update", "source.delete", "source.refresh":
		if len(envelope.Source) == 0 {
			return action, errConfigInvalidRequest
		}
		var source configSourceAction
		if err := decodeExactConfigObject(envelope.Source, &source, "source_type", "id", "name", "description", "location", "tuner", "buffer"); err != nil {
			return action, err
		}
		source.SourceType = strings.TrimSpace(source.SourceType)
		source.ID = strings.TrimSpace(source.ID)
		normalizeConfigString(source.Name)
		normalizeConfigString(source.Description)
		normalizeConfigString(source.Location)
		normalizeConfigString(source.Buffer)
		if configSourceMap(source.SourceType) == nil || !configStringValid(source.ID, 64, false) {
			return action, errConfigInvalidRequest
		}
		if source.Name != nil && !configStringValid(*source.Name, 128, false) || source.Description != nil && !configStringValid(*source.Description, 512, false) || source.Location != nil && !configSourceLocationValid(source.SourceType, *source.Location) {
			return action, errConfigInvalidRequest
		}
		if source.Tuner != nil && (*source.Tuner < 1 || *source.Tuner > 100 || source.SourceType == "xmltv") {
			return action, errConfigInvalidRequest
		}
		if source.Buffer != nil && (source.SourceType == "xmltv" || (*source.Buffer != "-" && *source.Buffer != "ffmpeg" && *source.Buffer != "vlc")) {
			return action, errConfigInvalidRequest
		}
		mutable := source.Name != nil || source.Description != nil || source.Location != nil || source.Tuner != nil || source.Buffer != nil
		if envelope.Type == "source.create" {
			if source.ID != "" || source.Name == nil || source.Location == nil || !configStringValid(*source.Name, 128, true) {
				return action, errConfigInvalidRequest
			}
		} else {
			if source.ID == "" {
				return action, errConfigInvalidRequest
			}
			if _, exists := configSourceMap(source.SourceType)[source.ID]; !exists {
				return action, errConfigInvalidRequest
			}
			if envelope.Type == "source.update" && !mutable || (envelope.Type == "source.delete" || envelope.Type == "source.refresh") && mutable {
				return action, errConfigInvalidRequest
			}
		}
		action.Source = &source
	case "filter.create", "filter.update", "filter.delete":
		if len(envelope.Filter) == 0 {
			return action, errConfigInvalidRequest
		}
		var filter configFilterAction
		if err := decodeExactConfigObject(envelope.Filter, &filter, "id", "type", "name", "description", "filter", "case_sensitive", "live_event", "include", "exclude", "starting_number", "category", "active"); err != nil {
			return action, err
		}
		filter.ID = strings.TrimSpace(filter.ID)
		filter.Type = strings.TrimSpace(filter.Type)
		for _, value := range []*string{filter.Name, filter.Description, filter.Filter, filter.Include, filter.Exclude, filter.StartingNumber, filter.Category} {
			normalizeConfigString(value)
		}
		if !configStringValid(filter.ID, 64, false) || filter.Name != nil && !configStringValid(*filter.Name, 128, false) || filter.Description != nil && !configStringValid(*filter.Description, 512, false) || filter.Filter != nil && !configStringValid(*filter.Filter, 2048, false) || filter.Include != nil && !configStringValid(*filter.Include, 2048, false) || filter.Exclude != nil && !configStringValid(*filter.Exclude, 2048, false) || filter.StartingNumber != nil && !configStringValid(*filter.StartingNumber, 256, false) || filter.Category != nil && !configStringValid(*filter.Category, 256, false) {
			return action, errConfigInvalidRequest
		}
		mutable := filter.Name != nil || filter.Description != nil || filter.Filter != nil || filter.CaseSensitive != nil || filter.LiveEvent != nil || filter.Include != nil || filter.Exclude != nil || filter.StartingNumber != nil || filter.Category != nil || filter.Active != nil
		if envelope.Type == "filter.create" {
			if filter.ID != "" || (filter.Type != "group-title" && filter.Type != "custom-filter") || filter.Name == nil || filter.Filter == nil || !configStringValid(*filter.Name, 128, true) || !configStringValid(*filter.Filter, 2048, true) {
				return action, errConfigInvalidRequest
			}
		} else {
			_, existing, exists := configFilterByID(filter.ID)
			if !exists {
				return action, errConfigInvalidRequest
			}
			existingType := configString(existing, "type")
			if filter.Type != "" && filter.Type != existingType {
				return action, errConfigInvalidRequest
			}
			filter.Type = existingType
			if envelope.Type == "filter.update" && !mutable || envelope.Type == "filter.delete" && mutable {
				return action, errConfigInvalidRequest
			}
		}
		action.Filter = &filter
	case "mapping.patch":
		if len(envelope.MappingPatch) == 0 {
			return action, errConfigInvalidRequest
		}
		var mapping configMappingPatchAction
		if err := decodeExactConfigObject(envelope.MappingPatch, &mapping, "id", "channel_id", "active", "name", "description", "update_channel_name", "update_channel_icon", "category", "group_title", "xmltv_file", "mapping", "ppv_extra", "backup_channel_1", "backup_channel_2", "backup_channel_3"); err != nil {
			return action, err
		}
		mapping.ID = strings.TrimSpace(mapping.ID)
		for _, value := range []*string{mapping.ChannelID, mapping.Name, mapping.Description, mapping.Category, mapping.GroupTitle, mapping.XMLTVFile, mapping.Mapping, mapping.PPVExtra, mapping.BackupChannel1, mapping.BackupChannel2, mapping.BackupChannel3} {
			normalizeConfigString(value)
		}
		if !configStringValid(mapping.ID, 64, true) {
			return action, errConfigInvalidRequest
		}
		existing, exists := Data.XEPG.Channels[mapping.ID].(map[string]interface{})
		if !exists {
			return action, errConfigInvalidRequest
		}
		if mapping.ChannelID != nil && !configStringValid(*mapping.ChannelID, 32, true) || mapping.Name != nil && !configStringValid(*mapping.Name, 128, false) || mapping.Description != nil && !configStringValid(*mapping.Description, 512, false) {
			return action, errConfigInvalidRequest
		}
		for _, value := range []*string{mapping.Category, mapping.GroupTitle, mapping.XMLTVFile, mapping.Mapping, mapping.PPVExtra, mapping.BackupChannel1, mapping.BackupChannel2, mapping.BackupChannel3} {
			if value != nil && !configStringValid(*value, 256, false) {
				return action, errConfigInvalidRequest
			}
		}
		referencesChanged := mapping.XMLTVFile != nil || mapping.Mapping != nil || mapping.BackupChannel1 != nil || mapping.BackupChannel2 != nil || mapping.BackupChannel3 != nil
		if referencesChanged && !configMappingReferencesValid(existing, mapping) {
			return action, errConfigInvalidRequest
		}
		if mapping.ChannelID == nil && mapping.Active == nil && mapping.Name == nil && mapping.Description == nil && mapping.UpdateChannelName == nil && mapping.UpdateChannelIcon == nil && mapping.Category == nil && mapping.GroupTitle == nil && mapping.XMLTVFile == nil && mapping.Mapping == nil && mapping.PPVExtra == nil && mapping.BackupChannel1 == nil && mapping.BackupChannel2 == nil && mapping.BackupChannel3 == nil {
			return action, errConfigInvalidRequest
		}
		action.MappingPatch = &mapping
	case "mapping.bulk_channel_ids":
		if len(envelope.MappingBulkChannelIDs) == 0 {
			return action, errConfigInvalidRequest
		}
		var bulkEnvelope struct {
			Items []json.RawMessage `json:"items"`
		}
		if err := decodeExactConfigObject(envelope.MappingBulkChannelIDs, &bulkEnvelope, "items"); err != nil || len(bulkEnvelope.Items) == 0 || len(bulkEnvelope.Items) > 480 {
			return action, errConfigInvalidRequest
		}
		bulk := configMappingBulkAction{Items: make([]configMappingBulkItem, len(bulkEnvelope.Items))}
		for index, rawItem := range bulkEnvelope.Items {
			if err := decodeExactConfigObject(rawItem, &bulk.Items[index], "id", "channel_id"); err != nil {
				return action, errConfigInvalidRequest
			}
		}
		seen := make(map[string]struct{}, len(bulk.Items))
		for index := range bulk.Items {
			bulk.Items[index].ID = strings.TrimSpace(bulk.Items[index].ID)
			bulk.Items[index].ChannelID = strings.TrimSpace(bulk.Items[index].ChannelID)
			if !configStringValid(bulk.Items[index].ID, 64, true) || !configStringValid(bulk.Items[index].ChannelID, 32, true) {
				return action, errConfigInvalidRequest
			}
			if _, exists := Data.XEPG.Channels[bulk.Items[index].ID]; !exists {
				return action, errConfigInvalidRequest
			}
			if _, duplicate := seen[bulk.Items[index].ID]; duplicate {
				return action, errConfigInvalidRequest
			}
			seen[bulk.Items[index].ID] = struct{}{}
		}
		action.MappingBulkChannelIDs = &bulk
	default:
		return action, errConfigInvalidRequest
	}
	return action, nil
}

func configNormalizedActionJSON(action configNormalizedAction) (json.RawMessage, error) {
	encoded, err := json.Marshal(action)
	return json.RawMessage(encoded), err
}

func configPlanActionJSON(action configNormalizedAction) (json.RawMessage, error) {
	redacted := action
	if action.Source != nil {
		source := *action.Source
		if source.Location != nil {
			location := sanitizedConfigLocation(source.SourceType, *source.Location)
			source.Location = &location
		}
		redacted.Source = &source
	}
	return configNormalizedActionJSON(redacted)
}

func configActionEffects(action configNormalizedAction) ([]string, []string) {
	switch action.Type {
	case "onboarding.update":
		return []string{"Persist the selected tuner capacity and guide mode."}, []string{"Changing guide mode rebuilds provider and output data."}
	case "source.create", "source.update", "source.refresh":
		return []string{"Fetch the selected source and persist validated provider state."}, []string{"This action performs a network fetch, can reach authorized private-LAN sources, and rebuilds downstream data."}
	case "source.delete":
		return []string{"Delete the selected source and its cached provider file."}, []string{"This action removes source data and rebuilds downstream data."}
	case "filter.create", "filter.update", "filter.delete":
		return []string{"Persist the selected filter change."}, []string{"Filter changes rebuild channel and guide output data."}
	case "mapping.patch", "mapping.bulk_channel_ids":
		return []string{"Persist the complete mapping with only the selected fields changed."}, []string{"Mapping changes rebuild outputs immediately or queue an output rebuild while a scan is active."}
	default:
		return nil, nil
	}
}

type configDomainRequest struct {
	Context  context.Context
	Kind     string
	FileType string
	Request  RequestStruct
	Action   configNormalizedAction
}

var configExecuteDomainRequest = executeConfigDomainRequest
var configProviderOptions = configProviderFetchOptions

func configSetString(values map[string]interface{}, key string, value *string) {
	if value != nil {
		values[key] = *value
	}
}

func configSetBool(values map[string]interface{}, key string, value *bool) {
	if value != nil {
		values[key] = *value
	}
}

func buildConfigDomainRequest(action configNormalizedAction) (configDomainRequest, error) {
	domain := configDomainRequest{Action: action}
	switch action.Type {
	case "onboarding.update":
		domain.Kind = "onboarding"
		domain.Request.Settings.Tuner = action.Onboarding.Tuner
		domain.Request.Settings.EpgSource = action.Onboarding.EpgSource
	case "source.create", "source.update", "source.delete", "source.refresh":
		domain.FileType = action.Source.SourceType
		values := make(map[string]interface{})
		id := action.Source.ID
		if action.Type == "source.create" {
			id = "-"
			values["type"] = action.Source.SourceType
		}
		configSetString(values, "name", action.Source.Name)
		configSetString(values, "description", action.Source.Description)
		configSetString(values, "file.source", action.Source.Location)
		if action.Source.Tuner != nil {
			values["tuner"] = *action.Source.Tuner
		}
		configSetString(values, "buffer", action.Source.Buffer)
		if action.Type == "source.delete" {
			values["delete"] = true
		}
		if action.Type == "source.refresh" {
			domain.Kind = "source.refresh"
		} else {
			domain.Kind = "source.save"
		}
		switch action.Source.SourceType {
		case "m3u":
			domain.Request.Files.M3U = map[string]interface{}{id: values}
		case "hdhr":
			domain.Request.Files.HDHR = map[string]interface{}{id: values}
		case "xmltv":
			domain.Request.Files.XMLTV = map[string]interface{}{id: values}
		default:
			return configDomainRequest{}, errConfigInvalidRequest
		}
	case "filter.create", "filter.update", "filter.delete":
		domain.Kind = "filter"
		id := int64(-1)
		if action.Type != "filter.create" {
			var ok bool
			id, _, ok = configFilterByID(action.Filter.ID)
			if !ok {
				return configDomainRequest{}, errConfigInvalidRequest
			}
		}
		values := make(map[string]interface{})
		if action.Type == "filter.delete" {
			values["delete"] = true
		} else {
			if action.Filter.Type != "" {
				values["type"] = action.Filter.Type
			}
			configSetString(values, "name", action.Filter.Name)
			configSetString(values, "description", action.Filter.Description)
			configSetString(values, "filter", action.Filter.Filter)
			configSetBool(values, "caseSensitive", action.Filter.CaseSensitive)
			configSetBool(values, "liveEvent", action.Filter.LiveEvent)
			configSetString(values, "include", action.Filter.Include)
			configSetString(values, "exclude", action.Filter.Exclude)
			configSetString(values, "startingNumber", action.Filter.StartingNumber)
			configSetString(values, "x-category", action.Filter.Category)
			configSetBool(values, "active", action.Filter.Active)
		}
		domain.Request.Filter = map[int64]interface{}{id: values}
	case "mapping.patch", "mapping.bulk_channel_ids":
		domain.Kind = "mapping"
		encoded, err := json.Marshal(Data.XEPG.Channels)
		if err != nil {
			return configDomainRequest{}, err
		}
		mapping := make(map[string]interface{})
		if err := json.Unmarshal(encoded, &mapping); err != nil {
			return configDomainRequest{}, err
		}
		if action.Type == "mapping.patch" {
			values, ok := mapping[action.MappingPatch.ID].(map[string]interface{})
			if !ok {
				return configDomainRequest{}, errConfigInvalidRequest
			}
			configSetString(values, "x-channelID", action.MappingPatch.ChannelID)
			configSetBool(values, "x-active", action.MappingPatch.Active)
			configSetString(values, "x-name", action.MappingPatch.Name)
			configSetString(values, "x-description", action.MappingPatch.Description)
			configSetBool(values, "x-update-channel-name", action.MappingPatch.UpdateChannelName)
			configSetBool(values, "x-update-channel-icon", action.MappingPatch.UpdateChannelIcon)
			configSetString(values, "x-category", action.MappingPatch.Category)
			configSetString(values, "x-group-title", action.MappingPatch.GroupTitle)
			configSetString(values, "x-xmltv-file", action.MappingPatch.XMLTVFile)
			configSetString(values, "x-mapping", action.MappingPatch.Mapping)
			configSetString(values, "x-ppv-extra", action.MappingPatch.PPVExtra)
			configSetString(values, "x-backup-channel-1", action.MappingPatch.BackupChannel1)
			configSetString(values, "x-backup-channel-2", action.MappingPatch.BackupChannel2)
			configSetString(values, "x-backup-channel-3", action.MappingPatch.BackupChannel3)
		} else {
			for _, item := range action.MappingBulkChannelIDs.Items {
				values, ok := mapping[item.ID].(map[string]interface{})
				if !ok {
					return configDomainRequest{}, errConfigInvalidRequest
				}
				values["x-channelID"] = item.ChannelID
			}
		}
		domain.Request.EpgMapping = mapping
	default:
		return configDomainRequest{}, errConfigInvalidRequest
	}
	return domain, nil
}

func configDomainPostprocessing(domain configDomainRequest, mappingResult MappingSaveResult) string {
	switch domain.Kind {
	case "onboarding":
		if domain.Action.Onboarding.EpgSource != nil {
			return "rebuilt"
		}
		return "not_required"
	case "source.save":
		if domain.Action.Type == "source.update" {
			return "not_required"
		}
		return "rebuilt"
	case "filter":
		return "rebuilt"
	case "source.refresh":
		return "queued"
	case "mapping":
		if mappingResult == MappingOutputRebuildRequested {
			return "queued"
		}
		if mappingResult == MappingOutputsRebuilt {
			return "rebuilt"
		}
	}
	return "failed"
}

func configDomainRequestPersisted(domain configDomainRequest) bool {
	switch domain.Kind {
	case "onboarding":
		return (domain.Action.Onboarding.Tuner == nil || Settings.Tuner == *domain.Action.Onboarding.Tuner) && (domain.Action.Onboarding.EpgSource == nil || Settings.EpgSource == *domain.Action.Onboarding.EpgSource)
	case "mapping":
		return reflect.DeepEqual(Data.XEPG.Channels, domain.Request.EpgMapping)
	case "source.save":
		if domain.Action.Type == "source.delete" {
			_, exists := configSourceMap(domain.FileType)[domain.Action.Source.ID]
			return !exists
		}
		if domain.Action.Type == "source.update" {
			stored, ok := configSourceMap(domain.FileType)[domain.Action.Source.ID].(map[string]interface{})
			if !ok {
				return false
			}
			return (domain.Action.Source.Name == nil || configString(stored, "name") == *domain.Action.Source.Name) && (domain.Action.Source.Description == nil || configString(stored, "description") == *domain.Action.Source.Description) && (domain.Action.Source.Location == nil || configString(stored, "file.source") == *domain.Action.Source.Location)
		}
		for _, value := range configSourceMap(domain.FileType) {
			stored, ok := value.(map[string]interface{})
			if ok && configString(stored, "name") == *domain.Action.Source.Name && configString(stored, "file.source") == *domain.Action.Source.Location {
				return true
			}
		}
	case "source.refresh":
		return false
	case "filter":
		if domain.Action.Type == "filter.delete" {
			_, _, exists := configFilterByID(domain.Action.Filter.ID)
			return !exists
		}
		if domain.Action.Type == "filter.update" {
			_, _, exists := configFilterByID(domain.Action.Filter.ID)
			return exists
		}
		return len(Settings.Filter) > 0
	}
	return false
}

func executeConfigDomainRequest(domain configDomainRequest) (configApplyResult, error) {
	var err error
	var mappingResult MappingSaveResult
	switch domain.Kind {
	case "onboarding":
		_, err = updateServerSettings(domain.Request)
	case "source.save":
		err = saveConfigFiles(domain.Context, domain.Request, domain.FileType)
	case "source.refresh":
		err = updateConfigFile(domain.Context, domain.Request, domain.FileType)
	case "filter":
		_, err = saveFilter(domain.Request)
	case "mapping":
		mappingResult, err = saveXEpgMapping(domain.Request)
	default:
		err = errConfigInvalidRequest
	}
	persisted := err == nil || configDomainRequestPersisted(domain)
	if domain.Kind == "source.refresh" {
		persisted = err == nil
	}
	result := configApplyResult{Persisted: persisted, Postprocessing: configDomainPostprocessing(domain, mappingResult)}
	if err != nil {
		result.Postprocessing = "failed"
	}
	return result, err
}

func saveConfigFiles(ctx context.Context, request RequestStruct, fileType string) error {
	return saveFilesWithOptions(request, fileType, configProviderOptions(ctx))
}

func updateConfigFile(ctx context.Context, request RequestStruct, fileType string) error {
	return updateFileWithOptions(request, fileType, configProviderOptions(ctx))
}

func applyConfigActionDomain(ctx context.Context, action configNormalizedAction) (configApplyResult, error) {
	domain, err := buildConfigDomainRequest(action)
	if err != nil {
		return configApplyResult{}, err
	}
	domain.Context = ctx
	return configExecuteDomainRequest(domain)
}

func isConfigCommandBody(body []byte) bool {
	var envelope struct {
		Cmd string `json:"cmd"`
	}
	if json.Unmarshal(body, &envelope) != nil {
		return false
	}
	return strings.HasPrefix(envelope.Cmd, "config.")
}

func rejectDuplicateConfigFields(body []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	var readValue func() error
	readValue = func() error {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		delimiter, ok := token.(json.Delim)
		if !ok {
			return nil
		}
		switch delimiter {
		case '{':
			seen := make(map[string]struct{})
			for decoder.More() {
				keyToken, err := decoder.Token()
				if err != nil {
					return err
				}
				key, ok := keyToken.(string)
				if !ok {
					return errConfigInvalidRequest
				}
				if _, exists := seen[key]; exists {
					return errConfigInvalidRequest
				}
				seen[key] = struct{}{}
				if err := readValue(); err != nil {
					return err
				}
			}
			_, err = decoder.Token()
			return err
		case '[':
			for decoder.More() {
				if err := readValue(); err != nil {
					return err
				}
			}
			_, err = decoder.Token()
			return err
		default:
			return errConfigInvalidRequest
		}
	}
	if err := readValue(); err != nil {
		return err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return errConfigInvalidRequest
	}
	return nil
}

func decodeConfigAPIRequest(body []byte) (configAPIRequest, error) {
	var request configAPIRequest
	if len(body) == 0 {
		return request, errConfigInvalidRequest
	}
	if len(body) > configAPIRequestLimit {
		return request, errConfigRequestTooLarge
	}
	if err := rejectDuplicateConfigFields(body); err != nil {
		return request, errConfigInvalidRequest
	}
	var exactFields map[string]json.RawMessage
	if err := json.Unmarshal(body, &exactFields); err != nil || exactFields == nil {
		return request, errConfigInvalidRequest
	}
	commandRaw, exists := exactFields["cmd"]
	if !exists || bytes.Equal(bytes.TrimSpace(commandRaw), []byte("null")) {
		return request, errConfigInvalidRequest
	}
	var command string
	if err := json.Unmarshal(commandRaw, &command); err != nil || command == "" {
		return request, errConfigInvalidRequest
	}
	requiredFieldsByCommand := map[string][]string{
		"config.describe": {"cmd", "token"},
		"config.get":      {"cmd", "token"},
		"config.plan":     {"cmd", "token", "action"},
		"config.apply":    {"cmd", "token", "action", "config_revision", "plan_digest"},
	}
	requiredFields, knownCommand := requiredFieldsByCommand[command]
	if !knownCommand || len(exactFields) != len(requiredFields) {
		return request, errConfigInvalidRequest
	}
	for _, field := range requiredFields {
		value, present := exactFields[field]
		if !present || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return request, errConfigInvalidRequest
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return request, errConfigInvalidRequest
	}
	if request.Cmd == "" {
		return request, errConfigInvalidRequest
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return request, errConfigInvalidRequest
	}
	return request, nil
}

func writeConfigJSON(w http.ResponseWriter, status int, response configAPIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response)
}

func writeConfigForbidden(w http.ResponseWriter) {
	writeConfigJSON(w, http.StatusForbidden, configAPIResponse{
		Status: false,
		Error:  &configAPIError{Code: "forbidden", Message: "configuration access denied"},
	})
}

func authorizeConfigRequest(request configAPIRequest) (principal, rotatedToken string, ok bool) {
	if !Settings.API || !Settings.AuthenticationAPI || System.ConfigurationWizard || request.Token == "" || len(request.Token) > 128 {
		return "", "", false
	}
	rotatedToken, principal, err := authentication.AuthorizeTokenPermissions(request.Token, "authentication.api", "authentication.config")
	if err != nil {
		return "", "", false
	}
	return principal, rotatedToken, true
}

func describeConfigAPI() configDescription {
	integer := func(minimum, maximum int) configFieldSchema {
		return configFieldSchema{Type: "integer", Minimum: &minimum, Maximum: &maximum}
	}
	stringField := func(maximum int, enum ...string) configFieldSchema {
		return configFieldSchema{Type: "string", MaxLength: maximum, Enum: enum}
	}
	boolean := configFieldSchema{Type: "boolean"}
	object := func(required, optional []string, properties map[string]configFieldSchema) configObjectSchema {
		return configObjectSchema{Type: "object", Required: required, Optional: optional, Properties: properties, AdditionalProperties: false}
	}
	action := func(payload string, required, optional []string, fields map[string]configFieldSchema, minProperties int) configActionSchema {
		return configActionSchema{PayloadKey: payload, Required: required, Optional: optional, Fields: fields, AdditionalProperties: false, MinProperties: minProperties}
	}
	selectFields := func(fields map[string]configFieldSchema, names ...string) map[string]configFieldSchema {
		selected := make(map[string]configFieldSchema, len(names))
		for _, name := range names {
			selected[name] = fields[name]
		}
		return selected
	}
	ref := func(name string) configFieldSchema { return configFieldSchema{Ref: name} }
	arrayRef := func(name string, max int) configFieldSchema {
		return configFieldSchema{Type: "array", ItemsRef: name, MaxItems: max}
	}

	sourceFields := map[string]configFieldSchema{
		"source_type": stringField(16, "m3u", "hdhr", "xmltv"), "id": stringField(64), "name": stringField(128),
		"description": stringField(512), "location": stringField(2048), "tuner": integer(1, 100), "buffer": stringField(16, "-", "ffmpeg", "vlc"),
	}
	locationField := sourceFields["location"]
	locationField.Description = "Sensitive IPTV URL input. URL userinfo, query, and fragment are accepted, digest-bound, and never returned or logged."
	locationField.WriteOnly = true
	locationField.Sensitive = true
	sourceFields["location"] = locationField
	filterFields := map[string]configFieldSchema{
		"id": stringField(64), "type": stringField(32, "group-title", "custom-filter"), "name": stringField(128), "description": stringField(512),
		"filter": stringField(2048), "case_sensitive": boolean, "live_event": boolean, "include": stringField(2048), "exclude": stringField(2048),
		"starting_number": stringField(256), "category": stringField(256), "active": boolean,
	}
	mappingFields := map[string]configFieldSchema{
		"id": stringField(64), "channel_id": stringField(32), "active": boolean, "name": stringField(128), "description": stringField(512),
		"update_channel_name": boolean, "update_channel_icon": boolean, "category": stringField(256), "group_title": stringField(256),
		"xmltv_file": stringField(256), "mapping": stringField(256), "ppv_extra": stringField(256), "backup_channel_1": stringField(256),
		"backup_channel_2": stringField(256), "backup_channel_3": stringField(256),
	}
	bulkItem := object([]string{"id", "channel_id"}, []string{}, map[string]configFieldSchema{"id": stringField(64), "channel_id": stringField(32)})
	actionSchemas := map[string]configActionSchema{
		"onboarding.update":        action("onboarding", []string{}, []string{"tuner", "epg_source"}, map[string]configFieldSchema{"tuner": integer(1, 100), "epg_source": stringField(8, "PMS", "XEPG")}, 1),
		"source.create":            action("source", []string{"source_type", "name", "location"}, []string{"description", "tuner", "buffer"}, selectFields(sourceFields, "source_type", "name", "location", "description", "tuner", "buffer"), 0),
		"source.update":            action("source", []string{"source_type", "id"}, []string{"name", "description", "location", "tuner", "buffer"}, sourceFields, 3),
		"source.delete":            action("source", []string{"source_type", "id"}, []string{}, map[string]configFieldSchema{"source_type": sourceFields["source_type"], "id": sourceFields["id"]}, 0),
		"source.refresh":           action("source", []string{"source_type", "id"}, []string{}, map[string]configFieldSchema{"source_type": sourceFields["source_type"], "id": sourceFields["id"]}, 0),
		"filter.create":            action("filter", []string{"type", "name", "filter"}, []string{"description", "case_sensitive", "live_event", "include", "exclude", "starting_number", "category", "active"}, selectFields(filterFields, "type", "name", "filter", "description", "case_sensitive", "live_event", "include", "exclude", "starting_number", "category", "active"), 0),
		"filter.update":            action("filter", []string{"id"}, []string{"type", "name", "description", "filter", "case_sensitive", "live_event", "include", "exclude", "starting_number", "category", "active"}, filterFields, 2),
		"filter.delete":            action("filter", []string{"id"}, []string{}, map[string]configFieldSchema{"id": filterFields["id"]}, 0),
		"mapping.patch":            action("mapping_patch", []string{"id"}, []string{"channel_id", "active", "name", "description", "update_channel_name", "update_channel_icon", "category", "group_title", "xmltv_file", "mapping", "ppv_extra", "backup_channel_1", "backup_channel_2", "backup_channel_3"}, mappingFields, 2),
		"mapping.bulk_channel_ids": action("mapping_bulk_channel_ids", []string{"items"}, []string{}, map[string]configFieldSchema{"items": {Type: "array", MinItems: 1, MaxItems: 480, Items: &bulkItem}}, 0),
	}
	for actionName, atLeastOne := range map[string][]string{
		"onboarding.update": {"tuner", "epg_source"},
		"source.update":     {"name", "description", "location", "tuner", "buffer"},
		"filter.update":     {"name", "description", "filter", "case_sensitive", "live_event", "include", "exclude", "starting_number", "category", "active"},
		"mapping.patch":     {"channel_id", "active", "name", "description", "update_channel_name", "update_channel_icon", "category", "group_title", "xmltv_file", "mapping", "ppv_extra", "backup_channel_1", "backup_channel_2", "backup_channel_3"},
	} {
		schema := actionSchemas[actionName]
		schema.AtLeastOne = atLeastOne
		actionSchemas[actionName] = schema
	}
	for _, actionName := range []string{"source.create", "source.update"} {
		schema := actionSchemas[actionName]
		schema.Constraints = append(schema.Constraints, "location must be an HTTP(S) URL or an HDHR host; local paths and separate credential, proxy, and header fields are forbidden; URL userinfo, query, and fragment remain accepted write-only input")
		actionSchemas[actionName] = schema
	}
	mappingSchema := actionSchemas["mapping.patch"]
	mappingSchema.Constraints = append(mappingSchema.Constraints, "xmltv_file and mapping must be a current guide-file/channel pair or a valid Threadfin Dummy pair; backup channels must name current channels")
	actionSchemas["mapping.patch"] = mappingSchema
	requestProperties := map[string]configFieldSchema{
		"cmd": stringField(32), "token": stringField(128), "action": ref("action_envelope"), "config_revision": stringField(67), "plan_digest": stringField(67),
	}
	revisionField := requestProperties["config_revision"]
	revisionField.MinLength, revisionField.Pattern = 67, `^r1:[0-9a-f]{64}$`
	requestProperties["config_revision"] = revisionField
	digestField := requestProperties["plan_digest"]
	digestField.MinLength, digestField.Pattern = 67, `^p1:[0-9a-f]{64}$`
	requestProperties["plan_digest"] = digestField
	requests := map[string]configObjectSchema{
		"config.describe": object([]string{"cmd", "token"}, []string{}, selectFields(requestProperties, "cmd", "token")),
		"config.get":      object([]string{"cmd", "token"}, []string{}, selectFields(requestProperties, "cmd", "token")),
		"config.plan":     object([]string{"cmd", "token", "action"}, []string{}, selectFields(requestProperties, "cmd", "token", "action")),
		"config.apply":    object([]string{"cmd", "token", "action", "config_revision", "plan_digest"}, []string{}, requestProperties),
	}
	for command, schema := range requests {
		commandField := schema.Properties["cmd"]
		commandField.Enum = []string{command}
		schema.Properties["cmd"] = commandField
		requests[command] = schema
	}
	responseModels := map[string]configObjectSchema{
		"error":              object([]string{"code", "message"}, []string{}, map[string]configFieldSchema{"code": stringField(64), "message": stringField(256)}),
		"onboarding":         object([]string{"tuner", "epg_source"}, []string{}, map[string]configFieldSchema{"tuner": {Type: "integer"}, "epg_source": {Type: "string"}}),
		"source":             object([]string{"id", "type", "name", "location"}, []string{"description", "tuner", "buffer"}, map[string]configFieldSchema{"id": {Type: "string"}, "type": {Type: "string"}, "name": {Type: "string"}, "description": {Type: "string"}, "location": {Type: "string"}, "tuner": {Type: "integer"}, "buffer": {Type: "string"}}),
		"filter":             object([]string{"id", "type", "name", "filter", "case_sensitive", "live_event", "active"}, []string{"description", "include", "exclude", "starting_number", "category"}, map[string]configFieldSchema{"id": {Type: "string"}, "type": {Type: "string"}, "name": {Type: "string"}, "description": {Type: "string"}, "filter": {Type: "string"}, "case_sensitive": boolean, "live_event": boolean, "include": {Type: "string"}, "exclude": {Type: "string"}, "starting_number": {Type: "string"}, "category": {Type: "string"}, "active": boolean}),
		"mapping":            object([]string{"id", "active", "update_channel_name", "update_channel_icon"}, []string{"channel_id", "name", "description", "category", "group_title", "xmltv_file", "mapping", "ppv_extra", "backup_channel_1", "backup_channel_2", "backup_channel_3"}, mappingFields),
		"state":              object([]string{"config_revision", "onboarding", "sources", "filters", "mappings"}, []string{}, map[string]configFieldSchema{"config_revision": {Type: "string"}, "onboarding": ref("onboarding"), "sources": arrayRef("source", 0), "filters": arrayRef("filter", 0), "mappings": arrayRef("mapping", 0)}),
		"plan":               object([]string{"action", "consequences", "warnings", "config_revision", "plan_digest"}, []string{}, map[string]configFieldSchema{"action": {Ref: "action_envelope", Description: "Redacted display only; do not replay. Apply must resend the caller-retained original action.", DisplayOnly: true}, "consequences": {Type: "array", ItemsRef: "string"}, "warnings": {Type: "array", ItemsRef: "string"}, "config_revision": {Type: "string"}, "plan_digest": {Type: "string"}}),
		"apply":              object([]string{"persisted", "postprocessing", "config_revision"}, []string{}, map[string]configFieldSchema{"persisted": boolean, "postprocessing": stringField(16, "not_required", "queued", "rebuilt", "failed"), "config_revision": {Type: "string"}}),
		"config_description": object([]string{"schema_version", "workflow", "authority", "token_policy", "sensitive_input_policy", "authentication", "requests", "action_envelope", "action_schemas", "responses", "response_models", "actions", "bounds", "exclusions"}, []string{}, map[string]configFieldSchema{"schema_version": {Type: "string"}, "workflow": {Type: "array", ItemsRef: "string"}, "authority": {Type: "string"}, "token_policy": {Type: "string"}, "sensitive_input_policy": {Type: "object"}, "authentication": {Type: "object"}, "requests": {Type: "object"}, "action_envelope": ref("action_envelope"), "action_schemas": {Type: "object"}, "responses": {Type: "object"}, "response_models": {Type: "object"}, "actions": {Type: "array", ItemsRef: "string"}, "bounds": {Type: "object"}, "exclusions": {Type: "array", ItemsRef: "string"}}),
	}
	responses := map[string]configObjectSchema{
		"config.describe": object([]string{"status", "token", "description"}, []string{}, map[string]configFieldSchema{"status": boolean, "token": {Type: "string"}, "description": ref("config_description")}),
		"config.get":      object([]string{"status", "token", "state"}, []string{}, map[string]configFieldSchema{"status": boolean, "token": {Type: "string"}, "state": ref("state")}),
		"config.plan":     object([]string{"status", "token", "plan"}, []string{}, map[string]configFieldSchema{"status": boolean, "token": {Type: "string"}, "plan": ref("plan")}),
		"config.apply":    object([]string{"status", "token", "apply"}, []string{}, map[string]configFieldSchema{"status": boolean, "token": {Type: "string"}, "apply": ref("apply")}),
		"error":           object([]string{"status", "error"}, []string{"token", "apply"}, map[string]configFieldSchema{"status": boolean, "token": {Type: "string"}, "apply": ref("apply"), "error": ref("error")}),
	}
	actionEnvelope := object([]string{"type"}, []string{"onboarding", "source", "filter", "mapping_patch", "mapping_bulk_channel_ids"}, map[string]configFieldSchema{
		"type": stringField(64, "onboarding.update", "source.create", "source.update", "source.delete", "source.refresh", "filter.create", "filter.update", "filter.delete", "mapping.patch", "mapping.bulk_channel_ids"), "onboarding": {Type: "object"}, "source": {Type: "object"}, "filter": {Type: "object"}, "mapping_patch": {Type: "object"}, "mapping_bulk_channel_ids": {Type: "object"},
	})
	actionEnvelope.MinProperties = 2
	actionEnvelope.ExactlyOne = []string{"onboarding", "source", "filter", "mapping_patch", "mapping_bulk_channel_ids"}
	return configDescription{
		SchemaVersion: configAPISchemaVersion,
		Workflow:      []string{"config.describe", "config.get", "config.plan", "config.apply"},
		Authority:     "CONFIG grants authority to fetch configured HTTP sources, including private-LAN sources; restricted destinations remain blocked.",
		TokenPolicy:   "The authentication token rotates on every successful call; use only the token returned by the immediately preceding response.",
		SensitiveInputPolicy: configSensitiveInputPolicy{
			Fields:                      []string{"action.source.location"},
			AcceptedURLComponents:       []string{"userinfo", "query", "fragment"},
			Returned:                    false,
			Logged:                      false,
			BoundTo:                     []string{"plan_digest", "config_revision"},
			PlanActionReplayable:        false,
			ApplyRequiresOriginalAction: true,
		},
		Authentication: configAuthenticationSchema{RequiredSettings: []string{"api", "authentication.api"}, RequiredPermissions: []string{"authentication.api", "authentication.config"}, SetupComplete: true, FailClosed: true},
		Requests:       requests,
		ActionEnvelope: actionEnvelope,
		ActionSchemas:  actionSchemas,
		Responses:      responses,
		ResponseModels: responseModels,
		Actions: []string{
			"onboarding.update",
			"source.create", "source.update", "source.delete", "source.refresh",
			"filter.create", "filter.update", "filter.delete",
			"mapping.patch", "mapping.bulk_channel_ids",
		},
		Bounds: map[string]int{
			"request_bytes":    configAPIRequestLimit,
			"id_bytes":         64,
			"name_bytes":       128,
			"description":      512,
			"expression":       2048,
			"string_bytes":     256,
			"token_bytes":      128,
			"mapping_targets":  480,
			"channel_id_bytes": 32,
		},
		Exclusions: []string{
			"users and authentication",
			"local paths, binaries, and runtime options",
			"bind, port, TLS, and network settings",
			"backup, restore, application updates, and uploads",
			"logos, separate credential fields, proxy or header controls, arbitrary settings, and raw maps",
		},
	}
}

func handleConfigAPI(w http.ResponseWriter, r *http.Request, body []byte) {
	if r.URL.Path != "/api/" {
		writeConfigJSON(w, http.StatusNotFound, configAPIResponse{Status: false, Error: &configAPIError{Code: "not_found", Message: "not found"}})
		return
	}
	if r.Method != http.MethodPost {
		writeConfigJSON(w, http.StatusMethodNotAllowed, configAPIResponse{Status: false, Error: &configAPIError{Code: "method_not_allowed", Message: "method not allowed"}})
		return
	}
	request, err := decodeConfigAPIRequest(body)
	if err != nil {
		status := http.StatusBadRequest
		code := "invalid_request"
		if errors.Is(err, errConfigRequestTooLarge) {
			status = http.StatusRequestEntityTooLarge
			code = "request_too_large"
		}
		writeConfigJSON(w, status, configAPIResponse{Status: false, Error: &configAPIError{Code: code, Message: "configuration request rejected"}})
		return
	}
	configMutationMutex.Lock()
	defer configMutationMutex.Unlock()
	principal, rotatedToken, ok := authorizeConfigRequest(request)
	if !ok {
		writeConfigForbidden(w)
		return
	}
	response := configAPIResponse{Status: true, Token: rotatedToken}
	switch request.Cmd {
	case "config.describe":
		if len(request.Action) != 0 || request.ConfigRevision != "" || request.PlanDigest != "" {
			writeConfigJSON(w, http.StatusBadRequest, configAPIResponse{Status: false, Token: rotatedToken, Error: &configAPIError{Code: "invalid_request", Message: "configuration request rejected"}})
			return
		}
		description := describeConfigAPI()
		response.Description = &description
	case "config.get":
		if len(request.Action) != 0 || request.ConfigRevision != "" || request.PlanDigest != "" {
			writeConfigJSON(w, http.StatusBadRequest, configAPIResponse{Status: false, Token: rotatedToken, Error: &configAPIError{Code: "invalid_request", Message: "configuration request rejected"}})
			return
		}
		state, stateErr := buildConfigState()
		if stateErr != nil {
			writeConfigJSON(w, http.StatusInternalServerError, configAPIResponse{Status: false, Token: rotatedToken, Error: &configAPIError{Code: "internal_error", Message: "configuration state unavailable"}})
			return
		}
		response.State = &state
	case "config.plan", "config.apply":
		if len(request.Action) == 0 || (request.Cmd == "config.plan" && (request.ConfigRevision != "" || request.PlanDigest != "")) || (request.Cmd == "config.apply" && (request.ConfigRevision == "" || request.PlanDigest == "")) {
			writeConfigJSON(w, http.StatusBadRequest, configAPIResponse{Status: false, Token: rotatedToken, Error: &configAPIError{Code: "invalid_request", Message: "configuration request rejected"}})
			return
		}
		normalized, normalizeErr := normalizeConfigAction(request.Action)
		if normalizeErr != nil {
			writeConfigJSON(w, http.StatusBadRequest, configAPIResponse{Status: false, Token: rotatedToken, Error: &configAPIError{Code: "invalid_action", Message: "configuration action rejected"}})
			return
		}
		normalizedJSON, marshalErr := configNormalizedActionJSON(normalized)
		if marshalErr != nil {
			writeConfigJSON(w, http.StatusInternalServerError, configAPIResponse{Status: false, Token: rotatedToken, Error: &configAPIError{Code: "internal_error", Message: "configuration state unavailable"}})
			return
		}
		state, stateErr := buildConfigState()
		if stateErr != nil {
			writeConfigJSON(w, http.StatusInternalServerError, configAPIResponse{Status: false, Token: rotatedToken, Error: &configAPIError{Code: "internal_error", Message: "configuration state unavailable"}})
			return
		}
		consequences, warnings := configActionEffects(normalized)
		digest, digestErr := configPlanDigestWithKey(configHMACKey, principal, state.ConfigRevision, normalizedJSON, consequences, warnings)
		if digestErr != nil {
			writeConfigJSON(w, http.StatusInternalServerError, configAPIResponse{Status: false, Token: rotatedToken, Error: &configAPIError{Code: "internal_error", Message: "configuration state unavailable"}})
			return
		}
		if request.Cmd == "config.plan" {
			planAction, planActionErr := configPlanActionJSON(normalized)
			if planActionErr != nil {
				writeConfigJSON(w, http.StatusInternalServerError, configAPIResponse{Status: false, Token: rotatedToken, Error: &configAPIError{Code: "internal_error", Message: "configuration state unavailable"}})
				return
			}
			response.Plan = &configPlanDTO{Action: planAction, Consequences: consequences, Warnings: warnings, ConfigRevision: state.ConfigRevision, PlanDigest: digest}
			break
		}
		if !configPlanCredentialsMatch(state.ConfigRevision, request.ConfigRevision, digest, request.PlanDigest) {
			writeConfigJSON(w, http.StatusConflict, configAPIResponse{Status: false, Token: rotatedToken, Error: &configAPIError{Code: "plan_conflict", Message: "configuration plan is stale or invalid"}})
			return
		}
		result, applyErr := configDomainApply(r.Context(), normalized)
		after, revisionErr := buildConfigState()
		if revisionErr != nil {
			writeConfigJSON(w, http.StatusInternalServerError, configAPIResponse{Status: false, Token: rotatedToken, Error: &configAPIError{Code: "internal_error", Message: "configuration state unavailable"}})
			return
		}
		apply := &configApplyDTO{Persisted: result.Persisted, Postprocessing: result.Postprocessing, ConfigRevision: after.ConfigRevision}
		if apply.Postprocessing == "" {
			apply.Postprocessing = "failed"
		}
		if applyErr != nil {
			writeConfigJSON(w, http.StatusInternalServerError, configAPIResponse{Status: false, Token: rotatedToken, Apply: apply, Error: &configAPIError{Code: "apply_failed", Message: "configuration action failed"}})
			return
		}
		response.Apply = apply
	default:
		writeConfigJSON(w, http.StatusBadRequest, configAPIResponse{Status: false, Token: rotatedToken, Error: &configAPIError{Code: "invalid_request", Message: "configuration request rejected"}})
		return
	}
	writeConfigJSON(w, http.StatusOK, response)
}
