package test

import (
	"log"
	"os"
	"testing"

	"github.com/shashimalcse/cronuseo/internal/config"
	db "github.com/shashimalcse/cronuseo/internal/db/mongo"
)

func DB(t *testing.T) *db.MongoDB {

	cfg, err := config.Load("../../config/run-test.yml")
	if err != nil {
		log.Fatal("Error while loading config for test.")
		os.Exit(-1)
	}
	logger := InitLogger()
	mongo := db.Init(cfg, logger)

	return mongo

}
