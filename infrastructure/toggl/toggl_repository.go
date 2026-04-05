package toggl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/SkipEveryLunch/slack-toggle-syncer/config"
	"github.com/SkipEveryLunch/slack-toggle-syncer/domain"
)

const baseURL = "https://api.track.toggl.com/api/v9"

type togglRepository struct {
	apiToken    string
	workspaceID int64
}

func NewTogglRepository(cfg config.Toggl) domain.TogglRepository {
	return &togglRepository{
		apiToken:    cfg.APIToken,
		workspaceID: cfg.WorkspaceID,
	}
}

type createTimeEntryRequest struct {
	Description string `json:"description"`
	Start       string `json:"start"`        // RFC3339
	Stop        string `json:"stop"`         // RFC3339
	Duration    int64  `json:"duration"`     // 秒数
	WorkspaceID int64  `json:"workspace_id"`
	ProjectID   int64  `json:"project_id,omitempty"`
	CreatedWith string `json:"created_with"`
}

func (r *togglRepository) CreateTimeEntry(ctx context.Context, entry *domain.TimeEntry) error {
	body := createTimeEntryRequest{
		Description: entry.Description,
		Start:       entry.Start.Format("2006-01-02T15:04:05-07:00"),
		Stop:        entry.End.Format("2006-01-02T15:04:05-07:00"),
		Duration:    int64(entry.End.Sub(entry.Start).Seconds()),
		WorkspaceID: r.workspaceID,
		ProjectID:   entry.ProjectID,
		CreatedWith: "slack-toggle-syncer",
	}

	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	url := fmt.Sprintf("%s/workspaces/%d/time_entries", baseURL, r.workspaceID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("http.NewRequest: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Toggl APIの認証: Basic認証でapi_tokenをユーザー名、"api_token"をパスワードとして使う
	req.SetBasicAuth(r.apiToken, "api_token")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http.Do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("toggl API error: status=%d body=%s", resp.StatusCode, string(b))
	}
	return nil
}
