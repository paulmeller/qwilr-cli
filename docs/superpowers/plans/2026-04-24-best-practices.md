# Best Practices Improvements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Apply Go CLI best practices from steipete's openclaw tools: SilenceUsage/SilenceErrors, remove os.Exit from helpers, add context+timeout to API client, add --timeout flag, NO_COLOR support, structured exit codes, input validation, and --dry-run.

**Architecture:** Refactor existing code in-place. API client gains context.Context support. Root command gains SilenceUsage/SilenceErrors and --timeout flag. newClient returns error instead of calling os.Exit. All command RunE functions updated to handle newClient error.

**Tech Stack:** Go 1.26, existing deps (no new deps)

---

## File Structure

```
Modified files:
├── cmd/root.go              # SilenceUsage/SilenceErrors, --timeout flag, exit code handling
├── cmd/pages.go             # newClient() returns error, all RunE functions updated
├── cmd/templates.go         # newClient() call updated
├── cmd/blocks.go            # newClient() call updated
├── cmd/webhooks.go          # newClient() call updated, input validation before client
├── cmd/configure.go         # (no changes needed)
├── cmd/completion.go        # (no changes needed)
├── internal/api/client.go   # Add context.Context to all methods, default timeout
├── internal/api/client_test.go # Update tests for context parameter
└── internal/output/output.go   # NO_COLOR support
```

---

### Task 1: Add SilenceUsage and SilenceErrors to Root Command

**Files:**
- Modify: `cmd/root.go`

- [ ] **Step 1: Add SilenceUsage and SilenceErrors to rootCmd**

In `cmd/root.go`, change the rootCmd definition from:

```go
var rootCmd = &cobra.Command{
	Use:   "qwilr",
	Short: "CLI for the Qwilr API",
	Long:  "Manage Qwilr pages, templates, blocks, and webhooks from the command line.",
}
```

to:

```go
var rootCmd = &cobra.Command{
	Use:           "qwilr",
	Short:         "CLI for the Qwilr API",
	Long:          "Manage Qwilr pages, templates, blocks, and webhooks from the command line.",
	SilenceUsage:  true,
	SilenceErrors: true,
}
```

- [ ] **Step 2: Verify it builds**

Run:
```bash
cd /Users/paulmeller/Projects/qwilr-cli && go build -o qwilr .
```
Expected: Build succeeds.

- [ ] **Step 3: Verify errors no longer dump usage**

Run:
```bash
./qwilr pages get 2>&1; echo "exit: $?"
```
Expected: Shows only the error message (e.g., "accepts 1 arg(s), received 0"), NOT the full usage/help text. Exit code 1.

- [ ] **Step 4: Commit**

```bash
git add cmd/root.go
git commit -m "fix: add SilenceUsage/SilenceErrors to prevent usage dump on errors"
```

---

### Task 2: Add Context with Timeout to API Client

**Files:**
- Modify: `internal/api/client.go`
- Modify: `internal/api/client_test.go`

- [ ] **Step 1: Update test to use context**

In `internal/api/client_test.go`, add the `"context"` import and update `TestClientGet` to pass context:

Replace the entire file with:

```go
package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", r.Header.Get("Authorization"), "Bearer test-token")
		}
		if r.Method != http.MethodGet {
			t.Errorf("Method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/v1/pages" {
			t.Errorf("Path = %q, want /v1/pages", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": "pg_123"})
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	var result map[string]string
	err := client.Get(context.Background(), "/pages", &result)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if result["id"] != "pg_123" {
		t.Errorf("id = %q, want %q", result["id"], "pg_123")
	}
}

func TestClientPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Method = %q, want POST", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", r.Header.Get("Content-Type"))
		}
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "Test Page" {
			t.Errorf("name = %q, want %q", body["name"], "Test Page")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": "pg_new"})
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	reqBody := map[string]string{"name": "Test Page"}
	var result map[string]string
	err := client.Post(context.Background(), "/pages", reqBody, &result)
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	if result["id"] != "pg_new" {
		t.Errorf("id = %q, want %q", result["id"], "pg_new")
	}
}

func TestClientError401(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer server.Close()

	client := NewClient("bad-token", server.URL)
	var result map[string]string
	err := client.Get(context.Background(), "/pages", &result)
	if err == nil {
		t.Fatal("expected error for 401, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("StatusCode = %d, want 401", apiErr.StatusCode)
	}
}

func TestClientError404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "not found"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	var result map[string]string
	err := client.Get(context.Background(), "/pages/nonexistent", &result)
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
	}
}

func TestClientDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Method = %q, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	err := client.Delete(context.Background(), "/pages/pg_123")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestClientPut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Method = %q, want PUT", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": "pg_123", "name": "Updated"})
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	reqBody := map[string]string{"name": "Updated"}
	var result map[string]string
	err := client.Put(context.Background(), "/pages/pg_123", reqBody, &result)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}
	if result["name"] != "Updated" {
		t.Errorf("name = %q, want %q", result["name"], "Updated")
	}
}

func TestClientTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": "pg_123"})
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	var result map[string]string
	err := client.Get(ctx, "/pages", &result)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:
```bash
cd /Users/paulmeller/Projects/qwilr-cli && go test ./internal/api/ -v
```
Expected: Compilation failure — `Get` takes wrong number of arguments (too many).

- [ ] **Step 3: Update client.go to accept context.Context**

Replace the entire `internal/api/client.go` with:

```go
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const DefaultBaseURL = "https://api.qwilr.com"

type Client struct {
	token   string
	baseURL string
	http    *http.Client
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	switch e.StatusCode {
	case 401:
		return "authentication failed — run \"qwilr configure\" to set your API token"
	case 404:
		return fmt.Sprintf("resource not found (HTTP %d)", e.StatusCode)
	case 429:
		return "rate limited — wait a moment and try again"
	default:
		if e.StatusCode >= 500 {
			return fmt.Sprintf("Qwilr API error (HTTP %d) — try again or check https://status.qwilr.com", e.StatusCode)
		}
		return fmt.Sprintf("API error (HTTP %d): %s", e.StatusCode, e.Message)
	}
}

func NewClient(token, baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		token:   token,
		baseURL: baseURL,
		http:    &http.Client{},
	}
}

func (c *Client) do(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	url := c.baseURL + "/v1" + path
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}
	return resp, nil
}

func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) Post(ctx context.Context, path string, body, result interface{}) error {
	resp, err := c.do(ctx, http.MethodPost, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) Put(ctx context.Context, path string, body, result interface{}) error {
	resp, err := c.do(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) Delete(ctx context.Context, path string) error {
	resp, err := c.do(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run:
```bash
cd /Users/paulmeller/Projects/qwilr-cli && go test ./internal/api/ -v
```
Expected: All 7 tests PASS (including new TestClientTimeout).

- [ ] **Step 5: Commit**

```bash
git add internal/api/client.go internal/api/client_test.go
git commit -m "feat: add context.Context to API client methods with timeout support"
```

---

### Task 3: Refactor newClient to Return Error and Add --timeout Flag

**Files:**
- Modify: `cmd/root.go`
- Modify: `cmd/pages.go`

- [ ] **Step 1: Add --timeout flag and cmdContext helper to root.go**

Replace the entire `cmd/root.go` with:

```go
package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var jsonOutput bool
var timeout time.Duration

var rootCmd = &cobra.Command{
	Use:           "qwilr",
	Short:         "CLI for the Qwilr API",
	Long:          "Manage Qwilr pages, templates, blocks, and webhooks from the command line.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 30*time.Second, "API request timeout")
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.config/qwilr")
	viper.SetEnvPrefix("QWILR")
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()
}

func cmdContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}
```

- [ ] **Step 2: Refactor newClient to return error, update all pages command callers**

Replace the entire `cmd/pages.go` with:

```go
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/paulmeller/qwilr-cli/internal/api"
	"github.com/paulmeller/qwilr-cli/internal/config"
	"github.com/paulmeller/qwilr-cli/internal/output"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var pagesCmd = &cobra.Command{
	Use:   "pages",
	Short: "Manage Qwilr pages",
}

var pagesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List pages",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		ctx, cancel := cmdContext()
		defer cancel()

		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		path := fmt.Sprintf("/pages?limit=%d&offset=%d", limit, offset)
		var result []map[string]interface{}
		if err := client.Get(ctx, path, &result); err != nil {
			return err
		}

		if jsonOutput {
			return output.PrintJSON(os.Stdout, result)
		}

		headers := []string{"ID", "NAME", "STATUS", "CREATED"}
		var rows [][]string
		for _, p := range result {
			rows = append(rows, []string{
				stringVal(p, "id"),
				stringVal(p, "name"),
				stringVal(p, "status"),
				stringVal(p, "createdAt"),
			})
		}
		output.PrintTable(os.Stdout, headers, rows)
		return nil
	},
}

var pagesGetCmd = &cobra.Command{
	Use:   "get <pageId>",
	Short: "Get page details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		ctx, cancel := cmdContext()
		defer cancel()

		var result map[string]interface{}
		if err := client.Get(ctx, "/pages/"+args[0], &result); err != nil {
			return err
		}

		if jsonOutput {
			return output.PrintJSON(os.Stdout, result)
		}

		fields := []output.KeyValue{
			{Key: "Name", Value: stringVal(result, "name")},
			{Key: "ID", Value: stringVal(result, "id")},
			{Key: "Status", Value: stringVal(result, "status")},
			{Key: "Created", Value: stringVal(result, "createdAt")},
			{Key: "URL", Value: stringVal(result, "url")},
		}
		output.PrintDetail(os.Stdout, fields)
		return nil
	},
}

type pageCreateRequest struct {
	Name      string            `json:"name,omitempty" yaml:"name"`
	Template  string            `json:"template,omitempty" yaml:"template"`
	Published bool              `json:"published,omitempty" yaml:"published"`
	Variables map[string]string `json:"variables,omitempty" yaml:"variables"`
}

var pagesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new page",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		ctx, cancel := cmdContext()
		defer cancel()

		var req pageCreateRequest

		fromFile, _ := cmd.Flags().GetString("from")
		if fromFile != "" {
			data, err := os.ReadFile(fromFile)
			if err != nil {
				return fmt.Errorf("reading file %s: %w", fromFile, err)
			}
			if err := yaml.Unmarshal(data, &req); err != nil {
				return fmt.Errorf("parsing file %s: %w", fromFile, err)
			}
		}

		// Flags override file values
		if name, _ := cmd.Flags().GetString("name"); name != "" {
			req.Name = name
		}
		if tmpl, _ := cmd.Flags().GetString("template"); tmpl != "" {
			req.Template = tmpl
		}
		if cmd.Flags().Changed("publish") {
			pub, _ := cmd.Flags().GetBool("publish")
			req.Published = pub
		}

		var result map[string]interface{}
		if err := client.Post(ctx, "/pages", req, &result); err != nil {
			return err
		}

		if jsonOutput {
			return output.PrintJSON(os.Stdout, result)
		}

		fmt.Fprintf(os.Stderr, "Page created: %s\n", stringVal(result, "id"))
		fields := []output.KeyValue{
			{Key: "ID", Value: stringVal(result, "id")},
			{Key: "Name", Value: stringVal(result, "name")},
			{Key: "URL", Value: stringVal(result, "url")},
		}
		output.PrintDetail(os.Stdout, fields)
		return nil
	},
}

var pagesUpdateCmd = &cobra.Command{
	Use:   "update <pageId>",
	Short: "Update a page",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		ctx, cancel := cmdContext()
		defer cancel()

		body := make(map[string]interface{})
		if name, _ := cmd.Flags().GetString("name"); name != "" {
			body["name"] = name
		}
		if content, _ := cmd.Flags().GetString("content"); content != "" {
			body["content"] = content
		}

		var result map[string]interface{}
		if err := client.Put(ctx, "/pages/"+args[0], body, &result); err != nil {
			return err
		}

		if jsonOutput {
			return output.PrintJSON(os.Stdout, result)
		}

		fmt.Fprintf(os.Stderr, "Page updated: %s\n", args[0])
		return nil
	},
}

var pagesDeleteCmd = &cobra.Command{
	Use:   "delete <pageId>",
	Short: "Delete a page",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		ctx, cancel := cmdContext()
		defer cancel()

		if err := client.Delete(ctx, "/pages/"+args[0]); err != nil {
			return err
		}
		if !jsonOutput {
			fmt.Fprintf(os.Stderr, "Page deleted: %s\n", args[0])
		}
		return nil
	},
}

var pagesVariablesCmd = &cobra.Command{
	Use:   "variables <pageId>",
	Short: "Set token/substitution values on a page",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		ctx, cancel := cmdContext()
		defer cancel()

		vars := make(map[string]string)

		fromFile, _ := cmd.Flags().GetString("from")
		if fromFile != "" {
			data, err := os.ReadFile(fromFile)
			if err != nil {
				return fmt.Errorf("reading file %s: %w", fromFile, err)
			}
			if err := yaml.Unmarshal(data, &vars); err != nil {
				return fmt.Errorf("parsing file %s: %w", fromFile, err)
			}
		}

		setVals, _ := cmd.Flags().GetStringSlice("set")
		for _, kv := range setVals {
			parts := splitKeyValue(kv)
			if parts == nil {
				return fmt.Errorf("invalid --set value %q — expected key=value", kv)
			}
			vars[parts[0]] = parts[1]
		}

		var result map[string]interface{}
		if err := client.Put(ctx, "/pages/"+args[0]+"/variables", vars, &result); err != nil {
			return err
		}

		if jsonOutput {
			return output.PrintJSON(os.Stdout, result)
		}

		fmt.Fprintf(os.Stderr, "Variables updated on page %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pagesCmd)

	pagesListCmd.Flags().Int("limit", 20, "Maximum number of pages to return")
	pagesListCmd.Flags().Int("offset", 0, "Number of pages to skip")

	pagesCreateCmd.Flags().String("name", "", "Page name")
	pagesCreateCmd.Flags().String("template", "", "Template ID to use")
	pagesCreateCmd.Flags().String("from", "", "YAML/JSON file with page definition")
	pagesCreateCmd.Flags().Bool("publish", false, "Publish the page immediately")

	pagesUpdateCmd.Flags().String("name", "", "New page name")
	pagesUpdateCmd.Flags().String("content", "", "New page content")

	pagesVariablesCmd.Flags().String("from", "", "YAML/JSON file with variable values")
	pagesVariablesCmd.Flags().StringSlice("set", nil, "Set variable as key=value")

	pagesCmd.AddCommand(pagesListCmd)
	pagesCmd.AddCommand(pagesGetCmd)
	pagesCmd.AddCommand(pagesCreateCmd)
	pagesCmd.AddCommand(pagesUpdateCmd)
	pagesCmd.AddCommand(pagesDeleteCmd)
	pagesCmd.AddCommand(pagesVariablesCmd)
}

func newClient() (*api.Client, error) {
	token := config.ResolveToken(config.DefaultPath())
	if token == "" {
		return nil, fmt.Errorf("no API token configured — run \"qwilr configure\" or set QWILR_API_TOKEN")
	}
	return api.NewClient(token, ""), nil
}

func stringVal(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	case json.Number:
		return val.String()
	default:
		return fmt.Sprintf("%v", val)
	}
}

func splitKeyValue(s string) []string {
	for i, c := range s {
		if c == '=' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return nil
}
```

- [ ] **Step 3: Verify the project does NOT compile yet**

Run:
```bash
cd /Users/paulmeller/Projects/qwilr-cli && go build -o qwilr . 2>&1
```
Expected: Compilation errors in templates.go, blocks.go, webhooks.go — they still call `newClient()` with old signature and API methods without context.

- [ ] **Step 4: Commit pages and root changes**

```bash
git add cmd/root.go cmd/pages.go
git commit -m "feat: add --timeout flag, refactor newClient to return error, add context to pages commands"
```

---

### Task 4: Update Templates, Blocks, and Webhooks for Context + newClient Error

**Files:**
- Modify: `cmd/templates.go`
- Modify: `cmd/blocks.go`
- Modify: `cmd/webhooks.go`

- [ ] **Step 1: Update templates.go**

Replace the entire `cmd/templates.go` with:

```go
package cmd

import (
	"os"

	"github.com/paulmeller/qwilr-cli/internal/output"
	"github.com/spf13/cobra"
)

var templatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "Manage Qwilr templates",
}

var templatesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available templates",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		ctx, cancel := cmdContext()
		defer cancel()

		var result []map[string]interface{}
		if err := client.Get(ctx, "/templates", &result); err != nil {
			return err
		}

		if jsonOutput {
			return output.PrintJSON(os.Stdout, result)
		}

		headers := []string{"ID", "NAME"}
		var rows [][]string
		for _, t := range result {
			rows = append(rows, []string{
				stringVal(t, "id"),
				stringVal(t, "name"),
			})
		}
		output.PrintTable(os.Stdout, headers, rows)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(templatesCmd)
	templatesCmd.AddCommand(templatesListCmd)
}
```

- [ ] **Step 2: Update blocks.go**

Replace the entire `cmd/blocks.go` with:

```go
package cmd

import (
	"os"

	"github.com/paulmeller/qwilr-cli/internal/output"
	"github.com/spf13/cobra"
)

var blocksCmd = &cobra.Command{
	Use:   "blocks",
	Short: "Manage Qwilr blocks",
}

var blocksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved blocks",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		ctx, cancel := cmdContext()
		defer cancel()

		var result []map[string]interface{}
		if err := client.Get(ctx, "/blocks", &result); err != nil {
			return err
		}

		if jsonOutput {
			return output.PrintJSON(os.Stdout, result)
		}

		headers := []string{"ID", "NAME", "TYPE"}
		var rows [][]string
		for _, b := range result {
			rows = append(rows, []string{
				stringVal(b, "id"),
				stringVal(b, "name"),
				stringVal(b, "type"),
			})
		}
		output.PrintTable(os.Stdout, headers, rows)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(blocksCmd)
	blocksCmd.AddCommand(blocksListCmd)
}
```

- [ ] **Step 3: Update webhooks.go with context + newClient error + validation before client**

Replace the entire `cmd/webhooks.go` with:

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/paulmeller/qwilr-cli/internal/output"
	"github.com/spf13/cobra"
)

var webhooksCmd = &cobra.Command{
	Use:   "webhooks",
	Short: "Manage webhook subscriptions",
}

var webhooksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List webhook subscriptions",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		ctx, cancel := cmdContext()
		defer cancel()

		var result []map[string]interface{}
		if err := client.Get(ctx, "/webhooks", &result); err != nil {
			return err
		}

		if jsonOutput {
			return output.PrintJSON(os.Stdout, result)
		}

		headers := []string{"ID", "URL", "EVENT"}
		var rows [][]string
		for _, w := range result {
			rows = append(rows, []string{
				stringVal(w, "id"),
				stringVal(w, "url"),
				stringVal(w, "event"),
			})
		}
		output.PrintTable(os.Stdout, headers, rows)
		return nil
	},
}

var webhooksCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a webhook subscription",
	RunE: func(cmd *cobra.Command, args []string) error {
		url, _ := cmd.Flags().GetString("url")
		event, _ := cmd.Flags().GetString("event")

		if url == "" {
			return fmt.Errorf("--url is required")
		}
		if event == "" {
			return fmt.Errorf("--event is required")
		}

		client, err := newClient()
		if err != nil {
			return err
		}
		ctx, cancel := cmdContext()
		defer cancel()

		body := map[string]string{
			"url":   url,
			"event": event,
		}

		var result map[string]interface{}
		if err := client.Post(ctx, "/webhooks", body, &result); err != nil {
			return err
		}

		if jsonOutput {
			return output.PrintJSON(os.Stdout, result)
		}

		fmt.Fprintf(os.Stderr, "Webhook created: %s\n", stringVal(result, "id"))
		fields := []output.KeyValue{
			{Key: "ID", Value: stringVal(result, "id")},
			{Key: "URL", Value: stringVal(result, "url")},
			{Key: "Event", Value: stringVal(result, "event")},
		}
		output.PrintDetail(os.Stdout, fields)
		return nil
	},
}

var webhooksDeleteCmd = &cobra.Command{
	Use:   "delete <webhookId>",
	Short: "Delete a webhook subscription",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		ctx, cancel := cmdContext()
		defer cancel()

		if err := client.Delete(ctx, "/webhooks/"+args[0]); err != nil {
			return err
		}
		if !jsonOutput {
			fmt.Fprintf(os.Stderr, "Webhook deleted: %s\n", args[0])
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(webhooksCmd)

	webhooksCreateCmd.Flags().String("url", "", "Webhook callback URL")
	webhooksCreateCmd.Flags().String("event", "", "Event type (pageAccepted, pagePreviewAccepted, pageViewed, pageFirstViewed)")

	webhooksCmd.AddCommand(webhooksListCmd)
	webhooksCmd.AddCommand(webhooksCreateCmd)
	webhooksCmd.AddCommand(webhooksDeleteCmd)
}
```

- [ ] **Step 4: Verify the project compiles and all tests pass**

Run:
```bash
cd /Users/paulmeller/Projects/qwilr-cli && go build -o qwilr . && go test ./... -v
```
Expected: Build succeeds. All 12 tests pass (7 API + 5 config + 3 output... wait, we added TestClientTimeout so 7 API tests).

- [ ] **Step 5: Verify --timeout flag appears in help**

Run:
```bash
./qwilr --help
```
Expected: `--timeout` flag listed with default "30s".

- [ ] **Step 6: Commit**

```bash
git add cmd/templates.go cmd/blocks.go cmd/webhooks.go
git commit -m "feat: update templates, blocks, webhooks for context and newClient error handling"
```

---

### Task 5: Add NO_COLOR Support to Output Package

**Files:**
- Modify: `internal/output/output.go`
- Modify: `internal/output/output_test.go`

- [ ] **Step 1: Add NO_COLOR test**

Add this test to `internal/output/output_test.go`:

```go
func TestNoColorDetection(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	if !NoColor() {
		t.Error("NoColor() = false, want true when NO_COLOR is set")
	}
}

func TestNoColorUnset(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	if NoColor() {
		t.Error("NoColor() = true, want false when NO_COLOR is unset")
	}
}
```

Also add `"os"` to the imports in the test file.

- [ ] **Step 2: Run tests to verify they fail**

Run:
```bash
cd /Users/paulmeller/Projects/qwilr-cli && go test ./internal/output/ -v
```
Expected: Compilation failure — `NoColor` undefined.

- [ ] **Step 3: Add NoColor function to output.go**

Add this import and function to `internal/output/output.go`. Add `"os"` to the imports, then add after the `PrintDetail` function:

```go
func NoColor() bool {
	_, ok := os.LookupEnv("NO_COLOR")
	return ok
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run:
```bash
cd /Users/paulmeller/Projects/qwilr-cli && go test ./internal/output/ -v
```
Expected: All 5 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/output/output.go internal/output/output_test.go
git commit -m "feat: add NO_COLOR environment variable support"
```

---

### Task 6: Add Structured Exit Codes

**Files:**
- Modify: `cmd/root.go`

- [ ] **Step 1: Update Execute() with structured exit codes**

In `cmd/root.go`, change the `Execute` function from:

```go
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
```

to:

```go
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		if isValidationError(err) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}

func isValidationError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "is required") ||
		strings.Contains(msg, "cannot be empty") ||
		strings.Contains(msg, "invalid") ||
		strings.Contains(msg, "no API token configured")
}
```

Also add `"strings"` to the imports.

- [ ] **Step 2: Verify it builds**

Run:
```bash
cd /Users/paulmeller/Projects/qwilr-cli && go build -o qwilr .
```
Expected: Build succeeds.

- [ ] **Step 3: Verify exit code 2 for validation errors**

Run:
```bash
./qwilr webhooks create 2>/dev/null; echo "exit: $?"
```
Expected: `exit: 2` (because --url is required, which is a validation error).

- [ ] **Step 4: Commit**

```bash
git add cmd/root.go
git commit -m "feat: add structured exit codes (1=general, 2=validation)"
```

---

### Task 7: Final Verification

**Files:** None (verification only)

- [ ] **Step 1: Run all tests**

Run:
```bash
cd /Users/paulmeller/Projects/qwilr-cli && go test ./... -v
```
Expected: All tests pass across all packages.

- [ ] **Step 2: Build and verify help**

Run:
```bash
cd /Users/paulmeller/Projects/qwilr-cli && go build -o qwilr . && ./qwilr --help
```
Expected: Help shows --json, --timeout flags. All commands listed.

- [ ] **Step 3: Verify error behavior (no usage dump)**

Run:
```bash
./qwilr pages get 2>&1
```
Expected: Just the error message, no usage text.

- [ ] **Step 4: Verify validation exit code**

Run:
```bash
./qwilr webhooks create 2>/dev/null; echo $?
```
Expected: Exit code 2.
