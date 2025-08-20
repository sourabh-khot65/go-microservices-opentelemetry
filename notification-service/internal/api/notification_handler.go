package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type NotificationHandler struct{}

func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{}
}

func (h *NotificationHandler) Router() *gin.Engine {
	router := gin.Default()
	router.GET("/notifications", h.HandleNotification)
	return router
}

func (h *NotificationHandler) HandleNotification(c *gin.Context) {
	c.String(http.StatusOK, "notification endpoint")
}