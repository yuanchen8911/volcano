#!/bin/bash
kubectl delete jobs --all --namespace=default
kubectl delete podgroups --all --namespace=default
kubectl delete pods --all --namespace=default
