package middlewares

import (
  "net/http"
  "strings"
  
  "KANA-SPACE-BACKEND/internal/pkgs/jwt"

  "github.com/gin-gonic/gin"
  
)

func Authenticate(jwtService jwt.Interface) gin.HandlerFunc {
  return func(ctx *gin.Context) {
    authHeader := ctx.GetHeader("Authorization")
    
    if authHeader == "" {
      ctx.JSON(http.StatusUnauthorized, gin.H{
        "status": false,
        "message": "Token tidak ditemukan",
      })
      ctx.Abort()
      return
    }
    
    tokenParts := strings.Split(authHeader, " ")
    
    if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Format token salah, wajib Bearer Token)",
			})
			ctx.Abort()
			return
		}

		token := tokenParts[1]
		
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Sesi login kadaluarsa",
			})
			ctx.Abort()
			return
		}

		ctx.Set("user_id", claims.UserId)
		ctx.Set("role", claims.Role)
		ctx.Next()
	}
}
