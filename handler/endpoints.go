package handler

import (
	"database/sql"
	"net/http"

	"github.com/fabrianivan-id/technical-test-sawitpro/generated"
	"github.com/fabrianivan-id/technical-test-sawitpro/repository"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	maxDimension     = 50000
	maxTreeHeight    = 30
	errInvalidWidth  = "Width must be between 1 and 50000"
	errInvalidLength = "Length must be between 1 and 50000"
	errInvalidTree   = "Invalid X, Y position or Height (0-30 allowed)"
)

// Helper for validation
func validateDimension(value int, max int, field string) (bool, string) {
	if value <= 0 || value > max {
		return false, field + " must be between 1 and " + string(max)
	}
	return true, ""
}

// CREATE ESTATE DATA HANDLER
func (s *Server) CreateEstate(c echo.Context) error {
	ctx := c.Request().Context()

	var req generated.CreateEstateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: "Invalid request body"})
	}

	if ok, msg := validateDimension(req.Width, maxDimension, "Width"); !ok {
		return c.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: msg})
	}

	if ok, msg := validateDimension(req.Length, maxDimension, "Length"); !ok {
		return c.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: msg})
	}

	result, err := s.Repository.CreateEstate(ctx, repository.Estate{
		Id:     uuid.New().String(),
		Width:  req.Width,
		Length: req.Length,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: "Error creating estate: " + err.Error()})
	}

	return c.JSON(http.StatusCreated, generated.CreateEstateResponse{Id: result.Id})
}

// CREATE TREE DATA HANDLER
func (s *Server) CreateEstateIdTree(c echo.Context, id string) error {
	ctx := c.Request().Context()

	var req generated.CreateTreeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: "Invalid request body"})
	}

	if req.X < 0 || req.Y < 0 || req.Height < 0 || req.Height > maxTreeHeight {
		return c.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: errInvalidTree})
	}

	result, err := s.Repository.CreateEstateTree(ctx, repository.EstateTree{
		Id:       uuid.New().String(),
		EstateId: id,
		X:        req.X,
		Y:        req.Y,
		Height:   req.Height,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: "Error creating tree: " + err.Error()})
	}

	return c.JSON(http.StatusCreated, generated.CreateTreeResponse{Id: result.Id})
}

// GETTING ESTATE STATISTICS DATA HANDLER
func (s *Server) GetEstateIdStats(c echo.Context, id string) error {
	ctx := c.Request().Context()

	result, err := s.Repository.GetStatsByEstateId(ctx, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: "Error fetching estate stats: " + err.Error()})
	}

	return c.JSON(http.StatusOK, generated.GetEstateStatsResponse{
		Count:  result.Count,
		Max:    result.Max,
		Min:    result.Min,
		Median: int(result.Median),
	})
}

// GETTING ESTATE DRONE PLAN DATA HANBDLER
func (s *Server) GetDronePlanByEstateId(c echo.Context, id string) error {
	ctx := c.Request().Context()

	estateData, err := s.Repository.GetEstateById(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, generated.ErrorResponse{Message: "Estate not found"})
		}
		return c.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: "Error fetching estate: " + err.Error()})
	}

	horizontalDistance := (estateData.Width-1)*estateData.Length + (estateData.Length-1)*estateData.Width
	verticalDistance := 0

	treesData, err := s.Repository.GetTreesByEstateId(ctx, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, generated.ErrorResponse{Message: "Error fetching trees: " + err.Error()})
	}

	for _, tree := range treesData {
		verticalDistance += tree.Height
	}

	return c.JSON(http.StatusOK, generated.GetDronePlanResponse{
		Distance: horizontalDistance + verticalDistance,
	})
}
