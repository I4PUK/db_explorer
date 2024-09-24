package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func PutRequestHandler(w http.ResponseWriter, r *http.Request, db dbHandler) {
	tableNames, err := db.getTableList(w, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	urlParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(urlParts) > 0 {
		currTableName := urlParts[0]
		if !contains(tableNames, currTableName) {
			w.WriteHeader(http.StatusNotFound)
			responseJson, _ := json.Marshal(ServerResponse{
				"error": "unknown table",
			})
			w.Write(responseJson)
			return
		}
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
			w.WriteHeader(http.StatusBadRequest)
			responseJson, _ := json.Marshal(ServerResponse{
				"error": fmt.Sprintf(`"field %s have invalid type"`, invalidField),
			})
			w.Write(responseJson)
			return
		}
		removeUnknownFields(&bodyMap, typesForColumns)
		addFieldToCreate(&bodyMap, typesForColumns)
		idCreated, err := db.createRow(currTableName, bodyMap)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		responseJson, _ := json.Marshal(ServerResponse{
			"response": ServerResponse{
				fmt.Sprintf(`%s`, columnNamePK): idCreated,
			},
		})
		w.Write(responseJson)
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
