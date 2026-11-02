package service

import (
	"tree/backend/internal/common/mask"
	"tree/backend/internal/family/tree/vo"
)

func maskPublicTreeResult(result *vo.TreeResult) *vo.TreeResult {
	if result == nil {
		return nil
	}
	masked := *result
	masked.GrayAccessEnabled = false
	masked.Nodes = make([]vo.Node, len(result.Nodes))
	for i, node := range result.Nodes {
		masked.Nodes[i] = node
		masked.Nodes[i].DisplayName = mask.DisplayName(node.DisplayName)
		masked.Nodes[i].GenerationCharacter = nil
		masked.Nodes[i].StopReason = nil
	}
	masked.Edges = make([]vo.Edge, len(result.Edges))
	for i, edge := range result.Edges {
		masked.Edges[i] = edge
		masked.Edges[i].RelationNote = nil
	}
	return &masked
}
