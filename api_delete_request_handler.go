package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
)

func DeleteRequestHandler(w http.ResponseWriter, r *http.Request, db dbHandler) {
	tableNames, err := db.getTableList(w, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	reTableName := regexp.MustCompile(`(?P<tablename>\w+)`)
	matchStrings := reTableName.FindAllString(r.URL.Path, 2)
	if len(matchStrings) > 0 {
		currTableName := matchStrings[0]
		if !contains(tableNames, currTableName) {
			w.WriteHeader(http.StatusNotFound)
			responseJson, _ := json.Marshal(ServerResponse{
				"error": "unknown table",
			})
			w.Write(responseJson)
			return
		}
		currRowId, err := strconv.Atoi(matchStrings[1])
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
