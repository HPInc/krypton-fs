# File Service

## Purpose

Files service is a supporting service for a multi-tenanted
device management system. Files service will provide core
functionality by returning time limited signed urls under
a {tenant}/{device} path using S3. Devices can use files service
provided signed urls to upload or download content such as logs
(upload to server) or other operational payloads such as updates
(download from server).

By providing signed urls to a cloud storage system like S3,
files service detaches itself from the data flow path.
This allows clients to engage directly with cloud storage
apis for transfer of content.

## Control flow

1.	Device requests a download or upload url
	a.	Files service validates device token via device sts public keys (1.1)
2.	Files Service does metadata lookup for requested operation
3.	Files Service creates required signed url and returns it to caller
	a.	url is signed for method (GET, PUT) and shasum of expected content
4.	Caller does file operation with http multipart put/get directly to cloud storage
5.	Upon file uploads, files service gets notified and publishes a notification for internal consumption
6.	Services interested in file uploads will subscribe to the notification and act accordingly

## Device pre-requisites

-	Device must have completed enrollment. This will allow devices to get a token
from DSTS (device security token service)
-	Device must present a valid access token to request upload/download urls

# Files service functions

-	Verifies token with DSTS public key
-	Verify permissions for requested path
	o	Tenant root
	o	Device root
-	Create signed upload or download url with the following attributes
	o	method (PUT, GET)
	o	file shasum
-	Metrics
	o	API level metrics
	o	File operations per tenant
	o	File operations per device
	o	Cache metrics
-	Postgres metadata database
-	Redis cache with write through setup

## File entry in metadata db

- ID, (db unique uuid)
- TenantId, (device token claims)
- DeviceId, (device token claims)
- FileShaSum, (caller should provide, files service will add this to signed url)
- FileName (name of the file as defined by caller)

## File entry in Storage (S3)

Bucket/Tenant/Device/File -> not descriptive names, but guids for path so we can scale
Concerns for scalability. See - https://docs.aws.amazon.com/AmazonS3/latest/userguide/optimizing-performance.html

# References

- Signed url - https://docs.aws.amazon.com/AmazonS3/latest/userguide/ShareObjectPreSignedURL.html
- Local test – S3 clone - https://min.io/
