package logs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/soerenschneider/aether/internal"
	"github.com/soerenschneider/aether/internal/templates"
	"go.uber.org/multierr"
)

const (
	defaultQuery = "error AND _time:45m"
	defaultLimit = 20
)

type Opt func(datasource *VictorialogsClient) error

type Querier interface {
	QueryVictorialogs(ctx context.Context, args VictorialogsQuery) ([]LogEntry, error)
}

type VictorialogsClient struct {
	endpoint string
	limit    int
	query    string
	querier  Querier

	regularTemplate *template.Template
	simpleTemplate  *template.Template
	// replaces the default template if set, see WithTemplateFile
	customTemplate     []byte
	httpClient         *http.Client
	excludeFromSummary bool
}

func New(endpoint string, templateData templates.TemplateData, opts ...Opt) (*VictorialogsClient, error) {
	if err := templateData.Validate(); err != nil {
		return nil, fmt.Errorf("invalid template data: %w", err)
	}

	ds := &VictorialogsClient{
		endpoint:   endpoint,
		httpClient: http.DefaultClient,
		query:      defaultQuery,
		limit:      defaultLimit,
	}
	ds.querier = ds

	var errs error
	for _, opt := range opts {
		if err := opt(ds); err != nil {
			errs = multierr.Append(errs, err)
		}
	}
	if errs != nil {
		return nil, errs
	}
	if len(ds.customTemplate) > 0 {
		templateData.DefaultTemplate = ds.customTemplate
	}

	funcMap := template.FuncMap{
		"clock": formatClock,
	}

	var err error
	ds.regularTemplate, err = template.New("logs-regular").Funcs(funcMap).Parse(string(templateData.DefaultTemplate))
	if err != nil {
		return nil, err
	}

	if len(templateData.SimpleTemplate) > 0 {
		ds.simpleTemplate, err = template.New("logs-simple").Funcs(funcMap).Parse(string(templateData.SimpleTemplate))
		if err != nil {
			return nil, err
		}
	}

	return ds, nil
}

func (c *VictorialogsClient) Name() string {
	return "Logs"
}

func (c *VictorialogsClient) GetData(ctx context.Context) (*internal.Data, error) {
	queryReq := VictorialogsQuery{
		Address: c.endpoint,
		Query:   c.query,
		Limit:   c.limit,
	}

	logs, err := c.querier.QueryVictorialogs(ctx, queryReq)
	if err != nil {
		return nil, err
	}

	var regularTemplateData bytes.Buffer
	if err := c.regularTemplate.Execute(&regularTemplateData, logs); err != nil {
		return nil, fmt.Errorf("could not render 'regular' template for datasource %q: %w: %w", c.Name(), internal.ErrTemplate, err)
	}

	var simpleTemplateData bytes.Buffer
	if err := c.simpleTemplate.Execute(&simpleTemplateData, logs); err != nil {
		return nil, fmt.Errorf("could not render 'simple' template for datasource %q: %w: %w", c.Name(), internal.ErrTemplate, err)
	}

	var summary []string
	if !c.excludeFromSummary {
		summary = getSummary(logs)
	}

	tile, tileErr := internal.RenderTile(c.regularTemplate, logs)
	if tileErr != nil {
		return nil, fmt.Errorf("could not render tile for datasource %q: %w", c.Name(), tileErr)
	}

	return &internal.Data{
		Summary:                    summary,
		RenderedDefaultTemplate:    regularTemplateData.Bytes(),
		RenderedSimplifiedTemplate: simpleTemplateData.Bytes(),
		RenderedTile:               tile,
	}, nil
}

func getSummary(logs []LogEntry) []string {
	return nil
}

func (c *VictorialogsClient) QueryVictorialogs(ctx context.Context, args VictorialogsQuery) ([]LogEntry, error) {
	endpoint, err := buildURL(args.Address, "select/logsql/query")
	if err != nil {
		return nil, fmt.Errorf("could not build url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("query", args.Query)
	q.Set("limit", strconv.Itoa(args.GetLimit()))
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad response: %s - %s", resp.Status, body)
	}

	body, _ := io.ReadAll(resp.Body)
	lines := bytes.Split(body, []byte("\n"))
	logs := make([]LogEntry, 0, len(lines))

	for _, line := range lines {
		var entry LogEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}

		t, err := time.Parse(time.RFC3339, entry.Timestamp)
		if err == nil {
			entry.unix = t.Unix()
			entry.Timestamp = t.Format("01-02 15:04:05")
		}

		logs = append(logs, entry)
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].unix < logs[j].unix
	})

	return logs, nil
}
