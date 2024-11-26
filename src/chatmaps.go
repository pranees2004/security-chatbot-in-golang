package src

import (
	"encoding/json"
	"math/rand"
	"os"
	"sync"
	"time"
)

var userStates = make(map[string]string)
var intents []Intent
var followUpResponses map[string]string
var mu sync.Mutex

func LoadIntents(filePath string) error {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var data IntentData
	err = json.Unmarshal(file, &data)
	if err != nil {
		return err
	}

	intents = data.Intents
	followUpResponses = data.FollowUpResponses
	return nil
}

func SetUserState(userID, state string) {
	mu.Lock()
	defer mu.Unlock()
	userStates[userID] = state
}

func GetUserState(userID string) string {
	mu.Lock()
	defer mu.Unlock()
	return userStates[userID]
}

func GetResponseByState(userID string) string {
	state := GetUserState(userID)
	for _, intent := range intents {
		if intent.Name == state {
			rand.Seed(time.Now().UnixNano())
			randomIndex := rand.Intn(len(intent.Responses))
			followUp := followUpResponses[intent.FollowUpQuestion]
			return intent.Responses[randomIndex] + " " + followUp
		}
	}

	return "I'm not sure how to respond to that."
}
