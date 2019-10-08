package edb

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

// TODO use multiple external reference tables for each entity.
//		This solves some rather unfortunate many-to-many issues and should
//		also speed-up the query.

type ExternalRef struct {
	ID uint64

	Service    string `gorm:"UNIQUE_INDEX:uix_reference_id"`
	Identifier string `gorm:"UNIQUE_INDEX:uix_reference_id"`
}

func NewExternalRef(service, id string) ExternalRef {
	return ExternalRef{
		Service:    service,
		Identifier: id,
	}
}

const externalRefSelect = `
INNER JOIN %s as jt
	ON jt.%s = %s.id
INNER JOIN external_refs as er
	ON er.id = jt.%s
`

// JoinModelExternalRef joins the ExternalRef rows under the alias "er".
// The join table is also available under the alias "jt".
// Assumes that the model stores the external references in an
// "ExternalReferences" field.
func JoinModelExternalRef(model interface{}) func(*gorm.DB) *gorm.DB {
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

		query := fmt.Sprintf(externalRefSelect, joinTableName, joinSourceKey, quotedTableName, joinDestKey)
		return db.Joins(query)
	}
}

func commaOkGORMErr(err error) (bool, error) {
	if gorm.IsRecordNotFoundError(err) {
		return false, nil
	}

	return err == nil, err
}

// GetModelByExternalRef searches a model out based on an external reference.
func GetModelByExternalRef(db *gorm.DB, service, identifier string, out interface{}) (bool, error) {
	err := db.Scopes(JoinModelExternalRef(out)).
		Where("er.service = ? AND er.identifier = ?", service, identifier).
		Take(out).Error

	return commaOkGORMErr(err)
}

// GetModelByExternalRefs returns the first model found using one of the
// references.
func GetModelByExternalRefs(db *gorm.DB, out interface{}, refs []ExternalRef) (bool, error) {
	db = db.Scopes(JoinModelExternalRef(out))

	for _, ref := range refs {
		err := db.Take(out, "er.service = ? AND er.identifier = ?", ref.Service, ref.Identifier).Error
		if err == nil {
			return true, nil
		} else if gorm.IsRecordNotFoundError(err) {
			continue
		} else {
			return false, err
		}
	}

	return false, nil
}
