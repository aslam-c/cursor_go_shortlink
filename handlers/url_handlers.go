package handlers

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"url-shortener/models"
)

type URLHandler struct {
	urlService *models.URLService
	baseURL    string
}

type CreateURLRequest struct {
	OriginalURL string `json:"original_url" binding:"required"`
	CustomCode  string `json:"custom_code,omitempty"`
	ExpiresIn   *int   `json:"expires_in,omitempty"` // hours from now
}

type CreateURLResponse struct {
	ID          int       `json:"id"`
	OriginalURL string    `json:"original_url"`
	ShortCode   string    `json:"short_code"`
	ShortURL    string    `json:"short_url"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type URLStatsResponse struct {
	ID          int       `json:"id"`
	OriginalURL string    `json:"original_url"`
	ShortCode   string    `json:"short_code"`
	ShortURL    string    `json:"short_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Clicks      int       `json:"clicks"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type ListURLsResponse struct {
	URLs       []URLStatsResponse `json:"urls"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	PerPage    int               `json:"per_page"`
	HasNext    bool              `json:"has_next"`
}

func NewURLHandler(urlService *models.URLService, baseURL string) *URLHandler {
	return &URLHandler{
		urlService: urlService,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
	}
}

// CreateShortURL creates a new short URL
func (h *URLHandler) CreateShortURL(c *gin.Context) {
	var req CreateURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate URL
	if !isValidURL(req.OriginalURL) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL format"})
		return
	}

	// Validate custom code if provided
	if req.CustomCode != "" {
		if !isValidShortCode(req.CustomCode) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Custom code must be 3-20 characters, alphanumeric and dashes only"})
			return
		}
	}

	// Calculate expiration time
	var expiresAt *time.Time
	if req.ExpiresIn != nil && *req.ExpiresIn > 0 {
		expiry := time.Now().Add(time.Duration(*req.ExpiresIn) * time.Hour)
		expiresAt = &expiry
	}

	// Create the short URL
	urlModel, err := h.urlService.CreateShortURL(req.OriginalURL, req.CustomCode, expiresAt)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create short URL", "details": err.Error()})
		return
	}

	response := CreateURLResponse{
		ID:          urlModel.ID,
		OriginalURL: urlModel.OriginalURL,
		ShortCode:   urlModel.ShortCode,
		ShortURL:    h.baseURL + "/" + urlModel.ShortCode,
		CreatedAt:   urlModel.CreatedAt,
		ExpiresAt:   urlModel.ExpiresAt,
	}

	c.JSON(http.StatusCreated, response)
}

// RedirectURL redirects to the original URL
func (h *URLHandler) RedirectURL(c *gin.Context) {
	shortCode := c.Param("shortCode")
	
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Short code is required"})
		return
	}

	urlModel, err := h.urlService.GetOriginalURL(shortCode)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "expired") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found or expired"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve URL", "details": err.Error()})
		return
	}

	// Redirect to the original URL
	c.Redirect(http.StatusMovedPermanently, urlModel.OriginalURL)
}

// GetURLStats returns statistics for a short URL
func (h *URLHandler) GetURLStats(c *gin.Context) {
	shortCode := c.Param("shortCode")
	
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Short code is required"})
		return
	}

	urlModel, err := h.urlService.GetURLStats(shortCode)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve URL stats", "details": err.Error()})
		return
	}

	response := URLStatsResponse{
		ID:          urlModel.ID,
		OriginalURL: urlModel.OriginalURL,
		ShortCode:   urlModel.ShortCode,
		ShortURL:    h.baseURL + "/" + urlModel.ShortCode,
		CreatedAt:   urlModel.CreatedAt,
		UpdatedAt:   urlModel.UpdatedAt,
		Clicks:      urlModel.Clicks,
		ExpiresAt:   urlModel.ExpiresAt,
	}

	c.JSON(http.StatusOK, response)
}

// ListURLs returns a paginated list of URLs
func (h *URLHandler) ListURLs(c *gin.Context) {
	// Parse query parameters
	pageStr := c.DefaultQuery("page", "1")
	perPageStr := c.DefaultQuery("per_page", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	perPage, err := strconv.Atoi(perPageStr)
	if err != nil || perPage < 1 || perPage > 100 {
		perPage = 10
	}

	offset := (page - 1) * perPage

	urls, err := h.urlService.ListURLs(perPage+1, offset) // Get one extra to check if there's a next page
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve URLs", "details": err.Error()})
		return
	}

	hasNext := len(urls) > perPage
	if hasNext {
		urls = urls[:perPage] // Remove the extra item
	}

	// Convert to response format
	responseURLs := make([]URLStatsResponse, len(urls))
	for i, urlModel := range urls {
		responseURLs[i] = URLStatsResponse{
			ID:          urlModel.ID,
			OriginalURL: urlModel.OriginalURL,
			ShortCode:   urlModel.ShortCode,
			ShortURL:    h.baseURL + "/" + urlModel.ShortCode,
			CreatedAt:   urlModel.CreatedAt,
			UpdatedAt:   urlModel.UpdatedAt,
			Clicks:      urlModel.Clicks,
			ExpiresAt:   urlModel.ExpiresAt,
		}
	}

	response := ListURLsResponse{
		URLs:    responseURLs,
		Total:   len(responseURLs),
		Page:    page,
		PerPage: perPage,
		HasNext: hasNext,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteURL deletes a short URL
func (h *URLHandler) DeleteURL(c *gin.Context) {
	shortCode := c.Param("shortCode")
	
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Short code is required"})
		return
	}

	err := h.urlService.DeleteURL(shortCode)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete URL", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Short URL deleted successfully"})
}

// HealthCheck endpoint
func (h *URLHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "url-shortener",
		"time":    time.Now(),
	})
}

// Helper functions

func isValidURL(str string) bool {
	if str == "" {
		return false
	}

	// Add protocol if missing
	if !strings.HasPrefix(str, "http://") && !strings.HasPrefix(str, "https://") {
		str = "http://" + str
	}

	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func isValidShortCode(code string) bool {
	if len(code) < 3 || len(code) > 20 {
		return false
	}

	for _, char := range code {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '-' || char == '_') {
			return false
		}
	}

	return true
}