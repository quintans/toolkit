package app

import (
	"github.com/quintans/goSQL/db"
	"github.com/quintans/goSQL/dbx"
	coll "github.com/quintans/toolkit/collection"
	. "github.com/quintans/toolkit/ext"
	"github.com/quintans/toolkit/log"

	"strings"
)

const NOT_DELETED int64 = 0

var logger = log.LoggerFor("pqp/toolkit/app")

// the result of the query is put in the passed struct.
// returns true if a result was found, false if no result
func FindById(DB db.IDb, table *db.Table, instance interface{}, id int64) (bool, error) {
	logger.CallerAt(1).Debugf("DAOUtils.FindById: %v", id)

	keyColumn := table.GetKeyColumns().Enumerator().Next().(*db.Column)

	return DB.Query(table).
		All().
		Where(keyColumn.Matches(id)).
		SelectTo(instance)
}

func FindAll(DB db.IDb, table *db.Table, instance interface{}) (coll.Collection, error) {
	logger.CallerAt(1).Debugf("DAOUtils.FetchAll")

	deletion := table.GetDeletionColumn()

	q := DB.Query(table).All()
	if deletion != nil {
		q.Where(deletion.Matches(NOT_DELETED))
	}

	return q.ListOf(instance)
}

func FindAllWithDeleted(DB db.IDb, table *db.Table, instance interface{}) (coll.Collection, error) {
	logger.CallerAt(1).Debugf("DAOUtils.FindAllWithDeleted")

	return DB.Query(table).All().ListOf(instance)
}

func Save(DB db.IDb, table *db.Table, entity IEntity) error {
	logger.CallerAt(1).Debugf("DAOUtils.Save: %s", entity)

	if entity.GetVersion() == nil {
		entity.SetVersion(Int64Ptr(1))
		entity.SetCreation(NOW())
		id, err := DB.Insert(table).Submit(entity)
		if err != nil {
			return err
		}
		entity.SetId(&id)
	} else {
		entity.SetModification(NOW())
		_, err := DB.Update(table).Submit(entity)
		if err != nil {
			return err
		}
	}
	return nil
}

func DeleteById(DB db.IDb, table *db.Table, id int64) (bool, error) {
	logger.CallerAt(1).Debugf("DAOUtils.deleteById: %v", id)

	keyColumn := table.GetKeyColumns().Enumerator().Next().(*db.Column)

	result, err := DB.Delete(table).
		Where(keyColumn.Matches(id)).
		Execute()

	return result != 0, err
}

func DeleteByIdAndVersion(DB db.IDb, table *db.Table, id int64, version int64) error {
	logger.CallerAt(1).Debugf("DAOUtils.removeByIdVersion: id: %v, version: %v", id, version)

	keyColumn := table.GetKeyColumns().Enumerator().Next().(*db.Column)
	versionColumn := table.GetVersionColumn()

	result, err := DB.Delete(table).
		Where(
		keyColumn.Matches(id),
		versionColumn.Matches(version),
	).
		Execute()

	if err != nil {
		return err
	}

	if result == 0 {
		return dbx.NewOptimisticLockFail("DAOUtils.removeByIdVersion: Unable to delete by id and version for the table " + table.GetName())
	}

	return nil
}

func Delete(DB db.IDb, table *db.Table, entity IEntity) error {
	return DeleteByIdAndVersion(DB, table, *entity.GetId(), *entity.GetVersion())
}

func SoftDeleteByIdAndVersion(DB db.IDb, table *db.Table, id int64, version int64) error {
	deletion := table.GetDeletionColumn()
	if deletion == nil {
		return dbx.NewPersistenceFail("DAOUtils.SoftDeleteByIdAndVersion", "Table "+table.GetName()+" does not have a deletion type column to do a soft delete.")
	}

	keyColumn := table.GetKeyColumns().Enumerator().Next().(*db.Column)
	versionColumn := table.GetVersionColumn()

	result, err := DB.Update(table).
		Set(deletion, nil).
		Set(versionColumn, version+1).
		Where(
		keyColumn.Matches(id),
		versionColumn.Matches(version),
	).
		Execute()

	if err != nil {
		return err
	}

	if result == 0 {
		return dbx.NewOptimisticLockFail("DAOUtils.removeByIdVersion: Unable to soft delete by id and version for the table " + table.GetName())
	}

	return nil
}

func SoftDelete(DB db.IDb, table *db.Table, entity IEntity) error {
	return SoftDeleteByIdAndVersion(DB, table, *entity.GetId(), *entity.GetVersion())
}

func AddEqualCriteria(conditions []*db.Criteria, column *db.Column, value interface{}) ([]*db.Criteria, bool) {
	if !IsNil(value) {
		conditions = append(conditions, column.Matches(value))
		return conditions, true
	}
	return conditions, false
}

func HasWildcards(value *string) bool {
	if !IsEmpty(value) {
		return strings.ContainsAny(*value, "%")
	}
	return false
}

func AddCriteria(conditions []*db.Criteria, column *db.Column, value *string) ([]*db.Criteria, bool) {
	if !IsEmpty(value) {
		var criteria *db.Criteria
		if HasWildcards(value) {
			criteria = column.Like(value)
		} else {
			criteria = column.Matches(value)
		}
		conditions = append(conditions, criteria)
		return conditions, true
	}
	return conditions, false
}

func AddWildCriteria(conditions []*db.Criteria, column *db.Column, value *string) ([]*db.Criteria, bool) {
	if !IsEmpty(value) {
		conditions = append(conditions, column.Like("%"+*value+"%"))
		return conditions, true
	}
	return conditions, false
}

func AddWildNoCaseCriteria(conditions []*db.Criteria, column *db.Column, value *string) ([]*db.Criteria, bool) {
	if !IsEmpty(value) {
		conditions = append(conditions, column.ILike("%"+*value+"%"))
		return conditions, true
	}
	return conditions, false
}

func AddNoCaseCriteria(conditions []*db.Criteria, column *db.Column, value *string) ([]*db.Criteria, bool) {
	if !IsEmpty(value) {
		var criteria *db.Criteria
		if HasWildcards(value) {
			criteria = column.ILike(value)
		} else {
			criteria = column.IMatches(value)
		}
		conditions = append(conditions, criteria)
		return conditions, true
	}
	return conditions, false
}

func AddRangeCriteria(conditions []*db.Criteria, column *db.Column, leftBound interface{}, rightBound interface{}) ([]*db.Criteria, bool) {
	if !IsNil(leftBound) || !IsNil(rightBound) {
		conditions = append(conditions, column.Range(leftBound, rightBound))
		return conditions, true
	}
	return conditions, false
}

func QueryForPage(
	query *db.Query,
	criteria Criteria,
	instance interface{},
	transformer func(interface{}) (interface{}, error),
) (Page, error) {
	max := criteria.PageSize
	first := (criteria.Page - 1) * max
	// for the first page the offset is zero
	query.Skip(first)
	if max > 0 {
		query.Limit(max + 1)
	}

	entities, err := query.ListFlatTreeOf(instance)
	if err != nil {
		return Page{}, err
	}

	results := entities.Elements()
	if transformer != nil {
		for k, v := range results {
			results[k], err = transformer(v)
			if err != nil {
				return Page{}, err
			}
		}
	}

	page := Page{}
	size := int64(entities.Size())
	if max > 0 && size > max {
		page.Last = false
		page.Results = results[:max]
	} else {
		page.Last = true
		page.Results = results
	}

	// count records
	if criteria.CountRecords {
		DB := query.GetDb()
		cnt := DB.Query(query.GetTable())
		cnt.Copy(query)
		cnt.ColumnsReset()
		cnt.Column(db.Count(nil))
		cnt.OrdersReset()
		var recs int64
		_, err = cnt.SelectInto(&recs)
		if err != nil {
			return Page{}, err
		}
		page.Count = Int64Ptr(recs)
	}

	return page, nil
}
