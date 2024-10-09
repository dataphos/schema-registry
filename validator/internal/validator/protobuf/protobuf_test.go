package protobuf

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/validator/protobuf/testdata/person"
	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/validator/protobuf/testdata/testpb3"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestValidate(t *testing.T) {
	dir := "./schemas"
	v, err := New(dir, 10)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	tt := []struct {
		name           string
		data           string
		schemaId       string
		schemaVersion  string
		schemaFilename string
		valid          bool
	}{
		{
			"Proto2-valid-1",
			"valid2-1.pb",
			"1",
			"1",
			"schema2-1.proto",
			true,
		},
		{
			"Proto2-valid-2",
			"valid2-2.pb",
			"1",
			"1",
			"schema2-1.proto",
			true,
		},
		{
			"Proto2-valid-3",
			"valid2-3.pb",
			"1",
			"1",
			"schema2-1.proto",
			true,
		},
		{
			"Proto2-valid-4",
			"valid2-4.pb",
			"1",
			"1",
			"schema2-1.proto",
			true,
		},
		{
			"Proto2-invalid-1",
			"invalid2-1.pb",
			"1",
			"1",
			"schema2-1.proto",
			false,
		},
		{
			"Proto2-invalid-2",
			"invalid2-2.pb",
			"1",
			"1",
			"schema2-1.proto",
			false,
		},
		{
			"Proto2-invalid-3",
			"valid3-8.pb",
			"1",
			"1",
			"schema2-1.proto",
			false,
		},
		{
			"Proto3-valid-1",
			"valid3-1.pb",
			"2",
			"1",
			"schema3-1.proto",
			true,
		},
		{
			"Proto3-valid-2",
			"valid3-2.pb",
			"2",
			"1",
			"schema3-1.proto",
			true,
		},
		{
			"Proto3-valid-3",
			"valid3-3.pb",
			"2",
			"1",
			"schema3-1.proto",
			true,
		},
		{
			"Proto3-valid-4",
			"valid3-4.pb",
			"2",
			"1",
			"schema3-1.proto",
			true,
		},
		{
			"Proto3-valid-5",
			"valid3-5.pb",
			"2",
			"1",
			"schema3-1.proto",
			true,
		},
		{
			"Proto3-valid-6",
			"valid3-6.pb",
			"2",
			"1",
			"schema3-1.proto",
			true,
		},
		{
			"Proto3-valid-7",
			"valid3-7.pb",
			"2",
			"1",
			"schema3-1.proto",
			true,
		},
		{
			"Proto3-invalid-1",
			"valid3-8.pb",
			"2",
			"1",
			"schema3-1.proto",
			false,
		},
		{
			"Proto3-valid-8",
			"valid3-8.pb",
			"3",
			"1",
			"schema3-2.proto",
			true,
		},
		{
			"Proto3-invalid-2",
			"valid3-8.pb",
			"2",
			"1",
			"schema3-1.proto",
			false,
		},
	}

	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	testdataDir := filepath.Join(basepath, "testdata")
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(testdataDir, tc.data))
			if err != nil {
				t.Errorf("schema read error: %s", err)
			}

			schema, err := os.ReadFile(filepath.Join(testdataDir, tc.schemaFilename))
			if err != nil {
				t.Errorf("schema read error: %s", err)
			}

			valid, err := v.Validate(data, schema, tc.schemaId, tc.schemaVersion)
			if err != nil {
				t.Error(err)
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

//nolint:deadcode,unused
func generateData() error {
	r1 := &testpb3.Record{
		Name:       "test1",
		Age:        1,
		Collection: []string{"1", "2", "3"},
		Foo:        &testpb3.Record_Foo{Bar: "bar1"},
	}
	data1, err := proto.Marshal(r1)
	if err != nil {
		return err
	}
	if err = saveToFile(data1, "testdata/valid3-1.pb"); err != nil {
		return err
	}

	r2 := &testpb3.Record{
		Name:       "test2",
		Collection: []string{"1", "2"},
		Foo:        &testpb3.Record_Foo{Bar: "bar2"},
	}
	data2, err := proto.Marshal(r2)
	if err != nil {
		return err
	}
	if err = saveToFile(data2, "testdata/valid3-2.pb"); err != nil {
		return err
	}

	r3 := &testpb3.Record{
		Name:       "test3",
		Collection: []string{"1", "3"},
		Foo:        &testpb3.Record_Foo{},
	}
	data3, err := proto.Marshal(r3)
	if err != nil {
		return err
	}
	if err = saveToFile(data3, "testdata/valid3-3.pb"); err != nil {
		return err
	}

	r4 := &testpb3.Record{
		Name:       "test4",
		Collection: []string{"1", "3"},
		Foo:        &testpb3.Record_Foo{},
	}
	data4, err := proto.Marshal(r4)
	if err != nil {
		return err
	}
	if err = saveToFile(data4, "testdata/valid3-4.pb"); err != nil {
		return err
	}

	r5 := &testpb3.Record{
		Name: "test5",
		Foo:  &testpb3.Record_Foo{},
	}
	data5, err := proto.Marshal(r5)
	if err != nil {
		return err
	}
	if err = saveToFile(data5, "testdata/valid3-5.pb"); err != nil {
		return err
	}

	r6 := &testpb3.Record{
		Collection: []string{"1", "3"},
		Foo:        &testpb3.Record_Foo{},
	}
	data6, err := proto.Marshal(r6)
	if err != nil {
		return err
	}
	if err = saveToFile(data6, "testdata/valid3-6.pb"); err != nil {
		return err
	}

	r7 := &testpb3.Record{
		Name:       "test7",
		Collection: []string{"1", "3"},
		Foo:        &testpb3.Record_Foo{},
	}
	data7, err := proto.Marshal(r7)
	if err != nil {
		return err
	}
	if err = saveToFile(data7, "testdata/valid3-7.pb"); err != nil {
		return err
	}

	p := &person.Person{
		Name:  "person",
		Id:    1,
		Email: "person@real.human",
		Phones: []*person.Person_PhoneNumber{
			{
				Number: "123456",
				Type:   person.Person_HOME,
			},
			{
				Number: "123457",
				Type:   person.Person_WORK,
			},
		},
		LastUpdated: timestamppb.Now(),
	}
	data8, err := proto.Marshal(p)
	if err != nil {
		return err
	}
	if err = saveToFile(data8, "testdata/valid3-8.pb"); err != nil {
		return err
	}

	return nil
}

//nolint:deadcode,unused
func saveToFile(data []byte, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	w := bufio.NewWriter(file)
	if _, err = w.Write(data); err != nil {
		return err
	}
	if err = w.Flush(); err != nil {
		return err
	}

	return file.Close()
}
