package avro

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/hamba/avro"
)

func TestAvroValidator_Validate(t *testing.T) {
	avroV := New()

	tt := []struct {
		name                        string
		data                        interface{}
		serializationSchemaFilename string
		validationSchemaFilename    string
		valid                       bool
	}{
		{
			"valid-1",
			struct {
				Name       string      `avro:"name"`
				Age        int         `avro:"age"`
				Collection interface{} `avro:"collection"`
				Foo        interface{} `avro:"foo"`
			}{
				"Syntio",
				4,
				[]string{"data engineering", "avro"},
				struct {
					Bar string `avro:"bar"`
				}{
					"hello world",
				},
			},
			"valid-1-serialization-schema.avsc",
			"valid-1-validation-schema.avsc",
			true,
		},
		{
			"invalid-1",
			struct {
				Name       string      `avro:"name"`
				Age        int         `avro:"age"`
				Collection interface{} `avro:"collection"`
			}{
				"Syntio",
				4,
				[]string{"data engineering", "avro"},
			},
			"invalid-1-serialization-schema.avsc",
			"invalid-1-validation-schema.avsc",
			false,
		},
		{
			"invalid-2",
			struct {
				Age      int  `avro:"age"`
				Tall     bool `avro:"tall"`
				Handsome bool `avro:"handsome"`
			}{
				2,
				true,
				true,
			},
			"invalid-2-serialization-schema.avsc",
			"invalid-2-validation-schema.avsc",
			// IT IS FUNDAMENTALLY IMPOSSIBLE TO COVER THIS CASE
			// BECAUSE OF THE WAY AVRO WORKS
			// THIS IS A KNOWN AND ACCEPTABLE LIMITATION
			true,
		},
		{
			"invalid-3",
			struct {
				Age      int  `avro:"age"`
				Tall     bool `avro:"tall"`
				Handsome bool `avro:"handsome"`
			}{
				2,
				true,
				false,
			},
			"invalid-3-serialization-schema.avsc",
			"invalid-3-validation-schema.avsc",
			// IT IS FUNDAMENTALLY IMPOSSIBLE TO COVER THIS CASE
			// BECAUSE OF THE WAY AVRO WORKS
			// THIS IS A KNOWN AND ACCEPTABLE LIMITATION
			true,
		},
		{
			"invalid-4",
			struct {
				Age    int `avro:"age"`
				Height int `avro:"height"`
				Length int `avro:"length"`
			}{
				4,
				64,
				-65,
			},
			"invalid-4-serialization-schema.avsc",
			"invalid-4-validation-schema.avsc",
			// IT IS FUNDAMENTALLY IMPOSSIBLE TO COVER THIS CASE
			// BECAUSE OF THE WAY AVRO WORKS
			// THIS IS A KNOWN AND ACCEPTABLE LIMITATION
			true,
		},
		{
			"invalid-5",
			struct {
				Age  int `avro:"age"`
				Pos0 int `avro:"pos0"`
				Pos1 int `avro:"pos1"`
				Pos2 int `avro:"pos2"`
				Pos3 int `avro:"pos3"`
				Pos4 int `avro:"pos4"`
			}{
				5,
				// 36, // H
				// -35, // E
				// 38, // L
				// 38, // L
				// -40, // O
				40,  // P
				-44, // W
				39,  // N
				-35, // E
				34,  // D
			},
			"invalid-5-serialization-schema.avsc",
			"invalid-5-validation-schema.avsc",
			// IT IS FUNDAMENTALLY IMPOSSIBLE TO COVER THIS CASE
			// BECAUSE OF THE WAY AVRO WORKS
			// THIS IS A KNOWN AND ACCEPTABLE LIMITATION
			true,
		},
		{
			"invalid-6",
			struct {
				Pos1 string `avro:"pos1"`
				Pos0 string `avro:"pos0"`
			}{
				"SYNTIO!",
				"HELLO, ",
			},
			"invalid-6-serialization-schema.avsc",
			"invalid-6-validation-schema.avsc",
			// IT IS FUNDAMENTALLY IMPOSSIBLE TO COVER THIS CASE
			// BECAUSE OF THE WAY AVRO WORKS
			// THIS IS A KNOWN AND ACCEPTABLE LIMITATION
			true,
		},
		{
			"invalid-7",
			struct {
				Pos0 string `avro:"pos0"`
			}{
				"HELLO",
			},
			"invalid-7-serialization-schema.avsc",
			"invalid-7-validation-schema.avsc",
			false,
		},
	}

	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	testdataDir := filepath.Join(basepath, "testdata")
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			serializationSchemaTxt, err := os.ReadFile(filepath.Join(testdataDir, tc.serializationSchemaFilename))
			if err != nil {
				t.Fatalf("serialization schema read error")
			}

			serializationSchema, err := avro.Parse(string(serializationSchemaTxt))
			if err != nil {
				t.Fatalf("avro serialization parse error: %s", err)
			}

			data, err := avro.Marshal(serializationSchema, tc.data)
			if err != nil {
				t.Fatalf("avro serialization error: %s", err)
			}

			validationSchema, err := os.ReadFile(filepath.Join(testdataDir, tc.validationSchemaFilename))
			if err != nil {
				t.Fatalf("validation schema read error")
			}

			valid, err := avroV.Validate(data, validationSchema, "", "")
			if err != nil {
				t.Fatalf("validator error: %s", err)
			}
			if valid != tc.valid {
				if valid {
					t.Errorf("message valid, invalid expected")
				} else {
					t.Errorf("message invalid, valid expected")
				}
			}
		})
	}
}
