#!/bin/sh
set -eu

kubectl set image deployment/app app=ngnix:1.31.5
