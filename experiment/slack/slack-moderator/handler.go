/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"k8s.io/test-infra/experiment/slack/slack"
)

type handler struct {
	client     *slack.Client
	adminToken string
}

func logError(rw http.ResponseWriter, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	log.Println(s)
	http.Error(rw, s, 500)
}

// ServeHTTP handles Slack webhook requests.
func (h *handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logError(rw, "Failed to read incoming request body: %v", err)
		return
	}
	if err := h.client.VerifySignature(body, r.Header); err != nil {
		logError(rw, "Failed validation: %v", err)
		return
	}
	f, err := url.ParseQuery(string(body))
	if err != nil {
		logError(rw, "Failed to parse incoming content: %v", err)
		return
	}
	content := f.Get("payload")
	if content == "" {
		logError(rw, "Payload was blank.")
		return
	}
	interaction := slackInteraction{}
	if err := json.Unmarshal([]byte(content), &interaction); err != nil {
		logError(rw, "Failed to unmarshal payload: %v", err)
		return
	}
	if interaction.Type == "message_action" && interaction.CallbackID == "report_message" {
		if isMod, err := h.userHasModerationPowers(interaction.User.ID); err != nil && isMod {
			h.handleModerateMessage(interaction, rw)
		} else {
			h.handleReportMessage(interaction, rw)
		}
	} else if interaction.Type == "dialog_submission" {
		switch interaction.CallbackID {
		case "send_report":
			h.handleReportSubmission(interaction, rw)
		case "moderate_user":
			h.handleModerateSubmission(interaction, rw)
		}
	}
}

func (h *handler) handleModerateMessage(interaction slackInteraction, rw http.ResponseWriter) {
	deactivateElement := slack.SelectElement{
		Name:  "deactivate",
		Label: fmt.Sprintf("Would you like to deactivate %s?", slack.EscapeMessage(interaction.User.Name)),
		Options: []slack.SelectOption{
			{
				Label: "No",
				Value: "no",
			},
			{
				Label: "Yes",
				Value: "yes",
			},
		},
		Value: "no",
	}
	removeContentElement := slack.SelectElement{
		Name:  "remove_content",
		Label: fmt.Sprintf("How much content would you like to remove?"),
		Options: []slack.SelectOption{
			{
				Label: "None",
				Value: "",
			},
			{
				Label: "1 hour",
				Value: "1h",
			},
			{
				Label: "6 hours",
				Value: "6h",
			},
			{
				Label: "12 hours",
				Value: "12h",
			},
			{
				Label: "24 hours",
				Value: "24h",
			},
			{
				Label: "48 hours",
				Value: "48h",
			},
		},
	}
	dialog := slack.DialogWrapper{
		TriggerID: interaction.TriggerID,
		Dialog: slack.Dialog{
			CallbackID:     "moderate_user",
			NotifyOnCancel: false,
			Title:          "Moderate User",
			Elements:       []interface{}{deactivateElement, removeContentElement},
			State:          interaction.User.ID,
		},
	}
	if err := h.client.CallMethod("dialog.open", dialog, nil); err != nil {
		logError(rw, "Failed to call dialog.open: %v", err)
		return
	}
}

func (h *handler) handleModerateSubmission(interaction slackInteraction, rw http.ResponseWriter) {
	isMod, err := h.userHasModerationPowers(interaction.User.ID)
	if err != nil || !isMod {
		logError(rw, "User %s (%s) does not seem to be a mod: %v", interaction.User.ID, interaction.User.Name, err)
		return
	}
	var messages []string
	targetUser := interaction.State
	if interaction.Submission["deactivate"] == "yes" {
		if err := h.deactivateUser(interaction, targetUser); err != nil {
			messages = append(messages, fmt.Sprintf("Failed to deactivate user %s (%s): %v", interaction.User.ID, interaction.User.Name, err))
		} else {
			messages = append(messages, fmt.Sprintf("Successfully deactivated user %s (%s)", interaction.User.ID, interaction.User.Name))
		}
	}
	if remove, ok := interaction.Submission["remove_content"]; ok && remove != "" {
		duration, err := time.ParseDuration(remove)
		if err != nil {
			messages = append(messages, fmt.Sprintf("Failed to parse removal duration, and therefore could not remove any content: %v", err))
		}
		removedFiles, remainingFiles, removedMessages, remainingMessages, err := h.removeUserContent(interaction, duration, targetUser)

		if err != nil {
			messages = append(messages, fmt.Sprintf("Failed to remove any content: %v", err))
		}

		if remainingFiles == 0 && remainingMessages == 0 {
			messages = append(messages, fmt.Sprintf("Successfully removed %d messages and %d files", removedMessages, removedFiles))
		} else {
			messages = append(messages, fmt.Sprintf("Couldn't remove all content. Removed %d messages and %d files, but there are %d messages and %d files left.", removedMessages, removedFiles, remainingMessages, remainingFiles))
		}
	}

	response := map[string]interface{}{
		"text":             strings.Join(messages, "\n"),
		"response_type":    "ephemeral",
		"replace_original": false,
	}

	if h.client.CallMethod(interaction.ResponseURL, response, nil) != nil {
		logError(rw, "Failed to send response: %v.", err)
		return
	}
}

func (h *handler) deactivateUser(interaction slackInteraction, targetUser string) error {
	result := struct {
		Ok    bool   `json:"ok"`
		Error string `json:"error,omitempty"`
	}{}

	resp, err := http.Post("https://slack.com/api/users.admin.setInactive", "application/x-www-form-urlencoded", bytes.NewBufferString("token="+h.adminToken+"&user="+targetUser))
	if err != nil {
		return fmt.Errorf("couldn't deactivate user %s: %v", targetUser, err)
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode json: %v", err)
	}
	if !result.Ok {
		return fmt.Errorf("couldn't update membership: %s", result.Error)
	}

	return nil
}

func (h *handler) handleReportMessage(interaction slackInteraction, rw http.ResponseWriter) {
	textArea := slack.TextArea{
		Name:  "message",
		Label: "Why are you reporting this message?",
		Hint:  "Moderators will see whatever you write here, along with the message being reported.",
	}
	selectElement := slack.SelectElement{
		Name:  "anonymous",
		Label: "Would you like to report anonymously?",
		Options: []slack.SelectOption{
			{
				Label: "No, report with my username",
				Value: "no",
			},
			{
				Label: "Yes, report anonymously",
				Value: "yes",
			},
		},
		Value: "no",
	}
	var elements []interface{}
	if interaction.Channel.Name == "directmessage" {
		elements = []interface{}{textArea}
	} else {
		elements = []interface{}{textArea, selectElement}
	}
	state, err := json.Marshal(dialogState{
		Sender:  interaction.Message.User,
		TS:      interaction.Message.Timestamp,
		Content: shortenString(interaction.Message.Text, 2900),
	})
	if err != nil {
		logError(rw, "Failed to serialise state for dialog: %v", err)
		return
	}
	dialog := slack.DialogWrapper{
		TriggerID: interaction.TriggerID,
		Dialog: slack.Dialog{
			CallbackID:     "send_report",
			NotifyOnCancel: false,
			Title:          "Report Message",
			Elements:       elements,
			State:          string(state),
		},
	}
	if err := h.client.CallMethod("dialog.open", dialog, nil); err != nil {
		logError(rw, "Failed to call dialog.open: %v", err)
		return
	}
}

func (h *handler) handleReportSubmission(interaction slackInteraction, rw http.ResponseWriter) {
	anonymous := interaction.Submission["anonymous"] == "yes"
	message := interaction.Submission["message"]
	state := dialogState{}
	if err := json.Unmarshal([]byte(interaction.State), &state); err != nil {
		logError(rw, "Failed to parse provided state: %v.", err)
		return
	}

	// Construct summary string
	var who string
	if anonymous {
		who = "An anonymous user"
	} else {
		who = fmt.Sprintf("<@%s|%s>", interaction.User.ID, interaction.User.Name)
	}

	var where string
	if interaction.Channel.Name == "directmessage" {
		where = "a direct message"
	} else {
		where = fmt.Sprintf("<#%s|%s>", interaction.Channel.ID, interaction.Channel.Name)
	}

	summary := fmt.Sprintf("%s *reported a message* in %s:", who, where)

	// Figure out a timestamp from the combined timestamp/message ID
	ts, err := strconv.ParseFloat(state.TS, 64)
	if err != nil {
		logError(rw, "Failed to parse provided timestamp: %v.", err)
		return
	}

	messageLink := "message they reported"
	if interaction.Channel.Name != "directmessage" {
		permalink, err := h.getPermalink(interaction.Channel.ID, state.TS)
		if err != nil {
			log.Printf("Failed to get a permalink: %v.", err)
		} else {
			messageLink = fmt.Sprintf("<%s|message they reported>", permalink)
		}
	}

	var author string
	if senderName, err := h.getDisplayName(state.Sender); err == nil {
		author = fmt.Sprintf("<@%s|%s>", state.Sender, senderName)
	} else {
		author = fmt.Sprintf("<@%s>", state.Sender)
		log.Printf("Failed to look up sender: %v", err)
	}

	report := map[string]interface{}{
		"text": summary,
		"attachments": []map[string]interface{}{
			{
				"pretext":   "They said:",
				"text":      message,
				"mrkdwn_in": []string{"text"},
				"fallback":  "They said: " + message,
			},
			{
				"pretext":     fmt.Sprintf("The %s was:", messageLink),
				"author_name": author,
				"text":        state.Content,
				"ts":          ts,
				"mrkdwn_in":   []string{"text", "pretext", "author_name"},
				"fallback":    fmt.Sprintf("The message they reported was: %s", state.Content),
			},
		},
	}
	if err := h.client.CallMethod(h.client.Config.WebhookURL, report, nil); err != nil {
		logError(rw, "Failed to send report: %v.", err)
		return
	}

	response := map[string]interface{}{
		"text":             "Thank you! Your report has been submitted.",
		"response_type":    "ephemeral",
		"replace_original": false,
	}

	if h.client.CallMethod(interaction.ResponseURL, response, nil) != nil {
		logError(rw, "Failed to send response: %v.", err)
		return
	}
}

func (h *handler) getPermalink(channel string, ts string) (string, error) {
	permalink := struct {
		Channel   string `json:"string"`
		Permalink string `json:"permalink"`
	}{}

	args := map[string]string{
		"channel":    channel,
		"message_ts": ts,
	}

	if err := h.client.CallOldMethod("chat.getPermalink", args, &permalink); err != nil {
		return "", fmt.Errorf("failed get permalink: %v", err)
	}
	return permalink.Permalink, nil
}

func (h *handler) getDisplayName(id string) (string, error) {
	user, err := h.getUserInfo(id)
	if err != nil {
		return "", err
	}
	return user.Name, nil
}

func (h *handler) getUserInfo(id string) (slack.User, error) {
	user := struct {
		User slack.User `json:"user"`
	}{}

	if err := h.client.CallOldMethod("users.info", map[string]string{"user": id}, &user); err != nil {
		return slack.User{}, fmt.Errorf("failed get user: %v", err)
	}
	return user.User, nil
}

func (h *handler) userHasModerationPowers(id string) (bool, error) {
	user, err := h.getUserInfo(id)
	if err != nil {
		return false, err
	}
	return user.IsAdmin || user.IsOwner || user.IsPrimaryOwner, nil
}

// The JSON strings here are short because we can only put a limited amount of information in
// the dialog state.
type dialogState struct {
	Sender  string `json:"s"`
	TS      string `json:"t"`
	Content string `json:"c"`
}

type slackInteraction struct {
	Token       string `json:"token"`
	CallbackID  string `json:"callback_id"`
	Type        string `json:"type"`
	TriggerID   string `json:"trigger_id"`
	ResponseURL string `json:"response_url"`
	Team        struct {
		ID     string `json:"id"`
		Domain string `json:"string"`
	} `json:"team"`
	Channel struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"channel"`
	User struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"user"`
	Message struct {
		Type      string `json:"type"`
		User      string `json:"user"`
		Timestamp string `json:"ts"`
		Text      string `json:"text"`
	}
	Submission map[string]string `json:"submission"`
	State      string            `json:"state"`
}

// shortenString returns the first N slice of a string.
func shortenString(str string, n int) string {
	if len(str) <= n {
		return str
	}
	return str[:n]
}
