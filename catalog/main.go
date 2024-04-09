package catalog

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// init
func init() {
	initEn(language.Make("en"))
	initZhhans(language.Make("zh-hans"))
	initZhhant(language.Make("zh-hant"))
}
// initEn will init en support.
func initEn(tag language.Tag) {
	message.SetString(tag, "%s gen_bot_events [module] [outfile]", "%s gen_bot_events [module] [outfile]")
	message.SetString(tag, "Echo framework's CLI scaffolding tool", "Echo framework's CLI scaffolding tool")
	message.SetString(tag, "Generate Bot Events", "Generate Bot Events")
	message.SetString(tag, "[module] can't be empty.", "[module] can't be empty.")
	message.SetString(tag, "[project name] can't be empty.", "[project name] can't be empty.")
	message.SetString(tag, "a tool for managing message translations.", "a tool for managing message translations.")
	message.SetString(tag, "create a project", "create a project")
	message.SetString(tag, "print only the version", "print only the version")
}
// initZhhans will init zh-hans support.
func initZhhans(tag language.Tag) {
	message.SetString(tag, "%s gen_bot_events [module] [outfile]", "%s gen_bot_events [module] [outfile]")
	message.SetString(tag, "Echo framework's CLI scaffolding tool", "Echo framework's CLI scaffolding tool")
	message.SetString(tag, "Generate Bot Events", "Generate Bot Events")
	message.SetString(tag, "[module] can't be empty.", "[module] can't be empty.")
	message.SetString(tag, "[project name] can't be empty.", "[project name] can't be empty.")
	message.SetString(tag, "a tool for managing message translations.", "a tool for managing message translations.")
	message.SetString(tag, "create a project", "create a project")
	message.SetString(tag, "print only the version", "print only the version")
}
// initZhhant will init zh-hant support.
func initZhhant(tag language.Tag) {
	message.SetString(tag, "%s gen_bot_events [module] [outfile]", "%s gen_bot_events [module] [outfile]")
	message.SetString(tag, "Echo framework's CLI scaffolding tool", "Echo framework's CLI scaffolding tool")
	message.SetString(tag, "Generate Bot Events", "Generate Bot Events")
	message.SetString(tag, "[module] can't be empty.", "[module] can't be empty.")
	message.SetString(tag, "[project name] can't be empty.", "[project name] can't be empty.")
	message.SetString(tag, "a tool for managing message translations.", "a tool for managing message translations.")
	message.SetString(tag, "create a project", "create a project")
	message.SetString(tag, "print only the version", "print only the version")
}
