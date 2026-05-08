package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

func main() {
	// Server per Render (Keep-Alive)
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, "Bot ER:HC & FDO Online") })
		port := os.Getenv("PORT")
		if port == "" { port = "8080" }
		http.ListenAndServe(":"+port, nil)
	}()

	token := os.Getenv("DISCORD_TOKEN")
	s, err := discordgo.New("Bot " + token)
	if err != nil { log.Fatalf("Errore sessione: %v", err) }

	// --- DEFINIZIONE DI TUTTI I COMANDI (FDO + VILLE + AZIENDE) ---
	commands := []*discordgo.ApplicationCommand{
		// COMANDI FDO ORIGINALI
		{Name: "chiama-fdo", Description: "Invia notifica alla Categoria FDO"},
		{
			Name: "arresto",
			Description: "Registra un arresto",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionUser, Name: "discord-civile", Description: "Tag Discord civile", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "roblox-civile", Description: "Nome Roblox civile", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "roblox-agente", Description: "Tuo nome Roblox", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "motivo", Description: "Motivo", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "verbale", Description: "Verbale", Required: true},
			},
		},
		{
			Name: "multa",
			Description: "Registra una multa",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionUser, Name: "discord-civile", Description: "Tag Discord civile", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "roblox-civile", Description: "Nome Roblox civile", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "roblox-agente", Description: "Tuo nome Roblox", Required: true},
				{Type: discordgo.ApplicationCommandOptionInteger, Name: "valore", Description: "Importo (1000-8000)", Required: true, MinValue: &[]float64{1000}[0], MaxValue: 8000},
				{Type: discordgo.ApplicationCommandOptionString, Name: "motivo", Description: "Motivo", Required: true},
			},
		},
		// NUOVI COMANDI VILLE E AZIENDE
		{Name: "villa-gestione", Description: "Gestione permessi villa"},
		{Name: "villa-allarme", Description: "Attiva allarme villa"},
		{Name: "contratto", Description: "Genera contratto casa", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionInteger, Name: "numero-casa", Description: "Civico", Required: true},
			{Type: discordgo.ApplicationCommandOptionString, Name: "tipo", Description: "Grandezza", Required: true, Choices: []*discordgo.ApplicationCommandOptionChoice{
				{Name: "Normale", Value: "Normale"}, {Name: "Large", Value: "Large"}, {Name: "Extra Large", Value: "Extra Large"},
			}},
		}},
		{Name: "setup-assistenza", Description: "Pannello Ticket Ville/Aziende"},
		{Name: "regolamento", Description: "Mostra regolamento"},
		{Name: "ban", Description: "Banna utente", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionUser, Name: "utente", Description: "Da bannare", Required: true},
			{Type: discordgo.ApplicationCommandOptionString, Name: "motivo", Description: "Motivo"},
		}},
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			data := i.ApplicationCommandData()
			switch data.Name {

			// LOGICA FDO
			case "chiama-fdo":
				ruoloFDO := "1492918778885963836"
				mappa := "ijisma95"
				msg := fmt.Sprintf("<@&%s>\n🚨 **CHIAMATA FORZE DELL'ORDINE** 🚨\n\n👤 **Mittente:** <@%s>\n📍 **Cod Mappa EH:** `%s`\n⚠️ Intervento richiesto!", ruoloFDO, i.Member.User.ID, mappa)
				s.ChannelMessageSend(i.ChannelID, msg)
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: 4, Data: &discordgo.InteractionResponseData{Content: "✅ Inviata", Flags: 64}})

			case "arresto":
				res := fmt.Sprintf("⚖️ **ARRESTO**\nCivile: %s (%s)\nAgente: %s\nMotivo: %s\nVerbale: %s", data.Options[0].UserValue(s).Mention(), data.Options[1].StringValue(), data.Options[2].StringValue(), data.Options[3].StringValue(), data.Options[4].StringValue())
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: 4, Data: &discordgo.InteractionResponseData{Content: res}})

			// LOGICA VILLE/AZIENDE
			case "contratto":
				res := fmt.Sprintf("📄 **CONTRATTO**\nCasa N: %d\nTipo: %s\nFirmato: %s", data.Options[0].IntValue(), data.Options[1].StringValue(), i.Member.User.Mention())
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: 4, Data: &discordgo.InteractionResponseData{Content: res}})

			case "regolamento":
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: 4, Data: &discordgo.InteractionResponseData{Content: "📜 **Regolamento:** Cosa c'è scritto qui..."}})

			case "setup-assistenza":
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: 4,
					Data: &discordgo.InteractionResponseData{
						Content: "🏢 **PANNELLO ASSET**",
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{Components: []discordgo.MessageComponent{
								discordgo.SelectMenu{
									CustomID: "ticket_asset",
									Options: []discordgo.SelectMenuOption{
										{Label: "Acquisto Villa/Azienda", Value: "acquisto"},
										{Label: "Gestione Ville/Aziende", Value: "gestione"},
									},
								},
							}},
						},
					},
				})
			}
		}

		// LOGICA TICKET ASSET (PING RUOLI 1495179869061906602 e 1495180574627860621)
		if i.Type == discordgo.InteractionMessageComponent && i.MessageComponentData().CustomID == "ticket_asset" {
			cat := i.MessageComponentData().Values[0]
			ch, _ := s.GuildChannelCreateComplex(i.GuildID, discordgo.GuildChannelCreateData{
				Name: "asset-" + cat + "-" + i.Member.User.Username,
				PermissionOverwrites: []*discordgo.PermissionOverwrite{
					{ID: i.GuildID, Type: 0, Deny: 1024},
					{ID: i.Member.User.ID, Type: 1, Allow: 3072},
				},
			})
			s.ChannelMessageSend(ch.ID, fmt.Sprintf("🎫 **Ticket %s**\nUtente: %s\nPing: <@&1495179869061906602> <@&1495180574627860621>", cat, i.Member.User.Mention()))
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: 4, Data: &discordgo.InteractionResponseData{Content: "✅ Aperto: <#"+ch.ID+">", Flags: 64}})
		}
	})

	s.Open()
	s.ApplicationCommandBulkOverwrite(s.State.User.ID, "", commands)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}
