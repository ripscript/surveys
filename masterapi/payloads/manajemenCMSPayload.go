package payloads

type ManajemenCMSPayload struct {
	NamaSection string `json:"nama_section" validate:"required,max=191"`
	TipeSection string `json:"tipe_section" validate:"required,oneof=text picture"`
}

type UpdateStatusSectionPayload struct {
	Status *bool `json:"status" validate:"required"`
}

type UpdateNameSectionPayload struct {
	Name string `json:"name" validate:"required,max=191"`
}

type UpdateContentSectionPayload struct {
	Contents []UpdateContentItemPayload `json:"contents" validate:"required,min=1,dive"`
}

type UpdateContentItemPayload struct {
	Key       string  `json:"key" validate:"required"`
	ValueText *string `json:"value_text"`
}

type UpdateItemsSectionPayload struct {
	Items []UpdateItemPayload `json:"items" validate:"required,min=1,dive"`
}

type UpdateItemPayload struct {
	ItemOrder   int     `json:"item_order" validate:"required"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
	Image       *string `json:"image"`
}

type UpdateMediaSectionPayload struct {
	Media []UpdateMediaItemPayload `json:"media" validate:"required,min=1,dive"`
}

type UpdateMediaItemPayload struct {
	ItemOrder int     `json:"item_order" validate:"required"`
	ImageURL  string  `json:"image_url" validate:"required"`
	Caption   *string `json:"caption"`
}

type UpdateOrderSectionPayload struct {
	Direction string `json:"direction" validate:"required,oneof=up down"`
}
