package main

import (
	"encoding/json"
	"fmt"
	"github.com/mattermost/mattermost-server/v5/model"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func writeJsonResponse(w http.ResponseWriter, data model.CommandResponse) {
	w.Header().Set("Content-Type", "application/json")
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Errorf("failed to marshall response: %v", err)
		jsonData = []byte(fmt.Sprintf("error marshalling the response: %v", err))
	}
	_, _ = w.Write(jsonData)
}
