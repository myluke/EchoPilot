GOPATH=$(shell go env GOPATH)

run: install-deps
	@echo "run app ..."
	@$(MAKE) generate
	@$(MAKE) install 

install-deps:
	@ls $(GOPATH)/bin/gin > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		echo "install gin ..."; \
		go install -mod=mod github.com/codegangsta/gin; \
	fi; \

	@ls $(GOPATH)/bin/codetool > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		echo "install codetool ..."; \
		go install -mod=mod github.com/mylukin/EchoPilot/codetool; \
	fi; \
	
	@ls $(GOPATH)/bin/easyi18n > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		echo "install easyi18n ..."; \
		go install -mod=mod github.com/mylukin/easy-i18n/easyi18n; \
	fi; \

generate: install-deps	
	@export PATH="$(GOPATH)/bin:$(PATH)"; \
	 go mod tidy; \
	 go mod vendor; \
	 go generate

install:
	@go build -v -mod=readonly -buildvcs=false -o ./EchoPilot; \
	 chmod a+x ./EchoPilot && mv ./EchoPilot $(GOPATH)/bin/