package helper

import "context"

type authCtxKey int

const ctxAuthKey authCtxKey = 1

type AuthInfo struct {
	MemberID int64
	Role     string
	Email    string
	Nama     string
}

func WithAuth(ctx context.Context, a AuthInfo) context.Context {
	return context.WithValue(ctx, ctxAuthKey, a)
}

func GetAuth(ctx context.Context) (AuthInfo, bool) {
	v := ctx.Value(ctxAuthKey)
	if v == nil {
		return AuthInfo{}, false
	}
	a, ok := v.(AuthInfo)
	return a, ok
}
