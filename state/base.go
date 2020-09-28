package state

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"github.com/akhileshh/state-server/utils"
	"github.com/labstack/echo"
)

const (
	// JSONStateEP retreive endpoint
	JSONStateEP = "/json"
	// JSONStatePostEP create endpoint
	JSONStatePostEP = "/json/post"
)

// SaveJSON compress and save neuroglancer JSON state.
// Currently supports saving states in GCS buckets.
func SaveJSON(c echo.Context) error {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)

	user := c.Request().Header.Get("X-Forwarded-User")
	if user == "" {
		user = c.Request().RemoteAddr
	}
	zw.Comment = fmt.Sprintf("Generated by user: %v", user)

	// copy from request body to compressed buffer
	n, err := io.Copy(zw, c.Request().Body)
	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if err := zw.Close(); err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// write compressed content into a unique object
	log.Printf("Raw state size: %v bytes.", n)
	uniqueID, err := writeToBucket(buf.Bytes())
	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(
		http.StatusOK,
		fmt.Sprintf(
			"%v/%v", utils.GetRequestSchemeAndHostURL(c)+JSONStateEP, uniqueID),
	)
}

// writeToBucket writes state to an object.
// Returns unique state ID.
func writeToBucket(data []byte) (string, error) {
	ctx := context.Background()
	id, key, err := GetUniqueObjectID()
	if err != nil {
		return "", fmt.Errorf("Creating unique ID: %v", err)
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	bkt := client.Bucket("state-server")
	obj := bkt.Object(key)

	log.Printf("Creating object %v\n", obj.ObjectName())

	w := obj.NewWriter(ctx)
	n, err := fmt.Fprint(w, base64.StdEncoding.EncodeToString(data))
	if err != nil {
		return "", fmt.Errorf(
			"Error writing to object %v: %v", obj.ObjectName(), err)
	}
	log.Printf("Compressed state size: %v bytes.", n)

	// Close immediately to confirm object has been created.
	if err := w.Close(); err != nil {
		return "", fmt.Errorf(
			"Error writing to object %v: %v", obj.ObjectName(), err)
	}

	log.Printf("Created object %v\n", obj.ObjectName())
	return id, nil
}

// GetJSON return neuroglancer JSON state of given ID.
func GetJSON(c echo.Context) error {
	ctx := context.Background()
	objectName := fmt.Sprintf("states/%v", c.Param("id"))

	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	rc, err := client.Bucket("state-server").Object(objectName).NewReader(ctx)
	if err != nil {
		return fmt.Errorf("Object(%q).NewReader: %v", objectName, err)
	}
	defer rc.Close()

	log.Printf("Read object %v\n", objectName)
	raw, err := ioutil.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("ioutil.ReadAll: %v", err)
	}
	log.Printf("Read object %v complete, size %v.\n", objectName, len(raw))

	compressed, err := base64.StdEncoding.DecodeString(string(raw))
	if err != nil {
		log.Printf("base64.StdEncoding.DecodeString: %v", err)
		return fmt.Errorf("base64.StdEncoding.DecodeString: %v", err)
	}

	zr, err := gzip.NewReader(bytes.NewBuffer(compressed))
	if err != nil {
		log.Printf("gzip.NewReader: %v", err)
		return fmt.Errorf("gzip.NewReader: %v", err)
	}

	log.Println(zr.Comment)

	jsonState := make(map[string]interface{})
	err = json.NewDecoder(zr).Decode(&jsonState)
	if err != nil {
		log.Printf("json.NewDecoder: %v", err)
		return fmt.Errorf("json.NewDecoder: %v", err)
	}

	if err := zr.Close(); err != nil {
		log.Fatal(err)
	}
	log.Println(jsonState)

	return c.JSON(http.StatusOK, jsonState)
}
