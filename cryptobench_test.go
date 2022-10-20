////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package main

import (
	"testing"

	"crypto/rand"
	"strconv"

	"gitlab.com/elixxir/crypto/cyclic"
	dh "gitlab.com/elixxir/crypto/diffieHellman"
	"gitlab.com/elixxir/crypto/rsa"
	"gitlab.com/xx_network/crypto/csprng"
	"gitlab.com/xx_network/crypto/large"
)

// //tests Session keys are generated correctly
// func TestGenerateSessionKey(t *testing.T) {
// 	const numTests = 50

// 	primeString := "FFFFFFFFFFFFFFFFC90FDAA22168C234C4C6628B80DC1CD129024E088A67CC74020BBEA63B139B22514A08798E3404DDEF9519B3CD3A431B302B0A6DF25F14374FE1356D6D51C245E485B576625E7EC6F44C42E9A637ED6B0BFF5CB6F406B7EDEE386BFB5A899FA5AE9F24117C4B1FE649286651ECE45B3DC2007CB8A163BF0598DA48361C55D39A69163FA8FD24CF5F83655D23DCA3AD961C62F356208552BB9ED529077096966D670C354E4ABC9804F1746C08CA18217C32905E462E36CE3BE39E772C180E86039B2783A2EC07A28FB5C55DF06F4C52C9DE2BCBF6955817183995497CEA956AE515D2261898FA051015728E5A8AAAC42DAD33170D04507A33A85521ABDF1CBA64ECFB850458DBEF0A8AEA71575D060C7DB3970F85A6E1E4C7ABF5AE8CDB0933D71E8C94E04A25619DCEE3D2261AD2EE6BF12FFA06D98A0864D87602733EC86A64521F2B18177B200CBBE117577A615D6C770988C0BAD946E208E24FA074E5AB3143DB5BFCE0FD108E4B82D120A93AD2CAFFFFFFFFFFFFFFFF"
// 	p := large.NewInt(1)
// 	p.SetString(primeString, 16)
// 	g := large.NewInt(2)
// 	grp := cyclic.NewGroup(p, g)

// 	rng := csprng.NewSystemRNG()

// 	for i := 0; i < numTests; i++ {
// 		//create session key
// 		privKey := dh.GeneratePrivateKey(dh.DefaultPrivateKeyLength, grp, rng)
// 		publicKey := dh.GeneratePublicKey(privKey, grp)
// 		session := dh.GenerateSessionKey(privKey, publicKey, grp)

// 		//create public key manually
// 		sessionExpected := grp.NewInt(1)
// 		grp.Exp(publicKey, privKey, sessionExpected)

// 		if session.Cmp(sessionExpected) != 0 {
// 			t.Errorf("Session key generated on attempt %v incorrect;"+
// 				"\n\tExpected: %s \n\tRecieved: %s \n\tPrivate key: %s", i,
// 				sessionExpected.TextVerbose(16, 0),
// 				session.TextVerbose(16, 0),
// 				privKey.TextVerbose(16, 0))
// 		}
// 	}
// }

// Benchmarks session key creation
func BenchmarkCreateDHSessionKey(b *testing.B) {
	// primeString := "FFFFFFFFFFFFFFFFC90FDAA22168C234C4C6628B80DC1CD129024E088A67CC74020BBEA63B139B22514A08798E3404DDEF9519B3CD3A431B302B0A6DF25F14374FE1356D6D51C245E485B576625E7EC6F44C42E9A637ED6B0BFF5CB6F406B7EDEE386BFB5A899FA5AE9F24117C4B1FE649286651ECE45B3DC2007CB8A163BF0598DA48361C55D39A69163FA8FD24CF5F83655D23DCA3AD961C62F356208552BB9ED529077096966D670C354E4ABC9804F1746C08CA18217C32905E462E36CE3BE39E772C180E86039B2783A2EC07A28FB5C55DF06F4C52C9DE2BCBF6955817183995497CEA956AE515D2261898FA051015728E5A8AAAC42DAD33170D04507A33A85521ABDF1CBA64ECFB850458DBEF0A8AEA71575D060C7DB3970F85A6E1E4C7ABF5AE8CDB0933D71E8C94E04A25619DCEE3D2261AD2EE6BF12FFA06D98A0864D87602733EC86A64521F2B18177B200CBBE117577A615D6C770988C0BAD946E208E24FA074E5AB3143DB5BFCE0FD108E4B82D120A93AD2CAFFFFFFFFFFFFFFFF"
	primeString := "FFFFFFFFFFFFFFFFC90FDAA22168C234C4C6628B80DC1CD129024E088A67CC74020BBEA63B139B22514A08798E3404DDEF9519B3CD3A431B302B0A6DF25F14374FE1356D6D51C245E485B576625E7EC6F44C42E9A637ED6B0BFF5CB6F406B7EDEE386BFB5A899FA5AE9F24117C4B1FE649286651ECE45B3DC2007CB8A163BF0598DA48361C55D39A69163FA8FD24CF5F83655D23DCA3AD961C62F356208552BB9ED529077096966D670C354E4ABC9804F1746C08CA18217C32905E462E36CE3BE39E772C180E86039B2783A2EC07A28FB5C55DF06F4C52C9DE2BCBF6955817183995497CEA956AE515D2261898FA051015728E5A8AAAC42DAD33170D04507A33A85521ABDF1CBA64ECFB850458DBEF0A8AEA71575D060C7DB3970F85A6E1E4C7ABF5AE8CDB0933D71E8C94E04A25619DCEE3D2261AD2EE6BF12FFA06D98A0864D87602733EC86A64521F2B18177B200CBBE117577A615D6C770988C0BAD946E208E24FA074E5AB3143DB5BFCE0FD108E4B82D120A92108011A723C12A787E6D788719A10BDBA5B2699C327186AF4E23C1A946834B6150BDA2583E9CA2AD44CE8DBBBC2DB04DE8EF92E8EFC141FBECAA6287C59474E6BC05D99B2964FA090C3A2233BA186515BE7ED1F612970CEE2D7AFB81BDD762170481CD0069127D5B05AA993B4EA988D8FDDC186FFB7DC90A6C08F4DF435C934063199FFFFFFFFFFFFFFFF"
	p := large.NewInt(1)
	p.SetString(primeString, 16)
	g := large.NewInt(2)
	grp := cyclic.NewGroup(p, g)

	pubkeys := make([]*cyclic.Int, b.N)
	privkeys := make([]*cyclic.Int, b.N)

	rng := csprng.NewSystemRNG()

	for i := 0; i < b.N; i++ {
		// Creation of two different DH Key Pairs with valid parameters
		privkeys[i] = dh.GeneratePrivateKey(dh.DefaultPrivateKeyLength, grp, rng)
		tmpPrivKey := dh.GeneratePrivateKey(dh.DefaultPrivateKeyLength, grp, rng)
		pubkeys[i] = dh.GeneratePublicKey(tmpPrivKey, grp)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dh.GenerateSessionKey(privkeys[i], pubkeys[i], grp)
	}
}

func BenchmarkRSASigCreate(b *testing.B) {
	// Generate keys
	sLocal := rsa.GetScheme()
	rng := rand.Reader

	privKey, err := sLocal.GenerateDefault(rng)
	if err != nil {
		b.Errorf("GenerateDefault: %v", err)
	}

	// pubKey := privKey.Public()

	// Construct signing options
	opts := rsa.NewDefaultPSSOptions()
	hashFunc := opts.HashFunc()

	for i := 0; i < b.N; i++ {
		// Create hash
		h := hashFunc.New()
		h.Write([]byte(strconv.Itoa(i) + "test12345"))
		hashed := h.Sum(nil)

		// Construct signature
		_, err := privKey.SignPSS(rng, hashFunc, hashed, opts)
		if err != nil {
			b.Fatalf("SignPSS error: %+v", err)
		}
	}
}

func BenchmarkRSASigVerify(b *testing.B) {
	// Generate keys
	sLocal := rsa.GetScheme()
	rng := rand.Reader

	privKey, err := sLocal.GenerateDefault(rng)
	if err != nil {
		b.Errorf("GenerateDefault: %v", err)
	}

	pubKey := privKey.Public()

	// Construct signing options
	opts := rsa.NewDefaultPSSOptions()
	hashFunc := opts.HashFunc()
	// Create hash
	h := hashFunc.New()
	h.Write([]byte(strconv.Itoa(0) + "test12345"))
	hashed := h.Sum(nil)

	// Construct signature
	signed, err := privKey.SignPSS(rng, hashFunc, hashed, opts)
	if err != nil {
		b.Fatalf("SignPSS error: %+v", err)
	}

	for i := 0; i < b.N; i++ {
		// Verify signature
		err = pubKey.VerifyPSS(hashFunc, hashed, signed, opts)
		if err != nil {
			b.Fatalf("VerifyPSS error: %+v", err)
		}
	}
}
