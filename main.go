package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

func main() {
	// Server per tenere vivo il bot su Render
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, "Bot Online") })
		port := os.Getenv("PORT")
		if port == "" { port = "8080" }
		http.ListenAndServe(":"+port, nil)
	}()

	token := os.Getenv("DISCORD_TOKEN")
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Errore sessione: %v", err)
	}

	// ID CONFIGURAZIONE
	serverID := "1495170590947016996"
	supervisoreID := "1502996558340296754"
	staffRole1 := "1495179869061906602"
	staffRole2 := "1495179516094709850"

	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.Bot { return }
		content := strings.ToLower(m.Content)

		// COMANDO !RECLAMA
		if content == "!reclama" {
			s.ChannelEditComplex(m.ChannelID, &discordgo.ChannelEdit{
				PermissionOverwrites: []*discordgo.PermissionOverwrite{
					{ID: m.GuildID, Type: discordgo.PermissionOverwriteTypeRole, Deny: 1024},
					{ID: m.Author.ID, Type: discordgo.PermissionOverwriteTypeMember, Allow: 3072},
					{ID: supervisoreID, Type: discordgo.PermissionOverwriteTypeRole, Allow: 3072},
				},
			})
			s.ChannelMessageSend(m.ChannelID, "✅ **TICKET RECLAMATO**\nL'operatore "+m.Author.Mention()+" ha preso in carico la richiesta.\nSolo lui e i supervisori possono ora vedere questo canale.")
		}

		// COMANDO !CHIUDI
		if content == "!chiudi" {
			s.ChannelMessageSend(m.ChannelID, "🔒 **CHIUSURA**\nIl ticket verrà eliminato tra 5 secondi...")
			time.Sleep(5 * time.Second)
			s.ChannelDelete(m.ChannelID)
		}
	})

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// COMANDO SLASH SETUP
		if i.Type == discordgo.InteractionApplicationCommand && i.ApplicationCommandData().Name == "setup-assistenza" {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: 4,
				Data: &discordgo.InteractionResponseData{
					Content: "🏢 **PANNELLO TICKET**\nSeleziona una categoria per aprire un ticket.",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.SelectMenu{
									CustomID: "ticket_asset",
									Placeholder: "Scegli categoria...",
									Options: []discordgo.SelectMenuOption{
										{Label: "Acquisto Asset", Value: "acquisto"},
										{Label: "Supporto", Value: "supporto"},
									},
								},
							},
						},
					},
				},
			})
		}

		// APERTURA TICKET DAL MENU
		if i.Type == discordgo.InteractionMessageComponent && i.MessageComponentData().CustomID == "ticket_asset" {
			utente := i.Member.User
			ch, err := s.GuildChannelCreateComplex(i.GuildID, discordgo.GuildChannelCreateData{
				Name: "ticket-" + utente.Username,
				Type: discordgo.ChannelTypeGuildText,
				PermissionOverwrites: []*discordgo.PermissionOverwrite{
					{ID: i.GuildID, Type: discordgo.PermissionOverwriteTypeRole, Deny: 1024},
					{ID: utente.ID, Type: discordgo.PermissionOverwriteTypeMember, Allow: 3072},
					{ID: staffRole1, Type: discordgo.PermissionOverwriteTypeRole, Allow: 3072},
					{ID: staffRole2, Type: discordgo.PermissionOverwriteTypeRole, Allow: 3072},
					{ID: supervisoreID, Type: discordgo.PermissionOverwriteTypeRole, Allow: 3072},
				},
			})
			if err != nil { return }

			s.ChannelMessageSendComplex(ch.ID, &discordgo.MessageSend{
				Content: "🎫 **TICKET APERTO**\nUtente: "+utente.Mention()+"\n\nStaff: <@&"+staffRole1+"> <@&"+staffRole2+">",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{Label: "Reclama ✋", Style: discordgo.PrimaryButton, CustomID: "btn_reclama"},
							discordgo.Button{Label: "Chiudi 🔒", Style: discordgo.DangerButton, CustomID: "btn_chiudi"},
						},
					},
				},
			})

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: 4, Data: &discordgo.InteractionResponseData{Content: "✅ Ticket creato: <#"+ch.ID+">", Flags: 64},
			})
		}

		// GESTIONE PULSANTI
		if i.Type == discordgo.InteractionMessageComponent {
			if i.MessageComponentData().CustomID == "btn_reclama" {
				s.ChannelEditComplex(i.ChannelID, &discordgo.ChannelEdit{
					PermissionOverwrites: []*discordgo.PermissionOverwrite{
						{ID: i.GuildID, Type: discordgo.PermissionOverwriteTypeRole, Deny: 1024},
						{ID: i.Member.User.ID, Type: discordgo.PermissionOverwriteTypeMember, Allow: 3072},
						{ID: supervisoreID, Type: discordgo.PermissionOverwriteTypeRole, Allow: 3072},
					},
				})
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: 4, Data: &discordgo.InteractionResponseData{Content: "✅ Hai reclamato questo ticket!"},
				})
			}
			if i.MessageComponentData().CustomID == "btn_chiudi" {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: 4, Data: &discordgo.InteractionResponseData{Content: "🔒 Chiusura in corso..."},
				})
				time.Sleep(2 * time.Second)
				s.ChannelDelete(i.ChannelID)
			}
		}
	})

	s.Open()
	s.ApplicationCommandBulkOverwrite(s.State.User.ID, serverID, []*discordgo.ApplicationCommand{
		{Name: "setup-assistenza", Description: "Pannello Ticket"},
		{Name: "regolamento", Description: "Regolamento"},
		{Name: "villa-allarme", Description: "Allarme"},
	})
	
	fmt.Println("Bot Online! 🚀")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop
}
