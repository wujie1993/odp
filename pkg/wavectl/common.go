package wavectl

import (
	"encoding/json"

	"github.com/ghodss/yaml"
)

const (
	OutputFormatJSON       = "json"
	OutputFormatJSONPretty = "json-pretty"
	OutputFormatYAML       = "yaml"
	OutputFormatTable      = "table"
)

func ToJSON(obj interface{}, pretty bool) ([]byte, error) {
	if pretty {
		return json.MarshalIndent(obj, "", "\t")
	}
	return json.Marshal(obj)
}

func ToYAML(obj interface{}) ([]byte, error) {
	return yaml.Marshal(obj)
}
