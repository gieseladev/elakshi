package api

import (
	"encoding/base32"
	"encoding/binary"
	"errors"
	"github.com/gieseladev/elakshi/pkg/edb"
	"github.com/jinzhu/gorm"
)

var (
	ErrEIDInvalid  = errors.New("eid invalid")
	ErrEIDNotFound = errors.New("eid was not found")
)

var NoPaddingEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

// EncodeEID encodes an id into an eid.
func EncodeEID(id uint64) string {
	var data = make([]byte, 8)
	binary.LittleEndian.PutUint64(data, id)

	return NoPaddingEncoding.EncodeToString(data)
}

// DecodeEID converts the encoded id into its integer representation.
// Returns ErrEIDInvalid if the eid is invalid.
func DecodeEID(eid string) (uint64, error) {
	data, err := NoPaddingEncoding.DecodeString(eid)
	if err != nil {
		return 0, ErrEIDInvalid
	}

	return binary.LittleEndian.Uint64(data), nil
}

func GetTrack(db *gorm.DB, eid string) (edb.Track, error) {
	trackID, err := DecodeEID(eid)
	if err != nil {
		return edb.Track{}, err
	}

	ok, track := edb.GetTrack(db, trackID)
	if !ok {
		return track, ErrEIDNotFound
	}

	return track, nil
}
