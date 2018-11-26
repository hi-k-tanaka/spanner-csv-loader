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
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/spanner/admin/database/apiv1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	adminpb "google.golang.org/genproto/googleapis/spanner/admin/database/v1"
)

func Test_run(t *testing.T) {
	// spannerProject not found
	{
		err := run("", "", "", "", "", "", "", "", false, false)
		require.Error(t, err)
		assert.Equal(t, "source is not set", err.Error())
	}

	// spannerProject not found
	{
		err := run("gcs", "", "", "", "", "", "", "", false, false)
		require.Error(t, err)
		assert.Equal(t, "spanner-project-id is not set", err.Error())
	}

	// spannerInstance not found
	{
		err := run("gcs", "", "", "project", "", "", "", "", false, false)
		require.Error(t, err)
		assert.Equal(t, "spanner-instance-id is not set", err.Error())
	}

	// spannerDatabase not found
	{
		err := run("gcs", "", "", "project", "instance", "", "", "", false, false)
		require.Error(t, err)
		assert.Equal(t, "spanner-database-id is not set", err.Error())
	}

	// spannerTable not found
	{
		err := run("gcs", "", "", "project", "instance", "database", "", "", false, false)
		require.Error(t, err)
		assert.Equal(t, "spanner-table is not set", err.Error())
	}

	// delimiter is invalid
	{
		err := run("gcs", "", "", "project", "instance", "database", "table", "", false, false)
		require.Error(t, err)
		assert.Equal(t, "invalid delimiter type. You can only use: comma or tab", err.Error())
	}

	// bucket not found
	{
		err := run("gcs", "", "", "project", "instance", "database", "table", "comma", false, false)
		require.Error(t, err)
		assert.Equal(t, "bucket is not set", err.Error())
	}

	// path not found
	{
		err := run("gcs", "test", "", "project", "instance", "database", "table", "comma", false, false)
		require.Error(t, err)
		assert.Equal(t, "path is not set", err.Error())
	}

	// gcs not found
	{
		err := run("gcs", "test", "path", "project", "instance", "database", "table", "comma", false, false)
		require.Error(t, err)
		assert.True(t, strings.Index(err.Error(), "error: createGCSReader") == 0)
	}

	// invalid csv file not found
	{
		err := run("notfound.csv", "test", "path", "project", "instance", "database", "table", "comma", false, false)
		require.Error(t, err)
		assert.True(t, strings.Index(err.Error(), "error: os.Open") == 0)
	}

	// load error
	{
		err := run("examples/example.csv", "", "", "project", "instance", "database", "table", "tab", false, false)
		require.Error(t, err)
		assert.True(t, strings.Index(err.Error(), "error: load") == 0)
	}
}

func Test_load(t *testing.T) {
	ctx := context.Background()

	project := os.Getenv("TEST_SPANNER_PROJECT_ID")
	instance := os.Getenv("TEST_SPANNER_INSTANCE_ID")
	dbID := "for-test-" + strconv.FormatInt(time.Now().Unix(), 10)

	adminClient, err := database.NewDatabaseAdminClient(ctx)
	require.NoError(t, err)

	op, err := adminClient.CreateDatabase(ctx, &adminpb.CreateDatabaseRequest{
		Parent:          fmt.Sprintf("projects/%s/instances/%s", project, instance),
		CreateStatement: "CREATE DATABASE `" + dbID + "`",
		ExtraStatements: []string{
			`CREATE TABLE Testing (
                                ColInt64 INT64 NOT NULL,
                                ColString STRING(MAX) NOT NULL,
                                ColFloat64 FLOAT64 NOT NULL,
                                ColBool BOOL NOT NULL,
                                ColDate DATE NOT NULL,
                                ColTimestamp Timestamp NOT NULL,
                        ) PRIMARY KEY (ColInt64)`,
		},
	})
	defer func() {
		err := adminClient.DropDatabase(ctx, &adminpb.DropDatabaseRequest{
			Database: fmt.Sprintf("projects/%s/instances/%s/databases/%s", project, instance, dbID),
		})
		require.NoError(t, err)
	}()
	require.NoError(t, err)

	_, err = op.Wait(ctx)
	require.NoError(t, err)

	st := `ColInt64,ColString,ColFloat64,ColBool,ColDate,ColTimestamp
int64,string,float64,bool,date,timestamp
1,foo,1.23,true,2018-02-03,2014-09-27T12:30:00.45Z`

	r := strings.NewReader(st)

	err = load(r, project, instance, dbID, "Testing", &loadOptions{
		comma:            ',',
		trimLeadingSpace: true,
		lazyQuotes:       true,
	})
	require.NoError(t, err)
}

func Test_makeVals(t *testing.T) {
	// success
	{
		types := []string{
			"int64", "float64", "string", "bool", "bool", "date", "timestamp",
		}
		record := []string{
			"1", "1.23", "foo", "false", "true", "2018-11-15", "2014-09-27T12:30:00.45Z",
		}

		r, err := makeVals(types, record)
		require.NoError(t, err)
		require.NotNil(t, r)

		assert.Equal(t, int64(1), r[0])
		assert.Equal(t, float64(1.23), r[1])
		assert.Equal(t, "foo", r[2])
		assert.Equal(t, false, r[3])
		assert.Equal(t, true, r[4])
		assert.Equal(t, "2018-11-15", r[5])
		assert.Equal(t, "2014-09-27T12:30:00.45Z", r[6])
	}

	// error invalid type
	{
		types := []string{
			"int32",
		}
		record := []string{
			"1",
		}

		r, err := makeVals(types, record)
		require.Error(t, err)
		assert.Nil(t, r)
	}

	// error invalid value
	{
		types := []string{
			"int64",
		}
		record := []string{
			"foo",
		}

		r, err := makeVals(types, record)
		require.Error(t, err)
		assert.Nil(t, r)
	}

	// error invalid value
	{
		types := []string{
			"float64",
		}
		record := []string{
			"foo",
		}

		r, err := makeVals(types, record)
		require.Error(t, err)
		assert.Nil(t, r)
	}

	// error invalid value
	{
		types := []string{
			"bool",
		}
		record := []string{
			"foo",
		}

		r, err := makeVals(types, record)
		require.Error(t, err)
		assert.Nil(t, r)
	}
}

func Test_createGCSReader(t *testing.T) {
	ctx := context.Background()
	bucket := os.Getenv("TEST_GCS_BUCKET")
	path := os.Getenv("TEST_GCS_PATH")

	r, err := createGCSReader(ctx, bucket, path)
	require.NoError(t, err)
	require.NotNil(t, r)
}
