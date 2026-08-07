package response

import "backend/reportapi/request"

type LaporanCoverResponse struct {
	DeskripsiHalamanDepan  string  `json:"deskripsi_halaman_depan" validate:"required"`
	BackgroundHalamanDepan *string `json:"background_halaman_depan"`

	DeskripsiHalamanBelakang  string  `json:"deskripsi_halaman_belakang" validate:"required"`
	BackgroundHalamanBelakang *string `json:"background_halaman_belakang"`

	KataPengantar string `json:"kata_pengantar" validate:"required"`
}

type CoverLaporanResponse struct {
	NamaLaporan            string  `json:"nama_laporan"`
	DeskripsiHalamanDepan  *string `json:"deskripsi_halaman_depan" validate:"required"`
	BackgroundHalamanDepan *string `json:"background_halaman_depan"`

	DeskripsiHalamanBelakang  *string `json:"deskripsi_halaman_belakang" validate:"required"`
	BackgroundHalamanBelakang *string `json:"background_halaman_belakang"`

	KataPengantar *string `json:"kata_pengantar" validate:"required"`
}

// KEBUTUHAN SECTION ==============
type SectionDetailResponse struct {
	ID            int64                      `json:"id"`
	Title         string                     `json:"title"`
	HasSubSection *bool                      `json:"has_sub_section"`
	Sequence      int                        `json:"sequence"`
	SubSections   []SubSectionDetailResponse `json:"sub_sections,omitempty"`

	// Dipakai HANYA jika HasSubSection == false
	ComponentType   string                      `json:"component_type,omitempty"`
	ComponentConfig *request.TableConfigPayload `json:"component_config,omitempty"`

	NarrativePosition *string                        `json:"narrative_position,omitempty"`
	NarrativeTemplate *string                        `json:"narrative_template,omitempty"`
	NarrativeLogic    *request.NarrativeLogicPayload `json:"narrative_logic,omitempty"`
}

type SubSectionDetailResponse struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Sequence int    `json:"sequence"`

	ComponentType   string                      `json:"component_type"`
	ComponentConfig *request.TableConfigPayload `json:"component_config,omitempty"`

	NarrativePosition *string                        `json:"narrative_position,omitempty"`
	NarrativeTemplate *string                        `json:"narrative_template,omitempty"`
	NarrativeLogic    *request.NarrativeLogicPayload `json:"narrative_logic,omitempty"`
}
