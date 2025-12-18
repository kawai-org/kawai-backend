package controller

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/kawai-org/kawai-backend/helper/atdb"
	"github.com/kawai-org/kawai-backend/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func PostInboxNomor(w http.ResponseWriter, r *http.Request) {
	var msg model.IteungMessage
	var resp model.Response

	// Set header konten selalu JSON
	w.Header().Set("Content-Type", "application/json")

	// 1. Decode pesan masuk
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp.Response = "Error: Format pesan tidak valid"
		json.NewEncoder(w).Encode(resp)
		return
	}

	// 2. Normalisasi pesan
	pesan := strings.ToLower(strings.TrimSpace(msg.Message))

	// 3. Logika Perintah: Simpan Catatan
	if strings.HasPrefix(pesan, "simpan") || strings.HasPrefix(pesan, "catat") {
		content := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(pesan, "simpan"), "catat"))

		if content == "" {
			resp.Response = "Waduh, isi catatannya kosong. Coba: simpan [isi catatan]"
		} else {
			newNote := model.Note{
				ID:        primitive.NewObjectID(),
				UserPhone: msg.Phone,
				Title:     "Catatan dari WhatsApp",
				Content:   content,
				UpdatedAt: time.Now(),
			}

			_, err := atdb.InsertNote(newNote)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				resp.Response = "Gagal menyimpan ke database."
			} else {
				resp.Response = "Siap! Catatan kamu sudah disimpan aman di Kawai. ✨"
			}
		}
	} else if pesan == "halo" || pesan == "hi" {
		resp.Response = "Halo! Aku Kawai Assistant. Ketik 'simpan [sesuatu]' untuk mencatat ya!"
	} else {
		resp.Response = "Aku belum paham perintah itu. Coba ketik 'simpan' diikuti catatanmu."
	}

	// 4. Kirim Respon Sukses
	json.NewEncoder(w).Encode(resp)
}

func GetNewToken(respw http.ResponseWriter, req *http.Request) {
	resp := model.Response{Response: "Feature Refresh Token coming soon"}
	WriteJSON(respw, http.StatusOK, resp)
}

func NotFound(respw http.ResponseWriter, req *http.Request) {
	resp := model.Response{Response: "Error 404: Path Not Found"}
	WriteJSON(respw, http.StatusNotFound, resp)
}