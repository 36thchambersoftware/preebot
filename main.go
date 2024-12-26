package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"

	mongo "preebot/pkg/db"
	"preebot/pkg/discord"

	mongodb "go.mongodb.org/mongo-driver/mongo"

	"github.com/bwmarrin/discordgo"
)

var (
	mdb *mongodb.Client
	dbctx context.Context
	dbcancel context.CancelFunc
)

var (
	integerOptionMinValue          = 1.0
	dmPermission                   = false
	defaultMemberPermissions int64 = discordgo.PermissionManageServer

	commands = []*discordgo.ApplicationCommand{
		&discord.ENGAGE_ROLE_COMMAND,
		&discord.LINK_WALLET_COMMAND,
		&discord.CHECK_MY_WALLETS_COMMAND,
		&discord.CONFIGURE_POOL_ID_COMMAND,
		&discord.CONFIGURE_POLICY_ID_COMMAND,
		&discord.CONFIGURE_DELEGATOR_ROLE_COMMAND,
		&discord.CONFIGURE_POLICY_ROLE_COMMAND,
		&discord.CONFIGURE_ENGAGE_ROLE_COMMAND,
		&discord.CONFIGURE_CUSTODIAL_COMMAND,
		&discord.LIST_DELEGATOR_ROLES_COMMAND,
		&discord.LIST_POLICY_ROLES_COMMAND,
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		discord.ENGAGE_ROLE_COMMAND.Name:              discord.ENGAGE_ROLE_HANDLER,
		discord.LINK_WALLET_COMMAND.Name:              discord.LINK_WALLET_HANDLER,
		discord.CHECK_MY_WALLETS_COMMAND.Name:         discord.CHECK_MY_WALLETS_HANDLER,
		discord.CONFIGURE_POOL_ID_COMMAND.Name:        discord.CONFIGURE_POOL_ID_HANDLER,
		discord.CONFIGURE_POLICY_ID_COMMAND.Name:      discord.CONFIGURE_POLICY_ID_HANDLER,
		discord.CONFIGURE_DELEGATOR_ROLE_COMMAND.Name: discord.CONFIGURE_DELEGATOR_ROLE_HANDLER,
		discord.CONFIGURE_POLICY_ROLE_COMMAND.Name:    discord.CONFIGURE_POLICY_ROLE_HANDLER,
		discord.CONFIGURE_ENGAGE_ROLE_COMMAND.Name:    discord.CONFIGURE_ENGAGE_ROLE_HANDLER,
		discord.CONFIGURE_CUSTODIAL_COMMAND.Name:      discord.CONFIGURE_CUSTODIAL_HANDLER,
		discord.LIST_DELEGATOR_ROLES_COMMAND.Name:     discord.LIST_DELEGATOR_ROLES_HANDLER,
		discord.LIST_POLICY_ROLES_COMMAND.Name:        discord.LIST_POLICY_ROLES_HANDLER,
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

func init() {
	// Setup DB
    mdb, ctx, cancel, err := mongo.Connect()
    if err != nil {
        panic(err)
    }

	dbctx = ctx
	dbcancel = cancel
	mongo.DB = mdb
}

func main() {
	defer mongo.Close(mdb, dbctx, dbcancel)

	// Setup discord
	discord.S.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if strings.Contains(strings.ToUpper(m.Author.GlobalName), "ANNOUNCEMENTS") || strings.Contains(strings.ToUpper(m.Author.GlobalName), "ADMIN") {
			s.ChannelMessageDelete(m.ChannelID, m.ID)
		}

		if m.Author.Bot {
			return
		}
	})

	// Setup Command Handler
	discord.S.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	discord.S.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	err := discord.S.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := discord.S.ApplicationCommandCreate(discord.S.State.User.ID, *discord.GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer discord.S.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	if *discord.RemoveCommands {
		log.Println("Removing commands...")

		for _, v := range registeredCommands {
			err := discord.S.ApplicationCommandDelete(discord.S.State.User.ID, *discord.GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}

	log.Println("Gracefully shutting down.")
}
