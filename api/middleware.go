package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var HARD_CODED_PUBLIC_KEY = []byte("MY_SECRET_KEY")

var VALIDATION_ALGO = jwt.SigningMethodRS256

// AuthCtx sets the organization ID in the context from the authorization header
func (api *API) jwtValidator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		if len(authHeader) != 2 {
			api.logger.ErrorCtx(ctx, "Malformed Token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		jwtToken := authHeader[1]
		token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
			method, ok := token.Method.(*jwt.SigningMethodRSA)
			if !ok || method != VALIDATION_ALGO {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			_, publicKey, err := api.signingSecretAccessor.ReadSigningSecrets(ctx)
			if err != nil {
				return nil, err
			}
			return publicKey, nil
		})
		if err != nil {
			api.logger.ErrorCtx(ctx, err.Error())
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			ctx := context.WithValue(r.Context(), contextKeyClaims, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			api.logger.ErrorCtx(ctx, err.Error())
			w.WriteHeader(http.StatusUnauthorized)
		}

	})
}

func (api *API) authMiddlewareFactory(middlewareParams map[TokenType]Role) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// first, extract the token claims from the context
			claims, ok := ctx.Value(contextKeyClaims).(jwt.MapClaims)
			if !ok {
				api.logger.ErrorCtx(ctx, "claims not found in context")
				w.WriteHeader(http.StatusForbidden)
				return
			}

			organizationId, ok := claims["organization_id"].(string)
			if !ok {
				api.logger.ErrorCtx(ctx, "organization_id not found in claims")
				w.WriteHeader(http.StatusForbidden)
				return
			}
			tokenType, ok := claims["type"].(string)
			if !ok {
				api.logger.ErrorCtx(ctx, "Token type not found in claims")
				w.WriteHeader(http.StatusForbidden)
				return
			}
			tokenRoleString, ok := claims["role"].(string)
			if !ok {
				api.logger.ErrorCtx(ctx, "Role not found in claims")
				w.WriteHeader(http.StatusForbidden)
				return
			}
			tokenRole := RoleFromString(tokenRoleString)

			// Next, check if the endpoint allows this type of token
			middlewareParamsMinimumRole, ok := middlewareParams[TokenType(tokenType)]
			if !ok {
				api.logger.WarnCtx(ctx, "Token type not allowed for this endpoint")
				w.WriteHeader(http.StatusForbidden)
				return
			}
			if tokenRole < middlewareParamsMinimumRole {
				api.logger.WarnCtx(ctx, "Token role not allowed for this endpoint")
				w.WriteHeader(http.StatusForbidden)
				return
			}

			ctx = context.WithValue(ctx, contextKeyOrgID, organizationId)
			ctx = context.WithValue(ctx, contextKeyTokenType, tokenType)
			ctx = context.WithValue(ctx, contextKeyTokenRole, tokenRole)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

var ErrOrgNotInContext = fmt.Errorf("organization ID not found in request context")

func orgIDFromCtx(ctx context.Context) (id string, err error) {

	orgID, found := ctx.Value(contextKeyOrgID).(string)

	if !found {
		return "", ErrOrgNotInContext
	}

	return orgID, nil
}
