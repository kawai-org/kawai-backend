package atdb

import (
	"context"

	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/model"
)

// InsertNote adalah fungsi untuk menyimpan catatan user ke MongoDB
func InsertNote(note model.Note) (interface{}, error) {
	collection := config.Mongoconn.Collection("notes")
	return collection.InsertOne(context.TODO(), note)
}