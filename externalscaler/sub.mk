PROTOFILE_URL = https://raw.githubusercontent.com/kedacore/keda/v2.8.1/pkg/scalers/externalscaler/externalscaler.proto

.PHONY: externalscaler/externalscaler.proto
externalscaler/externalscaler.proto:
	curl -sL $(PROTOFILE_URL) -o $@ -z $@
