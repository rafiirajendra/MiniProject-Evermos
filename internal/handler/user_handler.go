package handler

import (
	"go-evermos/config"
	"go-evermos/internal/entities"
	"go-evermos/pkg"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *fiber.Ctx) error {
	var input struct {
		Nama         string `json:"nama"`
		Email        string `json:"email"`
		NoTelp       string `json:"no_telp"`
		KataSandi    string `json:"kata_sandi"`
		TanggalLahir string `json:"tanggal_lahir"`
		JenisKelamin string `json:"jenis_kelamin"`
		Tentang      string `json:"tentang"`
		Pekerjaan    string `json:"pekerjaan"`
		IDProvinsi   string   `json:"id_provinsi"`
		IDKota       string   `json:"id_kota"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Cek no_telp unik
	var existingUser entities.User
	if err := config.DB.Where("notelp = ?", input.NoTelp).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No telepon sudah terdaftar"})
	}

	// Cek email unik
	if err := config.DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email sudah terdaftar"})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.KataSandi), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal hash password"})
	}

	// parse tanggal lahir
	parsedDate, err := time.Parse("2006-01-02", input.TanggalLahir)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format tanggal_lahir harus YYYY-MM-DD"})
	}

	user := entities.User{
		Nama:         input.Nama,
		KataSandi:    string(hashedPassword),
		Notelp:       input.NoTelp,
		TanggalLahir: parsedDate,
		JenisKelamin: input.JenisKelamin,
		Tentang:      &input.Tentang,
		Pekerjaan:    input.Pekerjaan,
		Email:        input.Email,
		IDProvinsi:   input.IDProvinsi,
		IDKota:       input.IDKota,
	}
	if err := config.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	store := entities.Store{
		IDUser:  user.ID,
		NamaToko: &user.Nama,
	}
	config.DB.Create(&store)

	return c.JSON(fiber.Map{"message": "Register sukses"})
}

func Login(c *fiber.Ctx) error {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	var user entities.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Email tidak ditemukan"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.KataSandi), []byte(input.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Password salah"})
	}

	token, err := pkg.GenerateToken(user.ID, user.IsAdmin)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuat token"})
	}

	return c.JSON(fiber.Map{"token": token})
}

func Profile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var user entities.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User tidak ditemukan"})
	}

	return c.JSON(user)
}

func UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var input struct {
		Nama        string `json:"nama"`
		NoTelp      string `json:"no_telp"`
		JenisKelamin string `json:"jenis_kelamin"`
		Tentang     string `json:"tentang"`
		Pekerjaan   string `json:"pekerjaan"`
		IDProvinsi  string `json:"id_provinsi"`
		IDKota      string `json:"id_kota"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	var user entities.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User tidak ditemukan"})
	}

	// Update field
	user.Nama = input.Nama
	user.Notelp = input.NoTelp
	user.JenisKelamin = input.JenisKelamin
	user.Tentang = &input.Tentang
	user.Pekerjaan = input.Pekerjaan
	user.IDProvinsi = input.IDProvinsi
	user.IDKota = input.IDKota

	if err := config.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal update profil"})
	}

	return c.JSON(fiber.Map{
		"message": "Profil berhasil diupdate",
		"user":    user,
	})
}

func GetMyTransactions(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(uint)

    var trxs []entities.Trx
    if err := config.DB.Where("id_user = ?", userID).
        Preload("TrxDetail").
        Preload("TrxDetail.ProductLog").
        Find(&trxs).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal ambil transaksi"})
    }

    return c.JSON(trxs)
}

func GetTransactionByID(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(uint)
    id := c.Params("id")

    var trx entities.Trx
    if err := config.DB.Where("id = ? AND id_user = ?", id, userID).
        Preload("TrxDetail").
        Preload("TrxDetail.ProductLog").
        First(&trx).Error; err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Transaksi tidak ditemukan"})
    }

    return c.JSON(trx)
}
