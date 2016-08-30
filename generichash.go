package sodium

// #cgo pkg-config: libsodium
// #include <stdlib.h>
// #include <sodium.h>
import "C"
import "fmt"

var (
	cryptoGenericHashBytesMin    = int(C.crypto_generichash_bytes_min())
	cryptoGenericHashBytesMax    = int(C.crypto_generichash_bytes_max())
	cryptoGenericHashKeyBytesMin = int(C.crypto_generichash_keybytes_min())
	cryptoGenericHashKeyBytesMax = int(C.crypto_generichash_keybytes_max())
	cryptoGenericHashBytes       = int(C.crypto_generichash_bytes())
	cryptoGenericHashKeyBytes    = int(C.crypto_generichash_keybytes())
	cryptoGenericHashPrimitive   = C.GoString(C.crypto_generichash_primitive())
	cryptoGenericHashStateBytes  = int(C.crypto_generichash_statebytes())
)

//GenericHash provides a BLAKE2b (RFC7693) hash, in interface of crypto/hash.Hash.
//
//The Hash's and key's size can be any between 16 bytes (128 bits) to
// 64 bytes (512 bits) based on different application.
type GenericHash struct {
	size      int
	blocksize int
	key       *GenericHashKey
	sum       []byte
	state     *C.struct_crypto_generichash_blake2b_state
}

type GenericHashKey struct {
	Bytes
}

func (GenericHashKey) Size() int {
	return cryptoGenericHashKeyBytes
}

//Unkeyed version with default output length.
func NewGenericHashDefault() GenericHash {
	return NewGenericHash(cryptoGenericHashBytes)
}

//Keyed version with default output length.
func NewGenericHashDefaultKeyed(key GenericHashKey) GenericHash {
	return NewGenericHashKeyed(cryptoGenericHashBytes, key)
}

//Unkeyed version, output length should between 16 (128-bit) to 64 (512-bit).
func NewGenericHash(outlen int) GenericHash {
	checkSizeInRange(outlen, cryptoGenericHashBytesMin, cryptoGenericHashBytesMax, "out")
	state := new(C.struct_crypto_generichash_blake2b_state)
	hash := GenericHash{
		size:      outlen,
		blocksize: 128,
		key:       nil,
		sum:       nil,
		state:     state,
	}
	hash.Reset()
	return hash
}

//Keyed version, output length in bytes should between 16 (128-bit) to 64 (512-bit).
func NewGenericHashKeyed(outlen int, key GenericHashKey) GenericHash {
	checkSizeInRange(outlen, cryptoGenericHashBytesMin, cryptoGenericHashBytesMax, "out")
	state := new(C.struct_crypto_generichash_blake2b_state)
	checkTypedSize(&key, "generic hash key")
	hash := GenericHash{
		size:      outlen,
		blocksize: 128,
		key:       &key,
		sum:       nil,
		state:     state,
	}
	hash.Reset()
	return hash
}

//Output length in bytes.
//
//Implements crypto/hash.Hash
func (g GenericHash) Size() int {
	return g.size
}

//Implements crypto/hash.Hash
func (g GenericHash) BlockSize() int {
	return g.blocksize
}

//Implements crypto/hash.Hash
func (g *GenericHash) Reset() {
	if g.key != nil {
		if int(C.crypto_generichash_init(
			g.state,
			(*C.uchar)(&g.key.Bytes[0]),
			(C.size_t)(g.key.Length()),
			(C.size_t)(g.size))) != 0 {
			panic("see libsodium")
		}
	} else {
		if int(C.crypto_generichash_init(
			g.state,
			(*C.uchar)(nil),
			(C.size_t)(0),
			(C.size_t)(g.size))) != 0 {
			panic("see libsodium")
		}
	}
}

//Use GenericHash.Write([]byte) to hash chunks of message.
//
//Implements crypto/hash.Hash
func (g *GenericHash) Write(p []byte) (n int, err error) {
	if g.sum != nil {
		return 0, fmt.Errorf("hash finalized")
	}
	i := p[:]

	for len(i) > g.blocksize {
		c := i[:g.blocksize]
		i = i[g.blocksize:]
		if int(C.crypto_generichash_update(
			g.state,
			(*C.uchar)(&c[0]),
			(C.ulonglong)(g.blocksize))) != 0 {
			panic("see libsodium")
		}
	}
	if len(i) > 0 {
		if int(C.crypto_generichash_update(
			g.state,
			(*C.uchar)(&i[0]),
			(C.ulonglong)(len(i)))) != 0 {
			panic("see libsodium")
		}
	}
	return len(p), nil
}

//Return appended the Sum before b.
//
//Implements crypto/hash.Hash.
//NOTE: Repeated call is allowed. But can't call Write() after Sum().
// Sum() will change the underlying state.
// It is not consistent with the definition of crypto/hash.Hash.
func (g *GenericHash) Sum(b []byte) []byte {
	if g.sum != nil {
		return append(g.sum, b...)
	}
	g.sum = make([]byte, g.size)
	if int(C.crypto_generichash_final(
		g.state,
		(*C.uchar)(&g.sum[0]),
		(C.size_t)(g.size))) != 0 {
		panic("see libsodium")
	}
	return append(g.sum, b...)
}
