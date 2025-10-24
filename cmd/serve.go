package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
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
			slog.Error("slack-bot-token is required")
			os.Exit(1)
		}
		signingSecret := viper.GetString("slack-signing-secret")
		if signingSecret == "" {
			slog.Error("slack-signing-secret is required")
			os.Exit(1)
		}
		apiKey := viper.GetString("api-key")

		api := slack.New(botToken)

		http.HandleFunc("/slack/events", createSlackEventHandler(signingSecret, api, cmd))
		http.HandleFunc("/slack/commands", createSlashCommandHandler(signingSecret, api, cmd, apiKey))

		slog.Info("Server starting", "port", port)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			slog.Error("Error starting server", "error", err)
			os.Exit(1)
		}
	},
}

func createSlashCommandHandler(signingSecret string, api *slack.Client, cmd *cobra.Command, apiKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		verifier, err := slack.NewSecretsVerifier(r.Header, signingSecret)
		if err != nil {
			slog.Error("Error creating secrets verifier", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("Error reading request body", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		verifier.Write(body)
		if err := verifier.Ensure(); err != nil {
			slog.Error("Error verifying request signature", "error", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(body))
		s, err := slack.SlashCommandParse(r)
		if err != nil {
			slog.Error("Error parsing slash command", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		go handleSlashCommand(cmd, api, s, apiKey)
		w.WriteHeader(http.StatusOK)
	}
}

func handleSlashCommand(cmd *cobra.Command, api *slack.Client, s slack.SlashCommand, apiKey string) {
	slog.Debug("Handling slash command", "command", s.Command, "text", s.Text, "user_id", s.UserID, "channel_id", s.ChannelID)
	args := strings.Fields(s.Text)
	if len(args) < 4 || args[0] != "interview" || args[1] != "start" || args[2] != "--topic" {
		slog.Warn("Invalid slash command usage", "text", s.Text)
		api.PostEphemeral(s.ChannelID, s.UserID, slack.MsgOptionText("Usage: /vox interview start --topic <topic-id>", false))
		return
	}
	topicID := args[3]
	slog.Debug("Parsed topic ID", "topic_id", topicID)

	var config interview.Config
	if err := viper.Unmarshal(&config); err != nil {
		slog.Error("Error unmarshalling config", "error", err)
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
		slog.Warn("Topic not found", "topic_id", topicID)
		api.PostEphemeral(s.ChannelID, s.UserID, slack.MsgOptionText(fmt.Sprintf("Error: topic '%s' not found", topicID), false))
		return
	}

	mu.Lock()
	if _, ok := activeInterviews[s.UserID]; ok {
		slog.Warn("Interview already in progress for user", "user_id", s.UserID)
		api.PostEphemeral(s.ChannelID, s.UserID, slack.MsgOptionText("You already have an interview in progress.", false))
		mu.Unlock()
		return
	}
	mu.Unlock()

	var questionProvider interview.QuestionProvider
	var err error
	switch strings.ToLower(selectedTopic.Provider) {
	case "static":
		slog.Debug("Creating static question provider")
		questionProvider = interview.NewStaticQuestionProvider(selectedTopic.Questions)
	case "gemini":
		slog.Debug("Creating gemini question provider")
		if apiKey == "" {
			slog.Error("api-key is required for gemini provider")
			return
		}
		model := viper.GetString("model")
		if !cmd.Flags().Changed("model") && config.Providers.Gemini.Model != "" {
			model = config.Providers.Gemini.Model
		}
		questionProvider, err = interview.NewGeminiQuestionProvider(model, apiKey, selectedTopic.Prompt)
		if err != nil {
			slog.Error("Error creating gemini provider", "error", err)
			return
		}
	default:
		slog.Error("Unknown provider", "provider", selectedTopic.Provider)
		return
	}

	convParams := &slack.OpenConversationParameters{Users: []string{s.UserID}}
	slog.Debug("Opening conversation with user", "user_id", s.UserID)
	channel, _, _, err := api.OpenConversation(convParams)
	if err != nil {
		slog.Error("Failed to open conversation", "error", err, "user_id", s.UserID)
		return
	}
	slog.Debug("Conversation opened", "channel_id", channel.ID)

	ui := interview.NewSlackUI(api, channel.ID, s.UserID)
	mu.Lock()
	activeInterviews[s.UserID] = ui
	mu.Unlock()

	defer func() {
		mu.Lock()
		delete(activeInterviews, s.UserID)
		mu.Unlock()
		slog.Info("Interview finished for user", "user_id", s.UserID)
	}()

	slog.Info("Starting interview for user", "user_id", s.UserID)
	if err := runInterview(cmd.OutOrStdout(), questionProvider, ui); err != nil {
		slog.Error("Error running interview", "error", err, "user_id", s.UserID)
	}
}

func createSlackEventHandler(signingSecret string, api *slack.Client, cmd *cobra.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("Received slack event")
		verifier, err := slack.NewSecretsVerifier(r.Header, signingSecret)
		if err != nil {
			slog.Error("Error creating secrets verifier", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("Error reading request body", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		verifier.Write(body)
		if err := verifier.Ensure(); err != nil {
			slog.Error("Error verifying request signature", "error", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			slog.Error("Error parsing event", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		slog.Debug("Parsed slack event", "type", eventsAPIEvent.Type)

		if eventsAPIEvent.Type == slackevents.URLVerification {
			var r *slackevents.ChallengeResponse
			if err := json.Unmarshal(body, &r); err != nil {
				slog.Error("Error unmarshalling challenge response", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(r.Challenge))
			slog.Debug("Responded to URL verification challenge")
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
	slog.Debug("Handling callback event", "type", innerEvent.Type)
	switch ev := innerEvent.Data.(type) {
	case *slackevents.MessageEvent:
		slog.Debug("Received message event", "user_id", ev.User, "text", ev.Text, "bot_id", ev.BotID)
		if ev.BotID != "" {
			slog.Debug("Ignoring message from bot")
			return
		}
		mu.Lock()
		ui, ok := activeInterviews[ev.User]
		mu.Unlock()

		if ok {
			slog.Debug("Found active interview for user", "user_id", ev.User)
			ui.AnswerChan <- ev.Text
		} else {
			slog.Debug("No active interview found for user", "user_id", ev.User)
		}
	}
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().Int("port", 3000, "The port to listen on for Slack events")
	serveCmd.Flags().String("slack-bot-token", "", "The Slack bot token")
	serveCmd.Flags().String("slack-signing-secret", "", "The Slack signing secret")
	serveCmd.Flags().String("api-key", "", "The API key for the gemini provider")
	viper.BindPFlags(serveCmd.Flags())
}
