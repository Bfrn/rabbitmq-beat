BEAT_NAME=rabbitmq-beat
BEAT_PATH=hummer/rabbitmq-beat
BEAT_GOPATH=$(firstword $(subst :, ,${GOPATH}))
SYSTEM_TESTS=false
TEST_ENVIRONMENT=false
ES_BEATS?=./vendor/github.com/elastic/beats
LIBBEAT_MAKEFILE=$(ES_BEATS)/libbeat/scripts/Makefile
GOPACKAGES=$(shell govendor list -no-status +local)
GOBUILD_FLAGS=-i -ldflags "-X $(BEAT_PATH)/vendor/github.com/elastic/beats/libbeat/version.buildTime=$(NOW) -X $(BEAT_PATH)/vendor/github.com/elastic/beats/libbeat/version.commit=$(COMMIT_ID)"
MAGE_IMPORT_PATH=${BEAT_PATH}/vendor/github.com/magefile/mage
NO_COLLECT=true

# Path to the libbeat Makefile
-include $(LIBBEAT_MAKEFILE)

# Initial beat setup
.PHONY: setup
setup: pre-setup

pre-setup:
	$(MAKE) -f $(LIBBEAT_MAKEFILE) mage ES_BEATS=$(ES_BEATS)
	$(MAKE) -f $(LIBBEAT_MAKEFILE) update BEAT_NAME=$(BEAT_NAME) ES_BEATS=$(ES_BEATS) NO_COLLECT=$(NO_COLLECT)