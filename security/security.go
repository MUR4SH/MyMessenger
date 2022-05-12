package security

import (
	"crypto/rand"
	"crypto/rsa"
	"log"
	"math/big"

	"github.com/MUR4SH/MyMessenger/structures"
)

func Encrypt(s string, key *rsa.PublicKey) []byte {
	crypt, err := rsa.EncryptPKCS1v15(rand.Reader, key, []byte(s))

	if err != nil {
		log.Println("err")
		log.Println(err)
	}

	return crypt
}

func PublicKeyTransform(key *rsa.PublicKey) structures.EditedPublicKey {
	var res structures.EditedPublicKey
	res.E = key.E
	res.N = key.N.String()

	return res
}

func CRTValueTransform(key *rsa.CRTValue) structures.EditedCRTValue {
	var res structures.EditedCRTValue

	res.Exp = key.Exp.String()
	res.Coeff = key.Coeff.String()
	res.R = key.R.String()

	return res
}

func PrecomputedTransform(key *rsa.PrecomputedValues) structures.EditedPrecomputedValues {
	var res structures.EditedPrecomputedValues

	res.Dp = key.Dp.String()
	res.Dq = key.Dq.String()
	res.Qinv = key.Qinv.String()

	var arr []structures.EditedCRTValue
	for i := 0; i < len(key.CRTValues); i++ {
		arr = append(arr, CRTValueTransform(&key.CRTValues[i]))
	}
	res.CRTValues = arr

	return res
}

func PrivateKeyTransform(key *rsa.PrivateKey) structures.EditedPrivateKey {
	var res structures.EditedPrivateKey
	res.PublicKey = PublicKeyTransform(&key.PublicKey)
	res.D = key.D.String()

	var arr []string
	for i := 0; i < len(key.Primes); i++ {
		arr = append(arr, key.Primes[i].String())
	}
	res.Primes = arr

	res.Precomputed = PrecomputedTransform(&key.Precomputed)

	return res
}

func PublicKeyDecode(key *structures.EditedPublicKey) rsa.PublicKey {
	var res rsa.PublicKey
	n := new(big.Int)

	res.E = key.E
	n, _ = n.SetString(key.N, 10)
	res.N = n

	return res
}

func CRTValueDecode(key *structures.EditedCRTValue) rsa.CRTValue {
	var res rsa.CRTValue
	n := new(big.Int)

	n, _ = n.SetString(key.Exp, 10)
	res.Exp = n

	n, _ = n.SetString(key.Coeff, 10)
	res.Coeff = n

	n, _ = n.SetString(key.R, 10)
	res.R = n

	return res
}

func PrecomputedDecode(key *structures.EditedPrecomputedValues) rsa.PrecomputedValues {
	var res rsa.PrecomputedValues
	n := new(big.Int)

	n, _ = n.SetString(key.Dp, 10)
	res.Dp = n

	n, _ = n.SetString(key.Dq, 10)
	res.Dq = n

	n, _ = n.SetString(key.Qinv, 10)
	res.Qinv = n

	var arr []rsa.CRTValue
	for i := 0; i < len(key.CRTValues); i++ {
		arr = append(arr, CRTValueDecode(&key.CRTValues[i]))
	}
	res.CRTValues = arr

	return res
}

func PrivateKeyDecode(key *structures.EditedPrivateKey) rsa.PrivateKey {
	var res rsa.PrivateKey
	n := new(big.Int)

	res.PublicKey = PublicKeyDecode(&key.PublicKey)

	n, _ = n.SetString(key.D, 10)
	res.D = n

	var arr []*big.Int
	for i := 0; i < len(key.Primes); i++ {
		n, _ = n.SetString(key.Primes[i], 10)
		arr = append(arr, n)
	}
	res.Primes = arr

	res.Precomputed = PrecomputedDecode(&key.Precomputed)

	return res
}
