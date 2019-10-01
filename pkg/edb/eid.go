package edb

import (
	"encoding/base32"
	"encoding/binary"
	"errors"
)

var noPaddingEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

// EncodeEID encodes an id into an eid.
func EncodeEID(id uint64) string {
	var data = make([]byte, 8)
	binary.LittleEndian.PutUint64(data, id)

	return noPaddingEncoding.EncodeToString(data)
}

var (
	// ErrEIDInvalid is the error returned if an invalid eid is being parsed.
	ErrEIDInvalid = errors.New("eid invalid")
)

// DecodeEID converts the encoded id into its integer representation.
// Returns ErrEIDInvalid if the eid is invalid.
func DecodeEID(eid string) (uint64, error) {
	if noPaddingEncoding.DecodedLen(len(eid)) > 8 {
		return 0, ErrEIDInvalid
	}

	data := make([]byte, 8)
	_, err := noPaddingEncoding.Decode(data, []byte(eid))
	if err != nil {
		return 0, ErrEIDInvalid
	}

	return binary.LittleEndian.Uint64(data), nil
}
