package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gopi.com/api/http/dto"
)

// UploadProfileImage handles authenticated profile image upload
// @Summary Upload/Update Profile Image
// @Description Upload or update the authenticated user's profile image
// @Tags Users
// @Accept mpfd
// @Produce json
// @Security Bearer
// @Param image formData file true "Profile image file (png, jpg, jpeg, webp, gif)"
// @Success 200 {object} dto.UserProfileResponse "Profile image updated"
// @Failure 400 {object} dto.AuthErrorResponse "Invalid request or file"
// @Failure 401 {object} dto.AuthErrorResponse "Unauthorized"
// @Failure 500 {object} dto.AuthErrorResponse "Internal server error"
// @Router /user/profile/image [post]
func (h *UserHandler) UploadProfileImage(c *gin.Context) {
    userID := c.GetString("user_id")
    if userID == "" {
        c.JSON(http.StatusUnauthorized, dto.AuthErrorResponse{Error: "Unauthorized", Success: false, StatusCode: http.StatusUnauthorized})
        return
    }

    fileHeader, err := c.FormFile("image")
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.AuthErrorResponse{Error: "image file is required", Success: false, StatusCode: http.StatusBadRequest})
        return
    }

    if fileHeader.Size <= 0 || fileHeader.Size > 10*1024*1024 { // 10MB limit
        c.JSON(http.StatusBadRequest, dto.AuthErrorResponse{Error: "file too large (max 10MB)", Success: false, StatusCode: http.StatusBadRequest})
        return
    }

    // Open and sniff content type
    src, err := fileHeader.Open()
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.AuthErrorResponse{Error: "cannot open uploaded file", Success: false, StatusCode: http.StatusBadRequest})
        return
    }
    defer src.Close()

    // Read first 512 bytes for content-type detection
    buf := make([]byte, 512)
    n, _ := io.ReadFull(src, buf)
    mimeType := http.DetectContentType(buf[:n])

    // Reset reader by re-opening (multipart file doesn't support Seek reliably)
    src.Close()

    // Validate MIME type
    allowed := map[string]string{
        "image/png":  ".png",
        "image/jpeg": ".jpg",
        "image/webp": ".webp",
        "image/gif":  ".gif",
    }
    ext, ok := allowed[mimeType]
    if !ok {
        // Fallback: allow by filename extension if header failed
        lower := strings.ToLower(fileHeader.Filename)
        switch {
        case strings.HasSuffix(lower, ".png"):
            ext = ".png"
        case strings.HasSuffix(lower, ".jpg"), strings.HasSuffix(lower, ".jpeg"):
            ext = ".jpg"
        case strings.HasSuffix(lower, ".webp"):
            ext = ".webp"
        case strings.HasSuffix(lower, ".gif"):
            ext = ".gif"
        default:
            c.JSON(http.StatusBadRequest, dto.AuthErrorResponse{Error: "unsupported image type", Success: false, StatusCode: http.StatusBadRequest})
            return
        }
    }

    // Build filename and storage key
    filename := fmt.Sprintf("%s-%d%s", userID, time.Now().UnixNano(), ext)
    key := "profile/" + filename

    // Re-open reader for upload
    src2, err := fileHeader.Open()
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.AuthErrorResponse{Error: "failed to read file", Success: false, StatusCode: http.StatusInternalServerError})
        return
    }
    defer src2.Close()

    // Save via configured storage backend
    publicURL, err := h.storage.Save(c.Request.Context(), key, src2, fileHeader.Size, mimeType)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.AuthErrorResponse{Error: "failed to store file", Success: false, StatusCode: http.StatusInternalServerError})
        return
    }

    // Update user profile image URL
    user, err := h.userService.GetUserByID(userID)
    if err != nil {
        // clean up uploaded file
        _ = h.storage.Delete(c.Request.Context(), key)
        c.JSON(http.StatusNotFound, dto.AuthErrorResponse{Error: "user not found", Success: false, StatusCode: http.StatusNotFound})
        return
    }
    user.ProfileImageURL = publicURL

    if err := h.userService.UpdateUser(user); err != nil {
        // clean up uploaded file
        _ = h.storage.Delete(c.Request.Context(), key)
        c.JSON(http.StatusInternalServerError, dto.AuthErrorResponse{Error: "failed to update user", Success: false, StatusCode: http.StatusInternalServerError})
        return
    }

    c.JSON(http.StatusOK, dto.UserProfileResponse{
        Success:    true,
        StatusCode: http.StatusOK,
        Data:       h.userModelToDTO(user),
    })
}
