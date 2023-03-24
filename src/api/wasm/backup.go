////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/elixxir/client/v4/bindings"
)

////////////////////////////////////////////////////////////////////////////////
// Structs and Interfaces                                                     //
////////////////////////////////////////////////////////////////////////////////

// Backup wraps the [bindings.Backup] object so its methods can be wrapped to be
// Javascript compatible.
type Backup struct {
	*bindings.Backup
}

// UpdateBackupFunc wraps a Javascript object so that it implements the
// [bindings.UpdateBackupFunc] interface.
type UpdateBackupFunc struct {
	UpdateBackupFn func(encryptedBackup []byte) `wasm:"UpdateBackup"`
}

// UpdateBackup is a function callback that returns new backups.
//
// Parameters:
//   - encryptedBackup - Returns the bytes of the encrypted backup (Uint8Array).
func (ubf *UpdateBackupFunc) UpdateBackup(encryptedBackup []byte) {
	ubf.UpdateBackupFn(encryptedBackup)
}

////////////////////////////////////////////////////////////////////////////////
// Backup functions                                                           //
////////////////////////////////////////////////////////////////////////////////

// InitializeBackup creates a bindings-layer [Backup] object.
//
// Parameters:
//   - e2eID - ID of [E2e] object in tracker.
//   - udID - ID of [UserDiscovery] object in tracker.
//   - backupPassPhrase - [Backup] passphrase provided by the user. Used to
//     decrypt the backup.
//   - cb - The callback to be called when a backup is triggered. Javascript
//     object that matches the [UpdateBackupFunc] struct.
func InitializeBackup(e2eID, udID int, backupPassPhrase string,
	cb *UpdateBackupFunc) (*Backup, error) {
	api, err := bindings.InitializeBackup(e2eID, udID, backupPassPhrase, cb)
	return &Backup{api}, err
}

// ResumeBackup resumes the backup processes with a new callback.
// Call this function only when resuming a backup that has already been
// initialized or to replace the callback.
// To start the backup for the first time or to use a new password, use
// [InitializeBackup].
//
// Parameters:
//   - e2eID - ID of [E2e] object in tracker.
//   - udID - ID of [UserDiscovery] object in tracker.
//   - cb - The callback to be called when a backup is triggered. Javascript
//     object that matches the [UpdateBackupFunc] struct. This will replace any
//     callback that has been passed into [InitializeBackup].
func ResumeBackup(e2eID, udID int, cb *UpdateBackupFunc) (*Backup, error) {
	api, err := bindings.ResumeBackup(e2eID, udID, cb)
	return &Backup{api}, err
}
