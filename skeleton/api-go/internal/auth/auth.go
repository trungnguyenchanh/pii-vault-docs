package auth

import (
	"context"
	"net/http"
	"strings"
)

// Principal là danh tính người gọi sau khi xác thực.
type Principal struct {
	Subject string   // định danh người dùng (sub)
	Roles   []string // vai trò: admin | dpo | security ...
}

type ctxKey int

const principalKey ctxKey = 0

// Roles admin hợp lệ cho trang quản trị.
const (
	RoleAdmin    = "admin"
	RoleDPO      = "dpo"
	RoleSecurity = "security"
)

// Authenticator xác minh token và trả về Principal.
// TODO(API-ADM-02): hiện thực xác minh JWT qua JWKS (issuer, audience, exp, chữ ký).
// Skeleton: nếu jwksURL rỗng => "dev mode", chấp nhận header X-Debug-User/X-Debug-Roles.
type Authenticator struct {
	JWKSURL  string
	Issuer   string
	Audience string
}

func NewAuthenticator(jwksURL, issuer, audience string) *Authenticator {
	return &Authenticator{JWKSURL: jwksURL, Issuer: issuer, Audience: audience}
}

// Middleware xác thực mọi request; gắn Principal vào context.
func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p *Principal

		if a.JWKSURL == "" {
			// DEV MODE — KHÔNG dùng ở production.
			user := r.Header.Get("X-Debug-User")
			if user == "" {
				user = "dev-admin"
			}
			roles := r.Header.Get("X-Debug-Roles")
			if roles == "" {
				roles = RoleAdmin
			}
			p = &Principal{Subject: user, Roles: strings.Split(roles, ",")}
		} else {
			// PRODUCTION
			token := bearer(r)
			if token == "" {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}
			// TODO: verifyJWT(token) -> claims -> Principal
			// claims, err := a.verify(token); if err != nil { 401 }
			http.Error(w, "JWT verification not implemented (set JWKS_URL empty for dev mode)", http.StatusNotImplemented)
			return
		}

		ctx := context.WithValue(r.Context(), principalKey, p)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole bọc handler, chỉ cho qua nếu Principal có một trong các vai trò.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p := FromContext(r.Context())
			if p == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if !hasAny(p.Roles, roles) {
				http.Error(w, "forbidden: insufficient role", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// FromContext lấy Principal đã xác thực.
func FromContext(ctx context.Context) *Principal {
	p, _ := ctx.Value(principalKey).(*Principal)
	return p
}

func bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return h[7:]
	}
	return ""
}

func hasAny(have, want []string) bool {
	set := map[string]struct{}{}
	for _, h := range have {
		set[strings.TrimSpace(h)] = struct{}{}
	}
	for _, w := range want {
		if _, ok := set[w]; ok {
			return true
		}
	}
	return false
}
