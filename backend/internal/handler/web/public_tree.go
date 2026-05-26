package web

import (
	"sort"

	"tree/backend/internal/database"
	"tree/backend/internal/model"
	"tree/backend/internal/response"

	"github.com/gin-gonic/gin"
)

func PublicTree(c *gin.Context) {
	var config model.PublicTreeConfig
	if err := database.DB.Where("status = ?", 1).Order("id asc").First(&config).Error; err != nil {
		response.Error(c, 404, "public tree config not found")
		return
	}

	var nodes []model.PublicTreeNode
	database.DB.
		Where("status = ?", 1).
		Order("generation asc, sort desc, id asc").
		Find(&nodes)

	var relationships []model.PublicTreeRelationship
	database.DB.
		Where("status = ?", 1).
		Order("sort desc, id asc").
		Find(&relationships)

	var relationshipTypes []model.RelationshipType
	database.DB.Where("status = ?", 1).Find(&relationshipTypes)

	relationshipTypeMap := map[string]model.RelationshipType{}
	for _, item := range relationshipTypes {
		relationshipTypeMap[item.Code] = item
	}

	nodeIDs := map[uint]bool{}
	for _, node := range nodes {
		nodeIDs[node.ID] = true
	}

	filteredRelationships := make([]model.PublicTreeRelationship, 0, len(relationships))
	for _, relationship := range relationships {
		if nodeIDs[relationship.FromNodeID] && nodeIDs[relationship.ToNodeID] {
			if relationship.RelationName == "" {
				relationship.RelationName = relationshipTypeMap[relationship.RelationType].Name
			}
			filteredRelationships = append(filteredRelationships, relationship)
		}
	}

	generationMap := map[int][]model.PublicTreeNode{}
	var generationKeys []int
	seen := map[int]bool{}

	for _, node := range nodes {
		generationMap[node.Generation] = append(generationMap[node.Generation], node)

		if !seen[node.Generation] {
			seen[node.Generation] = true
			generationKeys = append(generationKeys, node.Generation)
		}
	}

	sort.Ints(generationKeys)

	response.Success(c, gin.H{
		"config":          config,
		"nodes":           nodes,
		"relationships":   filteredRelationships,
		"generation_keys": generationKeys,
		"generations":     generationMap,
	})
}
