package pkg

import (

	"github.com/gofiber/fiber/v2"
	"strings"
)

func JWTMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        authHeader := c.Get("Authorization")
        if authHeader == "" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing or invalid token"})
        }

        // Format: Bearer <token>
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        claims, err := ValidateToken(tokenString)
        if err != nil {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired token"})
        }

        // Simpan data user ke context
        c.Locals("user_id", claims.UserID)
        c.Locals("is_admin", claims.IsAdmin)

        return c.Next()
    }
}

func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		isAdmin, ok := c.Locals("is_admin").(bool)
		if !ok || !isAdmin {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Hanya admin yang boleh mengakses resource ini",
			})
		}
		return c.Next()
	}
}