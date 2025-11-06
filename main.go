package main

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

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
	model := c.Param("model")

	var req GenerateContentRequest
	if err := c.Bind(&req); err != nil {
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
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "no text prompt found in contents.parts"})
	}

	// Pass API key to Cursor Agent
	apiKey := c.Request().Header.Get("x-cursor-api-key")

	ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Minute)
	defer cancel()

	agentCmd := os.Getenv("CURSOR_AGENT_CMD")
	if agentCmd == "" {
		agentCmd = "cursor-agent"
	}
	cmd := exec.CommandContext(ctx, agentCmd, "chat", "--model", model, prompt)
	env := os.Environ()
	if apiKey != "" {
		env = append(env, "CURSOR_API_KEY="+apiKey)
	}
	cmd.Env = env

	out, err := cmd.CombinedOutput()
	if err != nil {
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

	return c.JSON(http.StatusOK, resp)
}

func main() {
	e := echo.New()
	e.POST("/v1beta/models/:model", generate)
	// Dùng StartTLS nếu bạn đã cấu hình cert/key; ở đây ví dụ http để tối giản
	e.Logger.Fatal(e.Start(":1994"))
}
