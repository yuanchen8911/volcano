#!/bin/bash
spark-submit \
--master k8s://https://192.168.99.112:8443 \
--deploy-mode cluster \
--name spark-pi \
--class org.apache.spark.examples.SparkPi \
--conf spark.executor.instances=2 \
--conf spark.kubernetes.container.image=spark:spark \
local:///opt/spark/examples/jars/spark-examples_2.11-2.4.4.jar 2000
