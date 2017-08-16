package toolkit

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
)

const UUID_SIZE = 16

func NewUUID() UUID {
	var a [UUID_SIZE]byte
	rand.Read(a[:]) // slice convertion
	return a
}

func ToUUID(b []byte) (UUID, error) {
	if len(b) != UUID_SIZE {
		return UUID{}, errors.New("Slice length must be 16.")
	}
	var a [UUID_SIZE]byte
	copy(a[:], b)
	return a, nil
}

type UUID [UUID_SIZE]byte

func (a UUID) String() string {
	return hex.EncodeToString(a[:])
}

func (a UUID) Bytes() []byte {
	//	var b = make([]byte, UUID_SIZE)
	//	copy(b, a[:])
	//	return b
	return a[:]
}

func (a UUID) IsZero() bool {
	for i := 0; i < UUID_SIZE; i++ {
		if a[i] != 0 {
			return false
		}
	}
	return true
}
