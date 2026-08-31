package agenda

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed schema/agenda.schema.json
var schemaFS embed.FS

var (
	agendaSchemaOnce sync.Once
	agendaSchema     *jsonschema.Schema
	agendaSchemaErr  error
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
}

func LoadData(path string) (Data, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Data{}, fmt.Errorf("open data: %w", err)
	}
	if err := validateData(raw); err != nil {
		return Data{}, err
	}

	var data Data
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&data); err != nil {
		return Data{}, fmt.Errorf("decode data: %w", err)
	}

	return data, nil
}

func validateData(raw []byte) error {
	schema, err := loadAgendaSchema()
	if err != nil {
		return err
	}

	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("decode data: %w", err)
	}
	if err := schema.Validate(instance); err != nil {
		return fmt.Errorf("validate data: %w", err)
	}

	return nil
}

func loadAgendaSchema() (*jsonschema.Schema, error) {
	agendaSchemaOnce.Do(func() {
		schemaFile, err := schemaFS.Open("schema/agenda.schema.json")
		if err != nil {
			agendaSchemaErr = fmt.Errorf("open schema: %w", err)
			return
		}
		defer func() {
			if err := schemaFile.Close(); err != nil && agendaSchemaErr == nil {
				agendaSchemaErr = fmt.Errorf("close schema: %w", err)
			}
		}()

		schemaData, err := io.ReadAll(schemaFile)
		if err != nil {
			agendaSchemaErr = fmt.Errorf("read schema: %w", err)
			return
		}
		schemaDoc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaData))
		if err != nil {
			agendaSchemaErr = fmt.Errorf("decode schema: %w", err)
			return
		}

		compiler := jsonschema.NewCompiler()
		if err := compiler.AddResource("agenda.schema.json", schemaDoc); err != nil {
			agendaSchemaErr = fmt.Errorf("add schema: %w", err)
			return
		}
		agendaSchema, agendaSchemaErr = compiler.Compile("agenda.schema.json")
		if agendaSchemaErr != nil {
			agendaSchemaErr = fmt.Errorf("compile schema: %w", agendaSchemaErr)
		}
	})
	return agendaSchema, agendaSchemaErr
}
