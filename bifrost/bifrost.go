package bifrost

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/eirini/models/cf"
	"code.cloudfoundry.org/eirini/opi"
)

type Bifrost struct {
	Converter Converter
	Desirer   opi.Desirer
}

func (b *Bifrost) Transfer(ctx context.Context, request cf.DesireLRPRequest) error {
	desiredLRP, err := b.Converter.Convert(request)
	if err != nil {
		return errors.Wrap(err, "failed to convert request")
	}
	fmt.Printf("\n Desired: %#v\n", desiredLRP)
	return errors.Wrap(b.Desirer.Desire(&desiredLRP), "failed to desire")
}

func (b *Bifrost) List(ctx context.Context) ([]*models.DesiredLRPSchedulingInfo, error) {
	lrps, err := b.Desirer.List()
	if err != nil {
		return nil, errors.Wrap(err, "failed to list desired LRPs")
	}

	return toDesiredLRPSchedulingInfo(lrps), nil
}

func toDesiredLRPSchedulingInfo(lrps []*opi.LRP) []*models.DesiredLRPSchedulingInfo {
	infos := []*models.DesiredLRPSchedulingInfo{}
	for _, l := range lrps {
		info := &models.DesiredLRPSchedulingInfo{}
		info.DesiredLRPKey.ProcessGuid = l.Metadata[cf.ProcessGUID]
		info.Annotation = l.Metadata[cf.LastUpdated]
		infos = append(infos, info)
	}
	return infos
}

func (b *Bifrost) Update(ctx context.Context, update cf.UpdateDesiredLRPRequest) error {
	identifier := opi.LRPIdentifier{
		GUID:    update.GUID,
		Version: update.Version,
	}

	lrp, err := b.Desirer.Get(identifier)
	if err != nil {
		return errors.Wrap(err, "failed to get app")
	}

	lrp.TargetInstances = int(*update.Update.Instances)
	lrp.Metadata[cf.LastUpdated] = *update.Update.Annotation

	lrp.Metadata[cf.VcapAppUris] = getURIs(update)
	return errors.Wrap(b.Desirer.Update(lrp), "failed to update")
}

func (b *Bifrost) GetApp(ctx context.Context, identifier opi.LRPIdentifier) (*models.DesiredLRP, error) {
	lrp, err := b.Desirer.Get(identifier)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get app")
	}

	desiredLRP := &models.DesiredLRP{
		ProcessGuid: identifier.ProcessGUID(),
		Instances:   int32(lrp.TargetInstances),
	}

	return desiredLRP, nil
}

func (b *Bifrost) Stop(ctx context.Context, identifier opi.LRPIdentifier) error {
	return errors.Wrap(b.Desirer.Stop(identifier), "failed to stop app")
}

func (b *Bifrost) StopInstance(ctx context.Context, identifier opi.LRPIdentifier, index uint) error {
	if err := b.Desirer.StopInstance(identifier, index); err != nil {
		return errors.Wrap(err, "failed to stop instance")
	}
	return nil
}

func (b *Bifrost) GetInstances(ctx context.Context, identifier opi.LRPIdentifier) ([]*cf.Instance, error) {
	opiInstances, err := b.Desirer.GetInstances(identifier)
	if err != nil {
		return []*cf.Instance{}, errors.Wrap(err, "failed to get instances for app")
	}

	cfInstances := make([]*cf.Instance, 0, len(opiInstances))
	for _, i := range opiInstances {
		cfInstances = append(cfInstances, &cf.Instance{
			Since:          i.Since,
			Index:          i.Index,
			State:          i.State,
			PlacementError: i.PlacementError,
		})
	}

	return cfInstances, nil
}

func getURIs(update cf.UpdateDesiredLRPRequest) string {
	if !routesAvailable(update.Update.Routes) {
		return ""
	}

	cfRouterRoutes := (*update.Update.Routes)["cf-router"]
	data, err := cfRouterRoutes.MarshalJSON()
	if err != nil {
		panic("This should never happen")
	}

	return string(data)
}

func routesAvailable(routes *models.Routes) bool {
	if routes == nil {
		return false
	}

	if _, ok := (*routes)["cf-router"]; !ok {
		return false
	}

	return true
}
