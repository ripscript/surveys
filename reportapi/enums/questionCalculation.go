package enums

// IsNarrativeSource menandakan apakah tipe pertanyaan ini boleh dijadikan
// sumber data untuk Narrative (Kelompok 3 tidak boleh).
func (t QuestionType) IsNarrativeSource() bool {
	switch t {
	case LONG_ANSWER, MAPS, IMAGE_TEMPLATE:
		return false
	default:
		return true
	}
}

// CalculationType merepresentasikan jenis kalkulasi/agregasi yang bisa
// dipakai untuk variabel narasi, tergantung tipe pertanyaannya.
type CalculationType string

const (
	// Kelompok 1: Data Kuantitatif (number)
	MAX_ROW_NAME  CalculationType = "max_row_name"  // Nama wilayah dengan angka tertinggi
	MAX_ROW_VALUE CalculationType = "max_row_value" // Angka tertingginya
	MIN_ROW_NAME  CalculationType = "min_row_name"  // Nama wilayah dengan angka terendah
	MIN_ROW_VALUE CalculationType = "min_row_value" // Angka terendahnya
	TOTAL_SUM     CalculationType = "total_sum"     // Total keseluruhan
	AVERAGE       CalculationType = "average"       // Rata-rata

	// Kelompok 2: Data Kategorial (multiple-choices, dropdown, checkbox)
	MOST_FREQUENT_OPTION  CalculationType = "most_frequent_option"  // Opsi paling banyak dipilih
	MOST_FREQUENT_COUNT   CalculationType = "most_frequent_count"   // Jumlah responden opsi terbanyak
	LEAST_FREQUENT_OPTION CalculationType = "least_frequent_option" // Opsi paling sedikit dipilih
	LEAST_FREQUENT_COUNT  CalculationType = "least_frequent_count"  // Jumlah responden opsi tersedikit
)

func (c CalculationType) IsValid() bool {
	switch c {
	case MAX_ROW_NAME, MAX_ROW_VALUE, MIN_ROW_NAME, MIN_ROW_VALUE, TOTAL_SUM, AVERAGE,
		MOST_FREQUENT_OPTION, MOST_FREQUENT_COUNT, LEAST_FREQUENT_OPTION, LEAST_FREQUENT_COUNT:
		return true
	}
	return false
}

// QuestionTypeCalculationTypes adalah mapping QuestionType -> daftar
// CalculationType yang valid untuk tipe pertanyaan tersebut.
// Tipe di Kelompok 3 (long-answer, maps, image-template) sengaja tidak
// dimasukkan / bernilai slice kosong, karena tidak boleh dijadikan
// sumber data agregat untuk Narrative.
var QuestionTypeCalculationTypes = map[QuestionType][]CalculationType{
	// Kelompok 1: Data Kuantitatif
	NUMBER: {
		MAX_ROW_NAME,
		MAX_ROW_VALUE,
		MIN_ROW_NAME,
		MIN_ROW_VALUE,
		TOTAL_SUM,
		AVERAGE,
	},

	// Kelompok 2: Data Kategorial
	MULTIPLE_CHOICES: {
		MOST_FREQUENT_OPTION,
		MOST_FREQUENT_COUNT,
		LEAST_FREQUENT_OPTION,
		LEAST_FREQUENT_COUNT,
	},

	// Kelompok 3: Data Tidak Terstruktur — tidak ada calculation_type yang valid
	LONG_ANSWER:    {},
	MAPS:           {},
	IMAGE_TEMPLATE: {},
}

// GetCalculationTypesByQuestionType mengembalikan daftar CalculationType
// yang valid untuk sebuah QuestionType. Berguna buat isi dropdown FE.
func GetCalculationTypesByQuestionType(qt QuestionType) []CalculationType {
	if types, ok := QuestionTypeCalculationTypes[qt]; ok {
		return types
	}
	return []CalculationType{}
}

// IsCalculationTypeValidForQuestionType mengecek apakah kombinasi
// QuestionType + CalculationType itu valid (dipakai untuk validasi saat
// user membuat/menyimpan Narrative).
func IsCalculationTypeValidForQuestionType(qt QuestionType, ct CalculationType) bool {
	types, ok := QuestionTypeCalculationTypes[qt]
	if !ok {
		return false
	}
	for _, t := range types {
		if t == ct {
			return true
		}
	}
	return false
}

var CalculationTypeLabel = map[CalculationType]string{
	MAX_ROW_NAME:  "Nama Wilayah Tertinggi",
	MAX_ROW_VALUE: "Angka Tertinggi",
	MIN_ROW_NAME:  "Nama Wilayah Terendah",
	MIN_ROW_VALUE: "Angka Terendah",
	TOTAL_SUM:     "Total Keseluruhan",
	AVERAGE:       "Rata-rata",

	MOST_FREQUENT_OPTION:  "Opsi Terbanyak Dipilih",
	MOST_FREQUENT_COUNT:   "Jumlah Responden Opsi Terbanyak",
	LEAST_FREQUENT_OPTION: "Opsi Tersedikit Dipilih",
	LEAST_FREQUENT_COUNT:  "Jumlah Responden Opsi Tersedikit",
}
