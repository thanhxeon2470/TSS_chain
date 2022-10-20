package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"
	"math/big"

	"github.com/thanhxeon2470/TSS_chain/utils"

	"golang.org/x/crypto/ripemd160"
)

const Version = byte(0x00)
const AddressChecksumLen = 4

// Wallet stores private and public keys
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// NewWallet creates and returns a Wallet
func NewWallet() (*Wallet, error) {
	private, public := newKeyPair()
	wallet := Wallet{private, public}

	return &wallet, nil
}

// GetAddress returns wallet address
func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey)

	versionedPayload := append([]byte{Version}, pubKeyHash...)
	checksum := Checksum(versionedPayload)

	fullPayload := append(versionedPayload, checksum...)
	address := utils.Base58Encode(fullPayload)

	return address
}

func DecodeAddress(addr string) []byte {
	rawAddr := utils.Base58Decode([]byte(addr))
	if len(rawAddr) <= 4 {
		rawAddr = []byte{0x00, 0x00, 0x00, 0x00}
	}
	pubkeyHash := rawAddr[1 : len(rawAddr)-AddressChecksumLen]

	return pubkeyHash
}

// HashPubKey hashes public key
func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}

// Private Key econder
func EncodePrivKey(privKey ecdsa.PrivateKey) []byte {
	return utils.Base58Encode(privKey.D.Bytes())
}

func EncodePubkey(pubkey []byte) []byte {
	return utils.Base58Encode(pubkey)
}

// Private Key decoder
func DecodePrivKey(encoded []byte) *Wallet {
	// Decode base58
	bytes := utils.Base58Decode(encoded)
	// Allocate new Private Key
	var key ecdsa.PrivateKey
	key.D = new(big.Int).SetBytes(bytes)
	// Compute compatiable PublicKey
	key.PublicKey.Curve = elliptic.P256()
	key.PublicKey.X, key.PublicKey.Y = key.PublicKey.Curve.ScalarBaseMult(key.D.Bytes())

	pubKey := append(key.PublicKey.X.Bytes(), key.PublicKey.Y.Bytes()...)

	w := Wallet{key, pubKey}
	return &w
}

// ValidateAddress check if address if valid
func ValidateAddress(address string) bool {
	pubKeyHash := utils.Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-AddressChecksumLen:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-AddressChecksumLen]
	targetChecksum := Checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

// Checksum generates a checksum for a public key
func Checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:AddressChecksumLen]
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pubKey
}
