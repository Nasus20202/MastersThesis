#!/bin/sh
set -eu

kubectl patch configmap app-config --type=merge -p '{"data":{"message":"beta"}}'
