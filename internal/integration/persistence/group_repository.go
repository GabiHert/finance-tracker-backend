// Package persistence implements repository interfaces for database operations.
package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	"github.com/finance-tracker/backend/internal/integration/persistence/model"
)

// groupRepository implements the adapter.GroupRepository interface.
type groupRepository struct {
	db *gorm.DB
}

// NewGroupRepository creates a new group repository instance.
func NewGroupRepository(db *gorm.DB) adapter.GroupRepository {
	return &groupRepository{
		db: db,
	}
}

// CreateGroup creates a new group in the database.
func (r *groupRepository) CreateGroup(ctx context.Context, group *entity.Group) error {
	groupModel := model.GroupFromEntity(group)
	result := r.db.WithContext(ctx).Create(groupModel)
	return result.Error
}

// FindGroupByID retrieves a group by its ID.
func (r *groupRepository) FindGroupByID(ctx context.Context, id uuid.UUID) (*entity.Group, error) {
	var groupModel model.GroupModel
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&groupModel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return groupModel.ToEntity(), nil
}

// FindGroupsByUserID retrieves all groups a user belongs to.
func (r *groupRepository) FindGroupsByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.GroupListItem, error) {
	var results []struct {
		ID          uuid.UUID
		Name        string
		MemberCount int
		Role        string
		CreatedAt   string
	}

	query := `
		SELECT
			g.id,
			g.name,
			(SELECT COUNT(*) FROM group_members gm2 WHERE gm2.group_id = g.id) as member_count,
			gm.role,
			g.created_at
		FROM groups g
		INNER JOIN group_members gm ON gm.group_id = g.id
		WHERE gm.user_id = ?
		ORDER BY g.created_at DESC
	`

	if err := r.db.WithContext(ctx).Raw(query, userID).Scan(&results).Error; err != nil {
		return nil, err
	}

	groups := make([]*entity.GroupListItem, len(results))
	for i, res := range results {
		groups[i] = &entity.GroupListItem{
			ID:          res.ID,
			Name:        res.Name,
			MemberCount: res.MemberCount,
			Role:        entity.MemberRole(res.Role),
		}
	}

	return groups, nil
}

// UpdateGroup updates an existing group in the database.
func (r *groupRepository) UpdateGroup(ctx context.Context, group *entity.Group) error {
	groupModel := model.GroupFromEntity(group)
	result := r.db.WithContext(ctx).Save(groupModel)
	return result.Error
}

// DeleteGroup removes a group from the database.
func (r *groupRepository) DeleteGroup(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.GroupModel{}, "id = ?", id)
	return result.Error
}

// CreateMember adds a new member to a group.
func (r *groupRepository) CreateMember(ctx context.Context, member *entity.GroupMember) error {
	memberModel := model.GroupMemberFromEntity(member)
	result := r.db.WithContext(ctx).Create(memberModel)
	return result.Error
}

// FindMemberByID retrieves a group member by their ID.
func (r *groupRepository) FindMemberByID(ctx context.Context, id uuid.UUID) (*entity.GroupMember, error) {
	var memberModel model.GroupMemberModel
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&memberModel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	// Get user info
	var userModel model.UserModel
	if err := r.db.WithContext(ctx).Where("id = ?", memberModel.UserID).First(&userModel).Error; err == nil {
		memberModel.UserName = userModel.Name
		memberModel.UserEmail = userModel.Email
	}

	return memberModel.ToEntity(), nil
}

// FindMemberByGroupAndUser retrieves a member by group and user ID.
func (r *groupRepository) FindMemberByGroupAndUser(ctx context.Context, groupID, userID uuid.UUID) (*entity.GroupMember, error) {
	var memberModel model.GroupMemberModel
	result := r.db.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		First(&memberModel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	// Get user info
	var userModel model.UserModel
	if err := r.db.WithContext(ctx).Where("id = ?", memberModel.UserID).First(&userModel).Error; err == nil {
		memberModel.UserName = userModel.Name
		memberModel.UserEmail = userModel.Email
	}

	return memberModel.ToEntity(), nil
}

// FindMembersByGroupID retrieves all members of a group.
func (r *groupRepository) FindMembersByGroupID(ctx context.Context, groupID uuid.UUID) ([]*entity.GroupMember, error) {
	var memberModels []model.GroupMemberModel
	result := r.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		Order("joined_at ASC").
		Find(&memberModels)
	if result.Error != nil {
		return nil, result.Error
	}

	// Get user info for all members
	userIDs := make([]uuid.UUID, len(memberModels))
	for i, m := range memberModels {
		userIDs[i] = m.UserID
	}

	var userModels []model.UserModel
	if err := r.db.WithContext(ctx).Where("id IN ?", userIDs).Find(&userModels).Error; err == nil {
		userMap := make(map[uuid.UUID]model.UserModel)
		for _, u := range userModels {
			userMap[u.ID] = u
		}
		for i := range memberModels {
			if user, ok := userMap[memberModels[i].UserID]; ok {
				memberModels[i].UserName = user.Name
				memberModels[i].UserEmail = user.Email
			}
		}
	}

	members := make([]*entity.GroupMember, len(memberModels))
	for i, mm := range memberModels {
		members[i] = mm.ToEntity()
	}

	return members, nil
}

// UpdateMember updates a group member.
func (r *groupRepository) UpdateMember(ctx context.Context, member *entity.GroupMember) error {
	memberModel := model.GroupMemberFromEntity(member)
	result := r.db.WithContext(ctx).Save(memberModel)
	return result.Error
}

// DeleteMember removes a member from a group.
func (r *groupRepository) DeleteMember(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.GroupMemberModel{}, "id = ?", id)
	return result.Error
}

// CountAdminsByGroupID counts the number of admins in a group.
func (r *groupRepository) CountAdminsByGroupID(ctx context.Context, groupID uuid.UUID) (int, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&model.GroupMemberModel{}).
		Where("group_id = ? AND role = ?", groupID, entity.MemberRoleAdmin).
		Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return int(count), nil
}

// CreateInvite creates a new group invitation.
func (r *groupRepository) CreateInvite(ctx context.Context, invite *entity.GroupInvite) error {
	inviteModel := model.GroupInviteFromEntity(invite)
	result := r.db.WithContext(ctx).Create(inviteModel)
	return result.Error
}

// FindInviteByToken retrieves an invitation by its token.
func (r *groupRepository) FindInviteByToken(ctx context.Context, token string) (*entity.GroupInvite, error) {
	var inviteModel model.GroupInviteModel
	result := r.db.WithContext(ctx).Where("token = ?", token).First(&inviteModel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return inviteModel.ToEntity(), nil
}

// FindPendingInviteByGroupAndEmail retrieves a pending invite by group and email.
func (r *groupRepository) FindPendingInviteByGroupAndEmail(ctx context.Context, groupID uuid.UUID, email string) (*entity.GroupInvite, error) {
	var inviteModel model.GroupInviteModel
	result := r.db.WithContext(ctx).
		Where("group_id = ? AND email = ? AND status = ?", groupID, email, entity.InviteStatusPending).
		First(&inviteModel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return inviteModel.ToEntity(), nil
}

// FindPendingInvitesByGroupID retrieves all pending invites for a group.
func (r *groupRepository) FindPendingInvitesByGroupID(ctx context.Context, groupID uuid.UUID) ([]*entity.GroupInvite, error) {
	var inviteModels []model.GroupInviteModel
	result := r.db.WithContext(ctx).
		Where("group_id = ? AND status = ?", groupID, entity.InviteStatusPending).
		Order("created_at DESC").
		Find(&inviteModels)
	if result.Error != nil {
		return nil, result.Error
	}

	invites := make([]*entity.GroupInvite, len(inviteModels))
	for i, im := range inviteModels {
		invites[i] = im.ToEntity()
	}

	return invites, nil
}

// UpdateInvite updates an invitation.
func (r *groupRepository) UpdateInvite(ctx context.Context, invite *entity.GroupInvite) error {
	inviteModel := model.GroupInviteFromEntity(invite)
	result := r.db.WithContext(ctx).Save(inviteModel)
	return result.Error
}

// IsUserMemberOfGroup checks if a user is a member of a group.
func (r *groupRepository) IsUserMemberOfGroup(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&model.GroupMemberModel{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// GetGroupWithMembers retrieves a group with its members and pending invites.
func (r *groupRepository) GetGroupWithMembers(ctx context.Context, groupID uuid.UUID) (*entity.GroupWithMembers, error) {
	// Get group
	group, err := r.FindGroupByID(ctx, groupID)
	if err != nil || group == nil {
		return nil, err
	}

	// Get members
	members, err := r.FindMembersByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Get pending invites
	invites, err := r.FindPendingInvitesByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	return &entity.GroupWithMembers{
		Group:          group,
		Members:        members,
		PendingInvites: invites,
		MemberCount:    len(members),
	}, nil
}

// GetGroupDashboard retrieves comprehensive dashboard data for a group.
func (r *groupRepository) GetGroupDashboard(ctx context.Context, groupID uuid.UUID, startDate, endDate time.Time) (*entity.GroupDashboardData, error) {
	// Get members for member breakdown and count
	members, err := r.FindMembersByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Build a map of member info for quick lookup
	memberMap := make(map[uuid.UUID]*entity.GroupMember)
	memberUserIDs := make([]uuid.UUID, len(members))
	for i, m := range members {
		memberMap[m.UserID] = m
		memberUserIDs[i] = m.UserID
	}

	// Get dashboard summary (totals)
	summary, err := r.getGroupDashboardSummary(ctx, groupID, startDate, endDate, len(members))
	if err != nil {
		return nil, err
	}

	// Get category breakdown
	categoryBreakdown, err := r.getGroupCategoryBreakdown(ctx, groupID, startDate, endDate, summary.TotalExpenses)
	if err != nil {
		return nil, err
	}

	// Get member breakdown
	memberBreakdown, err := r.getGroupMemberBreakdown(ctx, groupID, startDate, endDate, memberMap, summary.TotalExpenses)
	if err != nil {
		return nil, err
	}

	// Get trends data
	trends, err := r.getGroupTrends(ctx, groupID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Get recent transactions
	recentTransactions, err := r.getGroupRecentTransactions(ctx, groupID, memberMap, 5)
	if err != nil {
		return nil, err
	}

	return &entity.GroupDashboardData{
		Summary:            summary,
		CategoryBreakdown:  categoryBreakdown,
		MemberBreakdown:    memberBreakdown,
		Trends:             trends,
		RecentTransactions: recentTransactions,
	}, nil
}

// getGroupDashboardSummary gets the summary totals for the dashboard.
func (r *groupRepository) getGroupDashboardSummary(ctx context.Context, groupID uuid.UUID, startDate, endDate time.Time, memberCount int) (*entity.GroupDashboardSummary, error) {
	var result struct {
		TotalExpenses float64
		TotalIncome   float64
	}

	// Query for transactions from all group members within the period
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN t.type = 'expense' THEN ABS(t.amount) ELSE 0 END), 0) as total_expenses,
			COALESCE(SUM(CASE WHEN t.type = 'income' THEN ABS(t.amount) ELSE 0 END), 0) as total_income
		FROM transactions t
		WHERE t.user_id IN (SELECT user_id FROM group_members WHERE group_id = ?)
		  AND t.date >= ?
		  AND t.date <= ?
		  AND t.deleted_at IS NULL
	`

	if err := r.db.WithContext(ctx).Raw(query, groupID, startDate, endDate).Scan(&result).Error; err != nil {
		return nil, err
	}

	return &entity.GroupDashboardSummary{
		TotalExpenses:  result.TotalExpenses,
		TotalIncome:    result.TotalIncome,
		NetBalance:     result.TotalIncome - result.TotalExpenses,
		MemberCount:    memberCount,
		ExpensesChange: 0, // Will be calculated by the use case
		IncomeChange:   0, // Will be calculated by the use case
	}, nil
}

// getGroupCategoryBreakdown gets the category breakdown for expenses.
func (r *groupRepository) getGroupCategoryBreakdown(ctx context.Context, groupID uuid.UUID, startDate, endDate time.Time, totalExpenses float64) ([]*entity.GroupCategoryBreakdown, error) {
	var results []struct {
		CategoryID    uuid.UUID
		CategoryName  string
		CategoryColor string
		Amount        float64
	}

	query := `
		SELECT
			c.id as category_id,
			COALESCE(c.name, 'Sem categoria') as category_name,
			COALESCE(c.color, '#9CA3AF') as category_color,
			COALESCE(SUM(ABS(t.amount)), 0) as amount
		FROM transactions t
		LEFT JOIN categories c ON c.id = t.category_id
		WHERE t.user_id IN (SELECT user_id FROM group_members WHERE group_id = ?)
		  AND t.type = 'expense'
		  AND t.date >= ?
		  AND t.date <= ?
		  AND t.deleted_at IS NULL
		GROUP BY c.id, c.name, c.color
		ORDER BY amount DESC
	`

	if err := r.db.WithContext(ctx).Raw(query, groupID, startDate, endDate).Scan(&results).Error; err != nil {
		return nil, err
	}

	breakdown := make([]*entity.GroupCategoryBreakdown, len(results))
	for i, r := range results {
		percentage := 0.0
		if totalExpenses > 0 {
			percentage = (r.Amount / totalExpenses) * 100
		}
		breakdown[i] = &entity.GroupCategoryBreakdown{
			CategoryID:    r.CategoryID,
			CategoryName:  r.CategoryName,
			CategoryColor: r.CategoryColor,
			Amount:        r.Amount,
			Percentage:    percentage,
		}
	}

	return breakdown, nil
}

// getGroupMemberBreakdown gets the member contribution breakdown for expenses.
func (r *groupRepository) getGroupMemberBreakdown(ctx context.Context, groupID uuid.UUID, startDate, endDate time.Time, memberMap map[uuid.UUID]*entity.GroupMember, totalExpenses float64) ([]*entity.GroupMemberBreakdown, error) {
	var results []struct {
		UserID           uuid.UUID
		Total            float64
		TransactionCount int
	}

	query := `
		SELECT
			t.user_id,
			COALESCE(SUM(ABS(t.amount)), 0) as total,
			COUNT(*) as transaction_count
		FROM transactions t
		WHERE t.user_id IN (SELECT user_id FROM group_members WHERE group_id = ?)
		  AND t.type = 'expense'
		  AND t.date >= ?
		  AND t.date <= ?
		  AND t.deleted_at IS NULL
		GROUP BY t.user_id
		ORDER BY total DESC
	`

	if err := r.db.WithContext(ctx).Raw(query, groupID, startDate, endDate).Scan(&results).Error; err != nil {
		return nil, err
	}

	breakdown := make([]*entity.GroupMemberBreakdown, len(results))
	for i, r := range results {
		percentage := 0.0
		if totalExpenses > 0 {
			percentage = (r.Total / totalExpenses) * 100
		}

		memberName := "Unknown"
		memberID := r.UserID
		if member, ok := memberMap[r.UserID]; ok {
			memberName = member.UserName
			memberID = member.ID
		}

		breakdown[i] = &entity.GroupMemberBreakdown{
			MemberID:         memberID,
			MemberName:       memberName,
			AvatarURL:        "", // No avatar URLs in the current schema
			Total:            r.Total,
			Percentage:       percentage,
			TransactionCount: r.TransactionCount,
		}
	}

	return breakdown, nil
}

// getGroupTrends gets daily income/expense trends for the period.
func (r *groupRepository) getGroupTrends(ctx context.Context, groupID uuid.UUID, startDate, endDate time.Time) ([]*entity.GroupTrendPoint, error) {
	var results []struct {
		Date     time.Time
		Income   float64
		Expenses float64
	}

	query := `
		SELECT
			t.date,
			COALESCE(SUM(CASE WHEN t.type = 'income' THEN ABS(t.amount) ELSE 0 END), 0) as income,
			COALESCE(SUM(CASE WHEN t.type = 'expense' THEN ABS(t.amount) ELSE 0 END), 0) as expenses
		FROM transactions t
		WHERE t.user_id IN (SELECT user_id FROM group_members WHERE group_id = ?)
		  AND t.date >= ?
		  AND t.date <= ?
		  AND t.deleted_at IS NULL
		GROUP BY t.date
		ORDER BY t.date ASC
	`

	if err := r.db.WithContext(ctx).Raw(query, groupID, startDate, endDate).Scan(&results).Error; err != nil {
		return nil, err
	}

	trends := make([]*entity.GroupTrendPoint, len(results))
	for i, r := range results {
		trends[i] = &entity.GroupTrendPoint{
			Date:     r.Date,
			Income:   r.Income,
			Expenses: r.Expenses,
		}
	}

	return trends, nil
}

// getGroupRecentTransactions gets the most recent transactions for the group.
func (r *groupRepository) getGroupRecentTransactions(ctx context.Context, groupID uuid.UUID, memberMap map[uuid.UUID]*entity.GroupMember, limit int) ([]*entity.GroupDashboardTransaction, error) {
	var results []struct {
		ID            uuid.UUID
		Description   string
		Amount        float64
		Date          time.Time
		Type          string
		CategoryName  string
		CategoryColor string
		UserID        uuid.UUID
	}

	query := `
		SELECT
			t.id,
			t.description,
			t.amount,
			t.date,
			t.type,
			COALESCE(c.name, 'Sem categoria') as category_name,
			COALESCE(c.color, '#9CA3AF') as category_color,
			t.user_id
		FROM transactions t
		LEFT JOIN categories c ON c.id = t.category_id
		WHERE t.user_id IN (SELECT user_id FROM group_members WHERE group_id = ?)
		  AND t.deleted_at IS NULL
		ORDER BY t.date DESC, t.created_at DESC
		LIMIT ?
	`

	if err := r.db.WithContext(ctx).Raw(query, groupID, limit).Scan(&results).Error; err != nil {
		return nil, err
	}

	transactions := make([]*entity.GroupDashboardTransaction, len(results))
	for i, r := range results {
		memberName := "Unknown"
		memberID := r.UserID
		if member, ok := memberMap[r.UserID]; ok {
			memberName = member.UserName
			memberID = member.ID
		}

		// Format amount: negative for expenses, positive for income
		amount := r.Amount
		if r.Type == "expense" {
			amount = -amount
		}

		transactions[i] = &entity.GroupDashboardTransaction{
			ID:              r.ID,
			Description:     r.Description,
			Amount:          amount,
			Date:            r.Date,
			CategoryName:    r.CategoryName,
			CategoryColor:   r.CategoryColor,
			MemberID:        memberID,
			MemberName:      memberName,
			MemberAvatarURL: "", // No avatar URLs in the current schema
		}
	}

	return transactions, nil
}

// GetGroupDashboardPreviousPeriod retrieves dashboard totals for comparison period.
func (r *groupRepository) GetGroupDashboardPreviousPeriod(ctx context.Context, groupID uuid.UUID, startDate, endDate time.Time) (totalExpenses, totalIncome float64, err error) {
	var result struct {
		TotalExpenses float64
		TotalIncome   float64
	}

	query := `
		SELECT
			COALESCE(SUM(CASE WHEN t.type = 'expense' THEN ABS(t.amount) ELSE 0 END), 0) as total_expenses,
			COALESCE(SUM(CASE WHEN t.type = 'income' THEN ABS(t.amount) ELSE 0 END), 0) as total_income
		FROM transactions t
		WHERE t.user_id IN (SELECT user_id FROM group_members WHERE group_id = ?)
		  AND t.date >= ?
		  AND t.date <= ?
		  AND t.deleted_at IS NULL
	`

	if err := r.db.WithContext(ctx).Raw(query, groupID, startDate, endDate).Scan(&result).Error; err != nil {
		return 0, 0, err
	}

	return result.TotalExpenses, result.TotalIncome, nil
}
