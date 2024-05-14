package config

import (
	"context"
	"log"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

var App *firebase.App

func init() {
	initFirebase()
}

func initFirebase() {
	opt := option.WithCredentialsFile("secrets/share-compute-firebase-adminsdk-bolvl-c394b053cf.json")
	var err error
	App, err = firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing Firebase app: %v", err)
	}
}
