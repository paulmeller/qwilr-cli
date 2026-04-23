package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintTable(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"ID", "NAME", "STATUS"}
	rows := [][]string{
		{"pg_123", "My Page", "live"},
		{"pg_456", "Draft Page", "draft"},
	}
	PrintTable(&buf, headers, rows)
	out := buf.String()

	if !strings.Contains(out, "ID") {
		t.Errorf("output missing header 'ID': %s", out)
	}
	if !strings.Contains(out, "pg_123") {
		t.Errorf("output missing row data 'pg_123': %s", out)
	}
	if !strings.Contains(out, "draft") {
		t.Errorf("output missing row data 'draft': %s", out)
	}
}

func TestPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"id": "pg_123", "name": "My Page"}
	err := PrintJSON(&buf, data)
	if err != nil {
		t.Fatalf("PrintJSON failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"id": "pg_123"`) {
		t.Errorf("output missing id: %s", out)
	}
}

func TestPrintDetail(t *testing.T) {
	var buf bytes.Buffer
	fields := []KeyValue{
		{Key: "Name", Value: "My Page"},
		{Key: "ID", Value: "pg_123"},
		{Key: "Status", Value: "live"},
	}
	PrintDetail(&buf, fields)
	out := buf.String()
	if !strings.Contains(out, "Name:") {
		t.Errorf("output missing 'Name:': %s", out)
	}
	if !strings.Contains(out, "pg_123") {
		t.Errorf("output missing 'pg_123': %s", out)
	}
}
