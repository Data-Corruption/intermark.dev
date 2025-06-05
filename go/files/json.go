package files

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
)

// LoadJSON loads the JSON from the given file path into the given object.
// Use os.IsNotExist(err) on the returned error to handle non-existence case.
func LoadJSON(path string, obj any) error {
	if _, err := os.Stat(path); err != nil {
		return err
	}
	file, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("issue reading json file: %w", err)
	}
	if err := json.Unmarshal(file, obj); err != nil {
		return fmt.Errorf("issue unmarshalling json file: %w", err)
	}
	return nil
}

// SaveJSON saves the object to the given file path with the given permissions.
func SaveJSON(path string, obj any, perms fs.FileMode) error {
	data, err := json.MarshalIndent(obj, "", "  ") // with indentation
	if err != nil {
		return fmt.Errorf("error encoding obj: %w", err)
	}
	if err = os.WriteFile(path, data, perms); err != nil {
		return fmt.Errorf("error writing obj file: %w", err)
	}
	return nil
}
