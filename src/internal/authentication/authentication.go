package authentication

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"

	"time"
	//"fmt"
	//"log"
)

const tokenLength = 40
const saltLength = 20
const idLength = 10

var tokenValidity int
var database string

var databaseFile = "authentication.json"

var data = make(map[string]interface{})
var tokens = make(map[string]interface{})

var initAuthentication = false
var authenticationMutex sync.RWMutex

// Cookie : cookie
type Cookie struct {
	Name       string
	Value      string
	Path       string
	Domain     string
	Expires    time.Time
	RawExpires string
}

// Framework examples

/*
func main() {
  var err error

  var checkErr = func(err error) {
     log.Println(err)
     os.Exit(0)
  }

  err = Init("", 10)          // Path to save the data, Validity of tokens in minutes | (error)
  if err != nil {
    checkErr(err)
  }


  err = CreateDefaultUser("admin", "123")
  if err != nil {
    checkErr(err)
  }




  err = CreateNewUser("threadfin", "threadfin")          // Username, Password | (error)
  if err != nil {
    checkErr(err)
  }



  err, token := UserAuthentication("threadfin", "threadfin")          // Username, Password | (error, token)
  if err != nil {
    checkErr(err)
  } else {
    fmt.Println("UserAuthentication()")
    fmt.Println("Token:", token)
    fmt.Println("---")
  }

  err, newToken := CheckTheValidityOfTheToken(token)        // Current token | (error, new token)
  if err != nil {
    checkErr(err)
  } else {
    fmt.Println("CheckTheValidityOfTheToken()")
    fmt.Println("New Token:", newToken)
    fmt.Println("---")
  }

  err, userID := GetUserID(newToken)                        // Current token | (error, user id)
  if err != nil {
    checkErr(err)
  } else {
    fmt.Println("GetUserID()")
    fmt.Println("User ID:", userID)
    fmt.Println("---")
  }


  var userData = make(map[string]interface{})
  userData["type"] = "Administrator"
  err = WriteUserData(userID, userData)          // User id, user data | (error)
  if err != nil {
    checkErr(err)
  }

  err, userData = ReadUserData(userID)          // User id | (error, userData)
  if err != nil {
    checkErr(err)
  } else {
    fmt.Println("ReadUserData()")
    fmt.Println("User data:", userData)
    fmt.Println("---")
  }

  err = RemoveUser(userID)
  if err != nil {
    checkErr(err)
  }

}
*/

// Init : databasePath = Path to authentication.json
func Init(databasePath string, validity int) (err error) {
	authenticationMutex.Lock()
	defer authenticationMutex.Unlock()

	database = filepath.Dir(databasePath) + string(os.PathSeparator) + databaseFile

	// Check if the database already exists
	if _, err = os.Stat(database); os.IsNotExist(err) {
		// Create an empty database
		var defaults = make(map[string]interface{})
		defaults["dbVersion"] = "1.0"
		defaults["hash"] = "argon2id"
		defaults["users"] = make(map[string]interface{})

		if saveDatabase(defaults) != nil {
			return
		}
	}

	// Loading the database
	err = loadDatabase()
	browserSessions = make(map[string]browserSession)

	// Set Token Validity
	tokenValidity = validity
	initAuthentication = true
	return
}

// CreateDefaultUser = created efault user
func CreateDefaultUser(username, password string) (err error) {
	authenticationMutex.Lock()
	defer authenticationMutex.Unlock()

	err = checkInit()
	if err != nil {
		return
	}

	var users = data["users"].(map[string]interface{})
	// Check if the default user exists
	if len(users) > 0 {
		err = createError(001)
		return
	}

	defaults, err := defaultsForNewUser(username, password)
	if err != nil {
		return
	}
	users[defaults["_id"].(string)] = defaults
	if err = saveDatabase(data); err != nil {
		delete(users, defaults["_id"].(string))
		return
	}

	return
}

// CreateNewUser : create new user
func CreateNewUser(username, password string) (userID string, err error) {
	authenticationMutex.Lock()
	defer authenticationMutex.Unlock()

	err = checkInit()
	if err != nil {
		return
	}

	var checkIfTheUserAlreadyExists = func(username string, userData map[string]interface{}) (err error) {
		var salt = userData["_salt"].(string)
		var loginUsername = userData["_username"].(string)

		if SHA256(username, salt) == loginUsername {
			err = createError(020)
		}

		return
	}

	var users = data["users"].(map[string]interface{})
	for _, userData := range users {
		err = checkIfTheUserAlreadyExists(username, userData.(map[string]interface{}))
		if err != nil {
			return
		}
	}

	defaults, err := defaultsForNewUser(username, password)
	if err != nil {
		return
	}
	userID = defaults["_id"].(string)
	users[userID] = defaults

	if err = saveDatabase(data); err != nil {
		delete(users, userID)
		userID = ""
		return
	}

	return
}

// UserAuthentication : user authentication
func UserAuthentication(username, password string) (token string, err error) {
	authenticationMutex.Lock()
	defer authenticationMutex.Unlock()

	err = checkInit()
	if err != nil {
		return
	}

	userID, err := authenticateUserLocked(username, password)
	if err != nil {
		return "", err
	}
	return setToken(userID, "-"), nil
}

func authenticateUserLocked(username, password string) (userID string, err error) {
	users := data["users"].(map[string]interface{})
	for id, loginData := range users {
		user := loginData.(map[string]interface{})
		if SHA256(username, user["_salt"].(string)) != user["_username"].(string) {
			continue
		}
		loginPassword := user["_password"].(string)
		matched, legacy := verifyPassword(password, loginPassword)
		if !matched {
			return "", createError(010)
		}
		if legacy {
			migrated, hashErr := hashPassword(password)
			if hashErr != nil {
				return "", hashErr
			}
			user["_password"] = migrated
			if saveErr := saveDatabase(data); saveErr != nil {
				user["_password"] = loginPassword
				return "", saveErr
			}
		}
		return id, nil
	}
	return "", createError(010)
}

// CheckTheValidityOfTheToken : check token
func CheckTheValidityOfTheToken(token string) (newToken string, err error) {
	authenticationMutex.Lock()
	defer authenticationMutex.Unlock()

	err = checkInit()
	if err != nil {
		return
	}

	err = createError(011)

	if v, ok := tokens[token]; ok {
		var expires = v.(map[string]interface{})["expires"].(time.Time)
		var userID = v.(map[string]interface{})["id"].(string)

		if expires.Sub(time.Now().Local()) < 0 {
			return
		}

		newToken = setToken(userID, token)

		err = nil

	} else {
		return
	}

	return
}

// GetUserID : get user ID
func GetUserID(token string) (userID string, err error) {
	authenticationMutex.RLock()
	defer authenticationMutex.RUnlock()

	err = checkInit()
	if err != nil {
		return
	}

	err = createError(002)

	if v, ok := tokens[token]; ok {
		var expires = v.(map[string]interface{})["expires"].(time.Time)
		userID = v.(map[string]interface{})["id"].(string)

		if expires.Sub(time.Now().Local()) < 0 {
			return
		}

		err = nil
	}

	return
}

// WriteUserData : save user date
func WriteUserData(userID string, userData map[string]interface{}) (err error) {
	authenticationMutex.Lock()
	defer authenticationMutex.Unlock()

	err = checkInit()
	if err != nil {
		return
	}

	err = createError(030)

	if v, ok := data["users"].(map[string]interface{})[userID].(map[string]interface{}); ok {

		v["data"] = cloneAuthenticationMap(userData)
		err = saveDatabase(data)

	} else {
		return
	}

	return
}

// ReadUserData : load user date
func ReadUserData(userID string) (userData map[string]interface{}, err error) {
	authenticationMutex.RLock()
	defer authenticationMutex.RUnlock()

	err = checkInit()
	if err != nil {
		return
	}

	err = createError(031)

	if v, ok := data["users"].(map[string]interface{})[userID].(map[string]interface{}); ok {
		userData = cloneAuthenticationMap(v["data"].(map[string]interface{}))
		err = nil

		return
	}

	return
}

// PermissionGranted reads a boolean user permission without changing storage.
// Missing permissions are denied so newly introduced authority stays opt-in.
func PermissionGranted(userID, permission string) (bool, error) {
	authenticationMutex.RLock()
	defer authenticationMutex.RUnlock()
	if err := checkInit(); err != nil {
		return false, err
	}
	user, ok := data["users"].(map[string]interface{})[userID].(map[string]interface{})
	if !ok {
		return false, createError(031)
	}
	userData, ok := user["data"].(map[string]interface{})
	if !ok {
		return false, createError(031)
	}
	granted, ok := userData[permission].(bool)
	return ok && granted, nil
}

// AuthorizeTokenPermissions validates a token and all requested permissions
// against one authentication snapshot, then rotates the token only on success.
func AuthorizeTokenPermissions(token string, permissions ...string) (newToken, userID string, err error) {
	authenticationMutex.Lock()
	defer authenticationMutex.Unlock()
	if err = checkInit(); err != nil {
		return "", "", err
	}
	tokenData, ok := tokens[token].(map[string]interface{})
	if !ok {
		return "", "", createError(011)
	}
	expires, ok := tokenData["expires"].(time.Time)
	if !ok || expires.Before(time.Now().Local()) {
		return "", "", createError(011)
	}
	userID, ok = tokenData["id"].(string)
	if !ok {
		return "", "", createError(011)
	}
	user, ok := data["users"].(map[string]interface{})[userID].(map[string]interface{})
	if !ok {
		return "", "", createError(011)
	}
	userData, ok := user["data"].(map[string]interface{})
	if !ok {
		return "", "", createError(011)
	}
	for _, permission := range permissions {
		granted, exists := userData[permission].(bool)
		if !exists || !granted {
			return "", "", createError(011)
		}
	}
	newToken = setToken(userID, token)
	return newToken, userID, nil
}

// RemoveUser : remove user
func RemoveUser(userID string) (err error) {
	authenticationMutex.Lock()
	defer authenticationMutex.Unlock()

	err = checkInit()
	if err != nil {
		return
	}

	err = createError(032)

	if user, ok := data["users"].(map[string]interface{})[userID]; ok {

		delete(data["users"].(map[string]interface{}), userID)
		err = saveDatabase(data)
		if err != nil {
			data["users"].(map[string]interface{})[userID] = user
			return
		}
		invalidateUserBrowserSessionsLocked(userID)

		return
	}

	return
}

// SetDefaultUserData : set default user data
func SetDefaultUserData(defaults map[string]interface{}) (err error) {
	authenticationMutex.Lock()
	defer authenticationMutex.Unlock()
	if err = checkInit(); err != nil {
		return err
	}
	for _, d := range data["users"].(map[string]interface{}) {
		userData := d.(map[string]interface{})["data"].(map[string]interface{})
		for k, v := range defaults {
			if _, ok := userData[k]; ok {
				// Key exist
			} else {
				userData[k] = cloneAuthenticationValue(v)
			}
		}
	}
	return saveDatabase(data)
}

// ChangeCredentials : change credentials
func ChangeCredentials(userID, username, password string) (err error) {
	authenticationMutex.Lock()
	defer authenticationMutex.Unlock()
	err = checkInit()
	if err != nil {
		return
	}

	err = createError(032)

	if userData, ok := data["users"].(map[string]interface{})[userID]; ok {
		//var userData = tmp.(map[string]interface{})
		var salt = userData.(map[string]interface{})["_salt"].(string)

		if len(username) > 0 {
			userData.(map[string]interface{})["_username"] = SHA256(username, salt)
		}

		if len(password) > 0 {
			passwordHash, hashErr := hashPassword(password)
			if hashErr != nil {
				return hashErr
			}
			userData.(map[string]interface{})["_password"] = passwordHash
		}

		err = saveDatabase(data)
		if err == nil && len(password) > 0 {
			invalidateUserBrowserSessionsLocked(userID)
		}
	}

	return
}

// GetAllUserData : get all user data
func GetAllUserData() (allUserData map[string]interface{}, err error) {
	authenticationMutex.Lock()
	defer authenticationMutex.Unlock()

	err = checkInit()
	if err != nil {
		return
	}

	if len(data) == 0 {
		var defaults = make(map[string]interface{})
		defaults["dbVersion"] = "1.0"
		defaults["hash"] = "argon2id"
		defaults["users"] = make(map[string]interface{})
		if err = saveDatabase(defaults); err != nil {
			return nil, err
		}
		data = defaults
	}

	allUserData = cloneAuthenticationMap(data["users"].(map[string]interface{}))
	return
}

// CheckTheValidityOfTheTokenFromHTTPHeader : get token from HTTP header
func CheckTheValidityOfTheTokenFromHTTPHeader(w http.ResponseWriter, r *http.Request) (writer http.ResponseWriter, newToken string, err error) {
	err = createError(011)
	for _, cookie := range r.Cookies() {
		if cookie.Name == "Token" {
			var token string
			token, err = CheckTheValidityOfTheToken(cookie.Value)
			//fmt.Println("T", token, err)
			writer = SetCookieToken(w, token)
			newToken = token
		}
	}
	//fmt.Println(err)
	return
}

// Framework tools

func checkInit() (err error) {
	if initAuthentication == false {
		err = createError(000)
	}

	return
}

func saveDatabase(tmpMap interface{}) (err error) {

	jsonString, err := json.MarshalIndent(tmpMap, "", "  ")

	if err != nil {
		return
	}

	err = os.WriteFile(database, []byte(jsonString), 0600)
	if err != nil {
		return
	}

	return
}

func loadDatabase() (err error) {
	jsonString, err := os.ReadFile(database)
	if err != nil {
		return
	}

	err = json.Unmarshal([]byte(jsonString), &data)
	if err != nil {
		return
	}

	return
}

// SHA256 : password + salt = sha256 string
func SHA256(secret, salt string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte("_remote_db"))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func randomString(n int) string {
	const alphanum = "-AbCdEfGhIjKlMnOpQrStUvWxYz0123456789aBcDeFgHiJkLmNoPqRsTuVwXyZ_"

	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

func randomID(n int) string {
	const alphanum = "ABCDEFGHJKLMNOPQRSTUVWXYZ0123456789"

	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

func createError(errCode int) (err error) {
	var errMsg string
	switch errCode {
	case 000:
		errMsg = "Authentication has not yet been initialized"
	case 001:
		errMsg = "Default user already exists"
	case 002:
		errMsg = "No user id found for this token"
	case 010:
		errMsg = "User authentication failed"
	case 011:
		errMsg = "Session has expired"
	case 020:
		errMsg = "User already exists"
	case 030:
		errMsg = "User data could not be saved"
	case 031:
		errMsg = "User data could not be read"
	case 032:
		errMsg = "User ID was not found"
	}

	err = errors.New(errMsg)
	return
}

func defaultsForNewUser(username, password string) (map[string]interface{}, error) {
	passwordHash, err := hashPassword(password)
	if err != nil {
		return nil, err
	}
	var defaults = make(map[string]interface{})
	var salt = randomString(saltLength)
	defaults["_username"] = SHA256(username, salt)
	defaults["_password"] = passwordHash
	defaults["_salt"] = salt
	defaults["_id"] = "id-" + randomID(idLength)
	//defaults["_one.time.token"] = randomString(tokenLength)
	defaults["data"] = map[string]interface{}{"authentication.config": false}

	return defaults, nil
}

func setToken(id, oldToken string) (newToken string) {
	delete(tokens, oldToken)

loopToken:
	newToken = randomString(tokenLength)
	if _, ok := tokens[newToken]; ok {
		goto loopToken
	}

	var tmp = make(map[string]interface{})
	tmp["id"] = id
	tmp["expires"] = time.Now().Local().Add(time.Minute * time.Duration(tokenValidity))

	tokens[newToken] = tmp

	return
}

func cloneAuthenticationValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		return cloneAuthenticationMap(typed)
	case []interface{}:
		cloned := make([]interface{}, len(typed))
		for index, item := range typed {
			cloned[index] = cloneAuthenticationValue(item)
		}
		return cloned
	default:
		return typed
	}
}

func cloneAuthenticationMap(source map[string]interface{}) map[string]interface{} {
	cloned := make(map[string]interface{}, len(source))
	for key, value := range source {
		cloned[key] = cloneAuthenticationValue(value)
	}
	return cloned
}

// SetCookieToken : set cookie
func SetCookieToken(w http.ResponseWriter, token string) http.ResponseWriter {
	authenticationMutex.RLock()
	defer authenticationMutex.RUnlock()
	expiration := time.Now().Add(time.Minute * time.Duration(tokenValidity))
	cookie := http.Cookie{Name: "Token", Value: token, Expires: expiration}
	http.SetCookie(w, &cookie)
	return w
}
