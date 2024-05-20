package pghandling

import (
	"errors"
	"fmt"
	"log"

	"github.com/appblocks-hub/SHIELD/functions/general"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SetupDB initialize database connection to postgres db.
func SetupDB() (*gorm.DB, error) {

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s", general.Envs["SHIELD_POSTGRES_HOST"], general.Envs["SHIELD_POSTGRES_USER"], general.Envs["SHIELD_POSTGRES_PASSWORD"], general.Envs["SHIELD_POSTGRES_NAME"], general.Envs["SHIELD_POSTGRES_PORT"], general.Envs["SHIELD_POSTGRES_SSLMODE"], general.Envs["SHIELD_POSTGRES_TIMEZONE"])
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Println("DB connection err:", err)

		err = errors.New("DB connection err")
	}

	return db, err
}
