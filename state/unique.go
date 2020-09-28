package state

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/storage"
	"github.com/akhileshh/state-server/utils"
)

// GetUniqueObjectID creates a unique object ID (within a bucket).
// Returns the ID, full key and error, if any.
func GetUniqueObjectID(bucket string) (string, string, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", "", fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	bkt := client.Bucket(bucket)
	id, _ := utils.GenerateRandomString(12)
	obj := bkt.Object(fmt.Sprintf("states/%v", id))
	r, err := obj.NewReader(ctx)

	// repeat until ErrObjectNotExist is returned
	for err == nil {
		defer r.Close()
		log.Println("Retrying with new state ID.")
		id, _ := utils.GenerateRandomString(12)
		obj = bkt.Object(fmt.Sprintf("states/%v", id))
		r, err = obj.NewReader(ctx)
	}
	log.Printf("Using unique ID %v\n", id)
	return id, obj.ObjectName(), nil
}
