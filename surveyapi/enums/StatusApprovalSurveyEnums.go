package enums

type StatusApprovalSurvey string

const (
	STATUS_APPROVAL_SURVEY_NON_APPROVAL StatusApprovalSurvey = "non_approval"
	STATUS_APPROVAL_SURVEY_WAITING      StatusApprovalSurvey = "waiting"
	STATUS_APPROVAL_SURVEY_APPROVED     StatusApprovalSurvey = "approved"
	STATUS_APPROVAL_SURVEY_REJECTED     StatusApprovalSurvey = "rejected"
)
