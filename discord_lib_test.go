package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/melardev/discord-message-protect/utils"
	"log"
	"os"
	"os/signal"
	"strings"
	"testing"
	"time"
)

/*
The code here was just used to experiment and learn how to use the discordgo SDK, it is not real unit test.
*/

func TestSendMessage(t *testing.T) {

}

func TestRichMessage(t *testing.T) {

}

func TestPingPongCommand(t *testing.T) {

}

func TestSendEphemeralInteraction(t *testing.T) {

}

func TestListenOnMessageDetectInteraction(t *testing.T) {
	db, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	onMessageCreated := func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Nothing for now

		err := s.ChannelMessageDelete(m.ChannelID, m.ID)
		for _, c := range m.Components {
			if actionRow, ok := c.(*discordgo.ActionsRow); ok {
				for _, c2 := range actionRow.Components {
					if button, ok2 := c2.(*discordgo.Button); ok2 {
						if strings.HasPrefix(button.CustomID, "btn_unlock_") {

						}
					}
				}
			}
		}

		if err != nil {
			panic(err)
		}

	}
	if err != nil {
		panic(err)
	}

	appId := "1032609488563355678"
	guildId := ""
	db.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is up!")
	})

	protectCommand := func(sess *discordgo.Session, i *discordgo.InteractionCreate) {
		_, err = sess.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
			Content: "Click to Unlock Message",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Click to Unlock",
							CustomID: "unlock_btn",
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
		}

	}

	// Components are part of interactions, so we register InteractionCreate handler
	db.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if i.ApplicationCommandData().Name == "protect" {
				protectCommand(s, i)
			}
		case discordgo.InteractionMessageComponent:
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponsePong,
			})
		}
	})

	db.AddHandler(onMessageCreated)
	_, err = db.ApplicationCommandCreate(appId, guildId, &discordgo.ApplicationCommand{
		Name:        "buttons",
		Description: "Test the buttons if you got courage",
	})

	if err != nil {
		log.Fatalf("Cannot create slash command: %v", err)
	}

	err = db.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer db.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Graceful shutdown")
}

func TestButtonInteraction(t *testing.T) {
	// https://discord.com/api/oauth2/authorize?client_id=1032609488563355678&permissions=8&scope=bot
	s, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))

	if err != nil {
		panic(err)
	}

	appId := "1032609488563355678"
	guildId := ""

	componentsHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"fd_no": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Huh. I see, maybe some of these resources might help you?",
					Flags:   discordgo.MessageFlagsEphemeral,
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									Emoji: discordgo.ComponentEmoji{
										Name: "📜",
									},
									Label: "Documentation",
									Style: discordgo.LinkButton,
									URL:   "https://discord.com/developers/docs/interactions/message-components#buttons",
								},
								discordgo.Button{
									Emoji: discordgo.ComponentEmoji{
										Name: "🔧",
									},
									Label: "DiscordConfig developers",
									Style: discordgo.LinkButton,
									URL:   "https://discord.gg/discord-developers",
								},
								discordgo.Button{
									Emoji: discordgo.ComponentEmoji{
										Name: "🦫",
									},
									Label: "DiscordConfig Gophers",
									Style: discordgo.LinkButton,
									URL:   "https://discord.gg/7RuRrVHyXF",
								},
							},
						},
					},
				},
			})
			if err != nil {
				panic(err)
			}
		},
		"fd_yes": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Great! If you wanna know more or just have questions, feel free to visit DiscordConfig Devs and DiscordConfig Gophers server. " +
						"But now, when you know how buttons work, let's move onto select menus (execute `/selects single`)",
					Flags: discordgo.MessageFlagsEphemeral,
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									Emoji: discordgo.ComponentEmoji{
										Name: "🔧",
									},
									Label: "DiscordConfig developers",
									Style: discordgo.LinkButton,
									URL:   "https://discord.gg/discord-developers",
								},
								discordgo.Button{
									Emoji: discordgo.ComponentEmoji{
										Name: "🦫",
									},
									Label: "DiscordConfig Gophers",
									Style: discordgo.LinkButton,
									URL:   "https://discord.gg/7RuRrVHyXF",
								},
							},
						},
					},
				},
			})
			if err != nil {
				panic(err)
			}
		},
		"select": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var response *discordgo.InteractionResponse

			data := i.MessageComponentData()
			switch data.Values[0] {
			case "go":
				response = &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "This is the way.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				}
			default:
				response = &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "It is not the way to go.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				}
			}
			err := s.InteractionRespond(i.Interaction, response)
			if err != nil {
				panic(err)
			}
			time.Sleep(time.Second) // Doing that so user won't see instant response.
			_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "Anyways, now when you know how to use single select menus, let's see how multi select menus work. " +
					"Try calling `/selects multi` command.",
				Flags: discordgo.MessageFlagsEphemeral,
			})
			if err != nil {
				panic(err)
			}
		},
		"stackoverflow_tags": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			data := i.MessageComponentData()

			const stackoverflowFormat = `https://stackoverflow.com/questions/tagged/%s`

			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Here is your stackoverflow URL: " + fmt.Sprintf(stackoverflowFormat, strings.Join(data.Values, "+")),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				panic(err)
			}
			time.Sleep(time.Second) // Doing that so user won't see instant response.
			_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "Now you know everything about select component. If you want to know more or ask a question - feel free to.",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{
								Emoji: discordgo.ComponentEmoji{
									Name: "📜",
								},
								Label: "Documentation",
								Style: discordgo.LinkButton,
								URL:   "https://discord.com/developers/docs/interactions/message-components#select-menus",
							},
							discordgo.Button{
								Emoji: discordgo.ComponentEmoji{
									Name: "🔧",
								},
								Label: "DiscordConfig developers",
								Style: discordgo.LinkButton,
								URL:   "https://discord.gg/discord-developers",
							},
							discordgo.Button{
								Emoji: discordgo.ComponentEmoji{
									Name: "🦫",
								},
								Label: "DiscordConfig Gophers",
								Style: discordgo.LinkButton,
								URL:   "https://discord.gg/7RuRrVHyXF",
							},
						},
					},
				},
				Flags: discordgo.MessageFlagsEphemeral,
			})
			if err != nil {
				panic(err)
			}
		},
	}

	commandsHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"buttons": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Are you comfortable with buttons and other message components?",
					Flags:   discordgo.MessageFlagsEphemeral,
					// Buttons and other components are specified in Components field.
					Components: []discordgo.MessageComponent{
						// ActionRow is a container of all buttons within the same row.
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									// Label is what the user will see on the button.
									Label: "Yes",
									// Style provides coloring of the button. There are not so many styles tho.
									Style: discordgo.SuccessButton,
									// Disabled allows bot to disable some buttons for users.
									Disabled: false,
									// CustomID is a thing telling DiscordConfig which data to send when this button will be pressed.
									CustomID: "fd_yes",
								},
								discordgo.Button{
									Label:    "No",
									Style:    discordgo.DangerButton,
									Disabled: false,
									CustomID: "fd_no",
								},
								discordgo.Button{
									Label:    "I don't know",
									Style:    discordgo.LinkButton,
									Disabled: false,
									// Link buttons don't require CustomID and do not trigger the gateway/HTTP event
									URL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
									Emoji: discordgo.ComponentEmoji{
										Name: "🤷",
									},
								},
							},
						},
						// The message may have multiple actions rows.
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									Label:    "DiscordConfig Developers server",
									Style:    discordgo.LinkButton,
									Disabled: false,
									URL:      "https://discord.gg/discord-developers",
								},
							},
						},
					},
				},
			})
			if err != nil {
				panic(err)
			}
		},
		"selects": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var response *discordgo.InteractionResponse
			switch i.ApplicationCommandData().Options[0].Name {
			case "single":
				response = &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Now let's take a look on selects. This is single item select menu.",
						Flags:   discordgo.MessageFlagsEphemeral,
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{
								Components: []discordgo.MessageComponent{
									discordgo.SelectMenu{
										// Select menu, as other components, must have a customID, so we set it to this value.
										CustomID:    "select",
										Placeholder: "Choose your favorite programming language 👇",
										Options: []discordgo.SelectMenuOption{
											{
												Label: "Go",
												// As with components, this things must have their own unique "id" to identify which is which.
												// In this case such id is Value field.
												Value: "go",
												Emoji: discordgo.ComponentEmoji{
													Name: "🦦",
												},
												// You can also make it a default option, but in this case we won't.
												Default:     false,
												Description: "Go programming language",
											},
											{
												Label: "JS",
												Value: "js",
												Emoji: discordgo.ComponentEmoji{
													Name: "🟨",
												},
												Description: "JavaScript programming language",
											},
											{
												Label: "Python",
												Value: "py",
												Emoji: discordgo.ComponentEmoji{
													Name: "🐍",
												},
												Description: "Python programming language",
											},
										},
									},
								},
							},
						},
					},
				}
			case "multi":
				minValues := 1
				response = &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "The tastiest things are left for the end. Let's see how the multi-item select menu works: " +
							"try generating your own stackoverflow search link",
						Flags: discordgo.MessageFlagsEphemeral,
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{
								Components: []discordgo.MessageComponent{
									discordgo.SelectMenu{
										CustomID:    "stackoverflow_tags",
										Placeholder: "Select tags to search on StackOverflow",
										// This is where confusion comes from. If you don't specify these things you will get single item select.
										// These fields control the minimum and maximum amount of selected items.
										MinValues: &minValues,
										MaxValues: 3,
										Options: []discordgo.SelectMenuOption{
											{
												Label:       "Go",
												Description: "Simple yet powerful programming language",
												Value:       "go",
												// Default works the same for multi-select menus.
												Default: false,
												Emoji: discordgo.ComponentEmoji{
													Name: "🦦",
												},
											},
											{
												Label:       "JS",
												Description: "Multiparadigm OOP language",
												Value:       "javascript",
												Emoji: discordgo.ComponentEmoji{
													Name: "🟨",
												},
											},
											{
												Label:       "Python",
												Description: "OOP prototyping programming language",
												Value:       "python",
												Emoji: discordgo.ComponentEmoji{
													Name: "🐍",
												},
											},
											{
												Label:       "Web",
												Description: "Web related technologies",
												Value:       "web",
												Emoji: discordgo.ComponentEmoji{
													Name: "🌐",
												},
											},
											{
												Label:       "Desktop",
												Description: "Desktop applications",
												Value:       "desktop",
												Emoji: discordgo.ComponentEmoji{
													Name: "💻",
												},
											},
										},
									},
								},
							},
						},
					},
				}

			}
			err := s.InteractionRespond(i.Interaction, response)
			if err != nil {
				panic(err)
			}
		},
	}

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is up!")
	})
	// Components are part of interactions, so we register InteractionCreate handler
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := commandsHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		case discordgo.InteractionMessageComponent:

			if h, ok := componentsHandlers[i.MessageComponentData().CustomID]; ok {
				h(s, i)
			}
		}
	})

	_, err = s.ApplicationCommandCreate(appId, guildId, &discordgo.ApplicationCommand{
		Name:        "buttons",
		Description: "Test the buttons if you got courage",
	})

	if err != nil {
		log.Fatalf("Cannot create slash command: %v", err)
	}

	_, err = s.ApplicationCommandCreate(appId, guildId, &discordgo.ApplicationCommand{
		Name: "selects",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "multi",
				Description: "Multi-item select menu",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "single",
				Description: "Single-item select menu",
			},
		},
		Description: "Lo and behold: dropdowns are coming",
	})

	if err != nil {
		log.Fatalf("Cannot create slash command: %v", err)
	}

	err = s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Graceful shutdown")
}

func TestFlow(t *testing.T) {
	// https://discord.com/api/oauth2/authorize?client_id=1032609488563355678&permissions=8&scope=bot
	s, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))

	if err != nil {
		panic(err)
	}

	appId := "1032609488563355678"
	guildId := ""

	secrets := map[string]string{}

	handler := func(sess *discordgo.Session, i *discordgo.InteractionCreate) {
		// interaction tokens are valid for 3 seconds, we must respond in 3 seconds
		// to keep them valid, if not, then interaction is invalidated.
		// if we want the interaction response to be delayed more than that, we use deferred signals
		// we can take up to 15 minutes to respond.
		customId := i.MessageComponentData().CustomID
		if !strings.HasPrefix(customId, "unlock_btn_") {
			err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   discordgo.MessageFlagsEphemeral,
					Content: "I do not know what happened",
				},
			})
			return
		}

		secretId := strings.Replace(customId, "unlock_btn_", "", -1)
		if secret, secretFound := secrets[secretId]; secretFound {
			fmt.Printf("Secret found %s %s\n", secretId, secret)
			err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   discordgo.MessageFlagsEphemeral,
					Content: secret,
				},
			})

			if err != nil {
				panic(err)
			}
		} else {
			err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   discordgo.MessageFlagsEphemeral,
					Content: "Hmmmm secret not found!!",
				},
			})
		}
	}

	commandsHandlers := map[string]func(sess *discordgo.Session, i *discordgo.InteractionCreate){
		"secret": func(sess *discordgo.Session, i *discordgo.InteractionCreate) {
			// First send the interaction response, we have only 3 seconds and if the internet is slow
			// we may waste it sending the message rather than the interaction below.
			err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   discordgo.MessageFlagsEphemeral,
					Content: "Users are notified about secret, they must click on the unlock button now",
				},
			})

			if err != nil {
				panic(err)
			}

			messageContent := ""
			if data, ok := i.Interaction.Data.(discordgo.ApplicationCommandInteractionData); ok {
				for _, option := range data.Options {
					if option.Name == "content" {
						messageContent = option.Value.(string)
					}

					if option.Name == "images" {
						fmt.Printf("We got some image! %#v\n", option.Value)
					}
				}
			}

			secretId := utils.GetRandomString(12)
			secrets[secretId] = messageContent

			/*
				_, err = s.ChannelMessageSend(i.ChannelID, "Hello there!!!")
					if err != nil {
						panic(err)
					}
					return
			*/

			_, err = sess.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
				Content: "Click to Unlock",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{
								Label:    "Click to Unlock",
								CustomID: fmt.Sprintf("unlock_btn_%s", secretId),
								//URL:      "https://google.com/",
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
			}
		},
	}

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is up!")
	})
	// Components are part of interactions, so we register InteractionCreate handler
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := commandsHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		case discordgo.InteractionMessageComponent:
			handler(s, i)

		}
	})

	_, err = s.ApplicationCommandCreate(appId, guildId, &discordgo.ApplicationCommand{
		Name:        "secret",
		Description: "Test the buttons if you got courage",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "content",
				Description: "Content",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Name:        "images",
				Description: "Content",
				Required:    false,
			},
		},
	})

	if err != nil {
		log.Fatalf("Cannot create slash command: %v", err)
	}

	err = s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Graceful shutdown")
}

func TestFlowLinkThenButton(t *testing.T) {
	// https://discord.com/api/oauth2/authorize?client_id=1032609488563355678&permissions=8&scope=bot
	s, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))

	if err != nil {
		panic(err)
	}

	appId := "1032609488563355678"
	guildId := ""

	secrets := map[string]string{}
	users := map[string]time.Time{}

	handler := func(sess *discordgo.Session, i *discordgo.InteractionCreate) {
		// interaction tokens are valid for 3 seconds, we must respond in 3 seconds
		// to keep them valid, if not, then interaction is invalidated.
		// if we want the interaction response to be delayed more than that, we use deferred signals
		// we can take up to 15 minutes to respond.
		customId := i.MessageComponentData().CustomID
		if !strings.HasPrefix(customId, "unlock_btn_") {
			err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   discordgo.MessageFlagsEphemeral,
					Content: "I do not know what happened",
				},
			})
			return
		}

		secretId := strings.Replace(customId, "unlock_btn_", "", -1)
		if secret, secretFound := secrets[secretId]; secretFound {
			showChallenge := false
			if lastLogged, userFound := users[s.State.User.ID]; userFound {
				if time.Now().UTC().Sub(lastLogged) > time.Minute {
					fmt.Printf("Secret found %s %s but user login status expired\n", secretId, secret)
					showChallenge = true
				} else {
					fmt.Printf("Secret found, user logged in\n")
					showChallenge = false
				}
			} else {
				fmt.Printf("Secret found, user not logged in\n")
				showChallenge = true
				users[s.State.User.ID] = time.Now().UTC()
			}

			if showChallenge {
				err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Flags:   discordgo.MessageFlagsEphemeral,
						Content: "",
						Embeds: []*discordgo.MessageEmbed{
							{
								Type:        discordgo.EmbedTypeLink,
								Title:       "Security",
								Description: "Please click the following link, solve the captcha, then click on Unlock - https://yahoo.com",
								URL:         "https://google.com/",
							},
						},
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{
								Components: []discordgo.MessageComponent{
									discordgo.Button{
										Label:    "Click to Unlock",
										CustomID: fmt.Sprintf("unlock_btn_%s", secretId),
										//URL:      "https://google.com/",
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
					},
				})

				if err != nil {
					panic(err)
				}
			} else {
				err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Flags:   discordgo.MessageFlagsEphemeral,
						Content: secret,
					},
				})

				if err != nil {
					panic(err)
				}
			}

		} else {
			err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   discordgo.MessageFlagsEphemeral,
					Content: "Hmmmm secret not found!!",
				},
			})
		}
	}

	commandsHandlers := map[string]func(sess *discordgo.Session, i *discordgo.InteractionCreate){
		"secret": func(sess *discordgo.Session, i *discordgo.InteractionCreate) {
			// First send the interaction response, we have only 3 seconds and if the internet is slow
			// we may waste it sending the message rather than the interaction below.
			err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   discordgo.MessageFlagsEphemeral,
					Content: "Users are notified about secret, they must click on the unlock button now",
				},
			})

			if err != nil {
				panic(err)
			}

			messageContent := ""
			if data, ok := i.Interaction.Data.(discordgo.ApplicationCommandInteractionData); ok {
				for _, option := range data.Options {
					if option.Name == "content" {
						messageContent = option.Value.(string)
					}

					if option.Name == "images" {
						fmt.Printf("We got some image! %#v\n", option.Value)
					}
				}
			}

			secretId := utils.GetRandomString(12)
			secrets[secretId] = messageContent

			/*
				_, err = s.ChannelMessageSend(i.ChannelID, "Hello there!!!")
					if err != nil {
						panic(err)
					}
					return
			*/

			_, err = sess.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
				Content: "Click to Unlock",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{
								Label:    "Click to Unlock",
								CustomID: fmt.Sprintf("unlock_btn_%s", secretId),
								//URL:      "https://google.com/",
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
			}
		},
	}

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is up!")
	})
	// Components are part of interactions, so we register InteractionCreate handler
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := commandsHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		case discordgo.InteractionMessageComponent:
			handler(s, i)
		}
	})

	_, err = s.ApplicationCommandCreate(appId, guildId, &discordgo.ApplicationCommand{
		Name:        "secret",
		Description: "Test the buttons if you got courage",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "content",
				Description: "Content",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Name:        "images",
				Description: "Content",
				Required:    false,
			},
		},
	})

	if err != nil {
		log.Fatalf("Cannot create slash command: %v", err)
	}

	err = s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Graceful shutdown")
}

func TestMultiInteractionUseForEdit(t *testing.T) {
	// https://discord.com/api/oauth2/authorize?client_id=1032609488563355678&permissions=8&scope=bot
	s, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))

	if err != nil {
		panic(err)
	}

	appId := "1032609488563355678"
	guildId := ""

	// users := map[string]*discordgo.Interaction{}
	var protectMessage string

	handler := func(sess *discordgo.Session, i *discordgo.InteractionCreate) {
		customId := i.MessageComponentData().CustomID
		if customId == "unlock_btn" {
			err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags: discordgo.MessageFlagsEphemeral,
					Embeds: []*discordgo.MessageEmbed{
						{
							Type:        discordgo.EmbedTypeLink,
							Title:       "Security",
							Description: "Please click the following link, solve the captcha, then click on Unlock - https://yahoo.com",
							URL:         "https://google.com/",
						},
					},
				},
			})

			if err != nil {
				panic(err)
			}
			// Mimic the user has solved the captcha, the captcha app has notified the discord bot through
			// messaging architecture that the user is not a bot, now reveal the message
			go func(session *discordgo.Session, interaction *discordgo.Interaction) {
				time.Sleep(time.Second * 3)
				_, err = session.InteractionResponseEdit(interaction, &discordgo.WebhookEdit{
					Content: &protectMessage,
					Embeds:  nil,
				})

				if err != nil {
					panic(err)
				}
			}(s, i.Interaction)
		}
	}

	protectCommand := func(sess *discordgo.Session, i *discordgo.InteractionCreate) {
		// First send the interaction response, we have only 3 seconds and if the internet is slow
		// we may waste it sending the message rather than the interaction below.
		err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "Users are notified about secret, they must click on the unlock button now",
			},
		})

		if err != nil {
			panic(err)
		}

		messageContent := ""
		if data, ok := i.Interaction.Data.(discordgo.ApplicationCommandInteractionData); ok {
			for _, option := range data.Options {
				if option.Name == "content" {
					messageContent = option.Value.(string)
				}

				if option.Name == "images" {
					fmt.Printf("We got some image! %#v\n", option.Value)
				}
			}
		}

		protectMessage = messageContent

		_, err = sess.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
			Content: "Click to Unlock Message",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Click to Unlock",
							CustomID: "unlock_btn",
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
		}
	}

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is up!")
	})
	// Components are part of interactions, so we register InteractionCreate handler
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if i.ApplicationCommandData().Name == "protect" {
				protectCommand(s, i)
			}
		case discordgo.InteractionMessageComponent:
			handler(s, i)
		}
	})

	_, err = s.ApplicationCommandCreate(appId, guildId, &discordgo.ApplicationCommand{
		Name:        "protect",
		Description: "Protect messages with anti-bot mechanisms",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "content",
				Description: "Content",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Name:        "images",
				Description: "Content",
				Required:    false,
			},
		},
	})

	if err != nil {
		log.Fatalf("Cannot create slash command: %v", err)
	}

	err = s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Graceful shutdown")
}

func TestMultiInteractionUseForEdit2(t *testing.T) {
	// https://discord.com/api/oauth2/authorize?client_id=1032609488563355678&permissions=8&scope=bot
	s, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))

	if err != nil {
		panic(err)
	}

	appId := "1032609488563355678"
	guildId := ""

	// users := map[string]*discordgo.Interaction{}
	var protectMessage string

	handler := func(sess *discordgo.Session, i *discordgo.InteractionCreate) {
		customId := i.MessageComponentData().CustomID
		if customId == "unlock_btn" {
			err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags: discordgo.MessageFlagsEphemeral,
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									Label:    "Click to Unlock",
									Style:    discordgo.LinkButton,
									Disabled: false,
									URL:      "https://google.com/",
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
				panic(err)
			}
			// Mimic the user has solved the captcha, the captcha app has notified the discord bot through
			// messaging architecture that the user is not a bot, now reveal the message
			go func(interaction *discordgo.Interaction) {
				time.Sleep(time.Second * 3)
				// Please note that the global bot session does not work in real app, for some reason
				// in this unit test it does, probably because it happens too quick??
				_, err = s.InteractionResponseEdit(interaction, &discordgo.WebhookEdit{
					Content: &protectMessage,
					Embeds:  nil,
				})

				if err != nil {
					panic(err)
				}
			}(i.Interaction)
		}
	}

	protectCommand := func(sess *discordgo.Session, i *discordgo.InteractionCreate) {
		// First send the interaction response, we have only 3 seconds and if the internet is slow
		// we may waste it sending the message rather than the interaction below.
		err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "Users are notified about secret, they must click on the unlock button now",
			},
		})

		if err != nil {
			panic(err)
		}

		messageContent := ""
		if data, ok := i.Interaction.Data.(discordgo.ApplicationCommandInteractionData); ok {
			for _, option := range data.Options {
				if option.Name == "content" {
					messageContent = option.Value.(string)
				}

				if option.Name == "images" {
					fmt.Printf("We got some image! %#v\n", option.Value)
				}
			}
		}

		protectMessage = messageContent

		_, err = sess.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
			Content: "Click to Unlock Message",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Click to Unlock",
							CustomID: "unlock_btn",
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
		}
	}

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is up!")
	})
	// Components are part of interactions, so we register InteractionCreate handler
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if i.ApplicationCommandData().Name == "protect" {
				protectCommand(s, i)
			}
		case discordgo.InteractionMessageComponent:
			handler(s, i)
		}
	})

	_, err = s.ApplicationCommandCreate(appId, guildId, &discordgo.ApplicationCommand{
		Name:        "protect",
		Description: "Protect messages with anti-bot mechanisms",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "content",
				Description: "Content",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Name:        "images",
				Description: "Content",
				Required:    false,
			},
		},
	})

	if err != nil {
		log.Fatalf("Cannot create slash command: %v", err)
	}

	err = s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Graceful shutdown")
}

func TestButtonDeferredInteraction(t *testing.T) {

	// https://discord.com/api/oauth2/authorize?client_id=1032609488563355678&permissions=8&scope=bot
	s, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))

	if err != nil {
		panic(err)
	}

	appId := "1032609488563355678"
	guildId := ""

	var lastAuthenticated time.Time

	commandsHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"secret": func(sess *discordgo.Session, i *discordgo.InteractionCreate) {
			if time.Now().UTC().Sub(lastAuthenticated) > time.Minute*1 {
				// interaction tokens are valid for 3 seconds, we must respond in 3 seconds
				// to keep them valid, if not, then interaction is invalidated.
				// if we want the interaction response to be delayed more than that, we use deferred signals
				// we can take up to 15 minutes to respond.

				err := sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseDeferredMessageUpdate,
					Data: &discordgo.InteractionResponseData{
						Flags: discordgo.MessageFlagsEphemeral,
						// Buttons and other components are specified in Components field.
						Components: []discordgo.MessageComponent{
							// ActionRow is a container of all buttons within the same row.
							discordgo.Button{
								Label:    "Reveal Secret",
								Style:    discordgo.LinkButton,
								Disabled: false,
								// Link buttons don't require CustomID and do not trigger the gateway/HTTP event
								URL: "https://google.com/",
								Emoji: discordgo.ComponentEmoji{
									Name:     "🔒",
									Animated: false,
								},
							},
						},
					},
				})

				if err != nil {
					panic(err)
				}

				go func(i *discordgo.Interaction) {
					fmt.Printf("Waiting for challenge to be resolved!!")
					time.Sleep(time.Second * 5)
					fmt.Printf("CaptchaService solved!!")
					err = s.InteractionRespond(i, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Flags:   discordgo.MessageFlagsEphemeral,
							Content: "Secret revealed!!",
							// Buttons and other components are specified in Components field.
						},
					})
					if err != nil {
						panic(err)
					}
				}(i.Interaction)
			} else {
				lastAuthenticated = time.Now().UTC()
			}
		},
	}

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is up!")
	})

	// Components are part of interactions, so we register InteractionCreate handler
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := commandsHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		case discordgo.InteractionMessageComponent:
		}
	})

	_, err = s.ApplicationCommandCreate(appId, guildId, &discordgo.ApplicationCommand{
		Name:        "secret",
		Description: "Reveal secret",
	})

	if err != nil {
		log.Fatalf("Cannot create slash command: %v", err)
	}

	err = s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Graceful shutdown")

}

func TestUrlButtonInteraction(t *testing.T) {

	// https://discord.com/api/oauth2/authorize?client_id=1032609488563355678&permissions=8&scope=bot
	s, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))

	if err != nil {
		panic(err)
	}

	// Bot parameters
	appId := "1032609488563355678"

	// Important note: call every command in order it's placed in the example.
	commandsHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"secret": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			messageContent := ""
			if data, ok := i.Interaction.Data.(discordgo.ApplicationCommandInteractionData); ok {
				for _, option := range data.Options {
					if option.Name == "content" {
						messageContent = option.Value.(string)
					}

					if option.Name == "images" {
						fmt.Printf("We got some image! %#v\n", option.Value)
					}
				}
			}

			fmt.Println(messageContent)
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "",
					Flags:   discordgo.MessageFlagsEphemeral,
					// Buttons and other components are specified in Components field.
					Components: []discordgo.MessageComponent{
						// ActionRow is a container of all buttons within the same row.
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									// Label is what the user will see on the button.
									Label: "Resolve Challenge",
									// Style provides coloring of the button. There are not so many styles tho.
									Style: discordgo.LinkButton,
									// Disabled allows bot to disable some buttons for users.
									Disabled: false,
									URL:      "https://google.com/",
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
				panic(err)
			}
		},
	}

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is up!")
	})

	// Components are part of interactions, so we register InteractionCreate handler
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := commandsHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		}
	})

	_, err = s.ApplicationCommandCreate(appId, "", &discordgo.ApplicationCommand{
		Name:        "secret",
		Description: "Secret",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "content",
				Description: "Content",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Name:        "images",
				Description: "Content",
				Required:    false,
			},
		},
	})

	if err != nil {
		log.Fatalf("Cannot create slash command: %v", err)
	}

	if err != nil {
		log.Fatalf("Cannot create slash command: %v", err)
	}

	err = s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Graceful shutdown")
}

func TestEditMessageUsingId(t *testing.T) {

}
