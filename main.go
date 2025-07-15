package main

//go:generate easyi18n extract . ./locales/en.json
//go:generate easyi18n update -f ./locales/en.json ./locales/zh-hans.json
//go:generate easyi18n update -f ./locales/en.json ./locales/zh-hant.json
//go:generate easyi18n generate --pkg=catalog ./locales ./catalog/main.go

import (
	"log"
	"os"

	"github.com/Xuanwo/go-locale"
	"github.com/mylukin/EchoPilot/command"
	ei18n "github.com/mylukin/easy-i18n/i18n"
	"github.com/urfave/cli/v2"

	_ "github.com/mylukin/EchoPilot/catalog"
)

func main() {
	// Detect OS language
	tag, _ := locale.Detect()

	// Set Language
	ei18n.SetLang(tag)

	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   ei18n.Sprintf("print only the version"),
	}

	app := &cli.App{
		Name:    `EchoPilot`,
		Version: "v0.1.67",
		Usage:   ei18n.Sprintf(`Echo framework's CLI scaffolding tool`),
		Action: func(c *cli.Context) error {
			cli.ShowAppHelp(c)
			return nil
		},
	}

	// 注册命令
	command.RegisterCommands(app)

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
