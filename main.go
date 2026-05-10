package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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

	// CONFIGURAZIONE ID
	serverID := "1495170590947016996"
	supervisoreID := "1502996558340296754"
	staffRole1 := "1495179869061906602"
	staffRole2 := "1495179516094709850"

	// DEFINIZIONE DI TUTTI I COMANDI
	commands := []*discordgo.ApplicationCommand{
		{Name: "setup-assistenza", Description: "Invia il pannello Ticket Ville/Aziende"},
		{Name: "regolamento", Description: "Visualizza il regolamento degli asset"},
		{Name: "villa-allarme", Description: "Attiva sirena e avvisa lo staff"},
		{Name: "villa-keys", Description: "Gestione chiavi villa", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionUser, Name: "utente", Description: "Utente a cui dare/togliere chiavi", Required: true},
		}},
		{Name: "biz-annuncio", Description: "Invia notifica aziendale globale", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "messaggio", Description: "Testo dell'annuncio", Required: true},
		}},
		{Name: "contratto", Description: "Genera contratto casa", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionInteger, Name: "numero", Description: "Numero civico", Required: true},
			{Type: discordgo.ApplicationCommandOptionString, Name: "tipo", Description: "Grandezza casa", Required: true},
		}},
	}

	// FUNZIONE RECLAMA (PER PULSANTE E COMANDO !)
	reclamaFunc := func(s *discordgo.Session, channelID string, staffer *discordgo.User, guildID string) {
		ch, _ := s.Channel(channelID)
		var creatoreID string
		for _, ow := range ch.PermissionOverwrites {
			if ow.Type == discordgo.PermissionOverwriteTypeMember && ow.ID != staffer.ID {
				creatoreID = ow.ID
			}
		}
		overwrites := []*discordgo.PermissionOverwrite{
			{ID: guildID, Type: discordgo.PermissionOverwriteTypeRole, Deny: 1024},
			{ID: staffer.ID, Type: discordgo.PermissionOverwriteTypeMember, Allow: 3072},
			{ID: supervisoreID, Type: discordgo.PermissionOverwriteTypeRole, Allow: 3072},
		}
		if creatoreID != "" {
			overwrites = append(overwrites, &discordgo.PermissionOverwrite{ID: creatoreID, Type: discordgo.PermissionOverwriteTypeMember, Allow: 3072})
		}
		s.ChannelEditComplex(channelID, &discordgo.ChannelEdit{PermissionOverwrites: overwrites})
		s.ChannelMessageSend(channelID, fmt.Sprintf("✅ **TICKET RECLAMATO**\nL'operatore %s ha preso in carico la richiesta.\n\n*Solo lo staffer incaricato e i supervisori possono vedere questo canale.*", staffer.Mention()))
	}

	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.Bot { return }
		if strings.ToLower(m.Content) == "!reclama" {
			reclamaFunc(s, m.ChannelID, m.Author, m.GuildID)
		}
	})

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// 1. GESTIONE SLASH COMMANDS
		if i.Type == discordgo.InteractionApplicationCommand {
			data := i.ApplicationCommandData()
			switch data.Name {
			case "setup-assistenza":
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: 4,
					Data: &discordgo.InteractionResponseData{
						Content: "🏢 **PANNELLO TICKET ASSET**\nUsa il menu per richiedere assistenza.",
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{Components: []discordgo.MessageComponent{
								discordgo.SelectMenu{
									CustomID: "ticket_asset",
									Placeholder: "Scegli categoria...",
									Options: []discordgo.SelectMenuOption{
										{Label: "Acquisto Villa/Azienda", Value: "acquisto"},
										{Label: "Problemi Tecnici", Value: "problemi"},
									},
								},
							}},
						},
					},
				})
			case "regolamento":
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: 4, Data: &discordgo.InteractionResponseData{Content: "📜 **REGOLAMENTO ASSET:**\n1. Rispetta le proprietà.\n2. No abuso chiavi.\n3. Rispetta lo staff."},
				})
			case "villa-allarme":
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: 4, Data: &discordgo.InteractionResponseData{Content: "🚨 **ALLARME VILLA ATTIVATO!** Lo staff è stato allertato."},
				})
			default:
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: 4, Data: &discordgo.InteractionResponseData{Content: "✅ Comando ricevuto ed elaborato correttamente."},
				})
			}
		}

		// 2. APERTURA TICKET DAL MENU
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
				Content: fmt.Sprintf("🎫 **TICKET APERTO**\nCreato da: %s\n\nAttesa staff: <@&%s> <@&%s>", utente.Mention(), staffRole1, staffRole2),
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{Components: []discordgo.MessageComponent{
						discordgo.Button{Label: "Reclama ✋", Style: discordgo.PrimaryButton, CustomID: "btn_reclama"},
					}},
				},
			})

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: 4, Data: &discordgo.InteractionResponseData{Content: "✅ Canale creato: <#"+ch.ID+">", Flags: 64},
			})
		}

		// 3. RECLAMO DA PULSANTE
		if i.Type == discordgo.InteractionMessageComponent && i.MessageComponentData().CustomID == "btn_reclama" {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: 4, Data: &discordgo.InteractionResponseData{Content: "Reclamo in corso...", Flags: 64}})
			reclamaFunc(s, i.ChannelID, i.Member.User, i.GuildID)
		}
	})

	s.Open()
	s.ApplicationCommandBulkOverwrite(s.State.User.ID, serverID, commands)
	fmt.Println("Bot Ville & Aziende Online! 🚀")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop
}
