package stager

import (
	"errors"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/eirini/models/cf"
	"code.cloudfoundry.org/eirini/opi"
	"code.cloudfoundry.org/lager"
	"go.uber.org/multierr"
)

//counterfeiter:generate . StagingCompleter
type StagingCompleter interface {
	CompleteStaging(*models.TaskCallbackResponse) error
}

type Stager struct {
	Desirer          opi.TaskDesirer
	StagingCompleter StagingCompleter
	Logger           lager.Logger
}

func New(desirer opi.TaskDesirer, stagingCompleter StagingCompleter, logger lager.Logger) *Stager {
	return &Stager{
		Desirer:          desirer,
		StagingCompleter: stagingCompleter,
		Logger:           logger,
	}
}

func (s *Stager) Stage(string, cf.StagingRequest) error {
	return errors.New("Buildpack staging is now implemented via bifrost, do not call this method")
}

func (s *Stager) CompleteStaging(task *models.TaskCallbackResponse) error {
	l := s.Logger.Session("complete-staging", lager.Data{"task-guid": task.TaskGuid})
	l.Debug("Complete staging")
	return multierr.Combine(
		s.StagingCompleter.CompleteStaging(task),
		s.Desirer.Delete(task.TaskGuid),
	)
}
