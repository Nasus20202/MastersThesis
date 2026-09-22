#!/bin/sh
set -eu

kubectl scale deployment/app --replicas=3
