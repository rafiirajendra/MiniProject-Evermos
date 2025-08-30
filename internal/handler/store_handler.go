package handler

import (
	"go-evermos/config"
	"go-evermos/internal/entities"

	"github.com/gofiber/fiber/v2"
)

// Get toko milik user login
func GetMyStore(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(uint)

    var store entities.Store
    if err := config.DB.Where("id_user = ?", userID).First(&store).Error; err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Toko tidak ditemukan"})
    }

    return c.JSON(store)
}

// Update toko milik user login
func UpdateMyStore(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(uint)

    var store entities.Store
    if err := config.DB.Where("id_user = ?", userID).First(&store).Error; err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Toko tidak ditemukan"})
    }

    var input struct {
        NamaToko string `json:"nama_toko"`
        UrlFoto  string `json:"url_foto"`
    }
    if err := c.BodyParser(&input); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
    }

    if input.NamaToko != "" {
        store.NamaToko = &input.NamaToko
    }
    if input.UrlFoto != "" {
        store.UrlFoto = &input.UrlFoto
    }

    if err := config.DB.Save(&store).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal update toko"})
    }

    return c.JSON(fiber.Map{"message": "Toko berhasil diupdate", "store": store})
}
