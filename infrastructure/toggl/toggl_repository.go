package toggl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

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

func (r *togglRepository) FindTodayEntries(ctx context.Context) ([]*domain.DeleteTogglEntry, error) {
	now := time.Now().In(domain.JST)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, domain.JST)
	endOfDay := startOfDay.AddDate(0, 0, 1).Add(-time.Second)
	endpoint := fmt.Sprintf("%s/me/time_entries?start_date=%s&end_date=%s",
		baseURL,
		url.QueryEscape(startOfDay.Format(time.RFC3339)),
		url.QueryEscape(endOfDay.Format(time.RFC3339)),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: %w", err)
	}
	req.SetBasicAuth(r.apiToken, "api_token")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http.Do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("toggl API error: status=%d body=%s", resp.StatusCode, string(b))
	}

	var raw []struct {
		ID          int64 `json:"id"`
		WorkspaceID int64 `json:"workspace_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("json.Decode: %w", err)
	}

	entries := make([]*domain.DeleteTogglEntry, len(raw))
	for i, e := range raw {
		entries[i] = &domain.DeleteTogglEntry{ID: e.ID, WorkspaceID: e.WorkspaceID}
	}
	return entries, nil
}

func (r *togglRepository) DeleteEntry(ctx context.Context, id int64) error {
	url := fmt.Sprintf("%s/workspaces/%d/time_entries/%d", baseURL, r.workspaceID, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("http.NewRequest: %w", err)
	}
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

type createTimeEntryRequest struct {
	Description string `json:"description"`
	Start       string `json:"start"`        // RFC3339
	Stop        string `json:"stop"`         // RFC3339
	Duration    int64  `json:"duration"`     // 秒数
	WorkspaceID int64  `json:"workspace_id"`
	ProjectID   int64  `json:"project_id,omitempty"`
	CreatedWith string `json:"created_with"`
}

func (r *togglRepository) CreateTogglEntry(ctx context.Context, entry *domain.TogglEntry) error {
	body := toCreateRequest(entry, r.workspaceID)

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

func toCreateRequest(entry *domain.TogglEntry, workspaceID int64) createTimeEntryRequest {
	return createTimeEntryRequest{
		Description: entry.Description,
		Start:       entry.Start.Format("2006-01-02T15:04:05-07:00"),
		Stop:        entry.End.Format("2006-01-02T15:04:05-07:00"),
		Duration:    int64(entry.End.Sub(entry.Start).Seconds()),
		WorkspaceID: workspaceID,
		ProjectID:   int64(entry.ProjectID),
		CreatedWith: "slack-toggle-syncer",
	}
}
