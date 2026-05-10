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
	// Server per Render (Keep-alive)
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, "Bot Ville & Aziende Online") })
		port := os.Getenv("PORT")
		if port == "" { port = "8080" }
		http.ListenAndServe(":"+port, nil)
	}()

	token := os.Getenv("DISCORD_TOKEN")
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Errore sessione: %v", err)
	}

	// CONFIGURAZIONE ID
	serverID      := "1495170590947016996"
	supervisoreID := "1502996558340296754"
	staffRole1    := "1495179869061906602"
	staffRole2    := "1495179516094709850"

	// 1. DEFINIZIONE TUTTI I COMANDI SLASH
	commands := []*discordgo.ApplicationCommand{
		{Name: "setup-assistenza", Description: "Pannello Ticket Ville/Aziende"},
		{Name: "regolamento", Description: "Regolamento degli asset"},
		{Name: "villa-allarme", Description: "Attiva sirena e avvisa lo staff"},
		{Name: "villa-keys", Description: "Gestione chiavi", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionUser, Name: "utente", Description: "Utente target", Required: true},
		}},
		{Name: "biz-annuncio", Description: "Notifica aziendale", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "messaggio", Description: "Testo annuncio", Required: true},
		}},
		{Name: "contratto", Description: "Genera contratto casa", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionInteger, Name: "numero", Description: "Civico", Required: true},
			{Type: discordgo.ApplicationCommandOptionString, Name: "tipo", Description: "Grandezza", Required: true},
		}},
	}

	// 2. GESTORE COMANDI TESTUALI (!reclama e !chiudi)
	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.Bot { return }
		content := strings.ToLower(m.Content)

		if content == "!reclama" {
			s.ChannelEditComplex(m.ChannelID, &discordgo.ChannelEdit{
				PermissionOverwrites: []*discordgo.PermissionOverwrite{
					{ID: m.GuildID, Type: discordgo.PermissionOverwriteTypeRole, Deny: 1024},
					{ID: m.Author.ID, Type: discordgo.PermissionOverwriteTypeMember, Allow: 3072},
					{ID: supervisoreID, Type: discordgo.PermissionOverwriteTypeRole, Allow: 3072},
				},
			})
			s.ChannelMessageSend(m.ChannelID, "✅ **TICKET RECLAMATO**\nOperatore: " + m.Author.Mention() + "\nAccesso limitato a: Staffer e Supervisori.")
		}

		if content == "!chiudi" {
			s.ChannelMessageSend(m.ChannelID, "🔒 **CHIUSURA**\nEliminazione tra 5 secondi...")
			time.Sleep(5 * time.Second)
			s.ChannelDelete(m.ChannelID)
		}
	})

	// 3. GESTORE INTERAZIONI (Slash e Pulsanti)
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		
		// Gestione Slash Commands
		if i.Type == discordgo.InteractionApplicationCommand {
			data := i.ApplicationCommandData()
			switch data.Name {
			case "setup-assistenza":
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: 4,
					Data: &discordgo.InteractionResponseData{
						Content: "🏢 **PANNELLO TICKET**\nSeleziona una categoria per aprire un ticket.",
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{Components: []discordgo.MessageComponent{
								discordgo.SelectMenu{
									CustomID: "ticket_asset",
									Placeholder: "Scegli categoria...",
									Options: []discordgo.SelectMenuOption{
										{Label: "Acquisto Villa/Azienda", Value: "acquisto"},
										{Label: "Problemi/Gestione", Value: "gestione"},
									},
								},
							}},
						},
					},
				})
			case "regolamento":
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: 4, Data: &discordgo.InteractionResponseData{Content: "📜 **Regolamento:** Rispetta gli asset e lo staff."},
				})
			case "villa-allarme":
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: 4, Data: &discordgo.InteractionResponseData{Content: "🚨 **ALLARME VILLA:** Staff informato!"},
				})
			default:
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: 4, Data: &discordgo.InteractionResponseData{Content: "✅ Comando eseguito."},
				})
			}
		}

		// Apertura Ticket dal Menu
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
				Content: "🎫 **TICKET APERTO**\nCreato da: " + utente.Mention() + "\nStaff: <@&" + staffRole1 + "> <@&" + staffRole2 + ">",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{Components: []discordgo.MessageComponent{
						discordgo.Button{Label: "Reclama ✋", Style: discordgo.PrimaryButton, CustomID: "btn_reclama"},
						discordgo.Button{Label: "Chiudi 🔒", Style: discordgo.DangerButton, CustomID: "btn_chiudi"},
					}},
				},
			})

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: 4, Data: &discordgo.InteractionResponseData{Content: "✅ Ticket: <#"+ch.ID+">", Flags: 64},
			})
		}

		// Gestione Pulsanti Reclama/Chiudi
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
					Type: 4, Data: &discordgo.InteractionResponseData{Content: "✅ Ticket reclamato correttamente!"},
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
	s.ApplicationCommandBulkOverwrite(s.State.User.ID, serverID, commands)
	
	fmt.Println("Bot Ville & Aziende Online! 🚀")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop
}
