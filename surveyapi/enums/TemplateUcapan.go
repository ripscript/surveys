package enums

type TypeTemplateUcapan string

const (
	OPENING TypeTemplateUcapan = "opening"
	CLOSING TypeTemplateUcapan = "closing"
)

func (t TypeTemplateUcapan) IsValid() bool {
	switch t {
	case OPENING, CLOSING:
		return true
	}
	return false
}
