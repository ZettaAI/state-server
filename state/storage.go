package state

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"time"

	"cloud.google.com/go/storage"
)

// writeDataToBucket writes bytes to a bucket object.
// Returns unique state ID.
func writeDataToBucket(data []byte, bucket, key, user string) (string, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if user != "" {
		zw.Comment = fmt.Sprintf("Generated by user: %s", user)
	}

	// copy from data to compressed buffer
	_, err := io.Copy(zw, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	if err := zw.Close(); err != nil {
		return "", err
	}

	var id string
	if key == "" {
		id, key, err = GetUniqueObjectID(bucket)
		if err != nil {
			return "", fmt.Errorf("creating unique ID: %v", err)
		}
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	obj := client.Bucket(bucket).Object(key)
	w := obj.NewWriter(ctx)
	_, err = fmt.Fprint(w, base64.StdEncoding.EncodeToString(data))
	if err != nil {
		return "", fmt.Errorf(
			"error writing to object %v: %v", obj.ObjectName(), err)
	}

	// close immediately to confirm object has been created.
	if err := w.Close(); err != nil {
		return "", fmt.Errorf(
			"at Close(): error writing to object %v: %v", obj.ObjectName(), err)
	}
	log.Printf("created object %v\n", obj.ObjectName())
	return id, nil
}

// readFromBucket reads JSON state and returns raw bytes.
func readFromBucket(bucket, id string) ([]byte, error) {
	ctx := context.Background()
	ctxTO, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	objectName := fmt.Sprintf("states/%v", id)
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	rc, err := client.Bucket(bucket).Object(objectName).NewReader(ctxTO)
	if err != nil {
		return nil, fmt.Errorf("object(%q).NewReader: %v", objectName, err)
	}
	defer rc.Close()

	raw, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %v", err)
	}
	log.Printf("read object %v complete, size %v.\n", objectName, len(raw))
	return raw, nil
}
