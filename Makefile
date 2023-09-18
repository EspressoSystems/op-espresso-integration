COMPOSEFLAGS=-d
ITESTS_L2_HOST=http://localhost:9545
BEDROCK_TAGS_REMOTE?=origin
DEVNET_ESPRESSO_FLAGS=--espresso --deploy-config="devnetL1-espresso.json" --deployment="devnetL1-espresso" --devnet-dir=".devnet-espresso" --l2-provider-url="http://localhost:19090"
DEVNET_ESPRESSO_OP2_FLAGS=--espresso --deploy-config="devnetL1-espresso2.json" --deployment="devnetL1-espresso2" --devnet-dir=".devnet-espresso2" --l2-provider-url="http://localhost:29090" --l2="op2"
DEVNET_ESPRESSO_DEMO_FLAGS=--espresso --deploy-config="espresso-demo.json" --deployment="espresso-demo" --devnet-dir=".devnet-espresso" --l2-provider-url="http://localhost:19090" --l2="op --compose-file demo-docker-compose.yml"
monorepo-base := $(realpath .)

build: build-go build-ts
.PHONY: build

build-go: submodules op-node op-proposer op-batcher
.PHONY: build-go

build-ts: submodules
	if [ -n "$$NVM_DIR" ]; then \
		. $$NVM_DIR/nvm.sh && nvm use; \
	fi
	pnpm install
	pnpm build
.PHONY: build-ts

ci-builder:
	docker build -t ci-builder -f ops/docker/ci-builder/Dockerfile .

submodules:
	# CI will checkout submodules on its own (and fails on these commands)
	if [ -z "$$GITHUB_ENV" ]; then \
		git submodule init; \
		git submodule update; \
	fi
.PHONY: submodules

op-bindings:
	make -C ./op-bindings
.PHONY: op-bindings

make op-bindings-docker:
	docker run -v $(monorepo-base):/work -it us-docker.pkg.dev/oplabs-tools-artifacts/images/ci-builder bash -c "env FORGE_BUILD_ARGS='--force' make -C /work op-bindings"
	echo "Asking for root permissions to set owner of files to ${USER} after docker run"
	sudo chown -R ${USER} $(monorepo-base)

.PHONY: op-bindings-docker

op-node:
	make -C ./op-node op-node
.PHONY: op-node

generate-mocks-op-node:
	make -C ./op-node generate-mocks
.PHONY: generate-mocks-op-node

generate-mocks-op-service:
	make -C ./op-service generate-mocks
.PHONY: generate-mocks-op-service

op-batcher:
	make -C ./op-batcher op-batcher
.PHONY: op-batcher

op-proposer:
	make -C ./op-proposer op-proposer
.PHONY: op-proposer

op-challenger:
	make -C ./op-challenger op-challenger
.PHONY: op-challenger

op-program:
	make -C ./op-program op-program
.PHONY: op-program

op-geth-proxy:
	make -C ./op-geth-proxy op-geth-proxy
.PHONY: op-geth-proxy

cannon:
	make -C ./cannon cannon
.PHONY: cannon

cannon-prestate: op-program cannon
	./cannon/bin/cannon load-elf --path op-program/bin/op-program-client.elf --out op-program/bin/prestate.json --meta op-program/bin/meta.json
	./cannon/bin/cannon run --proof-at '=0' --stop-at '=1' --input op-program/bin/prestate.json --meta op-program/bin/meta.json --proof-fmt 'op-program/bin/%d.json' --output /dev/null
	mv op-program/bin/0.json op-program/bin/prestate-proof.json

mod-tidy:
	# Below GOPRIVATE line allows mod-tidy to be run immediately after
	# releasing new versions. This bypasses the Go modules proxy, which
	# can take a while to index new versions.
	#
	# See https://proxy.golang.org/ for more info.
	export GOPRIVATE="github.com/ethereum-optimism" && go mod tidy
.PHONY: mod-tidy

clean:
	rm -rf ./bin
.PHONY: clean

nuke: clean devnet-clean
	git clean -Xdf
.PHONY: nuke

devnet-up:
	@if [ ! -e op-program/bin ]; then \
		make cannon-prestate; \
	fi
	$(shell ./ops/scripts/newer-file.sh .devnet/allocs-l1.json ./packages/contracts-bedrock)
	if [ $(.SHELLSTATUS) -ne 0 ]; then \
		make devnet-allocs; \
	fi
	PYTHONPATH=./bedrock-devnet python3 ./bedrock-devnet/main.py --monorepo-dir=.
.PHONY: devnet-up

devnet-up-espresso:
	$(shell ./ops/scripts/newer-file.sh .devnet-espresso/allocs-l1.json ./packages/contracts-bedrock)
	if [ $(.SHELLSTATUS) -ne 0 ]; then \
		make devnet-allocs-espresso; \
	fi
	PYTHONPATH=./bedrock-devnet python3 ./bedrock-devnet/main.py --monorepo-dir=. $(DEVNET_ESPRESSO_FLAGS)
.PHONY: devnet-up-espresso

devnet-up-espresso2:
	PYTHONPATH=./bedrock-devnet python3 ./bedrock-devnet/main.py --monorepo-dir=. $(DEVNET_ESPRESSO_OP2_FLAGS) --deploy-l2
.PHONY: devnet-up-espresso2

devnet-up-espresso-demo:
	PYTHONPATH=./bedrock-devnet python3 ./bedrock-devnet/main.py --monorepo-dir=. $(DEVNET_ESPRESSO_DEMO_FLAGS) --deploy-l2
.PHONY: devnet-up-espresso-demo

# alias for devnet-up
devnet-up-deploy: devnet-up

devnet-test:
	PYTHONPATH=./bedrock-devnet python3 ./bedrock-devnet/main.py --monorepo-dir=. --test
.PHONY: devnet-test

devnet-test-espresso:
	PYTHONPATH=./bedrock-devnet python3 ./bedrock-devnet/main.py --monorepo-dir=. $(DEVNET_ESPRESSO_FLAGS) --test
.PHONY: devnet-test-espresso

devnet-test-espresso2:
	PYTHONPATH=./bedrock-devnet python3 ./bedrock-devnet/main.py --monorepo-dir=. $(DEVNET_ESPRESSO_OP2_FLAGS) --test
.PHONY: devnet-test-espresso2

devnet-down:
	@(cd ./ops-bedrock && GENESIS_TIMESTAMP=$(shell date +%s) docker compose down -v)
.PHONY: devnet-down

devnet-clean:
	rm -rf ./packages/contracts-bedrock/deployments/devnetL1
	rm -rf ./packages/contracts-bedrock/deployments/devnetL1-espresso
	rm -rf ./packages/contracts-bedrock/deployments/devnetL1-espresso2
	rm -rf ./.devnet
	rm -rf ./.devnet-espresso
	rm -rf ./.devnet-espresso2
	cd ./ops-bedrock && docker compose down -v
	docker image ls 'ops-bedrock*' --format='{{.Repository}}' | xargs -r docker rmi
	docker volume ls --filter name=ops-bedrock --format='{{.Name}}' | xargs -r docker volume rm
.PHONY: devnet-clean

devnet-allocs:
	PYTHONPATH=./bedrock-devnet python3 ./bedrock-devnet/main.py --monorepo-dir=. --allocs

devnet-allocs-espresso:
	PYTHONPATH=./bedrock-devnet python3 ./bedrock-devnet/main.py --monorepo-dir=. $(DEVNET_ESPRESSO_FLAGS) --allocs

devnet-logs:
	@(cd ./ops-bedrock && docker compose logs -f)
.PHONY: devnet-logs

devnet-build:
	@(cd ./ops-bedrock && docker compose build)
.PHONY: devnet-build

devnet-pull:
	@(cd ./ops-bedrock && docker compose pull)
.PHONY: devnet-pull

e2e-pull:
	@(cd ./op-e2e && docker-compose pull)
.PHONY: e2e-pull

test-unit:
	make -C ./op-node test
	make -C ./op-proposer test
	make -C ./op-batcher test
	make -C ./op-e2e test
	pnpm test
.PHONY: test-unit

test-integration:
	bash ./ops-bedrock/test-integration.sh \
		./packages/contracts-bedrock/deployments/devnetL1
.PHONY: test-integration

# Remove the baseline-commit to generate a base reading & show all issues
semgrep:
	$(eval DEV_REF := $(shell git rev-parse develop))
	SEMGREP_REPO_NAME=ethereum-optimism/optimism semgrep ci --baseline-commit=$(DEV_REF)
.PHONY: semgrep

clean-node-modules:
	rm -rf node_modules
	rm -rf packages/**/node_modules

tag-bedrock-go-modules:
	./ops/scripts/tag-bedrock-go-modules.sh $(BEDROCK_TAGS_REMOTE) $(VERSION)
.PHONY: tag-bedrock-go-modules

update-op-geth:
	./ops/scripts/update-op-geth.py
.PHONY: update-op-geth

bedrock-markdown-links:
	docker run --init -it -v `pwd`:/input lycheeverse/lychee --verbose --no-progress --exclude-loopback \
		--exclude twitter.com --exclude explorer.optimism.io --exclude linux-mips.org \
		--exclude-mail /input/README.md "/input/specs/**/*.md"

install-geth:
	go install github.com/ethereum/go-ethereum/cmd/geth@v1.12.0

generate-hotshot-binding:
	forge build --root espresso-sequencer --out ../out --extra-output-files abi
	mv ./out/HotShot.sol/HotShot.abi.json ./op-service/espresso/hotshot
	rm -rf out
	abigen --abi ./op-service/espresso/hotshot/HotShot.abi.json --pkg hotshot --out ./op-service/espresso/hotshot/hotshot.go
	rm ./op-service/espresso/hotshot/HotShot.abi.json

