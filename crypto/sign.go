package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"

	"github.com/gitferry/bamboo/identity"

	"github.com/gitferry/bamboo/config"
)

// SigningAlgorithm is an identifier for a signing algorithm and curve.

//type SigningAlgorithm string

// String returns the string representation of this signing algorithm.
// func (f SigningAlgorithm) String() string {
//	return [...]string{"UNKNOWN", "BLS_BLS12381", "ECDSA_P256", "ECDSA_SECp256k1"}[f]
//}

const (
	// Supported signing algorithms
	//UnknownSigningAlgorithm SigningAlgorithm = iota
	BLS_BLS12381    = "BLS_BLS12381"
	ECDSA_P256      = "ECDSA_P256"
	ECDSA_SECp256k1 = "ECDSA_SECp256k1"
)

var keys []PrivateKey
var pubKeys []PublicKey

// PrivateKey is an unspecified signature scheme private key
type PrivateKey interface {
	// Algorithm returns the signing algorithm related to the private key.
	Algorithm() string
	// KeySize return the key size in bytes.
	// KeySize() int
	// Sign generates a signature using the provided hasher.
	Sign([]byte, Hasher) (Signature, error)
	// PublicKey returns the public key.
	PublicKey() PublicKey
	// Encode returns a bytes representation of the private key
	//Encode() ([]byte, error)
}

// PublicKey is an unspecified signature scheme public key.
type PublicKey interface {
	// Algorithm returns the signing algorithm related to the public key.
	Algorithm() string
	// KeySize return the key size in bytes.
	//KeySize() int
	// Verify verifies a signature of an input message using the provided hasher.
	Verify(Signature, Hash) (bool, error)
	// Encode returns a bytes representation of the public key.
	//Encode() ([]byte, error)
}

func SetKeys() error {
	keys = make([]PrivateKey, config.GetConfig().N())
	pubKeys = make([]PublicKey, config.GetConfig().N())
	var err error
	for i := 0; i < config.GetConfig().N(); i++ {
		keys[i], err = GenerateKey(config.GetConfig().GetSignatureScheme())
		if err != nil {
			return err
		}
		pubKeys[i] = keys[i].PublicKey()
	}
	return nil
}

func GenerateKey(signer string) (PrivateKey, error) {
	if signer == ECDSA_P256 {
		pubkeyCurve := elliptic.P256()
		priv, err := ecdsa.GenerateKey(pubkeyCurve, rand.Reader)
		if err != nil {
			return nil, err
		}
		privKey := &ecdsa_p256_PrivateKey{SignAlg: signer, PrivateKey: priv}
		return privKey, nil
	} else if signer == ECDSA_SECp256k1 {
		return nil, nil
	} else if signer == BLS_BLS12381 {
		return nil, nil
	} else {
		return nil, errors.New("Invalid signature scheme!")
	}
}

// Use the following functions for signing and verification.

func PrivSign(data []byte, nodeID identity.NodeID, hasher Hasher) (Signature, error) {
	return keys[nodeID.Node()].Sign(data, hasher)
}

func PubVerify(sig Signature, data []byte, nodeID identity.NodeID) (bool, error) {
	return pubKeys[nodeID.Node()].Verify(sig, data)
}

func VerifyQuorumSignature(leaderSig Signature, qcID Identifier, leader identity.NodeID, aggregatedSigs AggSig, blockID Identifier, aggSigners []identity.NodeID) (bool, error) {
	leaderSigCorrect, err := PubVerify(leaderSig, IDToByte(qcID), leader)
	if err != nil {
		return false, err
	}
	if leaderSigCorrect == false {
		return false, nil
	}
	var sigIsCorrect bool
	var errAgg error
	for i, signer := range aggSigners {
		sigIsCorrect, errAgg = PubVerify(aggregatedSigs[i], IDToByte(blockID), signer)
		if errAgg != nil {
			return false, errAgg
		}
		if sigIsCorrect == false {
			return false, nil
		}
	}
	return true, nil
}
