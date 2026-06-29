package middlewares

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

func RequireRoles(allowedRoles ...string) gin.HandlerFunc {
  return func(ctx *gin.Context) {
    userRole, exist := ctx.Get("role")
    if !exist {
      ctx.JSON(http.StatusAccepted, gin.H{
        "status": false,
        "message": "Sesi tidak teridentifikasi",
      })
      ctx.Abort()
      return
    }

    isAllowed := slices.Contains(allowedRoles, userRole.(string))

    if !isAllowed {
      ctx.JSON(http.StatusUnauthorized, gin.H{
        "status": false,
        "message": "Role tidak ditemukan",
      })
      ctx.Abort()
      return
    }

    ctx.Next()
  }
}
