// Package controller implements HTTP handlers for the API endpoints.
package controller

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/application/usecase/category"
	"github.com/finance-tracker/backend/internal/application/usecase/group"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/dto"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/middleware"
)

// GroupController handles group endpoints.
type GroupController struct {
	createUseCase       *group.CreateGroupUseCase
	listUseCase         *group.ListGroupsUseCase
	getUseCase          *group.GetGroupUseCase
	deleteUseCase       *group.DeleteGroupUseCase
	inviteUseCase       *group.InviteMemberUseCase
	acceptInviteUseCase *group.AcceptInviteUseCase
	changeRoleUseCase   *group.ChangeMemberRoleUseCase
	removeMemberUseCase *group.RemoveMemberUseCase
	leaveUseCase        *group.LeaveGroupUseCase
	getDashboardUseCase *group.GetGroupDashboardUseCase
	// Category use cases for group categories
	listCategoriesUseCase   *category.ListCategoriesUseCase
	createCategoryUseCase   *category.CreateCategoryUseCase
	// Group repository for membership verification
	groupRepo adapter.GroupRepository
}

// NewGroupController creates a new group controller instance.
func NewGroupController(
	createUseCase *group.CreateGroupUseCase,
	listUseCase *group.ListGroupsUseCase,
	getUseCase *group.GetGroupUseCase,
	deleteUseCase *group.DeleteGroupUseCase,
	inviteUseCase *group.InviteMemberUseCase,
	acceptInviteUseCase *group.AcceptInviteUseCase,
	changeRoleUseCase *group.ChangeMemberRoleUseCase,
	removeMemberUseCase *group.RemoveMemberUseCase,
	leaveUseCase *group.LeaveGroupUseCase,
	getDashboardUseCase *group.GetGroupDashboardUseCase,
	listCategoriesUseCase *category.ListCategoriesUseCase,
	createCategoryUseCase *category.CreateCategoryUseCase,
	groupRepo adapter.GroupRepository,
) *GroupController {
	return &GroupController{
		createUseCase:         createUseCase,
		listUseCase:           listUseCase,
		getUseCase:            getUseCase,
		deleteUseCase:         deleteUseCase,
		inviteUseCase:         inviteUseCase,
		acceptInviteUseCase:   acceptInviteUseCase,
		changeRoleUseCase:     changeRoleUseCase,
		removeMemberUseCase:   removeMemberUseCase,
		leaveUseCase:          leaveUseCase,
		getDashboardUseCase:   getDashboardUseCase,
		listCategoriesUseCase: listCategoriesUseCase,
		createCategoryUseCase: createCategoryUseCase,
		groupRepo:             groupRepo,
	}
}

// Create handles POST /groups requests.
func (c *GroupController) Create(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Parse request body
	var req dto.CreateGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body: " + err.Error(),
			Code:  string(domainerror.ErrCodeMissingGroupFields),
		})
		return
	}

	// Build input
	input := group.CreateGroupInput{
		Name:   req.Name,
		UserID: userID,
	}

	// Execute use case
	output, err := c.createUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleGroupError(ctx, err)
		return
	}

	// Build response
	response := dto.ToGroupResponse(output.Group, output.Members)
	ctx.JSON(http.StatusCreated, response)
}

// List handles GET /groups requests.
func (c *GroupController) List(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Build input
	input := group.ListGroupsInput{
		UserID: userID,
	}

	// Execute use case
	output, err := c.listUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve groups",
		})
		return
	}

	// Build response
	response := dto.ToGroupListResponse(output.Groups)
	ctx.JSON(http.StatusOK, response)
}

// Get handles GET /groups/:id requests.
func (c *GroupController) Get(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Parse group ID from URL
	groupIDStr := ctx.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid group ID format",
		})
		return
	}

	// Build input
	input := group.GetGroupInput{
		GroupID: groupID,
		UserID:  userID,
	}

	// Execute use case
	output, err := c.getUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleGroupError(ctx, err)
		return
	}

	// Build response
	response := dto.ToGroupDetailResponse(output.Group, output.Members, output.PendingInvites)
	ctx.JSON(http.StatusOK, response)
}

// Delete handles DELETE /groups/:id requests.
func (c *GroupController) Delete(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Parse group ID from URL
	groupIDStr := ctx.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid group ID format",
		})
		return
	}

	// Build input
	input := group.DeleteGroupInput{
		GroupID:     groupID,
		RequesterID: userID,
	}

	// Execute use case
	_, err = c.deleteUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleGroupError(ctx, err)
		return
	}

	// Return no content on success
	ctx.Status(http.StatusNoContent)
}

// Invite handles POST /groups/:id/invite requests.
func (c *GroupController) Invite(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Parse group ID from URL
	groupIDStr := ctx.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid group ID format",
		})
		return
	}

	// Parse request body
	var req dto.InviteMemberRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body: " + err.Error(),
			Code:  string(domainerror.ErrCodeInvalidGroupEmail),
		})
		return
	}

	// Build input
	input := group.InviteMemberInput{
		GroupID:   groupID,
		Email:     req.Email,
		InviterID: userID,
	}

	// Execute use case
	output, err := c.inviteUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleGroupError(ctx, err)
		return
	}

	// Build response
	response := dto.ToGroupInviteResponse(output.Invite)
	ctx.JSON(http.StatusCreated, response)
}

// AcceptInvite handles POST /groups/invites/:token/accept requests.
func (c *GroupController) AcceptInvite(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Get token from URL
	token := ctx.Param("token")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Token is required",
		})
		return
	}

	// Build input
	input := group.AcceptInviteInput{
		Token:  token,
		UserID: userID,
	}

	// Execute use case
	output, err := c.acceptInviteUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleGroupError(ctx, err)
		return
	}

	// Build response
	response := dto.ToAcceptInviteResponse(output.GroupID.String(), output.GroupName)
	ctx.JSON(http.StatusOK, response)
}

// ChangeRole handles PUT /groups/:id/members/:member_id/role requests.
func (c *GroupController) ChangeRole(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Parse group ID from URL
	groupIDStr := ctx.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid group ID format",
		})
		return
	}

	// Parse member ID from URL
	memberIDStr := ctx.Param("member_id")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid member ID format",
		})
		return
	}

	// Parse request body
	var req dto.ChangeMemberRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body: " + err.Error(),
			Code:  string(domainerror.ErrCodeInvalidMemberRole),
		})
		return
	}

	// Build input
	input := group.ChangeMemberRoleInput{
		GroupID:     groupID,
		MemberID:    memberID,
		NewRole:     entity.MemberRole(req.Role),
		RequesterID: userID,
	}

	// Execute use case
	output, err := c.changeRoleUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleGroupError(ctx, err)
		return
	}

	// Build response
	response := dto.ToMemberRoleResponse(output.Member)
	ctx.JSON(http.StatusOK, response)
}

// RemoveMember handles DELETE /groups/:id/members/:member_id requests.
func (c *GroupController) RemoveMember(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Parse group ID from URL
	groupIDStr := ctx.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid group ID format",
		})
		return
	}

	// Parse member ID from URL
	memberIDStr := ctx.Param("member_id")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid member ID format",
		})
		return
	}

	// Build input
	input := group.RemoveMemberInput{
		GroupID:     groupID,
		MemberID:    memberID,
		RequesterID: userID,
	}

	// Execute use case
	_, err = c.removeMemberUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleGroupError(ctx, err)
		return
	}

	// Return no content on success
	ctx.Status(http.StatusNoContent)
}

// Leave handles DELETE /groups/:id/members/me requests.
func (c *GroupController) Leave(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Parse group ID from URL
	groupIDStr := ctx.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid group ID format",
		})
		return
	}

	// Build input
	input := group.LeaveGroupInput{
		GroupID: groupID,
		UserID:  userID,
	}

	// Execute use case
	_, err = c.leaveUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleGroupError(ctx, err)
		return
	}

	// Return no content on success
	ctx.Status(http.StatusNoContent)
}

// GetDashboard handles GET /groups/:id/dashboard requests.
func (c *GroupController) GetDashboard(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Parse group ID from URL
	groupIDStr := ctx.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid group ID format",
		})
		return
	}

	// Parse period from query parameter
	periodStr := ctx.Query("period")
	period := group.PeriodThisMonth // Default to this month
	switch periodStr {
	case "this_month":
		period = group.PeriodThisMonth
	case "last_month":
		period = group.PeriodLastMonth
	case "this_week":
		period = group.PeriodThisWeek
	case "last_week":
		period = group.PeriodLastWeek
	}

	// Build input
	input := group.GetGroupDashboardInput{
		GroupID: groupID,
		UserID:  userID,
		Period:  period,
	}

	// Execute use case
	output, err := c.getDashboardUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleGroupError(ctx, err)
		return
	}

	// Build response
	response := dto.ToGroupDashboardResponse(output.Dashboard)
	ctx.JSON(http.StatusOK, response)
}

// ListCategories handles GET /groups/:id/categories requests.
func (c *GroupController) ListCategories(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Parse group ID from URL
	groupIDStr := ctx.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid group ID format",
		})
		return
	}

	// Verify user is a member of the group
	isMember, err := c.groupRepo.IsUserMemberOfGroup(ctx.Request.Context(), groupID, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to verify group membership",
		})
		return
	}
	if !isMember {
		ctx.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error: "You are not a member of this group",
			Code:  string(domainerror.ErrCodeNotGroupMember),
		})
		return
	}

	// Build input for listing categories
	input := category.ListCategoriesInput{
		OwnerType: entity.OwnerTypeGroup,
		OwnerID:   groupID,
	}

	// Filter by category type if provided
	if categoryType := ctx.Query("type"); categoryType != "" {
		catType := entity.CategoryType(categoryType)
		input.CategoryType = &catType
	}

	// Parse date range for statistics
	if startDateStr := ctx.Query("startDate"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			input.StartDate = &startDate
		}
	}
	if endDateStr := ctx.Query("endDate"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			input.EndDate = &endDate
		}
	}

	// Execute use case
	output, err := c.listCategoriesUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve categories",
		})
		return
	}

	// Build response
	response := dto.ToCategoryListResponse(output.Categories)
	ctx.JSON(http.StatusOK, response)
}

// CreateCategory handles POST /groups/:id/categories requests.
func (c *GroupController) CreateCategory(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Parse group ID from URL
	groupIDStr := ctx.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid group ID format",
		})
		return
	}

	// Verify user is an admin of the group
	member, err := c.groupRepo.FindMemberByGroupAndUser(ctx.Request.Context(), groupID, userID)
	if err != nil {
		ctx.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error: "You are not a member of this group",
			Code:  string(domainerror.ErrCodeNotGroupMember),
		})
		return
	}
	if member.Role != entity.MemberRoleAdmin {
		ctx.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error: "Only group admins can create categories",
			Code:  string(domainerror.ErrCodeNotGroupAdmin),
		})
		return
	}

	// Parse request body
	var req dto.CreateCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body",
			Code:  string(domainerror.ErrCodeMissingCategoryFields),
		})
		return
	}

	// Set default icon to "tag" for group categories if not provided
	icon := req.Icon
	if icon == "" {
		icon = "tag"
	}

	// Build input
	input := category.CreateCategoryInput{
		Name:      req.Name,
		Color:     req.Color,
		Icon:      icon,
		OwnerType: entity.OwnerTypeGroup,
		OwnerID:   groupID,
		Type:      entity.CategoryType(req.Type),
	}

	// Execute use case
	output, err := c.createCategoryUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleCategoryError(ctx, err)
		return
	}

	// Build response
	response := dto.ToCategoryResponse(output.Category)
	ctx.JSON(http.StatusCreated, response)
}

// handleCategoryError handles category errors and returns appropriate HTTP responses.
func (c *GroupController) handleCategoryError(ctx *gin.Context, err error) {
	var catErr *domainerror.CategoryError
	if errors.As(err, &catErr) {
		statusCode := c.getStatusCodeForCategoryError(catErr.Code)
		ctx.JSON(statusCode, dto.ErrorResponse{
			Error: catErr.Message,
			Code:  string(catErr.Code),
		})
		return
	}

	// Generic server error
	ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
		Error: "An internal error occurred",
	})
}

// getStatusCodeForCategoryError maps category error codes to HTTP status codes.
func (c *GroupController) getStatusCodeForCategoryError(code domainerror.CategoryErrorCode) int {
	switch code {
	case domainerror.ErrCodeCategoryNotFound:
		return http.StatusNotFound
	case domainerror.ErrCodeCategoryNameExists:
		return http.StatusConflict
	case domainerror.ErrCodeNotAuthorizedCategory:
		return http.StatusForbidden
	case domainerror.ErrCodeCategoryNameTooLong,
		domainerror.ErrCodeInvalidColorFormat,
		domainerror.ErrCodeInvalidOwnerType,
		domainerror.ErrCodeInvalidCategoryType,
		domainerror.ErrCodeMissingCategoryFields:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// handleGroupError handles group errors and returns appropriate HTTP responses.
func (c *GroupController) handleGroupError(ctx *gin.Context, err error) {
	var groupErr *domainerror.GroupError
	if errors.As(err, &groupErr) {
		statusCode := c.getStatusCodeForGroupError(groupErr.Code)
		ctx.JSON(statusCode, dto.ErrorResponse{
			Error: groupErr.Message,
			Code:  string(groupErr.Code),
		})
		return
	}

	// Generic server error
	ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
		Error: "An internal error occurred",
	})
}

// getStatusCodeForGroupError maps group error codes to HTTP status codes.
func (c *GroupController) getStatusCodeForGroupError(code domainerror.GroupErrorCode) int {
	switch code {
	case domainerror.ErrCodeGroupNotFound,
		domainerror.ErrCodeMemberNotFound,
		domainerror.ErrCodeInviteNotFound:
		return http.StatusNotFound
	case domainerror.ErrCodeInviteAlreadyExists,
		domainerror.ErrCodeUserAlreadyMember:
		return http.StatusConflict
	case domainerror.ErrCodeNotGroupAdmin,
		domainerror.ErrCodeNotGroupMember:
		return http.StatusForbidden
	case domainerror.ErrCodeGroupNameTooLong,
		domainerror.ErrCodeGroupNameRequired,
		domainerror.ErrCodeInvalidMemberRole,
		domainerror.ErrCodeInvalidGroupEmail,
		domainerror.ErrCodeMissingGroupFields,
		domainerror.ErrCodeCannotRemoveSoleAdmin,
		domainerror.ErrCodeInviteExpired,
		domainerror.ErrCodeCannotInviteSelf:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
