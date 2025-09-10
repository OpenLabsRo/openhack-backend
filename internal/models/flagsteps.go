package models

type FlagSteps struct {
	Name    string   `json:"name" bson:"name"`
	TurnOff []string `json:"turnoff" bson:"turnoff"`
	TurnOn  []string `json:"turnon" bson:"turnon"`
}
