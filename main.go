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
	// Server per Render (Keep-Alive)
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, "Bot Ville & Aziende Online") })
		port := os.Getenv("PORT")
		if port == "" { port = "8080" }
		http.ListenAndServe(":"+port, nil)
	}()

	token := os.Getenv("DISCORD_TOKEN")
	s, err := discordgo.New("Bot " + token)
	if err != nil { log.Fatalf("Errore sessione: %v", err) }

	// --- DEFINIZIONE COMANDI SLASH VILLE & AZIENDE ---
	commands := []*discordgo.ApplicationCommand{
		{Name: "villa-gestione", Description: "Gestione permessi e stato della villa"},
		{Name: "villa-keys", Description: "Consegna o ritira chiavi digitali", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionUser, Name: "utente", Description: "Utente a cui dare/togliere chiavi", Required: true},
		}},
		{Name: "villa-allarme", Description: "Attiva sirena e avvisa lo staff"},
		{Name: "villa-ospiti", Description: "Mostra lista persone autorizzate nella villa"},
		{Name: "villa-sell", Description: "Avvia procedura vendita proprietà"},
		{Name: "biz-info", Description: "Visualizza dipendenti e capitale sociale"},
		{Name: "biz-assumi", Description: "Registra ufficialmente un nuovo lavoratore", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionUser, Name: "utente", Description: "Nuovo dipendente", Required: true},
		}},
		{Name: "biz-paga", Description: "Invia stipendio o bonus produzione"},
		{Name: "biz-turno", Description: "Segna l'entrata/uscita dal lavoro"},
		{Name: "biz-annuncio", Description: "Invia notifica globale per l'azienda", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "messaggio", Description: "Testo dell'annuncio", Required: true},
		}},
		{Name: "contratto", Description: "Genera contratto casa", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionInteger, Name: "numero-casa", Description: "Numero civico", Required: true},
			{Type: discordgo.ApplicationCommandOptionString, Name: "tipo", Description: "Grandezza casa", Required: true, Choices: []*discordgo.ApplicationCommandOptionChoice{
				{Name: "Normale", Value: "Normale"}, {Name: "Large", Value: "Large"}, {Name: "Extra Large", Value: "Extra Large"},
			}},
		}},
		{Name: "regolamento", Description: "Visualizza cosa c'è scritto nel regolamento"},
		{Name: "setup-assistenza", Description: "Invia il pannello Ticket Ville/Aziende"},
		{Name: "ban", Description: "Banna un utente", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionUser, Name: "utente", Description: "Utente da bannare", Required: true},
			{Type: discordgo.ApplicationCommandOptionString, Name: "motivo", Description: "Motivo del ban"},
		}},
		{Name: "kick", Description: "Espelli un utente", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionUser, Name: "utente", Description: "Utente da espellere", Required: true},
		}},
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			data := i.ApplicationCommandData()
			switch data.Name {

			case "regolamento":
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: 4,
					Data: &discordgo.InteractionResponseData{Content: "📜 **REGOLAMENTO SERVER**\nIl regolamento prevede rispetto e gioco corretto."},
				})

			case "contratto":
				num := data.Options[0].IntValue()
				tipo := data.Options[1].StringValue()
				res := fmt.Sprintf("📄 **CONTRATTO DI PROPRIETÀ**\n\n🏠 **Numero Casa:** %d\n📐 **Tipo:** %s\n👤 **Firmato da:** %s", num, tipo, i.Member.User.Mention())
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: 4, Data: &discordgo.InteractionResponseData{Content: res}})

			case "setup-assistenza":
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: 4,
					Data: &discordgo.InteractionResponseData{
						Content: "🏢 **PANNELLO GESTIONE ASSET**\nSeleziona una categoria per aprire un ticket.",
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{Components: []discordgo.MessageComponent{
								discordgo.SelectMenu{
									CustomID: "ticket_asset",
									Placeholder: "Scegli categoria...",
									Options: []discordgo.SelectMenuOption{
										{Label: "Acquisto Villa / Azienda", Value: "acquisto_asset"},
										{Label: "Richiesta Gestione", Value: "gestione_asset"},
									},
								},
							}},
						},
					},
				})

			case "ban":
				target := data.Options[0].UserValue(s)
				s.GuildBanCreate(i.GuildID, target.ID, 0)
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: 4, Data: &discordgo.InteractionResponseData{Content: "🔨 Utente " + target.Username + " bannato."}})
			}
		}

		if i.Type == discordgo.InteractionMessageComponent && i.MessageComponentData().CustomID == "ticket_asset" {
			cat := i.MessageComponentData().Values[0]
			ch, _ := s.GuildChannelCreateComplex(i.GuildID, discordgo.GuildChannelCreateData{
				Name: "asset-" + i.Member.User.Username,
				PermissionOverwrites: []*discordgo.PermissionOverwrite{
					{ID: i.GuildID, Type: 0, Deny: 1024},
					{ID: i.Member.User.ID, Type: 1, Allow: 3072},
				},
			})
			pingMsg := fmt.Sprintf("🎫 **NUOVO TICKET ASSET**\nCategoria: %s\nUtente: %s\n\nAttenzione: <@&1495179869061906602> <@&1495180574627860621>", cat, i.Member.User.Mention())
			s.ChannelMessageSend(ch.ID, pingMsg)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: 4, Data: &discordgo.InteractionResponseData{Content: "✅ Ticket creato: <#"+ch.ID+">", Flags: 64}})
		}
	})

	s.Open()
	// QUESTA RIGA SOTTO È QUELLA CHE CARICA I COMANDI ISTANTANEAMENTE NEL TUO SERVER
	s.ApplicationCommandBulkOverwrite(s.State.User.ID, "1495170590947016996", commands)
	
	fmt.Println("Bot Ville & Aziende Online! 🚀")
	
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop
}
