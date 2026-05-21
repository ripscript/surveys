package customValidator

import (
	"backend/surveyapi/payloads"
	"fmt"

	"github.com/go-playground/validator/v10"
)

func ManajemenAlurPayloadValidator(sl validator.StructLevel) {
	payload := sl.Current().Interface().(payloads.ManajemenAlurPayload)

	registeredQuestions := make(map[int]int)
	sectionNames := make(map[int]string)
	usedNames := make(map[string]int)

	// ==========================================================
	// TAMBAHAN 1: Peta Relasi Question ID ke Section Index
	// ==========================================================
	questionToSectionMap := make(map[int]*int)
	for _, flow := range payload.Flows {
		for _, qID := range flow.QuestionIDs {
			questionToSectionMap[qID] = flow.SectionIndex
		}
	}

	// ==========================================================
	// TAMBAHAN 2: Helper Closure untuk validasi Enkapsulasi Section
	// ==========================================================
	checkTargetSection := func(targetID *int, sourceSection *int, fieldPath string) {
		if targetID != nil && *targetID != 0 {
			targetSec, exists := questionToSectionMap[*targetID]
			if !exists {
				// Target ID bahkan tidak ada di payload saat ini
				sl.ReportError(targetID, fieldPath, "TargetQuestionID", "invalid_target_question_id", "")
				return
			}

			// Pengecekan kesamaan SectionIndex
			isSameSection := false
			if sourceSection == nil && targetSec == nil {
				isSameSection = true
			} else if sourceSection != nil && targetSec != nil && *sourceSection == *targetSec {
				isSameSection = true
			}

			if !isSameSection {
				sl.ReportError(targetID, fieldPath, "target_question_id", "cross_section_routing_not_allowed", "")
			}
		}
	}

	if payload.HasSection {
		for i, flow := range payload.Flows {
			if flow.SectionIndex != nil && flow.SectionName != nil && *flow.SectionName != "" {
				idx := *flow.SectionIndex
				name := *flow.SectionName

				if existingIdx, exists := usedNames[name]; exists && existingIdx != idx {
					sl.ReportError(flow.SectionName, fmt.Sprintf("Flows[%d].SectionName", i), "SectionName", "duplicate_section_name", fmt.Sprintf("%d", existingIdx))
				} else {
					sectionNames[idx] = name
					usedNames[name] = idx
				}
			}
		}
	}

	for i, flow := range payload.Flows {
		for _, qID := range flow.QuestionIDs {
			if firstSeenIndex, exists := registeredQuestions[qID]; exists {
				sl.ReportError(flow.QuestionIDs, fmt.Sprintf("Flows[%d].QuestionIDs", i), "QuestionIDs", "duplicate_question_id", fmt.Sprintf("%d", firstSeenIndex+1))
			} else {
				registeredQuestions[qID] = i
			}
		}

		isSectionEmpty := flow.SectionIndex == nil || *flow.SectionIndex == 0

		if payload.HasSection {
			if isSectionEmpty {
				sl.ReportError(flow.SectionIndex, fmt.Sprintf("Flows[%d].SectionIndex", i), "SectionIndex", "required_with_has_section", "")
			} else {
				if _, hasName := sectionNames[*flow.SectionIndex]; !hasName {
					sl.ReportError(flow.SectionName, fmt.Sprintf("Flows[%d].SectionName", i), "SectionName", "section_must_have_name", fmt.Sprintf("%d", *flow.SectionIndex))
				}
			}
		} else {
			if !isSectionEmpty {
				sl.ReportError(flow.SectionIndex, fmt.Sprintf("Flows[%d].SectionIndex", i), "SectionIndex", "must_be_null_if_no_section", "")
			}
			if flow.SectionName != nil && *flow.SectionName != "" {
				sl.ReportError(flow.SectionName, fmt.Sprintf("Flows[%d].SectionName", i), "SectionName", "must_be_null_if_not_using_section", "")
			}
		}

		if flow.IsGroup {
			if flow.GroupName == nil || *flow.GroupName == "" {
				sl.ReportError(flow.GroupName, fmt.Sprintf("Flows[%d].GroupName", i), "GroupName", "group_must_have_name", "")
			}
			if flow.Routing.IsBreakdown {
				sl.ReportError(flow.Routing.IsBreakdown, fmt.Sprintf("Flows[%d].Routing.IsBreakdown", i), "IsBreakdown", "group_cannot_breakdown", "")
			}
			if flow.Routing.Rule != nil && *flow.Routing.Rule == "logic" {
				sl.ReportError(flow.Routing.Rule, fmt.Sprintf("Flows[%d].Routing.Rule", i), "Rule", "group_cannot_use_logic", "")
			}
			if len(flow.QuestionIDs) < 2 {
				sl.ReportError(flow.QuestionIDs, fmt.Sprintf("Flows[%d].QuestionIDs", i), "QuestionIDs", "group_must_have_multiple_questions", "")
			}
		} else {
			if flow.GroupName != nil && *flow.GroupName != "" {
				sl.ReportError(flow.GroupName, fmt.Sprintf("Flows[%d].GroupName", i), "GroupName", "must_be_null_if_not_group", "")
			}
		}

		// ==========================================================
		// VALIDASI UTAMA: BREAKDOWN vs NON-BREAKDOWN
		// ==========================================================
		if flow.Routing.IsBreakdown {
			// RULE 1: Rule utama WAJIB NULL jika breakdown
			if flow.Routing.Rule != nil && *flow.Routing.Rule != "" {
				sl.ReportError(flow.Routing.Rule, fmt.Sprintf("Flows[%d].Routing.Rule", i), "Rule", "must_be_null_if_breakdown", "")
			}
			if len(flow.Routing.Options) == 0 {
				sl.ReportError(flow.Routing.Options, fmt.Sprintf("Flows[%d].Routing.Options", i), "Options", "required_if_breakdown_true", "")
			}
			if flow.Routing.TargetQuestionID != nil && *flow.Routing.TargetQuestionID != 0 {
				sl.ReportError(flow.Routing.TargetQuestionID, fmt.Sprintf("Flows[%d].Routing.TargetQuestionID", i), "TargetQuestionID", "must_be_null_if_breakdown", "")
			}

			// Validasi dinamis per Opsi Jawaban
			for j, opt := range flow.Routing.Options {
				if opt.Rule == nil || *opt.Rule == "" {
					sl.ReportError(opt.Rule, fmt.Sprintf("Flows[%d].Routing.Options[%d].Rule", i, j), "Rule", "required_for_option", "")
					continue
				}

				optRuleVal := *opt.Rule

				if optRuleVal == "logic" {
					if opt.TargetQuestionID != nil && *opt.TargetQuestionID != 0 {
						sl.ReportError(opt.TargetQuestionID, fmt.Sprintf("Flows[%d].Routing.Options[%d].TargetQuestionID", i, j), "TargetQuestionID", "must_be_null_if_rule_logic", "")
					}
					if len(opt.Logics) == 0 {
						sl.ReportError(opt.Logics, fmt.Sprintf("Flows[%d].Routing.Options[%d].Logics", i, j), "Logics", "required_if_option_rule_logic", "")
					}
					for k, logic := range opt.Logics {
						isTargetEmpty := logic.TargetQuestionID == nil || *logic.TargetQuestionID == 0
						if isTargetEmpty && !logic.IsEnd {
							sl.ReportError(logic.TargetQuestionID, fmt.Sprintf("Flows[%d].Routing.Options[%d].Logics[%d]", i, j, k), "TargetQuestionID", "must_have_target_or_end", "")
						}
						if !isTargetEmpty && logic.IsEnd {
							sl.ReportError(logic.TargetQuestionID, fmt.Sprintf("Flows[%d].Routing.Options[%d].Logics[%d]", i, j, k), "TargetQuestionID", "must_be_null_if_end", "")
						}
						// Eksekusi Helper Validasi Section (Logics pada Breakdown Option)
						checkTargetSection(logic.TargetQuestionID, flow.SectionIndex, fmt.Sprintf("Flows[%d].Routing.Options[%d].Logics[%d].TargetQuestionID", i, j, k))
					}
				} else if optRuleVal == "jump-to" {
					isTargetEmpty := opt.TargetQuestionID == nil || *opt.TargetQuestionID == 0
					if isTargetEmpty && !opt.IsEnd {
						sl.ReportError(opt.TargetQuestionID, fmt.Sprintf("Flows[%d].Routing.Options[%d]", i, j), "TargetQuestionID", "must_have_target_or_end", "")
					}
					if !isTargetEmpty && opt.IsEnd {
						sl.ReportError(opt.TargetQuestionID, fmt.Sprintf("Flows[%d].Routing.Options[%d]", i, j), "TargetQuestionID", "must_be_null_if_end", "")
					}
					// Eksekusi Helper Validasi Section (Jump-to pada Breakdown Option)
					checkTargetSection(opt.TargetQuestionID, flow.SectionIndex, fmt.Sprintf("Flows[%d].Routing.Options[%d].TargetQuestionID", i, j))
				}
			}

		} else {
			// RULE 2: Rule utama WAJIB ADA jika BUKAN breakdown
			if flow.Routing.Rule == nil || *flow.Routing.Rule == "" {
				sl.ReportError(flow.Routing.Rule, fmt.Sprintf("Flows[%d].Routing.Rule", i), "Rule", "required_if_not_breakdown", "")
				continue // Stop validasi ini agar tidak panic pointer
			}

			mainRuleVal := *flow.Routing.Rule

			if mainRuleVal == "logic" {
				if flow.Routing.TargetQuestionID != nil && *flow.Routing.TargetQuestionID != 0 {
					sl.ReportError(flow.Routing.TargetQuestionID, fmt.Sprintf("Flows[%d].Routing.TargetQuestionID", i), "TargetQuestionID", "must_be_null_if_rule_logic", "")
				}
				if len(flow.Routing.Logics) == 0 {
					sl.ReportError(flow.Routing.Logics, fmt.Sprintf("Flows[%d].Routing.Logics", i), "Logics", "required_if_rule_logic", "")
				}
				for j, logic := range flow.Routing.Logics {
					isTargetEmpty := logic.TargetQuestionID == nil || *logic.TargetQuestionID == 0
					if isTargetEmpty && !logic.IsEnd {
						sl.ReportError(logic.TargetQuestionID, fmt.Sprintf("Flows[%d].Routing.Logics[%d]", i, j), "TargetQuestionID", "must_have_target_or_end", "")
					}
					if !isTargetEmpty && logic.IsEnd {
						sl.ReportError(logic.TargetQuestionID, fmt.Sprintf("Flows[%d].Routing.Logics[%d]", i, j), "TargetQuestionID", "must_be_null_if_end", "")
					}
					// Eksekusi Helper Validasi Section (Logics pada Main Rule)
					checkTargetSection(logic.TargetQuestionID, flow.SectionIndex, fmt.Sprintf("Flows[%d].Routing.Logics[%d].TargetQuestionID", i, j))
				}
			} else if mainRuleVal == "jump-to" {
				isTargetEmpty := flow.Routing.TargetQuestionID == nil || *flow.Routing.TargetQuestionID == 0
				if isTargetEmpty && !flow.Routing.IsEnd {
					sl.ReportError(flow.Routing.TargetQuestionID, fmt.Sprintf("Flows[%d].Routing", i), "Routing", "must_have_target_or_end", "")
				}
				if !isTargetEmpty && flow.Routing.IsEnd {
					sl.ReportError(flow.Routing.TargetQuestionID, fmt.Sprintf("Flows[%d].Routing.TargetQuestionID", i), "TargetQuestionID", "must_be_null_if_end", "")
				}
				// Eksekusi Helper Validasi Section (Jump-to pada Main Rule)
				checkTargetSection(flow.Routing.TargetQuestionID, flow.SectionIndex, fmt.Sprintf("Flows[%d].Routing.TargetQuestionID", i))
			}
		}
	}
}

func ManajemenAlurUpdatePayloadValidator(sl validator.StructLevel) {
	payload := sl.Current().Interface().(payloads.ManajemenAlurPayload)

	registeredQuestions := make(map[int]int)
	sectionNames := make(map[int]string)
	usedNames := make(map[string]int)

	// ==========================================================
	// TAMBAHAN 1: Peta Relasi Question ID ke Section Index
	// ==========================================================
	questionToSectionMap := make(map[int]*int)
	for _, flow := range payload.Flows {
		for _, qID := range flow.QuestionIDs {
			questionToSectionMap[qID] = flow.SectionIndex
		}
	}

	// ==========================================================
	// TAMBAHAN 2: Helper Closure untuk validasi Enkapsulasi Section
	// ==========================================================
	checkTargetSection := func(targetID *int, sourceSection *int, fieldPath string) {
		if targetID != nil && *targetID != 0 {
			targetSec, exists := questionToSectionMap[*targetID]
			if !exists {
				// Target ID bahkan tidak ada di payload saat ini
				sl.ReportError(targetID, fieldPath, "TargetQuestionID", "invalid_target_question_id", "")
				return
			}

			// Pengecekan kesamaan SectionIndex
			isSameSection := false
			if sourceSection == nil && targetSec == nil {
				isSameSection = true
			} else if sourceSection != nil && targetSec != nil && *sourceSection == *targetSec {
				isSameSection = true
			}

			if !isSameSection {
				sl.ReportError(targetID, fieldPath, "target_question_id", "cross_section_routing_not_allowed", "")
			}
		}
	}

	if payload.HasSection {
		for i, flow := range payload.Flows {
			if flow.SectionIndex != nil && flow.SectionName != nil && *flow.SectionName != "" {
				idx := *flow.SectionIndex
				name := *flow.SectionName

				if existingIdx, exists := usedNames[name]; exists && existingIdx != idx {
					sl.ReportError(flow.SectionName, fmt.Sprintf("Flows[%d].SectionName", i), "SectionName", "duplicate_section_name", fmt.Sprintf("%d", existingIdx))
				} else {
					sectionNames[idx] = name
					usedNames[name] = idx
				}
			}
		}
	}

	for i, flow := range payload.Flows {
		for _, qID := range flow.QuestionIDs {
			if firstSeenIndex, exists := registeredQuestions[qID]; exists {
				sl.ReportError(flow.QuestionIDs, fmt.Sprintf("Flows[%d].QuestionIDs", i), "QuestionIDs", "duplicate_question_id", fmt.Sprintf("%d", firstSeenIndex+1))
			} else {
				registeredQuestions[qID] = i
			}
		}

		isSectionEmpty := flow.SectionIndex == nil || *flow.SectionIndex == 0

		if payload.HasSection {
			if isSectionEmpty {
				sl.ReportError(flow.SectionIndex, fmt.Sprintf("Flows[%d].SectionIndex", i), "SectionIndex", "required_with_has_section", "")
			} else {
				if _, hasName := sectionNames[*flow.SectionIndex]; !hasName {
					sl.ReportError(flow.SectionName, fmt.Sprintf("Flows[%d].SectionName", i), "SectionName", "section_must_have_name", fmt.Sprintf("%d", *flow.SectionIndex))
				}
			}
		} else {
			if !isSectionEmpty {
				sl.ReportError(flow.SectionIndex, fmt.Sprintf("Flows[%d].SectionIndex", i), "SectionIndex", "must_be_null_if_no_section", "")
			}
			if flow.SectionName != nil && *flow.SectionName != "" {
				sl.ReportError(flow.SectionName, fmt.Sprintf("Flows[%d].SectionName", i), "SectionName", "must_be_null_if_not_using_section", "")
			}
		}

		if flow.IsGroup {
			if flow.GroupName == nil || *flow.GroupName == "" {
				sl.ReportError(flow.GroupName, fmt.Sprintf("Flows[%d].GroupName", i), "GroupName", "group_must_have_name", "")
			}
			if flow.Routing.IsBreakdown {
				sl.ReportError(flow.Routing.IsBreakdown, fmt.Sprintf("Flows[%d].Routing.IsBreakdown", i), "IsBreakdown", "group_cannot_breakdown", "")
			}
			if flow.Routing.Rule != nil && *flow.Routing.Rule == "logic" {
				sl.ReportError(flow.Routing.Rule, fmt.Sprintf("Flows[%d].Routing.Rule", i), "Rule", "group_cannot_use_logic", "")
			}
			if len(flow.QuestionIDs) < 2 {
				sl.ReportError(flow.QuestionIDs, fmt.Sprintf("Flows[%d].QuestionIDs", i), "QuestionIDs", "group_must_have_multiple_questions", "")
			}
		} else {
			if flow.GroupName != nil && *flow.GroupName != "" {
				sl.ReportError(flow.GroupName, fmt.Sprintf("Flows[%d].GroupName", i), "GroupName", "must_be_null_if_not_group", "")
			}
		}

		// ==========================================================
		// VALIDASI UTAMA: BREAKDOWN vs NON-BREAKDOWN
		// ==========================================================
		if flow.Routing.IsBreakdown {
			// RULE 1: Rule utama WAJIB NULL jika breakdown
			if flow.Routing.Rule != nil && *flow.Routing.Rule != "" {
				sl.ReportError(flow.Routing.Rule, fmt.Sprintf("Flows[%d].Routing.Rule", i), "Rule", "must_be_null_if_breakdown", "")
			}
			if len(flow.Routing.Options) == 0 {
				sl.ReportError(flow.Routing.Options, fmt.Sprintf("Flows[%d].Routing.Options", i), "Options", "required_if_breakdown_true", "")
			}
			if flow.Routing.TargetQuestionID != nil && *flow.Routing.TargetQuestionID != 0 {
				sl.ReportError(flow.Routing.TargetQuestionID, fmt.Sprintf("Flows[%d].Routing.TargetQuestionID", i), "TargetQuestionID", "must_be_null_if_breakdown", "")
			}

			// Validasi dinamis per Opsi Jawaban
			for j, opt := range flow.Routing.Options {
				if opt.Rule == nil || *opt.Rule == "" {
					sl.ReportError(opt.Rule, fmt.Sprintf("Flows[%d].Routing.Options[%d].Rule", i, j), "Rule", "required_for_option", "")
					continue
				}

				optRuleVal := *opt.Rule

				if optRuleVal == "logic" {
					if opt.TargetQuestionID != nil && *opt.TargetQuestionID != 0 {
						sl.ReportError(opt.TargetQuestionID, fmt.Sprintf("Flows[%d].Routing.Options[%d].TargetQuestionID", i, j), "TargetQuestionID", "must_be_null_if_rule_logic", "")
					}
					if len(opt.Logics) == 0 {
						sl.ReportError(opt.Logics, fmt.Sprintf("Flows[%d].Routing.Options[%d].Logics", i, j), "Logics", "required_if_option_rule_logic", "")
					}
					for k, logic := range opt.Logics {
						isTargetEmpty := logic.TargetQuestionID == nil || *logic.TargetQuestionID == 0
						if isTargetEmpty && !logic.IsEnd {
							sl.ReportError(logic.TargetQuestionID, fmt.Sprintf("Flows[%d].Routing.Options[%d].Logics[%d]", i, j, k), "TargetQuestionID", "must_have_target_or_end", "")
						}
						if !isTargetEmpty && logic.IsEnd {
							sl.ReportError(logic.TargetQuestionID, fmt.Sprintf("Flows[%d].Routing.Options[%d].Logics[%d]", i, j, k), "TargetQuestionID", "must_be_null_if_end", "")
						}
						// Eksekusi Helper Validasi Section (Logics pada Breakdown Option)
						checkTargetSection(logic.TargetQuestionID, flow.SectionIndex, fmt.Sprintf("Flows[%d].Routing.Options[%d].Logics[%d].TargetQuestionID", i, j, k))
					}
				} else if optRuleVal == "jump-to" {
					isTargetEmpty := opt.TargetQuestionID == nil || *opt.TargetQuestionID == 0
					if isTargetEmpty && !opt.IsEnd {
						sl.ReportError(opt.TargetQuestionID, fmt.Sprintf("Flows[%d].Routing.Options[%d]", i, j), "TargetQuestionID", "must_have_target_or_end", "")
					}
					if !isTargetEmpty && opt.IsEnd {
						sl.ReportError(opt.TargetQuestionID, fmt.Sprintf("Flows[%d].Routing.Options[%d]", i, j), "TargetQuestionID", "must_be_null_if_end", "")
					}
					// Eksekusi Helper Validasi Section (Jump-to pada Breakdown Option)
					checkTargetSection(opt.TargetQuestionID, flow.SectionIndex, fmt.Sprintf("Flows[%d].Routing.Options[%d].TargetQuestionID", i, j))
				}
			}

		} else {
			// RULE 2: Rule utama WAJIB ADA jika BUKAN breakdown
			if flow.Routing.Rule == nil || *flow.Routing.Rule == "" {
				sl.ReportError(flow.Routing.Rule, fmt.Sprintf("Flows[%d].Routing.Rule", i), "Rule", "required_if_not_breakdown", "")
				continue // Stop validasi ini agar tidak panic pointer
			}

			mainRuleVal := *flow.Routing.Rule

			if mainRuleVal == "logic" {
				if flow.Routing.TargetQuestionID != nil && *flow.Routing.TargetQuestionID != 0 {
					sl.ReportError(flow.Routing.TargetQuestionID, fmt.Sprintf("Flows[%d].Routing.TargetQuestionID", i), "TargetQuestionID", "must_be_null_if_rule_logic", "")
				}
				if len(flow.Routing.Logics) == 0 {
					sl.ReportError(flow.Routing.Logics, fmt.Sprintf("Flows[%d].Routing.Logics", i), "Logics", "required_if_rule_logic", "")
				}
				for j, logic := range flow.Routing.Logics {
					isTargetEmpty := logic.TargetQuestionID == nil || *logic.TargetQuestionID == 0
					if isTargetEmpty && !logic.IsEnd {
						sl.ReportError(logic.TargetQuestionID, fmt.Sprintf("Flows[%d].Routing.Logics[%d]", i, j), "TargetQuestionID", "must_have_target_or_end", "")
					}
					if !isTargetEmpty && logic.IsEnd {
						sl.ReportError(logic.TargetQuestionID, fmt.Sprintf("Flows[%d].Routing.Logics[%d]", i, j), "TargetQuestionID", "must_be_null_if_end", "")
					}
					// Eksekusi Helper Validasi Section (Logics pada Main Rule)
					checkTargetSection(logic.TargetQuestionID, flow.SectionIndex, fmt.Sprintf("Flows[%d].Routing.Logics[%d].TargetQuestionID", i, j))
				}
			} else if mainRuleVal == "jump-to" {
				isTargetEmpty := flow.Routing.TargetQuestionID == nil || *flow.Routing.TargetQuestionID == 0
				if isTargetEmpty && !flow.Routing.IsEnd {
					sl.ReportError(flow.Routing.TargetQuestionID, fmt.Sprintf("Flows[%d].Routing", i), "Routing", "must_have_target_or_end", "")
				}
				if !isTargetEmpty && flow.Routing.IsEnd {
					sl.ReportError(flow.Routing.TargetQuestionID, fmt.Sprintf("Flows[%d].Routing.TargetQuestionID", i), "TargetQuestionID", "must_be_null_if_end", "")
				}
				// Eksekusi Helper Validasi Section (Jump-to pada Main Rule)
				checkTargetSection(flow.Routing.TargetQuestionID, flow.SectionIndex, fmt.Sprintf("Flows[%d].Routing.TargetQuestionID", i))
			}
		}
	}
}
