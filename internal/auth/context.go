package auth

import (
	"context"
	"net/http"
)

type contextKey string

const UserContextKey contextKey = "user"

type UserContext struct {
	UserID   int64
	Username string
	Email    string
	Role     string
}

func GetUserFromContext(r *http.Request) *UserContext {
	user, ok := r.Context().Value(UserContextKey).(*UserContext)
	if !ok {
		return nil
	}
	return user
}

func SetUserInContext(ctx context.Context, user *UserContext) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}
