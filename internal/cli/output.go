package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/samber/lo"
)

func WriteJSON(writer io.Writer, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(writer, string(raw))
	return err
}

func WriteTable(writer io.Writer, headers []string, rows [][]string) error {
	w := tabwriter.NewWriter(writer, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, strings.Join(headers, "\t"))
	lo.ForEach(rows, func(row []string, _ int) {
		_, _ = fmt.Fprintln(w, strings.Join(row, "\t"))
	})
	return w.Flush()
}
