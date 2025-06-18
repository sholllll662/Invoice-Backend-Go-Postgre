package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/sholllll662/invoice-backend/database"
	"github.com/sholllll662/invoice-backend/models"
	"gorm.io/gorm"
)

type CreateInvoiceRequest struct {
	ClientID  uint                 `json:"client_id"`
	IssueDate string               `json:"issue_date"`
	DueDate   string               `json:"due_date"`
	Status    string               `json:"status"`
	Note      string               `json:"note"`
	Items     []models.InvoiceItem `json:"items"`
}

func CreateInvoice(c *gin.Context) {
	var req CreateInvoiceRequest

	// Bind JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "detail": err.Error()})
		return
	}

	// Ambil userID dari context
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDInterface.(uint)

	// Hitung total amount dari item-item
	var totalAmount float64
	for i := range req.Items {
		req.Items[i].TotalPrice = float64(req.Items[i].Quantity) * req.Items[i].UnitPrice
		totalAmount += req.Items[i].TotalPrice
	}

	// Buat Invoice
	invoice := models.Invoice{
		UserID:    userID,
		ClientID:  req.ClientID,
		IssueDate: req.IssueDate,
		DueDate:   req.DueDate,
		Amount:    totalAmount,
		Status:    req.Status,
		Note:      req.Note,
		Items:     req.Items,
	}

	if err := database.DB.Create(&invoice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan invoice"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Invoice berhasil dibuat", "invoice": invoice})
}

func GetInvoices(c *gin.Context) {
	// Ambil user ID dari context
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDInterface.(uint)

	// Ambil query parameters
	status := c.Query("status")
	clientID := c.Query("client_id")

	var invoices []models.Invoice
	query := database.DB.Preload("Items").Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if clientID != "" {
		query = query.Where("client_id = ?", clientID)
	}

	if err := query.Order("created_at DESC").Find(&invoices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data invoice"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"invoices": invoices})
}

func GetInvoiceByID(c *gin.Context) {
	// Ambil userID dari context (middleware auth)
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDInterface.(uint)

	// Ambil invoice ID dari parameter URL
	invoiceID := c.Param("id")

	var invoice models.Invoice
	err := database.DB.Preload("Items").
		Where("id = ? AND user_id = ?", invoiceID, userID).
		First(&invoice).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invoice tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data invoice"})
		return
	}

	c.JSON(http.StatusOK, invoice)
}

func UpdateInvoiceByID(c *gin.Context) {
	// Ambil user ID dari context
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDInterface.(uint)

	// Ambil invoice ID dari parameter URL
	invoiceID := c.Param("id")

	var existingInvoice models.Invoice
	err := database.DB.Preload("Items").
		Where("id = ? AND user_id = ?", invoiceID, userID).
		First(&existingInvoice).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invoice tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data invoice"})
		return
	}

	// Ambil data dari request
	var req CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "detail": err.Error()})
		return
	}

	// Hitung ulang total amount dan total harga per item
	var totalAmount float64
	for i := range req.Items {
		req.Items[i].TotalPrice = float64(req.Items[i].Quantity) * req.Items[i].UnitPrice
		totalAmount += req.Items[i].TotalPrice
	}

	// Hapus item lama
	if err := database.DB.Where("invoice_id = ?", invoiceID).Delete(&models.InvoiceItem{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus item lama"})
		return
	}

	// Update invoice utama
	existingInvoice.ClientID = req.ClientID
	existingInvoice.Status = req.Status
	existingInvoice.Note = req.Note
	existingInvoice.IssueDate = req.IssueDate
	existingInvoice.DueDate = req.DueDate
	existingInvoice.Amount = totalAmount
	existingInvoice.Items = req.Items

	if err := database.DB.Save(&existingInvoice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui invoice"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invoice berhasil diperbarui", "invoice": existingInvoice})
}

func DeleteInvoiceByID(c *gin.Context) {
	// Ambil user ID dari context
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDInterface.(uint)

	// Ambil ID invoice dari parameter
	invoiceID := c.Param("id")

	// Cari invoice-nya terlebih dahulu
	var invoice models.Invoice
	err := database.DB.Where("id = ? AND user_id = ?", invoiceID, userID).First(&invoice).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invoice tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mencari invoice"})
		return
	}

	// Hapus semua item terkait
	if err := database.DB.Where("invoice_id = ?", invoiceID).Delete(&models.InvoiceItem{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus item invoice"})
		return
	}

	// Hapus invoice utama
	if err := database.DB.Delete(&invoice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus invoice"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invoice berhasil dihapus"})
}

func formatRupiah(amount float64) string {
	return fmt.Sprintf("Rp %s", humanize.Commaf(amount))
}

func ExportInvoicePDF(c *gin.Context) {
	invoiceID := c.Param("id")

	// Validasi user dan ambil invoice
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDInterface.(uint)

	var invoice models.Invoice
	if err := database.DB.Preload("Items").
		Where("id = ? AND user_id = ?", invoiceID, userID).
		First(&invoice).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invoice tidak ditemukan"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data user"})
		return
	}

	// // Generate PDF
	// pdf := gofpdf.New("P", "mm", "A4", "")
	// pdf.AddPage()
	// pdf.SetFont("Arial", "B", 30)
	// pdf.Cell(40, 10, "INVOICE")

	// // Info invoice
	// pdf.SetFont("Arial", "", 12)
	// pdf.Ln(10)
	// issueDate, _ := time.Parse("2006-01-02", invoice.IssueDate)
	// pdf.Cell(100, 10, "Tanggal: "+issueDate.Format("02 Jan 2006"))
	// pdf.Ln(6)
	// pdf.Cell(100, 10, "Status: "+invoice.Status)
	// pdf.Ln(6)

	// // Tabel item
	// pdf.Ln(10)
	// pdf.SetFont("Arial", "B", 12)
	// pdf.CellFormat(80, 10, "Item", "1", 0, "", false, 0, "")
	// pdf.CellFormat(30, 10, "Qty", "1", 0, "", false, 0, "")
	// pdf.CellFormat(40, 10, "Harga", "1", 0, "", false, 0, "")
	// pdf.CellFormat(40, 10, "Total", "1", 1, "", false, 0, "")

	// pdf.SetFont("Arial", "", 12)
	// for _, item := range invoice.Items {
	// 	pdf.CellFormat(80, 10, item.ItemName, "1", 0, "", false, 0, "")
	// 	pdf.CellFormat(30, 10, strconv.Itoa(item.Quantity), "1", 0, "", false, 0, "")
	// 	pdf.CellFormat(40, 10, formatRupiah(item.UnitPrice), "1", 0, "", false, 0, "")
	// 	pdf.CellFormat(40, 10, formatRupiah(item.TotalPrice), "1", 1, "", false, 0, "")
	// }

	// // Total
	// pdf.Ln(6)
	// pdf.SetFont("Arial", "B", 12)
	// pdf.Cell(150, 10, "TOTAL:")
	// pdf.Cell(40, 10, formatRupiah(invoice.Amount))

	// // Output PDF
	// err := pdf.Output(c.Writer)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat PDF"})
	// 	return
	// }
	// c.Header("Content-Disposition", "inline; filename=invoice.pdf")
	// c.Header("Content-Type", "application/pdf")

	// Buat PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Judul
	pdf.SetFont("Arial", "B", 20)
	pdf.Cell(0, 10, "INVOICE")
	pdf.Ln(12)

	// Informasi KEPADA dan TANGGAL
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(30, 6, "KEPADA :")
	pdf.Cell(90, 6, "")
	pdf.Cell(30, 6, "TANGGAL :")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(60, 6, user.Name) // Ganti sesuai data
	pdf.Cell(60, 6, "")
	issueDate, _ := time.Parse("2006-01-02", invoice.IssueDate)
	pdf.Cell(60, 6, issueDate.Format("Monday, 02 January 2006"))
	pdf.Ln(6)

	pdf.Cell(60, 6, user.Email) // Ganti sesuai data
	pdf.Ln(10)

	// // Nomor invoice
	// pdf.SetFont("Arial", "B", 12)
	// pdf.Cell(40, 6, "NO INVOICE :")
	// pdf.SetFont("Arial", "", 12)
	// pdf.Cell(40, 6, invoice.ID) // Ganti sesuai data
	// pdf.Ln(10)

	// Header tabel item
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(80, 10, "KETERANGAN", "0", 0, "", false, 0, "")
	pdf.CellFormat(40, 10, "HARGA", "0", 0, "C", false, 0, "")
	pdf.CellFormat(30, 10, "JML", "0", 0, "C", false, 0, "")
	pdf.CellFormat(40, 10, "TOTAL", "0", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 12)
	for i, item := range invoice.Items {
		isLast := i == len(invoice.Items)-1

		border := "B"
		fill := true

		if isLast {
			border = "" // Tidak ada garis untuk baris terakhir
			fill = true // Opsional: tidak ada background
		}

		pdf.SetFillColor(230, 230, 230)
		pdf.CellFormat(80, 10, item.ItemName, border, 0, "", fill, 0, "")
		pdf.CellFormat(40, 10, formatRupiah(item.UnitPrice), border, 0, "C", fill, 0, "")
		pdf.CellFormat(30, 10, strconv.Itoa(item.Quantity), border, 0, "C", fill, 0, "")
		pdf.CellFormat(40, 10, formatRupiah(item.TotalPrice), border, 1, "C", fill, 0, "")
	}

	// Subtotal
	pdf.Ln(4)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(80, 10, "", "0 ", 0, "", false, 0, "")
	pdf.CellFormat(40, 10, "", "0", 0, "", false, 0, "")
	pdf.CellFormat(30, 10, "Sub Total", "0", 0, "C", false, 0, "")
	pdf.CellFormat(40, 10, formatRupiah(invoice.Amount), "0", 1, "C", false, 0, "")
	// pdf.Ln(4)
	// pdf.SetFont("Arial", "B", 12)
	// pdf.Cell(120, 10, "")
	// pdf.Cell(30, 10, "SUB TOTAL :")
	// pdf.Cell(40, 10, formatRupiah(invoice.Amount))

	// Terima kasih
	pdf.Ln(20)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 10, "TERIMAKASIH ATAS")
	pdf.Ln(6)
	pdf.Cell(0, 10, "PEMBELIAN ANDA")

	// Output PDF
	err := pdf.Output(c.Writer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat PDF"})
		return
	}
	c.Header("Content-Disposition", "inline; filename=invoice.pdf")
	c.Header("Content-Type", "application/pdf")

}
