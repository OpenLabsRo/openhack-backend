package teams

import "backend/internal/models"

// TeamRenameRequest captures the payload for renaming a team.
type TeamRenameRequest struct {
	Name string `json:"name"`
}

// SubmissionNameRequest updates the submission name field.
type SubmissionNameRequest struct {
	Name string `json:"name"`
}

// SubmissionDescRequest updates the submission description field.
type SubmissionDescRequest struct {
	Desc string `json:"desc"`
}

// SubmissionRepoRequest updates the submission repository URL.
type SubmissionRepoRequest struct {
	Repo string `json:"repo"`
}

// SubmissionPresRequest updates the submission presentation link.
type SubmissionPresRequest struct {
	Pres string `json:"pres"`
}

// AccountMembersResponse reflects the token, account, and teammate list returned by join/leave operations.
type AccountMembersResponse struct {
	Token   string           `json:"token"`
	Account models.Account   `json:"account"`
	Members []models.Account `json:"members"`
}

// TeamMembersResponse lists the teammates after administrative updates.
type TeamMembersResponse struct {
	Members []models.Account `json:"members"`
}

// AccountTokenResponse mirrors the token/account envelope reused by several team endpoints.
type AccountTokenResponse struct {
	Token   string         `json:"token"`
	Account models.Account `json:"account"`
}
