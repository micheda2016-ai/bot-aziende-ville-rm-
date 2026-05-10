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
	// Server per tenere vivo il bot su Render
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

	// DEFINIZIONE COMANDI
	commands := []*discordgo.ApplicationCommand{
		{Name: "setup-assistenza", Description: "Invia il pannello Ticket Ville/Aziende"},
		{Name: "regolamento", Description: "Visualizza il regolamento"},
		{Name: "contratto", Description: "Genera contratto casa", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionInteger, Name: "numero-casa", Description: "Numero civico", Required: true},
			{Type: discordgo.ApplicationCommandOptionString, Name: "tipo", Description: "Grandezza casa", Required: true, Choices: []*discordgo.ApplicationCommandOptionChoice{
				{Name: "Normale", Value: "Normale"}, {Name: "Large", Value: "Large"}, {Name: "Extra Large", Value: "Extra Large"},
			}},
		}},
		{Name: "villa-allarme", Description: "Attiva sirena e avvisa lo staff"},
		{Name: "biz-annuncio", Description: "Invia notifica globale per l'azienda", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "messaggio", Description: "Testo dell'annuncio", Required: true},
		}},
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// GESTIONE COMANDI SLASH
		if i.Type == discordgo.InteractionApplicationCommand {
			data := i.ApplicationCommandData()
			switch data.Name {

			case "setup-assistenza":
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: 4,
					Data: &discordgo.InteractionResponseData{
						Content: "🏢 **PANNELLO GESTIONE ASSET & VILLE**\nUsa il menu qui sotto per richiedere assistenza o acquistare una proprietà.",
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{Components: []discordgo.MessageComponent{
								discordgo.SelectMenu{
									CustomID: "ticket_asset",
									Placeholder: "Seleziona il motivo del ticket...",
									Options: []discordgo.SelectMenuOption{
										{Label: "Acquisto Villa / Azienda", Value: "acquisto_asset", Description: "Per comprare nuove proprietà"},
										{Label: "Problemi Tecnici / Gestione", Value: "gestione_asset", Description: "Per chiavi smarrite o bug"},
									},
								},
							}},
						},
					},
				})

			case "regolamento":
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: 4,
					Data: &discordgo.InteractionResponseData{Content: "📜 **REGOLAMENTO ASSET**\n1. Ogni proprietà va mantenuta attiva.\n2. Non abusare dei permessi chiavi.\n3. Rispetta lo staff."},
				})
			}
		}

		// GESTIONE CREAZIONE TICKET (QUANDO USANO IL MENU)
		if i.Type == discordgo.InteractionMessageComponent && i.MessageComponentData().CustomID == "ticket_asset" {
			scelta := i.MessageComponentData().Values[0]
			
			// Crea il canale testuale
			ch, err := s.GuildChannelCreateComplex(i.GuildID, discordgo.GuildChannelCreateData{
				Name: "ticket-" + i.Member.User.Username,
				Type: discordgo.ChannelTypeGuildText,
				PermissionOverwrites: []*discordgo.PermissionOverwrite{
					{ID: i.GuildID, Type: 0, Deny: 1024}, // Nasconde a @everyone
					{ID: i.Member.User.ID, Type: 1, Allow: 3072}, // Permette visualizzazione all'utente
				},
			})

			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: 4,
					Data: &discordgo.InteractionResponseData{Content: "❌ Errore nella creazione del ticket. Controlla i permessi del bot!", Flags: 64},
				})
				return
			}

			// RISPOSTA E PING DELLO STAFF NEL NUOVO CANALE
			// Sostituisci gli ID qui sotto con quelli reali del tuo staff se necessario
			ruoloStaff1 := "1495179869061906602"
			ruoloStaff2 := "1495180574627860621"

			messaggioBenvenuto := fmt.Sprintf("🎫 **NUOVO TICKET APERTO**\n\n**Utente:** %s\n**Motivo:** %s\n\nAttenzione staff: <@&%s> <@&%s>", 
				i.Member.User.Mention(), scelta, ruoloStaff1, ruoloStaff2)

			s.ChannelMessageSend(ch.ID, messaggioBenvenuto)

			// Conferma all'utente che il ticket è stato aperto
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: 4,
				Data: &discordgo.InteractionResponseData{
					Content: "✅ Il tuo ticket è stato aperto qui: <#" + ch.ID + ">",
					Flags: 64,
				},
			})
		}
	})

	s.Open()
	// Registrazione comandi immediata per il tuo server specifico
	s.ApplicationCommandBulkOverwrite(s.State.User.ID, "1495170590947016996", commands)
	
	fmt.Println("Bot Ville & Aziende Online! 🚀")
	
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop
}
