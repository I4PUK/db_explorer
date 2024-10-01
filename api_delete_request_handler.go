package main

import (
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
		sendResponse(w, http.StatusOK, ServerResponse{
			"response": ServerResponse{
				"deleted": countDeleted,
			},
		})
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	return
}
