package web

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

	"github.com/andrewhowdencom/vox/internal/config"
	"github.com/andrewhowdencom/vox/internal/domain/interview"
	voxhttp "github.com/andrewhowdencom/vox/internal/http"
	"github.com/andrewhowdencom/vox/internal/adapters/providers/gemini"
	"github.com/andrewhowdencom/vox/internal/adapters/providers/static"
	"github.com/andrewhowdencom/vox/internal/adapters/storage/bbolt"
	"github.com/andrewhowdencom/vox/internal/domain/storage"
	"github.com/andrewhowdencom/vox/internal/adapters/ui/slack"

	goslack "github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Server is the HTTP server for the web port.
type Server struct {
	slackClient      *goslack.Client
	signingSecret    string
	apiKey           string
	config           *config.Config
	activeInterviews map[string]*slack.UI
	mu               sync.Mutex
	repo             storage.Repository
}

// NewServeCmd creates a new cobra command for the "serve" command.
func NewServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Starts a server to handle Slack events",
		Long:  `Starts a server to handle Slack events and run interviews.`,
		Run: func(cmd *cobra.Command, args []string) {
			port := viper.GetInt("port")
			botToken := viper.GetString("slack-bot-token")
			signingSecret := viper.GetString("slack-signing-secret")
			apiKey := viper.GetString("api-key")

			if botToken == "" || signingSecret == "" {
				slog.Error("slack-bot-token and slack-signing-secret are required")
				os.Exit(1)
			}

			var cfg config.Config
			if err := viper.Unmarshal(&cfg); err != nil {
				slog.Error("Error unmarshalling config", "error", err)
				os.Exit(1)
			}

			repo, err := bbolt.NewRepository()
			if err != nil {
				slog.Error("could not create repository", "error", err)
				os.Exit(1)
			}

			httpClient := voxhttp.NewClient(cfg.DNSServer)
			slackClient := goslack.New(botToken, goslack.OptionHTTPClient(httpClient))

			server := &Server{
				slackClient:      slackClient,
				signingSecret:    signingSecret,
				apiKey:           apiKey,
				config:           &cfg,
				activeInterviews: make(map[string]*slack.UI),
				repo:             repo,
			}

			server.Run(port)
		},
	}

	cmd.Flags().Int("port", 8080, "The port to listen on for Slack events")
	cmd.Flags().String("slack-bot-token", "", "The Slack bot token")
	cmd.Flags().String("slack-signing-secret", "", "The Slack signing secret")
	cmd.Flags().String("api-key", "", "The API key for the gemini provider")
	viper.BindPFlags(cmd.Flags())

	return cmd
}

// Run starts the HTTP server.
func (s *Server) Run(port int) {
	http.HandleFunc("/slack/events", s.createSlackEventHandler())
	http.HandleFunc("/slack/commands", s.createSlashCommandHandler())


	slog.Info("Server starting", "port", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		slog.Error("Error starting server", "error", err)
		os.Exit(1)
	}
}

func (s *Server) createSlashCommandHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		verifier, err := goslack.NewSecretsVerifier(r.Header, s.signingSecret)
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
		command, err := goslack.SlashCommandParse(r)
		if err != nil {
			slog.Error("Error parsing slash command", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		go s.handleSlashCommand(command)
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) handleSlashCommand(command goslack.SlashCommand) {
	slog.Debug("Handling slash command", "command", command.Command, "text", command.Text, "user_id", command.UserID, "channel_id", command.ChannelID)

	var topicID string
	var interviewCmd = &cobra.Command{Use: "interview"}
	var startCmd = &cobra.Command{
		Use: "start",
		Run: func(cmd *cobra.Command, args []string) {
			slog.Debug("Parsed topic ID", "topic_id", topicID)

			if topicID == "" {
				slog.Warn("Topic ID is required")
				cmd.Help()
				return
			}

			var selectedTopic *config.Topic
			for i, t := range s.config.Interviews {
				if strings.EqualFold(t.ID, topicID) {
					selectedTopic = &s.config.Interviews[i]
					break
				}
			}

			if selectedTopic == nil {
				slog.Warn("Topic not found", "topic_id", topicID)
				s.slackClient.PostEphemeral(command.ChannelID, command.UserID, goslack.MsgOptionText(fmt.Sprintf("Error: topic '%s' not found", topicID), false))
				return
			}

			s.mu.Lock()
			if _, ok := s.activeInterviews[command.UserID]; ok {
				slog.Warn("Interview already in progress for user", "user_id", command.UserID)
				s.slackClient.PostEphemeral(command.ChannelID, command.UserID, goslack.MsgOptionText("You already have an interview in progress.", false))
				s.mu.Unlock()
				return
			}
			s.mu.Unlock()

			questionProvider, err := newQuestionProvider(s.config, selectedTopic, s.apiKey, viper.GetString("model"))
			if err != nil {
				slog.Error("Error creating question provider", "error", err)
				return
			}

			convParams := &goslack.OpenConversationParameters{Users: []string{command.UserID}}
			slog.Debug("Opening conversation with user", "user_id", command.UserID)
			channel, _, _, err := s.slackClient.OpenConversation(convParams)
			if err != nil {
				slog.Error("Failed to open conversation", "error", err, "user_id", command.UserID)
				return
			}
			slog.Debug("Conversation opened", "channel_id", channel.ID)

			ui := slack.New(s.slackClient, slack.ChannelID(channel.ID), slack.UserID(command.UserID))
			s.mu.Lock()
			s.activeInterviews[command.UserID] = ui
			s.mu.Unlock()

			defer func() {
				s.mu.Lock()
				delete(s.activeInterviews, command.UserID)
				s.mu.Unlock()
				slog.Info("Interview finished for user", "user_id", command.UserID)
			}()

			slog.Info("Starting interview for user", "user_id", command.UserID)
			interviewToRun := interview.NewInterview(questionProvider, ui, s.repo)
			if err := interviewToRun.Run(command.UserID, topicID); err != nil {
				slog.Error("Error running interview", "error", err, "user_id", command.UserID)
			}
		},
	}
	startCmd.Flags().StringVar(&topicID, "topic", "", "The ID of the interview topic")
	interviewCmd.AddCommand(startCmd)

	// Create a root command to mimic the actual command structure for parsing
	rootCmd := &cobra.Command{Use: "vox"}
	rootCmd.AddCommand(interviewCmd)
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	args := strings.Fields(command.Text)
	rootCmd.SetArgs(args)

	if err := rootCmd.Execute(); err != nil {
		slog.Warn("Invalid slash command usage", "text", command.Text, "error", err)
		s.slackClient.PostEphemeral(command.ChannelID, command.UserID, goslack.MsgOptionText(buf.String(), false))
		return
	}

	// If there was any output, send it. This is useful for --help or if the Run
	// function writes to the buffer (e.g. on validation error)
	if buf.Len() > 0 {
		s.slackClient.PostEphemeral(command.ChannelID, command.UserID, goslack.MsgOptionText(buf.String(), false))
	}
}

func (s *Server) createSlackEventHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("Received slack event")
		verifier, err := goslack.NewSecretsVerifier(r.Header, s.signingSecret)
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
			s.handleCallbackEvent(eventsAPIEvent)
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) handleCallbackEvent(eventsAPIEvent slackevents.EventsAPIEvent) {
	innerEvent := eventsAPIEvent.InnerEvent
	slog.Debug("Handling callback event", "type", innerEvent.Type)
	switch ev := innerEvent.Data.(type) {
	case *slackevents.MessageEvent:
		slog.Debug("Received message event", "user_id", ev.User, "text", ev.Text, "bot_id", ev.BotID)
		if ev.BotID != "" {
			slog.Debug("Ignoring message from bot")
			return
		}
		s.mu.Lock()
		ui, ok := s.activeInterviews[ev.User]
		s.mu.Unlock()

		if ok {
			slog.Debug("Found active interview for user", "user_id", ev.User)
			ui.AnswerChan <- ev.Text
		} else {
			slog.Debug("No active interview found for user", "user_id", ev.User)
		}
	}
}

// newQuestionProvider creates a QuestionProvider based on the selected topic.
func newQuestionProvider(cfg *config.Config, topic *config.Topic, apiKey, model string) (interview.QuestionProvider, error) {
	switch strings.ToLower(topic.Provider) {
	case "static":
		return static.New(topic.Questions), nil
	case "gemini":
		if apiKey == "" {
			return nil, fmt.Errorf("api-key is required for gemini provider")
		}

		finalPrompt := buildGeminiPrompt(cfg, topic.Prompt)
		return gemini.New(cfg, gemini.Model(model), gemini.APIKey(apiKey), gemini.Prompt(finalPrompt))
	default:
		return nil, fmt.Errorf("unknown provider '%s'", topic.Provider)
	}
}

// buildGeminiPrompt constructs the final prompt for the Gemini provider.
func buildGeminiPrompt(cfg *config.Config, topicPrompt string) string {
	const DefaultSystemPrompt = "You are an interviewer." // This should be defined in a better place.
	systemPrompt := DefaultSystemPrompt
	if cfg.Providers.Gemini.Interviewer.Prompt != "" {
		systemPrompt = cfg.Providers.Gemini.Interviewer.Prompt
	}
	return fmt.Sprintf("%s\n\n%s", systemPrompt, topicPrompt)
}
