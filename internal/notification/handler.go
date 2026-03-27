package notification

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(e *echo.Echo) {
	v1 := e.Group("/api/v1")

	v1.POST("/notifications", h.Create)
	v1.POST("/notifications/batch", h.CreateBatch)
	v1.GET("/notifications/:id", h.Get)
	v1.GET("/notifications", h.List)
	v1.PATCH("/notifications/:id/cancel", h.Cancel)
}

func (h *Handler) Create(c echo.Context) error {
	var req CreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
	}
	resp, err := h.service.Create(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusCreated, resp)
}

func (h *Handler) CreateBatch(c echo.Context) error {
	var req BatchCreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
	}
	resp, err := h.service.CreateBatch(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusCreated, resp)
}

func (h *Handler) Get(c echo.Context) error {
	id := c.Param("id")
	resp, err := h.service.Get(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) List(c echo.Context) error {
	var params ListParams
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
	}
	resp, err := h.service.List(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) Cancel(c echo.Context) error {
	id := c.Param("id")
	resp, err := h.service.Cancel(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, resp)
}
