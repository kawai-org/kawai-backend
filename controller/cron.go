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

	// Setup Timezone WIB untuk konversi tampilan
	loc, _ := time.LoadLocation("Asia/Jakarta")

	profile, errProf := atdb.GetOneDoc[model.BotProfile](config.Mongoconn, "profile", bson.M{})
	if errProf != nil {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "Profile Bot not found"})
		return
	}

	// Cari Reminder yang 'Pending' DAN Waktunya Sudah Lewat
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

	count := 0
	for _, rem := range reminders {
		// 🔥 PERBAIKAN TIMEZONE 🔥
		// Konversi waktu dari DB (UTC) ke WIB sebelum ditampilkan
		waktuWIB := rem.ScheduledTime.In(loc)

		pesan := fmt.Sprintf("⏰ *Waktunya!*\n\n📌 Topik: %s\n⏳ Waktu: %s\n\n_Pengingat ini diset untuk: %s_", 
			rem.Title,
			waktuWIB.Format("15:04 WIB"), // Tampilkan Jam saja biar ringkas
			waktuWIB.Format("02 Jan 2006, 15:04 WIB"), // Tampilkan lengkap di bawah
		)

		kirim := model.PushWaSend{
			Token:   profile.Token,
			Target:  rem.UserPhone,
			Type:    "text",
			Delay:   "1",
			Message: pesan,
		}
		
		_, _, errSend := atapi.PostJSON[interface{}](kirim, profile.URLApi)
		
		if errSend == nil {
			update := bson.M{"$set": bson.M{"status": "sent"}}
			config.Mongoconn.Collection("reminders").UpdateOne(context.TODO(), bson.M{"_id": rem.ID}, update)
			count++
		}
	}

	resp := map[string]interface{}{
		"status":    "success",
		"processed": count,
		"server_time": now.Format(time.RFC3339),
	}
	json.NewEncoder(w).Encode(resp)
}