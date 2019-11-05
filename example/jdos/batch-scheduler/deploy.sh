#!/bin/bash
kubectl delete -f default-queue.yaml
kubectl delete -f scheduling_v1alpha1_queue.yaml
kubectl delete -f scheduling_v1alpha2_queue.yaml
kubectl delete -f scheduling_v1alpha1_podgroup.yaml
kubectl delete -f scheduling_v1alpha2_podgroup.yaml
cd ../../../manifest/batch-scheduler
kubectl apply -f scheduling_v1alpha1_queue.yaml
kubectl apply -f scheduling_v1alpha2_queue.yaml
kubectl apply -f scheduling_v1alpha1_podgroup.yaml
kubectl apply -f scheduling_v1alpha2_podgroup.yaml
kubectl apply -f default-queue.yaml
