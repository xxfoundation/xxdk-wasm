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
	"gitlab.com/elixxir/crypto/channel"
	"gitlab.com/elixxir/xxdk-wasm/src/api/utils"
	"gitlab.com/xx_network/crypto/csprng"
	"reflect"
	"syscall/js"
	"testing"
)

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
		_, _ = ConstructIdentity(pi.PubKey, int(pi.CodesetVersion))

		if i%(n/len(identities)) == 0 {
			identities[j] = jsIdentity{pubKey, codeset}
			j++
		}
	}

	b.ResetTimer()
	for i := range identities {
		go func(identity jsIdentity) {
			_, err := ConstructIdentity(
				utils.CopyBytesToGo(identity.pubKey), identity.codeset.Int())
			if err != nil {
				b.Error(err)
			}
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
			_, err := bindings.ConstructIdentity(
				utils.CopyBytesToGo(identity.pubKey), identity.codeset.Int())
			if err != nil {
				b.Error(err)
			}
		}(identities[i])
	}
}
