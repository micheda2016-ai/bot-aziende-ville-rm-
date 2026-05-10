package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func main() {
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

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "setup-assistenza",
			Description: "Invia il pannello Ticket Ville/Aziende",
		},
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			if i.ApplicationCommandData().Name == "setup-assistenza" {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: 4,
					Data: &discordgo.InteractionResponseData{
						Content: "🏢 **PANNELLO GESTIONE ASSET & VILLE**\nSeleziona una categoria qui sotto per aprire un ticket.",
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{Components: []discordgo.MessageComponent{
								discordgo.SelectMenu{
									CustomID:    "ticket_asset",
									Placeholder: "Scegli il motivo della richiesta...",
									Options: []discordgo.SelectMenuOption{
										{Label: "Acquisto Villa / Azienda", Value: "acquisto_asset"},
										{Label: "Problemi Tecnici / Gestione", Value: "gestione_asset"},
									},
								},
							}},
						},
					},
				})
			}
		}

		if i.Type == discordgo.InteractionMessageComponent && i.MessageComponentData().CustomID == "ticket_asset" {
			categoria := i.MessageComponentData().Values[0]
			utente := i.Member.User

			ch, err := s.GuildChannelCreateComplex(i.GuildID, discordgo.GuildChannelCreateData{
				Name: "ticket-" + utente.Username,
				Type: discordgo.ChannelTypeGuildText,
				PermissionOverwrites: []*discordgo.PermissionOverwrite{
					{ID: i.GuildID, Type: 0, Deny: 1024},
					{ID: utente.ID, Type: 1, Allow: 3072},
				},
			})
			if err != nil { return }

			messaggio := fmt.Sprintf("🎫 **TICKET CREATO**\n\n**Creato da:** %s\n**Categoria:** %s\n\nAttenzione Staff: <@&1495179869061906602> <@&1495179516094709850>", utente.Mention(), categoria)

			s.ChannelMessageSendComplex(ch.ID, &discordgo.MessageSend{
				Content: messaggio,
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Reclama ✋",
							Style:    discordgo.PrimaryButton,
							CustomID: "reclama_ticket_" + utente.ID,
						},
					}},
				},
			})

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: 4,
				Data: &discordgo.InteractionResponseData{Content: "✅ Ticket aperto: <#"+ch.ID+">", Flags: 64},
			})
		}

		if i.Type == discordgo.InteractionMessageComponent && len(i.MessageComponentData().CustomID) > 15 && i.MessageComponentData().CustomID[:15] == "reclama_ticket_" {
			staffer := i.Member.User
			creatoreID := i.MessageComponentData().CustomID[15:]
			supervisoreID := "1502996558340296754"

			s.ChannelEditComplex(i.ChannelID, &discordgo.ChannelEdit{
				PermissionOverwrites: []*discordgo.PermissionOverwrite{
					{ID: i.GuildID, Type: 0, Deny: 1024},
					{ID: staffer.ID, Type: 1, Allow: 3072},
					{ID: supervisoreID, Type: 0, Allow: 3072},
					{ID: creatoreID, Type: 1, Allow: 3072},
				},
			})

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: 4,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("✅ **Ticket reclamato da:** %s", staffer.Mention()),
				},
			})

			// Corretto il modo in cui modifichiamo il messaggio per rimuovere il pulsante
			nuovoContenuto := i.Message.Content + "\n\n📌 **STATO: RECLAMATO**"
			s.ChannelMessageEditComplex(&discordgo.MessageEdit{
				ID:         i.Message.ID,
				Channel:    i.ChannelID,
				Content:    &nuovoContenuto,
				Components: []discordgo.MessageComponent{},
			})
		}
	})

	s.Open()
	s.ApplicationCommandBulkOverwrite(s.State.User.ID, "1495170590947016996", commands)
	fmt.Println("Bot Ville & Aziende Online! 🚀")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop
}
