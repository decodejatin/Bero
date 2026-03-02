package api

import (
	"net/http"

	"github.com/decodejatin/bero-backend/internal/service"
	"github.com/labstack/echo/v4"
)

// AddressHandler handles address endpoints
type AddressHandler struct {
	addressService service.AddressService
}

func NewAddressHandler(addressService service.AddressService) *AddressHandler {
	return &AddressHandler{addressService: addressService}
}

type CreateAddressRequest struct {
	Label       string   `json:"label"`
	FullAddress string   `json:"full_address"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
	IsDefault   bool     `json:"is_default"`
}

type UpdateAddressRequest struct {
	Label       string   `json:"label"`
	FullAddress string   `json:"full_address"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
	IsDefault   bool     `json:"is_default"`
}

// GetAddresses returns all saved addresses for the user
func (h *AddressHandler) GetAddresses(c echo.Context) error {
	userID := c.Get("user_id").(string)

	addresses, err := h.addressService.GetAddresses(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get addresses"})
	}

	return c.JSON(http.StatusOK, addresses)
}

// CreateAddress creates a new saved address
func (h *AddressHandler) CreateAddress(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req CreateAddressRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	if req.Label == "" || req.FullAddress == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "label and full_address are required"})
	}

	address, err := h.addressService.CreateAddress(c.Request().Context(), userID, req.Label, req.FullAddress, req.Latitude, req.Longitude, req.IsDefault)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create address"})
	}

	return c.JSON(http.StatusCreated, address)
}

// UpdateAddress updates an existing address
func (h *AddressHandler) UpdateAddress(c echo.Context) error {
	userID := c.Get("user_id").(string)
	addressID := c.Param("id")

	var req UpdateAddressRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	address, err := h.addressService.UpdateAddress(c.Request().Context(), userID, addressID, req.Label, req.FullAddress, req.Latitude, req.Longitude, req.IsDefault)
	if err != nil {
		if err.Error() == "unauthorized" {
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: "not authorized"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update address"})
	}

	return c.JSON(http.StatusOK, address)
}

// DeleteAddress deletes an address
func (h *AddressHandler) DeleteAddress(c echo.Context) error {
	userID := c.Get("user_id").(string)
	addressID := c.Param("id")

	err := h.addressService.DeleteAddress(c.Request().Context(), userID, addressID)
	if err != nil {
		if err.Error() == "unauthorized" {
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: "not authorized"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete address"})
	}

	return c.JSON(http.StatusOK, SuccessResponse{Message: "address deleted"})
}
