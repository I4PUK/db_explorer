package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func PutRequestHandler(w http.ResponseWriter, r *http.Request, db dbHandler, tableNames []string) {
	urlParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(urlParts) > 0 {
		currTableName := urlParts[0]
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var bodyMap = make(map[string]interface{})
		err = json.Unmarshal(body, &bodyMap)
		if err != nil {
			panic(err)
		}
		columnNamePK, err := db.getPrimaryColumnName(currTableName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		delete(bodyMap, columnNamePK)
		typesForColumns, err := db.getTypesForColumns(currTableName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		invalidField, err := findInvalidTypeField(bodyMap, typesForColumns, columnNamePK)
		if err != nil {
			sendResponse(w, http.StatusBadRequest, ServerResponse{
				"error": fmt.Sprintf(`"field %s have invalid type"`, invalidField),
			})
			return
		}
		removeUnknownFields(&bodyMap, typesForColumns)
		addFieldToCreate(&bodyMap, typesForColumns)
		idCreated, err := db.createRow(currTableName, bodyMap)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		sendResponse(w, http.StatusOK, ServerResponse{
			"response": ServerResponse{
				fmt.Sprintf(`%s`, columnNamePK): idCreated,
			},
		})
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	return
}

func addFieldToCreate(bodyMap *map[string]interface{}, typesForColumns map[string]string) {
	existFields := make([]string, len(*bodyMap))
	for key := range *bodyMap {
		existFields = append(existFields, key)
	}
	for key, value := range typesForColumns {
		if !contains(existFields, key) && strings.Contains(value, "NO") {
			if strings.Contains(value, "text") || strings.Contains(value, "varchar") {
				(*bodyMap)[key] = ""
			} else {
				(*bodyMap)[key] = 0
			}
		}
	}
}
