package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
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

	fmt.Println("\033[1;36m🤖 Analyzing your changes to generate a commit message...\033[0m")

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
		diffBytes, readErr := os.ReadFile(testFile)
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

	fmt.Println("\033[1;33m⏳ Generating commit message based on your changes...\033[0m")

	currentUser, err := user.Current()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error get currentUser: %v\n", err)
	}

	username := currentUser.Username
	// Prepare request payload
	payload := &Payload{
		ResponseMode: "blocking",
		User:         username,
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

	// Display the generated commit message to the user
	fmt.Println("\033[1;32m✅ AI generated commit message:\033[0m")
	fmt.Println("\033[1m" + commitMessage + "\033[0m")
	fmt.Println("\033[3mYour editor will open next - you can edit this message before saving\033[0m")

	// 4) Write commit message to the file so Git can show it in the editor
	if err := os.WriteFile(commitMsgFilePath, []byte(commitMessage+"\n"), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing commit message: %v\n", err)
		os.Exit(1)
	}
}
