package accounts

import (
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/internal/utils"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

// GetVoucherHandler serves voucher images from the file system.
// @Summary Get voucher image
// @Description Returns a voucher image for the authenticated participant if they have the corresponding promotional code.
// @Tags Accounts Vouchers
// @Produce image/jpeg
// @Param token query string true "Authentication token"
// @Param voucherType path string true "Voucher type (e.g., vmax)"
// @Param index path int true "Voucher index"
// @Success 200 {file} binary
// @Failure 400 {object} errmsg._VoucherInvalidIndex
// @Failure 401 {object} errmsg._AccountNoToken
// @Failure 403 {object} errmsg._VoucherNoPromoCode
// @Failure 404 {object} errmsg._VoucherNotFound
// @Failure 500 {object} errmsg._InternalServerError
// @Router /accounts/vouchers/{voucherType}/{index} [get]
func GetVoucherHandler(c fiber.Ctx) error {
	// Extract token from query parameter
	token := c.Query("token")
	if token == "" {
		return utils.StatusError(c, errmsg.AccountNoToken)
	}

	// Validate and retrieve account from token
	account := models.Account{}
	err := account.ParseToken(token)
	if err != nil {
		return utils.StatusError(c, errmsg.AccountNoToken)
	}

	if account.ID == "" {
		return utils.StatusError(c, errmsg.AccountNoToken)
	}

	voucherType := c.Params("voucherType")
	indexStr := c.Params("index")

	// Validate index parameter
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 {
		return utils.StatusError(c, errmsg.VoucherInvalidIndex)
	}

	// Check if the account has the promotional code for this voucher type
	err = account.Get()
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}
	promo, exists := account.Promotionals[voucherType]
	if !exists || promo == "" {
		return utils.StatusError(c, errmsg.VoucherNoPromoCode)
	}

	// Build the file path using fixed /var/openhack directory
	voucherDir := filepath.Join("/var/openhack", fmt.Sprintf("%s-vouchers", voucherType))
	filePath := filepath.Join(voucherDir, fmt.Sprintf("%d.jpg", index))

	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return utils.StatusError(c, errmsg.VoucherNotFound)
		}
		return utils.StatusError(c, errmsg.InternalServerError(err))
	}

	// Send the file with proper content type
	return c.SendFile(filePath)
}
