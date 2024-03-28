package main

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/melardev/discord-message-protect/captchas"
	"github.com/melardev/discord-message-protect/core"
	"github.com/melardev/discord-message-protect/http_server"
	"github.com/melardev/discord-message-protect/logging"
	"github.com/melardev/discord-message-protect/pollution"
	"github.com/melardev/discord-message-protect/secrets"
	"github.com/melardev/discord-message-protect/sessions"
	"github.com/melardev/discord-message-protect/utils"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ApplicationContext struct {
	DiscordBot           discordgo.Session
	CaptchaValidator     captchas.ICaptchaValidator
	SessionManager       sessions.ISessionManager
	SecretManager        secrets.ISecretManager
	DefaultLogger        logging.ILogger
	ErrorLogger          logging.ILogger
	PollutionLogger      logging.ILogger
	PollutionStrategy    pollution.IPollutionStrategy
	Config               *core.Config
	HttpServer           *http_server.MuxHttpServer
	revealSecretMutex    sync.Mutex
	revealSecretRequests map[string]*secrets.RevealRequest
}

type RunArgs struct {
	ConfigPath string
	Verbose    bool
}

type Application struct {
	Hostname              string
	Context               *ApplicationContext
	RolesFilters          map[string][]string
	ErrorLogger           logging.ILogger
	EditOriginalChallenge bool
	cleanedReq            int
}

const (
	genericMessageRetry = "An error occurred, please retry or contact your admin if the issue persists"
)

var app *Application

func GetApplication(args *RunArgs) *Application {
	if app != nil {
		return app
	}

	hostname, _ := os.Hostname()
	app = &Application{
		Hostname:              hostname,
		EditOriginalChallenge: true,
	}

	app.Context = &ApplicationContext{
		revealSecretRequests: map[string]*secrets.RevealRequest{},
	}
	configPath := ""
	if args.ConfigPath != "" {
		configPath = args.ConfigPath
	} else {
		configPath = "config.json"
	}

	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(content, &app.Context.Config)
	if err != nil {
		panic(err)
	}

	if app.Context.Config.AppPath == "" {
		appPath, err := filepath.Abs(".")
		if err != nil {
			panic(err)
		}
		app.Context.Config.AppPath = appPath
	}

	if app.Context.Config.LogPath == "" {
		logPath := filepath.Join(app.Context.Config.AppPath, "logs")

		app.Context.Config.LogPath = logPath
	}

	if app.Context.Config.DatabaseConfig != nil {
		dbConfig := app.Context.Config.DatabaseConfig
		if dbConfig.Username != "" {
			err = os.Setenv("DB_USER", dbConfig.Username)
			if err != nil {
				panic(err)
			}
		}

		if dbConfig.Password != "" {
			err = os.Setenv("DB_PASSWORD", dbConfig.Password)
			if err != nil {
				panic(err)
			}
		}

		if dbConfig.Hostname != "" {
			err = os.Setenv("DB_HOST", dbConfig.Hostname)
			if err != nil {
				panic(err)
			}
		}

		if dbConfig.Port > 0 {
			err = os.Setenv("DB_Port", strconv.Itoa(dbConfig.Port))
			if err != nil {
				panic(err)
			}
		}
	}

	if !utils.DirExists(app.Context.Config.LogPath) {
		err := os.MkdirAll(app.Context.Config.LogPath, 0644)
		if err != nil {
			panic(err)
		}
	}

	app.Context.DefaultLogger = logging.NewCompositeLogger(
		&logging.ConsoleLogger{MinLevel: logging.Warn},
		logging.NewFileLogger(filepath.Join(app.Context.Config.LogPath, "app.log")),
	)

	app.Context.ErrorLogger = logging.NewCompositeLogger(
		&logging.ConsoleLogger{MinLevel: logging.Debug},
		logging.NewFileLogger(filepath.Join(app.Context.Config.LogPath, "errors.log")),
	)

	app.Context.PollutionLogger = logging.NewCompositeLogger(
		&logging.ConsoleLogger{},
		logging.NewFileLogger(filepath.Join(app.Context.Config.LogPath, "pollution.log")),
	)

	if args.Verbose {
		app.Context.DefaultLogger.SetMinLevel(logging.Debug)
	} else {
		app.Context.DefaultLogger.SetMinLevel(logging.Warn)
	}

	app.Context.PollutionLogger.SetMinLevel(logging.Warn)
	app.Context.SessionManager = sessions.NewInMemoryAuthenticator(app.Context.Config)
	app.Context.SecretManager = secrets.NewDbSecretManager(app.Context.Config)

	if app.Context.Config.HttpConfig != nil {
		// For now, we only use the HTTP server for captcha, so if captcha is disabled, then no point
		// on using it
		if app.Context.Config.HttpConfig.CaptchaService != "" {
			app.Context.HttpServer = http_server.NewMuxHttpServer(app.Context.Config, app)
			app.Context.HttpServer.Run()
		}
	}

	if app.Context.Config.PollutionConfig != nil {
		app.Context.PollutionStrategy = pollution.GetPollutionStrategy(app.Context.Config.PollutionConfig.StrategyName,
			app.Context.Config.PollutionConfig.Args)
	}

	dg, err := discordgo.New("Bot " + app.Context.Config.DiscordConfig.BotToken)
	if err != nil {
		message := fmt.Sprintf("error creating DiscordConfig session %v\n", err)
		panic(message)
	}

	// Just like the ping pong example, we only care about receiving message
	// events in this example.
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentMessageContent
	// Open a websocket connection to DiscordConfig and begin listening.
	err = dg.Open()
	if err != nil {
		message := fmt.Sprintf("error opening connection %v\n", err)
		panic(message)
	}

	dg.AddHandler(app.OnMessageCreate)
	dg.AddHandler(app.OnDiscordBotReady)
	dg.AddHandler(app.OnCreateInteraction)

	_, err = dg.ApplicationCommandCreate(
		app.Context.Config.DiscordConfig.AppId,
		app.Context.Config.DiscordConfig.GuildId, &discordgo.ApplicationCommand{
			Name:        app.Context.Config.DiscordConfig.ProtectCommandName,
			Description: "Protects a message by challenging the user",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "content",
					Description: "Content",
					Required:    true,
				},
				{
					// TODO: Implement me, for now it is just in paper, there is no implementation backing this
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "pollute",
					Description: "Pollute message",
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name: "incremental",
							Value: pollution.GetPollutionStrategy(pollution.IncrementIntStrategyName,
								map[string]interface{}{
									"position": "beginning",
								}),
						},
						{
							Name: "faker",
							Value: pollution.GetPollutionStrategy(pollution.FakerStrategyName,
								map[string]interface{}{
									"position":  pollution.Random,
									"min_words": 1,
									"max_words": 10,
								}),
						},
					},
					Required: false,
				},
				{
					// TODO: Implement me, for now it is just in paper, there is no implementation backing this
					Type:        discordgo.ApplicationCommandOptionAttachment,
					Name:        "image",
					Description: "Content",
					Required:    false,
				},
			},
		})

	app.Context.DefaultLogger.Info(fmt.Sprintf("Application started\n"))
	return app
}

func (a *Application) Run() {

	if a.Context.Config.HttpConfig.CaptchaService != "" {
		// for now we only do maintenance on Reveal Requests which are only
		// used when Captcha protection is enabled
		go a.DoMaintenance()
	}

	// TODO: Improve this
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	a.Context.DefaultLogger.Debug("Exiting application")
}

func (a *Application) OnDiscordBotReady(s *discordgo.Session, r *discordgo.Ready) {
	a.Context.DefaultLogger.Info(fmt.Sprintf("DiscordConfig bot initialized, logged in on %s as - %s#%s\n",
		strings.Join(utils.GetGuildNames(s.State.Guilds), ","),
		s.State.User.Username,
		s.State.User.Discriminator))
}

func (a *Application) OnCreateInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		// A command has been triggered
		switch i.ApplicationCommandData().Name {
		case a.Context.Config.DiscordConfig.ProtectCommandName:
			a.OnProtectCommand(s, i)
		}
	case discordgo.InteractionMessageComponent:
		// An interaction has been triggered
		customId := i.MessageComponentData().CustomID
		if strings.HasPrefix(customId, "btn_unlock_") {
			a.OnUnlockInteraction(s, i)
		} else if strings.HasPrefix(customId, "btn_edit_") {
			a.OnUpdateSecret(s, i)
		} else if strings.HasPrefix(customId, "btn_delete_") {
			a.OnDeleteSecret(s, i)
		} else if strings.HasPrefix(customId, "btn_reveal_") {
			a.OnRevealSecret(s, i)
		}
	}
}

// OnProtectCommand - handles the protect command.
func (a *Application) OnProtectCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if a.RolesFilters != nil {
		commandName := app.Context.Config.DiscordConfig.ProtectCommandName
		member := i.Interaction.Member

		if rules, found := a.RolesFilters[commandName]; found {
			isAuthorized := false
			roles := member.Roles
			isAuthorized = false
			for _, whitelist := range rules {
				for _, r := range roles {
					if r == whitelist {
						isAuthorized = true
						break
					}
				}

				if isAuthorized {
					break
				}
			}

			if !isAuthorized {
				message := fmt.Sprintf("User %s#%s tried to use protect command yet he is not authorized to do so.\n",
					member.User.Username, member.User.Discriminator)
				a.ErrorLogger.Error(message)
				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Flags: discordgo.MessageFlagsEphemeral,
						Content: "You are not allowed to use this command, this error has been logged, please stop to avoid " +
							"getting banned for spam.",
					},
				})

				if err != nil {
					a.Context.ErrorLogger.Error(fmt.Sprintf("An error occurred on !isAuthorized %v\n", err))
					return
				}

				return
			}
		}
	}

	var messageContent string
	var imageUrl string

	if data, ok := i.Interaction.Data.(discordgo.ApplicationCommandInteractionData); ok {
		for _, option := range data.Options {
			if option.Name == "content" {
				messageContent = option.Value.(string)
			}

			if option.Name == "images" {
				attachmentId := option.Value.(string)
				attachment := data.Resolved.Attachments[attachmentId]
				// messageContent += "\n" + attachment.URL
				imageUrl = attachment.URL
			}
		}
	}

	// it is documented that in some conditions i.User is nil instead we should use i.Member
	user := i.User
	if user == nil {
		user = i.Member.User
	}

	secret, err := a.Context.SecretManager.Create(&secrets.CreateSecretDto{
		User:      user,
		Message:   messageContent,
		ChannelId: i.ChannelID,
		ImageUrl:  imageUrl,
	})

	if err != nil {
		a.Context.ErrorLogger.Error(fmt.Sprintf("An error occurred Creating the secret - %v\n", err))
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "An error occurred Creating the secret, please retry or contact your admin if the issue persists",
			},
		})
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: "Users are notified about secret, they must click on the unlock button now",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Reveal",
							CustomID: fmt.Sprintf("btn_reveal_%s", secret.SecretId),
							Style:    discordgo.PrimaryButton,
							Disabled: false,
							Emoji: discordgo.ComponentEmoji{
								Name:     "👀",
								Animated: false,
							},
						},
						discordgo.Button{
							Label:    "Delete",
							CustomID: fmt.Sprintf("btn_delete_%s", secret.SecretId),
							Style:    discordgo.PrimaryButton,
							Disabled: false,
							Emoji: discordgo.ComponentEmoji{
								Name:     "❌",
								Animated: false,
							},
						},
					},
				},
			},
		},
	})

	if err != nil {
		a.Context.ErrorLogger.Error(fmt.Sprintf("An error occurred on OnProtectCommand::InteractionRespond - %v\n", err))
		err = a.Context.SecretManager.Delete(secret.SecretId)
		if err != nil {
			a.Context.ErrorLogger.Error("An error occurred on OnProtectCommand::InteractionRespond::Err::Delete - %v\n", err)
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "An error occurred after creating the secret, please retry or contact your admin if the issue persists",
			},
		})
		return
	}

	// Send a message to everyone indicating we have a new protected message
	// to read it they have to pass the challenge (a click for now, probably a captcha too)
	m, err := s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
		Content: "Click to Unlock",
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Click to Unlock",
						CustomID: fmt.Sprintf("btn_unlock_%s", secret.SecretId),
						Style:    discordgo.PrimaryButton,
						Disabled: false,
						Emoji: discordgo.ComponentEmoji{
							Name:     "🔒",
							Animated: false,
						},
					},
				},
			},
		},
	})

	if err != nil {
		panic(err)
	} else {
		secret.MessageId = m.ID
		secret.Interaction = i.Interaction
		err = a.Context.SecretManager.UpdateMessageId(secret)
		if err != nil {
			a.Context.ErrorLogger.Error(fmt.Sprintf("An error occured Updating MessageId - %v\n", err))
		}
	}
}

// OnUnlockInteraction is a callback triggered when the user wants to see a protected message
// if the user is authenticated (passed challenge recently) he is gonna get the message without further interaction
// if not, the user would need to pass an additional challenge(if configured)
func (a *Application) OnUnlockInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customId := i.MessageComponentData().CustomID
	secretId := strings.Replace(customId, "btn_unlock_", "", -1)

	secret, err := a.Context.SecretManager.GetById(secretId)
	if err != nil {
		a.Context.ErrorLogger.Error(fmt.Sprintf("An error occurred on UnlockInteraction::GetById %v\n",
			err))

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: genericMessageRetry,
			},
		})

		if err != nil {
			a.Context.ErrorLogger.Error(fmt.Sprintf("An error occurred on UnlockInteraction::ErrInteractionRespond %v\n",
				err))
		}
		return
	}

	if secret == nil {
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: fmt.Sprintf("Protected message has expired or no longer available"),
			},
		})

		if err != nil {
			a.Context.ErrorLogger.Error(fmt.Sprintf("An error occurred on OnUnlockInteraction::SecretNull - %v\n", err))
		}

		return
	}

	// If we are not using captcha based protection, or we do but the user already passed it previously
	// then just show the protected message without further action from the user
	if a.Context.Config.HttpConfig.CaptchaService == "" ||
		a.Context.SessionManager.IsAuthenticated(s.State.SessionID) {
		secretContent := secret.Message

		if a.Context.PollutionStrategy != nil {
			user := i.User
			if user == nil {
				user = i.Member.User
			}

			modifiedMessage, indicators := a.Context.PollutionStrategy.Apply(secretContent,
				user.Username, user.Discriminator)

			a.Context.PollutionLogger.Info(fmt.Sprintf("Applied %s strategy, User: %s, Id: %s, Indicators: %s\n",
				a.Context.PollutionStrategy.GetName(),
				fmt.Sprintf("%s#%s", user.Username, user.Discriminator),
				secret.SecretId, indicators))
			secretContent = modifiedMessage
		}

		interactionResponse := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: secretContent,
			},
		}

		if secret.ImageUrl != "" {
			// If secret had an image we need to attach it as Embed
			interactionResponse.Data.Embeds = []*discordgo.MessageEmbed{
				{
					Image: &discordgo.MessageEmbedImage{
						URL: secret.ImageUrl,
					},
				},
			}
		}

		// User is authenticated, reveal the secret
		err = s.InteractionRespond(i.Interaction, interactionResponse)

		if err != nil {
			a.Context.ErrorLogger.Error(fmt.Sprintf("An error occurred on OnUnlockInteraction::InteractionRespond for authenticated user - %v\n",
				err))
		}
	} else {
		// Send another interaction, the ideal would be to use a LinkButton with CustomID to get its interaction
		// but it is not possible, so we must notify the user to click on the embed's link and only click
		// the unlock button after he resolved the captcha challenge, I have to find a better solution for this
		// but for now it does not seem to be possible otherwise.
		reqId := uuid.Must(uuid.NewRandom()).String()

		challengePath := a.Context.Config.HttpConfig.ChallengePath
		if !strings.HasPrefix(challengePath, "/") {
			challengePath = "/" + challengePath
		}

		now := time.Now().UTC()
		challengeUrl := fmt.Sprintf("%s://%s:%d%s?req_id=%s",
			a.Context.Config.HttpConfig.Scheme,
			a.Context.Config.HttpConfig.Hostname,
			a.Context.Config.HttpConfig.Port,
			challengePath,
			reqId)

		user := i.User
		if user == nil {
			user = i.Member.User
		}

		req := &secrets.RevealRequest{
			User: &core.DiscordUser{
				Id:       user.Discriminator,
				Username: user.Username,
			},
			Secret:    secret,
			ChannelId: secret.ChannelId,
			CreatedAt: now,
			UpdatedAt: now,
		}

		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
				/*Embeds: []*discordgo.MessageEmbed{
					{
						Type:        discordgo.EmbedTypeLink,
						Title:       "Security Check",
						Description: "Please click the following link, solve the captcha, then click on Unlock",
						URL:         challengeUrl,
					},
				},*/
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{
								Label:    "Resolve Captcha",
								Style:    discordgo.LinkButton,
								Disabled: false,
								URL:      challengeUrl,
								Emoji: discordgo.ComponentEmoji{
									Name:     "🔒",
									Animated: false,
								},
							},
						},
					},
				},
			},
		})

		if err != nil {
			a.Context.ErrorLogger.Error(fmt.Sprintf("An error occurred on OnUnlockInteraction::CreateChallenge - %v\n", err))
		} else {
			a.Context.revealSecretMutex.Lock()
			a.Context.revealSecretRequests[reqId] = req
			a.Context.revealSecretMutex.Unlock()
			// The interaction object is going to be used later to send the message to the user
			req.Interaction = i.Interaction
			req.Session = s
		}
	}

}

func (a *Application) OnMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Nothing for now
}

func (a *Application) OnUpdateSecret(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// TODO: Implement
}

// OnDeleteSecret - Is called when the secret owner clicks on the Delete button.
func (a *Application) OnDeleteSecret(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customId := i.MessageComponentData().CustomID
	secretId := strings.Replace(customId, "btn_delete_", "", -1)

	secret, err := a.Context.SecretManager.GetById(secretId)
	if err != nil {
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: genericMessageRetry,
			},
		})
		a.Context.ErrorLogger.Error(fmt.Sprintf("An error occurred on OnDelete::GetById(%s) - %v\n",
			secretId, err))
		return
	}

	if secret != nil {
		a.deleteSecret(s, i, secret)

		if a.Context.Config.AckActionOnProtectedMessage {
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   discordgo.MessageFlagsEphemeral,
					Content: "Protected Message deleted",
				},
			})

			if err != nil {
				message := fmt.Sprintf("An error occurred OnDeleteSecret::InteractionRespond %v\n", err)
				a.Context.ErrorLogger.Error(message)
			}
		}
	} else {
		a.Context.ErrorLogger.Error(fmt.Sprintf("Trying to delete secret %s but does not exist\n", secretId))
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "Protected Message does not exist, if the error persists please contact with your admin",
			},
		})
	}
}

// OnRevealSecret - Is called when the reveal button is clicked. This method will make the secret available
// to everybody without protection
func (a *Application) OnRevealSecret(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customId := i.MessageComponentData().CustomID
	secretId := strings.Replace(customId, "btn_reveal_", "", -1)
	secret, err := a.Context.SecretManager.GetById(secretId)
	if err != nil {
		message := fmt.Sprintf("An error occurred OnRevealSecret::GetById(%s) %v\n",
			secretId, err)
		a.Context.ErrorLogger.Error(message)
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "An error occurred trying to read secret\n",
			},
		})
		return
	}

	if secret == nil {
		message := fmt.Sprintf("Tried to reveal a secret(%s) that does not exist\n",
			secretId)
		a.Context.DefaultLogger.Warn(message)
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "Can not reveal secret, as it does not exist...",
			},
		})
		return
	}

	a.Context.revealSecretMutex.Lock()
	for reqId, b := range a.Context.revealSecretRequests {
		if b.Secret.SecretId == secretId {
			a.cleanedReq++
			delete(a.Context.revealSecretRequests, reqId)
		}
	}
	a.Context.revealSecretMutex.Unlock()
	m := &discordgo.MessageSend{
		Content: secret.Message,
	}

	if secret.ImageUrl != "" {
		m.Embeds = []*discordgo.MessageEmbed{
			{
				URL: secret.ImageUrl,
			},
		}
	}

	_, err = s.ChannelMessageSendComplex(i.ChannelID, m)

	if err != nil {
		message := fmt.Sprintf("An error occurred OnRevealSecret::ChannelMessageSend %v\n", err)
		a.Context.ErrorLogger.Error(message)
	}

	if err != nil {
		message := fmt.Sprintf("An error occurred OnRevealSecret::ChannelMessageSendReply %v\n", err)
		a.Context.ErrorLogger.Error(message)
	}

	if a.Context.Config.AckActionOnProtectedMessage {
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "Protected message revealed for everyone",
			},
		})

		if err != nil {
			message := fmt.Sprintf("An error occurred OnRevealSecret::ChannelMessageSendReply %v\n", err)
			a.Context.ErrorLogger.Error(message)
		}
	}

	a.deleteSecret(s, i, secret)
}

// OnValidCaptcha - Application implements the interface http server needs to notify us about a user succeeding the
// captcha challenge
func (a *Application) OnValidCaptcha(requestId string) {

	a.Context.revealSecretMutex.Lock()
	if revealReq, found := a.Context.revealSecretRequests[requestId]; !found {
		// Reveal request not found,
		// probably came too late, or a user just forged/tampered the request
		a.Context.revealSecretMutex.Unlock()
		return
	} else {

		a.Context.SessionManager.Authenticate(revealReq.User.Id, revealReq.User.Username)

		secretContent := revealReq.Secret.Message
		if a.Context.PollutionStrategy != nil {
			modifiedMessage, indicators := a.Context.PollutionStrategy.Apply(revealReq.Secret.Message,
				revealReq.User.Username, revealReq.User.Id)
			a.Context.PollutionLogger.Info(fmt.Sprintf("Applied %s strategy, User: %s, Id: %s, Indicators: %s\n",
				a.Context.PollutionStrategy.GetName(),
				fmt.Sprintf("%s#%s", revealReq.User.Username, revealReq.User.Id),
				revealReq.Secret.SecretId, indicators))
			secretContent = modifiedMessage
		}

		var err error
		if a.EditOriginalChallenge {
			_, err = revealReq.Session.InteractionResponseEdit(revealReq.Interaction, &discordgo.WebhookEdit{
				Content: &secretContent,
			})
		} else {
			err = revealReq.Session.InteractionRespond(revealReq.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   discordgo.MessageFlagsEphemeral,
					Content: secretContent,
				},
			})
		}

		if err != nil {
			a.Context.ErrorLogger.Error(fmt.Sprintf("An error occurred on OnValidCaptcha::InteractionResponseEdit - %v\n", err))
		}

		// The request is fulfilled, delete it
		a.cleanedReq++
		delete(a.Context.revealSecretRequests, requestId)
		a.Context.revealSecretMutex.Unlock()
	}
}

// DoMaintenance will perform maintenance tasks such as clean up the HTTP requests to solve the captcha challenge
// we may be expecting
func (a *Application) DoMaintenance() {
	for {
		a.Context.revealSecretMutex.Lock()
		now := time.Now().UTC()
		for reqId, b := range a.Context.revealSecretRequests {
			if now.Sub(b.UpdatedAt) >= time.Minute*10 {
				a.cleanedReq++
				delete(a.Context.revealSecretRequests, reqId)
			}
		}

		if a.cleanedReq >= 500 {
			a.Context.revealSecretRequests = utils.RotateMap(a.Context.revealSecretRequests)
			a.cleanedReq = 0
		}

		a.Context.revealSecretMutex.Unlock()
		time.Sleep(time.Second * 20)
	}
}

// deleteSecret - will delete the message containing the unlock button, the secret from the datasource, and
// the message shown to the creator that contains the delete button itself.
func (a *Application) deleteSecret(s *discordgo.Session, i *discordgo.InteractionCreate, secret *secrets.Secret) {
	// 1. Delete the Message containing the Unlock button
	if secret.MessageId != "" {
		err := s.ChannelMessageDelete(secret.ChannelId, secret.MessageId)
		if err != nil {
			a.Context.ErrorLogger.Error(fmt.Sprintf("An error occurred on deleteSecret::ChannelMessageDelete %v\n",
				err))
		}
	}

	// 2. Delete the secret from the datasource (memory +/- database)
	err := a.Context.SecretManager.Delete(secret.SecretId)
	if err != nil {
		a.Context.ErrorLogger.Error(fmt.Sprintf("An error occurred on deleteSecret::Delete %v\n",
			err))
	}

	// 3. Delete the message shown to the secret creator on creation time.
	// The issue is as of now we only can do so if we have the Secret.Interaction field filled
	// and that only happens when the secret was created during this Application lifetime.
	// If the Application is restarted, then we will not have the Interaction field anymore.
	// I must think about how to fix it, if possible.
	if secret.Interaction != nil {
		err = s.InteractionResponseDelete(secret.Interaction)
		if err != nil {
			a.Context.ErrorLogger.Error(fmt.Sprintf("An error occurred on "+
				"deleteSecret::InteractionResponseDelete %v\n",
				err))
		}
	} else {
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})

		if err != nil {
			a.Context.ErrorLogger.Error(fmt.Sprintf("An error occurred on "+
				"deleteSecret::InteractionResponseDelete %v\n",
				err))
		}
	}
}
