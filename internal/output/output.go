package output

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"
)

type KeyValue struct {
	Key   string
	Value string
}

func PrintTable(w io.Writer, headers []string, rows [][]string) {
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	for i, h := range headers {
		if i > 0 {
			fmt.Fprint(tw, "\t")
		}
		fmt.Fprint(tw, h)
	}
	fmt.Fprintln(tw)

	for _, row := range rows {
		for i, col := range row {
			if i > 0 {
				fmt.Fprint(tw, "\t")
			}
			fmt.Fprint(tw, col)
		}
		fmt.Fprintln(tw)
	}
	tw.Flush()
}

func PrintJSON(w io.Writer, data interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func PrintDetail(w io.Writer, fields []KeyValue) {
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	for _, f := range fields {
		fmt.Fprintf(tw, "%s:\t%s\n", f.Key, f.Value)
	}
	tw.Flush()
}
