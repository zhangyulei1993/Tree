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
	SupportsGenerationNaming *bool `json:"supportsGenerationNaming"`
}

type ImpactPreviewRequest struct {
	MaxOwnedFamilies         *int  `json:"maxOwnedFamilies"`
	MaxMembersPerOwnedFamily *int  `json:"maxMembersPerOwnedFamily"`
	MaxJoinedFamilies        *int  `json:"maxJoinedFamilies"`
	SupportsGenerationNaming *bool `json:"supportsGenerationNaming"`
}

type ConfigValues struct {
	MaxOwnedFamilies         int
	MaxMembersPerOwnedFamily int
	MaxJoinedFamilies        int
	SupportsGenerationNaming bool
}

type UpdateFeatureOverridesRequest struct {
	Phones []string `json:"phones"`
}

func (r UpdateConfigRequest) Parse() (ConfigValues, error) {
	return parseConfigValues(r.MaxOwnedFamilies, r.MaxMembersPerOwnedFamily, r.MaxJoinedFamilies, r.SupportsGenerationNaming)
}

func (r ImpactPreviewRequest) Parse() (ConfigValues, error) {
	return parseConfigValues(r.MaxOwnedFamilies, r.MaxMembersPerOwnedFamily, r.MaxJoinedFamilies, r.SupportsGenerationNaming)
}

func parseConfigValues(owned, members, joined *int, generationNaming *bool) (ConfigValues, error) {
	if owned == nil {
		return ConfigValues{}, fmt.Errorf("%w: maxOwnedFamilies", ErrMissingConfigField)
	}
	if members == nil {
		return ConfigValues{}, fmt.Errorf("%w: maxMembersPerOwnedFamily", ErrMissingConfigField)
	}
	if joined == nil {
		return ConfigValues{}, fmt.Errorf("%w: maxJoinedFamilies", ErrMissingConfigField)
	}
	if generationNaming == nil {
		return ConfigValues{}, fmt.Errorf("%w: supportsGenerationNaming", ErrMissingConfigField)
	}
	return ConfigValues{
		MaxOwnedFamilies:         *owned,
		MaxMembersPerOwnedFamily: *members,
		MaxJoinedFamilies:        *joined,
		SupportsGenerationNaming: *generationNaming,
	}, nil
}
