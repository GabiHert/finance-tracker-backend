// Package dto defines data transfer objects for API requests and responses.
package dto

import (
	"fmt"
	"time"

	"github.com/finance-tracker/backend/internal/domain/entity"
)

// CreateGroupRequest represents the request body for group creation.
type CreateGroupRequest struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

// InviteMemberRequest represents the request body for inviting a member.
type InviteMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ChangeMemberRoleRequest represents the request body for changing a member's role.
type ChangeMemberRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin member"`
}

// GroupResponse represents a single group in API responses.
type GroupResponse struct {
	ID        string               `json:"id"`
	Name      string               `json:"name"`
	CreatedBy string               `json:"created_by"`
	CreatedAt time.Time            `json:"created_at"`
	Members   []GroupMemberResponse `json:"members,omitempty"`
}

// GroupListResponse represents the response for listing groups.
type GroupListResponse struct {
	Groups []GroupListItemResponse `json:"groups"`
}

// GroupListItemResponse represents a group in list view.
type GroupListItemResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	MemberCount int       `json:"member_count"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
}

// GroupDetailResponse represents detailed group information.
type GroupDetailResponse struct {
	ID             string                `json:"id"`
	Name           string                `json:"name"`
	CreatedBy      string                `json:"created_by"`
	CreatedAt      time.Time             `json:"created_at"`
	Members        []GroupMemberResponse `json:"members"`
	PendingInvites []GroupInviteResponse `json:"pending_invites,omitempty"`
}

// GroupMemberResponse represents a group member in API responses.
type GroupMemberResponse struct {
	ID       string    `json:"id"`
	UserID   string    `json:"user_id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

// GroupInviteResponse represents a group invitation in API responses.
type GroupInviteResponse struct {
	ID        string    `json:"id,omitempty"`
	Email     string    `json:"email"`
	Token     string    `json:"token,omitempty"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// AcceptInviteResponse represents the response for accepting an invitation.
type AcceptInviteResponse struct {
	GroupID   string `json:"group_id"`
	GroupName string `json:"group_name"`
}

// MemberRoleResponse represents the response for role change.
type MemberRoleResponse struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

// ToGroupResponse converts a domain Group entity to a GroupResponse DTO.
func ToGroupResponse(group *entity.Group, members []*entity.GroupMember) GroupResponse {
	response := GroupResponse{
		ID:        group.ID.String(),
		Name:      group.Name,
		CreatedBy: group.CreatedBy.String(),
		CreatedAt: group.CreatedAt,
		Members:   make([]GroupMemberResponse, len(members)),
	}

	for i, m := range members {
		response.Members[i] = ToGroupMemberResponse(m)
	}

	return response
}

// ToGroupListResponse converts a list of GroupListItem to GroupListResponse.
func ToGroupListResponse(groups []*entity.GroupListItem) GroupListResponse {
	items := make([]GroupListItemResponse, len(groups))
	for i, g := range groups {
		items[i] = GroupListItemResponse{
			ID:          g.ID.String(),
			Name:        g.Name,
			MemberCount: g.MemberCount,
			Role:        string(g.Role),
			CreatedAt:   g.CreatedAt,
		}
	}
	return GroupListResponse{
		Groups: items,
	}
}

// ToGroupDetailResponse converts group data to a detailed response.
func ToGroupDetailResponse(group *entity.Group, members []*entity.GroupMember, invites []*entity.GroupInvite) GroupDetailResponse {
	response := GroupDetailResponse{
		ID:             group.ID.String(),
		Name:           group.Name,
		CreatedBy:      group.CreatedBy.String(),
		CreatedAt:      group.CreatedAt,
		Members:        make([]GroupMemberResponse, len(members)),
		PendingInvites: make([]GroupInviteResponse, len(invites)),
	}

	for i, m := range members {
		response.Members[i] = ToGroupMemberResponse(m)
	}

	for i, inv := range invites {
		response.PendingInvites[i] = GroupInviteResponse{
			ID:        inv.ID.String(),
			Email:     inv.Email,
			Status:    string(inv.Status),
			ExpiresAt: inv.ExpiresAt,
			CreatedAt: inv.CreatedAt,
		}
	}

	return response
}

// ToGroupMemberResponse converts a domain GroupMember entity to a response DTO.
func ToGroupMemberResponse(member *entity.GroupMember) GroupMemberResponse {
	return GroupMemberResponse{
		ID:       member.ID.String(),
		UserID:   member.UserID.String(),
		Name:     member.UserName,
		Email:    member.UserEmail,
		Role:     string(member.Role),
		JoinedAt: member.JoinedAt,
	}
}

// ToGroupInviteResponse converts a domain GroupInvite entity to a response DTO.
func ToGroupInviteResponse(invite *entity.GroupInvite) GroupInviteResponse {
	return GroupInviteResponse{
		ID:        invite.ID.String(),
		Email:     invite.Email,
		Token:     invite.Token,
		Status:    string(invite.Status),
		ExpiresAt: invite.ExpiresAt,
		CreatedAt: invite.CreatedAt,
	}
}

// ToAcceptInviteResponse converts accept invite output to a response DTO.
func ToAcceptInviteResponse(groupID, groupName string) AcceptInviteResponse {
	return AcceptInviteResponse{
		GroupID:   groupID,
		GroupName: groupName,
	}
}

// ToMemberRoleResponse converts member role change output to a response DTO.
func ToMemberRoleResponse(member *entity.GroupMember) MemberRoleResponse {
	return MemberRoleResponse{
		ID:     member.ID.String(),
		UserID: member.UserID.String(),
		Role:   string(member.Role),
	}
}

// GroupDashboardResponse represents the comprehensive group dashboard data.
type GroupDashboardResponse struct {
	Summary           GroupDashboardSummary         `json:"summary"`
	CategoryBreakdown []GroupCategoryBreakdown      `json:"category_breakdown"`
	MemberBreakdown   []GroupMemberBreakdown        `json:"member_breakdown"`
	Trends            []GroupTrendPoint             `json:"trends"`
	RecentTransactions []GroupDashboardTransaction  `json:"recent_transactions"`
}

// GroupDashboardSummary represents the summary section of the group dashboard.
type GroupDashboardSummary struct {
	TotalExpenses  string  `json:"total_expenses"`
	TotalIncome    string  `json:"total_income"`
	NetBalance     string  `json:"net_balance"`
	MemberCount    int     `json:"member_count"`
	ExpensesChange float64 `json:"expenses_change"`
	IncomeChange   float64 `json:"income_change"`
}

// GroupCategoryBreakdown represents a category's contribution to the group expenses/income.
type GroupCategoryBreakdown struct {
	CategoryID    string  `json:"category_id"`
	CategoryName  string  `json:"category_name"`
	CategoryColor string  `json:"category_color"`
	Amount        string  `json:"amount"`
	Percentage    float64 `json:"percentage"`
}

// GroupMemberBreakdown represents a member's contribution to the group expenses/income.
type GroupMemberBreakdown struct {
	MemberID         string  `json:"member_id"`
	MemberName       string  `json:"member_name"`
	AvatarURL        string  `json:"avatar_url"`
	Total            string  `json:"total"`
	Percentage       float64 `json:"percentage"`
	TransactionCount int     `json:"transaction_count"`
}

// GroupTrendPoint represents a single data point in the trends chart.
type GroupTrendPoint struct {
	Date     string `json:"date"`
	Income   string `json:"income"`
	Expenses string `json:"expenses"`
}

// GroupDashboardTransaction represents a transaction in the recent transactions list.
type GroupDashboardTransaction struct {
	ID              string `json:"id"`
	Description     string `json:"description"`
	Amount          string `json:"amount"`
	Date            string `json:"date"`
	CategoryName    string `json:"category_name"`
	CategoryColor   string `json:"category_color"`
	MemberName      string `json:"member_name"`
	MemberAvatarURL string `json:"member_avatar_url"`
}

// ToGroupDashboardResponse converts domain GroupDashboardData to a DTO response.
func ToGroupDashboardResponse(data *entity.GroupDashboardData) GroupDashboardResponse {
	response := GroupDashboardResponse{
		CategoryBreakdown:  make([]GroupCategoryBreakdown, 0),
		MemberBreakdown:    make([]GroupMemberBreakdown, 0),
		Trends:             make([]GroupTrendPoint, 0),
		RecentTransactions: make([]GroupDashboardTransaction, 0),
	}

	// Convert summary
	if data.Summary != nil {
		response.Summary = GroupDashboardSummary{
			TotalExpenses:  formatAmount(data.Summary.TotalExpenses),
			TotalIncome:    formatAmount(data.Summary.TotalIncome),
			NetBalance:     formatAmount(data.Summary.NetBalance),
			MemberCount:    data.Summary.MemberCount,
			ExpensesChange: roundToTwoDecimals(data.Summary.ExpensesChange),
			IncomeChange:   roundToTwoDecimals(data.Summary.IncomeChange),
		}
	}

	// Convert category breakdown
	for _, cat := range data.CategoryBreakdown {
		response.CategoryBreakdown = append(response.CategoryBreakdown, GroupCategoryBreakdown{
			CategoryID:    cat.CategoryID.String(),
			CategoryName:  cat.CategoryName,
			CategoryColor: cat.CategoryColor,
			Amount:        formatAmount(cat.Amount),
			Percentage:    roundToTwoDecimals(cat.Percentage),
		})
	}

	// Convert member breakdown
	for _, member := range data.MemberBreakdown {
		response.MemberBreakdown = append(response.MemberBreakdown, GroupMemberBreakdown{
			MemberID:         member.MemberID.String(),
			MemberName:       member.MemberName,
			AvatarURL:        member.AvatarURL,
			Total:            formatAmount(member.Total),
			Percentage:       roundToTwoDecimals(member.Percentage),
			TransactionCount: member.TransactionCount,
		})
	}

	// Convert trends
	for _, trend := range data.Trends {
		response.Trends = append(response.Trends, GroupTrendPoint{
			Date:     trend.Date.Format("2006-01-02"),
			Income:   formatAmount(trend.Income),
			Expenses: formatAmount(trend.Expenses),
		})
	}

	// Convert recent transactions
	for _, txn := range data.RecentTransactions {
		response.RecentTransactions = append(response.RecentTransactions, GroupDashboardTransaction{
			ID:              txn.ID.String(),
			Description:     txn.Description,
			Amount:          formatAmount(txn.Amount),
			Date:            txn.Date.Format("2006-01-02"),
			CategoryName:    txn.CategoryName,
			CategoryColor:   txn.CategoryColor,
			MemberName:      txn.MemberName,
			MemberAvatarURL: txn.MemberAvatarURL,
		})
	}

	return response
}

// formatAmount formats a float64 amount to a string with 2 decimal places.
func formatAmount(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}

// roundToTwoDecimals rounds a float64 to 2 decimal places.
func roundToTwoDecimals(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}
