package state

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

const (
	CREATE = "CREATE:"
	READ   = "READ:"
	UPDATE = "UPDATE:"
	DELETE = "DELETE:"
)

// RemoteLayer layer to be persisted
type RemoteLayer struct {
	name string
	body map[string]interface{}
}

// runLayerActions iterates through layers and perform action.
func runLayerActions(layersObject interface{}) []map[string]interface{} {
	var actionsRan bool
	layers := layersObject.([]interface{})
	var layers2 = make([]map[string]interface{}, len(layers))
	for idx, layer := range layers {
		layerMap := layer.(map[string]interface{})
		remoteLayer := RemoteLayer{layerMap["name"].(string), layerMap}
		ran, err := remoteLayer.runAction()
		actionsRan = actionsRan || ran
		if err != nil {
			log.Println(err)
		}
		layerMap["name"] = remoteLayer.name
		layers2[idx] = layerMap
	}
	return layers2
}

func (layer *RemoteLayer) runAction() (bool, error) {
	switch {
	case strings.HasPrefix(layer.name, CREATE):
		layer.name = layer.name[len(CREATE):]
		return true, layer.create()
	case strings.HasPrefix(layer.name, READ):
		layer.name = layer.name[len(READ):]
		return true, layer.read()
	case strings.HasPrefix(layer.name, UPDATE):
		layer.name = layer.name[len(UPDATE):]
		return true, layer.update()
	case strings.HasPrefix(layer.name, DELETE):
		layer.name = layer.name[len(DELETE):]
		return true, layer.delete()
	default:
		return false, nil
	}
}

func (layer *RemoteLayer) create() error {
	bytes, err := json.Marshal(layer.body)
	if err != nil {
		return err
	}
	_, err = writeObject(
		bytes,
		os.Getenv("REMOTE_ANNOTATIONS_BUCKET"),
		layer.name,
		true,
	)
	return err
}

func (layer *RemoteLayer) read() error {
	bytes, err := readObject(
		os.Getenv("REMOTE_ANNOTATIONS_BUCKET"),
		layer.name,
	)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, &layer.body)
}

func (layer *RemoteLayer) update() error {
	bytes, err := json.Marshal(layer.body)
	if err != nil {
		return err
	}
	_, err = writeObject(
		bytes,
		os.Getenv("REMOTE_ANNOTATIONS_BUCKET"),
		layer.name,
		false,
	)
	return err
}

func (layer *RemoteLayer) delete() error {
	return nil
}
