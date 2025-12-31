package timeparse

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Helper Map untuk Nama Bulan 
var monthMap = map[string]time.Month{
	"januari": 1, "jan": 1, "january": 1,
	"februari": 2, "feb": 2, "pebruari": 2, "february": 2,
	"maret": 3, "mar": 3, "march": 3,
	"april": 4, "apr": 4,
	"mei": 5, "may": 5,
	"juni": 6, "jun": 6, "june": 6,
	"juli": 7, "jul": 7, "july": 7,
	"agustus": 8, "agt": 8, "aug": 8, "agst": 8, "agus": 8, "agu": 8,
	"september": 9, "sep": 9, "sept": 9,
	"oktober": 10, "okt": 10, "oct": 10, "october": 10,
	"november": 11, "nov": 11, "nopember": 11, "nop": 11,
	"desember": 12, "des": 12, "dec": 12, "december": 12,
}

// Helper Map untuk Nama Hari (Termasuk singkatan chat)
var dayMap = map[string]time.Weekday{
	"minggu": time.Sunday, "ahad": time.Sunday, "mg": time.Sunday, "mgg": time.Sunday, "sun": time.Sunday,
	"senin": time.Monday, "sen": time.Monday, "sn": time.Monday, "mon": time.Monday,
	"selasa": time.Tuesday, "sls": time.Tuesday, "slasa": time.Tuesday, "tue": time.Tuesday,
	"rabu": time.Wednesday, "rab": time.Wednesday, "rb": time.Wednesday, "wed": time.Wednesday,
	"kamis": time.Thursday, "kam": time.Thursday, "kms": time.Thursday, "thu": time.Thursday,
	"jumat": time.Friday, "jum'at": time.Friday, "jum": time.Friday, "jmt": time.Friday, "fri": time.Friday,
	"sabtu": time.Saturday, "sab": time.Saturday, "sbt": time.Saturday, "sat": time.Saturday,
}

func ParseNaturalTime(text string) (time.Time, string) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	targetTime := now
	textLower := strings.ToLower(text)
	
	isDateSet := false 

	// ==========================================
	// A. DETEKSI DURASI ("... LAGI")
	// Support: "5 mnt lg", "sejam lg", "10 dtk lagi"
	// ==========================================
	reDuration := regexp.MustCompile(`(\d+|sebuah|satu|setengah|se)\s*(jam|menit|detik|jm|mnt|dtk)\s*(lagi|lg)?`)
	matchDur := reDuration.FindStringSubmatch(textLower)
	
	if len(matchDur) > 0 {
		angkaStr := matchDur[1]
		satuan := matchDur[2]
		
		var nilai float64
		if angkaStr == "se" || angkaStr == "satu" || angkaStr == "sebuah" {
			nilai = 1
		} else if angkaStr == "setengah" {
			nilai = 0.5
		} else {
			nilai, _ = strconv.ParseFloat(angkaStr, 64)
		}

		if satuan == "jam" || satuan == "jm" {
			targetTime = targetTime.Add(time.Duration(nilai * float64(time.Hour)))
		} else if satuan == "menit" || satuan == "mnt" {
			targetTime = targetTime.Add(time.Duration(nilai * float64(time.Minute)))
		} else if satuan == "detik" || satuan == "dtk" {
			targetTime = targetTime.Add(time.Duration(nilai * float64(time.Second)))
		}
		
		// Daftar kata sampah untuk dibersihkan dari judul
		sampah := []string{matchDur[0], "ingatkan", "remind", "ingat", "lg", "lagi"}
		return targetTime, cleanText(text, sampah)
	}

	// ==========================================
	// B. DETEKSI TANGGAL FORMAL (25/12)
	// ==========================================
	reDateFormal := regexp.MustCompile(`(\d{1,2})[-/](\d{1,2})([-./]\d{2,4})?`)
	matchDate := reDateFormal.FindStringSubmatch(textLower)
	
	if len(matchDate) > 0 {
		day, _ := strconv.Atoi(matchDate[1])
		month, _ := strconv.Atoi(matchDate[2])
		year := now.Year()
		if len(matchDate) > 3 && matchDate[3] != "" {
			yearStr := strings.TrimLeft(matchDate[3], "-./")
			if len(yearStr) == 2 { yearStr = "20" + yearStr }
			year, _ = strconv.Atoi(yearStr)
		} else {
			tempDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc)
			if tempDate.Before(now) { year++ }
		}
		targetTime = time.Date(year, time.Month(month), day, targetTime.Hour(), targetTime.Minute(), 0, 0, loc)
		isDateSet = true
		textLower = strings.Replace(textLower, matchDate[0], "", 1)
	}

	// ==========================================
	// C. DETEKSI TANGGAL TEXT ("17 Agt")
	// ==========================================
	if !isDateSet {
		for monthName, monthIdx := range monthMap {
			reTextDate := regexp.MustCompile(fmt.Sprintf(`(tgl|tanggal|tgll)?\s*(\d{1,2})\s*%s`, monthName))
			matchTextDate := reTextDate.FindStringSubmatch(textLower)
			
			if len(matchTextDate) > 0 {
				day, _ := strconv.Atoi(matchTextDate[2])
				year := now.Year()
				tempDate := time.Date(year, monthIdx, day, 0, 0, 0, 0, loc)
				if tempDate.Before(now.AddDate(0, 0, -1)) { year++ }
				targetTime = time.Date(year, monthIdx, day, targetTime.Hour(), targetTime.Minute(), 0, 0, loc)
				isDateSet = true
				textLower = strings.Replace(textLower, matchTextDate[0], "", 1)
				break
			}
		}
	}

	// ==========================================
	// D. DETEKSI HARI ("Senin", "Jumat")
	// ==========================================
	if !isDateSet {
		for dayName, dayIdx := range dayMap {
			// Pakai boundary \b agar "senin" tidak match dengan "seninan"
			if strings.Contains(textLower, dayName) {
				currentDay := int(now.Weekday())
				targetDay := int(dayIdx)
				daysToAdd := (targetDay - currentDay + 7) % 7
				if daysToAdd == 0 { daysToAdd = 7 }
				targetTime = targetTime.AddDate(0, 0, daysToAdd)
				isDateSet = true
				textLower = strings.ReplaceAll(textLower, dayName, "")
				textLower = strings.ReplaceAll(textLower, "hari", "")
				textLower = strings.ReplaceAll(textLower, "hri", "")
				break
			}
		}
	}

	// ==========================================
	// E. DETEKSI RELATIF ("Bsk", "Lsa", "Tgl 17")
	// ==========================================
	if !isDateSet {
		addedDays := 0
		// Cek regex Besok/Lusa dan singkatannya
		if regexp.MustCompile(`\b(besok|bsk|bsok)\b`).MatchString(textLower) {
			addedDays = 1
		} else if regexp.MustCompile(`\b(lusa|lsa)\b`).MatchString(textLower) {
			addedDays = 2
		} else {
			reTgl := regexp.MustCompile(`(tgl|tanggal)\s*(\d{1,2})`)
			matchTgl := reTgl.FindStringSubmatch(textLower)
			if len(matchTgl) > 0 {
				day, _ := strconv.Atoi(matchTgl[2])
				year, month, _ := now.Date()
				if day < now.Day() { month++ }
				targetTime = time.Date(year, month, day, targetTime.Hour(), targetTime.Minute(), 0, 0, loc)
				isDateSet = true
			}
		}

		if addedDays > 0 {
			targetTime = targetTime.AddDate(0, 0, addedDays)
			isDateSet = true
		}
	}

	// ==========================================
	// F. DETEKSI JAM ("Jam 10", "Pkl 9", "10:30")
	// ==========================================
	// Support: jam, jm, pukul, pkl
	reJam := regexp.MustCompile(`(jam|pukul|pkl|jm)?\s*(\d{1,2})[.:](\d{1,2})`) 
	reJamSimple := regexp.MustCompile(`(jam|pukul|pkl|jm)\s*(\d{1,2})`)       
	
	finalHour := 9
	finalMin := 0
	foundTime := false

	matchClock := reJam.FindStringSubmatch(textLower)
	if len(matchClock) > 0 {
		finalHour, _ = strconv.Atoi(matchClock[2])
		finalMin, _ = strconv.Atoi(matchClock[3])
		foundTime = true
		textLower = strings.Replace(textLower, matchClock[0], "", 1)
	} else {
		matchSimple := reJamSimple.FindStringSubmatch(textLower)
		if len(matchSimple) > 0 {
			finalHour, _ = strconv.Atoi(matchSimple[2])
			foundTime = true
			textLower = strings.Replace(textLower, matchSimple[0], "", 1)
		}
	}

	if !isDateSet && foundTime && time.Date(now.Year(), now.Month(), now.Day(), finalHour, finalMin, 0, 0, loc).Before(now) {
		targetTime = targetTime.AddDate(0, 0, 1)
	}

	if foundTime || isDateSet {
		if !foundTime && isDateSet {
			finalHour = 9
			finalMin = 0
		}
		targetTime = time.Date(targetTime.Year(), targetTime.Month(), targetTime.Day(), finalHour, finalMin, 0, 0, loc)
	} else {
		return time.Time{}, ""
	}

	// ==========================================
	// G. BERSIHKAN JUDUL (TERMASUK SINGKATAN)
	// ==========================================
	// Daftar kata sampah yang harus dibuang dari judul
	trashWords := []string{
		"ingatkan", "remind", "ingat", 
		"pada", "hari", "hri", "tgl", "tanggal", "tgll", 
		"besok", "bsk", "bsok", "lusa", "lsa", 
		"jam", "jm", "pukul", "pkl", 
		"nanti", "nt", "lagi", "lg",
	}
	cleanTitle := cleanText(text, trashWords)
	return targetTime, cleanTitle
}

func cleanText(text string, removeList []string) string {
	lower := strings.ToLower(text)
	reTime := regexp.MustCompile(`\d{1,2}[:.]\d{1,2}`)
	lower = reTime.ReplaceAllString(lower, "")
	
	for _, word := range removeList {
		// \b boundary biar gak ngehapus "jam" dari kata "jambret"
		re := regexp.MustCompile(fmt.Sprintf(`\b%s\b`, word)) 
		lower = re.ReplaceAllString(lower, "")
	}
	
	// Kembalikan huruf besar di awal (Title Case) sederhana
	cleaned := strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(lower, " "))
	if len(cleaned) > 1 {
		return strings.ToUpper(string(cleaned[0])) + cleaned[1:]
	}
	return cleaned
}