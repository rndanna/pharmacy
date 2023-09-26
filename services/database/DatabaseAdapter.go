package database

import (
	"log"
	"os"
	"pharmacy/models"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"golang.org/x/crypto/bcrypt"
)

type DatabaseService struct {
	DB *gorm.DB
}

var Database DatabaseService

func (db *DatabaseService) Init() bool {
	err := db.dbInit()
	if err != nil {
		return false
	}

	db.migrationsInit()

	db.createSuperUser()

	return true
}

func (db *DatabaseService) Close() error {
	return db.DB.Close()
}

func (db *DatabaseService) dbInit() error {
	var err error
	db.DB, err = gorm.Open(
		os.Getenv("DB_DRIVER"),
		os.Getenv("DB_DSN"),
	)
	if err != nil {
		log.Fatalf("Cant connect to DB: %s\n", err.Error())
	}
	db.DB.Debug()
	db.DB = db.DB.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4 auto_increment=1")

	return nil
}
func (db *DatabaseService) migrationsInit() {
	db.DB.Debug().AutoMigrate(
		&models.Product{},
		&models.Order{},
		&models.User{},
		&models.Feedback{},
		&models.Basket{},
		&models.BasketContent{},
		&models.PharmacyAddress{},
		&models.Category{},
		&models.Favorites{},
	)
}

func (db *DatabaseService) createSuperUser() {
	superUser := models.User{
		Email:    "superuser@mail.ru",
		Password: "superuser",
		Login:    "superuser",
		Role:     models.ROLE_ADMIN,
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(superUser.Password), 14)
	if err != nil {
		err.Error()
		return
	}
	superUser.Password = string(bytes)

	db.DB.Where("login = ?", "superuser").Find(&superUser)
	if superUser.ID != 0 {
		return
	}

	if createErr := db.DB.Create(&superUser); createErr.Error != nil {
		return
	}

	return
}
