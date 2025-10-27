package badges

import (
	"backend/internal/env"
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/internal/utils"
	"fmt"
	"sort"
	"strconv"
	"time"
)

func computeAndPersistBadgePileSalt(accounts []models.Account, trials int) (uint32, []int, errmsg.StatusError) {
	if env.BADGE_PILES <= 0 {
		return 0, nil, errmsg.InternalServerError(fmt.Errorf("badge piles misconfigured"))
	}

	if trials <= 0 {
		trials = 10_000
	}

	ids := make([]string, 0, len(accounts))
	for _, acc := range accounts {
		if acc.ID != "" {
			ids = append(ids, acc.ID)
		}
	}

	if len(ids) == 0 {
		return 0, nil, errmsg.InternalServerError(fmt.Errorf("no accounts available to compute badge piles"))
	}

	sort.Strings(ids)

	salt, counts := utils.ChooseBestSalt(ids, env.BADGE_PILES, 0, time.Second)

	setting := &models.Setting{
		Name:  models.SettingBadgePileSalt,
		Value: strconv.FormatUint(uint64(salt), 10),
	}

	if serr := setting.Save(); serr != errmsg.EmptyStatusError {
		return 0, nil, serr
	}

	env.BADGE_PILES_SALT = fmt.Sprintf("%s", setting.Value)

	return salt, counts, errmsg.EmptyStatusError
}

func loadBadgePileSalt() (uint32, errmsg.StatusError) {
	setting := &models.Setting{Name: models.SettingBadgePileSalt}

	serr := setting.Get()
	if serr == errmsg.SettingNotFound {
		return utils.BadgePileSalt(), errmsg.EmptyStatusError
	}

	if serr != errmsg.EmptyStatusError {
		return 0, serr
	}

	saltValue, err := strconv.ParseUint(fmt.Sprintf("%s", setting.Value), 10, 32)
	if err != nil {
		return 0, errmsg.InternalServerError(err)
	}

	env.BADGE_PILES_SALT = fmt.Sprintf("%s", setting.Value)

	return uint32(saltValue), errmsg.EmptyStatusError
}
