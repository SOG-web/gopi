package handler

import (
	"net/http"
	"strconv"
    "fmt"
    "io"
    "strings"
    "time"

	"github.com/gin-gonic/gin"
	"gopi.com/api/http/dto"
	postApp "gopi.com/internal/app/post"
    "gopi.com/internal/lib/storage"
)

type PostHandler struct {
    service *postApp.Service
    storage storage.Storage
}

func NewPostHandler(svc *postApp.Service, st storage.Storage) *PostHandler {
    return &PostHandler{service: svc, storage: st}
}

// Public endpoints

// ListPublishedPosts godoc
// @Summary List published posts
// @Description Get a paginated list of published posts
// @Tags posts
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(20)
// @Success 200 {object} dto.ListPostsResponse
// @Failure 500 {object} dto.PostResponse
// @Router /posts [get]
func (h *PostHandler) ListPublishedPosts(c *gin.Context) {
    limit, offset := parsePagination(c, 20)
    posts, err := h.service.ListPublished(limit, offset)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.PostResponse{Success: false, StatusCode: http.StatusInternalServerError, Message: err.Error()})
        return
    }
    c.JSON(http.StatusOK, dto.ListPostsResponse{Success: true, StatusCode: http.StatusOK, Data: posts, Count: len(posts)})
}

// GetPostBySlug godoc
// @Summary Get post by slug
// @Description Retrieve a single published post by slug
// @Tags posts
// @Produce json
// @Param slug path string true "Post slug"
// @Success 200 {object} dto.PostResponse
// @Failure 404 {object} dto.PostResponse
// @Router /posts/{slug} [get]
func (h *PostHandler) GetPostBySlug(c *gin.Context) {
    slug := c.Param("slug")
    post, err := h.service.GetPostBySlug(slug)
    if err != nil {
        c.JSON(http.StatusNotFound, dto.PostResponse{Success: false, StatusCode: http.StatusNotFound, Message: "post not found"})
        return
    }
    c.JSON(http.StatusOK, dto.PostResponse{Success: true, StatusCode: http.StatusOK, Data: post})
}

// Admin endpoints

// CreatePost godoc
// @Summary Create a post
// @Description Create a new post (staff only)
// @Tags posts
// @Security Bearer
// @Accept json
// @Produce json
// @Param post body dto.CreatePostRequest true "Post payload"
// @Success 201 {object} dto.PostResponse
// @Failure 400 {object} dto.PostResponse
// @Failure 401 {object} dto.PostResponse
// @Failure 403 {object} dto.PostResponse
// @Router /posts/admin [post]
func (h *PostHandler) CreatePost(c *gin.Context) {
    var req dto.CreatePostRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.PostResponse{Success: false, StatusCode: http.StatusBadRequest, Message: err.Error()})
        return
    }
    authorID := c.GetString("user_id")
    post, err := h.service.CreatePost(authorID, req.Title, req.Content, req.CoverImageURL, req.Publish)
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.PostResponse{Success: false, StatusCode: http.StatusBadRequest, Message: err.Error()})
        return
    }
    c.JSON(http.StatusCreated, dto.PostResponse{Success: true, StatusCode: http.StatusCreated, Data: post})
}

// UpdatePost godoc
// @Summary Update a post
// @Description Update a post (staff only)
// @Tags posts
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path string true "Post ID"
// @Param post body dto.UpdatePostRequest true "Post update payload"
// @Success 200 {object} dto.PostResponse
// @Failure 400 {object} dto.PostResponse
// @Failure 401 {object} dto.PostResponse
// @Failure 403 {object} dto.PostResponse
// @Router /posts/admin/{id} [put]
func (h *PostHandler) UpdatePost(c *gin.Context) {
    postID := c.Param("id")
    var req dto.UpdatePostRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.PostResponse{Success: false, StatusCode: http.StatusBadRequest, Message: err.Error()})
        return
    }
    post, err := h.service.UpdatePost(postID, req.Title, req.Content, req.CoverImageURL)
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.PostResponse{Success: false, StatusCode: http.StatusBadRequest, Message: err.Error()})
        return
    }
    c.JSON(http.StatusOK, dto.PostResponse{Success: true, StatusCode: http.StatusOK, Data: post})
}

// PublishPost godoc
// @Summary Publish a post
// @Description Publish a post (staff only)
// @Tags posts
// @Security Bearer
// @Produce json
// @Param id path string true "Post ID"
// @Success 200 {object} dto.PostResponse
// @Failure 400 {object} dto.PostResponse
// @Failure 401 {object} dto.PostResponse
// @Failure 403 {object} dto.PostResponse
// @Router /posts/admin/{id}/publish [post]
func (h *PostHandler) PublishPost(c *gin.Context) {
    postID := c.Param("id")
    post, err := h.service.PublishPost(postID)
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.PostResponse{Success: false, StatusCode: http.StatusBadRequest, Message: err.Error()})
        return
    }
    c.JSON(http.StatusOK, dto.PostResponse{Success: true, StatusCode: http.StatusOK, Data: post})
}

// UnpublishPost godoc
// @Summary Unpublish a post
// @Description Unpublish a post (staff only)
// @Tags posts
// @Security Bearer
// @Produce json
// @Param id path string true "Post ID"
// @Success 200 {object} dto.PostResponse
// @Failure 400 {object} dto.PostResponse
// @Failure 401 {object} dto.PostResponse
// @Failure 403 {object} dto.PostResponse
// @Router /posts/admin/{id}/unpublish [post]
func (h *PostHandler) UnpublishPost(c *gin.Context) {
    postID := c.Param("id")
    post, err := h.service.UnpublishPost(postID)
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.PostResponse{Success: false, StatusCode: http.StatusBadRequest, Message: err.Error()})
        return
    }
    c.JSON(http.StatusOK, dto.PostResponse{Success: true, StatusCode: http.StatusOK, Data: post})
}

// DeletePost godoc
// @Summary Delete a post
// @Description Delete a post. Only the author or a superuser may delete.
// @Tags posts
// @Security Bearer
// @Produce json
// @Param id path string true "Post ID"
// @Success 200 {object} dto.PostResponse
// @Failure 400 {object} dto.PostResponse
// @Failure 401 {object} dto.PostResponse
// @Failure 403 {object} dto.PostResponse
// @Failure 404 {object} dto.PostResponse
// @Router /posts/{id} [delete]
func (h *PostHandler) DeletePost(c *gin.Context) {
    postID := c.Param("id")
    requesterID := c.GetString("user_id")
    isSuperuserAny, _ := c.Get("is_superuser")
    isSuperuser := false
    if v, ok := isSuperuserAny.(bool); ok {
        isSuperuser = v
    }

    p, err := h.service.GetPostByID(postID)
    if err != nil {
        c.JSON(http.StatusNotFound, dto.PostResponse{Success: false, StatusCode: http.StatusNotFound, Message: "post not found"})
        return
    }

    if p.AuthorID != requesterID && !isSuperuser {
        c.JSON(http.StatusForbidden, dto.PostResponse{Success: false, StatusCode: http.StatusForbidden, Message: "only author or superuser can delete the post"})
        return
    }

    if err := h.service.DeletePost(postID); err != nil {
        c.JSON(http.StatusBadRequest, dto.PostResponse{Success: false, StatusCode: http.StatusBadRequest, Message: err.Error()})
        return
    }
    c.JSON(http.StatusOK, dto.PostResponse{Success: true, StatusCode: http.StatusOK, Message: "post deleted"})
}

// Comments endpoints

// CreateComment godoc
// @Summary Create a comment
// @Description Create a comment on any target (auth required)
// @Tags comments
// @Security Bearer
// @Accept json
// @Produce json
// @Param comment body dto.CreateCommentRequest true "Comment payload"
// @Success 201 {object} dto.CommentResponse
// @Failure 400 {object} dto.CommentResponse
// @Failure 401 {object} dto.CommentResponse
// @Router /comments [post]
func (h *PostHandler) CreateComment(c *gin.Context) {
    var req dto.CreateCommentRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.CommentResponse{Success: false, StatusCode: http.StatusBadRequest, Message: err.Error()})
        return
    }
    authorID := c.GetString("user_id")
    comment, err := h.service.CreateComment(authorID, req.TargetType, req.TargetID, req.Content, req.ParentID)
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.CommentResponse{Success: false, StatusCode: http.StatusBadRequest, Message: err.Error()})
        return
    }
    c.JSON(http.StatusCreated, dto.CommentResponse{Success: true, StatusCode: http.StatusCreated, Data: comment})
}

// UpdateComment godoc
// @Summary Update a comment
// @Description Update a comment (author only)
// @Tags comments
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path string true "Comment ID"
// @Param comment body dto.UpdateCommentRequest true "Comment update payload"
// @Success 200 {object} dto.CommentResponse
// @Failure 400 {object} dto.CommentResponse
// @Failure 401 {object} dto.CommentResponse
// @Failure 403 {object} dto.CommentResponse
// @Failure 404 {object} dto.CommentResponse
// @Router /comments/{id} [put]
func (h *PostHandler) UpdateComment(c *gin.Context) {
    commentID := c.Param("id")
    var req dto.UpdateCommentRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.CommentResponse{Success: false, StatusCode: http.StatusBadRequest, Message: err.Error()})
        return
    }
    requesterID := c.GetString("user_id")
    comment, err := h.service.UpdateComment(commentID, requesterID, req.Content)
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.CommentResponse{Success: false, StatusCode: http.StatusBadRequest, Message: err.Error()})
        return
    }
    c.JSON(http.StatusOK, dto.CommentResponse{Success: true, StatusCode: http.StatusOK, Data: comment})
}

// DeleteComment godoc
// @Summary Delete a comment
// @Description Delete a comment (author only)
// @Tags comments
// @Security Bearer
// @Produce json
// @Param id path string true "Comment ID"
// @Success 200 {object} dto.CommentResponse
// @Failure 400 {object} dto.CommentResponse
// @Failure 401 {object} dto.CommentResponse
// @Failure 403 {object} dto.CommentResponse
// @Failure 404 {object} dto.CommentResponse
// @Router /comments/{id} [delete]
func (h *PostHandler) DeleteComment(c *gin.Context) {
    commentID := c.Param("id")
    requesterID := c.GetString("user_id")
    if err := h.service.DeleteComment(commentID, requesterID); err != nil {
        c.JSON(http.StatusBadRequest, dto.CommentResponse{Success: false, StatusCode: http.StatusBadRequest, Message: err.Error()})
        return
    }
    c.JSON(http.StatusOK, dto.CommentResponse{Success: true, StatusCode: http.StatusOK, Message: "comment deleted"})
}

// ListCommentsByTarget godoc
// @Summary List comments for a target
// @Description Get comments for a target type and ID
// @Tags comments
// @Produce json
// @Param targetType path string true "Target type"
// @Param targetID path string true "Target ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(50)
// @Success 200 {object} dto.ListCommentsResponse
// @Failure 500 {object} dto.CommentResponse
// @Router /comments/{targetType}/{targetID} [get]
func (h *PostHandler) ListCommentsByTarget(c *gin.Context) {
    targetType := c.Param("targetType")
    targetID := c.Param("targetID")
    limit, offset := parsePagination(c, 50)
    comments, err := h.service.ListCommentsByTarget(targetType, targetID, limit, offset)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.CommentResponse{Success: false, StatusCode: http.StatusInternalServerError, Message: err.Error()})
        return
    }
    c.JSON(http.StatusOK, dto.ListCommentsResponse{Success: true, StatusCode: http.StatusOK, Data: comments, Count: len(comments)})
}

// UploadCoverImage godoc
// @Summary Upload post cover image
// @Description Upload or update the cover image for a post. Only the author, staff, or superuser may upload.
// @Tags posts
// @Security Bearer
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Post ID"
// @Param image formData file true "Cover image file (png, jpg, jpeg, webp, gif)"
// @Success 200 {object} dto.PostResponse
// @Failure 400 {object} dto.PostResponse
// @Failure 401 {object} dto.PostResponse
// @Failure 403 {object} dto.PostResponse
// @Failure 404 {object} dto.PostResponse
// @Failure 500 {object} dto.PostResponse
// @Router /posts/{id}/cover [post]
func (h *PostHandler) UploadCoverImage(c *gin.Context) {
    userID := c.GetString("user_id")
    if userID == "" {
        c.JSON(http.StatusUnauthorized, dto.PostResponse{Success: false, StatusCode: http.StatusUnauthorized, Message: "unauthorized"})
        return
    }

    postID := c.Param("id")
    p, err := h.service.GetPostByID(postID)
    if err != nil {
        c.JSON(http.StatusNotFound, dto.PostResponse{Success: false, StatusCode: http.StatusNotFound, Message: "post not found"})
        return
    }

    isStaffAny, _ := c.Get("is_staff")
    isSuperuserAny, _ := c.Get("is_superuser")
    isStaff := false
    isSuperuser := false
    if v, ok := isStaffAny.(bool); ok { isStaff = v }
    if v, ok := isSuperuserAny.(bool); ok { isSuperuser = v }
    if p.AuthorID != userID && !isStaff && !isSuperuser {
        c.JSON(http.StatusForbidden, dto.PostResponse{Success: false, StatusCode: http.StatusForbidden, Message: "forbidden"})
        return
    }

    fileHeader, err := c.FormFile("image")
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.PostResponse{Success: false, StatusCode: http.StatusBadRequest, Message: "image file is required"})
        return
    }
    if fileHeader.Size <= 0 || fileHeader.Size > 10*1024*1024 { // 10MB limit
        c.JSON(http.StatusBadRequest, dto.PostResponse{Success: false, StatusCode: http.StatusBadRequest, Message: "file too large (max 10MB)"})
        return
    }

    src, err := fileHeader.Open()
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.PostResponse{Success: false, StatusCode: http.StatusBadRequest, Message: "cannot open uploaded file"})
        return
    }
    defer src.Close()

    // Sniff MIME type
    buf := make([]byte, 512)
    n, _ := io.ReadFull(src, buf)
    mimeType := http.DetectContentType(buf[:n])
    _ = src.Close()

    allowed := map[string]string{
        "image/png":  ".png",
        "image/jpeg": ".jpg",
        "image/webp": ".webp",
        "image/gif":  ".gif",
    }
    ext, ok := allowed[mimeType]
    if !ok {
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
            c.JSON(http.StatusBadRequest, dto.PostResponse{Success: false, StatusCode: http.StatusBadRequest, Message: "unsupported image type"})
            return
        }
    }

    filename := fmt.Sprintf("%s-%d%s", postID, time.Now().UnixNano(), ext)
    key := "posts/" + filename

    src2, err := fileHeader.Open()
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.PostResponse{Success: false, StatusCode: http.StatusInternalServerError, Message: "failed to read file"})
        return
    }
    defer src2.Close()

    publicURL, err := h.storage.Save(c.Request.Context(), key, src2, fileHeader.Size, mimeType)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.PostResponse{Success: false, StatusCode: http.StatusInternalServerError, Message: "failed to store file"})
        return
    }

    // Update post cover image URL
    updated, err := h.service.UpdatePost(postID, "", "", publicURL)
    if err != nil {
        // best-effort cleanup
        _ = h.storage.Delete(c.Request.Context(), key)
        c.JSON(http.StatusInternalServerError, dto.PostResponse{Success: false, StatusCode: http.StatusInternalServerError, Message: "failed to update post"})
        return
    }

    c.JSON(http.StatusOK, dto.PostResponse{Success: true, StatusCode: http.StatusOK, Data: updated})
}

// helper
func parsePagination(c *gin.Context, defaultLimit int) (int, int) {
    pageStr := c.Query("page")
    limitStr := c.Query("limit")
    page := 1
    limit := defaultLimit
    var err error
    if pageStr != "" {
        if page, err = strconv.Atoi(pageStr); err != nil || page < 1 {
            page = 1
        }
    }
    if limitStr != "" {
        if limit, err = strconv.Atoi(limitStr); err != nil || limit < 1 || limit > 100 {
            limit = defaultLimit
        }
    }
    offset := (page - 1) * limit
    return limit, offset
}
