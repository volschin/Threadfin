package src

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	m3u "threadfin/src/internal/m3u-parser"
)

const (
	providerResponseBodyLimit    int64 = 64 * 1024 * 1024
	configProviderMaximumTimeout       = 30 * time.Second
)

var (
	errProviderResponseTooLarge = errors.New("provider response exceeds size limit")
	errProviderProxyForbidden   = errors.New("provider proxy is not allowed")
)

type providerResolveFunc func(context.Context, string) ([]net.IP, error)
type providerDialFunc func(context.Context, string, string) (net.Conn, error)

type providerFetchOptions struct {
	AllowLocal              bool
	AllowProxy              bool
	RedactSource            bool
	PropagateProviderErrors bool
	Resolve                 providerResolveFunc
	Dial                    providerDialFunc
	Transport               http.RoundTripper
	Timeout                 time.Duration
	MaxTimeout              time.Duration
	Context                 context.Context
}

func browserProviderFetchOptions() providerFetchOptions {
	return providerFetchOptions{AllowLocal: true, AllowProxy: true, Context: context.TODO()}
}

func configProviderFetchOptions(ctx context.Context) providerFetchOptions {
	if ctx == nil {
		ctx = context.TODO()
	}
	return providerFetchOptions{RedactSource: true, PropagateProviderErrors: true, Context: ctx, MaxTimeout: configProviderMaximumTimeout}
}

func providerSourceForLog(source string, options providerFetchOptions) string {
	if options.RedactSource {
		return "[configured source]"
	}
	return sanitizeProviderSourceForLog(source)
}

// fileType: Welcher Dateityp soll aktualisiert werden (m3u, hdhr, xml) | fileID: Update einer bestimmten Datei (Provider ID)
func getProviderData(fileType, fileID string) (err error) {
	return getProviderDataWithOptions(fileType, fileID, browserProviderFetchOptions())
}

func getProviderDataWithOptions(fileType, fileID string, fetchOptions providerFetchOptions) (err error) {

	var fileExtension, serverFileName string
	var body = make([]byte, 0)
	var newProvider bool
	var dataMap = make(map[string]interface{})
	var providerErr error

	var saveDateFromProvider = func(fileSource, serverFileName, id string, body []byte) (err error) {

		var data = make(map[string]interface{})

		if value, ok := dataMap[id].(map[string]interface{}); ok {
			data = value
		} else {
			data["id.provider"] = id
			dataMap[id] = data
		}

		// Default keys für die Providerdaten
		var keys = []string{"name", "description", "type", "file." + System.AppName, "file.source", "tuner", "http_proxy.ip", "http_proxy.port", "last.update", "compatibility", "counter.error", "counter.download", "provider.availability"}

		for _, key := range keys {

			if _, ok := data[key]; !ok {

				switch key {

				case "name":
					data[key] = serverFileName

				case "description":
					data[key] = ""

				case "type":
					data[key] = fileType

				case "file." + System.AppName:
					data[key] = id + fileExtension

				case "file.source":
					data[key] = fileSource

				case "http_proxy.ip":
					data[key] = ""

				case "http_proxy.port":
					data[key] = ""

				case "last.update":
					data[key] = time.Now().Format("2006-01-02 15:04:05")

				case "tuner":
					if fileType == "m3u" || fileType == "hdhr" {
						if _, ok := data[key].(float64); !ok {
							data[key] = 1
						}
					}

				case "compatibility":
					data[key] = make(map[string]interface{})

				case "counter.download":
					data[key] = 0.0

				case "counter.error":
					data[key] = 0.0

				case "provider.availability":
					data[key] = 100
				}

			}

		}

		if _, ok := data["id.provider"]; !ok {
			data["id.provider"] = id
		}

		// Datei extrahieren
		body, err = extractGZIP(body, providerSourceForLog(fileSource, fetchOptions))
		if err != nil {
			ShowError(err, 000)
			return
		}

		// Daten überprüfen
		showInfo("Check File:" + providerSourceForLog(fileSource, fetchOptions))

		switch fileType {

		case "m3u":
			newM3u, err := m3u.MakeInterfaceFromM3U(body)
			if err != nil {
				return err
			}

			var m3uContent strings.Builder
			m3uContent.WriteString("#EXTM3U\n")

			for _, channel := range newM3u {
				channelMap := channel.(map[string]string)

				extinf := fmt.Sprintf(`#EXTINF:-1 tvg-id="%s" tvg-name="%s" tvg-chno="%s" tvg-logo="%s" group-title="%s",%s`,
					channelMap["tvg-id"],
					channelMap["tvg-name"],
					channelMap["tvg-chno"],
					channelMap["tvg-logo"],
					channelMap["group-title"],
					channelMap["name"],
				)

				m3uContent.WriteString(extinf + "\n" + channelMap["url"] + "\n")
			}

			m3uBytes := []byte(m3uContent.String())
			body = m3uBytes

		case "hdhr":
			_, err = jsonToInterface(string(body))

		case "xmltv":
			err = checkXMLCompatibility(id, body)

		}

		if err != nil {
			return
		}

		var filePath = System.Folder.Data + data["file."+System.AppName].(string)

		err = writeByteToFile(filePath, body)

		if err == nil {
			data["last.update"] = time.Now().Format("2006-01-02 15:04:05")
			data["counter.download"] = data["counter.download"].(float64) + 1
		}

		return

	}

	switch fileType {

	case "m3u":
		dataMap = Settings.Files.M3U
		fileExtension = ".m3u"

	case "hdhr":
		dataMap = Settings.Files.HDHR
		fileExtension = ".json"

	case "xmltv":
		dataMap = Settings.Files.XMLTV
		fileExtension = ".xml"

	}

	for dataID, d := range dataMap {

		var data = d.(map[string]interface{})
		var fileSource = data["file.source"].(string)
		var httpProxyIp = ""
		if data["http_proxy.ip"] != nil {
			httpProxyIp = data["http_proxy.ip"].(string)
		}
		var httpProxyPort = ""
		if data["http_proxy.port"] != nil {
			httpProxyPort = data["http_proxy.port"].(string)
		}
		var httpProxyUrl = ""
		if httpProxyIp != "" && httpProxyPort != "" {
			httpProxyUrl = fmt.Sprintf("http://%s:%s", httpProxyIp, httpProxyPort)
		}

		newProvider = false

		if _, ok := data["new"]; ok {
			newProvider = true
			delete(data, "new")
		}

		// Wenn eine ID vorhanden ist und nicht mit der aus der Datanbank übereinstimmt, wird die Aktualisierung übersprungen (goto)
		if len(fileID) > 0 && newProvider == false {
			if dataID != fileID {
				goto Done
			}
		}

		switch fileType {

		case "hdhr":

			// Laden vom HDHomeRun Tuner
			showInfo("Tuner:" + providerSourceForLog(fileSource, fetchOptions))
			var tunerURL = "http://" + fileSource + "/lineup.json"
			serverFileName, body, err = downloadFileFromServerWithOptions(tunerURL, httpProxyUrl, fetchOptions)

		default:

			if isRemoteProviderSource(fileSource) {

				// Laden vom Remote Server
				showInfo("Download:" + providerSourceForLog(fileSource, fetchOptions))
				serverFileName, body, err = downloadFileFromServerWithOptions(fileSource, httpProxyUrl, fetchOptions)

			} else {
				if !fetchOptions.AllowLocal {
					err = errors.New("local provider sources are not available for configuration actions")
					break
				}

				// Laden einer lokalen Datei
				showInfo("Open:" + providerSourceForLog(fileSource, fetchOptions))

				err = checkFile(fileSource)
				if err == nil {
					body, err = readByteFromFile(fileSource)
					serverFileName = getFilenameFromPath(fileSource)
				}

			}

		}

		if err == nil {

			err = saveDateFromProvider(fileSource, serverFileName, dataID, body)
			if err == nil {
				showInfo("Save File:" + providerSourceForLog(fileSource, fetchOptions) + " [ID: " + dataID + "]")
			}

		}

		if err != nil {

			ShowError(err, 000)
			var downloadErr = err
			providerErr = errors.Join(providerErr, downloadErr)

			if newProvider == false {

				// Prüfen ob ältere Datei vorhanden ist
				var file = System.Folder.Data + dataID + fileExtension

				err = checkFile(file)
				if err == nil {

					if len(fileID) == 0 {
						showWarning(1011)
					}

				}

				// Fehler Counter um 1 erhöhen
				if value, ok := dataMap[dataID].(map[string]interface{}); ok {

					data := value
					data["counter.error"] = data["counter.error"].(float64) + 1
					data["counter.download"] = data["counter.download"].(float64) + 1

				}

			} else {
				return downloadErr
			}

		}

		// Berechnen der Fehlerquote
		if newProvider == false {

			if value, ok := dataMap[dataID].(map[string]interface{}); ok {

				data := value

				if data["counter.error"].(float64) == 0 {
					data["provider.availability"] = 100
				} else {
					data["provider.availability"] = int(data["counter.error"].(float64)*100/data["counter.download"].(float64)*-1 + 100)
				}

			}

		}

		switch fileType {

		case "m3u":
			Settings.Files.M3U = dataMap

		case "hdhr":
			Settings.Files.HDHR = dataMap

		case "xmltv":
			Settings.Files.XMLTV = dataMap
			delete(Data.Cache.XMLTV, System.Folder.Data+dataID+fileExtension)

		}

		if saveErr := saveSettings(Settings); saveErr != nil {
			if fetchOptions.PropagateProviderErrors {
				return errors.Join(providerErr, saveErr)
			}
			return saveErr
		}

	Done:
	}

	if fetchOptions.PropagateProviderErrors {
		return providerErr
	}
	return nil
}

func providerDestinationAllowed(ip net.IP) bool {
	if ip == nil || ip.IsLoopback() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() || ip.IsMulticast() {
		return false
	}
	for _, blocked := range []string{"100.100.100.200", "168.63.129.16", "fd00:ec2::254"} {
		if ip.Equal(net.ParseIP(blocked)) {
			return false
		}
	}
	return true
}

func defaultProviderResolve(ctx context.Context, host string) ([]net.IP, error) {
	return net.DefaultResolver.LookupIP(ctx, "ip", host)
}

func providerResolveAllowed(ctx context.Context, host string, resolve providerResolveFunc) ([]net.IP, error) {
	trimmedHost := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
	switch trimmedHost {
	case "", "localhost", "metadata.google.internal", "metadata.azure.internal", "metadata.azure.com":
		return nil, errors.New("provider destination is restricted")
	}
	if resolve == nil {
		resolve = defaultProviderResolve
	}
	addresses, err := resolve(ctx, trimmedHost)
	if err != nil || len(addresses) == 0 {
		return nil, errors.New("provider destination could not be resolved")
	}
	for _, address := range addresses {
		if !providerDestinationAllowed(address) {
			return nil, errors.New("provider destination is restricted")
		}
	}
	return addresses, nil
}

func providerSafeDialContext(resolve providerResolveFunc, dial providerDialFunc) providerDialFunc {
	if dial == nil {
		dialer := &net.Dialer{}
		dial = dialer.DialContext
	}
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, errors.New("provider destination is invalid")
		}
		addresses, err := providerResolveAllowed(ctx, host, resolve)
		if err != nil {
			return nil, err
		}
		var lastErr error
		for _, resolved := range addresses {
			connection, dialErr := dial(ctx, network, net.JoinHostPort(resolved.String(), port))
			if dialErr == nil {
				return connection, nil
			}
			lastErr = dialErr
		}
		return nil, lastErr
	}
}

func validateProviderURLDestination(ctx context.Context, providerURL *url.URL, resolve providerResolveFunc) error {
	if providerURL == nil || (providerURL.Scheme != "http" && providerURL.Scheme != "https") || providerURL.Host == "" {
		return errors.New("provider destination is invalid")
	}
	_, err := providerResolveAllowed(ctx, providerURL.Hostname(), resolve)
	return err
}

func providerRedirectPolicy(resolve providerResolveFunc) func(*http.Request, []*http.Request) error {
	return func(request *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("provider redirect limit exceeded")
		}
		if err := validateProviderURLDestination(request.Context(), request.URL, resolve); err != nil {
			return errors.New("provider redirect destination is restricted")
		}
		return nil
	}
}

func providerRedirectLimitPolicy(request *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("provider redirect limit exceeded")
	}
	return nil
}

func downloadFileFromServer(providerURL string, proxyURL string) (filename string, body []byte, err error) {
	return downloadFileFromServerWithOptions(providerURL, proxyURL, browserProviderFetchOptions())
}

func downloadFileFromServerWithOptions(providerURL string, proxyURL string, options providerFetchOptions) (filename string, body []byte, err error) {
	logSource := providerSourceForLog(providerURL, options)
	parsedProviderURL, err := url.ParseRequestURI(providerURL)
	if err != nil || parsedProviderURL.Host == "" || (parsedProviderURL.Scheme != "http" && parsedProviderURL.Scheme != "https") {
		return "", nil, fmt.Errorf("invalid provider source: %s", logSource)
	}
	if proxyURL != "" && !options.AllowProxy {
		return "", nil, errProviderProxyForbidden
	}
	requestTimeout := options.Timeout
	if requestTimeout <= 0 {
		requestTimeout = configuredProviderRequestTimeout(Settings.BufferTimeout)
	}
	if options.MaxTimeout > 0 && requestTimeout > options.MaxTimeout {
		requestTimeout = options.MaxTimeout
	}
	requestContext := options.Context
	if requestContext == nil {
		requestContext = context.TODO()
	}
	requestContext, cancel := context.WithTimeout(requestContext, requestTimeout)
	defer cancel()

	resolve := options.Resolve
	if resolve == nil {
		resolve = defaultProviderResolve
	}
	proxied := proxyURL != ""
	if !proxied {
		if err := validateProviderURLDestination(requestContext, parsedProviderURL, resolve); err != nil {
			return "", nil, errors.New("provider destination is restricted")
		}
	}
	transport := options.Transport
	if transport == nil {
		httpTransport := &http.Transport{}
		if proxied {
			parsedProxyURL, parseErr := url.Parse(proxyURL)
			if parseErr != nil || parsedProxyURL.Host == "" || (parsedProxyURL.Scheme != "http" && parsedProxyURL.Scheme != "https") {
				return "", nil, errors.New("provider proxy is invalid")
			}
			httpTransport.Proxy = http.ProxyURL(parsedProxyURL)
		} else {
			httpTransport.DialContext = providerSafeDialContext(resolve, options.Dial)
		}
		transport = httpTransport
	}
	redirectPolicy := providerRedirectPolicy(resolve)
	if proxied {
		redirectPolicy = providerRedirectLimitPolicy
	}
	httpClient := &http.Client{Transport: transport, CheckRedirect: redirectPolicy}
	req, err := http.NewRequestWithContext(requestContext, http.MethodGet, providerURL, nil)
	if err != nil {
		return "", nil, fmt.Errorf("invalid provider request for %s", logSource)
	}
	req.Header.Set("User-Agent", Settings.UserAgent)
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("provider request failed for %s", logSource)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("%d: %s %s", resp.StatusCode, logSource, http.StatusText(resp.StatusCode))
	}
	index := strings.Index(resp.Header.Get("Content-Disposition"), "filename")
	if index > -1 {
		headerFilename := resp.Header.Get("Content-Disposition")[index:]
		value := strings.SplitN(headerFilename, `=`, 2)
		if len(value) == 2 {
			filename = strings.ReplaceAll(strings.ReplaceAll(value[1], `"`, ""), `;`, "")
		}
	}
	if filename == "" {
		cleanFilename := strings.SplitN(getFilenameFromPath(providerURL), "?", 2)
		filename = cleanFilename[0]
	}
	body, err = io.ReadAll(io.LimitReader(resp.Body, providerResponseBodyLimit+1))
	if err != nil {
		return "", nil, errors.New("provider response could not be read")
	}
	if int64(len(body)) > providerResponseBodyLimit {
		return "", nil, errProviderResponseTooLarge
	}
	return filename, body, nil
}

func isRemoteProviderSource(source string) bool {
	lower := strings.ToLower(strings.TrimSpace(source))
	return strings.HasPrefix(lower, "http:") || strings.HasPrefix(lower, "https:")
}

func sanitizeProviderSourceForLog(source string) string {
	trimmed := strings.TrimSpace(source)
	if !isRemoteProviderSource(trimmed) {
		return source
	}

	providerURL, err := url.Parse(trimmed)
	if err != nil || providerURL.Host == "" {
		return "[redacted source]"
	}
	providerURL.User = nil
	providerURL.RawQuery = ""
	providerURL.ForceQuery = false
	providerURL.Fragment = ""
	return providerURL.String()
}

func configuredProviderRequestTimeout(bufferTimeout float64) time.Duration {
	if math.IsNaN(bufferTimeout) {
		return 30 * time.Second
	}
	if bufferTimeout <= 0 {
		return 30 * time.Second
	}
	milliseconds := bufferTimeout * 1000
	// Preserve a finite timeout for every positive configured value. One
	// millisecond is the smallest duration represented by this conversion;
	// zero would instead disable http.Client's timeout entirely.
	if milliseconds < 1 {
		return time.Millisecond
	}
	const maxDurationMilliseconds = float64((1<<63 - 1) / int64(time.Millisecond))
	if milliseconds >= maxDurationMilliseconds {
		return time.Duration(1<<63 - 1)
	}
	return time.Duration(milliseconds) * time.Millisecond
}
