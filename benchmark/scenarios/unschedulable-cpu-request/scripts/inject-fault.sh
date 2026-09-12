#!/bin/sh
set -eu

kubectl set resources deployment/app --requests=cpu=100
