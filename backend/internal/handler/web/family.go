package web

import (
	"sort"

	"tree/backend/internal/database"
	"tree/backend/internal/model"
	"tree/backend/internal/response"

	"github.com/gin-gonic/gin"
)

func FamilyTree(c *gin.Context) {
	familyID := c.Query("family_id")

	var family model.Family
	db := database.DB.Model(&model.Family{}).Where("status = ?", 1)

	if familyID != "" {
		db = db.Where("id = ?", familyID)
	}

	if err := db.Order("id asc").First(&family).Error; err != nil {
		response.Error(c, 404, "family not found")
		return
	}

	var members []model.FamilyMember
	database.DB.
		Where("family_id = ? AND status = ?", family.ID, 1).
		Order("generation asc, sort desc, id asc").
		Find(&members)

	var relationships []model.FamilyRelationship
	database.DB.
		Where("family_id = ?", family.ID).
		Order("id asc").
		Find(&relationships)

	generationMap := map[int][]model.FamilyMember{}
	var generationKeys []int
	seen := map[int]bool{}

	for _, member := range members {
		generationMap[member.Generation] = append(generationMap[member.Generation], member)

		if !seen[member.Generation] {
			seen[member.Generation] = true
			generationKeys = append(generationKeys, member.Generation)
		}
	}

	sort.Ints(generationKeys)

	response.Success(c, gin.H{
		"family":          family,
		"members":         members,
		"relationships":   relationships,
		"generation_keys": generationKeys,
		"generations":     generationMap,
	})
}

func FamilyMemberDetail(c *gin.Context) {
	id := c.Param("id")

	var member model.FamilyMember
	if err := database.DB.Where("id = ? AND status = ?", id, 1).First(&member).Error; err != nil {
		response.Error(c, 404, "member not found")
		return
	}

	var relationships []model.FamilyRelationship
	database.DB.
		Where("from_member_id = ? OR to_member_id = ?", member.ID, member.ID).
		Order("id asc").
		Find(&relationships)

	response.Success(c, gin.H{
		"member":        member,
		"relationships": relationships,
	})
}
