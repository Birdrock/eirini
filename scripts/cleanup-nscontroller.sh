#!/bin/bash

kubectl delete podsecuritypolicy eirini-namespace-controller
kubectl -n scf delete serviceaccount eirini-namespace-controller
kubectl -n scf delete role eirini-namespace-controller-psp
kubectl -n scf delete rolebinding eirini-namespace-controller-psp
kubectl -n scf delete role eirini-namespace-controller
kubectl -n scf delete rolebinding eirini-namespace-controller

kubectl -n eirini delete serviceaccount eirini-namespace-controller
kubectl -n eirini delete role eirini-namespace-controller-psp
kubectl -n eirini delete rolebinding eirini-namespace-controller-psp
kubectl -n eirini delete role eirini-namespace-controller
kubectl -n eirini delete rolebinding eirini-namespace-controller
