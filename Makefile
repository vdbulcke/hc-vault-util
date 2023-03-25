

.PHONY:	scan
scan: 
	trivy fs . 

.PHONY: build
build: 
	goreleaser build --clean

.PHONY: build-snapshot
build-snapshot: 
	goreleaser build --clean --snapshot --single-target


.PHONY: release-skip-publish
release-skip-publish: 
	goreleaser release --clean --skip-publish  --skip-sign

.PHONY: release-snapshot
release-snapshot: 
	goreleaser release --clean --skip-publish --snapshot  --skip-sign


.PHONY: lint
lint: 
	golangci-lint run ./... 


.PHONY: changelog
changelog: 
	git-chglog -o CHANGELOG.md 

.PHONY: view-changelog
view-changelog: 
	git-chglog 

.PHONY: test
test:
	
	


.PHONY: gen-doc
gen-doc: 
	mkdir -p ./doc
	./dist/hc-vault-util_linux_amd64_v1/hc-vault-util  documentation  --dir ./doc

.PHONY: doc-site
doc-site: 
	podman  run --rm -it -p 8000:8000 -v ${PWD}/www:/docs:z squidfunk/mkdocs-material 
