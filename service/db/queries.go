package db

// Database queries.
const (
	// Bucket lifecycle management queries
	queryInsertNewBucket = `INSERT INTO buckets(bucket_name,is_archived,
		created_at,updated_at) VALUES($1,$2,now(),now())`

	queryGetBucketByName = `SELECT bucket_name,is_archived,created_at,
	updated_at FROM buckets WHERE buckets.bucket_name=$1`

	queryGetEnabledBuckets = "SELECT * FROM buckets WHERE buckets.is_archived=false"

	queryArchiveBucket = `UPDATE buckets SET buckets.updated_at=now(),
	buckets.is_archived=false WHERE buckets.bucket_name=$1`

	// File lifecycle management queries
	queryInsertNewFile = `INSERT INTO files(tenant_id,device_id,name,checksum,
		size,status,created_at,updated_at,bucket_name) 
		VALUES($1,$2,$3,$4,$5,$6,now(),now(),$7)
		RETURNING file_id,tenant_id,device_id,name,checksum,size,status,
		created_at,updated_at,bucket_name`

	queryFileByID = `SELECT file_id,tenant_id,device_id,name,checksum,size,status,
	created_at,updated_at,bucket_name FROM files WHERE files.file_id=$1`

	queryUpdateFileStatus = `UPDATE files SET updated_at=now(), size=$2, status=$3 
	WHERE file_id=$1 RETURNING file_id`

	queryFilesForSpecificDevice = `SELECT * FROM files WHERE files.tenant_id=$1 and 
	files.device_id=$2`

	queryDeleteExpiredFiles = "DELETE FROM files WHERE files.created_at <= $1 LIMIT 100"

	deleteFileByID = `DELETE FROM files WHERE files.file_id=$1`
)
