package state

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

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
	user := c.Request().Header.Get("X-Forwarded-User")
	if user == "" {
		user = c.Request().RemoteAddr
	}

	body, err := exctractAndParseJSONState(c)
	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// write compressed content into a unique object
	log.Printf("raw state size: %d bytes.", len(body))
	uniqueID, err := writeDataToBucket(body, os.Getenv("STATE_SERVER_BUCKET_GCS"), "", user)
	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(
		http.StatusOK,
		fmt.Sprintf(
			"%s/%s", utils.GetRequestSchemeAndHostURL(c)+JSONStateEP, uniqueID),
	)
}

// GetJSON return neuroglancer JSON state of given ID.
func GetJSON(c echo.Context) error {
	raw, err := readFromBucket(os.Getenv("STATE_SERVER_BUCKET_GCS"), c.Param("id"))
	if err != nil {
		log.Printf("readFromBucket: %v", err)
		return fmt.Errorf("readFromBucket: %v", err)
	}

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

	log.Printf("%v: %v", c.Param("id"), zr.Comment)
	jsonState := make(map[string]interface{})
	err = json.NewDecoder(zr).Decode(&jsonState)
	if err != nil {
		log.Printf("json.NewDecoder: %v", err)
		return fmt.Errorf("json.NewDecoder: %v", err)
	}

	parseRemoteLayers(jsonState["layers"])
	if err := zr.Close(); err != nil {
		log.Printf("gzip.NewReader.Close: %v", err)
		return fmt.Errorf("gzip.NewReader.Close: %v", err)
	}
	return c.JSON(http.StatusOK, jsonState)
}

// exctractAndParseJSONState extract state/JSON from request body
// parse state for remote layers and write them to bucket
func exctractAndParseJSONState(c echo.Context) ([]byte, error) {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return nil, err
	}

	var stateMap map[string]interface{}
	err = json.Unmarshal(body, &stateMap)
	if err != nil {
		return nil, err
	}
	parseRemoteLayers(stateMap["layers"])
	return body, nil
}
