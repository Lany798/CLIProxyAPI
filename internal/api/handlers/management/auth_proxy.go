package management

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// HandleAuthProxyInfo returns the proxy information for a specific auth entry
// GET /v0/management/auth/:id/proxy
func (h *Handler) HandleAuthProxyInfo(c *gin.Context) {
	authID := strings.TrimSpace(c.Param("id"))
	if authID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "auth ID is required"})
		return
	}

	if h.authManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "auth manager not available"})
		return
	}

	// Get the auth entry
	auth, found := h.authManager.GetByID(authID)
	if !found || auth == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "auth not found"})
		return
	}

	// Build proxy info response
	response := gin.H{
		"auth_id":       auth.ID,
		"provider":      auth.Provider,
		"proxy_pool_id": auth.ProxyPoolID,
		"proxy_url":     auth.ProxyURL,
	}

	// Get the actual proxy URL from pool if proxy_pool_id is set
	if auth.ProxyPoolID != "" && h.cfg != nil {
		actualProxyURL := h.cfg.GetProxyFromPool(auth.ProxyPoolID)
		response["resolved_proxy_url"] = actualProxyURL
		if actualProxyURL == "" {
			response["warning"] = "proxy pool ID not found in configuration"
		}
	}

	// Add proxy info message
	proxyInfo := auth.ProxyInfo()
	if proxyInfo != "" {
		response["proxy_info"] = proxyInfo
	} else {
		response["proxy_info"] = "no proxy configured (using global proxy or direct connection)"
	}

	// Get account info
	authType, account := auth.AccountInfo()
	if account != "" {
		response["auth_type"] = authType
		response["account"] = account
	}

	c.JSON(http.StatusOK, response)
}

// HandleListAuthProxies returns proxy information for all auth entries
// GET /v0/management/auth/proxy-list
func (h *Handler) HandleListAuthProxies(c *gin.Context) {
	if h.authManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "auth manager not available"})
		return
	}

	auths := h.authManager.List()

	result := make([]gin.H, 0, len(auths))
	for _, auth := range auths {
		if auth == nil {
			continue
		}

		entry := gin.H{
			"auth_id":       auth.ID,
			"provider":      auth.Provider,
			"proxy_pool_id": auth.ProxyPoolID,
			"status":        string(auth.Status),
		}

		// Get account info
		authType, account := auth.AccountInfo()
		if account != "" {
			entry["auth_type"] = authType
			entry["account"] = account
		}

		// Get proxy info
		proxyInfo := auth.ProxyInfo()
		if proxyInfo != "" {
			entry["proxy_info"] = proxyInfo
		}

		// Resolve proxy pool ID if set
		if auth.ProxyPoolID != "" && h.cfg != nil {
			actualProxyURL := h.cfg.GetProxyFromPool(auth.ProxyPoolID)
			if actualProxyURL != "" {
				entry["resolved_proxy_url"] = maskProxyURL(actualProxyURL)
			} else {
				entry["warning"] = "proxy pool ID not found"
			}
		}

		result = append(result, entry)
	}

	c.JSON(http.StatusOK, gin.H{
		"auths": result,
		"total": len(result),
	})
}

// maskProxyURL masks sensitive information in proxy URL (password)
func maskProxyURL(proxyURL string) string {
	if proxyURL == "" {
		return ""
	}
	// Simple masking: hide password part
	// socks5://user:password@host:port -> socks5://user:****@host:port
	parts := strings.Split(proxyURL, "@")
	if len(parts) != 2 {
		return proxyURL // no auth, return as-is
	}

	credentials := parts[0]
	hostPort := parts[1]

	credParts := strings.Split(credentials, ":")
	if len(credParts) >= 2 {
		// Has password, mask it
		scheme := ""
		username := credParts[len(credParts)-1]
		if len(credParts) == 3 {
			scheme = credParts[0] + "://"
			username = credParts[1]
		}
		return scheme + username + ":****@" + hostPort
	}

	return proxyURL
}
