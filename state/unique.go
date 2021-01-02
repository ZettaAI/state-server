package state

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"

	"cloud.google.com/go/storage"
)

const (
	idSpaceLow  = 100000000000000000 // 18 digits
	idSpaceHigh = 999999999999999999
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

	// id, _ := utils.GenerateRandomString(12)
	id := idSpaceLow + rand.Intn(idSpaceHigh-idSpaceLow)

	bkt := client.Bucket(bucket)
	obj := bkt.Object(fmt.Sprintf("states/%d", id))
	r, err := obj.NewReader(ctx)

	// repeat until ErrObjectNotExist is returned
	for err == nil {
		defer r.Close()
		log.Println("Retrying with new state ID.")
		id := idSpaceLow + rand.Intn(idSpaceHigh-idSpaceLow)
		obj = bkt.Object(fmt.Sprintf("states/%d", id))
		r, err = obj.NewReader(ctx)
	}
	log.Printf("Using unique ID %v\n", id)
	return strconv.Itoa(id), obj.ObjectName(), nil
}
