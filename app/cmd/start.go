package cmd

import (
	"net/http"
	"os"
	"os/signal"

	"github.com/glendsoza/sprobe/prober"
	"github.com/glendsoza/sprobe/spec"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		config := cmd.Flag("config")
		fileData, err := os.ReadFile(config.Value.String())
		if err != nil {
			log.Fatal().
				Str("file_name", config.Value.String()).
				Err(err).
				Msg("unable to read the file")
		}
		var specs []*spec.LivenessProbe
		err = yaml.Unmarshal(fileData, &specs)
		if err != nil {
			log.Fatal().
				Str("file_name", config.Value.String()).
				Err(err).
				Msg("error loading the file")
		}
		sp, err := prober.NewProberManager(prober.NewServiceProber())
		if err != nil {
			log.Fatal().
				Str("file_name", config.Value.String()).
				Err(err).
				Msg("unable to create prober manager")
		}
		for _, spec := range specs {
			err := sp.Add(spec)
			if err != nil {
				log.Fatal().
					Str("file_name", config.Value.String()).
					Err(err).
					Msg("unable to load the spec")
			} else {
				log.Info().Str("file_name", config.Value.String()).
					Str("service_name", spec.ServiceName).
					Msg("loaded service for monitoring")
			}
		}
		log.Info().Str("file_name", config.Value.String()).Msg("monitoring")
		go func() {
			http.Handle("/metrics", promhttp.Handler())
			http.ListenAndServe(":2112", nil)
		}()
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		log.Info().Str("file_name", config.Value.String()).Msg("stopping monitoring, received sig int")
	},
}
