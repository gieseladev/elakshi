package edb

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

type ExternalRef struct {
	ID uint64

	Service    string `gorm:"UNIQUE_INDEX:uix_service_id"`
	Identifier string `gorm:"UNIQUE_INDEX:uix_service_id"`
}

func NewExternalRef(service, id string) ExternalRef {
	return ExternalRef{
		Service:    service,
		Identifier: id,
	}
}

const externalRefSelect = `
SELECT * FROM %s
INNER JOIN %s as jt
	ON jt.%s = %s.id
INNER JOIN external_refs as er
	ON er.id = jt.%s
WHERE er.service = ? AND er.identifier = ?
`

// GetModelByExternalRef returns a scope which selects rows from the given model
// which have the given external reference.
func GetModelByExternalRef(model interface{}, service, identifier string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		mScope := db.NewScope(model)
		quotedTableName := mScope.QuotedTableName()

		joinField, ok := mScope.FieldByName("ExternalReferences")
		if !ok {
			panic("ExternalReferences field not found")
		}

		if joinField.Relationship == nil || joinField.Relationship.Kind != "many_to_many" {
			panic("ExternalReferences not a many2many relation")
		}

		joinTableHandler := joinField.Relationship.JoinTableHandler
		joinTableName := mScope.Quote(joinTableHandler.Table(db))

		sourceKeys := joinTableHandler.SourceForeignKeys()
		if len(sourceKeys) != 1 {
			panic("invalid number of source keys")
		}

		destKeys := joinTableHandler.DestinationForeignKeys()
		if len(destKeys) != 1 {
			panic("invalid number of destination keys")
		}

		joinDestKey := mScope.Quote(destKeys[0].DBName)
		joinSourceKey := mScope.Quote(sourceKeys[0].DBName)

		query := fmt.Sprintf(externalRefSelect,
			quotedTableName, joinTableName, joinSourceKey, quotedTableName, joinDestKey)
		return db.Raw(query, service, identifier)
	}
}
