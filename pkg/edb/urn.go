package edb

import (
	"errors"
	"fmt"
	"strings"
)

const (
	elakshiNID = "elakshi"
)

// ElakshiURN represents a URN for an Elakshi entity.
type ElakshiURN struct {
	Namespace string
	EID       string
	id        uint64
}

// NewElakshiURN builds an ElakshiURN from the namespace and eid.
func NewElakshiURN(namespace, eid string) ElakshiURN {
	return ElakshiURN{
		Namespace: namespace,
		EID:       eid,
	}
}

func (u ElakshiURN) URN() string {
	return fmt.Sprintf("%s:%s:%s", elakshiNID, u.Namespace, u.EID)
}

func (u ElakshiURN) String() string {
	return u.URN()
}

func (u ElakshiURN) DecodeEID() (uint64, error) {
	if u.id != 0 {
		return u.id, nil
	}

	id, err := DecodeEID(u.EID)
	u.id = id

	return id, err
}

var (
	ErrInvalidURN = errors.New("invalid elakshi urn")
	ErrInvalidNID = errors.New("urn has invalid nid")
)

// ParseURN parses a urn string into an ElakshiURN string.
// Note that it only accepts Elakshi urns.
func ParseURN(u string) (ElakshiURN, error) {
	parts := strings.SplitN(u, ":", 3)
	if len(parts) != 3 {
		return ElakshiURN{}, ErrInvalidURN
	}

	nid := parts[0]
	if nid != elakshiNID {
		return ElakshiURN{}, ErrInvalidNID
	}

	namespace, eid := parts[1], parts[2]
	return NewElakshiURN(namespace, eid), nil
}

type URNPartsProvider interface {
	EID() string
	Namespace() string
}

// URNFromParts creates an ElakshiURN from a URNPartsProvider.
func URNFromParts(parts URNPartsProvider) ElakshiURN {
	return NewElakshiURN(parts.Namespace(), parts.EID())
}
