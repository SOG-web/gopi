package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gopi.com/api/http/dto"
	"gopi.com/internal/apperr"
)

// AdminSearchGroups godoc
// @Summary Admin: search chat groups by name
// @Description Case-insensitive search over group names with pagination (staff only)
// @Tags chat,admin
// @Security BearerAuth
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.GroupListResponse "Groups retrieved successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - staff only"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /chat/admin/groups/search [get]
func (h *ChatHandler) AdminSearchGroups(c *gin.Context) {
	var req dto.SearchGroupsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		respondError(c, apperr.E("AdminSearchGroups", apperr.InvalidInput, err, "Invalid query parameters"))
		return
	}

	offset := (req.Page - 1) * req.Limit
	groups, err := h.chatService.SearchGroupsByName(req.Query, req.Limit, offset)
	if err != nil {
		respondError(c, apperr.E("AdminSearchGroups", apperr.Internal, err, "Failed to search groups"))
		return
	}

	responses := make([]dto.GroupResponse, 0, len(groups))
	for _, g := range groups {
		responses = append(responses, h.groupToResponse(g))
	}

	c.JSON(http.StatusOK, dto.GroupListResponse{
		Groups: responses,
		Total:  len(responses),
		Page:   req.Page,
		Limit:  req.Limit,
	})
}
