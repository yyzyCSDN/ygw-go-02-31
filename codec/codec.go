package codec

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"example.com/eventledger/model"
)

func EncodeJSON(values []model.EventRecord) ([]byte, error) {
	cloned := make([]model.EventRecord, len(values))
	for index, value := range values {
		cloned[index] = value.Clone()
	}
	data, err := json.MarshalIndent(cloned, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode eventRecord list: %w", err)
	}
	return append(data, '\n'), nil
}

func DecodeJSON(reader io.Reader) ([]model.EventRecord, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	var values []model.EventRecord
	if err := decoder.Decode(&values); err != nil {
		return nil, fmt.Errorf("decode eventRecord list: %w", err)
	}
	for index := range values {
		if err := values[index].Validate(); err != nil {
			return nil, fmt.Errorf("eventRecord %d: %w", index+1, err)
		}
		values[index] = values[index].Clone()
	}
	return values, nil
}

func EncodeCSV(values []model.EventRecord) ([]byte, error) {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	_ = writer.Write([]string{"id", "name", "status", "priority", "amount", "active", "version", "tags", "updated_at"})
	for _, value := range values {
		row := []string{
			value.ID,
			value.Name,
			value.Status,
			strconv.Itoa(value.Priority),
			strconv.FormatInt(value.Amount, 10),
			strconv.FormatBool(value.Active),
			strconv.FormatUint(value.Version, 10),
			strings.Join(model.NormalizeTags(value.Tags), "|"),
			value.UpdatedAt.UTC().Format(time.RFC3339Nano),
		}
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("write eventRecord CSV: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("flush eventRecord CSV: %w", err)
	}
	return buffer.Bytes(), nil
}

func DecodeCSV(reader io.Reader) ([]model.EventRecord, error) {
	rows, err := csv.NewReader(reader).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read eventRecord CSV: %w", err)
	}
	if len(rows) == 0 {
		return []model.EventRecord{}, nil
	}
	wantHeader := []string{"id", "name", "status", "priority", "amount", "active", "version", "tags", "updated_at"}
	if len(rows[0]) != len(wantHeader) {
		return nil, fmt.Errorf("unexpected CSV header width")
	}
	for index := range wantHeader {
		if rows[0][index] != wantHeader[index] {
			return nil, fmt.Errorf("unexpected CSV header %q", rows[0][index])
		}
	}
	result := make([]model.EventRecord, 0, len(rows)-1)
	for rowNumber, row := range rows[1:] {
		if len(row) != len(wantHeader) {
			return nil, fmt.Errorf("row %d has unexpected width", rowNumber+2)
		}
		priority, err := strconv.Atoi(row[3])
		if err != nil {
			return nil, fmt.Errorf("row %d priority: %w", rowNumber+2, err)
		}
		amount, err := strconv.ParseInt(row[4], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("row %d amount: %w", rowNumber+2, err)
		}
		active, err := strconv.ParseBool(row[5])
		if err != nil {
			return nil, fmt.Errorf("row %d active: %w", rowNumber+2, err)
		}
		version, err := strconv.ParseUint(row[6], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("row %d version: %w", rowNumber+2, err)
		}
		updated, err := time.Parse(time.RFC3339Nano, row[8])
		if err != nil {
			return nil, fmt.Errorf("row %d updated_at: %w", rowNumber+2, err)
		}
		value := model.EventRecord{ID: row[0], Name: row[1], Status: row[2], Priority: priority, Amount: amount, Active: active, Version: version, Tags: strings.Split(row[7], "|"), UpdatedAt: updated}
		if err := value.Validate(); err != nil {
			return nil, fmt.Errorf("row %d: %w", rowNumber+2, err)
		}
		result = append(result, value)
	}
	return result, nil
}
