package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"flag"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

const OpenAIURL = "https://api.openai.com/v1/chat/completions"

type Message struct {
	Role string `json:"role"`
	Content string `json:"content"`
}

type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func callOpenAI(question string, fileContent string) (string, error) {
	systemMessage := "You are a helpful assistant.\nFile content:\n" + fileContent
	requestData := OpenAIRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{Role: "system", Content: systemMessage},
			{Role: "user", Content: question},
		},
	}
	requestBody, _ := json.Marshal(requestData)

	request, _ := http.NewRequest("POST", OpenAIURL, bytes.NewBuffer(requestBody))
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("OPENAI_API_KEY")))
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	responseBody := &OpenAIResponse{}
	err = json.NewDecoder(response.Body).Decode(responseBody)
	if err != nil {
		return "", err
	}

	// Check if Choices is not empty
	if len(responseBody.Choices) == 0 {
		bodyBytes, _ := ioutil.ReadAll(response.Body)
		bodyString := string(bodyBytes)
		fmt.Println(bodyString)
		return "", errors.New("no choices in OpenAI response")
	}

	return strings.TrimSpace(responseBody.Choices[0].Message.Content), nil
}

func sanitize(s string) string {
    result := ""
    for _, r := range s {
        if r >= 32 && r <= 126 {
            result += string(r)
        }
    }
    return result
}

func main() {
	attachFile := flag.String("attach", "", "Path to the file to attach as extra context")
    flag.Parse()

	fileContentStr := ""
    // Check if the user provided a file to attach
    if *attachFile != "" {
        fileContent, err := ioutil.ReadFile(*attachFile)
        if err != nil {
            fmt.Println("Error reading file:", err)
            return
        }

        // Limit the size of the file content to 3000 characters
        fileContentStr = string(fileContent)
        if len(fileContentStr) > 3000 {
            fileContentStr = fileContentStr[:3000]
        }

        // Now you can use `fileContentStr` in your call to the OpenAI API,
        // for example by adding it as a new system message before the user's message.
    }

	nonFlagArgs := flag.Args()
	question := strings.Join(nonFlagArgs, " ")
	question = strings.TrimSpace(question)  // Remove white space from both ends of the string

	fmt.Println("Asking GPT-3:", question)
	if len(question) == 0 {
		fmt.Println("Please type your question. Press control+d to finish.")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		answer, err := callOpenAI(input, fileContentStr)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		fmt.Println(answer)
	} else {
		answer, err := callOpenAI(question, fileContentStr)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		fmt.Println(sanitize(answer))
	}
}
