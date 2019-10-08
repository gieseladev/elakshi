package edb

import "time"

// DBModel contains common fields shared by most Elakshi models.
type DBModel struct {
	ID        uint64 `gorm:"primary_key"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// EID returns the eid of the model.
func (m DBModel) EID() string {
	return EncodeEID(m.ID)
}
