////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"github.com/pkg/errors"
	"gitlab.com/elixxir/client/v4/bindings"
	"gitlab.com/elixxir/client/v4/storage/utility"
	"gitlab.com/elixxir/crypto/indexedDb"
	"gitlab.com/elixxir/wasm-utils/exception"
	"gitlab.com/elixxir/wasm-utils/utils"
	"sync"
	"syscall/js"
)

// dbCipherTrackerSingleton is used to track DbCipher objects
// so that they can be referenced by ID back over the bindings.
var dbCipherTrackerSingleton = &DbCipherTracker{
	tracked: make(map[int]*DbCipher),
	count:   0,
}

// DbCipherTracker is a singleton used to keep track of extant
// DbCipher objects, preventing race conditions created by passing it
// over the bindings.
type DbCipherTracker struct {
	tracked map[int]*DbCipher
	count   int
	mux     sync.RWMutex
}

// create creates a DbCipher from a [indexedDb.Cipher], assigns it a unique
// ID, and adds it to the DbCipherTracker.
func (ct *DbCipherTracker) create(c indexedDb.Cipher) *DbCipher {
	ct.mux.Lock()
	defer ct.mux.Unlock()

	chID := ct.count
	ct.count++

	ct.tracked[chID] = &DbCipher{
		api: c,
		id:  chID,
	}

	return ct.tracked[chID]
}

// get an DbCipher from the DbCipherTracker given its ID.
func (ct *DbCipherTracker) get(id int) (*DbCipher, error) {
	ct.mux.RLock()
	defer ct.mux.RUnlock()

	c, exist := ct.tracked[id]
	if !exist {
		return nil, errors.Errorf(
			"Cannot get DbCipher for ID %d, does not exist", id)
	}

	return c, nil
}

// delete removes a DbCipherTracker from the DbCipherTracker.
func (ct *DbCipherTracker) delete(id int) {
	ct.mux.Lock()
	defer ct.mux.Unlock()

	delete(ct.tracked, id)
}

// DbCipher wraps the [indexedDb.Cipher] object so its methods
// can be wrapped to be Javascript compatible.
type DbCipher struct {
	api  indexedDb.Cipher
	salt []byte
	id   int
}

// newDbCipherJS creates a new Javascript compatible object
// (map[string]any) that matches the [DbCipher] structure.
func newDbCipherJS(c *DbCipher) map[string]any {
	DbCipherMap := map[string]any{
		"GetID":         js.FuncOf(c.GetID),
		"Encrypt":       js.FuncOf(c.Encrypt),
		"Decrypt":       js.FuncOf(c.Decrypt),
		"MarshalJSON":   js.FuncOf(c.MarshalJSON),
		"UnmarshalJSON": js.FuncOf(c.UnmarshalJSON),
	}

	return DbCipherMap
}

// NewDatabaseCipher constructs a [DbCipher] object.
//
// Parameters:
//   - args[0] - The tracked [Cmix] object ID (int).
//   - args[1] - The password for storage. This should be the same password
//     passed into [NewCmix] (Uint8Array).
//   - args[2] - The maximum size of a payload to be encrypted. A payload passed
//     into [DbCipher.Encrypt] that is larger than this value will result
//     in an error (int).
//
// Returns:
//   - JavaScript representation of the [DbCipher] object.
//   - Throws an error if creating the cipher fails.
func NewDatabaseCipher(_ js.Value, args []js.Value) any {
	cmixId := args[0].Int()
	password := utils.CopyBytesToGo(args[1])
	plaintTextBlockSize := args[2].Int()

	// Get user from singleton
	user, err := bindings.GetCMixInstance(cmixId)
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	// Generate RNG
	stream := user.GetRng().GetStream()

	// Load or generate a salt
	salt, err := utility.NewOrLoadSalt(
		user.GetStorage().GetKV(), stream)
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	// Construct a cipher
	c, err := indexedDb.NewCipher(
		password, salt, plaintTextBlockSize, stream)
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	// Add to singleton and return
	return newDbCipherJS(dbCipherTrackerSingleton.create(c))
}

// GetID returns the ID for this [DbCipher] in the
// DbCipherTracker.
//
// Returns:
//   - Tracker ID (int).
func (c *DbCipher) GetID(js.Value, []js.Value) any {
	return c.id
}

// Encrypt will encrypt the raw data. It will return a ciphertext. Padding is
// done on the plaintext so all encrypted data looks uniform at rest.
//
// Parameters:
//   - args[0] - The data to be encrypted (Uint8Array). This must be smaller
//     than the block size passed into [NewDatabaseCipher]. If it is
//     larger, this will return an error.
//
// Returns:
//   - The ciphertext of the plaintext passed in (String).
//   - Throws an error if it fails to encrypt the plaintext.
func (c *DbCipher) Encrypt(_ js.Value, args []js.Value) any {
	ciphertext, err := c.api.Encrypt(utils.CopyBytesToGo(args[0]))
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	return ciphertext
}

// Decrypt will decrypt the passed in encrypted value. The plaintext will be
// returned by this function. Any padding will be discarded within this
// function.
//
// Parameters:
//   - args[0] - the encrypted data returned by [DbCipher.Encrypt]
//     (String).
//
// Returns:
//   - The plaintext of the ciphertext passed in (Uint8Array).
//   - Throws an error if it fails to encrypt the plaintext.
func (c *DbCipher) Decrypt(_ js.Value, args []js.Value) any {
	plaintext, err := c.api.Decrypt(args[0].String())
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	return utils.CopyBytesToJS(plaintext)
}

// MarshalJSON marshals the cipher into valid JSON.
//
// Returns:
//   - JSON of the cipher (Uint8Array).
//   - Throws an error if marshalling fails.
func (c *DbCipher) MarshalJSON(js.Value, []js.Value) any {
	data, err := c.api.MarshalJSON()
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	return utils.CopyBytesToJS(data)
}

// UnmarshalJSON unmarshalls JSON into the cipher.
//
// Note that this function does not transfer the internal RNG. Use
// [indexedDb.NewCipherFromJSON] to properly reconstruct a cipher from JSON.
//
// Parameters:
//   - args[0] - JSON data to unmarshal (Uint8Array).
//
// Returns:
//   - JSON of the cipher (Uint8Array).
//   - Throws an error if marshalling fails.
func (c *DbCipher) UnmarshalJSON(_ js.Value, args []js.Value) any {
	err := c.api.UnmarshalJSON(utils.CopyBytesToGo(args[0]))
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}
	return nil
}
