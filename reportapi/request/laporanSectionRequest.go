package request // atau package payloads

import "encoding/json"

// ============================================================================
// LEVEL 1: SECTION (Bab Utama)
// ============================================================================
type UpsertSectionPayload struct {
	Title         string                    `json:"title" validate:"required,max=255"`
	HasSubSection *bool                     `json:"has_sub_section" validate:"required"`
	SubSections   []UpsertSubSectionPayload `json:"sub_sections" validate:"omitempty,dive"`

	ComponentType   string          `json:"component_type" validate:"required_if=HasSubSection false,omitempty,oneof=table chart"`
	ComponentConfig json.RawMessage `json:"component_config" validate:"omitempty"`

	NarrativePosition *string                `json:"narrative_position" validate:"omitempty,oneof=top bottom"`
	NarrativeTemplate *string                `json:"narrative_template" validate:"omitempty"`
	NarrativeLogic    *NarrativeLogicPayload `json:"narrative_logic" validate:"omitempty"`
}

// ============================================================================
// LEVEL 2: SUB-SECTION (Berisi Tabel/Grafik + Narasi)
// ============================================================================
type UpsertSubSectionPayload struct {
	Title    string `json:"title" validate:"required"`
	Sequence int    `json:"sequence" validate:"required,min=1"`

	ComponentType   string          `json:"component_type" validate:"required,oneof=table chart"`
	ComponentConfig json.RawMessage `json:"component_config" validate:"omitempty"`

	NarrativePosition *string                `json:"narrative_position" validate:"omitempty,oneof=top bottom"`
	NarrativeTemplate *string                `json:"narrative_template" validate:"omitempty"`
	NarrativeLogic    *NarrativeLogicPayload `json:"narrative_logic" validate:"omitempty"`
}

// ============================================================================
// LEVEL 3: CONFIG TABEL (JSONB Target)
// ============================================================================
type TableConfigPayload struct {
	TableStyle       string `json:"table_style" validate:"required,oneof=grouped_header simple"`
	ShowTerritoryCol bool   `json:"show_territory_col"`

	// Label kustom untuk kolom "Wilayah" (opsional, fallback ke "Wilayah" jika kosong)
	TerritoryColLabel *string `json:"territory_col_label" validate:"omitempty"`

	Groups  []TableGroupPayload   `json:"groups" validate:"omitempty,dive"`  // Hanya diisi jika style = grouped_header
	Columns []SimpleColumnPayload `json:"columns" validate:"omitempty,dive"` // Hanya diisi jika style = simple
}

type TableGroupPayload struct {
	GroupName string               `json:"group_name" validate:"required"` // Label header baris pertama (group)
	Columns   []GroupColumnPayload `json:"columns" validate:"required,min=1,dive"`
}

type GroupColumnPayload struct {
	FormFieldID int     `json:"form_field_id" validate:"required,min=1"`
	CustomLabel *string `json:"custom_label" validate:"omitempty,max=255"` // fallback ke question form_field jika kosong
}

type SimpleColumnPayload struct {
	FormFieldID int     `json:"form_field_id" validate:"required,min=1"`
	CustomLabel *string `json:"custom_label" validate:"omitempty,max=255"` // fallback ke question form_field jika kosong
}

// ============================================================================
// LEVEL 3: CONFIG NARASI (JSONB Target)
// ============================================================================
type NarrativeLogicPayload struct {
	Variables []NarrativeVariablePayload `json:"variables" validate:"required,dive"`
}

type NarrativeVariablePayload struct {
	Placeholder       string `json:"placeholder" validate:"required"`          // Contoh: "{test_variable_a}"
	SourceFormFieldID int    `json:"source_form_field_id" validate:"required"` // Contoh: 850
	CalculationType   string `json:"calculation_type" validate:"required,oneof=max_row_name max_row_value min_row_name min_row_value total_sum average most_frequent_option most_frequent_count"`
	// Perhatikan: "group_by" sudah dihapus dari struct validasi ini!
}

// ============================================================================
// LEVEL 3: CONFIG GRAFIK (Kolom flat di report_components, BUKAN JSONB)
// ============================================================================
type ChartConfigPayload struct {
	ChartType      string  `json:"chart_type" validate:"required,oneof=bar line pie"`
	ChartDirection *string `json:"chart_direction" validate:"omitempty,oneof=vertical horizontal"`
	IsMultipleData bool    `json:"is_multiple_data"`
	FormFieldIDs   []int   `json:"form_field_ids" validate:"required,min=1,dive,min=1"`
}

// KEBUTUHAN UPDATE SECTION
type UpdateSectionPayload struct {
	Title         string                    `json:"title" validate:"required,max=255"`
	HasSubSection *bool                     `json:"has_sub_section" validate:"required"`
	SubSections   []UpdateSubSectionPayload `json:"sub_sections" validate:"omitempty,dive"`

	ComponentType   string          `json:"component_type" validate:"required_if=HasSubSection false,omitempty,oneof=table chart"`
	ComponentConfig json.RawMessage `json:"component_config" validate:"omitempty"`

	NarrativePosition *string                `json:"narrative_position" validate:"omitempty,oneof=top bottom"`
	NarrativeTemplate *string                `json:"narrative_template" validate:"omitempty"`
	NarrativeLogic    *NarrativeLogicPayload `json:"narrative_logic" validate:"omitempty"`
}

// ID opsional: dikirim FE hanya untuk referensi/tracking di sisi FE (misal biar row tidak "loncat" di UI).
// Service TIDAK bergantung pada ID ini untuk logic replace — cukup diabaikan saat insert ulang.
type UpdateSubSectionPayload struct {
	ID       *int64 `json:"id,omitempty"`
	Title    string `json:"title" validate:"required"`
	Sequence int    `json:"sequence" validate:"required,min=1"`

	ComponentType   string          `json:"component_type" validate:"required,oneof=table chart"`
	ComponentConfig json.RawMessage `json:"component_config" validate:"omitempty"`

	NarrativePosition *string                `json:"narrative_position" validate:"omitempty,oneof=top bottom"`
	NarrativeTemplate *string                `json:"narrative_template" validate:"omitempty"`
	NarrativeLogic    *NarrativeLogicPayload `json:"narrative_logic" validate:"omitempty"`
}
