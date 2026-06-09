package utils

// ToPtr mengubah string biasa menjadi pointer ke string.
func ToPtr(s string) *string {
	return &s
}

// FromPtr mengubah pointer string menjadi string biasa.
// Mengembalikan string kosong jika pointer nil.
func FromPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// FromPtrOrDefault mengembalikan nilai default jika pointer nil.
func FromPtrOrDefault(s *string, defaultVal string) string {
	if s == nil {
		return defaultVal
	}
	return *s
}

// IsNilOrEmpty mengecek apakah pointer nil atau string-nya kosong.
func IsNilOrEmpty(s *string) bool {
	return s == nil || *s == ""
}
