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
	var chat bool

	flag.StringVar(&model, "model", "gpt-4o", "Model to use for OpenAI (default is gpt-4o)")
	flag.Var(&contexts, "context", "Context file (can be used multiple times, use -- for stdin)")
	flag.BoolVar(&chat, "chat", false, "Enable chat mode (conversational interaction with the model)")

	flag.Parse()

	if flag.NArg() == 0 {
		chat = true
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

	if chat {
		// Initialize conversation history
		var conversation []map[string]string
		conversation = append(conversation, map[string]string{"role": "system", "content": "You are a helpful assistant."})

		// Handle initial question if provided
		if userMessage != "" {
			// Add user's initial message to the conversation
			conversation = append(conversation, map[string]string{"role": "user", "content": userMessage})

			// Ask OpenAI
			answer, err := askOpenAI(apiKey, model, conversation)
			if err != nil {
				fmt.Println("Error interacting with OpenAI:", err)
				return
			}

			// Add OpenAI's response to the conversation
			conversation = append(conversation, map[string]string{"role": "assistant", "content": answer})

			// Print OpenAI's response
			fmt.Println("Assistant:", answer)
		}

		// Start chat mode loop
		// ... (same code as above for chat mode)
		// Start chat mode loop
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("You: ")
			userMessage, err := reader.ReadString('\n')
			if userMessage == "\n" { // Exit if Ctrl-D is pressed or input is empty
				fmt.Println("Exiting chat mode.")
				return
			}

			// Add user's message to the conversation
			conversation = append(conversation, map[string]string{"role": "user", "content": userMessage[:len(userMessage)-1]})

			// Ask OpenAI
			answer, err := askOpenAI(apiKey, model, conversation)
			if err != nil {
				fmt.Println("Error interacting with OpenAI:", err)
				return
			}

			// Add OpenAI's response to the conversation
			conversation = append(conversation, map[string]string{"role": "assistant", "content": answer})

			// Print OpenAI's response
			fmt.Println("Assistant:", answer)
		}
	} else {
		conversation := []map[string]string{
			{"role": "system", "content": "You are a helpful assistant."},
			{"role": "user", "content": userMessage},
		}

		// Call the askOpenAI function with the conversation history
		answer, err := askOpenAI(apiKey, model, conversation)
		if err != nil {
			fmt.Println("Error interacting with OpenAI:", err)
			return
		}

		fmt.Println("Answer:", answer)
	}
}

func askOpenAI(apiKey, model string, conversation []map[string]string) (string, error) {
	apiEndpoint := "https://api.openai.com/v1/chat/completions"

	payload := map[string]interface{}{
		"model":    model,
		"messages": conversation,
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

	// fmt.Println("Raw response from OpenAI:", string(body)) // Print the raw response

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("unexpected response format: no choices found")
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected response format: choice is not a map")
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected response format: message is not a map")
	}

	text, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("unexpected response format: content is not a string")
	}

	return text, nil
}

type multiFlag []string

func (f *multiFlag) String() string {
	return fmt.Sprint(*f)
}

func (f *multiFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}
