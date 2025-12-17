////////////////////////////////////////////////////////////////////////////////
// Copyright © 2024 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package kv

import (
	"github.com/Max-Sum/base32768"
	json "github.com/goccy/go-json"
	"github.com/pkg/errors"
)

// EncodedStore wraps a StringStore and provides a Store (bytes) interface
// by encoding/decoding values using base32768.
//
// base32768 is optimized for UTF-16 (JavaScript's native string encoding)
// and achieves 93.75% efficiency (15 bits per 16-bit character).
type EncodedStore struct {
	raw StringStore
}

// NewEncodedStore creates a Store that wraps a StringStore with base32768 encoding.
func NewEncodedStore(raw StringStore) Store {
	return &EncodedStore{raw: raw}
}

// Get implements Store.Get by decoding the base32768 string value to bytes.
func (s *EncodedStore) Get(key string) ([]byte, error) {
	value, err := s.raw.Get(key)
	if err != nil {
		return nil, err
	}

	// Decode from base32768
	decoded, err := base32768.SafeEncoding.DecodeString(value)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode value for key %q", key)
	}

	return decoded, nil
}

// Set implements Store.Set by encoding the bytes to base32768 string.
func (s *EncodedStore) Set(key string, value []byte) error {
	// Encode to base32768
	encoded := base32768.SafeEncoding.EncodeToString(value)
	return s.raw.Set(key, encoded)
}

// Delete implements Store.Delete.
func (s *EncodedStore) Delete(key string) error {
	return s.raw.Delete(key)
}

// Keys implements Store.Keys by returning JSON-encoded array of key names.
func (s *EncodedStore) Keys() ([]byte, error) {
	keys, err := s.raw.Keys()
	if err != nil {
		return nil, err
	}

	// Convert to JSON for backwards compatibility
	jsonKeys, err := json.Marshal(keys)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal keys to JSON")
	}

	return jsonKeys, nil
}
