#!/bin/bash

# run ./make for help

##############################################################################
# helpers
##############################################################################

assert_clean_repo() {
    if [[ -n $(git status --porcelain) ]]; then
        echo "⛔ Repository has uncommitted changes. Please commit or stash them before running this command."
        exit 1
    fi
}

confirm() {
    read -p "🚨 $1 Are you sure? [y/N] " ans
    [[ ${ans:-N} == "y" ]]
}

##############################################################################
# Q&A commands
##############################################################################

## [qa]: Q&A commands
## qa: run all quality control checks
qa() {
    qa_mod
    qa_test
    qa_fmt
    qa_vet
    qa_staticcheck
    qa_vulncheck
    qa_lint
}

## qa/mod: run go mod tidy and go mod verify
qa_mod() {
    echo "✔︎ Running go mod tidy and go mod verify..."
    go mod tidy -v
    go mod verify
}

## qa/fmt: run gofmt to test if all files are formatted
qa_fmt() {
    echo "✔️ Running gofmt..."
    test -z "$(gofmt -l ./...)"
}

## qa/vet: run go vet
qa_vet() {
    echo "✔️ Running go vet..."
    go vet -stdmethods=false $(go list ./...)
}

## qa/staticcheck: run staticcheck
qa_staticcheck() {
    echo "✔️ Running staticcheck..."
    go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
}

## qa/vulncheck: run govulncheck
qa_vulncheck() {
    echo "✔️ Running govulncheck..."
    go run golang.org/x/vuln/cmd/govulncheck@latest ./...
}

## qa/test: run all tests
qa_test() {
    echo "✔️ Running go test..."
    go test -race -buildvcs .
}

## qa/lint: run golangci-lint
qa_lint() {
    echo "✔️ Running golangci-lint..."
    go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0 run
}

## qa/test/cover: run all tests and create a coverage report
qa_test_cover() {
    echo "✔️ Running go test with coverage..."
    go test -race -buildvcs -coverprofile=art.coverage.prof .
    go tool cover -html=art.coverage.prof -o art.coverage.html
}

## qa/test/update: update golden files
qa_test_update() {
    echo "✔️ Running test golden files update..."
    go test -v -update-golden .
}

## qa/benchmarks: run all benchmarks
qa_benchmarks() {
    echo "✔️ Running go test with benchmarks..."
    go test -benchmem -bench=. -run=^a
}

##############################################################################
# development commands
##############################################################################

## [go]: go commands

## go/tidy: tidy modfiles and format .go files
go_tidy() {
    echo "✔︎ Running go mod tidy..."
    go mod tidy -v
}

## go/fmt: format .go files
go_fmt() {
    echo "✔︎ Running go fmt..."
    go fmt ./...
}

## go/build/asm: build the package and show the assembly output
go_build_asm() {
    echo "✔︎ Running go build asm..."
    go build -a -work -v -gcflags="-S -B -C" .
}

##############################################################################
# operation commands
##############################################################################

## [git]: git commands
## git/push: push changes to the github repository
git_push() {
    echo "✔︎ Running git push..."

    confirm "Pushing changes to the github repository." || exit 1
    qa
    assert_clean_repo
    # git push
}

## git/release: release a new version
git_release() {
    echo "✔︎ Running git release..."

    confirm "Releasing a new version." || exit 1
    qa
    assert_clean_repo
    # GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o="/tmp/bin/linux_amd64/${binary_name}" "$main_package_path"
    # upx -5 "/tmp/bin/linux_amd64/${binary_name}"
    # # Include additional deployment steps here...
}

help() {
    # Define a fixed width for alignment
    local column_width=15

    printf "Usage:\n"
    printf " $0 <command>\n"

    printf "\nCommands:\n"
    printf "\n  %-${column_width}s%s\n" "help" "show this help message"

    current_group=""
    
    while IFS= read -r line; do
        # Detect group lines and capture the description
        if [[ "$line" =~ ^##\ \[([^]]+)\]:\ (.*) ]]; then
            group="[${BASH_REMATCH[1]}]"
            description="${BASH_REMATCH[2]}"
            if [[ "$group" != "$current_group" ]]; then
                printf "\n%-${column_width}s  %s\n" "$group" "$description"
                current_group="$group"
            fi
        # Detect command lines and indent them if within a group
        elif [[ "$line" =~ ^##\ ([^:]+):\ (.*) ]]; then
            command="${BASH_REMATCH[1]}"
            description="${BASH_REMATCH[2]}"
            if [[ -n "$current_group" ]]; then
                printf "  %-${column_width}s%s\n" "$command" "$description"
            else
                printf "%-${column_width}s%s\n" "$command" "$description"
            fi
        fi
    done < "$0"
}

##############################################################################
# main entrypoint
##############################################################################

# Transform input to match function naming convention
# ie: "git/push" -> "git_push"
cmd="${1//\//_}"

# Check if the function exists and execute it, otherwise show help
if declare -f "$cmd" > /dev/null; then
    shift           # remove the first argument
    "$cmd" "$@" # call the function with any additional arguments
else
    if [[ -z "$1" ]]; then
        help
        exit 0
    fi
    echo "ERROR: unknown command: '$1'"
    help
    exit 1
fi
