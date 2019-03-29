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
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"
)

func (h *handler) removeUserContent(interaction slackInteraction, duration time.Duration, targetUser string) (removedFiles, remainingFiles, removedMessages, remainingMessages int, err error) {
	if duration > 48*time.Hour {
		return 0, 0, 0, 0, fmt.Errorf("unacceptably long content removal duration: %s", duration)
	}
	start := time.Now().Add(-duration)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		var err error
		removedFiles, remainingFiles, err = h.removeFilesFromUser(targetUser, start)
		if err != nil {
			log.Printf("Couldn't remove files: %v", err)
		}
		wg.Done()
	}()
	go func() {
		var err error
		removedMessages, remainingMessages, err = h.removeMessagesFromUser(targetUser, start)
		if err != nil {
			log.Printf("Couldn't remove messages: %v", err)
		}
		wg.Done()
	}()
	wg.Wait()

	err = nil
	return
}

func (h *handler) removeFilesFromUser(targetUser string, since time.Time) (removed, remaining int, err error) {
	cursor := ""
	for {
		var files []string
		files, cursor, err = h.searchForFiles(targetUser, since, cursor)
		if err != nil {
			return
		}
		for _, v := range files {
			if err := h.removeFile(v); err != nil {
				log.Printf("Failed to remove file %s: %v\n", v, err)
				remaining += 1
			} else {
				removed += 1
			}
		}
		if cursor == "" {
			break
		}
	}
	return removed, remaining, nil
}

func (h *handler) removeFile(id string) error {
	return h.client.CallMethod("files.delete", map[string]string{"file": id}, nil)
}

func (h *handler) searchForFiles(targetUser string, since time.Time, cursor string) ([]string, string, error) {
	args := map[string]string{
		"query":    fmt.Sprintf("from:%s after:%d", targetUser, since.Unix()-1),
		"count":    "100",
		"sort":     "timestamp",
		"sort_dir": "desc",
		"cursor":   cursor,
	}

	result := struct {
		Files struct {
			Matches []struct {
				ID      string `json:"id"`
				Created int64  `json:"created"`
			} `json:"matches"`
		} `json:"files"`
		Metadata struct {
			NextCursor string `json:"next_cursor"`
		} `json:"response_metadata"`
	}{}

	if err := h.client.CallOldMethod("search.files", args, &result); err != nil {
		return nil, "", fmt.Errorf("failed to find files: %v", err)
	}

	files := make([]string, 0, len(result.Files.Matches))
	for _, v := range result.Files.Matches {
		if time.Unix(v.Created, 0).Before(since) {
			continue
		}
		files = append(files, v.ID)
	}
	return files, result.Metadata.NextCursor, nil
}

type messageID struct {
	ts      string
	channel string
}

func (h *handler) removeMessagesFromUser(targetUser string, since time.Time) (removed, remaining int, err error) {
	cursor := ""
	for {
		var messages []messageID
		messages, cursor, err = h.searchForMessages(targetUser, since, cursor)
		if err != nil {
			return
		}
		for _, v := range messages {
			if err := h.removeMessage(v); err != nil {
				log.Printf("Failed to remove message %s: %v\n", v, err)
				remaining += 1
			} else {
				removed += 1
			}
		}
		if cursor == "" {
			break
		}
	}
	return removed, remaining, nil
}

func (h *handler) removeMessage(message messageID) error {
	req := map[string]interface{}{
		"channel": message.channel,
		"ts":      message.ts,
		"as_user": true,
	}
	return h.client.CallMethod("chat.delete", req, nil)
}

func (h *handler) searchForMessages(targetUser string, since time.Time, cursor string) ([]messageID, string, error) {
	args := map[string]string{
		"query":    fmt.Sprintf("from:%s after:%d", targetUser, since.Unix()-1),
		"count":    "100",
		"sort":     "timestamp",
		"sort_dir": "desc",
		"cursor":   cursor,
	}

	result := struct {
		Messages struct {
			Matches []struct {
				Channel struct {
					ID string `json:"id"`
				} `json:"channel"`
				TS string `json:"ts"`
			} `json:"matches"`
		} `json:"messages"`
		Metadata struct {
			NextCursor string `json:"next_cursor"`
		} `json:"response_metadata"`
	}{}

	if err := h.client.CallOldMethod("search.messages", args, &result); err != nil {
		return nil, "", fmt.Errorf("failed to find messages: %v", err)
	}

	messages := make([]messageID, 0, len(result.Messages.Matches))
	for _, v := range result.Messages.Matches {
		t, err := strconv.ParseInt(v.TS, 10, 64)
		if err != nil {
			log.Printf("Failed to parse timestamp %s: %v\n", v, err)
		}
		if time.Unix(t, 0).Before(since) {
			continue
		}
		messages = append(messages, messageID{
			ts:      v.TS,
			channel: v.Channel.ID,
		})
	}
	return messages, result.Metadata.NextCursor, nil
}
