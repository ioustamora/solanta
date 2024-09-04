//go get github.com/mattn/go-sqlite3
package main

import (
    "context"
    "database/sql"
    "fmt"
    "os"
    "os/signal"
    "github.com/go-telegram/bot"
    "github.com/go-telegram/bot/models"
    "github.com/blocto/solana-go-sdk/client"
    "github.com/blocto/solana-go-sdk/types"
    _ "github.com/mattn/go-sqlite3"
)

func main() {
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
    defer cancel()

    // Initialize SQLite database
    db, err := sql.Open("sqlite3", "./wallets.db")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    // Create table if it doesn't exist
    createTableSQL := `CREATE TABLE IF NOT EXISTS wallets (
        "id" INTEGER PRIMARY KEY AUTOINCREMENT,
        "user_id" TEXT,
        "wallet_address" TEXT,
        "public_key" TEXT,
        "private_key" TEXT,
        "balance" REAL
    );`
    _, err = db.Exec(createTableSQL)
    if err != nil {
        panic(err)
    }

    opts := []bot.Option{
        bot.WithDefaultHandler(walletHandler(db)),
    }

    b, err := bot.New("YOUR_BOT_TOKEN_FROM_BOTFATHER", opts...)
    if err != nil {
        panic(err)
    }

    b.Start(ctx)
}

func walletHandler(db *sql.DB) bot.HandlerFunc {
    return func(ctx context.Context, b *bot.Bot, update *models.Update) {
        if update.Message != nil {
            chatID := update.Message.Chat.ID
            userID := update.Message.From.ID

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

            // Store wallet details in the database
            insertWalletSQL := `INSERT INTO wallets (user_id, wallet_address, public_key, private_key, balance) VALUES (?, ?, ?, ?, ?)`
            _, err = db.Exec(insertWalletSQL, userID, wallet.PublicKey.ToBase58(), wallet.PublicKey.ToBase58(), wallet.PrivateKey, float64(balance)/1e9)
            if err != nil {
                b.SendMessage(ctx, &bot.SendMessageParams{
                    ChatID: chatID,
                    Text:   "Failed to store wallet details: " + err.Error(),
                })
                return
            }

            // Send wallet details to the user
            message := fmt.Sprintf(
                "Wallet Address: %s\nPublic Key: %s\nPrivate Key: %x\nBalance: %f SOL",
                wallet.PublicKey.ToBase58(),
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
}
