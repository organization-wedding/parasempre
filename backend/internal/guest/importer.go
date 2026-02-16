package guest

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/xuri/excelize/v2"
)

var requiredColumns = []string{"first_name", "last_name", "phone", "relationship"}

// ParseCSV reads CSV data and returns a slice of CreateGuestInput.
// Expects a header row with columns: first_name, last_name, phone, relationship.
func ParseCSV(r io.Reader) ([]CreateGuestInput, error) {
	reader := csv.NewReader(r)

	header, err := reader.Read()
	if err != nil {
		return nil, errors.New("failed to read CSV header")
	}

	colIndex, err := mapColumns(header)
	if err != nil {
		return nil, err
	}

	var guests []CreateGuestInput
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row: %w", err)
		}

		guests = append(guests, CreateGuestInput{
			FirstName:    strings.TrimSpace(record[colIndex["first_name"]]),
			LastName:     strings.TrimSpace(record[colIndex["last_name"]]),
			Phone:        strings.TrimSpace(record[colIndex["phone"]]),
			Relationship: strings.TrimSpace(record[colIndex["relationship"]]),
		})
	}

	return guests, nil
}

// ParseXLSX reads XLSX data and returns a slice of CreateGuestInput.
// Reads the first sheet; expects a header row with columns: first_name, last_name, phone, relationship.
func ParseXLSX(r io.Reader) ([]CreateGuestInput, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to open XLSX: %w", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read XLSX rows: %w", err)
	}

	if len(rows) == 0 {
		return nil, errors.New("XLSX file is empty")
	}

	colIndex, err := mapColumns(rows[0])
	if err != nil {
		return nil, err
	}

	var guests []CreateGuestInput
	for _, row := range rows[1:] {
		// Skip rows that don't have enough columns.
		maxIdx := 0
		for _, idx := range colIndex {
			if idx > maxIdx {
				maxIdx = idx
			}
		}
		if len(row) <= maxIdx {
			continue
		}

		guests = append(guests, CreateGuestInput{
			FirstName:    strings.TrimSpace(row[colIndex["first_name"]]),
			LastName:     strings.TrimSpace(row[colIndex["last_name"]]),
			Phone:        strings.TrimSpace(row[colIndex["phone"]]),
			Relationship: strings.TrimSpace(row[colIndex["relationship"]]),
		})
	}

	return guests, nil
}

func mapColumns(header []string) (map[string]int, error) {
	index := make(map[string]int)
	for i, col := range header {
		index[strings.TrimSpace(strings.ToLower(col))] = i
	}

	for _, required := range requiredColumns {
		if _, ok := index[required]; !ok {
			return nil, fmt.Errorf("missing required column: %s", required)
		}
	}

	return index, nil
}
