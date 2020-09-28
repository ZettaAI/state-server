package state

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"github.com/labstack/echo"
)

// SaveJSON compress and save neuroglancer JSON state.
func SaveJSON(c echo.Context) error {
	// Read request body
	reqBody := []byte{}
	if c.Request().Body != nil {
		reqBody, _ = ioutil.ReadAll(c.Request().Body)
	}
	// Reset
	// c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	n, err := zw.Write(reqBody)
	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if err := zw.Close(); err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	log.Printf("Raw state size: %v bytes.", n)
	uniqueID, err := writeToBucket(buf.Bytes())
	// uniqueID, err := writeToBucket(reqBody)

	if err != nil {
		log.Println(err)
		panic(err)
	}

	return c.JSON(
		http.StatusOK,
		fmt.Sprintf("localhost:8001/json/%v", uniqueID),
	)
}

// writeToBucket writes state to an object.
// Returns unique state ID.
func writeToBucket(data []byte) (string, error) {
	ctx := context.Background()
	id, err := GetUniqueObjectID()
	if err != nil {
		return "", fmt.Errorf("Creating unique ID: %v", err)
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	bkt := client.Bucket("state-server")
	obj := bkt.Object(id)

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
	return obj.ObjectName(), nil
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
