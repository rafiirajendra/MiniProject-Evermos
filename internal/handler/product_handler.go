package handler

import (
	"fmt"
	"go-evermos/config"
	"go-evermos/internal/entities"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gosimple/slug"
)

func CreateProduct(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	// ambil toko milik user
	var store entities.Store
	if err := config.DB.Where("id_user = ?", userID).First(&store).Error; err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Toko tidak ditemukan"})
	}

	// ambil form data
	namaProduk := c.FormValue("nama_produk")
	hargaReseller := c.FormValue("harga_reseller")
	hargaKonsumen := c.FormValue("harga_konsumen")
	deskripsi := c.FormValue("deskripsi")
	idCategory := c.FormValue("id_category")
	stok := c.FormValue("stok")

	if namaProduk == "" || hargaReseller == "" || hargaKonsumen == "" || idCategory == "" || stok == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Field wajib diisi"})
	}

	// upload file
	file, err := c.FormFile("foto")
	var fotoPath string
	if err == nil {
		fotoPath, err = saveFile(c, file)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal upload foto"})
		}
	}

	// simpan produk
	produk := entities.Product{
		NamaProduk:    namaProduk,
		Slug:          slug.Make(namaProduk),
		HargaReseller: hargaReseller,
		HargaKonsumen: hargaKonsumen,
		Deskripsi:     &deskripsi,
		IDToko:        store.ID,
		IDCategory:    parseUint(idCategory),
		Stok:          parseInt(stok),
	}
	if err := config.DB.Create(&produk).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal simpan produk"})
	}

	// simpan foto
	if fotoPath != "" {
		foto := entities.ProductPicture{
			IDProduk: produk.ID,
			Url:      fotoPath,
		}
		config.DB.Create(&foto)
	}

	return c.JSON(fiber.Map{
		"message": "Produk berhasil dibuat",
		"produk":  produk,
	})
}

// fungsi untuk simpan file
func saveFile(c *fiber.Ctx, file *multipart.FileHeader) (string, error) {
	dir := "./uploads"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, os.ModePerm)
	}
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), strings.ReplaceAll(file.Filename, " ", "_"))
	path := filepath.Join(dir, filename)

	// langsung pakai c.SaveFile
	if err := c.SaveFile(file, path); err != nil {
		return "", err
	}
	return path, nil
}

// helper untuk parse string -> int/uint
func parseUint(s string) uint {
	var u uint
	fmt.Sscan(s, &u)
	return u
}
func parseInt(s string) int {
	var i int
	fmt.Sscan(s, &i)
	return i
}

func GetAllProducts(c *fiber.Ctx) error {
	var products []entities.Product
	db := config.DB.Model(&entities.Product{})

	// Filtering
	if nama := c.Query("nama"); nama != "" {
		db = db.Where("nama_produk LIKE ?", "%"+nama+"%")
	}
	if category := c.Query("category"); category != "" {
		db = db.Where("id_category = ?", category)
	}
	if minPrice := c.Query("min_price"); minPrice != "" {
		db = db.Where("harga_konsumen >= ?", minPrice)
	}
	if maxPrice := c.Query("max_price"); maxPrice != "" {
		db = db.Where("harga_konsumen <= ?", maxPrice)
	}
	if toko := c.Query("toko"); toko != "" {
		db = db.Where("id_toko = ?", toko)
	}

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Hitung total data (setelah filter)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menghitung data"})
	}

	// Ambil data
	if err := db.Preload("ProductPicture").Preload("Category").Preload("Store").
		Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal ambil produk"})
	}

	totalPage := (total + int64(limit) - 1) / int64(limit)

	return c.JSON(fiber.Map{
		"page":        page,
		"limit":       limit,
		"total_data":  total,
		"total_page":  totalPage,
		"products":    products,
	})
}


func GetProductByID(c *fiber.Ctx) error {
	id := c.Params("id")

	var produk entities.Product
	if err := config.DB.Preload("ProductPicture").
		Preload("Category").
		Preload("Store").
		First(&produk, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Produk tidak ditemukan"})
	}

	return c.JSON(produk)
}

func UpdateProduct(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	id := c.Params("id")

	var produk entities.Product
	if err := config.DB.Joins("JOIN Toko ON Toko.id = Produk.id_toko").
		Where("Produk.id = ? AND Toko.id_user = ?", id, userID).
		First(&produk).Error; err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Produk tidak ditemukan atau bukan milik Anda"})
	}

	var input struct {
		NamaProduk    string `json:"nama_produk"`
		HargaReseller string `json:"harga_reseller"`
		HargaKonsumen string `json:"harga_konsumen"`
		Stok          int    `json:"stok"`
		Deskripsi     string `json:"deskripsi"`
		IDCategory    uint   `json:"id_category"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	produk.NamaProduk = input.NamaProduk
	produk.Slug = slug.Make(input.NamaProduk)
	produk.HargaReseller = input.HargaReseller
	produk.HargaKonsumen = input.HargaKonsumen
	produk.Stok = input.Stok
	produk.Deskripsi = &input.Deskripsi
	produk.IDCategory = input.IDCategory

	if err := config.DB.Save(&produk).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal update produk"})
	}

	return c.JSON(fiber.Map{"message": "Produk berhasil diupdate", "produk": produk})
}


func DeleteProduct(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	id := c.Params("id")

	var produk entities.Product
	if err := config.DB.Joins("JOIN Toko ON Toko.id = Produk.id_toko").
		Where("Produk.id = ? AND Toko.id_user = ?", id, userID).
		First(&produk).Error; err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Produk tidak ditemukan atau bukan milik Anda"})
	}

	if err := config.DB.Delete(&produk).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal hapus produk"})
	}

	return c.JSON(fiber.Map{"message": "Produk berhasil dihapus"})
}

