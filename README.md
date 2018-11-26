# spanner-csv-loader

spanner-csv-loader reads csv from Google Cloud Storage or the local file system and imports that data into Google Cloud Spanner.

# Installation
```sh
go get -u github.com/hi-k-tanaka/spanner-csv-loader
```

# Login
spanner-csv-loader uses Google Cloud Storage and Google Cloud Spanner. You have to login to Google Cloud Platform to access them.

```sh
gcloud auth application-default login
```

# CSV format
You can use the general format to CSV file that you want to create. However, The first and the second line is a little bit special. 

### First line
Each value in the first line must be a column name of the spanner table to imported.

### Second line
Each value in the second line must be a type name of the spanner table's column.

# How to use
```sh
spanner-csv-loader -h
```

# Run (load from Google Cloud Storage)
```sh
spanner-csv-loader --bucket=your-gcs-bucket --path=your-csv-path-on-gcs --spanner-project-id=your-gcp-project --spanner-instance-id=your-spanner-instance --spanner-database-id=your-spanner-database --spanner-table=your-spanner-table
```

# Run (load from local file)
```sh
spanner-csv-loader --source=examples/example.csv --spanner-project-id=your-gcp-project --spanner-instance-id=your-spanner-instance --spanner-database-id=your-spanner-database --spanner-table=your-spanner-table
```
