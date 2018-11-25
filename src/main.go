package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
)

func main() {
	bucket := flag.String("bucket", "", "Set Google Cloud Storage bucket name.")
	object := flag.String("object", "", "Set Google Cloud Storage object name.")

	flag.CommandLine.SetOutput(os.Stdout)
	flag.Parse()

	if err := run(*bucket, *object); err != nil {
		log.Fatalf("error: %+v\n", err)
	}
}

func run(bucket string, object string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	defer client.Close()
	if err != nil {
		return errors.Wrap(err, "failed storage.NewClient")
	}

	r, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return errors.Wrap(err, "failed NewReader")
	}
	reader := csv.NewReader(r)
	reader.Comma = '\t'

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		fmt.Printf("%+v %d\n", record, len(record))
	}
	return nil
}
