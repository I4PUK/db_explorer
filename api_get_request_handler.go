package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func GetRequestHandler(w http.ResponseWriter, r *http.Request, db dbHandler, tableNames []string) {
	parametersLen := strings.Count(r.URL.Path, "/")

	if r.URL.Path == "/" {
		getTableListHandler(w, r, db, tableNames)
	} else {
		switch parametersLen {
		case 1:
			getTableWithParameters(w, r, db, tableNames)
		case 2:
			getTableById(w, r, db, tableNames)
		}
	}
}

func getTableListHandler(w http.ResponseWriter, r *http.Request, db dbHandler, tableNames []string) {
	w.WriteHeader(http.StatusOK)
	responseJson, _ := json.Marshal(ServerResponse{
		"response": ServerResponse{
			"tables": tableNames,
		},
	})
	w.Write(responseJson)
}

func getTableWithParameters(w http.ResponseWriter, r *http.Request, db dbHandler, tableNames []string) {
	tableName := strings.ReplaceAll(r.URL.Path, "/", "")
	offset, offsetErr := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, limitErr := strconv.Atoi(r.URL.Query().Get("limit"))

	if offsetErr != nil {
		offset = 0
	}
	if limitErr != nil {
		limit = 5
	}

	if contains(tableNames, tableName) == false {
		w.WriteHeader(http.StatusNotFound)
		responseJson, _ := json.Marshal(ServerResponse{
			"error": "unknown table",
		})
		w.Write(responseJson)
		return
	}
	SRList, err := getRowsInTableByOffsetAndLimit(db, tableName, offset, limit)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	responseJson, _ := json.Marshal(ServerResponse{
		"response": ServerResponse{
			"records": SRList,
		},
	})
	w.Write(responseJson)
	return

}

func getRowsInTableByOffsetAndLimit(db dbHandler, tableName string, offset, limit int) ([]ServerResponse, error) {
	var records []ServerResponse
	rows, err := db.DB.Query(fmt.Sprintf(`SELECT * FROM %s LIMIT ? OFFSET ?`, tableName), limit, offset)

	defer rows.Close()
	if err != nil {
		return records, err
	}

	columns, err := rows.Columns()
	if err != nil {
		return records, err
	}
	items := make([]interface{}, len(columns))
	relatedItems := make([]interface{}, len(columns))

	for rows.Next() {
		record := ServerResponse{}
		for i := range columns {
			items[i] = &relatedItems[i]
		}
		if err := rows.Scan(items...); err != nil {
			return records, err
		}
		for i, column := range columns {
			if relatedItems[i] == nil {
				record[column] = nil
				continue
			}
			intValue64, ok := relatedItems[i].(int64)
			if ok {
				record[column] = intValue64
				continue
			}
			intValue32, ok := relatedItems[i].(int32)
			if ok {
				record[column] = intValue32
				continue
			}
			byteValue, ok := relatedItems[i].([]byte)
			if ok {
				record[column] = string(byteValue)
				continue
			}
			floatValue64, ok := relatedItems[i].(float64)
			if ok {
				record[column] = floatValue64
				continue
			}
			floatValue32, ok := relatedItems[i].(float32)
			if ok {
				record[column] = floatValue32
				continue
			}
		}
		records = append(records, record)
	}
	return records, nil
}

func getTableById(w http.ResponseWriter, r *http.Request, db dbHandler, tableNames []string) {
	urlParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(urlParts) <= 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	currTableName := urlParts[0]
	if !contains(tableNames, currTableName) {
		w.WriteHeader(http.StatusNotFound)
		responseJson, _ := json.Marshal(ServerResponse{
			"error": "unknown table",
		})
		w.Write(responseJson)
		return
	}
	currRowId, err := strconv.Atoi(urlParts[1])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	SRRow, err := db.getRowDetail(currTableName, currRowId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if SRRow == nil {
		w.WriteHeader(http.StatusNotFound)
		responseJson, _ := json.Marshal(ServerResponse{
			"error": "record not found",
		})
		w.Write(responseJson)
		return
	}
	w.WriteHeader(http.StatusOK)
	responseJson, _ := json.Marshal(ServerResponse{
		"response": ServerResponse{
			"record": SRRow,
		},
	})
	w.Write(responseJson)
	return
}
