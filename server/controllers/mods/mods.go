package mods

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"ark-server-commander/middleware"
	"ark-server-commander/models"
	modsservice "ark-server-commander/service/mods"
	"ark-server-commander/service/steam"

	"github.com/gin-gonic/gin"
)

var svc = modsservice.NewService()

// serverIDParam parses :id and reports whether it was usable.
func serverIDParam(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid server id"})
		return 0, false
	}
	return uint(id), true
}

// SearchWorkshop godoc
// @Summary Search the ARK Steam Workshop
// @Tags mods
// @Produce json
// @Router /mods/search [get]
func SearchWorkshop(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	items, err := svc.Search(c.Request.Context(), c.Query("q"), page, limit)
	if err != nil {
		// Missing API key is a configuration problem, not a server fault, and
		// the UI shows a specific "add by ID instead" message for it.
		if errors.Is(err, steam.ErrSearchUnavailable) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "workshop search unavailable: STEAM_API_KEY is not configured"})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": "workshop search failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Operation successful", "data": items})
}

// LookupWorkshopItems godoc
// @Summary Look up Workshop items by ID
// @Tags mods
// @Produce json
// @Router /mods/lookup [get]
func LookupWorkshopItems(c *gin.Context) {
	raw := strings.TrimSpace(c.Query("ids"))
	if raw == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids is required"})
		return
	}

	ids := strings.Split(raw, ",")
	if len(ids) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "too many ids (max 50)"})
		return
	}
	for i := range ids {
		ids[i] = strings.TrimSpace(ids[i])
	}

	items, err := svc.GetItems(c.Request.Context(), ids)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Operation successful", "data": items})
}

// ListServerMods godoc
// @Summary List a server's mods in load order
// @Tags mods
// @Produce json
// @Router /servers/{id}/mods [get]
func ListServerMods(c *gin.Context) {
	serverID, ok := serverIDParam(c)
	if !ok {
		return
	}

	list, err := svc.ListMods(serverID, c.GetUint("user_id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Operation successful", "data": list})
}

// AddServerMod godoc
// @Summary Attach a Workshop mod to a server
// @Tags mods
// @Produce json
// @Router /servers/{id}/mods [post]
func AddServerMod(c *gin.Context) {
	serverID, ok := serverIDParam(c)
	if !ok {
		return
	}

	var req models.ServerModRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workshop_id is required"})
		return
	}

	mod, err := svc.AddMod(c.Request.Context(), serverID, c.GetUint("user_id"), req.WorkshopID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	middleware.Log.Log(c.GetUint("user_id"), "mods.add", "server", req.WorkshopID, c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Operation successful", "data": mod})
}

// RemoveServerMod godoc
// @Summary Detach a mod from a server
// @Tags mods
// @Produce json
// @Router /servers/{id}/mods/{workshopId} [delete]
func RemoveServerMod(c *gin.Context) {
	serverID, ok := serverIDParam(c)
	if !ok {
		return
	}
	workshopID := c.Param("workshopId")

	if err := svc.RemoveMod(serverID, c.GetUint("user_id"), workshopID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	middleware.Log.Log(c.GetUint("user_id"), "mods.remove", "server", workshopID, c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Operation successful"})
}

// ToggleServerMod godoc
// @Summary Enable or disable a mod without losing its load position
// @Tags mods
// @Produce json
// @Router /servers/{id}/mods/{workshopId}/enabled [put]
func ToggleServerMod(c *gin.Context) {
	serverID, ok := serverIDParam(c)
	if !ok {
		return
	}

	var req struct {
		Enabled *bool `json:"enabled" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Enabled == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "enabled is required"})
		return
	}

	if err := svc.SetEnabled(serverID, c.GetUint("user_id"), c.Param("workshopId"), *req.Enabled); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Operation successful"})
}

// ReorderServerMods godoc
// @Summary Set the full mod load order
// @Tags mods
// @Produce json
// @Router /servers/{id}/mods/order [put]
func ReorderServerMods(c *gin.Context) {
	serverID, ok := serverIDParam(c)
	if !ok {
		return
	}

	var req models.ServerModReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workshop_ids is required"})
		return
	}

	if err := svc.Reorder(serverID, c.GetUint("user_id"), req.WorkshopIDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Operation successful"})
}
