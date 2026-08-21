package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
)

type contextKey struct{}

var ErrUnauthenticated = errors.New("unauthenticated")

type Principal struct {
	UserID string
	Role   domain.Role
}

func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, contextKey{}, principal)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(contextKey{}).(Principal)
	return principal, ok
}

type DevelopmentIdentity struct {
	Enabled bool
}

func (d DevelopmentIdentity) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !d.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		userID := strings.TrimSpace(r.Header.Get("X-GoreeCloud-User-ID"))
		if userID == "" {
			http.Error(w, ErrUnauthenticated.Error(), http.StatusUnauthorized)
			return
		}
		role := domain.Role(strings.TrimSpace(r.Header.Get("X-GoreeCloud-Role")))
		if role == "" {
			role = domain.RoleUser
		}
		if role != domain.RoleUser && role != domain.RoleAdmin {
			http.Error(w, ErrUnauthenticated.Error(), http.StatusUnauthorized)
			return
		}

		principal := Principal{UserID: userID, Role: role}
		next.ServeHTTP(w, r.WithContext(WithPrincipal(r.Context(), principal)))
	})
}
