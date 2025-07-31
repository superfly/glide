package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/superfly/fly-go/flaps"
	"github.com/superfly/flyctl/internal/appconfig"
	"github.com/superfly/flyctl/internal/cmdutil/preparers"
	"github.com/superfly/flyctl/internal/command"
	"github.com/superfly/flyctl/internal/command/deploy"
	"github.com/superfly/flyctl/internal/flag/flagctx"
	"github.com/superfly/flyctl/internal/flapsutil"
	"github.com/superfly/flyctl/internal/logger"
	"github.com/superfly/flyctl/internal/task"
	"github.com/superfly/flyctl/iostreams"
)

func deployApp(w http.ResponseWriter, r *http.Request) {
	var config appconfig.Config

	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		msg := fmt.Errorf("could not decode body into app config: %v", err)
		log.Println(msg)
		http.Error(w, msg.Error(), http.StatusBadRequest)
	}

	ctx := r.Context()

	var buf bytes.Buffer
	ctx = logger.NewContext(ctx, logger.New(&buf, logger.Info, true))

	// HACK: We need to have some flags for the deploy command to work. We just use a hack
	// here by using the flags set on the deploy command
	ctx = flagctx.NewContext(ctx, deploy.New().Flags())

	ctx = appconfig.WithName(ctx, config.AppName)

	ctx, err := command.DetermineHostname(ctx)
	if err != nil {
		msg := fmt.Errorf("could not determine hostname: %v", err)
		http.Error(w, msg.Error(), http.StatusInternalServerError)
	}
	ctx, err = command.DetermineWorkingDir(ctx)
	if err != nil {
		msg := fmt.Errorf("could not determine working directory: %v", err)
		http.Error(w, msg.Error(), http.StatusInternalServerError)

	}
	ctx, err = preparers.DetermineConfigDir(ctx)
	if err != nil {
		msg := fmt.Errorf("could not determine config directory: %v", err)
		http.Error(w, msg.Error(), http.StatusInternalServerError)
	}
	ctx, err = preparers.LoadConfig(ctx)
	if err != nil {
		msg := fmt.Errorf("could not load config: %v", err)
		http.Error(w, msg.Error(), http.StatusInternalServerError)
	}
	ctx, err = preparers.InitClient(ctx)
	if err != nil {
		msg := fmt.Errorf("could not create prepare client: %w", err)
		http.Error(w, msg.Error(), http.StatusInternalServerError)
	}

	ctx = task.NewWithContext(ctx)
	ctx = iostreams.NewContext(ctx, iostreams.System())

	// Instantiate FLAPS client if we haven't initialized one via a unit test.
	if flapsutil.ClientFromContext(ctx) == nil {
		flapsClient, err := flapsutil.NewClientWithOptions(ctx, flaps.NewClientOpts{
			AppName: config.AppName,
		})
		if err != nil {
			msg := fmt.Errorf("could not create flaps client: %w", err)
			http.Error(w, msg.Error(), http.StatusInternalServerError)
		}
		ctx = flapsutil.NewContextWithClient(ctx, flapsClient)
	}

	if err := deploy.DeployWithConfig(ctx, &config, 0, true); err != nil {
		msg := fmt.Errorf("could not deploy app: %v", err)
		log.Println(msg)
		http.Error(w, msg.Error(), http.StatusInternalServerError)
	}
}
