package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fox-one/mixin-sdk-go/v2"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
)

func main() {
	home, err := homedir.Dir()
	if err != nil {
		log.Println(err)
		return
	}

	app := &cli.App{
		Name:    "hook",
		Usage:   "messenger hook daemon",
		Version: "0.0.1",
		Action:  loop,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "dir",
				Value: fmt.Sprintf("%s/.mnm", home),
				Usage: "the webhook database",
			},
		},
		EnableBashCompletion: true,
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Println(err)
	}
}

func loop(c *cli.Context) error {
	dir := c.String("dir")
	conf, err := LoadConfig(dir + "/config.toml")
	if err != nil {
		return err
	}

	db, err := openDB(dir + "/db")
	if err != nil {
		return err
	}
	defer db.Close()

	ctx := context.Background()
	client, err := mixin.NewFromKeystore(&mixin.Keystore{
		AppID:             conf.Mixin.AppID,
		SessionID:         conf.Mixin.SessionID,
		SessionPrivateKey: conf.Mixin.SessionPrivateKey,
	})
	if err != nil {
		return err
	}

	hdr := &Handler{db: db, mixin: client, secret: conf.Mixin.OauthSecret}
	go func() {
		for {
			err := client.LoopBlaze(ctx, hdr)
			log.Println("LoopBlaze done with", err)
			time.Sleep(3 * time.Second)
		}
	}()
	return NewServer(hdr, conf.App.Port).ListenAndServe()
}
