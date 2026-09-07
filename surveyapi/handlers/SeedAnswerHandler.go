package handlers

import (
	pb "backend/siccore/pb"
	"backend/surveyapi/models"
	"backend/surveyapi/service"
	"context"
	"net/url"
)

type SeedAnswerHandler interface {
	SeedSurveyAnswer(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type seedAnswerHandler struct {
	svc service.SeedAnswerService
}

func NewSeedAnswerHandler(
	svc service.SeedAnswerService,
) SeedAnswerHandler {
	return &seedAnswerHandler{
		svc: svc,
	}
}

func (handler *seedAnswerHandler) SeedSurveyAnswer(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.svc.SeedSurveyAnswer(ctx, req, usr, param, slug)
}
