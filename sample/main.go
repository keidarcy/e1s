package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

const filename = "data.json"

type Data struct {
	// Your JSON data structure here
	ExampleField string `json:"example_field"`
}

func main() {
	// Load the JSON data from the file or create a new instance if the file doesn't exist.
	var data Data
	if _, err := os.Stat(filename); err == nil {
		file, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Println("Error reading file:", err)
			os.Exit(1)
		}
		if err := json.Unmarshal(file, &data); err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			os.Exit(1)
		}
	} else {
		data = Data{
			ExampleField: "Initial value",
		}
	}

	// Save the JSON data to a temporary file.
	tmpfile, err := ioutil.TempFile("", "data.json")
	if err != nil {
		fmt.Println("Error creating temporary file:", err)
		os.Exit(1)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		os.Exit(1)
	}

	if _, err := tmpfile.Write(jsonData); err != nil {
		fmt.Println("Error writing to temporary file:", err)
		os.Exit(1)
	}

	// Open the vi editor to allow the user to modify the JSON data.
	editor := os.Getenv("EDITOR")
	cmd := exec.Command(editor, tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("Error opening editor:", err)
		os.Exit(1)
	}

	// Reload the JSON data after editing.
	file, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		fmt.Println("Error reading temporary file:", err)
		os.Exit(1)
	}
	if err := json.Unmarshal(file, &data); err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		os.Exit(1)
	}

	// You can now use the 'data' variable with the updated JSON data.
	fmt.Println("Modified JSON Data:")
	fmt.Println(data)
}
