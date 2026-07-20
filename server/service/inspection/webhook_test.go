package inspection

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDingTalkPayloadContainsKeywords(t *testing.T) {
	var payload struct {
		MsgType string `json:"msgtype"`
		Text    struct {
			Content string `json:"content"`
		} `json:"text"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer server.Close()

	err := NotifyWebhook(server.URL+"/oapi.dingtalk.com", AlertWebhookPayload{
		TaskID:       7,
		TaskName:     "daily",
		ClusterName:  "prod",
		InspectionID: 9,
		AnomalyCount: 1,
		Summary:      "发现异常",
		Samples:      []string{"node-1 CPU 95%"},
	})
	if err != nil {
		t.Fatalf("NotifyWebhook: %v", err)
	}
	if payload.MsgType != "text" {
		t.Fatalf("msgtype = %q, want text", payload.MsgType)
	}
	for _, keyword := range []string{"日志", "信息", "daily", "node-1"} {
		if !strings.Contains(payload.Text.Content, keyword) {
			t.Errorf("content missing %q: %s", keyword, payload.Text.Content)
		}
	}
}

func TestNotifyWebhookReturnsDingTalkError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":310000,"errmsg":"invalid webhook"}`))
	}))
	defer server.Close()

	err := NotifyWebhook(server.URL+"/oapi.dingtalk.com", AlertWebhookPayload{})
	if err == nil || !strings.Contains(err.Error(), "310000") {
		t.Fatalf("err = %v, want DingTalk errcode", err)
	}
}
