// backend/wsapi/handlers/ws_handler.go
package handlers

import (
	"net/http"
	"time"

	"backend/wsapi/configs"
	"backend/wsapi/hub"
	"backend/wsapi/repository"
	"backend/wsapi/utils"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

const (
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

type WsHandler struct {
	hub         *hub.Hub
	channelRepo repository.ChannelRepo
}

func NewWsHandler(h *hub.Hub, channelRepo repository.ChannelRepo) *WsHandler {
	return &WsHandler{hub: h, channelRepo: channelRepo}
}

func (h *WsHandler) HandleWs(c echo.Context) error {
	ticketParam := c.QueryParam("ticket")
	if ticketParam == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Ticket wajib disertakan"})
	}

	wsTicket, err := h.channelRepo.GetTicketByTicket(ticketParam)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Ticket tidak valid"})
	}

	if wsTicket.ExpiresAt.Before(time.Now()) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Ticket sudah kedaluwarsa"})
	}

	conn, err := configs.WsUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		utils.LogErrors("Gagal upgrade WS: " + err.Error())
		return err
	}

	client := &hub.Client{
		Conn:   conn,
		UserID: wsTicket.UserID,
		Send:   make(chan []byte, 256),
	}

	h.hub.Register(client)

	go func() {
		if err := h.channelRepo.DeleteTicket(wsTicket.ID); err != nil {
			utils.LogErrors("Gagal menghapus ticket terpakai: " + err.Error())
		}
	}()

	go h.readPump(client)
	go h.writePump(client)

	return nil
}

func (h *WsHandler) readPump(client *hub.Client) {
	defer func() {
		h.hub.Unregister(client.UserID)
		client.Conn.Close()
	}()

	client.Conn.SetReadDeadline(time.Now().Add(pongWait))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (h *WsHandler) writePump(client *hub.Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
