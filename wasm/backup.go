////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/elixxir/client/bindings"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"syscall/js"
)

////////////////////////////////////////////////////////////////////////////////
// Structs and Interfaces                                                     //
////////////////////////////////////////////////////////////////////////////////

// Backup wraps the [bindings.Backup] object so its methods can be wrapped to be
// Javascript compatible.
type Backup struct {
	api *bindings.Backup
}

// newBackupJS creates a new Javascript compatible object
// (map[string]interface{}) that matches the Backup structure.
func newBackupJS(api *bindings.Backup) map[string]interface{} {
	b := Backup{api}
	backupMap := map[string]interface{}{
		"StopBackup":      js.FuncOf(b.StopBackup),
		"IsBackupRunning": js.FuncOf(b.IsBackupRunning),
		"AddJson":         js.FuncOf(b.AddJson),
	}

	return backupMap
}

// updateBackupFunc wraps Javascript callbacks to adhere to the
// [bindings.UpdateBackupFunc] interface.
type updateBackupFunc struct {
	updateBackup func(args ...interface{}) js.Value
}

func (ubf *updateBackupFunc) UpdateBackup(encryptedBackup []byte) {
	ubf.updateBackup(utils.CopyBytesToJS(encryptedBackup))
}

////////////////////////////////////////////////////////////////////////////////
// Client functions                                                           //
////////////////////////////////////////////////////////////////////////////////

// NewCmixFromBackup initializes a new e2e storage from an encrypted backup.
// Users of this function should delete the storage directory on error. Users of
// this function should call LoadCmix as normal once this call succeeds.
//
// Parameters:
//  - args[0] - JSON of the NDF (string).
//  - args[1] - storage directory (string).
//  - args[2] - backup passphrase (string).
//  - args[3] - session password (Uint8Array).
//  - args[4] - backup file contents (Uint8Array).
//
// Returns:
//  - JSON of [bindings.BackupReport] (Uint8Array).
//  - throws a TypeError if creating Cmix from backup fails.
func NewCmixFromBackup(_ js.Value, args []js.Value) interface{} {
	ndfJSON := args[0].String()
	storageDir := args[1].String()
	backupPassphrase := args[2].String()
	sessionPassword := utils.CopyBytesToGo(args[3])
	backupFileContents := utils.CopyBytesToGo(args[4])

	report, err := bindings.NewCmixFromBackup(ndfJSON, storageDir,
		backupPassphrase, sessionPassword, backupFileContents)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return utils.CopyBytesToJS(report)
}

////////////////////////////////////////////////////////////////////////////////
// Backup functions                                                           //
////////////////////////////////////////////////////////////////////////////////

// InitializeBackup creates a bindings-layer Backup object.
//
// Parameters:
//  - args[0] - ID of E2e object in tracker (int).
//  - args[1] - ID of UserDiscovery object in tracker (int).
//  - args[2] - backup passphrase provided by the user (string). Used to decrypt
//    backup.
//  - args[3] - the callback to be called when a backup is triggered. Must be
//    Javascript object that has functions that implement the
//    [bindings.UpdateBackupFunc] interface.
//
// Returns:
//  - Javascript representation of the Backup object
//  - Throws a TypeError if initializing the Backup fails.
func InitializeBackup(_ js.Value, args []js.Value) interface{} {
	cb := &updateBackupFunc{utils.WrapCB(args[3], "UpdateBackup")}
	api, err := bindings.InitializeBackup(
		args[0].Int(), args[1].Int(), args[2].String(), cb)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return newBackupJS(api)
}

// ResumeBackup resumes the backup processes with a new callback.
// Call this function only when resuming a backup that has already been
// initialized or to replace the callback.
// To start the backup for the first time or to use a new password, use
// InitializeBackup.
//
// Parameters:
//  - args[0] - ID of E2e object in tracker (int).
//  - args[1] - ID of UserDiscovery object in tracker (int).
//  - args[2] - the callback to be called when a backup is triggered. Must be
//    Javascript object that has functions that implement the
//    [bindings.UpdateBackupFunc] interface. This will replace any callback that
//    has been passed into InitializeBackup.
//
// Returns:
//  - Javascript representation of the Backup object
//  - Throws a TypeError if initializing the Backup fails.
func ResumeBackup(_ js.Value, args []js.Value) interface{} {
	cb := &updateBackupFunc{utils.WrapCB(args[2], "UpdateBackup")}
	api, err := bindings.ResumeBackup(args[0].Int(), args[1].Int(), cb)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return newBackupJS(api)
}

// StopBackup stops the backup processes and deletes the user's password from
// storage. To enable backups again, call InitializeBackup.
//
// Returns:
//  - Throws a TypeError if stopping the backup fails.
func (b *Backup) StopBackup(js.Value, []js.Value) interface{} {
	err := b.api.StopBackup()
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

// IsBackupRunning returns true if the backup has been initialized and is
// running. Returns false if it has been stopped.
//
// Returns:
//  - boolean
func (b *Backup) IsBackupRunning(js.Value, []js.Value) interface{} {
	return b.api.IsBackupRunning()
}

// AddJson stores the argument within the Backup structure.
//
// Parameters:
//  - args[0] - JSON to store (string).
func (b *Backup) AddJson(_ js.Value, args []js.Value) interface{} {
	b.api.AddJson(args[0].String())
	return nil
}
