package account

import "context"

// Identity is the authenticated request identity established by session
// middleware. Feature handlers must use it instead of client-provided fields.
type Identity struct {
	Nickname  string
	SessionID string
}

type contextKey struct{}

func WithIdentity(ctx context.Context, identity Identity) context.Context {
	return context.WithValue(ctx, contextKey{}, identity)
}

func IdentityFromContext(ctx context.Context) (Identity, bool) {
	identity, ok := ctx.Value(contextKey{}).(Identity)
	return identity, ok
}
