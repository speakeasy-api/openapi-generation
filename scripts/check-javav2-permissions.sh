#!/usr/bin/env bash

if [ "$(stat -c '%a' ./zSDKs/sdk-javav2/gradlew)" != "755" ]; then
    echo "./zSDKs/sdk-javav2/gradlew permissions are not 755"
    exit 1
fi

if [ "$(stat -c '%a' ./zSDKs/sdk-javav2/gradlew.bat)" != "755" ]; then
    echo "./zSDKs/sdk-javav2/gradlew.bat permissions are not 755"
    exit 1
fi

echo "Gradle permissions are correct"
