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
	// Server per Render
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
	serverID := "1495170590947016996"
	supervisoreID := "1502996558340296754"
	staffRole1 := "1495179869061906602"
	staffRole2 := "1495179516094709850"

	// --- FUNZIONI DI LOGICA ---

	// Funzione per reclamare il ticket
	reclamaTicket := func(s *discordgo.Session, channelID string, stafferID string, guildID string) {
		// Otteniamo i permessi attuali per non perdere l'accesso dell'utente che ha aperto il ticket
		ch, _ := s.Channel(channelID)
		
		overwrites := []*discordgo.PermissionOverwrite{
			{ID: guildID, Type: discordgo.PermissionOverwriteTypeRole, Deny: 1024},         // @everyone non vede
			{ID: stafferID, Type: discordgo.PermissionOverwriteTypeMember, Allow: 3072},     // Lo staffer vede
			{ID: supervisoreID, Type: discordgo.PermissionOverwriteTypeRole, Allow: 3072},  // Il supervisore vede
		}

		// Manteniamo l'accesso per chiunque avesse già il permesso esplicito (l'utente che ha aperto il ticket)
		for _, ow := range ch.PermissionOverwrites {
			if ow.Type == discordgo.PermissionOverwriteTypeMember && ow.ID != stafferID {
				overwrites = append(overwrites, ow)
			}
		}

		s.ChannelEditComplex(channelID, &discordgo.ChannelEdit{PermissionOverwrites: overwrites})
		s.ChannelMessageSend(channelID, fmt.Sprintf("✅ **TICKET RECLAMATO**\nL'operatore <@%s> ha preso in carico la richiesta.\n\n*Il supporto è ora limitato allo staffer incaricato e ai supervisori.*", stafferID))
	}

	// Funzione per chiudere il ticket
	chiudiTicket := func(s *discordgo.Session, channelID string) {
		s.ChannelMessageSend(channelID, "🔒 **CHIUSURA TICKET**\nIl canale verrà eliminato tra 5 secondi...")
		time.Sleep(5 * time.Second)
		s.ChannelDelete(channelID)
	}

	// --- HANDLERS ---

	// Gestore Messaggi (!reclama e !chiudi)
	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.Bot { return }
		content := strings.ToLower(m.Content)
		if content == "!reclama" {
			reclamaTicket(s, m.ChannelID, m.Author.ID, m.GuildID)
		} else if content == "!chiudi" {
			chiudiTicket(s, m.ChannelID)
		}
	})

	// Gestore Interazioni (Slash, Menu, Pulsanti)
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		
		// Comandi Slash
		if i.Type == discordgo.InteractionApplicationCommand {
			if i.ApplicationCommandData().Name == "setup-assistenza" {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: 4,
					Data: &discordgo.InteractionResponseData{
						Content: "🏢 **PANNELLO TICKET ASSET**\nSeleziona una categoria per parlare con lo Staff.",
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{Components: []discordgo.MessageComponent{
								discordgo.SelectMenu{
									CustomID: "ticket_asset",
									Placeholder: "Apri un ticket qui...",
									Options: []discordgo.SelectMenuOption{
										{Label: "Acquisto Asset", Value: "acquisto", Emoji: discordgo.ComponentEmoji{Name: "💰"}},
										{Label: "Supporto Tecnico", Value: "supporto", Emoji: discordgo.ComponentEmoji{Name: "🛠️"}},
									},
								},
							}},
						},
					},
				})
			}
		}

		// Creazione Ticket dal Menu
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
				Content: fmt.Sprintf("🎫 **NUOVO TICKET**\nUtente: %s\nStaff: <@&%s> <@&%s>", utente.Mention(), staffRole1, staffRole2),
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{Components: []discordgo.MessageComponent{
						discordgo.Button{Label: "Reclama ✋", Style: discordgo.PrimaryButton, CustomID: "btn_reclama"},
						discordgo.Button{Label: "Chiudi 🔒", Style: discordgo.DangerButton, CustomID: "btn_chiudi"},
					}},
				},
			})

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: 4, Data: &discordgo.InteractionResponseData{Content: "✅ Ticket aperto: <#"+ch.ID+">", Flags: 64},
			})
		}

		// Gestione Pulsanti Reclama/Chiudi
		if i.Type == discordgo.InteractionMessageComponent {
			switch i.MessageComponentData().CustomID {
			case "btn_reclama":
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: 4, Data: &discordgo.InteractionResponseData{Content: "Hai reclamato il ticket.", Flags: 64}})
				reclamaTicket(s, i.ChannelID, i.Member.User.ID, i.GuildID)
			case "btn_chiudi":
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: 4, Data: &discordgo.InteractionResponseData{Content: "Chiusura avviata.", Flags: 64}})
				chiudiTicket(s, i.ChannelID)
			}
		}
	})

	s.Open()
	// Registrazione Comandi
	commands := []*discordgo.ApplicationCommand{
		{Name: "setup-assistenza", Description: "Invia il pannello dei ticket"},
		{Name: "regolamento", Description: "Mostra regolamento asset"},
		{Name: "villa-allarme", Description: "Attiva allarme"},
	}
	s.ApplicationCommandBulkOverwrite(s.State.User.ID, serverID, commands)
	
	fmt.Println("Bot Ville & Aziende Online! 🚀")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop
}
