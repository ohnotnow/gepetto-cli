# Gepetto

Gepetto-cli is a simple golang program to answer questions via OpenAI's API from the CLI.

## Installation

To install Gepetto, you'll need to have Go installed on your system. You can download Go from the official website: https://golang.org/dl/.

Once you have Go installed and you have cloned this repo, you can run a quick test :
```
export OPENAI_API_KEY=sk....
go run gepetto 'What is the name of a famous Korean chilli sauce?'
```
And you should get a response

## Build & use the binary
```
go build gepetto
```
Now you should have `./gepetto` as a binary (which you can move somewhere in your $PATH if you like).  You can run `gepetto` either on it's own with no arguments - in which case it will prompt you to type one in.  Or if you put your question after the command it will just use that.
