package service

import "errors"

// 業務一意は DB の UNIQUE ではなく Service で拒否する（.sdd/principles.md）。
var (
	ErrDuplicateEmailInOrg       = errors.New("email already exists in organization")
	ErrDuplicateOrganizationName = errors.New("organization name already exists")
	ErrDuplicateSuperAdminEmail  = errors.New("super admin email already exists")
	ErrDuplicateProjectKey       = errors.New("project key already exists in organization")
	ErrDuplicateRoleName         = errors.New("role name already exists in this scope")
)
