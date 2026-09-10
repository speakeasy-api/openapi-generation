DETECTED_OS := $(shell sh -c 'uname 2>/dev/null || echo Unknown')
ifdef WSL_DISTRO_NAME
DETECTED_OS := WSL
endif

build-unity-sdks:
ifneq (,$(filter-out review,$(VARIANTS)))
	@EXTRA_ARGS="$(EXTRA_ARGS)" LOG_OUTPUT="$(LOG_OUTPUT)" ./scripts/build-target.sh unity $(filter-out review,$(VARIANTS))
endif
ifneq (,$(findstring review,$(VARIANTS)))
	./scripts/build-review-sdk.sh unity
endif

build-unity: build-unity-sdks
ifdef UNITY_PATH
	for TARGET in $(VARIANTS); do \
		if [ "$$TARGET" != "review" ]; then \
				echo $$TARGET && \
				pushd ./testSDKs/sdk-unity-$$TARGET && \
				find . -type f -regex ".*/bin/Debug/netstandard2.1/.*" -not -regex ".*UnityEngine.*" -exec cp {} ../sdk-unity-$$TARGET-tests/Assets/ \; && \
				popd ; \
		fi ; \
	done
else
	echo "UNITY_PATH is not set"
endif

test-unity: build-unity test-boot
ifdef UNITY_PATH
ifeq ($(DETECTED_OS),WSL)
	cp -r ./testSDKs/sdk-unity-*-tests /mnt/c/workspace/
	for TARGET in $(VARIANTS); do \
		if [ "$$TARGET" != "review" ] && [ "$$TARGET" != "client-credentials" ]; then \
			pushd /mnt/c/workspace/sdk-unity-$$TARGET-tests ; \
			rm test-results.xml || true ; \
			EXE_PATH=$$(dirname "$(UNITY_PATH)") ; \
			LOC=$$(pwd) ; \
			echo "Running tests in $${LOC} with Unity at $${EXE_PATH}" ; \
			"$${EXE_PATH}/Unity.exe" -force-free -runTests -batchmode -projectPath "$$(wslpath -w "$${LOC}")" -testPlatform PlayMode -logFile "$$(wslpath -w "$${LOC}/log.txt")" -testResults "$$(wslpath -w "$${LOC}/test-results.xml")" || exit 1 ; \
			popd ; \
		fi ; \
	done
endif
ifeq ($(DETECTED_OS),Darwin)
	for TARGET in $(VARIANTS); do \
		if [ "$$TARGET" != "review" ] && [ "$$TARGET" != "client-credentials" ]; then \
			pushd ./testSDKs/sdk-unity-$$TARGET-tests ; \
			"$(UNITY_PATH)/MacOS/Unity" -force-free -runTests -batchmode -projectPath "$$(pwd)" -testPlatform PlayMode -logFile "$$(pwd)/log.txt" -testResults "$$(pwd)/test-results.xml" || exit 1 ; \
			popd ; \
		fi ; \
	done
endif
ifeq ($(DETECTED_OS),Linux)
	for TARGET in $(VARIANTS); do \
		if [ "$$TARGET" != "review" ] && [ "$$TARGET" != "client-credentials" ]; then \
			pushd ./testSDKs/sdk-unity-$$TARGET-tests ; \
			"$(UNITY_PATH)/../Unity" -force-free -runTests -batchmode -projectPath "$$(pwd)" -testPlatform PlayMode -logFile "$$(pwd)/log.txt" -testResults "$$(pwd)/test-results.xml" || exit 1 ; \
			popd ; \
		fi ; \
	done
endif
ifneq (,$(filter-out 1,$(words $(VARIANTS))))
	go run ./tests -lang unity || exit 1
endif
else
	echo 'Skipping unity as $$UNITY_PATH not defined'
endif 
