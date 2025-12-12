package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/verbeux-ai/whatsmiau/utils"
	"go.mau.fi/whatsmeow"
)

func Root(ctx echo.Context) error {
	res, err := whatsmeow.GetLatestVersion(ctx.Request().Context(), &http.Client{Timeout: 5 * time.Second})
	if err != nil {
		return utils.HTTPFail(ctx, http.StatusBadRequest, err, "failed to fetch latest version")
	}

	jsonData := map[string]any{
		"status":             200,
		"message":            "Welcome to the Whatsmiau API, a Evolution API alternative, it is working!",
		"version":            "0.3.2",
		"clientName":         "whatsmiau",
		"documentation":      "https://doc.evolution-api.com",
		"whatsappWebVersion": res.String(),
	}

	return ctx.JSON(http.StatusOK, jsonData)
}

// Health check endpoint for Docker
func Health(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]any{
		"status":  "healthy",
		"service": "whatsmiau",
		"time":    time.Now().Unix(),
	})
}

// IPInfo response structure
type IPInfoResponse struct {
	IP       string `json:"ip"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"`
	Org      string `json:"org"`
	Timezone string `json:"timezone"`
}

// GetPublicIP fetches the public IP information from external API
func GetPublicIP(ctx echo.Context) error {
	client := &http.Client{Timeout: 10 * time.Second}

	// Request to ipinfo.io
	resp, err := client.Get("https://ipinfo.io/json")
	if err != nil {
		return utils.HTTPFail(ctx, http.StatusServiceUnavailable, err, "failed to fetch IP information")
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return utils.HTTPFail(ctx, http.StatusInternalServerError, err, "failed to read response")
	}

	// Parse JSON response
	var ipInfo IPInfoResponse
	if err := json.Unmarshal(body, &ipInfo); err != nil {
		return utils.HTTPFail(ctx, http.StatusInternalServerError, err, "failed to parse response")
	}

	// Return formatted response
	return ctx.JSON(http.StatusOK, map[string]any{
		"success": true,
		"message": "Public IP information retrieved successfully",
		"data": map[string]any{
			"ip":       ipInfo.IP,
			"city":     ipInfo.City,
			"region":   ipInfo.Region,
			"country":  ipInfo.Country,
			"location": ipInfo.Loc,
			"isp":      ipInfo.Org,
			"timezone": ipInfo.Timezone,
		},
	})
}
