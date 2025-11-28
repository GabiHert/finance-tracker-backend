package mock

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var once sync.Once
var db *Db

type Db struct {
	DbConn *gorm.DB
	models map[string]any
	schema string
}

// NewDb class is used to configure DB and create a connection pool using gorm
func NewDb(schema string, models map[string]any) *Db {
	if db == nil {
		once.Do(
			func() {
				db = open(schema, models)
			},
		)
	}

	return db
}

func open(schema string, models map[string]any) *Db {
	dbSQL, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		panic(err)
	}

	dbSQL.SetMaxOpenConns(1)

	dbConn, err := gorm.Open(sqlite.Dialector{Conn: dbSQL}, &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: false,
	})
	if err != nil {
		panic("failed to connect to database. err: " + err.Error())
	}

	newDbMock := &Db{
		DbConn: dbConn,
		schema: schema,
		models: models,
	}

	err = newDbMock.ClearDB()
	if err != nil {
		panic(fmt.Sprintf("failed to clear database. err: %s", err.Error()))
	}

	return newDbMock
}

func (d *Db) ClearDB() (err error) {
	retry := true
	retryCount := 0
	for retry {
		retryCount++
		if retryCount > 5 {
			return fmt.Errorf("failed to clear database after 3 attempts")
		}
		if err = d.DbConn.Exec("ATTACH ':memory:' AS " + d.schema).Error; err != nil {
			if !strings.Contains(err.Error(), "is already in use") {
				return err
			}
		} else {
			err = d.init()
			if err != nil {
				continue
			}

			time.Sleep(200 * time.Millisecond)

			_ = d.DbConn.Exec("PRAGMA schema_version").Error

			err = d.checkTables()
			if err != nil {
				continue
			}
		}

		err = d.reset()
		for err != nil {
			continue
		}

		retry = false
	}
	return nil
}

func (d *Db) init() (err error) {
	tx := d.DbConn.Exec("BEGIN EXCLUSIVE")
	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback()
			err = fmt.Errorf("panic occurred while clearing DB: %v", rec)
		} else if err != nil {
			errTx := tx.Exec("ROLLBACK").Error
			if errTx != nil {
				panic(errTx)
			}
		} else {
			errTx := tx.Exec("COMMIT").Error
			if errTx != nil {
				panic(errTx)
			}
		}
	}()

	modelList := make([]any, 0, len(d.models))
	for _, model := range d.models {
		modelList = append(modelList, model)

		stmt := &gorm.Statement{DB: tx}
		if err := stmt.Parse(model); err != nil {
			return err
		}
		tableName := stmt.Schema.Table

		sql := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
		if err := tx.Exec(sql).Error; err != nil {
			return err
		}
	}

	if err := tx.AutoMigrate(modelList...); err != nil {
		return err
	}

	for _, model := range modelList {
		if !tx.Migrator().HasTable(model) {
			return fmt.Errorf("table for model %T was not created", model)
		}
	}

	return nil
}

func (d *Db) reset() (err error) {
	modelList := make([]any, 0, len(d.models))
	for _, model := range d.models {
		modelList = append(modelList, model)

		err = d.DbConn.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(model).Error
		if err != nil {
			return err
		}

		stmt := &gorm.Statement{DB: d.DbConn}
		if err := stmt.Parse(model); err != nil {
			return err
		}
		tableName := stmt.Schema.Table

		err = d.DbConn.Exec("DELETE FROM sqlite_sequence WHERE name = ?", tableName).Error
		if err != nil && !strings.Contains(err.Error(), "no such table: sqlite_sequence") {
			return err
		}
	}

	return nil
}

func (d *Db) checkTables() (err error) {
	modelList := make([]any, 0, len(d.models))
	for _, model := range d.models {
		modelList = append(modelList, model)
	}

	for _, model := range modelList {
		if !d.DbConn.Migrator().HasTable(model) {
			return fmt.Errorf("table for model %T was not created", model)
		}
		if err := d.DbConn.Find(&model).Error; err != nil {
			return fmt.Errorf("failed to query table for model %T: %w", model, err)
		}
	}

	return nil
}

func (d *Db) GetModel(table string) (any, bool) {
	model, ok := d.models[table]
	return model, ok
}
