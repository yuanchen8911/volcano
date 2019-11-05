#!/bin/bash
kubectl delete -f default-queue.yaml
kubectl delete -f queue-crd.yaml
kubectl delete -f podgroup-crd.yaml
