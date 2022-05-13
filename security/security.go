package security

import (
	"crypto/rand"
	"crypto/rsa"
	"log"
	"math/big"

	"github.com/MUR4SH/MyMessenger/structures"
)

//Шифруем сообщение
func Encrypt(s string, key *rsa.PublicKey) []byte {
	crypt, err := rsa.EncryptPKCS1v15(rand.Reader, key, []byte(s))

	if err != nil {
		log.Println("encryption error")
		log.Println(err)
	}

	return crypt
}

//Расшифровываем сообщение
func Decrypt(s []byte, key *rsa.PrivateKey) []byte {
	decrypt, err := rsa.DecryptPKCS1v15(rand.Reader, key, s)

	if err != nil {
		log.Println("decryption error")
		log.Println(err)
	}

	return decrypt
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

//Кодируем приватный ключ в строчный для записи в бд
func PrivateKeyTransform(key *rsa.PrivateKey) *structures.EditedPrivateKey {
	var res structures.EditedPrivateKey
	res.PublicKey = PublicKeyTransform(&key.PublicKey)
	res.D = key.D.String()

	var arr []string
	for i := 0; i < len(key.Primes); i++ {
		arr = append(arr, key.Primes[i].String())
	}
	res.Primes = arr

	res.Precomputed = PrecomputedTransform(&key.Precomputed)
	return &res
}

func CRTValueDecode(key *structures.EditedCRTValue) rsa.CRTValue {
	var res rsa.CRTValue

	e := new(big.Int)
	e, _ = e.SetString(key.Exp, 10)
	res.Exp = e

	co := new(big.Int)
	co, _ = co.SetString(key.Coeff, 10)
	res.Coeff = co

	r := new(big.Int)
	r, _ = r.SetString(key.R, 10)
	res.R = r

	return res
}

func PrecomputedDecode(key *structures.EditedPrecomputedValues) rsa.PrecomputedValues {
	var res rsa.PrecomputedValues

	dp := new(big.Int)
	dp, _ = dp.SetString(key.Dp, 10)
	res.Dp = dp

	dq := new(big.Int)
	dq, _ = dq.SetString(key.Dq, 10)
	res.Dq = dq

	qi := new(big.Int)
	qi, _ = qi.SetString(key.Qinv, 10)
	res.Qinv = qi

	var arr []rsa.CRTValue
	for i := 0; i < len(key.CRTValues); i++ {
		arr = append(arr, CRTValueDecode(&key.CRTValues[i]))
	}
	res.CRTValues = arr

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

//Декодируем приватный ключ из бд в обычный
func PrivateKeyDecode(key *structures.EditedPrivateKey) *rsa.PrivateKey {
	var res rsa.PrivateKey

	res.PublicKey = PublicKeyDecode(&key.PublicKey)

	d := new(big.Int)
	d, _ = d.SetString(key.D, 10)
	res.D = d

	var arr []*big.Int
	for i := 0; i < len(key.Primes); i++ {
		pr := new(big.Int)
		pr, _ = pr.SetString(key.Primes[i], 10)
		arr = append(arr, pr)
	}
	res.Primes = arr

	res.Precomputed = PrecomputedDecode(&key.Precomputed)
	return &res
}
