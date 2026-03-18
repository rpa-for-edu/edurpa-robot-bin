package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

// Connection represents a connection object

type Credential struct {
	FileName string `json:"fileName"`
	Data     struct {
		AccessToken  string   `json:"access_token"`
		RefreshToken string   `json:"refresh_token"`
		ClientID     string   `json:"client_id"`
		ClientSecret string   `json:"client_secret"`
		TokenURI     string   `json:"token_uri"`
		Scopes       []string `json:"scopes"`
		BaseURL      string   `json:"base_url"`
	} `json:"data"`
}

func main() {

	err := checkRequiredVariables()

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Call the API and process the response
	connections, err := callAPIAndProcessResponse()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Create a folder with the specified name
	credentialFolderName := os.Getenv("ROBOT_CREDENTIAL_FOLDER")

	if _, err := os.Stat(credentialFolderName); os.IsNotExist(err) {
		// Directory does not exist, create it
		err := os.Mkdir(credentialFolderName, 0755)
		if err != nil {
			fmt.Printf("Error creating directory: %v\n", err)
			return
		}
		fmt.Println("Directory created successfully.")
	} else if err != nil {
		// Some other error occurred while checking directory existence
		fmt.Printf("Error checking directory existence: %v\n", err)
		return
	} else {
		// Directory already exists
		fmt.Println("Directory already exists.")
	}

	fmt.Println("Folder created successfully:", credentialFolderName)

	// Create JSON file
	for _, conn := range connections {
		jsonBytes, err := json.MarshalIndent(conn.Data, "", "    ")
		if err != nil {
			fmt.Println("Error marshalling JSON:", err)
			return
		}
		err = ioutil.WriteFile(credentialFolderName+"/"+conn.FileName, jsonBytes, 0644)
		if err != nil {
			fmt.Println("Error writing JSON file:", err)
			return
		}

		fmt.Println("JSON file created successfully: %s", conn.FileName)
	}
}

func callAPIAndProcessResponse() ([]Credential, error) {
	endpoint := fmt.Sprintf("%s/connection/for-robot/version", os.Getenv("MAIN_SERVER_API"))
	serviceKey := os.Getenv("SERVICE_KEY")

	userID, err := strconv.Atoi(os.Getenv("USER_ID"))
	if err != nil {
		return nil, fmt.Errorf("failed to convert USER_ID to integer: %w", err)
	}

	// Create request body
	requestData := map[string]interface{}{
		"userId":         userID,
		"processId":      os.Getenv("PROCESS_ID"),
		"processVersion": os.Getenv("PROCESS_VERSION"),
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}

	// Make HTTP POST request
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Service-Key", serviceKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API request failed with status code: %d, body: %s", resp.StatusCode, string(responseBody))
	}

	// Unmarshal response JSON
	var connections []Credential
	err = json.Unmarshal(responseBody, &connections)
	if err != nil {
		return nil, err
	}

	return connections, nil
}

func checkRequiredVariables() error {
	requiredVariables := []string{
		"ROBOT_CREDENTIAL_FOLDER",
		"MAIN_SERVER_API",
		"SERVICE_KEY",
		"USER_ID",
		"PROCESS_ID",
		"PROCESS_VERSION",
	}

	for _, variable := range requiredVariables {
		if value := os.Getenv(variable); value == "" {
			fmt.Print(variable, requiredVariables)
			return fmt.Errorf("required variable %s is not set", variable)
		}
	}

	userIDStr := os.Getenv("USER_ID")
	_, err := strconv.Atoi(userIDStr)
	if err != nil {
		return fmt.Errorf("failed to convert USER_ID to integer: %w", err)
	}

	return nil
}
