package supabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	baseURL   = "https://fcstpegkakfgjyhmkest.supabase.co"
	apiKey    = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImZjc3RwZWdrYWtmZ2p5aG1rZXN0Iiwicm9sZSI6ImFub24iLCJpYXQiOjE3NDYyNjEwMzgsImV4cCI6MjA2MTgzNzAzOH0.cSJpWvqJSP0Ka46dyoa3_DGHljFMlfyS4J2mI8WgJ34"
	tableName = "tasks"
)

type SupabaseClient struct{}

type Task struct {
	ID     int    `json:"id"`
	chatID int64  `json:"chat_id"`
	Text   string `json:"text"`
	isDone bool   `json:"is_done"`
}

func NewSupabaseClient() *SupabaseClient {
	return &SupabaseClient{}
}

// AddTask adds a new task for the user
func (s *SupabaseClient) AddTask(chatID int64, text string) error {
	task := Task{chatID: chatID, Text: text, isDone: false}
	data, _ := json.Marshal(task)

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/%s", baseURL, tableName), bytes.NewBuffer(data))
	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// ListTasks fetches all tasks for the user
func (s *SupabaseClient) ListTasks(chatID int64) ([]Task, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/%s?user_id=eq.%d", baseURL, tableName, chatID), nil)
	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var tasks []Task
	if err := json.Unmarshal(body, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

// DeleteTask deletes a task by ID and chatID
func (s *SupabaseClient) DeleteTask(chatID int64, taskID int) error {
	url := fmt.Sprintf("%s/%s?id=eq.%d&user_id=eq.%d", baseURL, tableName, taskID, chatID)
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// MarkisDone marks a task as completed
func (s *SupabaseClient) MarkisDone(chatID int64, taskID int) error {
	data := map[string]interface{}{"isDone": true}
	body, _ := json.Marshal(data)

	url := fmt.Sprintf("%s/%s?id=eq.%d&user_id=eq.%d", baseURL, tableName, taskID, chatID)
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(body))
	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
