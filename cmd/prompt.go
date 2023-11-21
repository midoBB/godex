package cmd

import (
	"godex/internal/config"
	"godex/internal/tui"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Fill out the configuration through a series of prompts.",
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(tui.InitialModel())

		if m, err := p.Run(); err != nil {
			log.Fatal(err)
		} else {
			model := m.(tui.TuiModel)
			cfg := model.GetConfig()
			err := config.SaveConfig(cfg)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}
