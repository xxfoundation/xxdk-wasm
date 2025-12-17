////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package storage

import (
	"crypto/cipher"
	"io"
	"io/fs"
	"syscall/js"

	json "github.com/goccy/go-json"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/chacha20poly1305"

	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/crypto/hash"
	utils "gitlab.com/elixxir/xxdk-wasm/jsutil"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb/worker/kv"
	"gitlab.com/xx_network/crypto/csprng"
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

	internalPasswordConstant = "XXInternalPassword"
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
	// initInternalPassword
	readInternalPasswordErr     = "could not generate"
	internalPasswordNumBytesErr = "expected %d bytes for internal password, found %d bytes"

	// getInternalPassword
	getPasswordStorageErr = "could not retrieve encrypted internal password from storage: %+v"
	getSaltStorageErr     = "could not retrieve salt from storage: %+v"
	getParamsStorageErr   = "could not retrieve encryption parameters from storage: %+v"
	paramsUnmarshalErr    = "failed to unmarshal encryption parameters loaded from storage: %+v"
	decryptPasswordErr    = "could not decrypt internal password: %+v"

	// decryptPassword
	readNonceLenErr        = "read %d bytes, too short to decrypt"
	decryptWithPasswordErr = "cannot decrypt with password: %+v"

	// makeSalt
	readSaltErr     = "could not generate salt: %+v"
	saltNumBytesErr = "expected %d bytes for salt, found %d bytes"
)

// GetOrInitPassword takes a user-provided password and returns its associated
// 256-bit internal password.
//
// If the internal password has not previously been created, then it is
// generated, saved to local storage, and returned. If the internal password has
// been previously generated, it is retrieved from local storage and returned.
//
// Any password saved to local storage is encrypted using the user-provided
// password.
//
// Parameters:
//   - args[0] - The user supplied password (string).
//
// Returns a promise:
//   - Resolves to internal password (Uint8Array).
//   - Rejects with an error on failure.
func GetOrInitPassword(_ js.Value, args []js.Value) any {
	// IMPORTANT: Copy args to Go types BEFORE CreatePromise.
	// The goroutine inside CreatePromise runs later when js.Value
	// references may be invalid.
	externalPassword := args[0].String()

	return utils.CreatePromise(func(resolve, reject func(...any) js.Value) {
		internalPassword, err := getOrInit(externalPassword)
		if err != nil {
			errorConstructor := js.Global().Get("Error")
			errorObject := errorConstructor.New(err.Error())
			reject(errorObject)
			return
		}
		resolve(utils.CopyBytesToJS(internalPassword))
	})
}

// ChangeExternalPassword allows a user to change their external password.
//
// Parameters:
//   - args[0] - The user's old password (string).
//   - args[1] - The user's new password (string).
//
// Returns a promise:
//   - Resolves on success.
//   - Rejects with an error on failure.
func ChangeExternalPassword(_ js.Value, args []js.Value) any {
	// IMPORTANT: Copy args to Go types BEFORE CreatePromise.
	oldPassword := args[0].String()
	newPassword := args[1].String()

	return utils.CreatePromise(func(resolve, reject func(...any) js.Value) {
		err := changeExternalPassword(oldPassword, newPassword)
		if err != nil {
			errorConstructor := js.Global().Get("Error")
			errorObject := errorConstructor.New(err.Error())
			reject(errorObject)
			return
		}
		resolve(js.Undefined())
	})
}

// VerifyPassword determines if the user-provided password is correct.
//
// Parameters:
//   - args[0] - The user supplied password (string).
//
// Returns:
//   - True if the password is correct and false if it is incorrect (boolean).
func VerifyPassword(_ js.Value, args []js.Value) any {
	return verifyPassword(args[0].String())
}

// getOrInit is the private function for GetOrInitPassword that is used for
// testing.
func getOrInit(externalPassword string) ([]byte, error) {
	store := kv.GetStore()
	if store == nil {
		return nil, errors.New("KV store not available")
	}

	internalPassword, err := getInternalPassword(externalPassword, store)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			rng := csprng.NewSystemRNG()
			return initInternalPassword(
				externalPassword, store, rng, defaultParams())
		}

		return nil, err
	}

	return internalPassword, nil
}

// changeExternalPassword is the private function for ChangeExternalPassword
// that is used for testing.
func changeExternalPassword(oldExternalPassword, newExternalPassword string) error {
	// NOTE: the following no longer works in synchronized environments, so
	// disabled in produciton.
	jww.FATAL.Panicf("cannot change password, unimplemented")
	store := kv.GetStore()
	if store == nil {
		return errors.New("KV store not available")
	}

	internalPassword, err := getInternalPassword(
		oldExternalPassword, store)
	if err != nil {
		return err
	}

	salt, err := makeSalt(csprng.NewSystemRNG())
	if err != nil {
		return err
	}
	if err = store.Set(saltKey, salt); err != nil {
		return errors.Wrapf(err, "kv: failed to set %q", saltKey)
	}

	key := deriveKey(newExternalPassword, salt, defaultParams())

	encryptedInternalPassword := encryptPassword(
		internalPassword, key, csprng.NewSystemRNG())
	if err = store.Set(passwordKey, encryptedInternalPassword); err != nil {
		return errors.Wrapf(err, "kv: failed to set %q", passwordKey)
	}

	return nil
}

// verifyPassword is the private function for VerifyPassword that is used for
// testing.
func verifyPassword(externalPassword string) bool {
	store := kv.GetStore()
	if store == nil {
		return false
	}
	_, err := getInternalPassword(externalPassword, store)
	return err == nil
}

// initInternalPassword generates a new internal password, stores an encrypted
// version in KV storage, and returns it.
func initInternalPassword(externalPassword string,
	store kv.Store, csprng io.Reader,
	params argonParams) ([]byte, error) {
	internalPassword := make([]byte, internalPasswordLen)

	// FIXME: The internal password is now just an expansion of
	// the users password text. We couldn't preserve the following
	// when doing cross-device sync.
	h := hash.CMixHash.New()
	h.Write([]byte(externalPassword))
	h.Write(internalPassword)
	copy(internalPassword, h.Sum(nil)[:internalPasswordLen])

	// Generate internal password
	// n, err := csprng.Read(internalPassword)
	// if err != nil {
	// 	return nil, errors.Errorf(readInternalPasswordErr, err)
	// } else if n != internalPasswordLen {
	// 	return nil, errors.Errorf(
	// 		internalPasswordNumBytesErr, internalPasswordLen, n)
	// }

	// Generate and store salt
	salt, err := makeSalt(csprng)
	if err != nil {
		return nil, err
	}
	if err = store.Set(saltKey, salt); err != nil {
		return nil,
			errors.Wrapf(err, "kv: failed to set %q", saltKey)
	}

	// Store argon2 parameters
	paramsData, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	if err = store.Set(argonParamsKey, paramsData); err != nil {
		return nil,
			errors.Wrapf(err, "kv: failed to set %q", argonParamsKey)
	}

	key := deriveKey(externalPassword, salt, params)

	encryptedInternalPassword := encryptPassword(internalPassword, key, csprng)
	if err = store.Set(passwordKey, encryptedInternalPassword); err != nil {
		return nil,
			errors.Wrapf(err, "kv: failed to set %q", passwordKey)
	}

	return internalPassword, nil
}

// getInternalPassword retrieves the internal password from KV storage,
// decrypts it, and returns it.
func getInternalPassword(
	externalPassword string, store kv.Store) ([]byte, error) {
	encryptedInternalPassword, err := store.Get(passwordKey)
	if err != nil {
		return nil, errors.WithMessage(err, getPasswordStorageErr)
	}

	salt, err := store.Get(saltKey)
	if err != nil {
		return nil, errors.WithMessage(err, getSaltStorageErr)
	}

	paramsData, err := store.Get(argonParamsKey)
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
// NOTE: Threads is set to 1 for WASM because multi-threaded argon2
// spawns multiple goroutines that can trigger Go WASM GC issues
// ("found pointer to free object") when combined with concurrent
// js.Value operations (like logging to a web worker).
func defaultParams() argonParams {
	return argonParams{
		Time:    1,
		Memory:  64 * 1024, // ~64 MB
		Threads: 1,         // Single-threaded for WASM stability
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
