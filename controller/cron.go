package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

		// --- TAMBAHAN: FILTER NOMOR ---
        // Jika nomor tidak berawalan "62" (atau kode negara lain yang valid), 
        // anggap itu LID/Invalid dan JANGAN KIRIM.
        if !strings.HasPrefix(rem.UserPhone, "62") {
            fmt.Printf("⚠️ Skip Reminder ke ID Laptop/Invalid: %s (Topik: %s)\n", rem.UserPhone, rem.Title)
            
            //  Tandai error di DB supaya tidak diproses lagi
             update := bson.M{"$set": bson.M{"status": "failed_invalid_number"}}
            config.Mongoconn.Collection("reminders").UpdateOne(context.TODO(), bson.M{"_id": rem.ID}, update)
            continue 
        }
		
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
		
		statusCode, apiRes, errSend := atapi.PostJSON[model.APIResponse](kirim, profile.URLApi)

		isSuccess := false
		
		if errSend == nil && statusCode == 200 {
		// Cek isi pesan respon dari API
            // Jika status sukses (biasanya true atau "success") dan tidak ada pesan error fatal
            if apiRes.Message != "Invalid number" && apiRes.Message != "Target not found" {
                isSuccess = true
            } else {
                fmt.Printf("Gagal Kirim (API Reject) ke %s: %s\n", rem.UserPhone, apiRes.Message)
            }
        } else {
            fmt.Printf("Gagal Kirim (Network Error) ke %s: %v\n", rem.UserPhone, errSend)
        }

        if isSuccess {
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