package helper

import (
	"errors"
	"fmt"
	"regexp"
	"unicode"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IsURL is test is url
func IsURL(token string) bool {
	return regexp.MustCompile(`^(?i)https?://[\w\-]+(\.[\w\-]+){1,}`).MatchString(token)
}

// IsObjectID is object id
func IsObjectID(v string) bool {
	_, err := primitive.ObjectIDFromHex(v)
	return err == nil
}

// 检测是否为英文
func IsEnglish(str string) bool {
	for _, ch := range str {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || unicode.IsNumber(ch) || unicode.IsSpace(ch) || unicode.IsPunct(ch)) {
			return false
		}
	}
	return true
}

// SignHash creates a hash for signing a message.
func SignHash(data []byte) []byte {
	return crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)))
}

// EnsureOwner verifies that the signature corresponds to the given address.
func EnsureOwner(address, message, signature string) (common.Address, error) {
	address1 := common.HexToAddress(address)
	rawSig := common.FromHex(signature)

	if len(rawSig) != 65 { // Ethereum signatures are 65 bytes
		return common.Address{}, errors.New("bad signature length")
	}

	rawSig[64] -= 27 // Adjust the recovery ID

	publicKey, err := crypto.SigToPub(SignHash([]byte(message)), rawSig)
	if err != nil {
		return common.Address{}, err
	}

	if owner := crypto.PubkeyToAddress(*publicKey); owner != address1 {
		return common.Address{}, errors.New("mismatch")
	}

	return address1, nil
}
