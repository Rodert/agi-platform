package model

import "gorm.io/datatypes"

type ImageModel struct {
	BaseModel
	Code                string         `gorm:"column:code" json:"code"`
	DisplayName         string         `gorm:"column:display_name" json:"display_name"`
	Description         string         `gorm:"column:description" json:"description"`
	CoverURL            string         `gorm:"column:cover_url" json:"cover_url"`
	PriceCredits        int64          `gorm:"column:price_credits" json:"price_credits"`
	SupportedSizes      datatypes.JSON `gorm:"column:supported_sizes" json:"supported_sizes"`
	SupportTextToImage  bool           `gorm:"column:support_text_to_image" json:"support_text_to_image"`
	SupportImageToImage bool           `gorm:"column:support_image_to_image" json:"support_image_to_image"`
	SupportEdit         bool           `gorm:"column:support_edit" json:"support_edit"`
	MaxImagesPerRequest int            `gorm:"column:max_images_per_request" json:"max_images_per_request"`
	AutoRefundOnFailure bool           `gorm:"column:auto_refund_on_failure" json:"auto_refund_on_failure"`
	Enabled             bool           `gorm:"column:enabled" json:"enabled"`
	Recommended         bool           `gorm:"column:recommended" json:"recommended"`
	SortOrder           int            `gorm:"column:sort_order" json:"sort_order"`
}

func (ImageModel) TableName() string {
	return "image_models"
}

type ImageModelRoute struct {
	ID                uint64         `gorm:"primaryKey;column:id" json:"id"`
	ModelID           uint64         `gorm:"column:model_id" json:"model_id"`
	ProviderID        uint64         `gorm:"column:provider_id" json:"provider_id"`
	ProviderKeyID     *uint64        `gorm:"column:provider_key_id" json:"provider_key_id,omitempty"`
	ProviderModelName string         `gorm:"column:provider_model_name" json:"provider_model_name"`
	Enabled           bool           `gorm:"column:enabled" json:"enabled"`
	Priority          int            `gorm:"column:priority" json:"priority"`
	Weight            int            `gorm:"column:weight" json:"weight"`
	ExtraConfig       datatypes.JSON `gorm:"column:extra_config" json:"extra_config"`
}

func (ImageModelRoute) TableName() string {
	return "image_model_routes"
}
