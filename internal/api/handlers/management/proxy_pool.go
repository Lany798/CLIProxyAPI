package management

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
)

// HandleProxyPoolList returns the list of all proxy pool entries
// GET /v0/management/proxy-pool
func (h *Handler) HandleProxyPoolList(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.cfg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "configuration not available"})
		return
	}

	// Return the proxy pool (sensitive ProxyURL data will be masked in response if needed)
	c.JSON(http.StatusOK, gin.H{
		"proxy_pool": h.cfg.ProxyPool,
	})
}

// HandleProxyPoolGet returns a single proxy pool entry by ID
// GET /v0/management/proxy-pool/:id
func (h *Handler) HandleProxyPoolGet(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.cfg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "configuration not available"})
		return
	}

	proxyID := strings.TrimSpace(c.Param("id"))
	if proxyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "proxy ID is required"})
		return
	}

	// Find the proxy entry
	for i := range h.cfg.ProxyPool {
		if h.cfg.ProxyPool[i].ID == proxyID {
			c.JSON(http.StatusOK, h.cfg.ProxyPool[i])
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "proxy not found"})
}

// HandleProxyPoolCreate creates a new proxy pool entry
// POST /v0/management/proxy-pool
func (h *Handler) HandleProxyPoolCreate(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.cfg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "configuration not available"})
		return
	}

	var entry config.ProxyPoolEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	// Validate required fields
	entry.ID = strings.TrimSpace(entry.ID)
	entry.ProxyURL = strings.TrimSpace(entry.ProxyURL)
	if entry.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "proxy ID is required"})
		return
	}
	if entry.ProxyURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "proxy URL is required"})
		return
	}

	// Check for duplicate ID
	for i := range h.cfg.ProxyPool {
		if h.cfg.ProxyPool[i].ID == entry.ID {
			c.JSON(http.StatusConflict, gin.H{"error": "proxy ID already exists"})
			return
		}
	}

	// Add to proxy pool
	h.cfg.ProxyPool = append(h.cfg.ProxyPool, entry)

	// Persist to disk
	if !h.persist(c) {
		return
	}

	c.JSON(http.StatusCreated, entry)
}

// HandleProxyPoolUpdate updates an existing proxy pool entry
// PUT /v0/management/proxy-pool/:id
func (h *Handler) HandleProxyPoolUpdate(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.cfg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "configuration not available"})
		return
	}

	proxyID := strings.TrimSpace(c.Param("id"))
	if proxyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "proxy ID is required"})
		return
	}

	var update config.ProxyPoolEntry
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	// Validate required fields
	update.ProxyURL = strings.TrimSpace(update.ProxyURL)
	if update.ProxyURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "proxy URL is required"})
		return
	}

	// Find and update the entry
	found := false
	for i := range h.cfg.ProxyPool {
		if h.cfg.ProxyPool[i].ID == proxyID {
			// Keep the original ID, update other fields
			h.cfg.ProxyPool[i].Name = strings.TrimSpace(update.Name)
			h.cfg.ProxyPool[i].ProxyURL = update.ProxyURL
			h.cfg.ProxyPool[i].Description = strings.TrimSpace(update.Description)
			found = true

			// Persist to disk
			if !h.persist(c) {
				return
			}

			c.JSON(http.StatusOK, h.cfg.ProxyPool[i])
			return
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "proxy not found"})
	}
}

// HandleProxyPoolDelete deletes a proxy pool entry
// DELETE /v0/management/proxy-pool/:id
func (h *Handler) HandleProxyPoolDelete(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.cfg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "configuration not available"})
		return
	}

	proxyID := strings.TrimSpace(c.Param("id"))
	if proxyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "proxy ID is required"})
		return
	}

	// Find and delete the entry
	for i := range h.cfg.ProxyPool {
		if h.cfg.ProxyPool[i].ID == proxyID {
			// Remove from slice
			h.cfg.ProxyPool = append(h.cfg.ProxyPool[:i], h.cfg.ProxyPool[i+1:]...)

			// Persist to disk
			if !h.persist(c) {
				return
			}

			c.JSON(http.StatusOK, gin.H{"status": "deleted", "id": proxyID})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "proxy not found"})
}
