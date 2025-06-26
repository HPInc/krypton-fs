## Overview
fs local infra
- db
- local storage (for s3 like local storage)
- sqs (for file upload notifications)

### db
postgres db

### local storage
s3 like local storage is provided by minio.
For a local ui, use http://localhost:9000

### sqs
for each file uploaded to local store, a corresponding notification is sent to this local sqs.
sqs is configured to recieve notifications in a queue named `fs-notification`
Notification has the following format
```json
{
  "EventName": "s3:ObjectCreated:Put",
  "Key": "fs-test1/0",
  "Records": [
    {
      "eventVersion": "2.0",
      "eventSource": "minio:s3",
      "awsRegion": "",
      "eventTime": "2022-12-14T22:51:06.323Z",
      "eventName": "s3:ObjectCreated:Put",
      "userIdentity": {
        "principalId": "fstestadmin"
      },
      "requestParameters": {
        "accessKey": "fstestadmin",
        "region": "",
        "sourceIPAddress": "192.168.0.29"
      },
      "responseElements": {
        "x-amz-request-id": "1730CAA3BA70CEF7",
        "x-minio-deployment-id": "8c276d2e-fc01-4e79-a743-25da2cfe856b",
        "x-minio-origin-endpoint": "http://172.17.0.2:9000"
      },
      "s3": {
        "s3SchemaVersion": "1.0",
        "configurationId": "Config",
        "bucket": {
          "name": "fs-test1",
          "ownerIdentity": {
            "principalId": "fstestadmin"
          },
          "arn": "arn:aws:s3:::fs-test1"
        },
        "object": {
          "key": "0",
          "size": 11973,
          "eTag": "7441a60e1611fefab926afb415de4e99",
          "contentType": "application/octet-stream",
          "userMetadata": {
            "content-type": "application/octet-stream"
          },
          "versionId": "1",
          "sequencer": "1730CAA3BB35A1E8"
        }
      },
      "source": {
        "host": "192.168.0.29",
        "port": "",
        "userAgent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36"
      }
    }
  ]
}
```
