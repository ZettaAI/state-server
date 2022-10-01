package state

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// RemoteLayer layer to be persisted
type RemoteLayer struct {
	name string
	body map[string]interface{}
}

// parseRemoteLayers iterates through layers and
// identifies which ones need to be read/written remotely.
func parseRemoteLayers(layersObject interface{}) {
	layers := layersObject.([]interface{})
	for _, layer := range layers {
		layerMap := layer.(map[string]interface{})
		remoteLayer := RemoteLayer{layerMap["name"].(string), layerMap}
		err := remoteLayer.save()
		if err != nil {
			log.Println(err)
		}
	}
}

func (l *RemoteLayer) save() error {
	bytes, err := json.Marshal(l.body)
	if err != nil {
		return err
	}
	fmt.Printf("layer name: %s, size: %d\n", l.name, len(bytes))
	_, err = writeDataToBucket(bytes, os.Getenv("REMOTE_LAYERS_BUCKET"), l.name, "")
	if err != nil {
		return err
	}
	return nil
}
