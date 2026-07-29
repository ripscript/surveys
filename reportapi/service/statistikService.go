package service

import (
	"backend/reportapi/enums"
	"backend/reportapi/models"
	"backend/reportapi/repository"
	"backend/reportapi/request"
	"backend/reportapi/utils"
	"backend/siccore/pb"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/speps/go-hashids/v2"
	"github.com/xuri/excelize/v2"
)

type StatistikService interface {
	GetStatistikIndex(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetStatistik(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ExportExcelStatistik(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type statistikService struct {
	userRepo      repository.UserRepo
	statistikRepo repository.StatistikRepo
}

func NewStatistikService(
	statistikRepo repository.StatistikRepo,
	userRepo repository.UserRepo,
) StatistikService {
	return &statistikService{
		statistikRepo: statistikRepo,
		userRepo:      userRepo,
	}
}

func (service *statistikService) GetStatistikIndex(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	search := param.Get("search")
	page, _ := strconv.Atoi(param.Get("page"))
	limit, _ := strconv.Atoi(param.Get("limit"))
	orderBy := param.Get("order_by")
	orderDir := param.Get("order_dir")

	statusSurvey := param.Get("status_survey")

	payload := request.SurveyWilayahDatatablePayload{
		Search:       search,
		Page:         page,
		Limit:        limit,
		OrderBy:      orderBy,
		OrderDir:     orderDir,
		StatusSurvey: statusSurvey,
	}

	if payload.Page <= 0 {
		payload.Page = 1
	}
	if payload.Limit <= 0 {
		payload.Limit = 5
	}

	respondentLogin, err := service.userRepo.GetRespondentById(ctx, usr.RespondentID)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	// spew.Dump(respondentLogin)

	permittedRoles := []int64{int64(enums.ROLE_ADMIN), int64(enums.ROLE_KECAMATAN), int64(enums.ROLE_KELURAHAN), int64(enums.ROLE_RW)}

	if respondentLogin.RoleId != nil && !utils.ContainsInt64(permittedRoles, *respondentLogin.RoleId) {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	data, totalData, err := service.statistikRepo.GetListAvailableSurvey(usr, respondentLogin, payload)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24

	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	for i := range data {
		surveyId := []int{data[i].ID}
		surveyCode, err := h.Encode(surveyId)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		data[i].SurveyCode = surveyCode

		now := time.Now()

		surveyDimulai := utils.ParseToWIB(data[i].SurveyDimulai)
		surveyBerakhir := utils.ParseToWIB(data[i].SurveyBerakhir)

		if surveyDimulai.After(now) {
			data[i].Status = string(enums.STATUS_SURVEY_UPCOMING)
		} else if surveyDimulai.Before(now) && surveyBerakhir.After(now) {
			data[i].Status = string(enums.STATUS_SURVEY_ONGOING)
		} else if surveyBerakhir.Before(now) {
			data[i].Status = string(enums.STATUS_SURVEY_FINISHED)
		}
	}

	showingFrom := (payload.Page-1)*payload.Limit + 1
	showingTo := showingFrom + len(data) - 1

	if totalData == 0 {
		showingFrom = 0
		showingTo = 0
	}

	result := map[string]interface{}{
		"data": data,
		"meta": map[string]interface{}{
			"total_entries": totalData,
			"current_page":  payload.Page,
			"per_page":      payload.Limit,
			"showing_from":  showingFrom,
			"showing_to":    showingTo,
		},
	}

	return utils.SendData(result, "Survey wilayah berhasil diambil")
}

func (service *statistikService) GetStatistik(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	codeStr, ok := slug["survey_code"].(string)
	if !ok {
		return utils.SendError(errors.New("Survey tidak valid"), http.StatusBadRequest)
	}

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24

	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	decodedSurveyIDs, err := h.DecodeWithError(codeStr)
	if err != nil || len(decodedSurveyIDs) == 0 {
		return utils.SendError(errors.New("Survey tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}
	surveyId := decodedSurveyIDs[0]

	// ambil data login untuk role-based scoping
	respondentLogin, err := service.userRepo.GetRespondentById(ctx, usr.RespondentID)
	if err != nil || respondentLogin.RoleId == nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	roleId := enums.RoleID(*respondentLogin.RoleId)

	if roleId == enums.ROLE_RT {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses ke fitur ini"), http.StatusUnauthorized)
	}

	typeWilayahStr := param.Get("type_wilayah")
	codeWilayah := param.Get("code_wilayah")

	typeWilayah := 0
	var parentId int64 = 0

	if typeWilayahStr != "" || codeWilayah != "" {
		if typeWilayahStr == "" || codeWilayah == "" {
			return utils.SendError(errors.New("type_wilayah dan code_wilayah harus diisi bersamaan"), http.StatusBadRequest)
		}

		parsedType, convErr := strconv.Atoi(typeWilayahStr)
		if convErr != nil {
			return utils.SendError(errors.New("type_wilayah tidak valid"), http.StatusBadRequest)
		}
		typeWilayah = parsedType

		switch enums.RoleID(typeWilayah) {
		case enums.ROLE_RT:
			return utils.SendError(errors.New("Wilayah ini sudah level terakhir (RT)"), http.StatusBadRequest)
		case enums.ROLE_RW, enums.ROLE_KELURAHAN, enums.ROLE_KECAMATAN:
			// valid, lanjut
		default:
			return utils.SendError(errors.New("type_wilayah tidak valid"), http.StatusBadRequest)
		}

		decodedIds, decErr := h.DecodeWithError(codeWilayah)
		if decErr != nil || len(decodedIds) == 0 {
			return utils.SendError(errors.New("code_wilayah tidak valid atau dimanipulasi"), http.StatusBadRequest)
		}
		parentId = int64(decodedIds[0])
	}

	// role-based scoping: override / validasi typeWilayah & parentId sesuai wilayah milik user login
	switch roleId {
	case enums.ROLE_ADMIN, enums.ROLE_PEMERINTAH_KOTA, enums.ROLE_WALIKOTA:
		// akses penuh: kosong = root (kecamatan), bebas drilldown ke wilayah manapun

	case enums.ROLE_KECAMATAN:
		if respondentLogin.KecamatanId == nil {
			return utils.SendError(errors.New("Data wilayah login tidak lengkap"), http.StatusUnauthorized)
		}
		ownKecamatanId := *respondentLogin.KecamatanId

		if typeWilayahStr == "" {
			// default: langsung tampilkan daftar kelurahan di kecamatan sendiri
			typeWilayah = int(enums.ROLE_KECAMATAN)
			parentId = ownKecamatanId
		} else {
			if typeWilayah == int(enums.ROLE_KECAMATAN) {
				return utils.SendError(errors.New("Anda tidak memiliki hak akses ke wilayah ini"), http.StatusForbidden)
			}
			kecId, _, _, ancErr := service.statistikRepo.GetWilayahAncestry(typeWilayah, parentId)
			if ancErr != nil || kecId != ownKecamatanId {
				return utils.SendError(errors.New("Anda tidak memiliki hak akses ke wilayah ini"), http.StatusForbidden)
			}
		}

	case enums.ROLE_KELURAHAN:
		if respondentLogin.KelurahanId == nil {
			return utils.SendError(errors.New("Data wilayah login tidak lengkap"), http.StatusUnauthorized)
		}
		ownKelurahanId := *respondentLogin.KelurahanId

		if typeWilayahStr == "" {
			// default: langsung tampilkan daftar rw di kelurahan sendiri
			typeWilayah = int(enums.ROLE_KELURAHAN)
			parentId = ownKelurahanId
		} else {
			if typeWilayah != int(enums.ROLE_RW) {
				return utils.SendError(errors.New("Anda tidak memiliki hak akses ke wilayah ini"), http.StatusForbidden)
			}
			_, kelId, _, ancErr := service.statistikRepo.GetWilayahAncestry(typeWilayah, parentId)
			if ancErr != nil || kelId != ownKelurahanId {
				return utils.SendError(errors.New("Anda tidak memiliki hak akses ke wilayah ini"), http.StatusForbidden)
			}
		}

	case enums.ROLE_RW:
		if respondentLogin.RWId == nil {
			return utils.SendError(errors.New("Data wilayah login tidak lengkap"), http.StatusUnauthorized)
		}
		// RW cuma bisa lihat RT di RW-nya sendiri; abaikan param FE, paksa ke RW sendiri
		typeWilayah = int(enums.ROLE_RW)
		parentId = *respondentLogin.RWId

	default:
		return utils.SendError(errors.New("Anda tidak memiliki hak akses ke fitur ini"), http.StatusUnauthorized)
	}

	search := param.Get("search")
	page, _ := strconv.Atoi(param.Get("page"))
	limit, _ := strconv.Atoi(param.Get("limit"))
	orderBy := param.Get("order_by")
	orderDir := param.Get("order_dir")

	payload := request.StatistikKewilayahanPayload{
		SurveyId:    surveyId,
		TypeWilayah: typeWilayah,
		ParentId:    parentId,
		Search:      search,
		Page:        page,
		Limit:       limit,
		OrderBy:     orderBy,
		OrderDir:    orderDir,
	}

	if payload.Page <= 0 {
		payload.Page = 1
	}
	if payload.Limit <= 0 {
		payload.Limit = 5
	}

	data, totalData, aggregate, err := service.statistikRepo.GetStatistikKewilayahan(payload)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	var childTypeWilayah int
	switch typeWilayah {
	case 0:
		childTypeWilayah = int(enums.ROLE_KECAMATAN) // 5
	case int(enums.ROLE_KECAMATAN):
		childTypeWilayah = int(enums.ROLE_KELURAHAN) // 4
	case int(enums.ROLE_KELURAHAN):
		childTypeWilayah = int(enums.ROLE_RW) // 3
	case int(enums.ROLE_RW):
		childTypeWilayah = int(enums.ROLE_RT) // 2
	}

	for i := range data {
		wc, encErr := h.Encode([]int{int(data[i].ID)})
		if encErr != nil {
			return utils.SendError(encErr, http.StatusInternalServerError)
		}
		data[i].WilayahCode = wc
		data[i].TypeWilayah = childTypeWilayah
		data[i].IsLeaf = childTypeWilayah == int(enums.ROLE_RT)
	}

	showingFrom := (payload.Page-1)*payload.Limit + 1
	showingTo := showingFrom + len(data) - 1

	if totalData == 0 {
		showingFrom = 0
		showingTo = 0
	}

	result := map[string]interface{}{
		"data": data,
		"meta": map[string]interface{}{
			"total_entries":       totalData,
			"current_page":        payload.Page,
			"per_page":            payload.Limit,
			"showing_from":        showingFrom,
			"showing_to":          showingTo,
			"type_wilayah":        typeWilayah,
			"total_wilayah":       aggregate.TotalWilayah,
			"total_sudah_mengisi": aggregate.TotalSudahMengisi,
			"total_belum_mengisi": aggregate.TotalBelumMengisi,
		},
	}

	return utils.SendData(result, "Statistik kewilayahan berhasil diambil")
}

func (service *statistikService) ExportExcelStatistik(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	codeStr, ok := slug["survey_code"].(string)
	if !ok {
		return utils.SendError(errors.New("Survey tidak valid"), http.StatusBadRequest)
	}

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24

	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	decodedSurveyIDs, err := h.DecodeWithError(codeStr)
	if err != nil || len(decodedSurveyIDs) == 0 {
		return utils.SendError(errors.New("Survey tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}
	surveyId := decodedSurveyIDs[0]

	respondentLogin, err := service.userRepo.GetRespondentById(ctx, usr.RespondentID)
	if err != nil || respondentLogin.RoleId == nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	roleId := enums.RoleID(*respondentLogin.RoleId)

	if roleId == enums.ROLE_RT {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses ke fitur ini"), http.StatusUnauthorized)
	}

	typeWilayahStr := param.Get("type_wilayah")
	codeWilayah := param.Get("code_wilayah")

	typeWilayah := 0
	var parentId int64 = 0

	if typeWilayahStr != "" || codeWilayah != "" {
		if typeWilayahStr == "" || codeWilayah == "" {
			return utils.SendError(errors.New("type_wilayah dan code_wilayah harus diisi bersamaan"), http.StatusBadRequest)
		}

		parsedType, convErr := strconv.Atoi(typeWilayahStr)
		if convErr != nil {
			return utils.SendError(errors.New("type_wilayah tidak valid"), http.StatusBadRequest)
		}
		typeWilayah = parsedType

		switch enums.RoleID(typeWilayah) {
		case enums.ROLE_RT:
			return utils.SendError(errors.New("Wilayah ini sudah level terakhir (RT)"), http.StatusBadRequest)
		case enums.ROLE_RW, enums.ROLE_KELURAHAN, enums.ROLE_KECAMATAN:
			// valid, lanjut
		default:
			return utils.SendError(errors.New("type_wilayah tidak valid"), http.StatusBadRequest)
		}

		decodedIds, decErr := h.DecodeWithError(codeWilayah)
		if decErr != nil || len(decodedIds) == 0 {
			return utils.SendError(errors.New("code_wilayah tidak valid atau dimanipulasi"), http.StatusBadRequest)
		}
		parentId = int64(decodedIds[0])
	}

	// 3. Override filter berdasarkan hak akses login user (sama dengan GetStatistik)
	switch roleId {
	case enums.ROLE_ADMIN, enums.ROLE_PEMERINTAH_KOTA, enums.ROLE_WALIKOTA:
		// Akses penuh
	case enums.ROLE_KECAMATAN:
		if respondentLogin.KecamatanId == nil {
			return utils.SendError(errors.New("Data wilayah login tidak lengkap"), http.StatusUnauthorized)
		}
		ownKecamatanId := *respondentLogin.KecamatanId
		if typeWilayahStr == "" {
			typeWilayah = int(enums.ROLE_KECAMATAN)
			parentId = ownKecamatanId
		} else {
			if typeWilayah == int(enums.ROLE_KECAMATAN) {
				return utils.SendError(errors.New("Anda tidak memiliki hak akses ke wilayah ini"), http.StatusForbidden)
			}
			kecId, _, _, ancErr := service.statistikRepo.GetWilayahAncestry(typeWilayah, parentId)
			if ancErr != nil || kecId != ownKecamatanId {
				return utils.SendError(errors.New("Anda tidak memiliki hak akses ke wilayah ini"), http.StatusForbidden)
			}
		}
	case enums.ROLE_KELURAHAN:
		if respondentLogin.KelurahanId == nil {
			return utils.SendError(errors.New("Data wilayah login tidak lengkap"), http.StatusUnauthorized)
		}
		ownKelurahanId := *respondentLogin.KelurahanId
		if typeWilayahStr == "" {
			typeWilayah = int(enums.ROLE_KELURAHAN)
			parentId = ownKelurahanId
		} else {
			if typeWilayah != int(enums.ROLE_RW) {
				return utils.SendError(errors.New("Anda tidak memiliki hak akses ke wilayah ini"), http.StatusForbidden)
			}
			_, kelId, _, ancErr := service.statistikRepo.GetWilayahAncestry(typeWilayah, parentId)
			if ancErr != nil || kelId != ownKelurahanId {
				return utils.SendError(errors.New("Anda tidak memiliki hak akses ke wilayah ini"), http.StatusForbidden)
			}
		}
	case enums.ROLE_RW:
		if respondentLogin.RWId == nil {
			return utils.SendError(errors.New("Data wilayah login tidak lengkap"), http.StatusUnauthorized)
		}
		typeWilayah = int(enums.ROLE_RW)
		parentId = *respondentLogin.RWId
	default:
		return utils.SendError(errors.New("Anda tidak memiliki hak akses ke fitur ini"), http.StatusUnauthorized)
	}

	// 4. Tarik data dari database
	data, err := service.statistikRepo.GetExportStatistikExcel(int(surveyId), typeWilayah, parentId)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	// 5. Inisialisasi dan tulis file Excel
	f := excelize.NewFile()
	defer f.Close()
	sheetName := "Statistik Survey"
	if errRename := f.SetSheetName("Sheet1", sheetName); errRename != nil {
		sheetName = "Sheet1"
	}

	// Buat Style untuk Header (Background #15406a, Font Putih & Bold, Border Lengkap, Teks Tengah)
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#15406a"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	// Buat Style untuk Data (Border Lengkap)
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	// Tulis Header & inisialisasi variabel untuk menghitung lebar kolom proporsional
	headers := []string{"Kecamatan", "Kelurahan", "RW", "Jumlah RT", "RT Sudah Mengisi", "RT Belum Mengisi"}
	colWidths := make(map[int]int)

	for i, headerText := range headers {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		cell := fmt.Sprintf("%s1", colName)
		f.SetCellValue(sheetName, cell, headerText)

		// Inisialisasi panjang string tiap kolom (berdasarkan panjang header text)
		colWidths[i] = len(headerText)
	}

	// Terapkan style ke baris header A1 - F1
	f.SetCellStyle(sheetName, "A1", "F1", headerStyle)

	// Tulis Baris Data
	for i, rowData := range data {
		rowIndex := i + 2
		rtBelumMengisi := rowData.JumlahRt - rowData.RtSudahMengisi
		if rtBelumMengisi < 0 {
			rtBelumMengisi = 0
		}

		// Array data untuk mempermudah looping penulisan per baris
		rowValues := []interface{}{
			rowData.Kecamatan,
			rowData.Kelurahan,
			rowData.Rw,
			rowData.JumlahRt,
			rowData.RtSudahMengisi,
			rtBelumMengisi,
		}

		for colIndex, val := range rowValues {
			colName, _ := excelize.ColumnNumberToName(colIndex + 1)
			cell := fmt.Sprintf("%s%d", colName, rowIndex)
			f.SetCellValue(sheetName, cell, val)

			// Kalkulasi panjang data teks untuk menyesuaikan width kolom nanti
			strVal := fmt.Sprintf("%v", val)
			if len(strVal) > colWidths[colIndex] {
				colWidths[colIndex] = len(strVal)
			}
		}
	}

	// Terapkan style data (border) untuk seluruh baris data yang terisi
	if len(data) > 0 {
		lastRowCell := fmt.Sprintf("F%d", len(data)+1)
		f.SetCellStyle(sheetName, "A2", lastRowCell, dataStyle)
	}

	// Atur lebar tiap kolom agar proporsional (menggunakan data colWidths tadi)
	for i, width := range colWidths {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		// Ditambah offset (+2) agar ada jarak / tidak terlalu mepet dengan border
		f.SetColWidth(sheetName, colName, colName, float64(width)+2.0)
	}

	// 6. Simpan Excel ke buffer lalu encode ke Base64 string
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return utils.SendError(errors.New("Gagal men-generate file excel"), http.StatusInternalServerError)
	}

	mimeType := "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	return utils.SetResponseData(buf.Bytes(), true, "Data File,"+mimeType, http.StatusOK, nil, ""), nil
}
