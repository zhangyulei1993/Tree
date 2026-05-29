package admin

import (
	"strconv"
	"strings"

	"tree/backend/internal/database"
	"tree/backend/internal/model"
	"tree/backend/internal/response"

	"github.com/gin-gonic/gin"
)

func (h *Handler) FamilyList(c *gin.Context) {
	var list []model.Family
	db := database.DB.Model(&model.Family{})

	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("name LIKE ? OR surname LIKE ? OR origin LIKE ?", like, like, like)
	}

	if status := strings.TrimSpace(c.Query("status")); status != "" {
		db = db.Where("status = ?", status)
	}

	db.Order("id desc").Find(&list)

	result := make([]gin.H, 0, len(list))
	for _, family := range list {
		var memberCount int64
		var relationshipCount int64

		database.DB.Model(&model.FamilyMember{}).
			Where("family_id = ?", family.ID).
			Count(&memberCount)
		database.DB.Model(&model.FamilyRelationship{}).
			Where("family_id = ?", family.ID).
			Count(&relationshipCount)

		result = append(result, gin.H{
			"id":                 family.ID,
			"name":               family.Name,
			"surname":            family.Surname,
			"origin":             family.Origin,
			"cover":              family.Cover,
			"description":        family.Description,
			"status":             family.Status,
			"created_at":         family.CreatedAt,
			"updated_at":         family.UpdatedAt,
			"member_count":       memberCount,
			"relationship_count": relationshipCount,
		})
	}

	response.Success(c, result)
}

func (h *Handler) FamilyCreate(c *gin.Context) {
	var req model.Family

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid params")
		return
	}

	if req.Name == "" {
		response.Error(c, 400, "name is required")
		return
	}

	if req.Status == 0 {
		req.Status = 1
	}

	if err := database.DB.Create(&req).Error; err != nil {
		response.Error(c, 500, "create failed")
		return
	}

	response.Success(c, req)
}

func (h *Handler) FamilyUpdate(c *gin.Context) {
	var req model.Family

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid params")
		return
	}

	if req.ID == 0 {
		response.Error(c, 400, "id is required")
		return
	}

	var item model.Family
	if err := database.DB.First(&item, req.ID).Error; err != nil {
		response.Error(c, 404, "not found")
		return
	}

	item.Name = req.Name
	item.Surname = req.Surname
	item.Origin = req.Origin
	item.Cover = req.Cover
	item.Description = req.Description
	item.Status = req.Status

	database.DB.Save(&item)

	response.Success(c, item)
}

func (h *Handler) FamilyDelete(c *gin.Context) {
	id := c.Param("id")

	database.DB.Delete(&model.Family{}, id)
	database.DB.Where("family_id = ?", id).Delete(&model.FamilyMember{})
	database.DB.Where("family_id = ?", id).Delete(&model.FamilyRelationship{})

	response.SuccessMsg(c, "deleted")
}

func (h *Handler) MemberList(c *gin.Context) {
	familyID := c.Query("family_id")

	var list []model.FamilyMember
	db := database.DB.Model(&model.FamilyMember{})

	if familyID != "" {
		db = db.Where("family_id = ?", familyID)
	}

	db.Order("generation asc, sort desc, id asc").Find(&list)

	response.Success(c, list)
}

func (h *Handler) MemberCreate(c *gin.Context) {
	var req model.FamilyMember

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid params")
		return
	}

	if req.FamilyID == 0 {
		response.Error(c, 400, "family_id is required")
		return
	}

	if req.Name == "" {
		response.Error(c, 400, "name is required")
		return
	}

	if req.Generation == 0 {
		req.Generation = 1
	}

	if req.Gender == "" {
		req.Gender = "unknown"
	}

	if req.Status == 0 {
		req.Status = 1
	}

	if err := database.DB.Create(&req).Error; err != nil {
		response.Error(c, 500, "create failed")
		return
	}

	response.Success(c, req)
}

func (h *Handler) MemberUpdate(c *gin.Context) {
	var req model.FamilyMember

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid params")
		return
	}

	if req.ID == 0 {
		response.Error(c, 400, "id is required")
		return
	}

	var item model.FamilyMember
	if err := database.DB.First(&item, req.ID).Error; err != nil {
		response.Error(c, 404, "not found")
		return
	}

	item.FamilyID = req.FamilyID
	item.Name = req.Name
	item.Gender = req.Gender
	item.Avatar = req.Avatar
	item.Birthday = req.Birthday
	item.BirthPlace = req.BirthPlace
	item.Generation = req.Generation
	item.Biography = req.Biography
	item.Sort = req.Sort
	item.Status = req.Status

	database.DB.Save(&item)

	response.Success(c, item)
}

func (h *Handler) MemberDelete(c *gin.Context) {
	id := c.Param("id")

	database.DB.Delete(&model.FamilyMember{}, id)
	database.DB.Where("from_member_id = ? OR to_member_id = ?", id, id).Delete(&model.FamilyRelationship{})

	response.SuccessMsg(c, "deleted")
}

func (h *Handler) RelationshipList(c *gin.Context) {
	familyID := c.Query("family_id")

	var list []model.FamilyRelationship
	db := database.DB.Model(&model.FamilyRelationship{})

	if familyID != "" {
		db = db.Where("family_id = ?", familyID)
	}

	db.Order("id desc").Find(&list)

	response.Success(c, list)
}

func (h *Handler) RelationshipCreate(c *gin.Context) {
	var req model.FamilyRelationship

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid params")
		return
	}

	if req.FamilyID == 0 || req.FromMemberID == 0 || req.ToMemberID == 0 {
		response.Error(c, 400, "family and members are required")
		return
	}

	if req.FromMemberID == req.ToMemberID {
		response.Error(c, 400, "members cannot be same")
		return
	}

	if req.RelationType == "" {
		response.Error(c, 400, "relation_type is required")
		return
	}

	if !relationshipTypeExists(req.RelationType) {
		response.Error(c, 400, "invalid relation_type")
		return
	}

	var count int64
	database.DB.Model(&model.FamilyRelationship{}).
		Where("family_id = ? AND from_member_id = ? AND to_member_id = ? AND relation_type = ?", req.FamilyID, req.FromMemberID, req.ToMemberID, req.RelationType).
		Count(&count)

	if count > 0 {
		response.Error(c, 400, "relationship exists")
		return
	}

	if err := database.DB.Create(&req).Error; err != nil {
		response.Error(c, 500, "create failed")
		return
	}

	response.Success(c, req)
}

func (h *Handler) RelationshipUpdate(c *gin.Context) {
	var req model.FamilyRelationship

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid params")
		return
	}

	if req.ID == 0 {
		response.Error(c, 400, "id is required")
		return
	}

	var item model.FamilyRelationship
	if err := database.DB.First(&item, req.ID).Error; err != nil {
		response.Error(c, 404, "not found")
		return
	}

	if req.FromMemberID == req.ToMemberID {
		response.Error(c, 400, "members cannot be same")
		return
	}

	if req.RelationType == "" || !relationshipTypeExists(req.RelationType) {
		response.Error(c, 400, "invalid relation_type")
		return
	}

	item.FamilyID = req.FamilyID
	item.FromMemberID = req.FromMemberID
	item.ToMemberID = req.ToMemberID
	item.RelationType = req.RelationType
	item.Remark = req.Remark

	database.DB.Save(&item)

	response.Success(c, item)
}

func (h *Handler) RelationshipDelete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	if id <= 0 {
		response.Error(c, 400, "invalid id")
		return
	}

	database.DB.Delete(&model.FamilyRelationship{}, id)

	response.SuccessMsg(c, "deleted")
}

func relationshipTypeExists(code string) bool {
	var count int64
	database.DB.Model(&model.RelationshipType{}).
		Where("code = ? AND status = ?", code, 1).
		Count(&count)

	return count > 0
}
