// package main

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"io/ioutil"
// 	"net/http"
// 	"os"
// 	"os/exec"
// 	"path/filepath"
// 	"strings"
// )

// // OuterResponse models the outer JSON structure returned by Dify's workflow
// type OuterResponse struct {
// 	Data struct {
// 		Status  string `json:"status"` // e.g. "succeeded"
// 		Outputs struct {
// 			Output string `json:"output"` // e.g. "{"commit_info": "..."}"
// 		} `json:"outputs"`
// 		Error interface{} `json:"error"` // might be null or some object/string
// 	} `json:"data"`
// }

// // OutputData models the inner JSON contained in data.outputs.output
// // e.g. { "commit_info": "Your commit message" }
// type OutputData struct {
// 	CommitInfo string `json:"commit_info"`
// }

// // Payload is the request body we send to Dify's /workflows/run
// type Payload struct {
// 	Inputs struct {
// 		RepoName string `json:"repo_name"`
// 		DiffText string `json:"diff_text"`
// 	} `json:"inputs"`
// 	ResponseMode string `json:"response_mode"` // e.g. "blocking"
// 	User         string `json:"user"`          // e.g. a user ID like "abc-123"
// }

// func main() {
// 	// 1) The Git hook passes the commit message file path as arg #1
// 	if len(os.Args) < 2 {
// 		fmt.Fprintln(os.Stderr, "Usage: commit_msg_generator <commit_msg_file>")
// 		os.Exit(1)
// 	}
// 	commitMsgFilePath := os.Args[1]

// 	// 2) Dynamically detect repo name from the working directory
// 	pwd, err := os.Getwd()
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Error getwd: %v\n", err)
// 		os.Exit(1)
// 	}
// 	repoName := filepath.Base(pwd)

// 	// 3) For debugging, allow reading a test diff file from env, else run git diff
// 	var diffText string
// 	if testFile := os.Getenv("TEST_DIFF_FILE"); testFile != "" {
// 		// Debug scenario
// 		diffBytes, readErr := ioutil.ReadFile(testFile)
// 		if readErr != nil {
// 			fmt.Fprintf(os.Stderr, "Error reading test diff file %s: %v\n", testFile, readErr)
// 			os.Exit(1)
// 		}
// 		diffText = string(diffBytes)
// 		fmt.Printf("[DEBUG] Using diff from file: %s\n", testFile)
// 	} else {
// 		// Production scenario
// 		diffBytes, cmdErr := exec.Command("git", "diff", "--staged").Output()
// 		if cmdErr != nil {
// 			fmt.Fprintf(os.Stderr, "Error running git diff --staged: %v\n", cmdErr)
// 			os.Exit(1)
// 		}
// 		diffText = string(diffBytes)
// 	}

// 	// If there's no diff text, just exit
// 	if strings.TrimSpace(diffText) == "" {
// 		fmt.Fprintln(os.Stderr, "No diff content found; skipping AI commit generation.")
// 		os.Exit(0)
// 	}

// 	// 4) Prepare the JSON payload for Dify
// 	payload := Payload{
// 		ResponseMode: "blocking",
// 		User:         "abc-123",
// 	}
// 	payload.Inputs.RepoName = repoName
// 	payload.Inputs.DiffText = diffText

// 	reqBody, err := json.Marshal(payload)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Error marshaling JSON payload: %v\n", err)
// 		os.Exit(1)
// 	}

// 	// 5) Create the HTTP request
// 	req, err := http.NewRequest("POST", "https://api.dify.ai/v1/workflows/run", bytes.NewBuffer(reqBody))
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Error creating HTTP request: %v\n", err)
// 		os.Exit(1)
// 	}

// 	apiKey := os.Getenv("API_KEY")
// 	if apiKey == "" {
// 		fmt.Fprintln(os.Stderr, "API_KEY environment variable not set.")
// 		os.Exit(1)
// 	}
// 	req.Header.Set("Authorization", "Bearer "+apiKey)
// 	req.Header.Set("Content-Type", "application/json")

// 	// 6) Send request
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Error doing HTTP request: %v\n", err)
// 		os.Exit(1)
// 	}
// 	defer resp.Body.Close()

// 	// If status code is not 200, handle it
// 	if resp.StatusCode != http.StatusOK {
// 		bodyBytes, _ := io.ReadAll(resp.Body)
// 		fmt.Fprintf(os.Stderr, "Dify returned status %d: %s\n", resp.StatusCode, string(bodyBytes))
// 		os.Exit(1)
// 	}

// 	// 7) Parse the outer JSON
// 	respBytes, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Error reading Dify response: %v\n", err)
// 		os.Exit(1)
// 	}

// 	var outer OuterResponse
// 	if err := json.Unmarshal(respBytes, &outer); err != nil {
// 		fmt.Fprintf(os.Stderr, "Error unmarshaling Dify response: %v\n", err)
// 		os.Exit(1)
// 	}

// 	fmt.Println("outer: ", outer)
// 	// 8) Check status in outer.Data
// 	if outer.Data.Status != "succeeded" {
// 		// If there's an error message
// 		if outer.Data.Error != nil {
// 			fmt.Fprintf(os.Stderr, "Dify returned an error: %+v\n", outer.Data.Error)
// 		} else {
// 			fmt.Fprintf(os.Stderr, "Dify returned unexpected status: %s\n", outer.Data.Status)
// 		}
// 		os.Exit(1)
// 	}

// 	// 9) Parse the "output" field which is itself JSON containing "commit_info"
// 	var outputData OutputData
// 	if err := json.Unmarshal([]byte(outer.Data.Outputs.Output), &outputData); err != nil {
// 		fmt.Println("outputData: ", outputData)
// 		fmt.Fprintf(os.Stderr, "Error parsing output JSON: %v\n", err)
// 		os.Exit(1)
// 	}

// 	commitMessage := strings.TrimSpace(outputData.CommitInfo)
// 	if commitMessage == "" {
// 		fmt.Fprintf(os.Stderr, "Empty commit_info in the output.\n")
// 		os.Exit(1)
// 	}

// 	// 10) Write the commit message so Git can show it to the user
// 	if err := ioutil.WriteFile(commitMsgFilePath, []byte(commitMessage+"\n"), 0644); err != nil {
// 		fmt.Fprintf(os.Stderr, "Error writing commit message to %s: %v\n", commitMsgFilePath, err)
// 		os.Exit(1)
// 	}

// 	fmt.Printf("[INFO] Successfully wrote AI commit to %s\n", commitMsgFilePath)
// }

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// OuterResponse models the outer JSON structure returned by Dify
type OuterResponse struct {
	Data struct {
		Status  string `json:"status"` // e.g. "succeeded"
		Outputs struct {
			Output string `json:"output"`
		} `json:"outputs"`
		Error interface{} `json:"error"`
	} `json:"data"`
}

// OutputData is the JSON we expect inside data.outputs.output
type OutputData struct {
	CommitInfo string `json:"commit_info"`
}

type Payload struct {
	Inputs struct {
		RepoName string `json:"repo_name"`
		DiffText string `json:"diff_text"`
	} `json:"inputs"`
	ResponseMode string `json:"response_mode"`
	User         string `json:"user"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: commit_msg_generator <commit_msg_file>")
		os.Exit(1)
	}
	commitMsgFilePath := os.Args[1]

	// Dynamically fetch repo name
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getwd: %v\n", err)
		os.Exit(1)
	}
	repoName := filepath.Base(pwd)

	// Decide whether to read diff from a file (debug) or from Git
	var diffText string
	if testFile := os.Getenv("TEST_DIFF_FILE"); testFile != "" {
		diffBytes, readErr := ioutil.ReadFile(testFile)
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "Error reading test diff file %s: %v\n", testFile, readErr)
			os.Exit(1)
		}
		diffText = string(diffBytes)
		fmt.Printf("[DEBUG] Using diff from file: %s\n", testFile)
	} else {
		diffBytes, cmdErr := exec.Command("git", "diff", "--staged").Output()
		if cmdErr != nil {
			fmt.Fprintf(os.Stderr, "Error running git diff --staged: %v\n", cmdErr)
			os.Exit(1)
		}
		diffText = string(diffBytes)
	}

	if strings.TrimSpace(diffText) == "" {
		fmt.Fprintln(os.Stderr, "No diff content found; skipping AI commit generation.")
		os.Exit(0)
	}

	// Prepare request payload
	payload := &Payload{
		ResponseMode: "blocking",
		User:         "abc-123",
	}
	payload.Inputs.RepoName = repoName
	payload.Inputs.DiffText = diffText

	reqBody, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling payload: %v\n", err)
		os.Exit(1)
	}

	req, err := http.NewRequest("POST", "https://api.dify.ai/v1/workflows/run", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		os.Exit(1)
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "API_KEY environment variable not set.")
		os.Exit(1)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "HTTP request error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "Dify returned status %d: %s\n", resp.StatusCode, string(bodyBytes))
		os.Exit(1)
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading Dify response: %v\n", err)
		os.Exit(1)
	}

	// 1) Parse outer JSON
	var outer OuterResponse
	if err := json.Unmarshal(respBytes, &outer); err != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshaling outer response: %v\n", err)
		os.Exit(1)
	}

	// If not succeeded, handle error
	if outer.Data.Status != "succeeded" {
		fmt.Fprintf(os.Stderr, "Dify returned status '%s'. Error detail: %+v\n", outer.Data.Status, outer.Data.Error)
		os.Exit(1)
	}

	// 2) The actual commit_info string is in outer.Data.Outputs.Output, which is itself JSON
	rawJSON := outer.Data.Outputs.Output

	// Remove code fences or backticks if the AI returned them
	rawJSON = strings.ReplaceAll(rawJSON, "```json", "")
	rawJSON = strings.ReplaceAll(rawJSON, "```", "")
	rawJSON = strings.ReplaceAll(rawJSON, "`", "")
	rawJSON = strings.TrimSpace(rawJSON)

	// 3) Parse the nested JSON
	var od OutputData
	if err := json.Unmarshal([]byte(rawJSON), &od); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing nested JSON: %v\n", err)
		os.Exit(1)
	}
	commitMessage := strings.TrimSpace(od.CommitInfo)
	if commitMessage == "" {
		fmt.Fprintf(os.Stderr, "Received empty commit_info.\n")
		os.Exit(1)
	}

	// 4) Write commit message to the file so Git can show it in the editor
	if err := ioutil.WriteFile(commitMsgFilePath, []byte(commitMessage+"\n"), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing commit message: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[INFO] AI-generated commit message inserted successfully: %s\n", commitMessage)
}
