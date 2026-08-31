package handlers

import (
	"backend/reportapi/models"
	"backend/reportapi/service"
	pb "backend/siccore/pb"
	"context"
	"net/url"
)

type RatingHandler interface {
	SaveRating(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	RatingOverview(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	RatingList(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetOldRating(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type ratingHandler struct {
	ratingService service.RatingService
}

func NewRatingHandler(
	ratingService service.RatingService,
) RatingHandler {
	return &ratingHandler{
		ratingService,
	}
}

func (handler *ratingHandler) SaveRating(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.ratingService.SaveRating(ctx, req, usr, param, slug)
}

func (handler *ratingHandler) RatingOverview(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.ratingService.RatingOverview(ctx, req, usr, param, slug)
}

func (handler *ratingHandler) RatingList(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.ratingService.RatingList(ctx, req, usr, param, slug)
}

func (handler *ratingHandler) GetOldRating(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.ratingService.GetOldRating(ctx, req, usr, param, slug)
}
