package cmd

import (
	"context"
	"errors"
	"fmt"
	"k8s-event-listener/pkg/eventlistener"
	"k8s-event-listener/pkg/resource"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/pflag"

	"github.com/heptiolabs/healthcheck"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// K8sEventListenerCommand main application
type K8sEventListenerCommand struct {
	rootCommand   *cobra.Command
	eventListener *eventlistener.EventListener
	ctx           context.Context
	cErr          chan error
	healthHandler healthcheck.Handler
}

// NewK8sEventListenerCommand returns a pointer to K8sEventListenerCommand
func NewK8sEventListenerCommand(ctx context.Context) *K8sEventListenerCommand {
	return &K8sEventListenerCommand{
		rootCommand:   getRootCommand(),
		ctx:           ctx,
		cErr:          make(chan error),
		healthHandler: healthcheck.NewHandler(),
	}
}

// Run the main application
func (k *K8sEventListenerCommand) Run() int {
	k.rootCommand.Flags().StringP("probe-port", "p", "8080", "HTTP port to listen for liveness/readiness probes")
	for i, item := range resource.Resources {
		k.rootCommand.Flags().String(strings.Join(item.Name, ", "), "", "Callback for k8s resource event")
	}

	k.rootCommand.PersistentPreRunE = func(cmd *cobra.Command, args []string) (err error) {
		k.rootCommand.Flags().VisitAll(bindFlags)

		k.eventListener = eventlistener.NewEventListener(
			k.ctx,
			viper.GetString("kube_config"),
			viper.GetString("kube_context"),
			k.handleError,
			viper.GetString("verbose"),
			k.healthHandler,
		)

		return k.eventListener.Init()
	}

	k.rootCommand.RunE = func(cmd *cobra.Command, args []string) (err error) {
		go func() {
			log.Println(fmt.Sprintf("Server started, listening in :%s", viper.GetString("probe_port")))
			k.cErr <- http.ListenAndServe(fmt.Sprintf(":%s", viper.GetString("probe_port")), k.healthHandler)
		}()

		go func() {
			listening := false
			for nil, item := range resource.Resources {
				for nil, name := range item.Name {
					if viper.IsSet(name) {
						r, err := resource.NewResource(item.Name[0], viper.GetString(item.Name[0]))
						if err != nil {
							k.cErr <- err
							return
						}

						err = k.eventListener.Listen(r)
						if err != nil {
							k.cErr <- err
							return
						}
						
						listening := true
					}
				}
			}
			
			if listening != true {
				k.cErr <- errors.New("no known resource set")
				return
			}
			
			return
		}()

		err = <-k.cErr
		return
	}

	if err := k.rootCommand.Execute(); err != nil {
		k.handleError(err)
		return 1
	}

	return 0
}

func (k *K8sEventListenerCommand) populateConfig() (err error) {
	viper.AddConfigPath(".")

	viper.SetConfigName(".config")
	viper.SetEnvPrefix("K8S_EVENT_LISTENER")
	viper.AutomaticEnv()

	return viper.ReadInConfig()
}

func (k *K8sEventListenerCommand) handleError(err error) {
	log.Println(fmt.Sprintf("[ERROR] %s",
		err.Error(),
	))
}

func bindFlags(flag *pflag.Flag) {
	if err := viper.BindPFlag(strings.ReplaceAll(flag.Name, "-", "_"), flag); err != nil {
		panic(err)
	}
}

func getRootCommand() (c *cobra.Command) {
	c = &cobra.Command{
		Use:           "k8s-event-listener",
		Short:         "Listen for specific kubernetes events",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	c.PersistentFlags().String("kube-config", "", "Path to kubeconfig file")
	c.PersistentFlags().String("kube-context", "", "Context to use")
	c.PersistentFlags().StringP("verbose", "v", "0", "Verbose level")

	return
}
