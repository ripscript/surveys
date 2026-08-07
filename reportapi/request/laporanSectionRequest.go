package request // atau package payloads

// ============================================================================
// LEVEL 1: SECTION (Bab Utama)
// ============================================================================
type UpsertSectionPayload struct {
	Title         string                    `json:"title" validate:"required,max=255"`
	HasSubSection bool                      `json:"has_sub_section"`
	SubSections   []UpsertSubSectionPayload `json:"sub_sections" validate:"omitempty,dive"`
}

// ============================================================================
// LEVEL 2: SUB-SECTION (Berisi Tabel/Grafik + Narasi)
// ============================================================================
type UpsertSubSectionPayload struct {
	Title    string `json:"title" validate:"required"`
	Sequence int    `json:"sequence" validate:"required,min=1"`

	// --- Visual (Tabel / Grafik) ---
	ComponentType   string              `json:"component_type" validate:"required,oneof=table chart"`
	ComponentConfig *TableConfigPayload `json:"component_config" validate:"omitempty"` // Dibuat pointer agar fleksibel jika tipe-nya 'chart'

	// --- Narasi (Opsional) ---
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

	// Gunakan omitempty agar validasi tidak bentrok
	Groups  []TableGroupPayload   `json:"groups" validate:"omitempty,dive"`  // Hanya diisi jika style = grouped_header
	Columns []SimpleColumnPayload `json:"columns" validate:"omitempty,dive"` // Hanya diisi jika style = simple
}

type TableGroupPayload struct {
	GroupName    string `json:"group_name" validate:"required"`
	FormFieldIDs []int  `json:"form_field_ids" validate:"required,min=1,dive,min=1"`
}

type SimpleColumnPayload struct {
	FormFieldID int `json:"form_field_id" validate:"required,min=1"`
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
