package main

import (
	"errors"
	"log"
	"os"

	"github.com/Xuanwo/go-locale"
	ei18n "github.com/mylukin/easy-i18n/i18n"
	"github.com/urfave/cli/v2"
)

func main() {
	// Detect OS language
	tag, _ := locale.Detect()

	// Set Language
	ei18n.SetLang(tag)

	appName := "codetool"

	app := &cli.App{
		HelpName: appName,
		Name:     appName,
		Usage:    ei18n.Sprintf(`a tool for managing message translations.`),
		Action: func(c *cli.Context) error {
			cli.ShowAppHelp(c)
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:      "gen_bot_events",
				Usage:     ei18n.Sprintf(`Generate Bot Events`),
				UsageText: ei18n.Sprintf(`%s gen_bot_events [module] [outfile]`, appName),
				Action: func(c *cli.Context) error {
					module := c.Args().Get(0)
					if module == "" {
						return errors.New(ei18n.Sprintf(`[module] can't be empty.`))
					}
					outFile := c.Args().Get(1)
					if outFile == "" {
						outFile = "./routers/bot_events.go"
					}

					return GenBotEvents(module, outFile)
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
