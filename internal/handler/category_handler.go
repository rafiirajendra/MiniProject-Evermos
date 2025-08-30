package handler

import (
	"go-evermos/config"
	"go-evermos/internal/entities"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// Create Category (Admin only)
func CreateCategory(c *fiber.Ctx) error {
	var input struct {
		NamaCategory string `json:"nama_category"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	category := entities.Category{NamaCategory: input.NamaCategory}
	if err := config.DB.Create(&category).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuat kategori"})
	}

	return c.JSON(category)
}

// Get all Categories (with pagination & filtering)
func GetCategories(c *fiber.Ctx) error {
	var categories []entities.Category
	db := config.DB

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	// Filtering by nama_category
	if nama := c.Query("nama"); nama != "" {
		db = db.Where("nama_category LIKE ?", "%"+nama+"%")
	}

	// Ambil data dengan pagination
	if err := db.Offset(offset).Limit(limit).Find(&categories).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil kategori"})
	}

	return c.JSON(fiber.Map{
		"page":       page,
		"limit":      limit,
		"categories": categories,
	})
}

// Update Category
func UpdateCategory(c *fiber.Ctx) error {
	id := c.Params("id")

	var category entities.Category
	if err := config.DB.First(&category, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Kategori tidak ditemukan"})
	}

	var input struct {
		NamaCategory string `json:"nama_category"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	category.NamaCategory = input.NamaCategory
	if err := config.DB.Save(&category).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal update kategori"})
	}

	return c.JSON(category)
}

// Delete Category
func DeleteCategory(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := config.DB.Delete(&entities.Category{}, id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal hapus kategori"})
	}

	return c.JSON(fiber.Map{"message": "Kategori berhasil dihapus"})
}
