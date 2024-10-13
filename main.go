package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"preebot/pkg/discord"

	"github.com/bwmarrin/discordgo"
)

// Bot parameters
var (
	GuildID        = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	BotToken       = flag.String("token", "", "Bot access token")
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutting down or not")
)

var s *discordgo.Session

func init() { flag.Parse() }

func init() {
	token, ok := os.LookupEnv("PREEBOT_TOKEN")
	if !ok {
		log.Fatalf("Missing token")
	}
	var err error
	s, err = discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

var (
	integerOptionMinValue          = 1.0
	dmPermission                   = false
	defaultMemberPermissions int64 = discordgo.PermissionManageServer

	commands = []*discordgo.ApplicationCommand{
		&discord.ENGAGE_ROLE_COMMAND,
		&discord.LINK_WALLET_COMMAND,
		&discord.CHECK_DELEGATION_COMMAND,
		&discord.CONFIGURE_POOL_ID_COMMAND,
		&discord.CONFIGURE_POLICY_ID_COMMAND,
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		discord.ENGAGE_ROLE_COMMAND.Name:         discord.ENGAGE_ROLE_HANDLER,
		discord.LINK_WALLET_COMMAND.Name:         discord.LINK_WALLET_HANDLER,
		discord.CHECK_DELEGATION_COMMAND.Name:    discord.CHECK_DELEGATION_HANDLER,
		discord.CONFIGURE_POOL_ID_COMMAND.Name:   discord.CONFIGURE_POOL_ID_HANDLER,
		discord.CONFIGURE_POLICY_ID_COMMAND.Name: discord.CONFIGURE_POLICY_ID_HANDLER,
	}
	lockout         = make(map[string]struct{})
	lockoutResponse = &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Please wait for your last command to finish. :D",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}
)

func main() {
	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.GlobalName == "ANNOUNCEMENTS" || m.Author.GlobalName == "ADMIN" {
			s.ChannelMessageDelete(m.ChannelID, m.ID)
		}

		if m.Author.Bot {
			return
		}
	})

	// Setup Command Handler
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			if _, ok := lockout[i.Member.User.ID]; !ok {
				lockout[i.Member.User.ID] = struct{}{}
				defer func() {
					delete(lockout, i.Member.User.ID)
				}()
				h(s, i)
			} else {
				s.InteractionRespond(i.Interaction, lockoutResponse)
			}
		}
	})

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, *GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	if *RemoveCommands {
		log.Println("Removing commands...")

		for _, v := range registeredCommands {
			err := s.ApplicationCommandDelete(s.State.User.ID, *GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}

	log.Println("Gracefully shutting down.")
}
