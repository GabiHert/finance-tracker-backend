// Package error defines domain-specific errors for the Finance Tracker application.
package error

import "errors"

// Group domain errors.
var (
	// ErrGroupNotFound is returned when a group is not found in the system.
	ErrGroupNotFound = errors.New("group not found")

	// ErrGroupNameTooLong is returned when the group name exceeds the maximum length.
	ErrGroupNameTooLong = errors.New("group name too long")

	// ErrGroupNameRequired is returned when the group name is empty.
	ErrGroupNameRequired = errors.New("group name is required")

	// ErrMemberNotFound is returned when a member is not found in the group.
	ErrMemberNotFound = errors.New("member not found")

	// ErrInviteNotFound is returned when an invitation is not found.
	ErrInviteNotFound = errors.New("invite not found")

	// ErrInviteExpired is returned when an invitation has expired.
	ErrInviteExpired = errors.New("invite has expired")

	// ErrInviteAlreadyExists is returned when an invitation already exists for the email.
	ErrInviteAlreadyExists = errors.New("invite already exists for this email")

	// ErrUserAlreadyMember is returned when a user is already a member of the group.
	ErrUserAlreadyMember = errors.New("user is already a member of this group")

	// ErrNotGroupAdmin is returned when a non-admin tries to perform admin actions.
	ErrNotGroupAdmin = errors.New("only group admins can perform this action")

	// ErrCannotRemoveSoleAdmin is returned when trying to remove the only admin.
	ErrCannotRemoveSoleAdmin = errors.New("cannot leave or be removed: you are the only admin")

	// ErrInvalidMemberRole is returned when an invalid member role is provided.
	ErrInvalidMemberRole = errors.New("invalid member role")

	// ErrNotGroupMember is returned when a user is not a member of the group.
	ErrNotGroupMember = errors.New("user is not a member of this group")

	// ErrInvalidEmail is returned when an invalid email is provided.
	ErrInvalidGroupEmail = errors.New("invalid email address")

	// ErrCannotInviteSelf is returned when a user tries to invite themselves.
	ErrCannotInviteSelf = errors.New("cannot invite yourself")

	// ErrUserNotRegistered is returned when the invited email is not a registered user.
	ErrUserNotRegistered = errors.New("user is not registered on the platform")
)

// GroupErrorCode defines error codes for group errors.
// Format: GRP-XXYYYY where XX is category and YYYY is specific error.
type GroupErrorCode string

const (
	// Resource not found errors (01XXXX)
	ErrCodeGroupNotFound  GroupErrorCode = "GRP-010001"
	ErrCodeMemberNotFound GroupErrorCode = "GRP-010002"

	// Validation errors (02XXXX)
	ErrCodeGroupNameTooLong   GroupErrorCode = "GRP-020001"
	ErrCodeGroupNameRequired  GroupErrorCode = "GRP-020002"
	ErrCodeInvalidMemberRole  GroupErrorCode = "GRP-020003"
	ErrCodeInvalidGroupEmail  GroupErrorCode = "GRP-020004"
	ErrCodeMissingGroupFields GroupErrorCode = "GRP-020005"

	// Conflict errors (03XXXX)
	ErrCodeInviteAlreadyExists GroupErrorCode = "GRP-010005"
	ErrCodeUserAlreadyMember   GroupErrorCode = "GRP-030002"

	// Authorization errors (04XXXX)
	ErrCodeNotGroupAdmin GroupErrorCode = "GRP-040001"
	ErrCodeNotGroupMember GroupErrorCode = "GRP-040002"

	// Invite errors (05XXXX)
	ErrCodeInviteNotFound    GroupErrorCode = "GRP-010006"
	ErrCodeInviteExpired     GroupErrorCode = "GRP-050002"
	ErrCodeCannotInviteSelf  GroupErrorCode = "GRP-050003"
	ErrCodeUserNotRegistered GroupErrorCode = "GRP-050004"

	// Business logic errors (06XXXX)
	ErrCodeCannotRemoveSoleAdmin GroupErrorCode = "GRP-010008"
)

// GroupError represents a group error with code and message.
type GroupError struct {
	Code    GroupErrorCode
	Message string
	Err     error
}

// Error implements the error interface.
func (e *GroupError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the underlying error.
func (e *GroupError) Unwrap() error {
	return e.Err
}

// NewGroupError creates a new GroupError with the given code and message.
func NewGroupError(code GroupErrorCode, message string, err error) *GroupError {
	return &GroupError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
