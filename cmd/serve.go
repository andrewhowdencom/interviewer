package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/andrewhowdencom/vox/interview"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// activeInterviews is a simple in-memory store for ongoing interviews, keyed by Slack User ID.
var activeInterviews = make(map[string]*interview.SlackUI)
var mu sync.Mutex

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts a server to handle Slack events",
	Long:  `Starts a server to handle Slack events and run interviews.`,
	Run: func(cmd *cobra.Command, args []string) {
		port := viper.GetInt("port")
		if port == 0 {
			port = 3000
		}

		botToken := viper.GetString("slack-bot-token")
		if botToken == "" {
			log.Fatal("slack-bot-token is required")
		}
		signingSecret := viper.GetString("slack-signing-secret")
		if signingSecret == "" {
			log.Fatal("slack-signing-secret is required")
		}

		api := slack.New(botToken)

		http.HandleFunc("/slack/events", createSlackEventHandler(signingSecret, api, cmd))
		http.HandleFunc("/slack/commands", createSlashCommandHandler(signingSecret, api, cmd))

		log.Printf("Server starting on port %d", port)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	},
}

func createSlashCommandHandler(signingSecret string, api *slack.Client, cmd *cobra.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		verifier, err := slack.NewSecretsVerifier(r.Header, signingSecret)
		if err != nil {
			log.Printf("Error creating secrets verifier: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		verifier.Write(body)
		if err := verifier.Ensure(); err != nil {
			log.Printf("Error verifying request signature: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(body))
		s, err := slack.SlashCommandParse(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		go handleSlashCommand(cmd, api, s)
		w.WriteHeader(http.StatusOK)
	}
}

func handleSlashCommand(cmd *cobra.Command, api *slack.Client, s slack.SlashCommand) {
	args := strings.Fields(s.Text)
	if len(args) < 4 || args[0] != "interview" || args[1] != "start" || args[2] != "--topic" {
		api.PostEphemeral(s.ChannelID, s.UserID, slack.MsgOptionText("Usage: /vox interview start --topic <topic-id>", false))
		return
	}
	topicID := args[3]

	var config interview.Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Printf("Error unmarshalling config: %v", err)
		return
	}

	var selectedTopic *interview.Topic
	for i, t := range config.Interviews {
		if strings.EqualFold(t.ID, topicID) {
			selectedTopic = &config.Interviews[i]
			break
		}
	}

	if selectedTopic == nil {
		api.PostEphemeral(s.ChannelID, s.UserID, slack.MsgOptionText(fmt.Sprintf("Error: topic '%s' not found", topicID), false))
		return
	}

	mu.Lock()
	if _, ok := activeInterviews[s.UserID]; ok {
		api.PostEphemeral(s.ChannelID, s.UserID, slack.MsgOptionText("You already have an interview in progress.", false))
		mu.Unlock()
		return
	}
	mu.Unlock()

	var questionProvider interview.QuestionProvider
	var err error
	switch strings.ToLower(selectedTopic.Provider) {
	case "static":
		questionProvider = interview.NewStaticQuestionProvider(selectedTopic.Questions)
	case "gemini":
		apiKey := viper.GetString("api-key")
		if apiKey == "" {
			log.Println("Error: api-key is required for gemini provider")
			return
		}
		model := viper.GetString("model")
		if !cmd.Flags().Changed("model") && config.Providers.Gemini.Model != "" {
			model = config.Providers.Gemini.Model
		}
		questionProvider, err = interview.NewGeminiQuestionProvider(model, apiKey, selectedTopic.Prompt)
		if err != nil {
			log.Printf("Error creating gemini provider: %v", err)
			return
		}
	default:
		log.Printf("Error: unknown provider '%s'", selectedTopic.Provider)
		return
	}

	convParams := &slack.OpenConversationParameters{Users: []string{s.UserID}}
	channel, _, _, err := api.OpenConversation(convParams)
	if err != nil {
		log.Printf("Failed to open conversation: %v", err)
		return
	}

	ui := interview.NewSlackUI(api, channel.ID, s.UserID)
	mu.Lock()
	activeInterviews[s.UserID] = ui
	mu.Unlock()

	defer func() {
		mu.Lock()
		delete(activeInterviews, s.UserID)
		mu.Unlock()
		log.Printf("Interview finished for user %s", s.UserID)
	}()

	log.Printf("Starting interview for user %s", s.UserID)
	if err := runInterview(cmd.OutOrStdout(), questionProvider, ui); err != nil {
		log.Printf("Error running interview: %v", err)
	}
}

func createSlackEventHandler(signingSecret string, api *slack.Client, cmd *cobra.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		verifier, err := slack.NewSecretsVerifier(r.Header, signingSecret)
		if err != nil {
			log.Printf("Error creating secrets verifier: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		verifier.Write(body)
		if err := verifier.Ensure(); err != nil {
			log.Printf("Error verifying request signature: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			log.Printf("Error parsing event: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if eventsAPIEvent.Type == slackevents.URLVerification {
			var r *slackevents.ChallengeResponse
			if err := json.Unmarshal(body, &r); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(r.Challenge))
			return
		}

		if eventsAPIEvent.Type == slackevents.CallbackEvent {
			handleCallbackEvent(api, eventsAPIEvent)
		}
		w.WriteHeader(http.StatusOK)
	}
}

func handleCallbackEvent(api *slack.Client, eventsAPIEvent slackevents.EventsAPIEvent) {
	innerEvent := eventsAPIEvent.InnerEvent
	switch ev := innerEvent.Data.(type) {
	case *slackevents.MessageEvent:
		if ev.BotID != "" {
			return
		}
		mu.Lock()
		ui, ok := activeInterviews[ev.User]
		mu.Unlock()

		if ok {
			ui.AnswerChan <- ev.Text
		}
	}
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().Int("port", 3000, "The port to listen on for Slack events")
	serveCmd.Flags().String("slack-bot-token", "", "The Slack bot token")
	serveCmd.Flags().String("slack-signing-secret", "", "The Slack signing secret")
	viper.BindPFlags(serveCmd.Flags())
}
