////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package main

import (
	"encoding/json"
	"github.com/pkg/errors"
	"gitlab.com/elixxir/client/v4/channels"
	cft "gitlab.com/elixxir/client/v4/channelsFileTransfer"
	"gitlab.com/elixxir/crypto/fileTransfer"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb/impl"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"strings"
	"time"
)

// ReceiveFile is called when a file upload or download beings.
//
// fileLink and fileData are nillable and may be updated based
// upon the UUID or file ID later.
//
// fileID is always unique to the fileData. fileLink is the JSON of
// channelsFileTransfer.FileLink.
//
// Returns any fatal errors.
func (w *wasmModel) ReceiveFile(fileID fileTransfer.ID, fileLink,
	fileData []byte, timestamp time.Time, status cft.Status) error {

	newFile := &File{
		Id:        fileID.Marshal(),
		Data:      fileData,
		Link:      fileLink,
		Timestamp: timestamp,
		Status:    uint8(status),
	}
	return w.upsertFile(newFile)
}

// UpdateFile is called when a file upload or download completes or changes.
//
// fileLink, fileData, timestamp, and status are all nillable and may be
// updated based upon the file ID at a later date. If a nil value is passed,
// then make no update.
//
// Returns an error if the file cannot be updated. It must return
// channels.NoMessageErr if the file does not exist.
func (w *wasmModel) UpdateFile(fileID fileTransfer.ID, fileLink,
	fileData []byte, timestamp *time.Time, status *cft.Status) error {
	parentErr := "[Channels indexedDB] failed to UpdateFile"

	// Get the File as it currently exists in storage
	fileObj, err := impl.Get(w.db, fileStoreName, impl.EncodeBytes(fileID.Marshal()))
	if err != nil {
		if strings.Contains(err.Error(), impl.ErrDoesNotExist) {
			return errors.WithMessage(channels.NoMessageErr, parentErr)
		}
		return errors.WithMessage(err, parentErr)
	}
	currentFile, err := valueToFile(fileObj)
	if err != nil {
		return errors.WithMessage(err, parentErr)
	}

	// Update the fields if specified
	if status != nil {
		currentFile.Status = uint8(*status)
	}
	if timestamp != nil {
		currentFile.Timestamp = *timestamp
	}
	if fileData != nil {
		currentFile.Data = fileData
	}
	if fileLink != nil {
		currentFile.Link = fileLink
	}

	return w.upsertFile(currentFile)
}

// upsertFile is a helper function that will update an existing File
// if File.Id is specified. Otherwise, it will perform an insert.
func (w *wasmModel) upsertFile(newFile *File) error {
	newFileJson, err := json.Marshal(&newFile)
	if err != nil {
		return err
	}
	fileObj, err := utils.JsonToJS(newFileJson)
	if err != nil {
		return err
	}

	_, err = impl.Put(w.db, fileStoreName, fileObj)
	return err
}

// GetFile returns the ModelFile containing the file data and download link
// for the given file ID.
//
// Returns an error if the file cannot be retrieved. It must return
// channels.NoMessageErr if the file does not exist.
func (w *wasmModel) GetFile(fileID fileTransfer.ID) (
	cft.ModelFile, error) {
	fileObj, err := impl.Get(w.db, fileStoreName,
		impl.EncodeBytes(fileID.Marshal()))
	if err != nil {
		if strings.Contains(err.Error(), impl.ErrDoesNotExist) {
			return cft.ModelFile{}, channels.NoMessageErr
		}
		return cft.ModelFile{}, err
	}

	resultFile, err := valueToFile(fileObj)
	if err != nil {
		return cft.ModelFile{}, err
	}

	result := cft.ModelFile{
		ID:        fileTransfer.NewID(resultFile.Data),
		Link:      resultFile.Link,
		Data:      resultFile.Data,
		Timestamp: resultFile.Timestamp,
		Status:    cft.Status(resultFile.Status),
	}
	return result, nil
}

// DeleteFile deletes the file with the given file ID.
//
// Returns fatal errors. It must return channels.NoMessageErr if the file
// does not exist.
func (w *wasmModel) DeleteFile(fileID fileTransfer.ID) error {
	err := impl.Delete(w.db, fileStoreName, impl.EncodeBytes(fileID.Marshal()))
	if err != nil {
		if strings.Contains(err.Error(), impl.ErrDoesNotExist) {
			return channels.NoMessageErr
		}
	}
	return err
}
