package template

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Service interface {
	Create(ctx context.Context, req CreateRequest) (*Response, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Response, error)
	List(ctx context.Context, params ListParams) (*PagedResponse, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*Response, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(e *echo.Echo) {
	v1 := e.Group("/api/v1")

	v1.POST("/templates", h.Create)
	v1.GET("/templates/:id", h.GetByID)
	v1.GET("/templates", h.List)
	v1.PUT("/templates/:id", h.Update)
	v1.DELETE("/templates/:id", h.Delete)
}

// @Summary Create a template
// @Tags templates
// @Accept json
// @Produce json
// @Param request body CreateRequest true "Template payload"
// @Success 201 {object} Response
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/templates [post]
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
		return HandleError(c, err)
	}
	return c.JSON(http.StatusCreated, resp)
}

// @Summary Get template by ID
// @Tags templates
// @Produce json
// @Param id path string true "Template ID"
// @Success 200 {object} Response
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/templates/{id} [get]
func (h *Handler) GetByID(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: ErrInvalidID.Error()})
	}
	resp, err := h.service.GetByID(c.Request().Context(), id)
	if err != nil {
		return HandleError(c, err)
	}
	return c.JSON(http.StatusOK, resp)
}

// @Summary List templates
// @Tags templates
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Success 200 {object} PagedResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/templates [get]
func (h *Handler) List(c echo.Context) error {
	var params ListParams
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
	}
	resp, err := h.service.List(c.Request().Context(), params)
	if err != nil {
		return HandleError(c, err)
	}
	return c.JSON(http.StatusOK, resp)
}

// @Summary Update a template
// @Tags templates
// @Accept json
// @Produce json
// @Param id path string true "Template ID"
// @Param request body UpdateRequest true "Template payload"
// @Success 200 {object} Response
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/v1/templates/{id} [put]
func (h *Handler) Update(c echo.Context) error {
	var req UpdateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: ErrInvalidID.Error()})
	}
	resp, err := h.service.Update(c.Request().Context(), id, req)
	if err != nil {
		return HandleError(c, err)
	}
	return c.JSON(http.StatusOK, resp)
}

// @Summary Delete a template
// @Tags templates
// @Param id path string true "Template ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/templates/{id} [delete]
func (h *Handler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: ErrInvalidID.Error()})
	}
	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		return HandleError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func HandleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return c.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
	case errors.Is(err, ErrNameConflict):
		return c.JSON(http.StatusConflict, ErrorResponse{Message: err.Error()})
	default:
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Internal server error"})
	}
}
