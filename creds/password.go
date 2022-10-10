////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package creds

import (
	"crypto/cipher"
	"encoding/json"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"gitlab.com/xx_network/crypto/csprng"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/chacha20poly1305"
	"io"
	"os"
	"syscall/js"
)

// Data lengths.
const (
	// Length of the internal password (256-bit)
	internalPasswordLen = 32

	// keyLen is the length of the key generated
	keyLen = chacha20poly1305.KeySize

	// saltLen is the length of the salt. Recommended to be 16 bytes here:
	// https://datatracker.ietf.org/doc/html/draft-irtf-cfrg-argon2-04#section-3.1
	saltLen = 16
)

// Storage keys.
const (
	// Key used to store the encrypted internal password salt in local storage.
	saltKey = "xxInternalPasswordSalt"

	// Key used to store the encrypted internal password in local storage.
	passwordKey = "xxEncryptedInternalPassword"

	// Key used to store the argon2 parameters used to encrypted/decrypt the
	// password.
	argonParamsKey = "xxEncryptedInternalPasswordParams"
)

// Error messages.
const (
	// getInternalPassword
	getPasswordStorageErr = "could not retrieve encrypted internal password from storage: %+v"
	getSaltStorageErr     = "could not retrieve salt from storage: %+v"
	getParamsStorageErr   = "could not retrieve encryption parameters from storage: %+v"
	paramsUnmarshalErr    = "failed to unmarshal encryption parameters loaded from storage: %+v"
	decryptPasswordErr    = "could not decrypt internal password: %+v"

	// initInternalPassword
	readInternalPasswordErr     = "could not generate internal password: %+v"
	internalPasswordNumBytesErr = "expected %d bytes for internal password, found %d bytes"

	// decryptPassword
	readNonceLenErr        = "read %d bytes, too short to decrypt"
	decryptWithPasswordErr = "cannot decrypt with password: %+v"

	// makeSalt
	readSaltErr     = "could not generate salt: %+v"
	saltNumBytesErr = "expected %d bytes for salt, found %d bytes"
)

// GetOrInitJS takes a user-provided password and returns its associated 256-bit
// internal password.
//
// If the internal password has not previously been created, then it is
// generated, saved to local storage, and returned. If the internal password has
// been previously generated, it is retrieved from local storage and returned.
//
// Any password saved to local storage is encrypted using the user-provided
// password.
//
// Parameters:
//  - args[0] - The user supplied password (string).
//
// Returns:
//  - Internal password (Uint8Array).
//  - Throws TypeError on failure.
func GetOrInitJS(_ js.Value, args []js.Value) interface{} {
	internalPassword, err := GetOrInit(args[0].String())
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return utils.CopyBytesToJS(internalPassword)
}

// ChangeExternalPasswordJS allows a user to change their external password.
//
// Parameters:
//  - args[0] - The user's old password (string).
//  - args[1] - The user's new password (string).
//
// Returns:
//  - Throws TypeError on failure.
func ChangeExternalPasswordJS(_ js.Value, args []js.Value) interface{} {
	err := ChangeExternalPassword(args[0].String(), args[1].String())
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

// GetOrInit takes a user-provided password and returns its associated 256-bit
// internal password.
//
// If the internal password has not previously been created, then it is
// generated, saved to local storage, and returned. If the internal password has
// been previously generated, it is retrieved from local storage and returned.
//
// Any password saved to local storage is encrypted using the user-provided
// password.
func GetOrInit(externalPassword string) ([]byte, error) {
	localStorage := utils.GetLocalStorage()
	internalPassword, err := getInternalPassword(externalPassword, localStorage)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			rng := csprng.NewSystemRNG()
			return initInternalPassword(
				externalPassword, localStorage, rng, defaultParams())
		}

		return nil, err
	}

	return internalPassword, nil
}

// ChangeExternalPassword allows a user to change their external password.
func ChangeExternalPassword(oldExternalPassword, newExternalPassword string) error {
	localStorage := utils.GetLocalStorage()
	internalPassword, err := getInternalPassword(oldExternalPassword, localStorage)
	if err != nil {
		return err
	}

	salt, err := makeSalt(csprng.NewSystemRNG())
	if err != nil {
		return err
	}
	localStorage.SetItem(saltKey, salt)

	key := deriveKey(newExternalPassword, salt, defaultParams())

	encryptedInternalPassword := encryptPassword(
		internalPassword, key, csprng.NewSystemRNG())
	localStorage.SetItem(passwordKey, encryptedInternalPassword)

	return nil
}

// initInternalPassword generates a new internal password, stores an encrypted
// version in local storage, and returns it.
func initInternalPassword(externalPassword string,
	localStorage *utils.LocalStorage, csprng io.Reader,
	params argonParams) ([]byte, error) {
	internalPassword := make([]byte, internalPasswordLen)

	// Generate internal password
	n, err := csprng.Read(internalPassword)
	if err != nil {
		return nil, errors.Errorf(readInternalPasswordErr, err)
	} else if n != internalPasswordLen {
		return nil, errors.Errorf(
			internalPasswordNumBytesErr, internalPasswordLen, n)
	}

	// Generate and store salt
	salt, err := makeSalt(csprng)
	if err != nil {
		return nil, err
	}
	localStorage.SetItem(saltKey, salt)

	// Store argon2 parameters
	paramsData, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	localStorage.SetItem(argonParamsKey, paramsData)

	key := deriveKey(externalPassword, salt, params)

	encryptedInternalPassword := encryptPassword(internalPassword, key, csprng)
	localStorage.SetItem(passwordKey, encryptedInternalPassword)

	return internalPassword, nil
}

// getInternalPassword retrieves the internal password from local storage,
// decrypts it, and returns it.
func getInternalPassword(
	externalPassword string, localStorage *utils.LocalStorage) ([]byte, error) {
	encryptedInternalPassword, err := localStorage.GetItem(passwordKey)
	if err != nil {
		return nil, errors.WithMessage(err, getPasswordStorageErr)
	}

	salt, err := localStorage.GetItem(saltKey)
	if err != nil {
		return nil, errors.WithMessage(err, getSaltStorageErr)
	}

	paramsData, err := localStorage.GetItem(argonParamsKey)
	if err != nil {
		return nil, errors.WithMessage(err, getParamsStorageErr)
	}

	var params argonParams
	err = json.Unmarshal(paramsData, &params)
	if err != nil {
		return nil, errors.Errorf(paramsUnmarshalErr, err)
	}

	key := deriveKey(externalPassword, salt, params)

	decryptedInternalPassword, err :=
		decryptPassword(encryptedInternalPassword, key)
	if err != nil {
		return nil, errors.Errorf(decryptPasswordErr, err)
	}

	return decryptedInternalPassword, nil
}

// encryptPassword encrypts the data for a shared URL using XChaCha20-Poly1305.
func encryptPassword(data, password []byte, csprng io.Reader) []byte {
	chaCipher := initChaCha20Poly1305(password)
	nonce := make([]byte, chaCipher.NonceSize())
	if _, err := io.ReadFull(csprng, nonce); err != nil {
		jww.FATAL.Panicf("Could not generate nonce %+v", err)
	}
	ciphertext := chaCipher.Seal(nonce, nonce, data, nil)
	return ciphertext
}

// decryptPassword decrypts the encrypted data from a shared URL using
// XChaCha20-Poly1305.
func decryptPassword(data, password []byte) ([]byte, error) {
	chaCipher := initChaCha20Poly1305(password)
	nonceLen := chaCipher.NonceSize()
	if (len(data) - nonceLen) <= 0 {
		return nil, errors.Errorf(readNonceLenErr, len(data))
	}
	nonce, ciphertext := data[:nonceLen], data[nonceLen:]
	plaintext, err := chaCipher.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.Errorf(decryptWithPasswordErr, err)
	}
	return plaintext, nil
}

// initChaCha20Poly1305 returns a XChaCha20-Poly1305 cipher.AEAD that uses the
// given password hashed into a 256-bit key.
func initChaCha20Poly1305(password []byte) cipher.AEAD {
	pwHash := blake2b.Sum256(password)
	chaCipher, err := chacha20poly1305.NewX(pwHash[:])
	if err != nil {
		jww.FATAL.Panicf("Could not init XChaCha20Poly1305 mode: %+v", err)
	}

	return chaCipher
}

// argonParams contains the cost parameters used by Argon2.
type argonParams struct {
	Time    uint32 // Number of passes over the memory
	Memory  uint32 // Amount of memory used in KiB
	Threads uint8  // Number of threads used
}

// defaultParams returns the recommended general purposes parameters.
func defaultParams() argonParams {
	return argonParams{
		Time:    1,
		Memory:  64 * 1024, // ~64 MB
		Threads: 4,
	}
}

// deriveKey derives a key from a user supplied password and a salt via the
// Argon2 algorithm.
func deriveKey(password string, salt []byte, params argonParams) []byte {
	return argon2.IDKey([]byte(password), salt,
		params.Time, params.Memory, params.Threads, keyLen)
}

// makeSalt generates a salt of the correct length of key generation.
func makeSalt(csprng io.Reader) ([]byte, error) {
	b := make([]byte, saltLen)
	size, err := csprng.Read(b)
	if err != nil {
		return nil, errors.Errorf(readSaltErr, err)
	} else if size != saltLen {
		return nil, errors.Errorf(saltNumBytesErr, saltLen, size)
	}

	return b, nil
}
