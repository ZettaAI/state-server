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

	data, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}
	state, err := parseStatesAndRunActions(data)
	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	data, err = json.Marshal(state)
	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	uniqueID, err := writeDataToBucket(data, os.Getenv("STATE_SERVER_BUCKET_GCS"), user)
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
	raw, err := readObject(
		os.Getenv("STATE_SERVER_BUCKET_GCS"), fmt.Sprintf("states/%s", c.Param("id")))
	if err != nil {
		err = fmt.Errorf("readObject: %v", err)
		log.Println(err)
		return err
	}

	data, err := base64.StdEncoding.DecodeString(string(raw))
	if err != nil {
		data = raw
	}

	zr, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("gzip.NewReader: %v", err)
	}

	log.Printf("%v: %v", c.Param("id"), zr.Comment)
	jsonState := make(map[string]interface{})

	err = json.NewDecoder(zr).Decode(&jsonState)
	if err != nil {
		return fmt.Errorf("json.NewDecoder: %v", err)
	}

	if err := zr.Close(); err != nil {
		err = fmt.Errorf("gzip.NewReader.Close: %v", err)
		log.Println(err)
		return err
	}
	return c.JSON(http.StatusOK, jsonState)
}

// parseStatesAndRunActions extract state/JSON from request body
// parse state for remote layers and write them to bucket
func parseStatesAndRunActions(body []byte) (map[string]interface{}, error) {
	var stateMap map[string]interface{}
	err := json.Unmarshal(body, &stateMap)
	if err != nil {
		return nil, err
	}
	stateMap["layers"] = runLayerActions(stateMap["layers"])
	return stateMap, nil
}
