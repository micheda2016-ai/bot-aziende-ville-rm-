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
	// Keep-alive per Render
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

	// --- TUTTI I COMANDI ---
	commands := []*discordgo.ApplicationCommand{
		{Name: "setup-assistenza", Description: "Invia il pannello Ticket Ville/Aziende"},
		{Name: "regolamento", Description: "Visualizza il regolamento degli asset"},
		{Name: "villa-allarme", Description: "Attiva sirena e avvisa lo staff"},
		{Name: "villa-keys", Description: "Gestione chiavi", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionUser, Name: "utente", Description: "Utente target", Required: true},
		}},
		{Name: "biz-annuncio", Description: "Invia notifica aziendale", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "messaggio", Description: "Testo", Required: true},
		}},
		{Name: "contratto", Description: "Genera contratto casa", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionInteger, Name: "numero", Description: "Civico", Required: true},
			{Type: discordgo.ApplicationCommandOptionString, Name: "tipo", Description: "Grandezza", Required: true},
		}},
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		
		// 1. GESTIONE COMANDI SLASH
		if i.Type == discordgo.InteractionApplicationCommand {
			data := i.ApplicationCommandData()
			switch data.Name {
			case "setup-assistenza":
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: 4,
					Data: &discordgo.InteractionResponseData{
						Content: "🏢 **PANNELLO GESTIONE ASSET & VILLE**\nSeleziona una categoria per aprire un ticket.",
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{Components: []discordgo.MessageComponent{
								discordgo.SelectMenu{
									CustomID: "ticket_asset",
									Placeholder: "Scegli motivo...",
									Options: []discordgo.SelectMenuOption{
										{Label: "Acquisto Villa / Azienda", Value: "acquisto_asset"},
										{Label: "Gestione / Problemi", Value: "gestione_asset"},
									},
								},
							}},
						},
					},
				})
			case "regolamento":
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: 4, Data: &discordgo.InteractionResponseData{Content: "📜 **REGOLAMENTO:** Rispetta le proprietà e lo staff."},
				})
			}
		}

		// 2. APERTURA TICKET (MENU)
		if i.Type == discordgo.InteractionMessageComponent && i.MessageComponentData().CustomID == "ticket_asset" {
			utente := i.Member.User
			ch, err := s.GuildChannelCreateComplex(i.GuildID, discordgo.GuildChannelCreateData{
				Name: "ticket-" + utente.Username,
				Type: discordgo.ChannelTypeGuildText,
				PermissionOverwrites: []*discordgo.PermissionOverwrite{
					{ID: i.GuildID, Type: discordgo.PermissionOverwriteTypeRole, Deny: 1024},
					{ID: utente.ID, Type: discordgo.PermissionOverwriteTypeMember, Allow: 3072},
				},
			})
			if err != nil { return }

			// Messaggio nel ticket con PING STAFF
			msg := fmt.Sprintf("🎫 **TICKET CREATO**\n**Creato da:** %s\n\nStaff: <@&1495179869061906602> <@&1495179516094709850>", utente.Mention())
			s.ChannelMessageSendComplex(ch.ID, &discordgo.MessageSend{
				Content: msg,
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{Components: []discordgo.MessageComponent{
						discordgo.Button{Label: "Reclama ✋", Style: discordgo.PrimaryButton, CustomID: "reclama_" + utente.ID},
					}},
				},
			})

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: 4, Data: &discordgo.InteractionResponseData{Content: "✅ Ticket aperto: <#"+ch.ID+">", Flags: 64},
			})
		}

		// 3. LOGICA RECLAMA
		if i.Type == discordgo.InteractionMessageComponent && len(i.MessageComponentData().CustomID) > 8 && i.MessageComponentData().CustomID[:8] == "reclama_" {
			staffer := i.Member.User
			creatoreID := i.MessageComponentData().CustomID[8:]
			supervisoreID := "1502996558340296754"

			// Cambia permessi canale
			s.ChannelEditComplex(i.ChannelID, &discordgo.ChannelEdit{
				PermissionOverwrites: []*discordgo.PermissionOverwrite{
					{ID: i.GuildID, Type: discordgo.PermissionOverwriteTypeRole, Deny: 1024},
					{ID: staffer.ID, Type: discordgo.PermissionOverwriteTypeMember, Allow: 3072},
					{ID: supervisoreID, Type: discordgo.PermissionOverwriteTypeRole, Allow: 3072},
					{ID: creatoreID, Type: discordgo.PermissionOverwriteTypeMember, Allow: 3072},
				},
			})

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: 4,
				Data: &discordgo.InteractionResponseData{Content: fmt.Sprintf("✅ **Ticket reclamato da:** %s", staffer.Mention())},
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
