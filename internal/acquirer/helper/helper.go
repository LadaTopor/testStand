package helper

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"hash"
	"net/url"
	"strconv"
	"testStand/internal/acquirer"
)

// JoinUrl
func JoinUrl(baseUrl string, endpointParts ...string) string {
	path, _ := url.JoinPath(baseUrl, endpointParts...)
	return path
}

// GenerateSHA1Hash.
// Deprecated: use GenerateHash istead it
func GenerateSHA1Hash(data string) []byte {
	hash := sha1.New()
	hash.Write([]byte(data))
	return hash.Sum(nil)
}

// GenerateHash. hashFunc must not be used before to avoid unpredictable results.
func GenerateHash(hashFunc hash.Hash, data []byte) []byte {
	hashFunc.Write(data)
	return hashFunc.Sum(nil)
}

// GenerateHMAC. hashFunc must not be used before to avoid unpredictable results.
func GenerateHMAC(hashFunc func() hash.Hash, data []byte, key string) []byte {
	if hashFunc == nil {
		return nil
	}

	hmacer := hmac.New(hashFunc, []byte(key))
	hmacer.Write(data)
	return hmacer.Sum(nil)
}

// GenerateMD5Hash
func GenerateMD5Hash(text string) string {
	hash := md5.New()
	hash.Write([]byte(text))
	return hex.EncodeToString(hash.Sum(nil))
}

// DecodeUnicode
func DecodeUnicode(unicodeString string) string {
	decodedString, _ := strconv.Unquote(`"` + unicodeString + `"`)
	return decodedString
}

// UnsupportedMethodError
func UnsupportedMethodError() (*acquirer.TransactionStatus, error) {
	return &acquirer.TransactionStatus{
		Status:   acquirer.REJECTED,
		TxnError: acquirer.NewTxnError(4001, "Method not implemented"),
	}, nil
}
