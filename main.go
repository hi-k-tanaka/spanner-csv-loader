// Copyright hi-k-tanaka
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"cloud.google.com/go/spanner"
	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
)

type (
	loadOptions struct {
		comma            rune
		trimLeadingSpace bool
		lazyQuotes       bool
	}
)

func main() {
	source := flag.String("source", "gcs", "Set source type: gcs or csv filepath")
	bucket := flag.String("bucket", "", "Set Google Cloud Storage bucket name.")
	path := flag.String("path", "", "Set Google Cloud Storage object path.")
	spannerProject := flag.String("spanner-project-id", "", "Set Spanner project-id.")
	spannerInstance := flag.String("spanner-instance-id", "", "Set Spanner instance-id.")
	spannerDatabase := flag.String("spanner-database-id", "", "Set Spanner database-id.")
	spannerTable := flag.String("spanner-table", "", "Set Spanner table name that the csv will be loaded.")
	delimiter := flag.String("delimiter", "comma", "Set delimiter type: comma or tab.")
	lazyQuotes := flag.Bool("lazyquotes", true, "If true, use LazyQuotes.")
	trimLeadingSpace := flag.Bool("trimleadingspace", true, "If true, use TrimLeadingSpace.")

	flag.CommandLine.SetOutput(os.Stdout)
	flag.Parse()

	if err := run(*source, *bucket, *path, *spannerProject, *spannerInstance, *spannerDatabase, *spannerTable, *delimiter, *lazyQuotes, *trimLeadingSpace); err != nil {
		if os.Getenv("DEBUG") != "" {
			// show stacktrace
			log.Fatalf("error: %+v\n", err)
		} else {
			log.Fatalf("error: %s\n", err)
		}
	}
}

func run(source, bucket, path, spannerProject, spannerInstance, spannerDatabase, spannerTable, delimiter string, lazyQuotes, trimLeadingSpace bool) error {
	if source == "" {
		return errors.New("source is not set")
	}
	if spannerProject == "" {
		return errors.New("spanner-project-id is not set")
	}
	if spannerInstance == "" {
		return errors.New("spanner-instance-id is not set")
	}
	if spannerDatabase == "" {
		return errors.New("spanner-database-id is not set")
	}
	if spannerTable == "" {
		return errors.New("spanner-table is not set")
	}

	var comma = ','
	switch delimiter {
	case "comma":
		{
			comma = ','
		}
	case "tab":
		{
			comma = '\t'
		}
	default:
		return errors.New("invalid delimiter type. You can only use: comma or tab")
	}

	var reader io.Reader
	if source == "gcs" {
		if bucket == "" {
			return errors.New("bucket is not set")
		}
		if path == "" {
			return errors.New("path is not set")
		}
		ctx := context.Background()
		r, err := createGCSReader(ctx, bucket, path)
		if err != nil {
			return errors.Wrap(err, "error: createGCSReader")
		}
		reader = r
	} else {
		file, err := os.Open(source)
		if err != nil {
			return errors.Wrap(err, "error: os.Open")
		}
		reader = file
	}

	if err := load(reader, spannerProject, spannerInstance, spannerDatabase, spannerTable, &loadOptions{
		comma:            comma,
		lazyQuotes:       lazyQuotes,
		trimLeadingSpace: trimLeadingSpace,
	}); err != nil {
		return errors.Wrapf(err, "error: load")
	}
	return nil
}

func load(r io.Reader, spannerProjectID string, spannerInstanceID string, spannerDatabaseID string, spannerTable string, opt *loadOptions) error {
	ctx := context.Background()

	reader := csv.NewReader(r)
	reader.Comma = opt.comma
	reader.ReuseRecord = true
	reader.LazyQuotes = opt.lazyQuotes
	reader.TrimLeadingSpace = opt.trimLeadingSpace

	cli, err := spanner.NewClient(ctx, fmt.Sprintf("projects/%s/instances/%s/databases/%s", spannerProjectID, spannerInstanceID, spannerDatabaseID))
	if err != nil {
		return errors.Wrap(err, "failed NewClient")
	}

	colNum := 0
	columns := make([]string, 0)
	types := make([]string, 0)
	m := make([]*spanner.Mutation, 0)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if colNum == 0 {
			for _, v := range record {
				columns = append(columns, v)
			}
		} else if colNum == 1 {
			for _, v := range record {
				types = append(types, v)
			}
		} else {
			vals, err := makeVals(types, record)
			if err != nil {
				return errors.Wrap(err, "error makeVals")
			}
			m = append(m, spanner.InsertOrUpdate(spannerTable, columns, vals))
		}
		colNum++
	}
	_, err = cli.Apply(ctx, m)
	if err != nil {
		return errors.Wrap(err, "error Apply")
	}
	return nil
}

func makeVals(types []string, record []string) ([]interface{}, error) {
	vals := make([]interface{}, 0, len(record))
	for i, v := range record {
		switch types[i] {
		case "int64":
			{
				val, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return nil, errors.Wrap(err, "error ParseInt")
				}
				vals = append(vals, val)
			}
		case "float64":
			{
				val, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return nil, errors.Wrap(err, "error ParseFloat")
				}
				vals = append(vals, val)
			}
		case "bool":
			{
				val, err := strconv.ParseBool(v)
				if err != nil {
					return nil, errors.Wrap(err, "error ParseBool")
				}
				vals = append(vals, val)
			}
		case "date":
			{
				vals = append(vals, v)
			}
		case "timestamp":
			{
				vals = append(vals, v)
			}
		case "string":
			{
				vals = append(vals, v)
			}
		default:
			{
				return nil, errors.New(fmt.Sprintf("invalid spanner data type %s", types[i]))
			}
		}
	}
	return vals, nil
}

func createGCSReader(ctx context.Context, bucket string, path string) (io.Reader, error) {
	client, err := storage.NewClient(ctx)
	defer func() {
		err := client.Close()
		if err != nil {
			log.Fatalf("error: %s\n", err)
		}
	}()
	if err != nil {
		return nil, errors.Wrap(err, "failed storage.NewClient")
	}
	r, err := client.Bucket(bucket).Object(path).NewReader(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed NewReader")
	}
	return r, nil
}
