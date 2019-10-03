package edb

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAutoMigrate(t *testing.T) {
	db := createDB(t)

	err := AutoMigrate(db)
	assert.NoError(t, err)
}
