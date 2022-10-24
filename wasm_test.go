////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

// This file is compiled for all architectures except WebAssembly.
//go:build !js || !wasm

package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
	"unicode"
)

// Tests that all public functions in client/bindings are implemented here in
// the WASM bindings.
func TestPublicFunctions(t *testing.T) {
	// Exclude these functions from the check. These functions are intentionally
	// not implemented.
	excludeList := map[string]struct{}{
		// Notifications are not available in the browser
		"GetNotificationsReport":     {},
		"RegisterForNotifications":   {},
		"UnregisterForNotifications": {},

		// UD not available in the browser
		"IsRegisteredWithUD":     {},
		"NewOrLoadUd":            {},
		"NewUdManagerFromBackup": {},
		"LookupUD":               {},
		"MultiLookupUD":          {},
		"SearchUD":               {},

		// These functions are used internally by the WASM bindings but are not
		// exposed
		"NewEventModel":                   {},
		"NewChannelsManagerGoEventModel":  {},
		"LoadChannelsManagerGoEventModel": {},
		"GetChannelDbCipherTrackerFromID": {},

		// Version functions were renamed to differentiate between WASM and
		// client versions
		"GetGitVersion":   {},
		"GetDependencies": {},
	}
	wasmFuncs := getPublicFunctions("wasm", t)
	bindingsFuncs := getPublicFunctions(
		"vendor/gitlab.com/elixxir/client/bindings", t)

	for fnName := range bindingsFuncs {
		if _, exists := wasmFuncs[fnName]; !exists {
			if _, exists = excludeList[fnName]; !exists {
				t.Errorf("Function %q does not exist in WASM bindings.", fnName)
			} else {
				delete(wasmFuncs, fnName)
			}
		}
	}
}

func getPublicFunctions(pkg string, t testing.TB) map[string]*ast.FuncDecl {
	set := token.NewFileSet()
	packs, err := parser.ParseDir(set, pkg, nil, 0)
	if err != nil {
		t.Fatalf("Failed to parse package: %+v", err)
	}

	funcs := make(map[string]*ast.FuncDecl)
	for _, pack := range packs {
		for _, f := range pack.Files {
			for _, d := range f.Decls {
				if fn, isFn := d.(*ast.FuncDecl); isFn {
					// Exclude type methods and private functions
					if fn.Recv == nil && unicode.IsUpper(rune(fn.Name.Name[0])) {
						funcs[fn.Name.Name] = fn
					}
				}
			}
		}
	}

	return funcs
}
