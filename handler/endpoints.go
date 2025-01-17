package handler

import (
	"database/sql"
	"net/http"

	"github.com/fabrianivan-id/technical-test-sawitpro/generated"
	"github.com/fabrianivan-id/technical-test-sawitpro/repository"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Handler to create a new estate
// POST  /estate
func (s *Server) CreateEstate(c echo.Context) error {
	ctx := c.Request().Context()

	var req generated.CreateEstateRequest
	var errResponse generated.ErrorResponse

	if err := c.Bind(&req); err != nil {
		errResponse.Message = "Invalid Request Body"
		return c.JSON(http.StatusBadRequest, errResponse)
	}

	if req.Width < 1 || req.Width > 50000 || req.Length < 1 || req.Length > 50000 {
		errResponse.Message = "Width and Length must be between 1 and 50000"
		return c.JSON(http.StatusBadRequest, errResponse)
	}

	result, err := s.Repository.CreateEstate(ctx, repository.Estate{
		Id:     uuid.New().String(),
		Width:  req.Width,
		Length: req.Length,
	})

	if err != nil {
		errResponse.Message = "Error to Create New Estate"
		return c.JSON(http.StatusBadRequest, errResponse)
	}

	return c.JSON(http.StatusCreated, generated.CreateEstateResponse{
		Id: result.Id,
	})
}

// Handler to create a new tree in an estate
// POST  /estate/{id}/tree
func (s *Server) CreateEstateIdTree(c echo.Context, id string) error {
	ctx := c.Request().Context()

	var req generated.CreateTreeRequest
	var errResponse generated.ErrorResponse

	if err := c.Bind(&req); err != nil {
		errResponse.Message = err.Error()
		return c.JSON(http.StatusBadRequest, errResponse)
	}

	if req.X < 0 || req.Y < 0 || req.Height < 0 || req.Height > 30 {
		errResponse.Message = "Invalid payload X or Y position or height"
		return c.JSON(http.StatusBadRequest, errResponse)
	}

	result, err := s.Repository.CreateEstateTree(ctx, repository.EstateTree{
		Id:       uuid.New().String(),
		EstateId: id,
		X:        req.X,
		Y:        req.Y,
		Height:   req.Height,
	})

	if err != nil {
		errResponse.Message = err.Error()
		return c.JSON(http.StatusBadRequest, errResponse)
	}

	return c.JSON(http.StatusCreated, generated.CreateTreeResponse{
		Id: result.Id,
	})
}

// Handler to get estate stats
// GET  /estate/{id}/stats
func (s *Server) GetEstateIdStats(c echo.Context, id string) error {
	ctx := c.Request().Context()

	result, err := s.Repository.GetStatsByEstateId(ctx, id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, generated.ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, generated.GetEstateStatsResponse{
		Count:  result.Count,
		Max:    result.Max,
		Min:    result.Min,
		Median: int(result.Median),
	})
}

// Handler to get drone plan by estate id
// GET  /estate/{id}/drone-plan
func (s *Server) GetDronePlanByEstateId(c echo.Context, id string) error {
	ctx := c.Request().Context()

	estateData, err := s.Repository.GetEstateById(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, generated.ErrorResponse{
				Message: "Estate id not found",
			})
		}

		return c.JSON(http.StatusBadRequest, generated.ErrorResponse{
			Message: err.Error(),
		})
	}

	horizontalDistance := (estateData.Width-1)*estateData.Length + (estateData.Length-1)*estateData.Width
	verticalDistance := 0

	treesData, err := s.Repository.GetTreesByEstateId(ctx, id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, generated.ErrorResponse{
			Message: err.Error(),
		})
	}

	if len(treesData) > 0 {
		for _, tree := range treesData {
			verticalDistance += tree.Height
		}
	}

	return c.JSON(http.StatusOK, generated.GetDronePlanResponse{
		Distance: horizontalDistance + verticalDistance,
	})
}
