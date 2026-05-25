package service

import (
	"backend/siccore/pb"
	"backend/surveyapi/customValidator"
	"backend/surveyapi/enums"
	"backend/surveyapi/models"
	"backend/surveyapi/payloads"
	"backend/surveyapi/repository"
	"backend/surveyapi/response"
	"backend/surveyapi/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/speps/go-hashids/v2"
	"gorm.io/gorm"
)

type ManajemenAlurService interface {
	CreateManajemenAlur(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error)
	GetDetailManajemenAlur(usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateManajemenAlur(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DeleteManajemenAlur(usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetListManajemenAlur(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	FlowPreviewIndex(usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error)
	PreviewAlurSurvey(ctx context.Context, usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error)
	AlurOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
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

	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}
	err = json.Unmarshal(jsonBytes, &payload)
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

	if payload.TemplateFormulirPertanyaan == nil || *payload.TemplateFormulirPertanyaan == "" {
		return utils.SendError(errors.New("Template formulir pertanyaan tidak boleh kosong"), http.StatusBadRequest)
	}

	form, err := service.templateFormulirPertanyaanRepo.GetFormByCode(*payload.TemplateFormulirPertanyaan)
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
		alurByName, err := txRepo.GetFlowDetailByNameCaseInsensitive(payload.NamaAlur)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if alurByName != nil {
			return utils.NewClientError(fmt.Sprintf("nama alur survey '%s' sudah digunakan, silakan gunakan nama lain", payload.NamaAlur))
		}

		statusSec := "0"
		if payload.HasSection {
			statusSec = "1"
		}

		flowDetail := &models.FlowDetail{
			FormId:        form.ID,
			Name:          payload.NamaAlur,
			OpeningId:     payload.Pembuka,
			ClosingId:     payload.Penutup,
			Code:          code,
			Version:       1,
			CreatedBy:     int(usr.ID),
			StatusSection: statusSec,
		}

		if err := txRepo.CreateFlowDetail(flowDetail); err != nil {
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
					newSection := &models.FlowSection{Name: secName}
					if err := txRepo.CreateSection(newSection); err != nil {
						return err
					}
					mapSectionIndexToDBID[*flow.SectionIndex] = newSection.ID
				}
			}

			if flow.IsGroup {
				newGroup := &models.FlowGroup{Name: *flow.GroupName}
				if err := txRepo.CreateGroup(newGroup); err != nil {
					return err
				}
				mapFirstQuestionIDToGroupID[flow.QuestionIDs[0]] = int(newGroup.ID)
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

			mainRuleVal := "jump-to"
			if flow.Routing.Rule != nil && *flow.Routing.Rule != "" {
				mainRuleVal = *flow.Routing.Rule
			}

			for indexQ, qID := range flow.QuestionIDs {
				formFieldDetail, err := service.templateFormulirPertanyaanRepo.GetFormFieldByID(qID)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return utils.NewClientError(fmt.Sprintf("pertanyaan dengan ID %d tidak ditemukan", qID))
					}
					return err
				}

				if formFieldDetail == nil {
					return utils.NewClientError(fmt.Sprintf("pertanyaan dengan ID %d tidak ditemukan", qID))
				}

				if formFieldDetail.Template != string(enums.MULTIPLE_CHOICES) {
					if len(flow.Routing.Options) > 0 {
						return utils.NewClientError(fmt.Sprintf("pertanyaan dengan ID %d tidak boleh memiliki opsi jawaban karena bukan bertipe multiple choices", qID))
					}
				}

				if flow.Routing.Rule != nil {
					if *flow.Routing.Rule != string(enums.ROUTING_RULE_LOGIC) {
						if len(flow.Routing.Logics) > 0 {
							return utils.NewClientError(
								fmt.Sprintf("pertanyaan dengan ID %d tidak boleh memiliki logics", qID),
							)
						}
					} else if *flow.Routing.Rule == string(enums.ROUTING_RULE_LOGIC) {
						if len(flow.Routing.Logics) < 1 {
							return utils.NewClientError(
								fmt.Sprintf("pertanyaan dengan ID %d harus setidaknya memiliki 1 logics", qID),
							)
						}
					}
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

				if flow.IsGroup {
					if formFieldDetail.Template != string(enums.MULTIPLE_CHOICES) {
						return utils.NewClientError(fmt.Sprintf("pertanyaan dengan ID %d harus bertipe Multiple Choices untuk digunakan dalam group", qID))
					}

					if len(flow.Routing.Options) > 0 {
						return utils.NewClientError(
							fmt.Sprintf("pertanyaan dengan ID %d bertipe group, tidak boleh memiliki opsi jawaban", qID),
						)
					}

					if len(flow.Routing.Logics) > 0 {
						return utils.NewClientError(
							fmt.Sprintf("pertanyaan dengan ID %d bertipe group, tidak boleh memiliki logika", qID),
						)
					}
				}

				if flow.IsGroup && indexQ < len(flow.QuestionIDs)-1 {
					nextQIDInGroup := flow.QuestionIDs[indexQ+1]
					field.ChildId, field.GroupChildId = resolveTarget(&nextQIDInGroup)
				} else {
					if flow.Routing.IsBreakdown {
						field.Breakdown = true
						field.ChildId = 0
						field.GroupChildId = 0

						for _, opt := range flow.Routing.Options {
							var child, grpChild int
							isAdvanced := false

							optRuleVal := "jump-to"
							if opt.Rule != nil && *opt.Rule != "" {
								optRuleVal = *opt.Rule
							}

							if optRuleVal == "logic" {
								isAdvanced = true
								child = 0
								grpChild = 0

								if len(opt.Logics) > 0 {
									for _, logic := range opt.Logics {
										formFieldDetail, err := service.templateFormulirPertanyaanRepo.GetFormFieldByID(logic.IfQuestionID)
										if err != nil {
											if errors.Is(err, gorm.ErrRecordNotFound) {
												return utils.NewClientError(fmt.Sprintf("pertanyaan dengan ID %d tidak ditemukan", logic.IfQuestionID))
											}
											return err
										}

										if formFieldDetail == nil {
											return utils.NewClientError(fmt.Sprintf("pertanyaan dengan ID %d tidak ditemukan", logic.IfQuestionID))
										}

										if formFieldDetail.Template != string(enums.MULTIPLE_CHOICES) {
											return utils.NewClientError(fmt.Sprintf("pertanyaan dengan ID %d harus bertipe Multiple Choices untuk digunakan dalam logic", logic.IfQuestionID))
										}
									}
								}

							} else {
								if opt.TargetQuestionID != nil && *opt.TargetQuestionID != 0 {
									questionExists, err := service.templateFormulirPertanyaanRepo.IsQuestionExistsById(*opt.TargetQuestionID)
									if err != nil {
										return err
									}
									if !questionExists {
										return utils.NewClientError(fmt.Sprintf("pertanyaan dengan ID %d tidak ditemukan", *opt.TargetQuestionID))
									}
								}
								child, grpChild = resolveTarget(opt.TargetQuestionID)
							}

							optID := opt.OptionID

							if optID != 0 {
								optionQuestionExists, err := service.templateFormulirPertanyaanRepo.IsQuestionOptionsExistById(optID)
								if err != nil {
									return err
								}
								if !optionQuestionExists {
									return utils.NewClientError(fmt.Sprintf("opsi pertanyaan dengan ID %d tidak ditemukan", optID))
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
								IsAdvancedOption:  isAdvanced,
							}

							if err := txRepo.CreateFlowField(optField); err != nil {
								return err
							}

							if optRuleVal == "logic" {
								for _, logic := range opt.Logics {
									logicTargetID := 0
									if logic.TargetQuestionID != nil && *logic.TargetQuestionID != 0 {
										_questionExists, err := service.templateFormulirPertanyaanRepo.IsQuestionExistsById(*logic.TargetQuestionID)
										if err != nil {
											return err
										}
										if !_questionExists {
											return utils.NewClientError("pertanyaan target opsi logic tidak ditemukan")
										}
										logicTargetID = *logic.TargetQuestionID
									}

									advanced := &models.AdvancedOptionFlow{
										FlowFieldId: optField.ID,
										FormFieldId: logic.IfQuestionID,
										Option:      logic.IfOptionID,
										ChildId:     logicTargetID,
									}

									if err := txRepo.CreateAdvancedOption(advanced); err != nil {
										return err
									}
								}
							}
						}
						continue
					} else {
						switch mainRuleVal {
						case "jump-to":
							if flow.Routing.TargetQuestionID != nil && *flow.Routing.TargetQuestionID != 0 {
								questionExists, err := service.templateFormulirPertanyaanRepo.IsQuestionExistsById(*flow.Routing.TargetQuestionID)
								if err != nil {
									return err
								}

								if !questionExists {
									return utils.NewClientError(fmt.Sprintf("pertanyaan dengan ID %d tidak ditemukan", *flow.Routing.TargetQuestionID))
								}
							}
							field.ChildId, field.GroupChildId = resolveTarget(flow.Routing.TargetQuestionID)
						case "logic":
							field.IsAdvancedOption = true
							field.ChildId = 0
							field.GroupChildId = 0
						}
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

				if mainRuleVal == "logic" && !flow.Routing.IsBreakdown && !flow.IsGroup && indexQ == len(flow.QuestionIDs)-1 {
					for _, logic := range flow.Routing.Logics {
						if logic.IfQuestionID != 0 {
							formFieldDetail, err := service.templateFormulirPertanyaanRepo.GetFormFieldByID(logic.IfQuestionID)
							if err != nil {
								if errors.Is(err, gorm.ErrRecordNotFound) {
									return utils.NewClientError(fmt.Sprintf("pertanyaan dengan ID %d tidak ditemukan", logic.IfQuestionID))
								}
								return err
							}

							if formFieldDetail == nil {
								return utils.NewClientError(fmt.Sprintf("pertanyaan dengan ID %d tidak ditemukan", logic.IfQuestionID))
							}

							if formFieldDetail.Template != string(enums.MULTIPLE_CHOICES) {
								return utils.NewClientError(fmt.Sprintf("pertanyaan dengan ID %d harus bertipe Multiple Choices untuk digunakan dalam logic", logic.IfQuestionID))
							}
						}

						logicTargetID := 0
						if logic.TargetQuestionID != nil && *logic.TargetQuestionID != 0 {
							_questionExists, err := service.templateFormulirPertanyaanRepo.IsQuestionExistsById(*logic.TargetQuestionID)
							if err != nil {
								return err
							}

							if !_questionExists {
								return utils.NewClientError(fmt.Sprintf("pertanyaan target dengan ID %d tidak ditemukan", *logic.TargetQuestionID))
							}
							logicTargetID = *logic.TargetQuestionID
						}

						advanced := &models.AdvancedOptionFlow{
							FlowFieldId: firstInsertedFieldID,
							FormFieldId: logic.IfQuestionID,
							Option:      logic.IfOptionID,
							ChildId:     logicTargetID,
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
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			return utils.SendError(err, appErr.Code)
		}
		return utils.SendError(err, http.StatusInternalServerError)
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

	flowDetail, err := service.manajemenAlurRepo.GetFlowDetailByCode(code)
	if err != nil {
		return utils.SendError(errors.New("alur survey tidak ditemukan"), http.StatusNotFound)
	}

	fields, err := service.manajemenAlurRepo.GetFlowFieldsByDetailID(flowDetail.ID)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	form, err := service.templateFormulirPertanyaanRepo.GetFormById(int64(flowDetail.FormId))
	if err != nil {
		return utils.SendError(errors.New("Formulir pertanyaan tidak ditemukan"), http.StatusNotFound)
	}

	if len(fields) == 0 {
		return utils.SendData(payloads.ManajemenAlurPayload{
			ID:                         flowDetail.ID,
			FlowCode:                   &flowDetail.Code,
			TemplateFormulirPertanyaan: &form.Code,
			NamaAlur:                   flowDetail.Name,
			Pembuka:                    flowDetail.OpeningId,
			Penutup:                    flowDetail.ClosingId,
			HasSection:                 flowDetail.StatusSection == "1",
			Flows:                      []payloads.FlowItem{},
		}, "Detail alur berhasil diambil")
	}

	var fieldIDs []int
	var sectionIDs []int
	var groupIDs []int

	uniqueSections := make(map[int]bool)
	uniqueGroups := make(map[int]bool)
	mapSequenceToFields := make(map[int][]models.FlowField)
	mapGroupIDToFirstQuestionID := make(map[int]int)

	for _, f := range fields {
		fieldIDs = append(fieldIDs, f.ID)

		if f.SectionId != nil && !uniqueSections[*f.SectionId] {
			uniqueSections[*f.SectionId] = true
			sectionIDs = append(sectionIDs, *f.SectionId)
		}
		if f.GroupId != 0 {
			if !uniqueGroups[f.GroupId] {
				uniqueGroups[f.GroupId] = true
				groupIDs = append(groupIDs, f.GroupId)
			}
			if _, exists := mapGroupIDToFirstQuestionID[f.GroupId]; !exists {
				mapGroupIDToFirstQuestionID[f.GroupId] = f.FormFieldId
			}
		}

		mapSequenceToFields[f.Sequence] = append(mapSequenceToFields[f.Sequence], f)
	}

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

	var reconstructedFlows []payloads.FlowItem
	var sequences []int
	for seq := range mapSequenceToFields {
		sequences = append(sequences, seq)
	}
	sort.Ints(sequences)

	printedSections := make(map[int]bool)
	sectionCounter := 1
	mapSectionDBIDToIndex := make(map[int]int)

	// Variabel global pembantu pointer
	rLogic := "logic"
	rJump := "jump-to"

	// Helper untuk merender pointer TargetQuestionID (menghindari memory shared bug)
	getTargetPtr := func(child int, grpChild int) *int {
		target := child
		if grpChild != 0 {
			if qID, exists := mapGroupIDToFirstQuestionID[grpChild]; exists {
				target = qID
			}
		}
		if target == 0 {
			return nil
		}
		val := target
		return &val
	}

	for _, seq := range sequences {
		seqFields := mapSequenceToFields[seq]
		firstField := seqFields[0]

		flowItem := payloads.FlowItem{
			Sequence: seq,
			IsGroup:  firstField.GroupId != 0,
		}

		for _, f := range seqFields {
			flowItem.FieldIDs = append(flowItem.FieldIDs, f.ID)
		}

		if firstField.SectionId != nil {
			secDBID := *firstField.SectionId
			flowItem.SectionID = &secDBID

			if _, exists := mapSectionDBIDToIndex[secDBID]; !exists {
				mapSectionDBIDToIndex[secDBID] = sectionCounter
				sectionCounter++
			}
			secIdx := mapSectionDBIDToIndex[secDBID]
			flowItem.SectionIndex = &secIdx

			if !printedSections[secDBID] {
				name := mapSectionNameByID[secDBID]
				flowItem.SectionName = &name
				printedSections[secDBID] = false // ini nanti bisa di ubah ke true kalau mau hanya tampilkan section name di field pertama tiap section, untuk sekarang biar muncul semua aja
			}
		}

		if flowItem.IsGroup {
			gID := firstField.GroupId
			flowItem.GroupID = &gID
			gName := mapGroupNameByID[gID]
			flowItem.GroupName = &gName
		}

		mapUniqueQID := make(map[int]bool)
		for _, f := range seqFields {
			if !mapUniqueQID[f.FormFieldId] {
				flowItem.QuestionIDs = append(flowItem.QuestionIDs, f.FormFieldId)
				mapUniqueQID[f.FormFieldId] = true
			}
		}

		lastQuestionID := flowItem.QuestionIDs[len(flowItem.QuestionIDs)-1]
		var exitField models.FlowField
		for _, f := range seqFields {
			if f.FormFieldId == lastQuestionID {
				exitField = f
				break
			}
		}

		flowItem.Routing = payloads.RoutingRule{
			IsBreakdown: firstField.Breakdown,
			Options:     []payloads.OptionRoute{},
			Logics:      []payloads.LogicRoute{},
		}

		if firstField.Breakdown {
			flowItem.Routing.Rule = nil
			flowItem.Routing.IsEnd = false

			for _, f := range seqFields {
				if f.FormAnswerFieldId != nil {
					optRule := &rJump
					if f.IsAdvancedOption {
						optRule = &rLogic
					}

					var optLogics []payloads.LogicRoute
					if f.IsAdvancedOption {
						if logics, exists := mapLogicsByFieldID[f.ID]; exists {
							for _, logic := range logics {
								optLogics = append(optLogics, payloads.LogicRoute{
									LogicID:          logic.ID,
									IfQuestionID:     logic.FormFieldId,
									IfOptionID:       logic.Option,
									TargetQuestionID: getTargetPtr(logic.ChildId, 0),
									IsEnd:            logic.ChildId == 0,
								})
							}
						}
					}

					flowItem.Routing.Options = append(flowItem.Routing.Options, payloads.OptionRoute{
						FieldID:          f.ID,
						OptionID:         *f.FormAnswerFieldId,
						Rule:             optRule,
						TargetQuestionID: getTargetPtr(f.ChildId, int(f.GroupChildId)),
						IsEnd:            f.ChildId == 0 && f.GroupChildId == 0,
						Logics:           optLogics,
					})
				}
			}

		} else {
			flowItem.Routing.TargetQuestionID = getTargetPtr(exitField.ChildId, int(exitField.GroupChildId))
			flowItem.Routing.IsEnd = exitField.ChildId == 0 && exitField.GroupChildId == 0

			if exitField.IsAdvancedOption {
				flowItem.Routing.Rule = &rLogic
				if logics, exists := mapLogicsByFieldID[exitField.ID]; exists {
					for _, logic := range logics {
						flowItem.Routing.Logics = append(flowItem.Routing.Logics, payloads.LogicRoute{
							LogicID:          logic.ID,
							IfQuestionID:     logic.FormFieldId,
							IfOptionID:       logic.Option,
							TargetQuestionID: getTargetPtr(logic.ChildId, 0),
							IsEnd:            logic.ChildId == 0,
						})
					}
				}
			} else {
				flowItem.Routing.Rule = &rJump
			}
		}

		reconstructedFlows = append(reconstructedFlows, flowItem)
	}

	finalResponse := payloads.ManajemenAlurPayload{
		ID:                         flowDetail.ID,
		FlowCode:                   &flowDetail.Code,
		TemplateFormulirPertanyaan: &form.Code,
		NamaAlur:                   flowDetail.Name,
		Pembuka:                    flowDetail.OpeningId,
		Penutup:                    flowDetail.ClosingId,
		HasSection:                 flowDetail.StatusSection == "1",
		Flows:                      reconstructedFlows,
	}

	return utils.SendData(finalResponse, "Detail manajemen alur berhasil diambil")
}

func (service *manajemenAlurService) UpdateManajemenAlur(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.ManajemenAlurPayload

	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}
	err = json.Unmarshal(jsonBytes, &payload)
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
		existingDetail.Version += 1 // Increment Version

		if err := txRepo.UpdateFlowDetail(existingDetail); err != nil {
			return err
		}

		// WIPE AND REPLACE: Hapus semua relasi lama
		if err := txRepo.DeleteRoutingByDetailID(existingDetail.ID); err != nil {
			return err
		}

		mapSectionIndexToDBID := make(map[int]int)
		mapFirstQuestionIDToGroupID := make(map[int]int)

		for _, flow := range payload.Flows {
			// UPSERT SECTION
			if payload.HasSection && flow.SectionIndex != nil {
				if _, exists := mapSectionIndexToDBID[*flow.SectionIndex]; !exists {
					secName := fmt.Sprintf("Section %d", *flow.SectionIndex)
					if flow.SectionName != nil && *flow.SectionName != "" {
						secName = *flow.SectionName
					}

					secModel := &models.FlowSection{Name: secName}
					// Jika ada SectionID dari payload (Data Lama), gunakan ID itu
					if flow.SectionID != nil && *flow.SectionID != 0 {
						secModel.ID = *flow.SectionID
					}

					if err := txRepo.SaveSection(secModel); err != nil {
						return err
					}
					mapSectionIndexToDBID[*flow.SectionIndex] = secModel.ID
				}
			}

			// UPSERT GROUP
			if flow.IsGroup {
				grpModel := &models.FlowGroup{Name: *flow.GroupName}
				// Jika ada GroupID dari payload (Data Lama), gunakan ID itu
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

			// Ekstraksi Rule Utama
			mainRuleVal := "jump-to"
			if flow.Routing.Rule != nil && *flow.Routing.Rule != "" {
				mainRuleVal = *flow.Routing.Rule
			}

			for indexQ, qID := range flow.QuestionIDs {
				formFieldDetail, err := service.templateFormulirPertanyaanRepo.GetFormFieldByID(qID)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return utils.NewClientError(fmt.Sprintf("pertanyaan dengan ID %d tidak ditemukan", qID))
					}
					return err
				}

				if formFieldDetail == nil {
					return utils.NewClientError(fmt.Sprintf("pertanyaan dengan ID %d tidak ditemukan", qID))
				}

				if formFieldDetail.Template != string(enums.MULTIPLE_CHOICES) {
					if len(flow.Routing.Options) > 0 {
						return utils.NewClientError(fmt.Sprintf("pertanyaan dengan ID %d tidak boleh memiliki opsi jawaban karena bukan bertipe multiple choices", qID))
					}
				}

				if flow.Routing.Rule != nil {
					if *flow.Routing.Rule != string(enums.ROUTING_RULE_LOGIC) {
						if len(flow.Routing.Logics) > 0 {
							return utils.NewClientError(
								fmt.Sprintf("pertanyaan dengan ID %d tidak boleh memiliki logics", qID),
							)
						}
					} else if *flow.Routing.Rule == string(enums.ROUTING_RULE_LOGIC) {
						if len(flow.Routing.Logics) < 1 {
							return utils.NewClientError(
								fmt.Sprintf("pertanyaan dengan ID %d harus setidaknya memiliki 1 logics", qID),
							)
						}
					}
				}

				field := &models.FlowField{
					FlowDetailId:     existingDetail.ID, // Gunakan ID existingDetail
					Sequence:         flow.Sequence,
					FormFieldId:      qID,
					GroupChildId:     0,
					GroupId:          currentGroupID,
					SectionId:        currentSectionID,
					Breakdown:        false,
					IsAdvancedOption: false,
				}

				if flow.IsGroup {
					if formFieldDetail.Template != string(enums.MULTIPLE_CHOICES) {
						return utils.NewClientError(fmt.Sprintf("pertanyaan dengan ID %d harus bertipe Multiple Choices untuk digunakan dalam group", qID))
					}

					if len(flow.Routing.Options) > 0 {
						return utils.NewClientError(
							fmt.Sprintf("pertanyaan dengan ID %d bertipe group, tidak boleh memiliki opsi jawaban", qID),
						)
					}

					if len(flow.Routing.Logics) > 0 {
						return utils.NewClientError(
							fmt.Sprintf("pertanyaan dengan ID %d bertipe group, tidak boleh memiliki logika", qID),
						)
					}
				}

				if flow.IsGroup && indexQ < len(flow.QuestionIDs)-1 {
					nextQIDInGroup := flow.QuestionIDs[indexQ+1]
					field.ChildId, field.GroupChildId = resolveTarget(&nextQIDInGroup)
				} else {
					if flow.Routing.IsBreakdown {
						field.Breakdown = true
						field.ChildId = 0
						field.GroupChildId = 0

						for _, opt := range flow.Routing.Options {
							var child, grpChild int
							isAdvanced := false

							optRuleVal := "jump-to"
							if opt.Rule != nil && *opt.Rule != "" {
								optRuleVal = *opt.Rule
							}

							if optRuleVal == "logic" {
								isAdvanced = true
								child = 0 // Logic tidak memiliki child di flow_fields
								grpChild = 0
							} else {
								if opt.TargetQuestionID != nil && *opt.TargetQuestionID != 0 {
									questionExists, err := service.templateFormulirPertanyaanRepo.IsQuestionExistsById(*opt.TargetQuestionID)
									if err != nil {
										return err
									}
									if !questionExists {
										return fmt.Errorf("pertanyaan dengan ID %d tidak ditemukan", *opt.TargetQuestionID)
									}
								}
								child, grpChild = resolveTarget(opt.TargetQuestionID)
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
								IsAdvancedOption:  isAdvanced,
							}

							if err := txRepo.CreateFlowField(optField); err != nil {
								return err
							}

							if optRuleVal == "logic" {
								for _, logic := range opt.Logics {
									// PERBAIKAN TARGET UNTUK ADVANCED OPTION FLOWS MENGGUNAKAN RAW ID
									logicTargetID := 0
									if logic.TargetQuestionID != nil && *logic.TargetQuestionID != 0 {
										_questionExists, err := service.templateFormulirPertanyaanRepo.IsQuestionExistsById(*logic.TargetQuestionID)
										if err != nil || !_questionExists {
											return fmt.Errorf("pertanyaan target opsi logic tidak ditemukan")
										}
										logicTargetID = *logic.TargetQuestionID
									}

									advanced := &models.AdvancedOptionFlow{
										FlowFieldId: optField.ID,
										FormFieldId: logic.IfQuestionID,
										Option:      logic.IfOptionID,
										ChildId:     logicTargetID,
									}

									if err := txRepo.CreateAdvancedOption(advanced); err != nil {
										return err
									}
								}
							}
						}
						continue
					} else {
						if mainRuleVal == "jump-to" {
							if flow.Routing.TargetQuestionID != nil && *flow.Routing.TargetQuestionID != 0 {
								questionExists, err := service.templateFormulirPertanyaanRepo.IsQuestionExistsById(*flow.Routing.TargetQuestionID)
								if err != nil || !questionExists {
									return fmt.Errorf("pertanyaan dengan ID %d tidak ditemukan", *flow.Routing.TargetQuestionID)
								}
							}
							field.ChildId, field.GroupChildId = resolveTarget(flow.Routing.TargetQuestionID)
						} else if mainRuleVal == "logic" {
							field.IsAdvancedOption = true
							field.ChildId = 0
							field.GroupChildId = 0
						}
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

				if mainRuleVal == "logic" && !flow.Routing.IsBreakdown && !flow.IsGroup && indexQ == len(flow.QuestionIDs)-1 {
					for _, logic := range flow.Routing.Logics {
						// PERBAIKAN TARGET UNTUK ADVANCED OPTION FLOWS MENGGUNAKAN RAW ID
						logicTargetID := 0
						if logic.TargetQuestionID != nil && *logic.TargetQuestionID != 0 {
							_questionExists, err := service.templateFormulirPertanyaanRepo.IsQuestionExistsById(*logic.TargetQuestionID)
							if err != nil || !_questionExists {
								return fmt.Errorf("pertanyaan target dengan ID %d tidak ditemukan", *logic.TargetQuestionID)
							}
							logicTargetID = *logic.TargetQuestionID
						}

						advanced := &models.AdvancedOptionFlow{
							FlowFieldId: firstInsertedFieldID,
							FormFieldId: logic.IfQuestionID,
							Option:      logic.IfOptionID,
							ChildId:     logicTargetID,
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

func (service *manajemenAlurService) GetListManajemenAlur(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	search := param.Get("search")
	page, _ := strconv.Atoi(param.Get("page"))
	limit, _ := strconv.Atoi(param.Get("limit"))
	orderBy := param.Get("order_by")
	orderDir := param.Get("order_dir")

	payload := payloads.DatatablePayload{
		Search:   search,
		Page:     page,
		Limit:    limit,
		OrderBy:  orderBy,
		OrderDir: orderDir,
	}

	if payload.Page <= 0 {
		payload.Page = 1
	}
	if payload.Limit <= 0 {
		payload.Limit = 5
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

	totalPages := int(math.Ceil(float64(totalData) / float64(payload.Limit)))

	result := map[string]interface{}{
		"data": data,
		"meta": map[string]interface{}{
			"total":      totalData,
			"page":       payload.Page,
			"limit":      payload.Limit,
			"totalPages": totalPages,
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
	rawSections, err := service.manajemenAlurRepo.GetPreviewSectionByFlowDetailId(flowDetail.ID, statusSectionStr, nil)
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
	flowFieldToOptionMap := make(map[int]int)                   // Pemetaan flow_field_id ke option_id khusus breakdown
	breakdownRawMap := make(map[int]map[int]models.RawNodeData) // nodeKey -> optionId -> rawData

	var formFieldIDs []int
	var flowFieldIDs []int

	totalRequired := 0
	totalOptional := 0
	entryNodeId := 0
	var activeSectionName *string

	rJump := "jump-to"
	rLogic := "logic"

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
				val := target
				return &val
			}
		}
		if childID != 0 {
			val := childID
			return &val
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

		// Catat pemetaan option untuk logic breakdown nanti
		if raw.Breakdown && raw.FormAnswerFieldId != nil {
			flowFieldToOptionMap[raw.FlowFieldId] = *raw.FormAnswerFieldId
		}

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
					Logics: []response.PreviewAlurSurveyAdvancedLogicItem{},
				},
			}
			breakdownRawMap[nodeKey] = make(map[int]models.RawNodeData)
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

		if raw.Breakdown {
			step.Routing.IsBreakdown = true
			step.Routing.Rule = nil // Rule utama dikosongkan karena rute ada di level opsi
			if raw.FormAnswerFieldId != nil {
				// Simpan raw node untuk diekstrak rutenya ke dalam Option di PART 3
				breakdownRawMap[nodeKey][*raw.FormAnswerFieldId] = raw
			}
		} else {
			step.Routing.IsBreakdown = false
			step.Routing.TargetQuestionId = target
			step.Routing.IsEnd = (target == nil)
			if raw.IsAdvancedOption {
				step.Routing.Rule = &rLogic
			} else {
				step.Routing.Rule = &rJump
			}
		}

		blueprintNodes[nodeKey] = step
	}

	// PART 3 Sisipkan Opsi (beserta Rute Breakdown)
	options, _ := service.manajemenAlurRepo.GetAnswerOptionsByQuestionIDList(formFieldIDs)
	for _, opt := range options {
		for nodeKey, step := range blueprintNodes {
			for i, q := range step.Questions {
				if q.QuestionId == opt.FormFieldId {

					optItem := response.PreviewAlurSurveyOptionItem{
						ID:     opt.ID,
						Label:  opt.Option,
						Logics: []response.PreviewAlurSurveyAdvancedLogicItem{},
					}

					// Jika soal breakdown, masukkan Rule dan Target khusus opsi tersebut
					if step.Routing.IsBreakdown {
						if rawOpt, ok := breakdownRawMap[nodeKey][opt.ID]; ok {
							optTarget := resolveTargetID(rawOpt.ChildId, int(rawOpt.GroupChildId))
							optItem.TargetQuestionId = optTarget
							optItem.IsEnd = (optTarget == nil)

							if rawOpt.IsAdvancedOption {
								optItem.Rule = &rLogic
							} else {
								optItem.Rule = &rJump
							}
						} else {
							// Fallback aman
							optItem.Rule = &rJump
							optItem.IsEnd = true
						}
					}

					step.Questions[i].Options = append(step.Questions[i].Options, optItem)
					blueprintNodes[nodeKey] = step
				}
			}
		}
	}

	// PART 3.5 Sisipkan Advanced Logics
	logics, _ := service.manajemenAlurRepo.GetAdvancedOptionsByFieldIDs(flowFieldIDs)
	for _, logic := range logics {
		nodeKey, valid := flowFieldToNodeKeyMap[logic.FlowFieldId]

		if valid {
			if step, exists := blueprintNodes[nodeKey]; exists {
				var logicTargetPtr *int
				if logic.ChildId != 0 {
					val := logic.ChildId
					logicTargetPtr = &val
				}

				logicItem := response.PreviewAlurSurveyAdvancedLogicItem{
					IfQuestionId:     logic.FormFieldId,
					IfOptionId:       logic.Option,
					TargetQuestionId: logicTargetPtr,
					IsEnd:            logic.ChildId == 0,
				}

				if step.Routing.IsBreakdown {
					// Jika breakdown, cari opsi spesifik dan letakkan logic ke dalam array opsi tersebut
					optID, hasOptMap := flowFieldToOptionMap[logic.FlowFieldId]
					if hasOptMap {
						for i, q := range step.Questions {
							for j, o := range q.Options {
								if o.ID == optID {
									step.Questions[i].Options[j].Logics = append(step.Questions[i].Options[j].Logics, logicItem)
								}
							}
						}
					}
				} else {
					// Jika bukan breakdown, letakkan di root routing
					step.Routing.Logics = append(step.Routing.Logics, logicItem)
				}

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

func (service *manajemenAlurService) AlurOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	page, err := strconv.Atoi(param.Get("page"))
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(param.Get("limit"))
	if err != nil || limit <= 0 {
		limit = 1000
	}

	var alurIds []string
	if len(param["id[]"]) > 0 {
		alurIds = param["id[]"]
	} else if len(param["id"]) > 0 {
		alurIds = param["id"]
	}

	var parsedIDs []string
	for _, alurId := range alurIds {
		parsedIDs = append(parsedIDs, alurId)
	}

	var excludeAlurIds []string
	if len(param["exclude_id[]"]) > 0 {
		excludeAlurIds = param["exclude_id[]"]
	} else if len(param["exclude_id"]) > 0 {
		excludeAlurIds = param["exclude_id"]
	}

	var parsedExcludeIDs []string
	for _, excludeID := range excludeAlurIds {
		parsedExcludeIDs = append(parsedExcludeIDs, excludeID)
	}

	_req := payloads.ManajemenAlurOptionsPayload{
		Q:          param.Get("q"),
		Page:       page,
		Limit:      limit,
		IDs:        parsedIDs,
		ExcludeIDs: parsedExcludeIDs,
	}

	data, totalData, err := service.manajemenAlurRepo.FlowOptions(_req)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	currentTotalLoaded := (page-1)*limit + len(data)
	hasMore := int64(currentTotalLoaded) < totalData

	responseData := response.StringOptionsResponse{
		Options: data,
		Meta: response.PaginationMeta{
			CurrentPage: page,
			PerPage:     limit,
			Total:       totalData,
			HasMore:     hasMore,
		},
	}

	return utils.SendData(responseData, "Berhasil mengambil opsi alur survey")
}
