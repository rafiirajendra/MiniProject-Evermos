package handler

import (
	"fmt"
	"go-evermos/config"
	"go-evermos/internal/entities"
	"time"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// Request body untuk checkout
type CheckoutRequest struct {
	IDAlamat    uint `json:"id_alamat"`
	MethodBayar string `json:"method_bayar"`
	Items       []struct {
		IDProduk uint `json:"id_produk"`
		Qty      int  `json:"qty"`
	} `json:"items"`
}

// Create Transaction (Checkout)
func CreateTransaction(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	// Parse body
	var req CheckoutRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	if len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Item tidak boleh kosong"})
	}

	var totalHarga int
	var trxDetails []entities.TrxDetail

	// Proses tiap produk
	for _, item := range req.Items {
		var produk entities.Product
		if err := config.DB.First(&produk, item.IDProduk).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": fmt.Sprintf("Produk %d tidak ditemukan", item.IDProduk)})
		}

		// Kurangi stok
		if produk.Stok < item.Qty {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("Stok produk %s tidak mencukupi", produk.NamaProduk)})
		}
		produk.Stok -= item.Qty
		config.DB.Save(&produk)

		// Simpan ke ProductLog
		prodLog := entities.ProductLog{
			IDProduk:      produk.ID,
			NamaProduk:    produk.NamaProduk,
			Slug:          produk.Slug,
			HargaReseller: produk.HargaReseller,
			HargaKonsumen: produk.HargaKonsumen,
			Deskripsi:     produk.Deskripsi,
			IDToko:        produk.IDToko,
			IDCategory:    produk.IDCategory,
		}
		config.DB.Create(&prodLog)

		// Hitung harga total per item
		// asumsi HargaKonsumen = string → kita pakai parseInt
		var hargaInt int
		fmt.Sscan(produk.HargaKonsumen, &hargaInt)
		hargaTotalItem := hargaInt * item.Qty
		totalHarga += hargaTotalItem

		// Buat detail
		trxDetails = append(trxDetails, entities.TrxDetail{
			IDLogProduk: prodLog.ID,
			IDToko:      produk.IDToko,
			Kuantitas:   item.Qty,
			HargaTotal:  hargaTotalItem,
		})
	}

	// Buat transaksi utama
	trx := entities.Trx{
		IDUser:           userID,
		AlamatPengiriman: req.IDAlamat,
		HargaTotal:       totalHarga,
		KodeInvoice:      fmt.Sprintf("INV-%d", time.Now().Unix()),
		MethodBayar:      req.MethodBayar,
	}
	if err := config.DB.Create(&trx).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal simpan transaksi"})
	}

	// Simpan detail transaksi
	for i := range trxDetails {
		trxDetails[i].IDTrx = trx.ID
		config.DB.Create(&trxDetails[i])
	}

	return c.JSON(fiber.Map{
		"message": "Transaksi berhasil dibuat",
		"trx":     trx,
		"details": trxDetails,
	})
}

// Ambil semua transaksi milik user (dengan pagination & filter)
func GetUserTransactions(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	db := config.DB

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	// Filtering
	if method := c.Query("method"); method != "" {
		db = db.Where("method_bayar = ?", method)
	}
	if invoice := c.Query("invoice"); invoice != "" {
		db = db.Where("kode_invoice LIKE ?", "%"+invoice+"%")
	}

	var trxs []entities.Trx
	if err := db.Preload("TrxDetail").
		Where("id_user = ?", userID).
		Offset(offset).Limit(limit).
		Find(&trxs).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal ambil transaksi"})
	}

	return c.JSON(fiber.Map{
		"page":         page,
		"limit":        limit,
		"transactions": trxs,
	})
}

// Ambil detail transaksi tertentu
func GetUserTransactionByID(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	id := c.Params("id")

	var trx entities.Trx
	if err := config.DB.Preload("TrxDetail").
		Where("id = ? AND id_user = ?", id, userID).
		First(&trx).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Transaksi tidak ditemukan"})
	}

	return c.JSON(trx)
}
