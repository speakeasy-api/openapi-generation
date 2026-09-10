# Required variables from parent Makefile:
# EXTRA_ARGS - Additional arguments for commands
# LANG - Target language for snippet generation
# TARGET - Target variant
# USAGE_GROUP - Usage group for snippet generation
# OPERATION - Operation ID for snippet generation

ifdef USAGE_GROUP
# To compile all usage snippets inside a testproject (csharp, go, pythonv2, php, typescriptv2, ruby):
# $ USAGE_GROUP="all" make build-<lang>
# You can also specify a subSDK (tag) instead:
# $ TARGET="primary" USAGE_GROUP="unions" make build-<lang>
override EXTRA_ARGS += -u $(USAGE_GROUP)

# To generate standalone usage snippets for a given subSDK:
# $ TARGET="<target>" USAGE_GROUP="<tag>" LANG="<lang>" make standalone-snippet
standalone-snippet:
	$(call generate-standalone-snippet,$(LANG),$(TARGET),-g $(USAGE_GROUP))
endif

ifdef OPERATION
# To generate a standalone usage snippet for a given operation:
# $ TARGET="<target>" OPERATION="<opID>" LANG="<lang>" make standalone-snippet
standalone-snippet:
	$(call generate-standalone-snippet,$(LANG),$(TARGET),-op $(OPERATION))
endif 
