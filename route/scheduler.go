package route

import (
	"fmt"

	"code.cloudfoundry.org/eirini/util"
	"github.com/pkg/errors"
)

//go:generate counterfeiter . Emitter
type Emitter interface {
	Emit(Message)
}

type CollectorScheduler struct {
	Collector Collector
	Scheduler util.TaskScheduler
	Emitter   Emitter
}

func (c CollectorScheduler) Start() {
	c.Scheduler.Schedule(func() error {
		routes, err := c.Collector.Collect()
		if err != nil {
			fmt.Println("SCHEDULER START COLLECT ERRORED", err)
			return errors.Wrap(err, "failed to collect routes")
		}
		for _, r := range routes {
			fmt.Printf("ABOUT TO EMIT: %#v", r)
			c.Emitter.Emit(r)
		}
		return nil
	})
}
