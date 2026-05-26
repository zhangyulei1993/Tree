package model

import "time"

type AdminUser struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Password  string    `gorm:"size:255;not null" json:"-"`
	Nickname  string    `gorm:"size:50" json:"nickname"`
	Role      string    `gorm:"size:30;default:admin" json:"role"`
	Status    int       `gorm:"default:1" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SiteConfig struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	SiteName    string    `gorm:"size:100" json:"site_name"`
	Logo        string    `gorm:"size:255" json:"logo"`
	Phone       string    `gorm:"size:50" json:"phone"`
	Email       string    `gorm:"size:100" json:"email"`
	Address     string    `gorm:"size:255" json:"address"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Banner struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"size:100" json:"title"`
	ImageURL  string    `gorm:"size:255;not null" json:"image_url"`
	LinkURL   string    `gorm:"size:255" json:"link_url"`
	Sort      int       `gorm:"default:0" json:"sort"`
	Status    int       `gorm:"default:1" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Article struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Title      string    `gorm:"size:200;not null" json:"title"`
	Cover      string    `gorm:"size:255" json:"cover"`
	Summary    string    `gorm:"size:500" json:"summary"`
	Content    string    `gorm:"type:longtext" json:"content"`
	CategoryID uint      `gorm:"default:0" json:"category_id"`
	Sort       int       `gorm:"default:0" json:"sort"`
	Status     int       `gorm:"default:1" json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ContactMessage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:50" json:"name"`
	Phone     string    `gorm:"size:50" json:"phone"`
	Email     string    `gorm:"size:100" json:"email"`
	Content   string    `gorm:"type:text" json:"content"`
	Status    int       `gorm:"default:0" json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type UploadFile struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:255" json:"name"`
	URL       string    `gorm:"size:255" json:"url"`
	Path      string    `gorm:"size:255" json:"path"`
	Size      int64     `json:"size"`
	MimeType  string    `gorm:"size:100" json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
}

type Family struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Surname     string    `gorm:"size:50" json:"surname"`
	Origin      string    `gorm:"size:255" json:"origin"`
	Cover       string    `gorm:"size:255" json:"cover"`
	Description string    `gorm:"type:text" json:"description"`
	Status      int       `gorm:"default:1" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type FamilyMember struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	FamilyID   uint      `gorm:"index;not null" json:"family_id"`
	Name       string    `gorm:"size:80;not null" json:"name"`
	Gender     string    `gorm:"size:20;default:unknown" json:"gender"`
	Avatar     string    `gorm:"size:255" json:"avatar"`
	Birthday   string    `gorm:"size:30" json:"birthday"`
	BirthPlace string    `gorm:"size:120" json:"birth_place"`
	Generation int       `gorm:"default:1" json:"generation"`
	Biography  string    `gorm:"type:text" json:"biography"`
	Sort       int       `gorm:"default:0" json:"sort"`
	Status     int       `gorm:"default:1" json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type FamilyRelationship struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	FamilyID     uint      `gorm:"index;not null" json:"family_id"`
	FromMemberID uint      `gorm:"index;not null" json:"from_member_id"`
	ToMemberID   uint      `gorm:"index;not null" json:"to_member_id"`
	RelationType string    `gorm:"size:40;not null" json:"relation_type"`
	Remark       string    `gorm:"size:255" json:"remark"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type PublicTreeConfig struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"size:100" json:"title"`
	Subtitle    string    `gorm:"size:255" json:"subtitle"`
	Description string    `gorm:"type:text" json:"description"`
	Status      int       `gorm:"default:1" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PublicTreeNode struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:80;not null" json:"name"`
	Role        string    `gorm:"size:100" json:"role"`
	Generation  int       `gorm:"default:1" json:"generation"`
	Description string    `gorm:"type:text" json:"description"`
	Sort        int       `gorm:"default:0" json:"sort"`
	Status      int       `gorm:"default:1" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PublicTreeRelationship struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	FromNodeID   uint      `gorm:"index;not null" json:"from_node_id"`
	ToNodeID     uint      `gorm:"index;not null" json:"to_node_id"`
	RelationType string    `gorm:"size:40;not null" json:"relation_type"`
	RelationName string    `gorm:"size:80" json:"relation_name"`
	Remark       string    `gorm:"size:255" json:"remark"`
	Sort         int       `gorm:"default:0" json:"sort"`
	Status       int       `gorm:"default:1" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type RelationshipType struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Code        string    `gorm:"size:40;uniqueIndex;not null" json:"code"`
	Name        string    `gorm:"size:80;not null" json:"name"`
	Category    string    `gorm:"size:40" json:"category"`
	Direction   string    `gorm:"size:20;default:one_way" json:"direction"`
	Description string    `gorm:"size:255" json:"description"`
	Sort        int       `gorm:"default:0" json:"sort"`
	Status      int       `gorm:"default:1" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
