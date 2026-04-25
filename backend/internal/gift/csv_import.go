package gift

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

var csvRequiredColumns = []string{"name", "price_brl"}

var httpsURLRegex = regexp.MustCompile(`^https://[^\s]+$`)

func ParseCSVRows(r io.Reader) ([]CSVPreviewRow, error) {
	bufR := bufio.NewReader(r)

	if head, _ := bufR.Peek(3); len(head) >= 3 && head[0] == 0xEF && head[1] == 0xBB && head[2] == 0xBF {
		_, _ = bufR.Discard(3)
	}

	reader := csv.NewReader(bufR)
	reader.Comma = detectDelimiter(bufR)
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err == io.EOF {
		return nil, errors.New("CSV file is empty")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	colIndex, err := mapCSVColumns(header)
	if err != nil {
		return nil, err
	}

	var rows []CSVPreviewRow
	lineNum := 2 // header was line 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV line %d: %w", lineNum, err)
		}

		rows = append(rows, parseCSVRow(record, colIndex, lineNum))
		lineNum++
	}

	if len(rows) == 0 {
		return nil, errors.New("CSV file has no data rows")
	}
	return rows, nil
}

func parseCSVRow(record []string, colIndex map[string]int, lineNum int) CSVPreviewRow {
	row := CSVPreviewRow{LineNumber: lineNum}

	col := func(key string) string {
		idx, ok := colIndex[key]
		if !ok || idx >= len(record) {
			return ""
		}
		return strings.TrimSpace(record[idx])
	}

	name := col("name")
	priceStr := col("price_brl")
	description := col("description")
	imageURL := col("image_url")
	storeURL := col("store_url")

	row.Input.Name = name
	if name == "" {
		row.Errors = append(row.Errors, "name is required")
	} else if len(name) > 200 {
		row.Errors = append(row.Errors, "name too long (max 200 chars)")
	}

	cents, err := parsePriceBRL(priceStr)
	switch {
	case err != nil:
		row.Errors = append(row.Errors, fmt.Sprintf("price_brl: %s", err.Error()))
	case cents <= 0:
		row.Errors = append(row.Errors, "price_brl must be greater than zero")
	default:
		row.Input.PriceCents = cents
	}

	if description != "" {
		if len(description) > 2000 {
			row.Errors = append(row.Errors, "description too long (max 2000 chars)")
		} else {
			d := description
			row.Input.Description = &d
		}
	}

	if imageURL != "" {
		if !httpsURLRegex.MatchString(imageURL) {
			row.Errors = append(row.Errors, "image_url must be a valid HTTPS URL")
		} else {
			u := imageURL
			row.Input.ImageURL = &u
		}
	}

	if storeURL != "" {
		if !httpsURLRegex.MatchString(storeURL) {
			row.Errors = append(row.Errors, "store_url must be a valid HTTPS URL")
		} else {
			u := storeURL
			row.Input.StoreURL = &u
		}
	}

	if name != "" {
		row.DedupeKey = NormalizeDedupeKey(name)
	}

	return row
}

func detectDelimiter(bufR *bufio.Reader) rune {
	peek, _ := bufR.Peek(4096)
	if idx := bytes.IndexAny(peek, "\r\n"); idx >= 0 {
		peek = peek[:idx]
	}
	if bytes.Count(peek, []byte(";")) > bytes.Count(peek, []byte(",")) {
		return ';'
	}
	return ','
}

func mapCSVColumns(header []string) (map[string]int, error) {
	index := make(map[string]int, len(header))
	for i, col := range header {
		index[strings.TrimSpace(strings.ToLower(col))] = i
	}
	for _, required := range csvRequiredColumns {
		if _, ok := index[required]; !ok {
			return nil, fmt.Errorf("missing required column: %q", required)
		}
	}
	return index, nil
}

func parsePriceBRL(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("required")
	}
	s = strings.ReplaceAll(s, "R$", "")
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", ".")

	parts := strings.Split(s, ".")
	var intPart, fracPart string
	switch len(parts) {
	case 1:
		intPart = parts[0]
		fracPart = "00"
	case 2:
		intPart = parts[0]
		fracPart = parts[1]
	default:
		return 0, fmt.Errorf("invalid format %q", s)
	}

	switch len(fracPart) {
	case 0:
		fracPart = "00"
	case 1:
		fracPart += "0"
	case 2:
		// ok
	default:
		return 0, fmt.Errorf("decimal part must have at most 2 digits, got %q", fracPart)
	}

	intVal, err := strconv.ParseInt(intPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid integer part %q", intPart)
	}
	fracVal, err := strconv.ParseInt(fracPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid decimal part %q", fracPart)
	}
	if intVal < 0 || fracVal < 0 {
		return 0, errors.New("must be positive")
	}
	return intVal*100 + fracVal, nil
}
