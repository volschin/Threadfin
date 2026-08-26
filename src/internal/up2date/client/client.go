package up2date

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"time"
)

// ClientInfo : Information about the key (NAME OS, ARCH, UUID, KEY)
type ClientInfo struct {
	Arch                          string        `json:"arch"`
	Branch                        string        `json:"branch"`
	CMD                           string        `json:"cmd,omitempty"`
	Name                          string        `json:"name"`
	OS                            string        `json:"os"`
	URL                           string        `json:"url"`
	WindowsUpdateReadinessTimeout time.Duration `json:"-"`

	Response ServerResponse `json:"response,omitempty"`
}

// ServerResponse : Response from server after client request
type ServerResponse struct {
	Status    bool   `json:"status,omitempty"`
	Reason    string `json:"reason,omitempty"`
	Version   string `json:"version,omitempty"`
	UpdateBIN string `json:"update.url.bin,omitempty"`
	UpdateZIP string `json:"update.url.zip,omitempty"`
	Filename  string `json:"filename.bin,omitempty"`
}

// Updater : Client infos
var Updater ClientInfo

// UpdateURL : URL for the new binary
var UpdateURL string

// Init : Init
func Init() {
	Updater.OS = runtime.GOOS
	Updater.Arch = runtime.GOARCH
}

// GetVersion : Information about the latest version
func GetVersion() (err error) {

	Updater.CMD = "getVersion"
	err = serverRequest()
	return
}

func decodeServerResponse(reader io.Reader) (response ServerResponse, err error) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return response, err
	}
	err = json.Unmarshal(body, &response)
	return response, err
}

func serverRequest() (err error) {

	var serverResponse ServerResponse
	jsonByte, err := json.MarshalIndent(Updater, "", "  ")
	if err != nil {
		return err
	}

	// Serververbindung prüfen
	u, err := url.Parse(Updater.URL)
	if err != nil {
		return err
	}
	server := u.Host

	timeout := time.Second
	connection, err := net.DialTimeout("tcp", server, timeout)
	if err != nil {
		return err
	}
	if err := connection.Close(); err != nil {
		return err
	}

	// Check redirect 301 <---> 308
	redirect, err := http.NewRequest("POST", Updater.URL, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	client.CheckRedirect = func(redirect *http.Request, via []*http.Request) error {
		return errors.New("Redirect")
	}

	resp, err := client.Do(redirect)
	if err != nil {
		if resp == nil || resp.StatusCode < 301 || resp.StatusCode > 308 {
			return err
		}
		Updater.URL = resp.Header.Get("Location")
	}
	// ---

	req, err := http.NewRequest("POST", Updater.URL, bytes.NewBuffer(jsonByte))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%d: %s (%s)", resp.StatusCode, http.StatusText(resp.StatusCode), Updater.URL)
	}

	Updater.CMD = ""
	serverResponse, err = decodeServerResponse(resp.Body)
	if err != nil {
		return err
	}

	Updater.Response = serverResponse
	return nil
}
