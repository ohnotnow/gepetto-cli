package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
)

func main() {
	var model string
	var contexts multiFlag

	flag.StringVar(&model, "model", "gpt-3.5-turbo-16k", "Model to use for OpenAI (default is gpt-3.5-turbo-16k)")
	flag.Var(&contexts, "context", "Context file (can be used multiple times, use -- for stdin)")

	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Println("Please provide a message to send to the model.")
		return
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	userMessage := flag.Arg(0)
	for _, contextFile := range contexts {
		var fileName, fileContent string
		if contextFile == "--" {
			fileName = "STDIN"
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				fileContent += scanner.Text() + "\n"
			}
		} else {
			if contextFile[0] == '~' {
				usr, _ := user.Current()
				contextFile = usr.HomeDir + contextFile[1:]
			}
			fileName = contextFile
			content, err := ioutil.ReadFile(contextFile)
			if err != nil {
				fmt.Println("Error reading context file:", err)
				return
			}
			fileContent = string(content)
		}
		userMessage += fmt.Sprintf(" -- context: %s -- ```%s```", fileName, fileContent)
	}

	const maxLength = 12000
	if len(userMessage) > maxLength {
		userMessage = userMessage[:maxLength]
	}

	answer, err := askOpenAI(apiKey, model, userMessage)
	if err != nil {
		fmt.Println("Error interacting with OpenAI:", err)
		return
	}

	fmt.Println("Answer:", answer)
}

func askOpenAI(apiKey, model, userMessage string) (string, error) {
	apiEndpoint := "https://api.openai.com/v1/chat/completions"
	messages := []map[string]string{
		{"role": "system", "content": "You are a helpful assistant."},
		{"role": "user", "content": userMessage},
	}

	payload := map[string]interface{}{
		"model":    model,
		"messages": messages,
	}
	payloadJSON, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", apiEndpoint, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		// Print the entire response to help diagnose the issue
		return "", fmt.Errorf("unexpected response from OpenAI: %v", result)
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected format for choice in response from OpenAI: %v", choices[0])
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected format for message in response from OpenAI: %v", choice)
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("unexpected format for content in response from OpenAI: %v", message)
	}

	return content, nil
}

type multiFlag []string

func (f *multiFlag) String() string {
	return fmt.Sprint(*f)
}

func (f *multiFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}
