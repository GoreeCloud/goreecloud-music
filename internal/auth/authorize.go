package auth

import "github.com/GoreeCloud/goreecloud-music/internal/domain"

func CanReadLibrary(principal Principal, membership domain.LibraryMembership) bool {
	if principal.Role == domain.RoleAdmin {
		return true
	}
	return principal.UserID != "" && membership.UserID == principal.UserID && membership.CanRead
}

func CanManageLibrary(principal Principal, membership domain.LibraryMembership) bool {
	if principal.Role == domain.RoleAdmin {
		return true
	}
	return principal.UserID != "" && membership.UserID == principal.UserID && membership.CanManage
}
