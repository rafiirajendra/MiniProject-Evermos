package handler

import (
	"go-evermos/config"
	"go-evermos/internal/entities"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// Create address
func CreateAddress(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var input struct {
		JudulAlamat  string `json:"judul_alamat"`
		NamaPenerima string `json:"nama_penerima"`
		NoTelp       string `json:"no_telp"`
		DetailAlamat string `json:"detail_alamat"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	address := entities.Address{
		IDUser:       userID,
		JudulAlamat:  input.JudulAlamat,
		NamaPenerima: input.NamaPenerima,
		NoTelp:       input.NoTelp,
		DetailAlamat: input.DetailAlamat,
	}

	if err := config.DB.Create(&address).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuat alamat"})
	}

	return c.JSON(address)
}

// Get all addresses for user
func GetAddresses(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var addresses []entities.Address
	db := config.DB

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	// Filtering
	if judul := c.Query("judul"); judul != "" {
		db = db.Where("judul_alamat LIKE ?", "%"+judul+"%")
	}
	if penerima := c.Query("penerima"); penerima != "" {
		db = db.Where("nama_penerima LIKE ?", "%"+penerima+"%")
	}
	if telp := c.Query("telp"); telp != "" {
		db = db.Where("no_telp LIKE ?", "%"+telp+"%")
	}

	// Ambil data sesuai user
	if err := db.Where("id_user = ?", userID).
		Offset(offset).
		Limit(limit).
		Find(&addresses).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal ambil alamat"})
	}

	return c.JSON(fiber.Map{
		"page":      page,
		"limit":     limit,
		"addresses": addresses,
	})
}

// Update address
func UpdateAddress(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(uint)
    id := c.Params("id")

    var address entities.Address
    if err := config.DB.Where("id = ? AND id_user = ?", id, userID).First(&address).Error; err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Alamat tidak ditemukan"})
    }

    if err := c.BodyParser(&address); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
    }

    if err := config.DB.Save(&address).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal update alamat"})
    }

    return c.JSON(fiber.Map{"message": "Alamat berhasil diupdate", "alamat": address})
}

// Delete address
func DeleteAddress(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(uint)
    id := c.Params("id")

    if err := config.DB.Where("id = ? AND id_user = ?", id, userID).Delete(&entities.Address{}).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal hapus alamat"})
    }

    return c.JSON(fiber.Map{"message": "Alamat berhasil dihapus"})
}
