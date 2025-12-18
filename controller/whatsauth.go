package controller

import (
	"encoding/json"
	"net/http"

	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/helper/atdb"
	"github.com/kawai-org/kawai-backend/model"
)

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func GetHome(respw http.ResponseWriter, req *http.Request) {
	resp := model.Response{Response: "It works! Kawai Assistant is Online."}
	WriteJSON(respw, http.StatusOK, resp)
}

func PostInboxNomor(respw http.ResponseWriter, req *http.Request) {
	var resp model.Response
	var msg model.IteungMessage

	err := json.NewDecoder(req.Body).Decode(&msg)
	if err != nil {
		resp.Response = "Error Decode: " + err.Error()
		WriteJSON(respw, http.StatusBadRequest, resp)
		return
	}

	// Pastikan config.Mongoconn sudah di-assign di main.go
	if config.Mongoconn != nil {
		_, err = atdb.InsertOneDoc(config.Mongoconn, "inbox", msg)
		if err != nil {
			resp.Response = "Error DB: " + err.Error()
		}
	}

	resp.Response = "Pesan diterima oleh Kawai"
	WriteJSON(respw, http.StatusOK, resp)
}

func GetNewToken(respw http.ResponseWriter, req *http.Request) {
	resp := model.Response{Response: "Feature Refresh Token coming soon"}
	WriteJSON(respw, http.StatusOK, resp)
}

func NotFound(respw http.ResponseWriter, req *http.Request) {
	resp := model.Response{Response: "Error 404: Path Not Found"}
	WriteJSON(respw, http.StatusNotFound, resp)
}