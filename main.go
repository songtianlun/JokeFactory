package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type Joke struct {
	Content  string `json:"content"`
	Language string `json:"language"`
}

func main() {
	http.HandleFunc("/", serveHTML)
	http.HandleFunc("/joke", getJoke)
	http.HandleFunc("/explain", explainJoke)

	fmt.Println("Server is running on http://0.0.0.0:8080")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func getJoke(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en"
	}
	joke := generateJoke(lang)
	json.NewEncoder(w).Encode(joke)
}

func explainJoke(w http.ResponseWriter, r *http.Request) {
	var joke Joke
	err := json.NewDecoder(r.Body).Decode(&joke)
	if err != nil {
		http.Error(w, "Invalid joke data", http.StatusBadRequest)
		return
	}
	explanation := explainJokeLogic(joke)
	json.NewEncoder(w).Encode(map[string]string{"explanation": explanation})
}

func generateJoke(language string) Joke {
	client := getOpenAIClient()

	prompt := fmt.Sprintf("Generate a joke in %s language.", language)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: getModelName(),
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		log.Printf("ChatCompletion error: %v\n", err)
		return Joke{Content: "Error generating joke. Please try again.", Language: language}
	}

	return Joke{
		Content:  resp.Choices[0].Message.Content,
		Language: language,
	}
}

func translateExplanation(explanation string, targetLang string) string {
	if targetLang == "zh" {
		// 这里可以调用翻译API或使用预定义的翻译映射
		// 简单起见,这里我们只翻译几个关键词
		explanation = strings.Replace(explanation, "Unexpected Twist", "意外转折", -1)
		explanation = strings.Replace(explanation, "Relatable Humor", "引人共鸣的幽默", -1)
		explanation = strings.Replace(explanation, "Wordplay", "文字游戏", -1)
		explanation = strings.Replace(explanation, "Cultural Context", "文化背景", -1)
	}
	return explanation
}

func explainJokeLogic(joke Joke) string {
	client := getOpenAIClient()

	var prompt string
	if joke.Language == "zh" {
		prompt = fmt.Sprintf("请用中文解释为什么这个笑话很有趣，并使用 markdown 格式来组织你的回答:\n%s", joke.Content)
	} else {
		prompt = fmt.Sprintf("Explain in English why this joke is funny, and use markdown format to organize your answer:\n%s", joke.Content)
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: getModelName(),
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		log.Printf("ChatCompletion error: %v\n", err)
		if joke.Language == "zh" {
			return "解释笑话时出错"
		}
		return "Error explaining joke"
	}

	return resp.Choices[0].Message.Content
}

func getOpenAIClient() *openai.Client {
	apiKey := os.Getenv("LLM_API_KEY")
	if apiKey == "" {
		log.Fatal("LLM_API_KEY environment variable is not set")
	}

	apiURL := os.Getenv("LLM_API_URL")
	if apiURL == "" {
		apiURL = "https://api.openai.com/v1" // 默认值
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = apiURL

	return openai.NewClientWithConfig(config)
}

func getModelName() string {
	modelName := os.Getenv("LLM_MODEL_NAME")
	if modelName == "" {
		modelName = openai.GPT3Dot5Turbo // 默认值
	}
	return modelName
}
