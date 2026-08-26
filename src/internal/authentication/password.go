package authentication

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	passwordMemory      uint32 = 19456
	passwordIterations  uint32 = 2
	passwordParallelism uint8  = 1
	passwordSaltLength         = 16
	passwordKeyLength   uint32 = 32
)

func hashPassword(password string) (string, error) {
	salt := make([]byte, passwordSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, passwordIterations, passwordMemory, passwordParallelism, passwordKeyLength)
	return serializeArgon2IDHash(salt, key), nil
}

func serializeArgon2IDHash(salt, key []byte) string {
	encoded := make([]byte, 0,
		len("$argon2id$v=19$m=19456,t=2,p=1$$")+
			base64.RawStdEncoding.EncodedLen(len(salt))+
			base64.RawStdEncoding.EncodedLen(len(key)),
	)
	encoded = append(encoded, "$argon2id$v="...)
	encoded = strconv.AppendInt(encoded, int64(argon2.Version), 10)
	encoded = append(encoded, "$m="...)
	encoded = strconv.AppendInt(encoded, int64(passwordMemory), 10)
	encoded = append(encoded, ",t="...)
	encoded = strconv.AppendInt(encoded, int64(passwordIterations), 10)
	encoded = append(encoded, ",p="...)
	encoded = strconv.AppendInt(encoded, int64(passwordParallelism), 10)
	encoded = append(encoded, '$')
	encoded = base64.RawStdEncoding.AppendEncode(encoded, salt)
	encoded = append(encoded, '$')
	encoded = base64.RawStdEncoding.AppendEncode(encoded, key)
	return string(encoded)
}

func verifyPassword(password, stored string) (matches bool, legacy bool) {
	if !strings.HasPrefix(stored, "$argon2id$") {
		legacyHash := SHA256(password, "")
		return subtle.ConstantTimeCompare([]byte(legacyHash), []byte(stored)) == 1, true
	}

	parts := strings.Split(stored, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" || parts[3] != "m=19456,t=2,p=1" {
		return false, false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) != passwordSaltLength {
		return false, false
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(expected) != int(passwordKeyLength) {
		return false, false
	}
	actual := argon2.IDKey([]byte(password), salt, passwordIterations, passwordMemory, passwordParallelism, passwordKeyLength)
	return subtle.ConstantTimeCompare(actual, expected) == 1, false
}
