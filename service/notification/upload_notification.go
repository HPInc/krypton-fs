// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package notification

type UploadNotification struct {
	ReceiptHandle string   `json:"-"`
	Records       []Record `json:"records"`
}

type Record struct {
	Storage struct {
		Bucket struct {
			Name string `json:"name"`
		} `json:"bucket"`
		Object struct {
			Key  string `json:"key"`
			Size int64  `json:"size"`
		} `json:"object"`
	} `json:"s3"`
	ScanStatus string `json:"scan_status"`
}
