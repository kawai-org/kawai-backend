package gcallapi

import (
	"context"
	"errors"

	"github.com/kawai-org/kawai-backend/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/oauth2"
)

// 1. Fungsi untuk mengambil Config dari DB
func credentialsFromDB(db *mongo.Database) (*oauth2.Config, error) {
	collection := db.Collection("credentials")
	var credentialRecord model.CredentialRecord
	err := collection.FindOne(context.TODO(), bson.M{}).Decode(&credentialRecord)
	if err != nil {
		return nil, err
	}

	if len(credentialRecord.RedirectURIs) == 0 {
		return nil, errors.New("no redirect URIs found in credentials")
	}

	config := &oauth2.Config{
		ClientID:     credentialRecord.ClientID,
		ClientSecret: credentialRecord.ClientSecret,
		Scopes:       credentialRecord.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  credentialRecord.AuthURI,
			TokenURL: credentialRecord.TokenURI,
		},
		RedirectURL: credentialRecord.RedirectURIs[0],
	}

	return config, nil
}

// 2. Fungsi untuk mengambil Token dari DB
func tokenFromDB(db *mongo.Database) (*oauth2.Token, error) {
	collection := db.Collection("tokens")
	var tokenRecord model.CredentialRecord
	err := collection.FindOne(context.TODO(), bson.M{}).Decode(&tokenRecord)
	if err != nil {
		return nil, err
	}

	token := &oauth2.Token{
		AccessToken:  tokenRecord.Token,
		RefreshToken: tokenRecord.RefreshToken,
		TokenType:    "Bearer",
		Expiry:       tokenRecord.Expiry,
	}

	if tokenRecord.Token == "" {
		return nil, errors.New("token tidak ada")
	}

	return token, nil
}

// 3. Fungsi UNTUK REFRESH TOKEN 
func refreshToken(config *oauth2.Config, token *oauth2.Token) (*oauth2.Token, error) {
	ts := config.TokenSource(context.Background(), token)
	newToken, err := ts.Token()
	if err != nil {
		return nil, err
	}
	return newToken, nil
}

// 4. Fungsi UNTUK SIMPAN TOKEN KE DB 
func saveToken(db *mongo.Database, token *oauth2.Token) error {
	collection := db.Collection("tokens")
	tokenRecord := bson.M{
		"token":         token.AccessToken,
		"refresh_token": token.RefreshToken,
		"expiry":        token.Expiry,
	}

	_, err := collection.UpdateOne(
		context.TODO(),
		bson.M{},
		bson.M{"$set": tokenRecord},
		options.Update().SetUpsert(true),
	)
	return err
}