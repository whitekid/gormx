package gormx

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	"github.com/whitekid/iter"
	"gorm.io/gorm"
)

var (
	errNotFound       = errors.New("not found")
	errMultipleResult = errors.New("multiple results")
)

// ModelService generic model CRUD operations
// M: database model struct, shuuld embeding gorm.Model
// R: user model struct
type ModelService[T any] struct {
	DB *gorm.DB
}

func NewModelService[T any](db *gorm.DB) *ModelService[T] {
	return &ModelService[T]{
		DB: db,
	}
}

func (s *ModelService[T]) Save(db *gorm.DB, m *T) (*T, error) {
	if db == nil {
		db = s.DB
	}

	if tx := db.Save(m); tx.Error != nil {
		return nil, tx.Error
	}

	return m, nil
}

// TODO update in batches
func (s *ModelService[T]) CreateInBatches(db *gorm.DB, m ...*T) ([]*T, error) {
	if db == nil {
		db = s.DB
	}

	if tx := db.CreateInBatches(m, 100); tx.Error != nil {
		return nil, tx.Error
	}

	return m, nil
}

// TODO delete in batches
func (s *ModelService[T]) Delete(db *gorm.DB, where *T) error {
	if db == nil {
		db = s.DB
	}

	tx := db
	if where != nil {
		tx = tx.Where(where)
	}

	tx = tx.Delete(new(T))
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

type ListOpt struct {
	Order    string
	MaxRowID int64
	Offset   int
	Count    int
}

type ListResult[T any] struct {
	Items iter.Iterator[*T]
	Count int64 // Rows Affected

	MaxRowID int64
	Offset   int
}

func TableName[T any](db *gorm.DB) string {
	var m T
	return db.NamingStrategy.TableName(reflect.TypeOf(&m).Elem().Name())
}

// listModel
// pagenation이 가능하게 만드는 것이었는데...
func (s *ModelService[T]) listModel(db *gorm.DB, where *T, opts ListOpt) (*ListResult[T], error) {
	if db == nil {
		db = s.DB
	}

	tableName := db.NamingStrategy.TableName(reflect.TypeOf(new(T)).Elem().Name())

	// psql은 rowid가 없음
	if opts.Count > 0 && opts.MaxRowID == 0 {
		var max sql.NullInt64

		if err := s.DB.Model(new(T)).
			Select(fmt.Sprintf("MAX(`%s`.rowid)", tableName)).
			Row().Scan(&max); err != nil {
			return nil, err
		}
		opts.MaxRowID = max.Int64
	}

	if opts.MaxRowID > 0 {
		db = db.Where(fmt.Sprintf("`%s`.rowid <= ?", tableName), opts.MaxRowID)
	}
	if opts.Count > 0 {
		db = db.Offset(opts.Offset).Limit(opts.Count)
	}

	if where != nil {
		db = db.Where(where)
	}

	if opts.Order != "" {
		db = db.Order(opts.Order)
	}

	results := []*T{}
	db = db.Find(&results)
	if db.Error != nil {
		return nil, db.Error
	}

	return &ListResult[T]{
		Items:    iter.S(results),
		Count:    db.RowsAffected,
		MaxRowID: opts.MaxRowID,
		Offset:   opts.Offset + int(db.RowsAffected),
	}, nil
}

func (s *ModelService[T]) List(db *gorm.DB, where *T, opts ListOpt) (*ListResult[T], error) {
	r, err := s.listModel(db, where, opts)
	if err != nil {
		return nil, err
	}

	return &ListResult[T]{
		Items:    r.Items,
		Count:    r.Count,
		MaxRowID: r.MaxRowID,
		Offset:   r.Offset,
	}, nil

}

func (s *ModelService[T]) getModel(db *gorm.DB, where *T) (*T, error) {
	it, err := s.listModel(db, where, ListOpt{})
	if err != nil {
		return nil, err
	}

	results := it.Items.Slice()

	switch len(results) {
	case 0:
		return nil, errNotFound
	case 1:
		return results[0], nil
	default:
		return nil, errMultipleResult
	}
}

func (s *ModelService[T]) Get(db *gorm.DB, where *T) (*T, error) {
	r, err := s.getModel(db, where)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func IsNotFound(err error) bool       { return errors.Is(err, errNotFound) }
func IsMultipleRecord(err error) bool { return errors.Is(err, errMultipleResult) }

func List[T any](tx *gorm.DB) ([]*T, error) {
	r := []*T{}
	if tx := tx.Find(&r); tx.Error != nil {
		return nil, tx.Error
	}

	return r, nil
}

func Get[T any](tx *gorm.DB) (*T, error) {
	r, err := List[T](tx)
	if err != nil {
		return nil, err
	}

	switch len(r) {
	case 0:
		return nil, errNotFound
	case 1:
		return r[0], nil
	default:
		return nil, errMultipleResult
	}
}
