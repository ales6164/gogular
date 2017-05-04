package gogular

import (
	"os"
	"encoding/json"
)

type Theme struct {
	Directory     string
	Configuration *ThemeConfiguration

	*Stylesheet
}

type ThemeConfiguration struct {
	Selector string
	Name     string

	StyleUrls []string
}

func (t *Theme) ReadConfiguration() {
	file, err := os.Open(t.Directory + "/config.json")
	if err != nil {
		ConsoleLog(err)
	}

	decoder := json.NewDecoder(file)

	t.Configuration = new(ThemeConfiguration)
	err = decoder.Decode(t.Configuration)
	if err != nil {
		ConsoleLog(err)
	}
}
