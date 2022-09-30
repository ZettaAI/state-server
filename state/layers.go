package state

import "fmt"

// type remoteLayer struct {
// 	name string
// 	body map[string]interface{}
// }

// parseRemoteLayers iterates through layers and
// identifies which ones need to be read/written remotely.
func parseRemoteLayers(layersObject interface{}) {
	layers := layersObject.([]interface{})
	for i, layer := range layers {
		layerMap := layer.(map[string]interface{})
		fmt.Printf("%d layer name: %s\n", i, layerMap["name"])
	}
}

// func (v *remoteLayer) save() {

// }
