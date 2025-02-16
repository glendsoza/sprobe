package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "sprobe",
		Short: "",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	cfgFile string
)

func Execute() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "./sprobe.yaml", "config file")
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err)
	}
}
