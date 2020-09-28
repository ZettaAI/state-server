package state

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/storage"
)

// GetUniqueObjectID returns a unique object ID (within a bucket).
func GetUniqueObjectID() (string, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	bkt := client.Bucket("state-server")

	id, _ := generateRandomString(14)
	obj := bkt.Object(fmt.Sprintf("states/%v", id))
	r, err := obj.NewReader(ctx)

	// repeat until ErrObjectNotExist is returned
	for err == nil {
		defer r.Close()
		log.Println("Retrying with new state ID.")
		id, _ := generateRandomString(14)
		obj = bkt.Object(fmt.Sprintf("states/%v", id))
		r, err = obj.NewReader(ctx)
	}
	log.Printf("Using unique ID %v\n", id)
	return obj.ObjectName(), nil
}
