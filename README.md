# spanner-csv-loader

spanner-csv-loader reads csv from Google Cloud Storage or the local file system and imports that data into Google Cloud Spanner.

# Installation
```sh
go get -u github.com/hi-k-tanaka/spanner-csv-loader
```

# Login
spanner-csv-loader uses Google Cloud Storage and Google Cloud Spanner. You have to login to Google Cloud Platform to access them.
You have to install [Google Cloud SDK](https://cloud.google.com/sdk/install) before do this.

```sh
gcloud auth application-default login
```

# CSV format
You can use the general format to CSV file that you want to create. However, the first and the second line is a little bit special. 

### First line
Each value in the first line must be a column name of the spanner table to imported.

### Second line
Each value in the second line must be a type name of the spanner table's column.

Supported spanner column types are below. [See also](https://cloud.google.com/spanner/docs/data-definition-language#data_types)
```
int64, string, float64, bool, date, timestamp
```

### Example
[example.sql](https://github.com/hi-k-tanaka/spanner-csv-loader/blob/master/examples/example.sql)
```
CREATE TABLE Students (
  StudentId INT64 NOT NULL,
  Name STRING(MAX) NOT NULL,
  Score INT64 NOT NULL,
  Average FLOAT64 NOT NULL,
  Valid BOOL NOT NULL,
  CreatedAt DATE NOT NULL,
  UpdatedAt TIMESTAMP NOT NULL
) PRIMARY KEY(StudentId);
```


[example.csv](https://github.com/hi-k-tanaka/spanner-csv-loader/blob/master/examples/example.csv)
```csv
StudentId,Name,Score,Average,Valid,CreatedAt,UpdatedAt
int64,string,int64,float64,bool,date,timestamp
1,Mark,180,150.3,true,2018-11-12,2014-09-27T12:30:00.45Z
```

# How to use
```sh
spanner-csv-loader -h
```

# Run
Load CSV from Google Cloud Storage
```sh
spanner-csv-loader --bucket=your-gcs-bucket --path=your-csv-path-on-gcs --spanner-project-id=your-gcp-project --spanner-instance-id=your-spanner-instance --spanner-database-id=your-spanner-database --spanner-table=your-spanner-table
```

Load CSV from local file
```sh
spanner-csv-loader --source=examples/example.csv --spanner-project-id=your-gcp-project --spanner-instance-id=your-spanner-instance --spanner-database-id=your-spanner-database --spanner-table=your-spanner-table
```
