#!/bin/bash -eu

source /build-common.sh

BINARY_NAME="ubackup"
COMPILE_IN_DIRECTORY="cmd/ubackup"
BINTRAY_PROJECT="function61/ubackup"

standardBuildProcess
