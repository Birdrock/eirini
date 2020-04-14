// Code generated by counterfeiter. DO NOT EDIT.
package eirinifakes

import (
	"context"
	"sync"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/eirini"
	"code.cloudfoundry.org/eirini/models/cf"
	"code.cloudfoundry.org/eirini/opi"
)

type FakeBifrost struct {
	GetAppStub        func(context.Context, opi.LRPIdentifier) (*models.DesiredLRP, error)
	getAppMutex       sync.RWMutex
	getAppArgsForCall []struct {
		arg1 context.Context
		arg2 opi.LRPIdentifier
	}
	getAppReturns struct {
		result1 *models.DesiredLRP
		result2 error
	}
	getAppReturnsOnCall map[int]struct {
		result1 *models.DesiredLRP
		result2 error
	}
	GetInstancesStub        func(context.Context, opi.LRPIdentifier) ([]*cf.Instance, error)
	getInstancesMutex       sync.RWMutex
	getInstancesArgsForCall []struct {
		arg1 context.Context
		arg2 opi.LRPIdentifier
	}
	getInstancesReturns struct {
		result1 []*cf.Instance
		result2 error
	}
	getInstancesReturnsOnCall map[int]struct {
		result1 []*cf.Instance
		result2 error
	}
	ListStub        func(context.Context) ([]*models.DesiredLRPSchedulingInfo, error)
	listMutex       sync.RWMutex
	listArgsForCall []struct {
		arg1 context.Context
	}
	listReturns struct {
		result1 []*models.DesiredLRPSchedulingInfo
		result2 error
	}
	listReturnsOnCall map[int]struct {
		result1 []*models.DesiredLRPSchedulingInfo
		result2 error
	}
	StopStub        func(context.Context, opi.LRPIdentifier) error
	stopMutex       sync.RWMutex
	stopArgsForCall []struct {
		arg1 context.Context
		arg2 opi.LRPIdentifier
	}
	stopReturns struct {
		result1 error
	}
	stopReturnsOnCall map[int]struct {
		result1 error
	}
	StopInstanceStub        func(context.Context, opi.LRPIdentifier, uint) error
	stopInstanceMutex       sync.RWMutex
	stopInstanceArgsForCall []struct {
		arg1 context.Context
		arg2 opi.LRPIdentifier
		arg3 uint
	}
	stopInstanceReturns struct {
		result1 error
	}
	stopInstanceReturnsOnCall map[int]struct {
		result1 error
	}
	TransferStub        func(context.Context, cf.DesireLRPRequest) error
	transferMutex       sync.RWMutex
	transferArgsForCall []struct {
		arg1 context.Context
		arg2 cf.DesireLRPRequest
	}
	transferReturns struct {
		result1 error
	}
	transferReturnsOnCall map[int]struct {
		result1 error
	}
	TransferStagingStub        func(context.Context, string, cf.StagingRequest) error
	transferStagingMutex       sync.RWMutex
	transferStagingArgsForCall []struct {
		arg1 context.Context
		arg2 string
		arg3 cf.StagingRequest
	}
	transferStagingReturns struct {
		result1 error
	}
	transferStagingReturnsOnCall map[int]struct {
		result1 error
	}
	TransferTaskStub        func(context.Context, string, cf.TaskRequest) error
	transferTaskMutex       sync.RWMutex
	transferTaskArgsForCall []struct {
		arg1 context.Context
		arg2 string
		arg3 cf.TaskRequest
	}
	transferTaskReturns struct {
		result1 error
	}
	transferTaskReturnsOnCall map[int]struct {
		result1 error
	}
	UpdateStub        func(context.Context, cf.UpdateDesiredLRPRequest) error
	updateMutex       sync.RWMutex
	updateArgsForCall []struct {
		arg1 context.Context
		arg2 cf.UpdateDesiredLRPRequest
	}
	updateReturns struct {
		result1 error
	}
	updateReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeBifrost) GetApp(arg1 context.Context, arg2 opi.LRPIdentifier) (*models.DesiredLRP, error) {
	fake.getAppMutex.Lock()
	ret, specificReturn := fake.getAppReturnsOnCall[len(fake.getAppArgsForCall)]
	fake.getAppArgsForCall = append(fake.getAppArgsForCall, struct {
		arg1 context.Context
		arg2 opi.LRPIdentifier
	}{arg1, arg2})
	fake.recordInvocation("GetApp", []interface{}{arg1, arg2})
	fake.getAppMutex.Unlock()
	if fake.GetAppStub != nil {
		return fake.GetAppStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.getAppReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeBifrost) GetAppCallCount() int {
	fake.getAppMutex.RLock()
	defer fake.getAppMutex.RUnlock()
	return len(fake.getAppArgsForCall)
}

func (fake *FakeBifrost) GetAppCalls(stub func(context.Context, opi.LRPIdentifier) (*models.DesiredLRP, error)) {
	fake.getAppMutex.Lock()
	defer fake.getAppMutex.Unlock()
	fake.GetAppStub = stub
}

func (fake *FakeBifrost) GetAppArgsForCall(i int) (context.Context, opi.LRPIdentifier) {
	fake.getAppMutex.RLock()
	defer fake.getAppMutex.RUnlock()
	argsForCall := fake.getAppArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeBifrost) GetAppReturns(result1 *models.DesiredLRP, result2 error) {
	fake.getAppMutex.Lock()
	defer fake.getAppMutex.Unlock()
	fake.GetAppStub = nil
	fake.getAppReturns = struct {
		result1 *models.DesiredLRP
		result2 error
	}{result1, result2}
}

func (fake *FakeBifrost) GetAppReturnsOnCall(i int, result1 *models.DesiredLRP, result2 error) {
	fake.getAppMutex.Lock()
	defer fake.getAppMutex.Unlock()
	fake.GetAppStub = nil
	if fake.getAppReturnsOnCall == nil {
		fake.getAppReturnsOnCall = make(map[int]struct {
			result1 *models.DesiredLRP
			result2 error
		})
	}
	fake.getAppReturnsOnCall[i] = struct {
		result1 *models.DesiredLRP
		result2 error
	}{result1, result2}
}

func (fake *FakeBifrost) GetInstances(arg1 context.Context, arg2 opi.LRPIdentifier) ([]*cf.Instance, error) {
	fake.getInstancesMutex.Lock()
	ret, specificReturn := fake.getInstancesReturnsOnCall[len(fake.getInstancesArgsForCall)]
	fake.getInstancesArgsForCall = append(fake.getInstancesArgsForCall, struct {
		arg1 context.Context
		arg2 opi.LRPIdentifier
	}{arg1, arg2})
	fake.recordInvocation("GetInstances", []interface{}{arg1, arg2})
	fake.getInstancesMutex.Unlock()
	if fake.GetInstancesStub != nil {
		return fake.GetInstancesStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.getInstancesReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeBifrost) GetInstancesCallCount() int {
	fake.getInstancesMutex.RLock()
	defer fake.getInstancesMutex.RUnlock()
	return len(fake.getInstancesArgsForCall)
}

func (fake *FakeBifrost) GetInstancesCalls(stub func(context.Context, opi.LRPIdentifier) ([]*cf.Instance, error)) {
	fake.getInstancesMutex.Lock()
	defer fake.getInstancesMutex.Unlock()
	fake.GetInstancesStub = stub
}

func (fake *FakeBifrost) GetInstancesArgsForCall(i int) (context.Context, opi.LRPIdentifier) {
	fake.getInstancesMutex.RLock()
	defer fake.getInstancesMutex.RUnlock()
	argsForCall := fake.getInstancesArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeBifrost) GetInstancesReturns(result1 []*cf.Instance, result2 error) {
	fake.getInstancesMutex.Lock()
	defer fake.getInstancesMutex.Unlock()
	fake.GetInstancesStub = nil
	fake.getInstancesReturns = struct {
		result1 []*cf.Instance
		result2 error
	}{result1, result2}
}

func (fake *FakeBifrost) GetInstancesReturnsOnCall(i int, result1 []*cf.Instance, result2 error) {
	fake.getInstancesMutex.Lock()
	defer fake.getInstancesMutex.Unlock()
	fake.GetInstancesStub = nil
	if fake.getInstancesReturnsOnCall == nil {
		fake.getInstancesReturnsOnCall = make(map[int]struct {
			result1 []*cf.Instance
			result2 error
		})
	}
	fake.getInstancesReturnsOnCall[i] = struct {
		result1 []*cf.Instance
		result2 error
	}{result1, result2}
}

func (fake *FakeBifrost) List(arg1 context.Context) ([]*models.DesiredLRPSchedulingInfo, error) {
	fake.listMutex.Lock()
	ret, specificReturn := fake.listReturnsOnCall[len(fake.listArgsForCall)]
	fake.listArgsForCall = append(fake.listArgsForCall, struct {
		arg1 context.Context
	}{arg1})
	fake.recordInvocation("List", []interface{}{arg1})
	fake.listMutex.Unlock()
	if fake.ListStub != nil {
		return fake.ListStub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.listReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeBifrost) ListCallCount() int {
	fake.listMutex.RLock()
	defer fake.listMutex.RUnlock()
	return len(fake.listArgsForCall)
}

func (fake *FakeBifrost) ListCalls(stub func(context.Context) ([]*models.DesiredLRPSchedulingInfo, error)) {
	fake.listMutex.Lock()
	defer fake.listMutex.Unlock()
	fake.ListStub = stub
}

func (fake *FakeBifrost) ListArgsForCall(i int) context.Context {
	fake.listMutex.RLock()
	defer fake.listMutex.RUnlock()
	argsForCall := fake.listArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeBifrost) ListReturns(result1 []*models.DesiredLRPSchedulingInfo, result2 error) {
	fake.listMutex.Lock()
	defer fake.listMutex.Unlock()
	fake.ListStub = nil
	fake.listReturns = struct {
		result1 []*models.DesiredLRPSchedulingInfo
		result2 error
	}{result1, result2}
}

func (fake *FakeBifrost) ListReturnsOnCall(i int, result1 []*models.DesiredLRPSchedulingInfo, result2 error) {
	fake.listMutex.Lock()
	defer fake.listMutex.Unlock()
	fake.ListStub = nil
	if fake.listReturnsOnCall == nil {
		fake.listReturnsOnCall = make(map[int]struct {
			result1 []*models.DesiredLRPSchedulingInfo
			result2 error
		})
	}
	fake.listReturnsOnCall[i] = struct {
		result1 []*models.DesiredLRPSchedulingInfo
		result2 error
	}{result1, result2}
}

func (fake *FakeBifrost) Stop(arg1 context.Context, arg2 opi.LRPIdentifier) error {
	fake.stopMutex.Lock()
	ret, specificReturn := fake.stopReturnsOnCall[len(fake.stopArgsForCall)]
	fake.stopArgsForCall = append(fake.stopArgsForCall, struct {
		arg1 context.Context
		arg2 opi.LRPIdentifier
	}{arg1, arg2})
	fake.recordInvocation("Stop", []interface{}{arg1, arg2})
	fake.stopMutex.Unlock()
	if fake.StopStub != nil {
		return fake.StopStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.stopReturns
	return fakeReturns.result1
}

func (fake *FakeBifrost) StopCallCount() int {
	fake.stopMutex.RLock()
	defer fake.stopMutex.RUnlock()
	return len(fake.stopArgsForCall)
}

func (fake *FakeBifrost) StopCalls(stub func(context.Context, opi.LRPIdentifier) error) {
	fake.stopMutex.Lock()
	defer fake.stopMutex.Unlock()
	fake.StopStub = stub
}

func (fake *FakeBifrost) StopArgsForCall(i int) (context.Context, opi.LRPIdentifier) {
	fake.stopMutex.RLock()
	defer fake.stopMutex.RUnlock()
	argsForCall := fake.stopArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeBifrost) StopReturns(result1 error) {
	fake.stopMutex.Lock()
	defer fake.stopMutex.Unlock()
	fake.StopStub = nil
	fake.stopReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeBifrost) StopReturnsOnCall(i int, result1 error) {
	fake.stopMutex.Lock()
	defer fake.stopMutex.Unlock()
	fake.StopStub = nil
	if fake.stopReturnsOnCall == nil {
		fake.stopReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.stopReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeBifrost) StopInstance(arg1 context.Context, arg2 opi.LRPIdentifier, arg3 uint) error {
	fake.stopInstanceMutex.Lock()
	ret, specificReturn := fake.stopInstanceReturnsOnCall[len(fake.stopInstanceArgsForCall)]
	fake.stopInstanceArgsForCall = append(fake.stopInstanceArgsForCall, struct {
		arg1 context.Context
		arg2 opi.LRPIdentifier
		arg3 uint
	}{arg1, arg2, arg3})
	fake.recordInvocation("StopInstance", []interface{}{arg1, arg2, arg3})
	fake.stopInstanceMutex.Unlock()
	if fake.StopInstanceStub != nil {
		return fake.StopInstanceStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.stopInstanceReturns
	return fakeReturns.result1
}

func (fake *FakeBifrost) StopInstanceCallCount() int {
	fake.stopInstanceMutex.RLock()
	defer fake.stopInstanceMutex.RUnlock()
	return len(fake.stopInstanceArgsForCall)
}

func (fake *FakeBifrost) StopInstanceCalls(stub func(context.Context, opi.LRPIdentifier, uint) error) {
	fake.stopInstanceMutex.Lock()
	defer fake.stopInstanceMutex.Unlock()
	fake.StopInstanceStub = stub
}

func (fake *FakeBifrost) StopInstanceArgsForCall(i int) (context.Context, opi.LRPIdentifier, uint) {
	fake.stopInstanceMutex.RLock()
	defer fake.stopInstanceMutex.RUnlock()
	argsForCall := fake.stopInstanceArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeBifrost) StopInstanceReturns(result1 error) {
	fake.stopInstanceMutex.Lock()
	defer fake.stopInstanceMutex.Unlock()
	fake.StopInstanceStub = nil
	fake.stopInstanceReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeBifrost) StopInstanceReturnsOnCall(i int, result1 error) {
	fake.stopInstanceMutex.Lock()
	defer fake.stopInstanceMutex.Unlock()
	fake.StopInstanceStub = nil
	if fake.stopInstanceReturnsOnCall == nil {
		fake.stopInstanceReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.stopInstanceReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeBifrost) Transfer(arg1 context.Context, arg2 cf.DesireLRPRequest) error {
	fake.transferMutex.Lock()
	ret, specificReturn := fake.transferReturnsOnCall[len(fake.transferArgsForCall)]
	fake.transferArgsForCall = append(fake.transferArgsForCall, struct {
		arg1 context.Context
		arg2 cf.DesireLRPRequest
	}{arg1, arg2})
	fake.recordInvocation("Transfer", []interface{}{arg1, arg2})
	fake.transferMutex.Unlock()
	if fake.TransferStub != nil {
		return fake.TransferStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.transferReturns
	return fakeReturns.result1
}

func (fake *FakeBifrost) TransferCallCount() int {
	fake.transferMutex.RLock()
	defer fake.transferMutex.RUnlock()
	return len(fake.transferArgsForCall)
}

func (fake *FakeBifrost) TransferCalls(stub func(context.Context, cf.DesireLRPRequest) error) {
	fake.transferMutex.Lock()
	defer fake.transferMutex.Unlock()
	fake.TransferStub = stub
}

func (fake *FakeBifrost) TransferArgsForCall(i int) (context.Context, cf.DesireLRPRequest) {
	fake.transferMutex.RLock()
	defer fake.transferMutex.RUnlock()
	argsForCall := fake.transferArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeBifrost) TransferReturns(result1 error) {
	fake.transferMutex.Lock()
	defer fake.transferMutex.Unlock()
	fake.TransferStub = nil
	fake.transferReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeBifrost) TransferReturnsOnCall(i int, result1 error) {
	fake.transferMutex.Lock()
	defer fake.transferMutex.Unlock()
	fake.TransferStub = nil
	if fake.transferReturnsOnCall == nil {
		fake.transferReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.transferReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeBifrost) TransferStaging(arg1 context.Context, arg2 string, arg3 cf.StagingRequest) error {
	fake.transferStagingMutex.Lock()
	ret, specificReturn := fake.transferStagingReturnsOnCall[len(fake.transferStagingArgsForCall)]
	fake.transferStagingArgsForCall = append(fake.transferStagingArgsForCall, struct {
		arg1 context.Context
		arg2 string
		arg3 cf.StagingRequest
	}{arg1, arg2, arg3})
	fake.recordInvocation("TransferStaging", []interface{}{arg1, arg2, arg3})
	fake.transferStagingMutex.Unlock()
	if fake.TransferStagingStub != nil {
		return fake.TransferStagingStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.transferStagingReturns
	return fakeReturns.result1
}

func (fake *FakeBifrost) TransferStagingCallCount() int {
	fake.transferStagingMutex.RLock()
	defer fake.transferStagingMutex.RUnlock()
	return len(fake.transferStagingArgsForCall)
}

func (fake *FakeBifrost) TransferStagingCalls(stub func(context.Context, string, cf.StagingRequest) error) {
	fake.transferStagingMutex.Lock()
	defer fake.transferStagingMutex.Unlock()
	fake.TransferStagingStub = stub
}

func (fake *FakeBifrost) TransferStagingArgsForCall(i int) (context.Context, string, cf.StagingRequest) {
	fake.transferStagingMutex.RLock()
	defer fake.transferStagingMutex.RUnlock()
	argsForCall := fake.transferStagingArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeBifrost) TransferStagingReturns(result1 error) {
	fake.transferStagingMutex.Lock()
	defer fake.transferStagingMutex.Unlock()
	fake.TransferStagingStub = nil
	fake.transferStagingReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeBifrost) TransferStagingReturnsOnCall(i int, result1 error) {
	fake.transferStagingMutex.Lock()
	defer fake.transferStagingMutex.Unlock()
	fake.TransferStagingStub = nil
	if fake.transferStagingReturnsOnCall == nil {
		fake.transferStagingReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.transferStagingReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeBifrost) TransferTask(arg1 context.Context, arg2 string, arg3 cf.TaskRequest) error {
	fake.transferTaskMutex.Lock()
	ret, specificReturn := fake.transferTaskReturnsOnCall[len(fake.transferTaskArgsForCall)]
	fake.transferTaskArgsForCall = append(fake.transferTaskArgsForCall, struct {
		arg1 context.Context
		arg2 string
		arg3 cf.TaskRequest
	}{arg1, arg2, arg3})
	fake.recordInvocation("TransferTask", []interface{}{arg1, arg2, arg3})
	fake.transferTaskMutex.Unlock()
	if fake.TransferTaskStub != nil {
		return fake.TransferTaskStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.transferTaskReturns
	return fakeReturns.result1
}

func (fake *FakeBifrost) TransferTaskCallCount() int {
	fake.transferTaskMutex.RLock()
	defer fake.transferTaskMutex.RUnlock()
	return len(fake.transferTaskArgsForCall)
}

func (fake *FakeBifrost) TransferTaskCalls(stub func(context.Context, string, cf.TaskRequest) error) {
	fake.transferTaskMutex.Lock()
	defer fake.transferTaskMutex.Unlock()
	fake.TransferTaskStub = stub
}

func (fake *FakeBifrost) TransferTaskArgsForCall(i int) (context.Context, string, cf.TaskRequest) {
	fake.transferTaskMutex.RLock()
	defer fake.transferTaskMutex.RUnlock()
	argsForCall := fake.transferTaskArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeBifrost) TransferTaskReturns(result1 error) {
	fake.transferTaskMutex.Lock()
	defer fake.transferTaskMutex.Unlock()
	fake.TransferTaskStub = nil
	fake.transferTaskReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeBifrost) TransferTaskReturnsOnCall(i int, result1 error) {
	fake.transferTaskMutex.Lock()
	defer fake.transferTaskMutex.Unlock()
	fake.TransferTaskStub = nil
	if fake.transferTaskReturnsOnCall == nil {
		fake.transferTaskReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.transferTaskReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeBifrost) Update(arg1 context.Context, arg2 cf.UpdateDesiredLRPRequest) error {
	fake.updateMutex.Lock()
	ret, specificReturn := fake.updateReturnsOnCall[len(fake.updateArgsForCall)]
	fake.updateArgsForCall = append(fake.updateArgsForCall, struct {
		arg1 context.Context
		arg2 cf.UpdateDesiredLRPRequest
	}{arg1, arg2})
	fake.recordInvocation("Update", []interface{}{arg1, arg2})
	fake.updateMutex.Unlock()
	if fake.UpdateStub != nil {
		return fake.UpdateStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.updateReturns
	return fakeReturns.result1
}

func (fake *FakeBifrost) UpdateCallCount() int {
	fake.updateMutex.RLock()
	defer fake.updateMutex.RUnlock()
	return len(fake.updateArgsForCall)
}

func (fake *FakeBifrost) UpdateCalls(stub func(context.Context, cf.UpdateDesiredLRPRequest) error) {
	fake.updateMutex.Lock()
	defer fake.updateMutex.Unlock()
	fake.UpdateStub = stub
}

func (fake *FakeBifrost) UpdateArgsForCall(i int) (context.Context, cf.UpdateDesiredLRPRequest) {
	fake.updateMutex.RLock()
	defer fake.updateMutex.RUnlock()
	argsForCall := fake.updateArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeBifrost) UpdateReturns(result1 error) {
	fake.updateMutex.Lock()
	defer fake.updateMutex.Unlock()
	fake.UpdateStub = nil
	fake.updateReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeBifrost) UpdateReturnsOnCall(i int, result1 error) {
	fake.updateMutex.Lock()
	defer fake.updateMutex.Unlock()
	fake.UpdateStub = nil
	if fake.updateReturnsOnCall == nil {
		fake.updateReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.updateReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeBifrost) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getAppMutex.RLock()
	defer fake.getAppMutex.RUnlock()
	fake.getInstancesMutex.RLock()
	defer fake.getInstancesMutex.RUnlock()
	fake.listMutex.RLock()
	defer fake.listMutex.RUnlock()
	fake.stopMutex.RLock()
	defer fake.stopMutex.RUnlock()
	fake.stopInstanceMutex.RLock()
	defer fake.stopInstanceMutex.RUnlock()
	fake.transferMutex.RLock()
	defer fake.transferMutex.RUnlock()
	fake.transferStagingMutex.RLock()
	defer fake.transferStagingMutex.RUnlock()
	fake.transferTaskMutex.RLock()
	defer fake.transferTaskMutex.RUnlock()
	fake.updateMutex.RLock()
	defer fake.updateMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeBifrost) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ eirini.Bifrost = new(FakeBifrost)
