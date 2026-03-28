package notification

import (
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Service interface {
	Create(ctx context.Context, req CreateRequest) (*Response, error)
	CreateBatch(ctx context.Context, req BatchCreateRequest) (*BatchCreateResponse, error)
	GetByID(ctx context.Context, id string) (*Response, error)
	List(ctx context.Context, params ListParams) (*PagedResponse, error)
	Cancel(ctx context.Context, id string) (*Response, error)
}

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
	v1.GET("/notifications/:id", h.GetByID)
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
		return handleError(c, err)
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
		return handleError(c, err)
	}
	return c.JSON(http.StatusCreated, resp)
}

func (h *Handler) GetByID(c echo.Context) error {
	id := c.Param("id")
	resp, err := h.service.GetByID(c.Request().Context(), id)
	if err != nil {
		return handleError(c, err)
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
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) Cancel(c echo.Context) error {
	id := c.Param("id")
	resp, err := h.service.Cancel(c.Request().Context(), id)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, resp)
}

func handleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrInvalidID):
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
	case errors.Is(err, ErrNotFound):
		return c.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
	case errors.Is(err, ErrNotCancellable):
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
	default:
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}
}
