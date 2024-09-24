package main

import (
	"database/sql"
	"fmt"
	"net/http"
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

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func handler(w http.ResponseWriter, r *http.Request, db dbHandler) {
	fmt.Printf("Received %s %s with: %s \n", r.Method, r.URL.Path, r.URL.RawQuery)

	tableNames, err := db.getTableList(w, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		GetRequestHandler(w, r, db, tableNames)
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

func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	dbHandler := dbHandler{DB: db}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, dbHandler)
	})

	return mux, nil
}
