#!/bin/sh
# GEN-6: write Gradle dependency verification metadata for the Telegram
# Android build, then commit gradle/verification-metadata.xml.
#
# Usage:
#   sh scripts/gen-gradle-verification.sh
#
# Runs the Android build with --write-verification-metadata sha256 from
# examples/telegram/android. Review the diff (every artifact hash), commit
# the file, and rebuild normally: unsigned or mismatched artifacts then
# fail the build instead of being trusted on TLS alone. Re-run after any
# AGP/dependency bump. Needs the Android SDK; README has the setup.
set -eu

cd examples/telegram/android
./gradlew --write-verification-metadata sha256 assembleDebug
echo "review the diff, then: git add gradle/verification-metadata.xml"
