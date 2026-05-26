package database

import (
	"fmt"
	"log"
	"time"

	"tree/backend/internal/config"
	"tree/backend/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(cfg *config.Config) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBCharset,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("mysql connect failed: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("mysql db pool failed: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db

	autoMigrate()
	seedAdmin()

	log.Println("mysql connected")
}

func autoMigrate() {
	err := DB.AutoMigrate(
		&model.AdminUser{},
		&model.SiteConfig{},
		&model.Banner{},
		&model.Article{},
		&model.ContactMessage{},
		&model.UploadFile{},
		&model.Family{},
		&model.FamilyMember{},
		&model.FamilyRelationship{},
	)

	if err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}
}

func seedAdmin() {
	var count int64
	DB.Model(&model.AdminUser{}).Where("username = ?", "admin").Count(&count)

	if count > 0 {
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)

	admin := model.AdminUser{
		Username: "admin",
		Password: string(hash),
		Nickname: "Admin",
		Role:     "admin",
		Status:   1,
	}

	DB.Create(&admin)

	log.Println("default admin created: admin / admin123")
}
