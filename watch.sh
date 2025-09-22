#!/usr/bin/env bash

WATCH_DIR="."
BUILD_CMD="go build"

black-on-green() {
    printf "\e[30;42m%s\e[0m" "$1"
}

yellow-bold() {
    printf "\e[93m%s\e[0m" "$1"
}

WATCH_MSG=$(yellow-bold "Watching .go files")
BUILD_MSG=$(black-on-green "✓ Build successful")

watch-go-files-mac() {
    echo "$WATCH_MSG"
    # fswatch --exclude ".*" --include "\\.go$" $WATCH_DIR |
    fswatch ./*.go ./**/*.go |
        while read -r changed_file; do
            clear
            echo "$WATCH_MSG"
            echo "$changed_file changed. Rebuilding..."
            go build . && echo "$BUILD_MSG"
        done
}

watch-go-files-linux() {
    while true; do
        changed_file=$(inotifywait \
            -q \
            -e modify,create,delete,move \
            -r --exclude '(^|/)\.git/' \
            --format '%w%f' \
            "$WATCH_DIR" \
        )

        if [[ "$changed_file" == *.go ]]; then
            clear
            echo "$changed_file changed. Rebuilding..."
            $BUILD_CMD . && echo "$BUILD_MSG"
            echo "$WATCH_MSG"
        fi
    done
}

if [[ $(uname) == "Linux" ]]; then
    command -v inotifywait &>/dev/null || (echo "'inotifywait' needs to be installed" && exit)
    echo "$WATCH_MSG"
    watch-go-files-linux
elif [[ $(uname) == "Darwin" ]]; then
    command -v fswatch &>/dev/null || (echo "'fswatch' needs to be installed" && exit)
    watch-go-files-mac
fi
