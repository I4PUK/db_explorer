package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func DeleteRequestHandler(w http.ResponseWriter, r *http.Request, db dbHandler, tableNames []string) {
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
		currRowId, err := strconv.Atoi(urlParts[1])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		columnNamePK, err := db.getPrimaryColumnName(currTableName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		countDeleted, err := db.deleteRow(currTableName, columnNamePK, currRowId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		responseJson, _ := json.Marshal(ServerResponse{
			"response": ServerResponse{
				"deleted": countDeleted,
			},
		})
		w.Write(responseJson)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	return
}
