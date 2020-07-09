package bifrost

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	"code.cloudfoundry.org/eirini"
	"code.cloudfoundry.org/eirini/models/cf"
	"code.cloudfoundry.org/eirini/opi"
)

//counterfeiter:generate . LRPConverter
type LRPConverter interface {
	ConvertLRP(request cf.DesireLRPRequest) (opi.LRP, error)
}

//counterfeiter:generate . LRPDesirer
type LRPDesirer interface {
	Desire(namespace string, lrp *opi.LRP) error
	List() ([]*opi.LRP, error)
	Get(identifier opi.LRPIdentifier) (*opi.LRP, error)
	GetInstances(identifier opi.LRPIdentifier) ([]*opi.Instance, error)
	Update(lrp *opi.LRP) error
	Stop(identifier opi.LRPIdentifier) error
	StopInstance(identifier opi.LRPIdentifier, index uint) error
}

type LRP struct {
	DefaultNamespace string
	Converter        LRPConverter
	Desirer          LRPDesirer
}

func (l *LRP) Transfer(ctx context.Context, request cf.DesireLRPRequest) error {
	desiredLRP, err := l.Converter.ConvertLRP(request)
	if err != nil {
		return errors.Wrap(err, "failed to convert request")
	}
	namespace := l.DefaultNamespace
	if request.Namespace != "" {
		namespace = request.Namespace
	}
	return errors.Wrap(l.Desirer.Desire(namespace, &desiredLRP), "failed to desire")
}

func (l *LRP) List(ctx context.Context) ([]cf.DesiredLRPSchedulingInfo, error) {
	lrps, err := l.Desirer.List()
	if err != nil {
		return nil, errors.Wrap(err, "failed to list desired LRPs")
	}

	return toDesiredLRPSchedulingInfo(lrps), nil
}

func toDesiredLRPSchedulingInfo(lrps []*opi.LRP) []cf.DesiredLRPSchedulingInfo {
	infos := []cf.DesiredLRPSchedulingInfo{}
	for _, l := range lrps {
		info := cf.DesiredLRPSchedulingInfo{}
		info.DesiredLRPKey.ProcessGUID = l.LRPIdentifier.ProcessGUID()
		info.GUID = l.LRPIdentifier.GUID
		info.Version = l.LRPIdentifier.Version
		info.Annotation = l.LastUpdated
		infos = append(infos, info)
	}
	return infos
}

func (l *LRP) Update(ctx context.Context, request cf.UpdateDesiredLRPRequest) error {
	identifier := opi.LRPIdentifier{
		GUID:    request.GUID,
		Version: request.Version,
	}

	lrp, err := l.Desirer.Get(identifier)
	if err != nil {
		return errors.Wrap(err, "failed to get app")
	}

	lrp.TargetInstances = request.Update.Instances
	lrp.LastUpdated = request.Update.Annotation

	lrp.AppURIs = getURIs(request.Update)
	return errors.Wrap(l.Desirer.Update(lrp), "failed to update")
}

func (l *LRP) GetApp(ctx context.Context, identifier opi.LRPIdentifier) (cf.DesiredLRP, error) {
	lrp, err := l.Desirer.Get(identifier)
	if err != nil {
		return cf.DesiredLRP{}, errors.Wrap(err, "failed to get app")
	}

	desiredLRP := cf.DesiredLRP{
		ProcessGUID: identifier.ProcessGUID(),
		Instances:   int32(lrp.TargetInstances),
		Annotation:  lrp.LastUpdated,
	}

	if lrp.AppURIs != "" {
		routes := json.RawMessage{}
		if err := routes.UnmarshalJSON([]byte(lrp.AppURIs)); err != nil {
			return cf.DesiredLRP{}, errors.Wrap(err, "failed to unmarshal routes")
		}
		lrpRoutes := map[string]json.RawMessage{"cf-router": routes}
		desiredLRP.Routes = lrpRoutes
	}

	return desiredLRP, nil
}

func (l *LRP) Stop(ctx context.Context, identifier opi.LRPIdentifier) error {
	return errors.Wrap(l.Desirer.Stop(identifier), "failed to stop app")
}

func (l *LRP) StopInstance(ctx context.Context, identifier opi.LRPIdentifier, index uint) error {
	if err := l.Desirer.StopInstance(identifier, index); err != nil {
		return errors.Wrap(err, "failed to stop instance")
	}
	return nil
}

func (l *LRP) GetInstances(ctx context.Context, identifier opi.LRPIdentifier) ([]*cf.Instance, error) {
	opiInstances, err := l.Desirer.GetInstances(identifier)
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

func (l *LRP) Reconcile(ctx context.Context, request cf.DesireLRPRequest, lastUpdated string) error {
	_, err := l.GetApp(ctx, opi.LRPIdentifier{
		GUID:    request.GUID,
		Version: request.Version,
	})
	if errors.Is(err, eirini.ErrNotFound) {
		return l.Transfer(ctx, request)
	}
	if err != nil {
		return err
	}

	updateRequest := cf.UpdateDesiredLRPRequest{
		GUID:    request.GUID,
		Version: request.Version,
		Update: cf.DesiredLRPUpdate{
			Instances:  request.NumInstances,
			Routes:     request.Routes,
			Annotation: lastUpdated,
		},
	}

	return l.Update(ctx, updateRequest)
}

func getURIs(update cf.DesiredLRPUpdate) string {
	cfRouterRoutes, hasRoutes := update.Routes["cf-router"]
	if !hasRoutes {
		return ""
	}

	data, err := cfRouterRoutes.MarshalJSON()
	if err != nil {
		panic("This should never happen")
	}

	return string(data)
}
