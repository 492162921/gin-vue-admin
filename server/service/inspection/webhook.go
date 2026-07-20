package inspection

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type AlertWebhookPayload struct {
	TaskID       uint
	TaskName     string
	ClusterName  string
	InspectionID uint
	AnomalyCount int
	Summary      string
	Samples      []string
}

func NotifyWebhook(webhookURL string, payload AlertWebhookPayload) error {
	webhookURL = strings.TrimSpace(webhookURL)
	if webhookURL == "" {
		return nil
	}

	content := formatAlertWebhookContent(payload)
	var body any
	if strings.Contains(webhookURL, "oapi.dingtalk.com") {
		body = map[string]any{
			"msgtype": "text",
			"text":    map[string]string{"content": content},
		}
	} else {
		body = map[string]any{
			"event":        "inspection_anomaly",
			"taskId":       payload.TaskID,
			"taskName":     payload.TaskName,
			"clusterName":  payload.ClusterName,
			"inspectionId": payload.InspectionID,
			"anomalyCount": payload.AnomalyCount,
			"summary":      payload.Summary,
			"samples":      payload.Samples,
			"message":      content,
		}
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	client := &http.Client{Timeout: 8 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	responseBody, _ := io.ReadAll(io.LimitReader(response.Body, 2048))
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("webhook http %d: %s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	if strings.Contains(webhookURL, "oapi.dingtalk.com") {
		var result struct {
			ErrCode int    `json:"errcode"`
			ErrMsg  string `json:"errmsg"`
		}
		if json.Unmarshal(responseBody, &result) == nil && result.ErrCode != 0 {
			return fmt.Errorf("dingtalk errcode=%d errmsg=%s", result.ErrCode, result.ErrMsg)
		}
	}
	return nil
}

func NotifyReportWebhook(webhookURL, title, digest string) error {
	webhookURL = strings.TrimSpace(webhookURL)
	if webhookURL == "" {
		return nil
	}
	content := fmt.Sprintf("【日志】【信息】K8s 巡检报告\n标题: %s\n摘要: %s", displayValue(title), displayValue(digest))
	var body any
	if strings.Contains(webhookURL, "oapi.dingtalk.com") {
		body = map[string]any{
			"msgtype": "text",
			"text":    map[string]string{"content": content},
		}
	} else {
		body = map[string]any{
			"event":   "inspection_report",
			"title":   title,
			"digest":  digest,
			"message": content,
		}
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	request, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	response, err := (&http.Client{Timeout: 8 * time.Second}).Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	responseBody, _ := io.ReadAll(io.LimitReader(response.Body, 2048))
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("webhook http %d: %s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	return nil
}

func formatAlertWebhookContent(payload AlertWebhookPayload) string {
	var content strings.Builder
	content.WriteString("【日志】【信息】K8s 巡检告警\n")
	content.WriteString(fmt.Sprintf("任务: %s (id=%d)\n", displayValue(payload.TaskName), payload.TaskID))
	content.WriteString(fmt.Sprintf("集群: %s\n", displayValue(payload.ClusterName)))
	content.WriteString(fmt.Sprintf("巡检ID: %d\n异常数: %d\n", payload.InspectionID, payload.AnomalyCount))
	if payload.Summary != "" {
		content.WriteString("摘要: " + payload.Summary + "\n")
	}
	for _, sample := range payload.Samples {
		content.WriteString("- " + sample + "\n")
	}
	return content.String()
}

func displayValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}
