package espresso

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/crypto"
)

type Commitment [32]byte

type RawCommitmentBuilder struct {
	hasher crypto.KeccakState
}

func NewRawCommitmentBuilder(name string) *RawCommitmentBuilder {
	b := new(RawCommitmentBuilder)
	b.hasher = crypto.NewKeccakState()
	return b.ConstantString(name)
}

// Append a constant string to the running hash.
//
// WARNING: The string `s` must be a constant. This function does not encode the length of `s` in
// the hash, which can lead to domain collisions when different strings with different lengths are
// used depending on the input object.
func (b *RawCommitmentBuilder) ConstantString(s string) *RawCommitmentBuilder {
	if _, err := io.WriteString(b.hasher, s); err != nil {
		panic(fmt.Sprintf("KeccakState Writer is not supposed to fail, but it did: %v", err))
	}

	// To denote the end of the string and act as a domain separator, include a byte sequence which
	// can never appear in a valid UTF-8 string.
	invalidUtf8 := []byte{0xC0, 0x7F}
	return b.FixedSizeBytes(invalidUtf8)
}

// Include a named field of another committable type.
func (b *RawCommitmentBuilder) Field(f string, c Commitment) *RawCommitmentBuilder {
	return b.ConstantString(f).FixedSizeBytes(c[:])
}

// Include a named field of type `uint256` in the hash.
func (b *RawCommitmentBuilder) Uint256Field(f string, n *U256) *RawCommitmentBuilder {
	return b.ConstantString(f).Uint256(n)
}

// Include a value of type `uint256` in the hash.
func (b *RawCommitmentBuilder) Uint256(n *U256) *RawCommitmentBuilder {
	bytes := make([]byte, 32)
	n.FillBytes(bytes)

	// `FillBytes` uses big endian byte ordering, but the Espresso commitment scheme uses little
	// endian, so we need to reverse the bytes.
	for i, j := 0, len(bytes)-1; i < j; i, j = i+1, j-1 {
		bytes[i], bytes[j] = bytes[j], bytes[i]
	}

	return b.FixedSizeBytes(bytes)
}

// Include a named field of type `uint64` in the hash.
func (b *RawCommitmentBuilder) Uint64Field(f string, n uint64) *RawCommitmentBuilder {
	return b.ConstantString(f).Uint64(n)
}

// Include a value of type `uint64` in the hash.
func (b *RawCommitmentBuilder) Uint64(n uint64) *RawCommitmentBuilder {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, n)
	return b.FixedSizeBytes(bytes)
}

// Append a fixed size byte array to the running hash.
//
// WARNING: Go's type system cannot express the requirement that `bytes` is a fixed size array of
// any size. The best we can do is take a dynamically sized slice. However, this function uses a
// fixed-size encoding; namely, it does not encode the length of `bytes` in the hash, which can lead
// to domain collisions when this function is called with a slice which can have different lengths
// depending on the input object.
//
// The caller must ensure that this function is only used with slices whose length is statically
// determined by the type being committed to.
func (b *RawCommitmentBuilder) FixedSizeBytes(bytes Bytes) *RawCommitmentBuilder {
	b.hasher.Write(bytes)
	return b
}

// Include a named field of dynamic length in the hash.
func (b *RawCommitmentBuilder) VarSizeField(f string, bytes Bytes) *RawCommitmentBuilder {
	return b.ConstantString(f).VarSizeBytes(bytes)
}

// Include a byte array whose length can be dynamic to the running hash.
func (b *RawCommitmentBuilder) VarSizeBytes(bytes Bytes) *RawCommitmentBuilder {
	// First commit to the length, to prevent length extension and domain collision attacks.
	b.Uint64(uint64(len(bytes)))
	b.hasher.Write(bytes)
	return b
}

func (b *RawCommitmentBuilder) Finalize() Commitment {
	var comm Commitment
	bytes := b.hasher.Sum(nil)
	copy(comm[:], bytes)
	return comm
}
