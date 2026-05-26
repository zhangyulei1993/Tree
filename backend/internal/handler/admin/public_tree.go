package admin

import (
	"strconv"

	"tree/backend/internal/database"
	"tree/backend/internal/model"
	"tree/backend/internal/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Handler) PublicTreeConfig(c *gin.Context) {
	var item model.PublicTreeConfig

	if err := database.DB.Order("id asc").First(&item).Error; err != nil {
		item = model.PublicTreeConfig{
			Title:       "公开关系树",
			Subtitle:    "家族人物关系展示",
			Description: "用于官网公开展示，不读取用户真实家庭数据。",
			Status:      1,
		}

		if err := database.DB.Create(&item).Error; err != nil {
			response.Error(c, 500, "create config failed")
			return
		}
	}

	response.Success(c, item)
}

func (h *Handler) PublicTreeConfigUpdate(c *gin.Context) {
	var req model.PublicTreeConfig

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid params")
		return
	}

	var item model.PublicTreeConfig
	if req.ID != 0 {
		database.DB.First(&item, req.ID)
	}

	if item.ID == 0 {
		database.DB.Order("id asc").First(&item)
	}

	item.Title = req.Title
	item.Subtitle = req.Subtitle
	item.Description = req.Description
	item.Status = req.Status

	if item.ID == 0 {
		if err := database.DB.Create(&item).Error; err != nil {
			response.Error(c, 500, "create config failed")
			return
		}
	} else if err := database.DB.Save(&item).Error; err != nil {
		response.Error(c, 500, "update config failed")
		return
	}

	response.Success(c, item)
}

func (h *Handler) RelationshipTypeList(c *gin.Context) {
	var list []model.RelationshipType
	database.DB.Order("sort desc, id asc").Find(&list)
	response.Success(c, list)
}

func (h *Handler) RelationshipTypeCreate(c *gin.Context) {
	var req model.RelationshipType

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid params")
		return
	}

	if req.Code == "" || req.Name == "" {
		response.Error(c, 400, "code and name are required")
		return
	}

	if req.Direction == "" {
		req.Direction = "one_way"
	}

	if err := database.DB.Create(&req).Error; err != nil {
		response.Error(c, 500, "create failed")
		return
	}

	response.Success(c, req)
}

func (h *Handler) RelationshipTypeUpdate(c *gin.Context) {
	var req model.RelationshipType

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid params")
		return
	}

	if req.ID == 0 {
		response.Error(c, 400, "id is required")
		return
	}

	var item model.RelationshipType
	if err := database.DB.First(&item, req.ID).Error; err != nil {
		response.Error(c, 404, "not found")
		return
	}

	item.Code = req.Code
	item.Name = req.Name
	item.Category = req.Category
	item.Direction = req.Direction
	item.Description = req.Description
	item.Sort = req.Sort
	item.Status = req.Status

	if err := database.DB.Save(&item).Error; err != nil {
		response.Error(c, 500, "update failed")
		return
	}

	response.Success(c, item)
}

func (h *Handler) RelationshipTypeDelete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	if id <= 0 {
		response.Error(c, 400, "invalid id")
		return
	}

	var item model.RelationshipType
	if err := database.DB.First(&item, id).Error; err != nil {
		response.Error(c, 404, "not found")
		return
	}

	var count int64
	database.DB.Model(&model.PublicTreeRelationship{}).Where("relation_type = ?", item.Code).Count(&count)

	if count > 0 {
		response.Error(c, 400, "relationship type is in use")
		return
	}

	if err := database.DB.Delete(&item).Error; err != nil {
		response.Error(c, 500, "delete failed")
		return
	}

	response.SuccessMsg(c, "deleted")
}

func (h *Handler) PublicTreeNodeList(c *gin.Context) {
	var list []model.PublicTreeNode
	database.DB.Order("generation asc, sort desc, id asc").Find(&list)
	response.Success(c, list)
}

func (h *Handler) PublicTreeNodeCreate(c *gin.Context) {
	var req model.PublicTreeNode

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid params")
		return
	}

	if req.Name == "" {
		response.Error(c, 400, "name is required")
		return
	}

	if req.Generation == 0 {
		req.Generation = 1
	}

	if err := database.DB.Create(&req).Error; err != nil {
		response.Error(c, 500, "create failed")
		return
	}

	response.Success(c, req)
}

func (h *Handler) PublicTreeNodeUpdate(c *gin.Context) {
	var req model.PublicTreeNode

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid params")
		return
	}

	if req.ID == 0 {
		response.Error(c, 400, "id is required")
		return
	}

	var item model.PublicTreeNode
	if err := database.DB.First(&item, req.ID).Error; err != nil {
		response.Error(c, 404, "not found")
		return
	}

	item.Name = req.Name
	item.Role = req.Role
	item.Generation = req.Generation
	item.Description = req.Description
	item.Sort = req.Sort
	item.Status = req.Status

	if err := database.DB.Save(&item).Error; err != nil {
		response.Error(c, 500, "update failed")
		return
	}

	response.Success(c, item)
}

func (h *Handler) PublicTreeNodeDelete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	if id <= 0 {
		response.Error(c, 400, "invalid id")
		return
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&model.PublicTreeNode{}, id).Error; err != nil {
			return err
		}
		return tx.Where("from_node_id = ? OR to_node_id = ?", id, id).Delete(&model.PublicTreeRelationship{}).Error
	})

	if err != nil {
		response.Error(c, 500, "delete failed")
		return
	}

	response.SuccessMsg(c, "deleted")
}

func (h *Handler) PublicTreeRelationshipList(c *gin.Context) {
	var list []model.PublicTreeRelationship
	database.DB.Order("sort desc, id asc").Find(&list)
	response.Success(c, list)
}

func (h *Handler) PublicTreeRelationshipCreate(c *gin.Context) {
	var req model.PublicTreeRelationship

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid params")
		return
	}

	if !validatePublicTreeRelationship(c, req) {
		return
	}

	var count int64
	database.DB.Model(&model.PublicTreeRelationship{}).
		Where("from_node_id = ? AND to_node_id = ? AND relation_type = ?", req.FromNodeID, req.ToNodeID, req.RelationType).
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

func (h *Handler) PublicTreeRelationshipUpdate(c *gin.Context) {
	var req model.PublicTreeRelationship

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid params")
		return
	}

	if req.ID == 0 {
		response.Error(c, 400, "id is required")
		return
	}

	if !validatePublicTreeRelationship(c, req) {
		return
	}

	var item model.PublicTreeRelationship
	if err := database.DB.First(&item, req.ID).Error; err != nil {
		response.Error(c, 404, "not found")
		return
	}

	item.FromNodeID = req.FromNodeID
	item.ToNodeID = req.ToNodeID
	item.RelationType = req.RelationType
	item.RelationName = req.RelationName
	item.Remark = req.Remark
	item.Sort = req.Sort
	item.Status = req.Status

	if err := database.DB.Save(&item).Error; err != nil {
		response.Error(c, 500, "update failed")
		return
	}

	response.Success(c, item)
}

func (h *Handler) PublicTreeRelationshipDelete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	if id <= 0 {
		response.Error(c, 400, "invalid id")
		return
	}

	if err := database.DB.Delete(&model.PublicTreeRelationship{}, id).Error; err != nil {
		response.Error(c, 500, "delete failed")
		return
	}

	response.SuccessMsg(c, "deleted")
}

func validatePublicTreeRelationship(c *gin.Context, req model.PublicTreeRelationship) bool {
	if req.FromNodeID == 0 || req.ToNodeID == 0 {
		response.Error(c, 400, "nodes are required")
		return false
	}

	if req.FromNodeID == req.ToNodeID {
		response.Error(c, 400, "nodes cannot be same")
		return false
	}

	if req.RelationType == "" {
		response.Error(c, 400, "relation_type is required")
		return false
	}

	var count int64
	database.DB.Model(&model.PublicTreeNode{}).
		Where("id IN ?", []uint{req.FromNodeID, req.ToNodeID}).
		Count(&count)

	if count != 2 {
		response.Error(c, 400, "node not found")
		return false
	}

	return true
}
