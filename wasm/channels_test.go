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
	"gitlab.com/elixxir/crypto/channel"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"gitlab.com/xx_network/crypto/csprng"
	"reflect"
	"syscall/js"
	"testing"
)

// Tests that the map representing ChannelsManager returned by
// newChannelsManagerJS contains all of the methods on ChannelsManager.
func Test_newChannelsManagerJS(t *testing.T) {
	cmType := reflect.TypeOf(&ChannelsManager{})

	e2e := newChannelsManagerJS(&bindings.ChannelsManager{})
	if len(e2e) != cmType.NumMethod() {
		t.Errorf("ChannelsManager JS object does not have all methods."+
			"\nexpected: %d\nreceived: %d", cmType.NumMethod(), len(e2e))
	}

	for i := 0; i < cmType.NumMethod(); i++ {
		method := cmType.Method(i)

		if _, exists := e2e[method.Name]; !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}

// Tests that ChannelsManager has all the methods that
// [bindings.ChannelsManager] has.
func Test_ChannelsManagerMethods(t *testing.T) {
	cmType := reflect.TypeOf(&ChannelsManager{})
	binCmType := reflect.TypeOf(&bindings.ChannelsManager{})

	if binCmType.NumMethod() != cmType.NumMethod() {
		t.Errorf("WASM ChannelsManager object does not have all methods from "+
			"bindings.\nexpected: %d\nreceived: %d",
			binCmType.NumMethod(), cmType.NumMethod())
	}

	for i := 0; i < binCmType.NumMethod(); i++ {
		method := binCmType.Method(i)

		if _, exists := cmType.MethodByName(method.Name); !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}

type jsIdentity struct {
	pubKey  js.Value
	codeset js.Value
}

// Benchmark times the ConstructIdentity, which uses a sync.Map to increase
// efficiency for previously generated identities.
func BenchmarkConstructIdentity(b *testing.B) {
	const n = 100_000
	identities, j := make([]jsIdentity, 1000), 0
	for i := 0; i < n; i++ {
		pi, err := channel.GenerateIdentity(csprng.NewSystemRNG())
		if err != nil {
			b.Fatalf("%+v", err)
		}

		pubKey := utils.CopyBytesToJS(pi.PubKey)
		codeset := js.ValueOf(int(pi.CodesetVersion))
		ConstructIdentity(js.Value{}, []js.Value{pubKey, codeset})

		if i%(n/len(identities)) == 0 {
			identities[j] = jsIdentity{pubKey, codeset}
			j++
		}
	}

	b.ResetTimer()
	for i := range identities {
		go func(identity jsIdentity) {
			ConstructIdentity(
				js.Value{}, []js.Value{identity.pubKey, identity.codeset})
		}(identities[i])
	}
}

// Benchmark times the constructIdentity, which generates each new identity.
func Benchmark_constructIdentity(b *testing.B) {
	identities := make([]jsIdentity, b.N)
	for i := range identities {
		pi, err := channel.GenerateIdentity(csprng.NewSystemRNG())
		if err != nil {
			b.Fatalf("%+v", err)
		}

		pubKey := utils.CopyBytesToJS(pi.PubKey)
		codeset := js.ValueOf(int(pi.CodesetVersion))
		identities[i] = jsIdentity{pubKey, codeset}
	}

	b.ResetTimer()
	for i := range identities {
		go func(identity jsIdentity) {
			constructIdentity(
				js.Value{}, []js.Value{identity.pubKey, identity.codeset})
		}(identities[i])
	}
}
