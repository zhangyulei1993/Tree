package dto

import (
	"errors"
	"fmt"
)

var ErrMissingConfigField = errors.New("quota config field missing")

type UpdateConfigRequest struct {
	MaxOwnedFamilies         *int  `json:"maxOwnedFamilies"`
	MaxMembersPerOwnedFamily *int  `json:"maxMembersPerOwnedFamily"`
	MaxJoinedFamilies        *int  `json:"maxJoinedFamilies"`
	SupportsFeaturePreview   *bool `json:"supportsFeaturePreview"`
}

type ImpactPreviewRequest struct {
	MaxOwnedFamilies         *int  `json:"maxOwnedFamilies"`
	MaxMembersPerOwnedFamily *int  `json:"maxMembersPerOwnedFamily"`
	MaxJoinedFamilies        *int  `json:"maxJoinedFamilies"`
	SupportsFeaturePreview   *bool `json:"supportsFeaturePreview"`
}

type ConfigValues struct {
	MaxOwnedFamilies         int
	MaxMembersPerOwnedFamily int
	MaxJoinedFamilies        int
	SupportsFeaturePreview   bool
}

type UpdateFeatureOverridesRequest struct {
	Phones []string `json:"phones"`
}

func (r UpdateConfigRequest) Parse() (ConfigValues, error) {
	return parseConfigValues(r.MaxOwnedFamilies, r.MaxMembersPerOwnedFamily, r.MaxJoinedFamilies, r.SupportsFeaturePreview)
}

func (r ImpactPreviewRequest) Parse() (ConfigValues, error) {
	return parseConfigValues(r.MaxOwnedFamilies, r.MaxMembersPerOwnedFamily, r.MaxJoinedFamilies, r.SupportsFeaturePreview)
}

func parseConfigValues(owned, members, joined *int, featurePreview *bool) (ConfigValues, error) {
	if owned == nil {
		return ConfigValues{}, fmt.Errorf("%w: maxOwnedFamilies", ErrMissingConfigField)
	}
	if members == nil {
		return ConfigValues{}, fmt.Errorf("%w: maxMembersPerOwnedFamily", ErrMissingConfigField)
	}
	if joined == nil {
		return ConfigValues{}, fmt.Errorf("%w: maxJoinedFamilies", ErrMissingConfigField)
	}
	if featurePreview == nil {
		return ConfigValues{}, fmt.Errorf("%w: supportsFeaturePreview", ErrMissingConfigField)
	}
	return ConfigValues{
		MaxOwnedFamilies:         *owned,
		MaxMembersPerOwnedFamily: *members,
		MaxJoinedFamilies:        *joined,
		SupportsFeaturePreview:   *featurePreview,
	}, nil
}
