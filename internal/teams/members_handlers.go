package teams

import (
	"backend/internal/errmsg"
	"backend/internal/events"
	"backend/internal/models"
	"backend/internal/utils"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
)

// TeamMembersGetHandler lists the teammates for the current account.
// @Summary List team members for the authenticated account
// @Description Reads the cached roster for the caller's team to power team management UIs.
// @Tags Teams Members
// @Security AccountAuth
// @Produce json
// @Success 200 {array} models.Account
// @Failure 401 {object} errmsg._AccountNoToken
// @Failure 409 {object} errmsg._AccountHasNoTeam
// @Failure 500 {object} errmsg._InternalServerError
// @Router /teams/members [get]
func TeamMembersGetHandler(c fiber.Ctx) error {
	account := models.Account{}
	utils.GetLocals(c, "account", &account)

	if account.TeamID == "" {
		return utils.StatusError(
			c, errmsg.AccountHasNoTeam,
		)
	}

	team := models.Team{ID: account.TeamID}
	err := team.Get()
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	members, err := team.GetMembers()
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	return c.JSON(members)
}

// TeamMembersJoinHandler lets the caller join a target team via query parameter.
// @Summary Join an existing team by ID
// @Description Validates capacity, attaches the caller, and returns a refreshed token plus updated roster.
// @Tags Teams Members
// @Security AccountAuth
// @Produce json
// @Param teamID query string true "Team ID"
// @Success 200 {object} AccountMembersResponse
// @Failure 401 {object} errmsg._AccountNoToken
// @Failure 404 {object} errmsg._TeamNotFound
// @Failure 409 {object} errmsg._AccountAlreadyHasTeam
// @Failure 409 {object} errmsg._TeamFull
// @Failure 500 {object} errmsg._InternalServerError
// @Router /teams/members/join [patch]
func TeamMembersJoinHandler(c fiber.Ctx) error {
	// get all info on the team
	team := models.Team{ID: c.Query("teamID")}
	err := team.Get()
	if err != nil {
		return utils.StatusError(
			c, errmsg.TeamNotFound,
		)
	}

	// unmarshal the body
	var account models.Account
	utils.GetLocals(c, "account", &account)

	if account.TeamID != "" {
		return utils.StatusError(
			c, errmsg.AccountAlreadyHasTeam,
		)
	}

	// adding the member to the team
	serr := team.AddMember(account.ID, account)
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(
			c, serr,
		)
	}

	err = account.AddToTeam(team.ID)
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	// getting all of the teammembers
	members, err := team.GetMembers()
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	token := account.GenToken()

	events.Em.TeamMemberJoin(
		account.ID,
		team.ID,
	)

	return c.JSON(bson.M{
		"token":   token,
		"account": account,
		"members": members,
	})
}

// TeamLeaveHandler removes the caller from their current team.
// @Summary Leave the current team
// @Description Detaches the caller from the roster and returns the remaining membership plus refreshed token.
// @Tags Teams Members
// @Security AccountAuth
// @Produce json
// @Success 200 {object} AccountMembersResponse
// @Failure 401 {object} errmsg._AccountNoToken
// @Failure 404 {object} errmsg._TeamNotFound
// @Failure 409 {object} errmsg._AccountHasNoTeam
// @Failure 500 {object} errmsg._InternalServerError
// @Router /teams/members/leave [patch]
func TeamMembersLeaveHandler(c fiber.Ctx) error {
	// unmarshal the body
	var account models.Account
	utils.GetLocals(c, "account", &account)

	if account.TeamID == "" {
		return utils.StatusError(
			c, errmsg.AccountHasNoTeam,
		)
	}

	team := models.Team{ID: account.TeamID}
	err := team.Get()
	if err != nil {
		return utils.StatusError(c, errmsg.TeamNotFound)
	}

	// removing the member to the team
	serr := team.RemoveMember(account.ID)
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(
			c, serr,
		)
	}

	err = account.RemoveFromTeam(team.ID)
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	// getting all of the teammembers
	members, err := team.GetMembers()
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	token := account.GenToken()

	events.Em.TeamMemberLeave(
		account.ID,
		team.ID,
	)
	events.Em.AccountTeamExit(
		account.ID,
		team.ID,
	)

	return c.JSON(bson.M{
		"token":   token,
		"account": account,
		"members": members,
	})
}

// TeamMembersKickHandler removes a teammate by account ID.
// @Summary Remove a teammate by account ID
// @Description Allows captains to prune roster members and returns the updated list for confirmation.
// @Tags Teams Members
// @Security AccountAuth
// @Produce json
// @Param accountID query string true "Account ID"
// @Success 200 {object} TeamMembersResponse
// @Failure 401 {object} errmsg._AccountNoToken
// @Failure 404 {object} errmsg._AccountNotFound
// @Failure 404 {object} errmsg._TeamNotFound
// @Failure 500 {object} errmsg._InternalServerError
// @Router /teams/members/kick [patch]
func TeamMembersKickHandler(c fiber.Ctx) error {
	// getting the local account
	var account models.Account
	utils.GetLocals(c, "account", &account)

	// finding the account to remove
	accountToRemove := models.Account{ID: c.Query("accountID")}
	err := accountToRemove.Get()
	if err != nil {
		return utils.StatusError(
			c, errmsg.AccountNotFound,
		)
	}

	// the team
	team := models.Team{ID: account.TeamID}
	err = team.Get()
	if err != nil {
		return utils.StatusError(
			c, errmsg.TeamNotFound,
		)
	}

	// removing the member to the team
	serr := team.RemoveMember(accountToRemove.ID)
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(
			c, serr,
		)
	}

	events.Em.TeamMemberKick(
		account.ID,
		team.ID,
		accountToRemove.ID,
	)

	err = accountToRemove.RemoveFromTeam(team.ID)
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	// getting all of the teammembers
	members, err := team.GetMembers()
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	events.Em.AccountTeamExit(
		accountToRemove.ID,
		team.ID,
	)

	return c.JSON(bson.M{
		"members": members,
	})
}
