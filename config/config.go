package config

import (
	"os"
)

// Sederhanakan saja, jangan panggil package lain di sini untuk menghindari Import Cycle
var PrivateKey string = os.Getenv("privateKey")
var PublicKey string = os.Getenv("publicKey")
var PhoneNumber string = os.Getenv("PHONENUMBER")

// Tambahkan variabel yang hilang agar controller tidak error
var WAAPIToken string
var WAAPIGetToken string = "https://api.whatsauth.my.id/v1/gettoken" // Contoh URL