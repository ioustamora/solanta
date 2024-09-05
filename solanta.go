package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/types"
)

// Send any text message to the bot after the bot has been started

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b, err := bot.New("", opts...)
	if err != nil {
		panic(err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/hello", bot.MatchTypeExact, helloHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/wallet", bot.MatchTypeExact, walletHandler)

	b.Start(ctx)
}

func helloHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      "Hello, *" + bot.EscapeMarkdown(update.Message.From.FirstName) + "*",
		ParseMode: models.ParseModeMarkdown,
	})
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   update.Message.Text,
	})
	// Print user information
	log.Printf("Chat ID: %d", update.Message.Chat.ID)
	log.Printf("User ID: %d", update.Message.From.ID)
	log.Printf("First Name: %s", update.Message.From.FirstName)
	log.Printf("Last Name: %s", update.Message.From.LastName)
	log.Printf("Username: @%s", update.Message.From.Username)
	log.Printf("Language Code: %s", update.Message.From.LanguageCode)
	log.Printf("Is Bot: %t", update.Message.From.IsBot)
	log.Printf("Is Premium: %t", update.Message.From.IsPremium)
	log.Printf("AddedToAttachmentMenu: %t", update.Message.From.AddedToAttachmentMenu)
	log.Printf("CanJoinGroups: %t", update.Message.From.CanJoinGroups)
	log.Printf("CanReadAllGroupMessages: %t", update.Message.From.CanReadAllGroupMessages)
	log.Printf("SupportInlineQueries: %t", update.Message.From.SupportInlineQueries)
	log.Printf("CanConnectToBusiness: %t", update.Message.From.CanConnectToBusiness)
}

func walletHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message != nil {
		chatID := update.Message.Chat.ID

		// Create a new Solana wallet
		wallet := types.NewAccount()

		// Airdrop 1 SOL to the new wallet
		c := client.NewClient("https://api.devnet.solana.com")
		_, err := c.RequestAirdrop(ctx, wallet.PublicKey.ToBase58(), 1e9) // 1 SOL = 1e9 lamports
		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Failed to airdrop 1 SOL: " + err.Error(),
			})
			return
		}

		// Get wallet balance
		balance, err := c.GetBalance(ctx, wallet.PublicKey.ToBase58())
		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Failed to get wallet balance: " + err.Error(),
			})
			return
		}

		// Send wallet details to the user
		message := fmt.Sprintf(
			"Wallet Address(Public Key): %s\nPrivate Key: %x\nBalance: %f SOL",
			wallet.PublicKey.ToBase58(),
			wallet.PrivateKey,
			float64(balance)/1e9,
		)

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   message,
		})
	}
}
