package main

import (
	"context"
	"cursor-api/log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func init() {
	os.Setenv("APP_NAME", "CURSOR_API")
	logger := log.InitLogger(false)
	// Check if KUBERNETES_SERVICE_HOST is set
	if _, exists := os.LookupEnv("KUBERNETES_SERVICE_HOST"); !exists {
		// If not in Kubernetes, set LOG_LEVEL to DEBUG
		os.Setenv("LOG_LEVEL", "DEBUG")
	}
	logger.SetLevel(log.GetLogLevel("LOG_LEVEL"))
	os.Setenv("TZ", "Asia/Ho_Chi_Minh")
}

type Part struct {
	Text string `json:"text,omitempty"`
}

type Content struct {
	Role  string `json:"role,omitempty"`
	Parts []Part `json:"parts"`
}

type GenerateContentRequest struct {
	Contents []Content `json:"contents"`
}

type Candidate struct {
	Content Content `json:"content"`
}

type GenerateContentResponse struct {
	Model      string      `json:"model"`
	Candidates []Candidate `json:"candidates"`
}

func generate(c echo.Context) error {
	start := time.Now()
	model := c.Param("model")
	log.Infof("Incoming generate request for model=%s", model)
	defer func() {
		log.Debugf("generate handler for model=%s finished in %s", model, time.Since(start))
	}()

	var req GenerateContentRequest
	if err := c.Bind(&req); err != nil {
		log.Errorf("Failed to bind request body for model=%s: %v", model, err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Extract first text prompt
	prompt := ""
	for _, con := range req.Contents {
		for _, p := range con.Parts {
			if p.Text != "" {
				prompt = p.Text
				break
			}
		}
		if prompt != "" {
			break
		}
	}
	if prompt == "" {
		log.Warnf("Request for model=%s did not include any text prompt", model)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "no text prompt found in contents.parts"})
	}
	if len(prompt) > 200 {
		log.Debugf("Prompt preview (first 200 chars) for model=%s: %q...", model, prompt[:200])
	} else {
		log.Debugf("Prompt for model=%s: %q", model, prompt)
	}

	// Pass API key to Cursor Agent
	apiKey := c.Request().Header.Get("x-cursor-api-key")
	if apiKey == "" {
		log.Debugf("No x-cursor-api-key header provided for model=%s; using default agent credentials", model)
	} else {
		log.Debugf("x-cursor-api-key header detected for model=%s", model)
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Minute)
	defer cancel()

	agentCmd := os.Getenv("CURSOR_AGENT_CMD")
	if agentCmd == "" {
		agentCmd = "cursor-agent"
	}
	cmd := exec.CommandContext(ctx, agentCmd, "chat", "--model", model)
	log.Debugf("Executing agent command: %s %v", agentCmd, cmd.Args[1:])
	// Avoid very long argv by sending the prompt via stdin
	cmd.Stdin = strings.NewReader(prompt)
	env := os.Environ()
	if apiKey != "" {
		env = append(env, "CURSOR_API_KEY="+apiKey)
	}
	cmd.Env = env

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Error executing agent command for model=%s: %v", model, err)
		log.Debugf("Agent combined output for model=%s: %s", model, string(out))
		return c.JSON(http.StatusBadGateway, map[string]string{
			"error":  err.Error(),
			"output": string(out),
		})
	}

	reply := strings.TrimSpace(string(out))

	resp := GenerateContentResponse{
		Model: model,
		Candidates: []Candidate{
			{
				Content: Content{
					Parts: []Part{{Text: reply}},
				},
			},
		},
	}

	log.Infof("Successfully generated response for model=%s (payload bytes=%d)", model, len(reply))

	return c.JSON(http.StatusOK, resp)
}

func main() {
	e := echo.New()
	e.Use(log.LoggerHandler)
	e.POST("/v1beta/models/:model", generate)
	// Dùng StartTLS nếu bạn đã cấu hình cert/key; ở đây ví dụ http để tối giản
	e.Logger.Fatal(e.Start(":1994"))
}
