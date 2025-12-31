package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/helper/atapi"
	"github.com/kawai-org/kawai-backend/helper/atdb"
	"github.com/kawai-org/kawai-backend/model"
	"go.mongodb.org/mongo-driver/bson"
)

// HandleCron adalah fungsi yang akan dipanggil setiap menit
func HandleCron(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Ambil Profile Bot (Buat dapet Token WA)
	profile, errProf := atdb.GetOneDoc[model.BotProfile](config.Mongoconn, "profile", bson.M{})
	if errProf != nil {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "Profile Bot not found"})
		return
	}

	// 2. Cari Reminder yang 'Pending' DAN Waktunya Sudah Lewat (<= Sekarang)
	now := time.Now()
	filter := bson.M{
		"status":         "pending",
		"scheduled_time": bson.M{"$lte": now},
	}

	cursor, err := config.Mongoconn.Collection("reminders").Find(context.TODO(), filter)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": err.Error()})
		return
	}
	defer cursor.Close(context.TODO())

	var reminders []model.Reminder
	if err = cursor.All(context.TODO(), &reminders); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "Cursor decode error"})
		return
	}

	// 3. Loop dan Kirim Pesan
	count := 0
	for _, rem := range reminders {
		// Format Pesan
		pesan := fmt.Sprintf("⏰ *Waktunya!*\n\nTopik: %s\n\n_Pengingat ini diset untuk: %s_", 
			rem.Title, 
			rem.ScheduledTime.Format("02 Jan 15:04"),
		)

		// Kirim WA
		kirim := model.PushWaSend{
			Token:   profile.Token,
			Target:  rem.UserPhone,
			Type:    "text",
			Delay:   "1",
			Message: pesan,
		}
		
		// 🔥 PERBAIKAN DISINI: Menangkap 3 return value (tambah satu _ lagi)
		_, _, errSend := atapi.PostJSON[interface{}](kirim, profile.URLApi)
		
		// 4. Update Status jadi 'sent' (Biar gak dikirim ulang terus)
		if errSend == nil {
			update := bson.M{"$set": bson.M{"status": "sent"}}
			config.Mongoconn.Collection("reminders").UpdateOne(context.TODO(), bson.M{"_id": rem.ID}, update)
			count++
		}
	}

	// Laporan Selesai
	resp := map[string]interface{}{
		"status":    "success",
		"processed": count,
		"time":      now.Format(time.RFC3339),
	}
	json.NewEncoder(w).Encode(resp)
}