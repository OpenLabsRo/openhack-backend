package teams

import "backend/internal/models"

// TeamRenameRequest captures the payload for renaming a team.
type TeamChangeNameRequest struct {
	Name string `json:"name" example:"Team Awesome"`
}

type TeamChangeTableRequest struct {
	Table string `json:"table" example:"A1"`
}

// SubmissionNameRequest updates the submission name field.
type SubmissionNameRequest struct {
	Name string `json:"name" example:"My Project"`
}

// SubmissionDescRequest updates the submission description field.
type SubmissionDescRequest struct {
	Desc string `json:"desc" example:"A brief description of my project"`
}

// SubmissionRepoRequest updates the submission repository URL.
type SubmissionRepoRequest struct {
	Repo string `json:"repo" example:"https://github.com/user/repo"`
}

// SubmissionPresRequest updates the submission presentation link.
type SubmissionPresRequest struct {
	Pres string `json:"pres" example:"https://www.youtube.com/watch?v=dQw4w9WgXcQ"`
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
