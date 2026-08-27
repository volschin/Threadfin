package src

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"threadfin/src/internal/authentication"
)

func activatedSystemAuthentication() (err error) {

	err = authentication.Init(System.Folder.Config, 60)
	if err != nil {
		return
	}

	var defaults = make(map[string]interface{})
	defaults["authentication.web"] = false
	defaults["authentication.pms"] = false
	defaults["authentication.xml"] = false
	defaults["authentication.api"] = false
	defaults["authentication.config"] = false
	err = authentication.SetDefaultUserData(defaults)

	return
}

func createFirstUserForAuthentication(username, password string) (sessionID string, err error) {
	if err = authentication.CreateDefaultUser(username, password); err != nil {
		return "", err
	}

	var userID string
	sessionID, userID, err = authentication.AuthenticateBrowser(username, password)
	if err != nil {
		return "", err
	}

	var userData = make(map[string]interface{})
	userData["username"] = username
	userData["authentication.web"] = true
	userData["authentication.pms"] = true
	userData["authentication.m3u"] = true
	userData["authentication.xml"] = true
	userData["authentication.api"] = false
	userData["authentication.config"] = false
	userData["defaultUser"] = true

	if err = authentication.WriteUserData(userID, userData); err != nil {
		authentication.InvalidateBrowserSession(sessionID)
		return "", err
	}
	if _, err = authentication.AuthorizeBrowserSession(sessionID, "authentication.web"); err != nil {
		authentication.InvalidateBrowserSession(sessionID)
		return "", err
	}

	return sessionID, nil
}

func authorizeBrowserRequest(r *http.Request, permissions ...string) (string, error) {
	cookie, err := r.Cookie(authentication.BrowserSessionCookieName)
	if err != nil {
		return "", err
	}
	return authentication.AuthorizeBrowserSession(cookie.Value, permissions...)
}

func browserCookieSecure(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	systemMutex.Lock()
	defer systemMutex.Unlock()
	return Settings.ForceHttps && Settings.HttpsThreadfinDomain != ""
}

func tokenAuthentication(token string) (newToken string, err error) {

	if System.ConfigurationWizard == true {
		return
	}

	newToken, err = authentication.CheckTheValidityOfTheToken(token)

	return
}

func basicAuth(r *http.Request, level string) (username string, err error) {

	err = errors.New("User authentication failed")

	auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

	if len(auth) != 2 || auth[0] != "Basic" {
		return
	}

	payload, _ := base64.StdEncoding.DecodeString(auth[1])
	pair := strings.SplitN(string(payload), ":", 2)

	username = pair[0]
	var password = pair[1]

	token, err := authentication.UserAuthentication(username, password)

	if err != nil {
		return
	}

	err = checkAuthorizationLevel(token, level)

	return
}

func urlAuth(r *http.Request, requestType string) (err error) {
	var level, token string

	var username = r.URL.Query().Get("username")
	var password = r.URL.Query().Get("password")

	switch requestType {

	case "m3u":
		level = "authentication.m3u"
		if Settings.AuthenticationM3U == true {
			token, err = authentication.UserAuthentication(username, password)
			if err != nil {
				return
			}
			err = checkAuthorizationLevel(token, level)
		}

	case "xml":
		level = "authentication.xml"
		if Settings.AuthenticationXML == true {
			token, err = authentication.UserAuthentication(username, password)
			if err != nil {
				return
			}
			err = checkAuthorizationLevel(token, level)
		}

	}

	return
}

func checkAuthorizationLevel(token, level string) (err error) {
	userID, err := authentication.GetUserID(token)
	if err != nil {
		return err
	}

	userData, err := authentication.ReadUserData(userID)
	if err != nil {
		return err
	}

	if len(userData) > 0 {

		if v, ok := userData[level].(bool); ok {

			if v == false {
				err = errors.New("No authorization")
			}

		} else {
			userData[level] = false
			if err = authentication.WriteUserData(userID, userData); err != nil {
				return err
			}
			err = errors.New("No authorization")
		}

	} else {
		if err = authentication.WriteUserData(userID, userData); err != nil {
			return err
		}
		err = errors.New("No authorization")
	}

	return
}
