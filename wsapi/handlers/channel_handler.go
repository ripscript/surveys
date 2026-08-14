package handlers

import (
	"backend/siccore/pb"
	"backend/wsapi/models"
	"backend/wsapi/service"
	"context"
	"net/url"
)

type ChannelHandler interface {
	GetTicket(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type channelHandler struct {
	channelService service.ChannelService
}

func NewChannelHandler(
	channelService service.ChannelService,
) ChannelHandler {
	return &channelHandler{
		channelService: channelService,
	}
}

func (handler *channelHandler) GetTicket(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.channelService.GetTicket(ctx, req, usr, param, slug)
}
