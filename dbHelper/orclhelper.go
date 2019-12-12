package dbHelper

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-oci8"
	log "github.com/sirupsen/logrus"
)

type OrclDB struct {
	DB *sql.DB
}

func NewOrclDB(openString string) (*OrclDB, error) {

	db, err := sql.Open("oci8", openString)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &OrclDB{
		DB: db,
	}, nil
}

func GetDSN(user string, password string, ip string, port int, serviceName string) string {
	return fmt.Sprintf("%s/%s@%s:%d/%s", user, password, ip, port, serviceName)
}

func (o *OrclDB) Close() {
	o.DB.Close()
}

func (o *OrclDB) FetchAll(query string, args ...interface{}) ([][]interface{}, error) {
	tx, err := o.DB.Begin()
	if err != nil {
		return nil, err
	}
	values, err := o.FetchAllTx(tx, query, args...)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return values, nil
}

func (o *OrclDB) FetchAllTx(tx *sql.Tx, query string, args ...interface{}) ([][]interface{}, error) {
	stmt, err := tx.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	var rows *sql.Rows
	rows, err = stmt.Query(args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// get column infomration
	var columns []string
	columns, err = rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("columns error: %v", err)
	}

	// create values
	values := make([][]interface{}, 0, 1)

	// get values
	pRowInterface := make([]interface{}, len(columns))

	for rows.Next() {
		rowInterface := make([]interface{}, len(columns))
		for i := 0; i < len(rowInterface); i++ {
			pRowInterface[i] = &rowInterface[i]
		}

		err = rows.Err()
		if err != nil {
			return nil, fmt.Errorf("rows error: %v", err)
		}

		err = rows.Scan(pRowInterface...)
		if err != nil {
			return nil, fmt.Errorf("scan error: %v", err)
		}

		values = append(values, rowInterface)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	err = rows.Close()
	if err != nil {
		return nil, fmt.Errorf("close error: %v", err)
	}

	// return values
	return values, nil
}

func (o *OrclDB) ExecTx(tx *sql.Tx, query string, args ...interface{}) (sql.Result, error) {
	stmt, err := tx.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	return stmt.Exec(args...)
}

func (o *OrclDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	tx, err := o.DB.Begin()
	if err != nil {
		return nil, err
	}
	result, err := o.ExecTx(tx, query, args...)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return result, err
}
