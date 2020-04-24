package bifrost

import (
	"context"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/eirini/models/cf"
	"code.cloudfoundry.org/eirini/opi"
	"code.cloudfoundry.org/lager"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

//counterfeiter:generate . StagingConverter
type StagingConverter interface {
	ConvertStaging(stagingGUID string, request cf.StagingRequest) (opi.StagingTask, error)
}

//counterfeiter:generate . StagingDesirer
type StagingDesirer interface {
	DesireStaging(task *opi.StagingTask) error
	Delete(name string) error
}

//counterfeiter:generate . StagingCompleter
type StagingCompleter interface {
	CompleteStaging(*models.TaskCallbackResponse) error
}

type BuildpackStaging struct {
	Converter        StagingConverter
	StagingDesirer   StagingDesirer
	StagingCompleter StagingCompleter
	Logger           lager.Logger
}

func (b *BuildpackStaging) TransferStaging(ctx context.Context, stagingGUID string, stagingRequest cf.StagingRequest) error {
	desiredStaging, err := b.Converter.ConvertStaging(stagingGUID, stagingRequest)
	if err != nil {
		return errors.Wrap(err, "failed to convert staging task")
	}

	return errors.Wrap(b.StagingDesirer.DesireStaging(&desiredStaging), "failed to desire")
}

func (b *BuildpackStaging) CompleteStaging(task *models.TaskCallbackResponse) error {
	l := b.Logger.Session("complete-staging", lager.Data{"task-guid": task.TaskGuid})
	l.Debug("Complete staging")
	return multierr.Combine(
		b.StagingCompleter.CompleteStaging(task),
		b.StagingDesirer.Delete(task.TaskGuid),
	)
}