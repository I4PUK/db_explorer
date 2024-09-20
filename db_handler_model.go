package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
)

type dbHandler struct {
	DB *sql.DB
}

// получаем список всех таблиц
func (h *dbHandler) getTableList(w http.ResponseWriter, r *http.Request) ([]string, error) {
	// ответ
	var tableNames []string

	// запрос к базе
	rows, err := h.DB.Query("SHOW TABLES")

	// закрываем запрос, чтобы не было течи
	defer rows.Close()

	if err != nil {
		return tableNames, err
	}
	// итерируемся по строкам
	for rows.Next() {
		var tableName string

		// читаем ячейку в переменную
		err := rows.Scan(&tableName)

		// если ошибка - возвращаем ее
		if err != nil {
			return tableNames, err
		}

		// добавляем к ответу
		tableNames = append(tableNames, tableName)
	}

	return tableNames, nil
}

func (h *dbHandler) getPrimaryColumnName(tableName string) (string, error) {
	rows, err := h.DB.Query(fmt.Sprintf(`SHOW FULL COLUMNS FROM %s`, tableName))
	defer rows.Close()

	if err != nil {
		return "", err
	}
	for rows.Next() {
		var (
			Field      string
			Type       interface{}
			Collation  interface{}
			Null       interface{}
			Key        string
			Default    interface{}
			Extra      interface{}
			Privileges interface{}
			Comment    interface{}
		)
		if err := rows.Scan(&Field, &Type, &Collation, &Null, &Key, &Default, &Extra, &Privileges, &Comment); err != nil {
			return "", err
		}
		if Key != "PRI" {
			continue
		}
		return Field, nil
	}

	return "", err
}

func (h *dbHandler) getRowDetail(tableName string, rowId int) (ServerResponse, error) {
	columnNamePK, err := h.getPrimaryColumnName(tableName)
	if err != nil {
		return nil, err
	}
	var records []ServerResponse
	rows, err := h.DB.Query(fmt.Sprintf(`SELECT * FROM %[1]s WHERE %[2]s = ?`, tableName, columnNamePK), rowId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	items := make([]interface{}, len(columns))
	itemsRelated := make([]interface{}, len(columns))
	for rows.Next() {
		record := ServerResponse{}
		for i := range columns {
			items[i] = &itemsRelated[i]
		}
		if err := rows.Scan(items...); err != nil {
			return nil, err
		}
		for i, column := range columns {
			if itemsRelated[i] == nil {
				record[column] = nil
				continue
			}
			intValue64, ok := itemsRelated[i].(int64)
			if ok {
				record[column] = intValue64
				continue
			}
			intValue32, ok := itemsRelated[i].(int32)
			if ok {
				record[column] = intValue32
				continue
			}
			floatValue64, ok := itemsRelated[i].(float64)
			if ok {
				record[column] = floatValue64
				continue
			}
			floatValue32, ok := itemsRelated[i].(float32)
			if ok {
				record[column] = floatValue32
				continue
			}
			byteValue, ok := itemsRelated[i].([]byte)
			if ok {
				record[column] = string(byteValue)
				continue
			}
		}
		records = append(records, record)
	}
	if len(records) == 0 {
		return nil, nil
	}
	return records[0], nil
}

func (h *dbHandler) updateRows(tableName string, bodyMap map[string]interface{}, columnNamePK string, currRowId int) (int, error) {
	columnNames := make([]string, len(bodyMap))
	columnValues := make([]interface{}, len(bodyMap))

	for key, value := range bodyMap {
		columnNames = append(columnNames, fmt.Sprintf(`%s=?`, key))
		columnValues = append(columnValues, value)
	}
	columnValues = append(columnValues, currRowId)

	result, err := h.DB.Exec(fmt.Sprintf(`UPDATE %[1]s SET %[2]s WHERE %[3]s=? ;`,
		tableName,
		strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(fmt.Sprintf("%s", columnNames), "[", ""), "]", ""), " ", ","),
		columnNamePK), columnValues...)

	if err != nil {
		return 0, err
	}
	lastAffected, err := result.RowsAffected()

	if err != nil {
		return 0, err
	}

	return int(lastAffected), nil
}

func (h *dbHandler) getTypesForColumns(tableName string) (map[string]string, error) {
	rows, err := h.DB.Query(fmt.Sprintf(`SHOW FULL COLUMNS FROM %s`, tableName))
	result := make(map[string]string)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			Field      string
			Type       string
			Collation  interface{}
			Null       string
			Key        interface{}
			Default    interface{}
			Extra      interface{}
			Privileges interface{}
			Comment    interface{}
		)
		if err := rows.Scan(&Field, &Type, &Collation, &Null, &Key, &Default, &Extra, &Privileges, &Comment); err != nil {
			return nil, err
		}
		result[Field] = fmt.Sprintf(`%[1]s,%[2]s`, Type, Null)

	}
	return result, err
}

func (h *dbHandler) createRow(tableName string, bodyMap map[string]interface{}) (int, error) {
	columnNames := make([]string, 0, len(bodyMap))
	columnValues := make([]interface{}, 0, len(bodyMap))
	questionMark := make([]string, 0, len(bodyMap))

	for key, value := range bodyMap {
		columnNames = append(columnNames, key)
		columnValues = append(columnValues, value)
		questionMark = append(questionMark, "?")
	}

	result, err := h.DB.Exec(fmt.Sprintf(`INSERT INTO %[1]s (%[2]s) VALUES(%[3]s);`,
		tableName,
		strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(fmt.Sprintf("%s", columnNames), "[", ""), "]", ""), " ", ","),
		strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(fmt.Sprintf("%s", questionMark), "[", ""), "]", ""), " ", ","),
	), columnValues...)

	if err != nil {
		return 0, err
	}
	lastInsertId, err := result.LastInsertId()

	if err != nil {
		return 0, err
	}

	return int(lastInsertId), nil
}

func (h *dbHandler) deleteRow(tableName string, columnNamePK string, currRowId int) (int, error) {
	result, err := h.DB.Exec(fmt.Sprintf(`DELETE FROM %[1]s WHERE %[2]s=? ;`,
		tableName, columnNamePK), currRowId)
	if err != nil {
		return 0, err
	}
	lastAffectedId, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(lastAffectedId), nil
}
