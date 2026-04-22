package service

import (
	"backend/siccore/pb"
	"backend/surveyapi/customValidator"
	"backend/surveyapi/models"
	"backend/surveyapi/payloads"
	"backend/surveyapi/repository"
	"backend/surveyapi/response"
	"backend/surveyapi/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/speps/go-hashids/v2"
)

type ManajemenAlurService interface {
	CreateManajemenAlur(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error)
	GetDetailManajemenAlur(usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateManajemenAlur(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DeleteManajemenAlur(usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetListManajemenAlur(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error)
	FlowPreviewIndex(usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error)
	PreviewAlurSurvey(ctx context.Context, usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type manajemenAlurService struct {
	manajemenAlurRepo              repository.ManajemenAlurRepo
	templateFormulirPertanyaanRepo repository.TemplateFormulirPertanyaanRepo
	templateUcapanRepo             repository.TemplateUcapanRepo
}

func NewManajemenAlurService(
	manajemenAlurRepo repository.ManajemenAlurRepo,
	templateFormulirPertanyaanRepo repository.TemplateFormulirPertanyaanRepo,
	templateUcapanRepo repository.TemplateUcapanRepo,
) ManajemenAlurService {
	return &manajemenAlurService{
		manajemenAlurRepo:              manajemenAlurRepo,
		templateFormulirPertanyaanRepo: templateFormulirPertanyaanRepo,
		templateUcapanRepo:             templateUcapanRepo,
	}
}

func (service *manajemenAlurService) CreateManajemenAlur(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.ManajemenAlurPayload

	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	var validate = validator.New()

	validate.RegisterStructValidation(customValidator.ManajemenAlurPayloadValidator, payloads.ManajemenAlurPayload{})

	err = validate.Struct(payload)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			customErrorMsg := utils.TranslateError(err)
			return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
		}
	}

	form, err := service.templateFormulirPertanyaanRepo.GetFormByCode(payload.TemplateFormulirPertanyaan)
	if err != nil {
		return utils.SendError(errors.New("formulir tidak ditemukan"), http.StatusNotFound)
	}

	code, err := utils.GenerateUniqueString(32)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	codeIsExist, err := service.manajemenAlurRepo.ManajemenAlurCodeIsExist(code)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	for codeIsExist {
		code, err = utils.GenerateUniqueString(32)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		codeIsExist, err = service.manajemenAlurRepo.ManajemenAlurCodeIsExist(code)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}
	}

	err = service.manajemenAlurRepo.RunInTransaction(func(txRepo repository.ManajemenAlurRepo) error {

		// PUTARAN 1: Bikin Master, Section, dan Group
		statusSec := "0"
		if payload.HasSection {
			statusSec = "1"
		}

		// A. Insert Flow Detail
		flowDetail := &models.FlowDetail{
			FormId:        form.ID,
			Name:          payload.NamaAlur,
			OpeningId:     payload.Pembuka,
			ClosingId:     payload.Penutup,
			Code:          code,
			Version:       1,
			CreatedBy:     1,
			StatusSection: statusSec,
		}

		if err := txRepo.CreateFlowDetail(flowDetail); err != nil {
			return err
		}

		// Dictionary untuk Putaran 2
		mapSectionIndexToDBID := make(map[int]int)
		mapFirstQuestionIDToGroupID := make(map[int]int)

		for _, flow := range payload.Flows {
			// B. Insert Section
			if payload.HasSection && flow.SectionIndex != nil {
				if _, exists := mapSectionIndexToDBID[*flow.SectionIndex]; !exists {
					secName := fmt.Sprintf("Section %d", *flow.SectionIndex)
					if flow.SectionName != nil && *flow.SectionName != "" {
						secName = *flow.SectionName
					}

					newSection := &models.FlowSection{Name: secName}
					if err := txRepo.CreateSection(newSection); err != nil {
						return err
					}
					mapSectionIndexToDBID[*flow.SectionIndex] = newSection.ID
				}
			}

			// C. Insert Group
			if flow.IsGroup {
				newGroup := &models.FlowGroup{Name: *flow.GroupName}
				if err := txRepo.CreateGroup(newGroup); err != nil {
					return err
				}

				mapFirstQuestionIDToGroupID[flow.QuestionIDs[0]] = int(newGroup.ID)
			}
		}

		// PUTARAN 2: Insert Flow Fields & Advanced Logic
		resolveTarget := func(targetQID *int) (childID int, groupChildID int) {
			if targetQID == nil || *targetQID == 0 {
				return 0, 0
			}
			if gID, isGroupTarget := mapFirstQuestionIDToGroupID[*targetQID]; isGroupTarget {
				return 0, gID
			}
			return *targetQID, 0
		}

		// Mulai Looping Putaran 2
		for _, flow := range payload.Flows {

			var currentSectionID *int
			if payload.HasSection && flow.SectionIndex != nil {
				if dbID, exists := mapSectionIndexToDBID[*flow.SectionIndex]; exists {
					currentSectionID = &dbID
				}
			}

			var currentGroupID int = 0
			if flow.IsGroup {
				if dbID, exists := mapFirstQuestionIDToGroupID[flow.QuestionIDs[0]]; exists {
					currentGroupID = dbID
				}
			}

			for indexQ, qID := range flow.QuestionIDs {
				questionExists, err := service.templateFormulirPertanyaanRepo.IsQuestionExistsById(qID)
				if err != nil {
					return err
				}
				if !questionExists {
					return fmt.Errorf("pertanyaan dengan ID %d tidak ditemukan", qID)
				}

				field := &models.FlowField{
					FlowDetailId:     flowDetail.ID,
					Sequence:         flow.Sequence,
					FormFieldId:      qID,
					GroupChildId:     0,
					GroupId:          currentGroupID,
					SectionId:        currentSectionID,
					Breakdown:        false,
					IsAdvancedOption: false,
				}

				// IMPLEMENTSI ROUTING
				if flow.IsGroup && indexQ < len(flow.QuestionIDs)-1 {
					nextQIDInGroup := flow.QuestionIDs[indexQ+1]

					questionExists, err := service.templateFormulirPertanyaanRepo.IsQuestionExistsById(nextQIDInGroup)
					if err != nil {
						return err
					}
					if !questionExists {
						return fmt.Errorf("pertanyaan dengan ID %d tidak ditemukan", nextQIDInGroup)
					}

					field.ChildId, field.GroupChildId = resolveTarget(&nextQIDInGroup)
				} else {
					if flow.Routing.Rule == "jump-to" {
						if flow.Routing.IsBreakdown {
							// Jump-to Breakdown
							field.Breakdown = true
							field.ChildId = 0
							field.GroupChildId = 0

							// if err := txRepo.CreateFlowField(field); err != nil {
							// 	return err
							// }

							for _, opt := range flow.Routing.Options {
								if opt.TargetQuestionID != nil && *opt.TargetQuestionID != 0 {
									questionExists, err := service.templateFormulirPertanyaanRepo.IsQuestionExistsById(*opt.TargetQuestionID)
									if err != nil {
										return err
									}
									if !questionExists {
										return fmt.Errorf("pertanyaan dengan ID %d tidak ditemukan", *opt.TargetQuestionID)
									}
								}

								child, grpChild := resolveTarget(opt.TargetQuestionID)
								optID := opt.OptionID

								if optID != 0 {
									optionQuestionExists, err := service.templateFormulirPertanyaanRepo.IsQuestionOptionsExistById(optID)
									if err != nil {
										return err
									}
									if !optionQuestionExists {
										return fmt.Errorf("opsi pertanyaan dengan ID %d tidak ditemukan", optID)
									}
								}

								optField := &models.FlowField{
									FlowDetailId:      flowDetail.ID,
									Sequence:          flow.Sequence,
									FormFieldId:       qID,
									FormAnswerFieldId: &optID,
									ChildId:           child,
									GroupChildId:      grpChild,
									GroupId:           currentGroupID,
									SectionId:         currentSectionID,
									Breakdown:         true,
								}
								if err := txRepo.CreateFlowField(optField); err != nil {
									return err
								}
							}
							continue
						} else {
							// Jump-to biasa
							if flow.Routing.TargetQuestionID != nil && *flow.Routing.TargetQuestionID != 0 {
								questionExists, err := service.templateFormulirPertanyaanRepo.IsQuestionExistsById(*flow.Routing.TargetQuestionID)
								if err != nil {
									return err
								}
								if !questionExists {
									return fmt.Errorf("pertanyaan dengan ID %d tidak ditemukan", *flow.Routing.TargetQuestionID)
								}
							}
							field.ChildId, field.GroupChildId = resolveTarget(flow.Routing.TargetQuestionID)
						}
					} else if flow.Routing.Rule == "logic" {
						// PERBAIKAN: Hanya set flag, jangan di-insert di sini!
						field.IsAdvancedOption = true
						field.ChildId = 0
						field.GroupChildId = 0
					}
				}

				// ====================================================
				// EKSEKUSI INSERT UTAMA `flow_fields`
				// ====================================================
				answerOptionIDs, err := txRepo.GetAnswerOptionsByQuestionID(qID)
				if err != nil {
					return err
				}

				// Variabel untuk menangkap ID flow_field yang berhasil dibuat
				// agar bisa disambungkan ke advanced_option_flows nanti
				var firstInsertedFieldID int

				if len(answerOptionIDs) > 0 {
					for i, optID := range answerOptionIDs {
						optIDCopy := optID
						fieldCopy := *field
						fieldCopy.FormAnswerFieldId = &optIDCopy

						if err := txRepo.CreateFlowField(&fieldCopy); err != nil {
							return err
						}

						// Ambil ID dari row pertama yang ter-insert sebagai acuan relasi logic
						if i == 0 {
							firstInsertedFieldID = fieldCopy.ID
						}
					}
				} else {
					if err := txRepo.CreateFlowField(field); err != nil {
						return err
					}
					firstInsertedFieldID = field.ID
				}

				// ====================================================
				// EKSEKUSI INSERT ADVANCED LOGIC (Jika Rule == logic)
				// ====================================================
				if flow.Routing.Rule == "logic" && !flow.IsGroup && indexQ == len(flow.QuestionIDs)-1 {
					for _, logic := range flow.Routing.Logics {
						child, _ := resolveTarget(logic.TargetQuestionID)

						if logic.TargetQuestionID != nil && *logic.TargetQuestionID != 0 {
							_questionExists, err := service.templateFormulirPertanyaanRepo.IsQuestionExistsById(*logic.TargetQuestionID)
							if err != nil {
								return err
							}
							if !_questionExists {
								return fmt.Errorf("pertanyaan target dengan ID %d tidak ditemukan", *logic.TargetQuestionID)
							}
						}

						questionExists, err := service.templateFormulirPertanyaanRepo.IsQuestionExistsById(logic.IfQuestionID)
						if err != nil {
							return err
						}
						if !questionExists {
							return fmt.Errorf("pertanyaan acuan dengan ID %d tidak ditemukan", logic.IfQuestionID)
						}

						optionQuestionExists, err := service.templateFormulirPertanyaanRepo.IsQuestionOptionsExistById(logic.IfOptionID)
						if err != nil {
							return err
						}
						if !optionQuestionExists {
							return fmt.Errorf("opsi pertanyaan acuan dengan ID %d tidak ditemukan", logic.IfOptionID)
						}

						advanced := &models.AdvancedOptionFlow{
							FlowFieldId: firstInsertedFieldID, // <-- SEKARANG MENDAPATKAN ID YANG VALID!
							FormFieldId: logic.IfQuestionID,
							Option:      logic.IfOptionID,
							ChildId:     child,
						}

						if err := txRepo.CreateAdvancedOption(advanced); err != nil {
							return err
						}
					}
				}

			}
		}

		return nil
	})

	if err != nil {
		return utils.SendError(errors.New("gagal menyimpan alur survey: "+err.Error()), http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Alur survey berhasil disimpan!")
}

func (service *manajemenAlurService) GetDetailManajemenAlur(usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	codeStr := slug["code"]
	code, ok := codeStr.(string)
	if !ok {
		return utils.SendError(errors.New("Kode alur survey tidak valid"), http.StatusBadRequest)
	}

	// 1. Ambil Master Data
	flowDetail, err := service.manajemenAlurRepo.GetFlowDetailByCode(code)
	if err != nil {
		return utils.SendError(errors.New("alur survey tidak ditemukan"), http.StatusNotFound)
	}

	// 2. Ambil semua Flow Fields
	fields, err := service.manajemenAlurRepo.GetFlowFieldsByDetailID(flowDetail.ID)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	// Jika belum ada flow-nya, langsung kembalikan kerangka dasar
	if len(fields) == 0 {
		return utils.SendData(payloads.ManajemenAlurDetailResponse{
			ID:         flowDetail.ID,
			FormCode:   flowDetail.Code,
			Name:       flowDetail.Name,
			OpeningID:  flowDetail.OpeningId,
			ClosingID:  flowDetail.ClosingId,
			HasSection: flowDetail.StatusSection == "1",
			Flows:      []payloads.FlowDetailItem{},
		}, "Detail alur berhasil diambil")
	}

	// 3. Kumpulkan ID untuk Relasi (Mencegah N+1 Query)
	var fieldIDs []int
	var sectionIDs []int
	var groupIDs []int

	uniqueSections := make(map[int]bool)
	uniqueGroups := make(map[int]bool)
	mapSequenceToFields := make(map[int][]models.FlowField)

	for _, f := range fields {
		fieldIDs = append(fieldIDs, f.ID)

		if f.SectionId != nil && !uniqueSections[*f.SectionId] {
			uniqueSections[*f.SectionId] = true
			sectionIDs = append(sectionIDs, *f.SectionId)
		}
		if f.GroupId != 0 && !uniqueGroups[f.GroupId] {
			uniqueGroups[f.GroupId] = true
			groupIDs = append(groupIDs, f.GroupId)
		}

		mapSequenceToFields[f.Sequence] = append(mapSequenceToFields[f.Sequence], f)
	}

	// 4. Tarik data relasi secara Bulk
	advancedLogics, _ := service.manajemenAlurRepo.GetAdvancedOptionsByFieldIDs(fieldIDs)
	sections, _ := service.manajemenAlurRepo.GetSectionsByIDs(sectionIDs)
	groups, _ := service.manajemenAlurRepo.GetGroupsByIDs(groupIDs)

	mapLogicsByFieldID := make(map[int][]models.AdvancedOptionFlow)
	for _, l := range advancedLogics {
		mapLogicsByFieldID[l.FlowFieldId] = append(mapLogicsByFieldID[l.FlowFieldId], l)
	}

	mapSectionNameByID := make(map[int]string)
	for _, s := range sections {
		mapSectionNameByID[s.ID] = s.Name
	}

	mapGroupNameByID := make(map[int]string)
	for _, g := range groups {
		mapGroupNameByID[g.ID] = g.Name
	}

	// =========================================================
	// 5. PROSES MERAKIT ULANG KE DTO BARU
	// =========================================================

	var reconstructedFlows []payloads.FlowDetailItem

	var sequences []int
	for seq := range mapSequenceToFields {
		sequences = append(sequences, seq)
	}
	sort.Ints(sequences) // Pastikan sudah import "sort"

	printedSections := make(map[int]bool)

	// Alat bantu untuk men-generate ulang urutan SectionIndex (1, 2, 3...)
	sectionCounter := 1
	mapSectionDBIDToIndex := make(map[int]int)

	for _, seq := range sequences {
		seqFields := mapSequenceToFields[seq]
		firstField := seqFields[0]

		flowItem := payloads.FlowDetailItem{
			Sequence: seq,
			IsGroup:  firstField.GroupId != 0,
		}

		// Kumpulkan semua ID flow_fields untuk node ini
		for _, f := range seqFields {
			flowItem.FieldIDs = append(flowItem.FieldIDs, f.ID)
		}

		// Set Section Index dan Name
		if firstField.SectionId != nil {
			secDBID := *firstField.SectionId
			flowItem.SectionID = &secDBID

			// Generate urutan 1, 2, 3...
			if _, exists := mapSectionDBIDToIndex[secDBID]; !exists {
				mapSectionDBIDToIndex[secDBID] = sectionCounter
				sectionCounter++
			}
			secIdx := mapSectionDBIDToIndex[secDBID]
			flowItem.SectionIndex = &secIdx

			if !printedSections[secDBID] {
				name := mapSectionNameByID[secDBID]
				flowItem.SectionName = &name
				printedSections[secDBID] = true
			}
		}

		// Set Group Name & ID
		if flowItem.IsGroup {
			gID := firstField.GroupId
			flowItem.GroupID = &gID
			gName := mapGroupNameByID[gID]
			flowItem.GroupName = &gName
		}

		// Kumpulkan QuestionIDs unik
		mapUniqueQID := make(map[int]bool)
		for _, f := range seqFields {
			if !mapUniqueQID[f.FormFieldId] {
				flowItem.QuestionIDs = append(flowItem.QuestionIDs, f.FormFieldId)
				mapUniqueQID[f.FormFieldId] = true
			}
		}

		// Merakit Routing
		flowItem.Routing = payloads.RoutingDetailRule{
			IsBreakdown: firstField.Breakdown,
			Options:     []payloads.RoutingOptionDetail{},
			Logics:      []payloads.RoutingLogicDetail{},
		}

		if firstField.IsAdvancedOption {
			// CASE 1: Logic
			flowItem.Routing.Rule = "logic"
			flowItem.Routing.IsEnd = false

			for _, f := range seqFields {
				if logics, exists := mapLogicsByFieldID[f.ID]; exists {
					for _, logic := range logics {

						isEnd := logic.ChildId == 0
						var target *int
						if !isEnd {
							target = &logic.ChildId
						}

						flowItem.Routing.Logics = append(flowItem.Routing.Logics, payloads.RoutingLogicDetail{
							LogicID:          logic.ID, // DB Asli
							IfQuestionID:     logic.FormFieldId,
							IfOptionID:       logic.Option,
							TargetQuestionID: target,
							IsEnd:            isEnd,
						})
					}
				}
			}

		} else if firstField.Breakdown {
			// CASE 2: Jump-to Breakdown
			flowItem.Routing.Rule = "jump-to"
			flowItem.Routing.IsEnd = false

			for _, f := range seqFields {
				if f.FormAnswerFieldId != nil {

					targetID := f.ChildId
					if f.GroupChildId != 0 {
						targetID = int(f.GroupChildId)
					}

					isEnd := targetID == 0
					var tPtr *int
					if !isEnd {
						tPtr = &targetID
					}

					flowItem.Routing.Options = append(flowItem.Routing.Options, payloads.RoutingOptionDetail{
						FieldID:          f.ID, // DB Asli flow_fields
						OptionID:         *f.FormAnswerFieldId,
						TargetQuestionID: tPtr,
						IsEnd:            isEnd,
					})
				}
			}

		} else {
			// CASE 3: Jump-to Normal
			flowItem.Routing.Rule = "jump-to"

			targetID := firstField.ChildId
			if firstField.GroupChildId != 0 {
				targetID = int(firstField.GroupChildId)
			}

			isEnd := targetID == 0

			if isEnd {
				flowItem.Routing.IsEnd = true
			} else {
				flowItem.Routing.IsEnd = false
				flowItem.Routing.TargetQuestionID = &targetID
			}
		}

		reconstructedFlows = append(reconstructedFlows, flowItem)
	}

	// 6. Bungkus menjadi Final Response menggunakan Struct Utama
	finalResponse := payloads.ManajemenAlurDetailResponse{
		ID:         flowDetail.ID,
		FormCode:   flowDetail.Code,
		Name:       flowDetail.Name,
		OpeningID:  flowDetail.OpeningId,
		ClosingID:  flowDetail.ClosingId,
		HasSection: flowDetail.StatusSection == "1",
		Flows:      reconstructedFlows,
	}

	return utils.SendData(finalResponse, "Detail manajemen alur berhasil diambil")
}

func (service *manajemenAlurService) UpdateManajemenAlur(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.ManajemenAlurUpdatePayload

	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	var validate = validator.New()
	validate.RegisterStructValidation(customValidator.ManajemenAlurUpdatePayloadValidator, payloads.ManajemenAlurUpdatePayload{})

	err = validate.Struct(payload)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			customErrorMsg := utils.TranslateError(err)
			return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
		}
	}

	codeStr := slug["code"]
	code, ok := codeStr.(string)
	if !ok {
		return utils.SendError(errors.New("Kode alur survey tidak valid"), http.StatusBadRequest)
	}

	// 1. Pastikan Master Alur Benar-benar Ada
	existingDetail, err := service.manajemenAlurRepo.GetFlowDetailByCode(code)
	if err != nil {
		return utils.SendError(errors.New("alur survey tidak ditemukan di database"), http.StatusNotFound)
	}

	err = service.manajemenAlurRepo.RunInTransaction(func(txRepo repository.ManajemenAlurRepo) error {

		statusSec := "0"
		if payload.HasSection {
			statusSec = "1"
		}

		existingDetail.Name = payload.NamaAlur
		existingDetail.OpeningId = payload.Pembuka
		existingDetail.ClosingId = payload.Penutup
		existingDetail.StatusSection = statusSec
		existingDetail.Version += 1

		if err := txRepo.UpdateFlowDetail(existingDetail); err != nil {
			return err
		}

		if err := txRepo.DeleteRoutingByDetailID(existingDetail.ID); err != nil {
			return err
		}

		mapSectionIndexToDBID := make(map[int]int)
		mapFirstQuestionIDToGroupID := make(map[int]int)

		for _, flow := range payload.Flows {

			if payload.HasSection && flow.SectionIndex != nil {
				if _, exists := mapSectionIndexToDBID[*flow.SectionIndex]; !exists {
					secName := fmt.Sprintf("Section %d", *flow.SectionIndex)
					if flow.SectionName != nil && *flow.SectionName != "" {
						secName = *flow.SectionName
					}

					secModel := &models.FlowSection{Name: secName}
					if flow.SectionID != nil && *flow.SectionID != 0 {
						secModel.ID = *flow.SectionID
					}

					if err := txRepo.SaveSection(secModel); err != nil {
						return err
					}
					mapSectionIndexToDBID[*flow.SectionIndex] = secModel.ID
				}
			}

			if flow.IsGroup {
				grpModel := &models.FlowGroup{Name: *flow.GroupName}
				if flow.GroupID != nil && *flow.GroupID != 0 {
					grpModel.ID = *flow.GroupID
				}

				if err := txRepo.SaveGroup(grpModel); err != nil {
					return err
				}
				mapFirstQuestionIDToGroupID[flow.QuestionIDs[0]] = int(grpModel.ID)
			}
		}

		resolveTarget := func(targetQID *int) (childID int, groupChildID int) {
			if targetQID == nil || *targetQID == 0 {
				return 0, 0
			}
			if gID, isGroupTarget := mapFirstQuestionIDToGroupID[*targetQID]; isGroupTarget {
				return 0, gID
			}
			return *targetQID, 0
		}

		for _, flow := range payload.Flows {
			var currentSectionID *int
			if payload.HasSection && flow.SectionIndex != nil {
				if dbID, exists := mapSectionIndexToDBID[*flow.SectionIndex]; exists {
					currentSectionID = &dbID
				}
			}

			var currentGroupID int = 0
			if flow.IsGroup {
				if dbID, exists := mapFirstQuestionIDToGroupID[flow.QuestionIDs[0]]; exists {
					currentGroupID = dbID
				}
			}

			for indexQ, qID := range flow.QuestionIDs {
				questionExists, err := service.templateFormulirPertanyaanRepo.IsQuestionExistsById(qID)
				if err != nil {
					return err
				}
				if !questionExists {
					return fmt.Errorf("pertanyaan dengan ID %d tidak ditemukan", qID)
				}

				field := &models.FlowField{
					FlowDetailId:     existingDetail.ID,
					Sequence:         flow.Sequence,
					FormFieldId:      qID,
					GroupChildId:     0,
					GroupId:          currentGroupID,
					SectionId:        currentSectionID,
					Breakdown:        false,
					IsAdvancedOption: false,
				}

				if flow.IsGroup && indexQ < len(flow.QuestionIDs)-1 {
					nextQIDInGroup := flow.QuestionIDs[indexQ+1]
					field.ChildId, field.GroupChildId = resolveTarget(&nextQIDInGroup)
				} else {
					if flow.Routing.Rule == "jump-to" {
						if flow.Routing.IsBreakdown {
							field.Breakdown = true
							field.ChildId = 0
							field.GroupChildId = 0

							for _, opt := range flow.Routing.Options {
								child, grpChild := resolveTarget(opt.TargetQuestionID)
								if child != 0 {
									questionExists, err := service.templateFormulirPertanyaanRepo.IsQuestionExistsById(child)
									if err != nil {
										return err
									}
									if !questionExists {
										return fmt.Errorf("pertanyaan dengan ID %d tidak ditemukan", child)
									}
								}

								optID := opt.OptionID

								if optID != 0 {
									optionQuestionExists, err := service.templateFormulirPertanyaanRepo.IsQuestionOptionsExistById(optID)
									if err != nil {
										return err
									}
									if !optionQuestionExists {
										return fmt.Errorf("opsi pertanyaan dengan ID %d tidak ditemukan", optID)
									}
								}

								optField := &models.FlowField{
									FlowDetailId:      existingDetail.ID,
									Sequence:          flow.Sequence,
									FormFieldId:       qID,
									FormAnswerFieldId: &optID,
									ChildId:           child,
									GroupChildId:      grpChild,
									GroupId:           currentGroupID,
									SectionId:         currentSectionID,
									Breakdown:         true,
								}
								if err := txRepo.CreateFlowField(optField); err != nil {
									return err
								}
							}
							continue
						} else {
							field.ChildId, field.GroupChildId = resolveTarget(flow.Routing.TargetQuestionID)
						}
					} else if flow.Routing.Rule == "logic" {
						field.IsAdvancedOption = true
						field.ChildId = 0
						field.GroupChildId = 0
					}
				}

				answerOptionIDs, err := txRepo.GetAnswerOptionsByQuestionID(qID)
				if err != nil {
					return err
				}

				var firstInsertedFieldID int

				if len(answerOptionIDs) > 0 {
					for i, optID := range answerOptionIDs {
						optIDCopy := optID
						fieldCopy := *field
						fieldCopy.FormAnswerFieldId = &optIDCopy

						if err := txRepo.CreateFlowField(&fieldCopy); err != nil {
							return err
						}

						if i == 0 {
							firstInsertedFieldID = fieldCopy.ID
						}
					}
				} else {
					if err := txRepo.CreateFlowField(field); err != nil {
						return err
					}
					firstInsertedFieldID = field.ID
				}

				if flow.Routing.Rule == "logic" && !flow.IsGroup && indexQ == len(flow.QuestionIDs)-1 {
					for _, logic := range flow.Routing.Logics {
						child, _ := resolveTarget(logic.TargetQuestionID)

						if logic.TargetQuestionID != nil && *logic.TargetQuestionID != 0 {
							questionExists, err := service.templateFormulirPertanyaanRepo.IsQuestionExistsById(*logic.TargetQuestionID)
							if err != nil {
								return err
							}
							if !questionExists {
								return fmt.Errorf("pertanyaan target dengan ID %d tidak ditemukan", *logic.TargetQuestionID)
							}
						}

						questionExists, err := service.templateFormulirPertanyaanRepo.IsQuestionExistsById(logic.IfQuestionID)
						if err != nil {
							return err
						}
						if !questionExists {
							return fmt.Errorf("pertanyaan acuan dengan ID %d tidak ditemukan", logic.IfQuestionID)
						}

						optionQuestionExists, err := service.templateFormulirPertanyaanRepo.IsQuestionOptionsExistById(logic.IfOptionID)
						if err != nil {
							return err
						}
						if !optionQuestionExists {
							return fmt.Errorf("opsi pertanyaan acuan dengan ID %d tidak ditemukan", logic.IfOptionID)
						}

						advanced := &models.AdvancedOptionFlow{
							FlowFieldId: firstInsertedFieldID,
							FormFieldId: logic.IfQuestionID,
							Option:      logic.IfOptionID,
							ChildId:     child,
						}
						if err := txRepo.CreateAdvancedOption(advanced); err != nil {
							return err
						}
					}
				}

			}
		}

		return nil
	})

	if err != nil {
		return utils.SendError(errors.New("gagal mengupdate alur survey: "+err.Error()), http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Alur survey berhasil diperbarui!")
}

func (service *manajemenAlurService) DeleteManajemenAlur(usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	codeStr := slug["code"]
	code, ok := codeStr.(string)
	if !ok {
		return utils.SendError(errors.New("Kode alur survey tidak valid"), http.StatusBadRequest)
	}

	existingDetail, err := service.manajemenAlurRepo.GetFlowDetailByCode(code)
	if err != nil {
		return utils.SendError(errors.New("alur survey tidak ditemukan di database"), http.StatusNotFound)
	}

	err = service.manajemenAlurRepo.RunInTransaction(func(txRepo repository.ManajemenAlurRepo) error {
		// if err := txRepo.DeleteRoutingByDetailID(existingDetail.ID); err != nil {
		// 	return err
		// }

		// if err := txRepo.DeleteFlowFieldsByDetailID(existingDetail.ID); err != nil {
		// 	return err
		// }

		if err := txRepo.DeleteAlurByID(existingDetail.ID); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return utils.SendError(errors.New("gagal menghapus alur survey: "+err.Error()), http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Alur survey berhasil dihapus!")
}

func (service *manajemenAlurService) GetListManajemenAlur(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.DatatablePayload
	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	if payload.Page <= 0 {
		payload.Page = 1
	}
	if payload.Limit <= 0 {
		payload.Limit = 25
	}

	data, totalData, err := service.manajemenAlurRepo.GetListAlur(payload)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	for i := range data {
		if data[i].CreatedBy == int(usr.ID) {
			data[i].IsMyOwn = true
		} else {
			data[i].IsMyOwn = false
		}

		isUsed := false

		if data[i].IsMyOwn && !isUsed {
			data[i].PosibleUpdate = true
			data[i].PosibleDelete = true
		} else {
			data[i].PosibleUpdate = false
			data[i].PosibleDelete = false
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

	return utils.SendData(result, "Berhasil mengambil list template formulir pertanyaan")
}

func (service *manajemenAlurService) FlowPreviewIndex(usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	codeStr := slug["code"]
	code, ok := codeStr.(string)
	if !ok {
		return utils.SendError(errors.New("Kode alur survey tidak valid"), http.StatusBadRequest)
	}

	// Mengambil flowDetail
	flowDetail, err := service.manajemenAlurRepo.GetFlowDetailByCode(code)
	if err != nil {
		return utils.SendError(errors.New("alur survey tidak ditemukan"), http.StatusNotFound)
	}

	statusSectionStr := flowDetail.StatusSection
	statusSectionInt, _ := strconv.Atoi(statusSectionStr)
	hasSectionBool := utils.IntToBool(statusSectionInt)

	var sections []models.FlowPreviewSection
	rawSections, err := service.manajemenAlurRepo.GetPreviewSectionByFlowDetailId(flowDetail.ID, statusSectionStr)
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

	for _, v := range rawSections {

		sectionId := []int{v.SectionId}

		sectionCode, err := h.Encode(sectionId)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		sections = append(sections, models.FlowPreviewSection{
			SectionCode:            &sectionCode,
			SectionName:            v.SectionName,
			TotalRequiredQuestions: v.TotalRequiredQuestions,
			TotalOptionalQuestions: v.TotalOptionalQuestions,
		})
	}

	data := models.FlowPreview{
		FlowName:   flowDetail.Name,
		HasSection: hasSectionBool,
		Sections:   sections,
	}

	return utils.SendData(data, "Berhasil mengambil data untuk preview alur survey")
}

func (service *manajemenAlurService) PreviewAlurSurvey(ctx context.Context, usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	hostUserAPI := os.Getenv("USERAPI_HOST") + ":" + os.Getenv("USERAPI_PORT")

	newSlug := map[string]interface{}{"id": strconv.FormatInt(usr.RespondentID, 10)}

	dataBytes, err := utils.HitBackend(ctx, hostUserAPI, "GET", "/respondent/:id", newSlug, nil)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	var respondentData models.DetailRespondent
	if err := json.Unmarshal(dataBytes, &respondentData); err != nil {
		return utils.SendError(errors.New("Gagal memparsing data respondent dari UserAPI"), http.StatusInternalServerError)
	}

	codeStr := slug["flow_code"]
	flowCode, ok := codeStr.(string)
	if !ok {
		return utils.SendError(errors.New("Kode alur survey tidak valid"), http.StatusBadRequest)
	}

	sectionCodeStr := slug["section_code"]
	sectionCode, ok := sectionCodeStr.(string)
	if !ok {
		return utils.SendError(errors.New("Kode bagian alur survey tidak valid"), http.StatusBadRequest)
	}

	flowDetail, err := service.manajemenAlurRepo.GetFlowDetailByCode(flowCode)
	if err != nil {
		return utils.SendError(errors.New("Alur survey tidak ditemukan"), http.StatusNotFound)
	}

	sectionID := 0
	if sectionCode != "0" && sectionCode != "" {
		hd := hashids.NewData()
		hd.Salt = os.Getenv("HASHID_SALT")
		hd.MinLength = 24

		h, err := hashids.NewWithData(hd)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		decodedIDs, err := h.DecodeWithError(sectionCode)
		if err != nil || len(decodedIDs) == 0 {
			return utils.SendError(errors.New("Kode bagian alur survey tidak valid atau dimanipulasi"), http.StatusBadRequest)
		}

		sectionID = decodedIDs[0]
	}

	rawNodes, err := service.manajemenAlurRepo.GetRawNodesForPreview(flowDetail.ID, sectionID)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	if len(rawNodes) == 0 {
		return utils.SendError(errors.New("Tidak ada pertanyaan pada alur atau bagian ini"), http.StatusNotFound)
	}

	blueprintNodes := make(map[int]response.PreviewAlurSurveyStep)
	groupMasterMap := make(map[int]int)

	flowFieldToNodeKeyMap := make(map[int]int)

	var formFieldIDs []int
	var flowFieldIDs []int

	totalRequired := 0
	totalOptional := 0
	entryNodeId := 0
	var activeSectionName *string

	// PART 1 Peta Grup & Pencatatan ID Master
	for _, raw := range rawNodes {
		formFieldIDs = append(formFieldIDs, raw.FormFieldId)
		flowFieldIDs = append(flowFieldIDs, raw.FlowFieldId)

		if raw.GroupId != nil && *raw.GroupId != 0 {
			if _, exists := groupMasterMap[*raw.GroupId]; !exists {
				groupMasterMap[*raw.GroupId] = raw.FormFieldId
			}
		}
	}

	resolveTargetID := func(childID int, groupChildID int) *int {
		if groupChildID != 0 {
			if target, ok := groupMasterMap[groupChildID]; ok {
				return &target
			}
		}
		if childID != 0 {
			return &childID
		}
		return nil
	}

	// PART 2 Bangun Kerangka Node & Kalkulasi Progress
	for i, raw := range rawNodes {
		var nodeKey int
		if raw.GroupId != nil && *raw.GroupId != 0 {
			nodeKey = groupMasterMap[*raw.GroupId]
		} else {
			nodeKey = raw.FormFieldId
		}

		flowFieldToNodeKeyMap[raw.FlowFieldId] = nodeKey

		if i == 0 {
			entryNodeId = nodeKey
			activeSectionName = raw.SectionName
		}

		step, exists := blueprintNodes[nodeKey]
		if !exists {
			stepType := "single"
			if raw.GroupId != nil && *raw.GroupId != 0 {
				stepType = "group"
			}

			step = response.PreviewAlurSurveyStep{
				StepType:  stepType,
				GroupId:   raw.GroupId,
				GroupName: raw.GroupName,
				Questions: []response.PreviewAlurSurveyQuestionDetail{},
				Routing: response.PreviewAlurSurveyRoutingDetail{
					BreakdownRoutes: make(map[int]int),
					AdvancedLogics:  []response.PreviewAlurSurveyAdvancedLogicItem{},
				},
			}
		}

		// Masukkan Pertanyaan
		isDuplicate := false
		for _, q := range step.Questions {
			if q.QuestionId == raw.FormFieldId {
				isDuplicate = true
				break
			}
		}

		if !isDuplicate {
			var expectedImageCount *int
			if raw.ImageQuantity != nil && *raw.ImageQuantity != "" {
				count, errParse := strconv.Atoi(*raw.ImageQuantity)
				if errParse == nil {
					expectedImageCount = &count
				}
			}

			step.Questions = append(step.Questions, response.PreviewAlurSurveyQuestionDetail{
				QuestionId:         raw.FormFieldId,
				Type:               raw.Template,
				Label:              raw.Label,
				IsRequired:         raw.IsRequired,
				ExpectedImageCount: expectedImageCount,
				Options:            []response.PreviewAlurSurveyOptionItem{},
			})

			if raw.IsRequired {
				totalRequired++
			} else {
				totalOptional++
			}
		}

		target := resolveTargetID(raw.ChildId, int(raw.GroupChildId))

		if raw.IsAdvancedOption {
			step.Routing.Rule = "logic"
		} else if raw.Breakdown && raw.FormAnswerFieldId != nil {
			step.Routing.Rule = "jump-to"
			if target != nil {
				step.Routing.BreakdownRoutes[*raw.FormAnswerFieldId] = *target
			}
		} else {
			step.Routing.Rule = "jump-to"
			if target == nil {
				step.Routing.IsEnd = true
			} else {
				step.Routing.DefaultNext = target
			}
		}

		blueprintNodes[nodeKey] = step
	}

	// PART 3 Sisipkan Opsi & Advanced Logics
	options, _ := service.manajemenAlurRepo.GetAnswerOptionsByQuestionIDList(formFieldIDs)
	for _, opt := range options {
		for nodeKey, step := range blueprintNodes {
			for i, q := range step.Questions {
				if q.QuestionId == opt.FormFieldId {
					step.Questions[i].Options = append(step.Questions[i].Options, response.PreviewAlurSurveyOptionItem{
						ID:    opt.ID,
						Label: opt.Option,
					})
					blueprintNodes[nodeKey] = step
				}
			}
		}
	}

	logics, _ := service.manajemenAlurRepo.GetAdvancedOptionsByFieldIDs(flowFieldIDs)
	for _, logic := range logics {
		nodeKey, valid := flowFieldToNodeKeyMap[logic.FlowFieldId]

		if valid {
			if step, exists := blueprintNodes[nodeKey]; exists {

				step.Routing.AdvancedLogics = append(step.Routing.AdvancedLogics, response.PreviewAlurSurveyAdvancedLogicItem{
					IfQuestionId:     logic.FormFieldId,
					IfOptionId:       logic.Option,
					TargetQuestionId: logic.ChildId,
				})

				blueprintNodes[nodeKey] = step
			}
		}
	}

	// PART 4 Ambil Konten Opening & Closing
	var openingMeta, closingMeta *string

	if flowDetail.OpeningId != 0 {
		if rawOpening, errOp := service.templateUcapanRepo.GetTemplateUcapanById(flowDetail.OpeningId); errOp == nil && rawOpening != nil {

			openingStr := utils.ReplaceStringRespondentVariable(rawOpening.Content, &respondentData)
			openingMeta = &openingStr
		}
	}

	if flowDetail.ClosingId != 0 {
		if rawClosing, errCl := service.templateUcapanRepo.GetTemplateUcapanById(flowDetail.ClosingId); errCl == nil && rawClosing != nil {
			closingStr := utils.ReplaceStringRespondentVariable(rawClosing.Content, &respondentData)
			closingMeta = &closingStr
		}
	}

	// FINAL Bungkus ke Data Response
	hasSectionBool := utils.StringToBool(flowDetail.StatusSection)

	var activeSecCode *string
	if hasSectionBool && sectionID != 0 {
		activeSecCode = &sectionCode
	}

	dataResponse := response.PreviewAlurSurveyBlueprintResponse{
		SurveyInfo: response.PreviewAlurSurveyInfo{
			Name:              flowDetail.Name,
			HasSection:        hasSectionBool,
			ActiveSectionCode: activeSecCode,
			ActiveSectionName: activeSectionName,
			TotalRequired:     totalRequired,
			TotalOptional:     totalOptional,
			EntryNodeId:       entryNodeId,
		},
		Opening: openingMeta,
		Closing: closingMeta,
		Nodes:   blueprintNodes,
	}

	return utils.SendData(dataResponse, "Berhasil memuat Blueprint Preview Alur Survey")
}
