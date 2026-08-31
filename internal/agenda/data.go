package agenda

import (
	"encoding/json"
	"fmt"
	"os"
)

type Data struct {
	Schema    string    `json:"$schema,omitempty"`
	Title     string    `json:"title,omitempty"`
	Week      string    `json:"week,omitempty"`
	Timezone  string    `json:"timezone,omitempty"`
	Days      []Day     `json:"days,omitempty"`
	People    []Person  `json:"people"`
	TimeRange TimeRange `json:"timeRange"`
	Slots     []Slot    `json:"slots,omitempty"`
	Events    []Event   `json:"events"`
}

type Day string

type Person struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type TimeRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type Slot struct {
	Label string `json:"label"`
	Start string `json:"start"`
	End   string `json:"end"`
}

type Event struct {
	Person   string `json:"person"`
	Day      Day    `json:"day"`
	Start    string `json:"start"`
	End      string `json:"end"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle,omitempty"`
	Note     string `json:"note,omitempty"`
	Style    string `json:"style,omitempty"`
	Stack    Stack  `json:"stack,omitempty"`
}

type Stack struct {
	Index int `json:"index"`
	Total int `json:"total"`
}

func LoadData(path string) (Data, error) {
	file, err := os.Open(path)
	if err != nil {
		return Data{}, fmt.Errorf("open data: %w", err)
	}

	var data Data
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&data); err != nil {
		if closeErr := file.Close(); closeErr != nil {
			return Data{}, fmt.Errorf("close data: %w", closeErr)
		}
		return Data{}, fmt.Errorf("decode data: %w", err)
	}
	if err := file.Close(); err != nil {
		return Data{}, fmt.Errorf("close data: %w", err)
	}

	return data, nil
}
