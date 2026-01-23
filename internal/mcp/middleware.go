package mcp

import (
	"net/http"

	"github.com/CaioMtho/pinoql-mcp/internal/credentials/claims"
	"github.com/gin-gonic/gin"
)

type MCPContextHandler struct {
	baseHandler http.Handler
}

func NewMCPContextHandler(baseHandler http.Handler) *MCPContextHandler {
	return &MCPContextHandler{
		baseHandler: baseHandler,
	}
}

func (h *MCPContextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if ginCtx, ok := ctx.Value(gin.ContextKey).(*gin.Context); ok {
		if claimsValue, exists := ginCtx.Get("claims"); exists {
			if pinoqlClaims, ok := claimsValue.(*claims.PinoQLClaims); ok {
				ctx = WithClaims(ctx, pinoqlClaims)
			}
		}

		if tenantID, exists := ginCtx.Get("tenant_id"); exists {
			if tid, ok := tenantID.(string); ok {
				ctx = WithTenantID(ctx, tid)
			}
		}
	}

	r = r.WithContext(ctx)
	h.baseHandler.ServeHTTP(w, r)
}
