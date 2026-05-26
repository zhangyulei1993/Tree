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
	seedRelationshipTypes()
	seedPublicTree()

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
		&model.PublicTreeConfig{},
		&model.PublicTreeNode{},
		&model.PublicTreeRelationship{},
		&model.RelationshipType{},
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

func seedRelationshipTypes() {
	defaults := []model.RelationshipType{
		{Code: "father_child", Name: "父子", Category: "blood", Direction: "one_way", Description: "父亲与子女关系，前一位为父亲，后一位为子女。", Sort: 100, Status: 1},
		{Code: "mother_child", Name: "母子", Category: "blood", Direction: "one_way", Description: "母亲与子女关系，前一位为母亲，后一位为子女。", Sort: 90, Status: 1},
		{Code: "spouse", Name: "配偶", Category: "marriage", Direction: "two_way", Description: "夫妻或配偶关系。", Sort: 80, Status: 1},
		{Code: "uncle_nephew", Name: "叔侄", Category: "blood", Direction: "one_way", Description: "叔叔与侄子或侄女关系。", Sort: 70, Status: 1},
		{Code: "maternal_uncle_nephew", Name: "舅甥", Category: "blood", Direction: "one_way", Description: "舅舅与外甥或外甥女关系。", Sort: 60, Status: 1},
		{Code: "brothers", Name: "兄弟", Category: "blood", Direction: "two_way", Description: "男性兄弟关系。", Sort: 50, Status: 1},
		{Code: "sisters", Name: "姐妹", Category: "blood", Direction: "two_way", Description: "女性姐妹关系。", Sort: 40, Status: 1},
		{Code: "brother_sister", Name: "兄妹", Category: "blood", Direction: "two_way", Description: "哥哥与妹妹关系。", Sort: 30, Status: 1},
		{Code: "sister_brother", Name: "姐弟", Category: "blood", Direction: "two_way", Description: "姐姐与弟弟关系。", Sort: 20, Status: 1},
		{Code: "siblings", Name: "兄弟姐妹", Category: "blood", Direction: "two_way", Description: "通用兄弟姐妹关系。", Sort: 10, Status: 1},
	}

	DB.Model(&model.RelationshipType{}).
		Where("code IN ?", []string{"parent_child", "affinity"}).
		Update("status", 0)

	for _, item := range defaults {
		var existing model.RelationshipType
		err := DB.Where("code = ?", item.Code).First(&existing).Error

		if err != nil {
			DB.Create(&item)
			continue
		}

		DB.Model(&existing).Updates(map[string]interface{}{
			"name":        item.Name,
			"category":    item.Category,
			"direction":   item.Direction,
			"description": item.Description,
			"sort":        item.Sort,
			"status":      item.Status,
		})
	}
}

func seedPublicTree() {
	DB.Model(&model.PublicTreeRelationship{}).
		Where("relation_type LIKE ?", "spouse%").
		Update("relation_type", "spouse")

	var nodeCount int64
	DB.Model(&model.PublicTreeNode{}).Count(&nodeCount)

	if nodeCount > 0 {
		return
	}

	config := model.PublicTreeConfig{
		Title:       "汉武帝重要亲属关系示意",
		Subtitle:    "公开历史人物亲属关系展示",
		Description: "本关系树仅用于官网公开展示，不读取用户真实家庭树数据。",
		Status:      1,
	}

	DB.Create(&config)

	nodes := []model.PublicTreeNode{
		{
			Name:        "汉景帝刘启",
			Role:        "汉武帝父亲",
			Generation:  1,
			Description: "西汉皇帝，汉武帝刘彻之父。",
			Sort:        90,
			Status:      1,
		},
		{
			Name:        "王娡",
			Role:        "汉武帝母亲",
			Generation:  1,
			Description: "汉武帝刘彻之母，后为皇太后。",
			Sort:        80,
			Status:      1,
		},
		{
			Name:        "田蚡",
			Role:        "汉武帝舅舅",
			Generation:  1,
			Description: "王娡之弟，汉武帝的重要外戚人物。",
			Sort:        70,
			Status:      1,
		},
		{
			Name:        "汉武帝刘彻",
			Role:        "核心人物",
			Generation:  2,
			Description: "西汉第七位皇帝，本示意树的核心人物。",
			Sort:        100,
			Status:      1,
		},
		{
			Name:        "陈阿娇",
			Role:        "第一任皇后",
			Generation:  2,
			Description: "汉武帝早期皇后。",
			Sort:        80,
			Status:      1,
		},
		{
			Name:        "卫子夫",
			Role:        "皇后",
			Generation:  2,
			Description: "汉武帝皇后，刘据之母。",
			Sort:        70,
			Status:      1,
		},
		{
			Name:        "卫青",
			Role:        "卫子夫之弟",
			Generation:  2,
			Description: "卫子夫之弟，与汉武帝形成重要姻亲关系。",
			Sort:        60,
			Status:      1,
		},
		{
			Name:        "刘据",
			Role:        "戾太子",
			Generation:  3,
			Description: "汉武帝与卫子夫之子。",
			Sort:        90,
			Status:      1,
		},
		{
			Name:        "刘弗陵",
			Role:        "汉昭帝",
			Generation:  3,
			Description: "汉武帝之子，后来即位为汉昭帝。",
			Sort:        80,
			Status:      1,
		},
	}

	if err := DB.Create(&nodes).Error; err != nil {
		log.Printf("seed public tree nodes failed: %v", err)
		return
	}

	relationships := []model.PublicTreeRelationship{
		{
			FromNodeID:   nodes[0].ID,
			ToNodeID:     nodes[3].ID,
			RelationType: "father_child",
			RelationName: "父子",
			Remark:       "父亲与儿子",
			Sort:         100,
			Status:       1,
		},
		{
			FromNodeID:   nodes[1].ID,
			ToNodeID:     nodes[3].ID,
			RelationType: "mother_child",
			RelationName: "母子",
			Remark:       "母亲与儿子",
			Sort:         90,
			Status:       1,
		},
		{
			FromNodeID:   nodes[2].ID,
			ToNodeID:     nodes[3].ID,
			RelationType: "maternal_uncle_nephew",
			RelationName: "舅甥",
			Remark:       "舅舅与外甥",
			Sort:         80,
			Status:       1,
		},
		{
			FromNodeID:   nodes[3].ID,
			ToNodeID:     nodes[4].ID,
			RelationType: "spouse",
			RelationName: "配偶",
			Remark:       "丈夫与配偶",
			Sort:         70,
			Status:       1,
		},
		{
			FromNodeID:   nodes[3].ID,
			ToNodeID:     nodes[5].ID,
			RelationType: "spouse",
			RelationName: "配偶",
			Remark:       "丈夫与配偶",
			Sort:         60,
			Status:       1,
		},
		{
			FromNodeID:   nodes[6].ID,
			ToNodeID:     nodes[3].ID,
			RelationType: "uncle_nephew",
			RelationName: "叔侄",
			Remark:       "叔叔与侄子或侄女",
			Sort:         50,
			Status:       1,
		},
		{
			FromNodeID:   nodes[5].ID,
			ToNodeID:     nodes[7].ID,
			RelationType: "mother_child",
			RelationName: "母子",
			Remark:       "母亲与儿子",
			Sort:         40,
			Status:       1,
		},
		{
			FromNodeID:   nodes[3].ID,
			ToNodeID:     nodes[7].ID,
			RelationType: "father_child",
			RelationName: "父子",
			Remark:       "父亲与儿子",
			Sort:         30,
			Status:       1,
		},
		{
			FromNodeID:   nodes[3].ID,
			ToNodeID:     nodes[8].ID,
			RelationType: "father_child",
			RelationName: "父子",
			Remark:       "父亲与儿子",
			Sort:         20,
			Status:       1,
		},
	}

	if err := DB.Create(&relationships).Error; err != nil {
		log.Printf("seed public tree relationships failed: %v", err)
	}
}
