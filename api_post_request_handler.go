package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func PostRequestHandler(w http.ResponseWriter, r *http.Request, db dbHandler, tableNames []string) {
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
		currRowId, err := strconv.Atoi(urlParts[1])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
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
		typesForColumns, err := db.getTypesForColumns(currTableName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		invalidField, err := findInvalidTypeField(bodyMap, typesForColumns, columnNamePK)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			responseJson, _ := json.Marshal(ServerResponse{
				"error": fmt.Sprintf(`field %s have invalid type`, invalidField),
			})
			w.Write(responseJson)
			return
		}
		removeUnknownFields(&bodyMap, typesForColumns)
		countUpdated, err := db.updateRows(currTableName, bodyMap, columnNamePK, currRowId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		responseJson, _ := json.Marshal(ServerResponse{
			"response": ServerResponse{
				"updated": countUpdated,
			},
		})
		w.Write(responseJson)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	return
}

func findInvalidTypeField(bodyMap map[string]interface{}, typesForColumns map[string]string, columnNamePK string) (string, error) {
	for bodyKey, bodyValue := range bodyMap {
		typeValueFromDB := typesForColumns[bodyKey]
		_, okByte := bodyValue.(string)
		_, okInt32 := bodyValue.(int32)
		_, okInt64 := bodyValue.(int64)
		_, okFloat32 := bodyValue.(float32)
		_, okFloat64 := bodyValue.(float64)
		if bodyKey == columnNamePK || (bodyValue == nil && !strings.Contains(typeValueFromDB, "YES")) ||
			(strings.Contains(typeValueFromDB, "int") && !(okInt32 || okInt64)) ||
			(strings.Contains(typeValueFromDB, "float") && !(okFloat32 || okFloat64)) ||
			((strings.Contains(typeValueFromDB, "text") || strings.Contains(typeValueFromDB, "varchar")) && !okByte && bodyValue != nil) {
			return bodyKey, errors.New("invalid type field")

		}
	}
	return "", nil
}

func removeUnknownFields(bodyMap *map[string]interface{}, typesForColumns map[string]string) {
	existFields := make([]string, len(typesForColumns))
	for key := range typesForColumns {
		existFields = append(existFields, key)
	}
	for key := range *bodyMap {
		if !contains(existFields, key) {
			delete(*bodyMap, key)
		}
	}
}
