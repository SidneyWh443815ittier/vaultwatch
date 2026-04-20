package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yourusername/vaultwatch/internal/alert"
)

const defaultJiraAPIPath = "/rest/api/2/issue"

type jiraSender struct {
	baseURL  string
	project  string
	issueType string
	username string
	token    string
	client   *http.Client
}

type jiraPayload struct {
	Fields jiraFields `json:"fields"`
}

type jiraFields struct {
	Project   jiraProject   `json:"project"`
	Summary   string        `json:"summary"`
	Description string      `json:"description"`
	Issuetype jiraIssueType `json:"issuetype"`
	Priority  jiraPriority  `json:"priority"`
}

type jiraProject struct {
	Key string `json:"key"`
}

type jiraIssueType struct {
	Name string `json:"name"`
}

type jiraPriority struct {
	Name string `json:"name"`
}

// NewJiraSender creates a Jira sender that opens issues for lease alerts.
func NewJiraSender(baseURL, project, issueType, username, token string) alert.Sender {
	return newJiraSenderWithURL(baseURL+defaultJiraAPIPath, project, issueType, username, token)
}

func newJiraSenderWithURL(url, project, issueType, username, token string) alert.Sender {
	return &jiraSender{
		baseURL:   url,
		project:   project,
		issueType: issueType,
		username:  username,
		token:     token,
		client:    &http.Client{},
	}
}

func (j *jiraSender) Send(a alert.Alert) error {
	priority := "Medium"
	if a.Level == alert.LevelCritical {
		priority = "High"
	}

	payload := jiraPayload{
		Fields: jiraFields{
			Project:   jiraProject{Key: j.project},
			Summary:   fmt.Sprintf("[VaultWatch] Lease expiring: %s", a.LeaseID),
			Description: fmt.Sprintf("Lease *%s* expires in %s.\nStatus: %s", a.LeaseID, a.TTL, a.Level),
			Issuetype: jiraIssueType{Name: j.issueType},
			Priority:  jiraPriority{Name: priority},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("jira: marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, j.baseURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("jira: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(j.username, j.token)

	resp, err := j.client.Do(req)
	if err != nil {
		return fmt.Errorf("jira: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("jira: unexpected status %d", resp.StatusCode)
	}
	return nil
}
