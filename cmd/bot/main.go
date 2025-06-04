package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/lenin884/cryptobot/internal/bot"
	"github.com/lenin884/cryptobot/internal/config"
	"github.com/lenin884/cryptobot/internal/market"
	"github.com/lenin884/cryptobot/internal/storage"
	"github.com/pkg/errors"
)

// main is the entry point of the bot application. It initializes the context for graceful shutdown,
// loads the configuration, creates the bot instance, and starts handling updates from the Telegram API.
// It also sets up signal handling to gracefully shut down the bot when an interrupt signal is received.
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	configPath := os.Getenv("CONFIG_PATH")
	config, err := config.NewConfig(configPath)
	if err != nil {
		log.Fatal(errors.Wrap(err, "can't load config"))
	}

	store, err := storage.New(config.DBPath)
	if err != nil {
		log.Fatal(errors.Wrap(err, "can't open db"))
	}

	bybitClient := market.NewBybit(config.Bybit.Key, config.Bybit.Secret)

	bot, err := bot.NewBot(config.Telegram.Token, bybitClient, store)
	if err != nil {
		log.Fatal(errors.Wrap(err, "can't create bot"))
	}
	log.Println("Bot created")

	updates, err := bot.GetUpdatesChan()
	if err != nil {
		log.Fatal(errors.Wrap(err, "can't get updates channel"))
	}
	log.Println("Updates channel getter")

	go func() {
		bot.HandleUpdates(ctx, updates)
	}()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
		<-sig
		cancel()
	}()

	<-ctx.Done()

	log.Println("shutting down")
}
