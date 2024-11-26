package src

type Intent struct {
	ID               int      `json:"id"`
	Name             string   `json:"name"`
	Examples         []string `json:"examples"`
	Responses        []string `json:"responses"`
	FollowUpQuestion string   `json:"follow_up_question"`
}

type IntentData struct {
	Intents           []Intent          `json:"intents"`
	FollowUpResponses map[string]string `json:"follow_up_responses"`
}