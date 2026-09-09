#!/bin/sh
set -eu

kubectl set image deployment/app app=nginx:does-not-exist
