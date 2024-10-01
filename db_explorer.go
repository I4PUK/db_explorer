package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

//GET /$table?limit=5&offset=7 - возвращает список из 5 записей (limit) начиная с 7-й (offset) из таблицы $table. limit по-умолчанию 5, offset 0
//GET /$table/$id - возвращает информацию о самой записи или 404
//PUT /$table - создаёт новую запись, данный по записи в теле запроса (POST-параметры)
//POST /$table/$id - обновляет запись, данные приходят в теле запроса (POST-параметры)
//DELETE /$table/$id - удаляет запись
//GET, PUT, POST, DELETE - это http-метод, которым был отправлен запрос

type ServerResponse map[string]interface{}

type Request struct {
	Table    *string
	RecordId *int
}

func sendResponse(w http.ResponseWriter, statusCode int, content ServerResponse) {
	w.WriteHeader(statusCode)
	responseJson, _ := json.Marshal(content)
	w.Write(responseJson)
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func handler(w http.ResponseWriter, r *http.Request, db dbHandler) {
	//fmt.Printf("Received %s %s with: %s \n", r.Method, r.URL.Path, r.URL.RawQuery)

	tableNames, err := db.getTableList(w, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	req := getUrlPartsFromRequest(r)

	if *req.Table != "" {
		if contains(tableNames, *req.Table) == false {
			sendResponse(w, http.StatusNotFound, ServerResponse{"error": "unknown table"})
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		if *req.Table == "" {
			sendResponse(w, http.StatusOK, ServerResponse{
				"response": ServerResponse{
					"tables": tableNames,
				},
			})
			return
		}
		if req.RecordId == nil {
			getTableWithParameters(w, r, db)
		} else {
			getTableById(w, r, db, tableNames)
		}

	case http.MethodPost:
		PostRequestHandler(w, r, db, tableNames)
		return
	case http.MethodPut:
		PutRequestHandler(w, r, db, tableNames)
		return
	case http.MethodDelete:
		DeleteRequestHandler(w, r, db, tableNames)
		return
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getUrlPartsFromRequest(r *http.Request) Request {
	urlParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	var req Request
	if len(urlParts) >= 1 {
		req.Table = &urlParts[0]
	}

	if len(urlParts) >= 2 {
		if id, err := strconv.Atoi(urlParts[1]); err == nil {
			req.RecordId = &id
		}
	}

	return req
}

func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	dbHandler := dbHandler{DB: db}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, dbHandler)
	})

	return mux, nil
}
