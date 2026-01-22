// Package select pkg/select/select_test.go
package s3select

import (
	"encoding/xml"
	"strings"
	"testing"
)

func TestOptionsCSVSerialization(t *testing.T) {
	csvIn := &CSVInputOptions{}
	csvIn.SetFileHeaderInfo(CSVFileHeaderInfoUse)
	csvIn.SetFieldDelimiter(",")
	csvOut := &CSVOutputOptions{}
	csvOut.SetRecordDelimiter("\n")

	opts := Options{
		Expression:     "SELECT * FROM S3Object",
		ExpressionType: QueryExpressionTypeSQL,
		InputSerialization: InputSerialization{
			CSV: csvIn,
		},
		OutputSerialization: OutputSerialization{
			CSV: csvOut,
		},
		RequestProgress: RequestProgress{Enabled: true},
	}

	data, err := xml.Marshal(opts)
	if err != nil {
		t.Fatalf("xml.Marshal error = %v", err)
	}
	encoded := string(data)
	if !strings.Contains(encoded, "<CSV>") {
		t.Fatalf("expected CSV input serialization, got %s", encoded)
	}
	if !strings.Contains(encoded, "<FileHeaderInfo>USE</FileHeaderInfo>") {
		t.Fatalf("expected FileHeaderInfo, got %s", encoded)
	}
	if !strings.Contains(encoded, "<RecordDelimiter>") {
		t.Fatalf("expected record delimiter output, got %s", encoded)
	}
}

func TestOptionsJSONSerialization(t *testing.T) {
	jsonIn := &JSONInputOptions{}
	jsonIn.SetType(JSONLinesType)
	jsonOut := &JSONOutputOptions{}
	jsonOut.SetRecordDelimiter("\n")

	opts := Options{
		Expression:     "SELECT * FROM S3Object",
		ExpressionType: QueryExpressionTypeSQL,
		InputSerialization: InputSerialization{
			JSON: jsonIn,
		},
		OutputSerialization: OutputSerialization{
			JSON: jsonOut,
		},
		RequestProgress: RequestProgress{Enabled: true},
	}

	data, err := xml.Marshal(opts)
	if err != nil {
		t.Fatalf("xml.Marshal error = %v", err)
	}
	encoded := string(data)
	if !strings.Contains(encoded, "<JSON>") {
		t.Fatalf("expected JSON input serialization, got %s", encoded)
	}
	if !strings.Contains(encoded, "<Type>LINES</Type>") {
		t.Fatalf("expected JSON type, got %s", encoded)
	}
}

func TestOptionsParquetSerialization(t *testing.T) {
	opts := Options{
		Expression:     "SELECT * FROM S3Object",
		ExpressionType: QueryExpressionTypeSQL,
		InputSerialization: InputSerialization{
			Parquet: &ParquetInputOptions{},
		},
		OutputSerialization: OutputSerialization{
			CSV: &CSVOutputOptions{},
		},
		RequestProgress: RequestProgress{Enabled: true},
	}

	data, err := xml.Marshal(opts)
	if err != nil {
		t.Fatalf("xml.Marshal error = %v", err)
	}
	encoded := string(data)
	if !strings.Contains(encoded, "<Parquet>") {
		t.Fatalf("expected Parquet input serialization, got %s", encoded)
	}
}
