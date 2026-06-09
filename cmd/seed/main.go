package main

import (
	"log"

	"github.com/kidkender/prohealth-api/internal/config"
	"github.com/kidkender/prohealth-api/internal/model"
	"github.com/kidkender/prohealth-api/pkg/database"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	config.LoadEnv()

	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("Admin@123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("failed to hash password: ", err)
	}
	err = db.Transaction(func(tx *gorm.DB) error {

		admin := model.User{
			Email:    "admin@prohealth.com",
			Password: string(hash),
			Role:     model.RoleAdmin,
		}

		if err := tx.Create(&admin).Error; err != nil {
			log.Fatal("failed to create admin: ", err)
			return err
		}

		adminProfile := model.UserDetail{
			UserId:   admin.ID,
			Fullname: "Admin",
			User:     admin,
		}

		if err := tx.Create(&adminProfile).Error; err != nil {
			log.Fatal("failed to create admin profile: ", err)
			return err
		}

		log.Println("Admin seeded successfully")
		return nil
	})
}
