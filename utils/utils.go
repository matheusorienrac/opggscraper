package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/matheusorienrac/opggscraper/model"
)

// Unmarshalls champion data to JSON and saves it to a file
func SaveJSON(champion *model.Champion, filename string) error {

	// Marshal the map into JSON
	jsonData, err := json.Marshal(champion)
	if err != nil {
		return err
	}

	// Save the JSON data to a file
	err = ioutil.WriteFile(filename+".json", jsonData, 0644)
	if err != nil {
		return err
	}

	fmt.Println("JSON data saved to " + filename + ".json")

	return nil
}
