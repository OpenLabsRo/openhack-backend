package flagstages

// FlagStageCreateRequest documents the payload to create a new stage.
type FlagStageCreateRequest struct {
	Name    string   `json:"name"`
	TurnOn  []string `json:"turnon"`
	TurnOff []string `json:"turnoff"`
}
