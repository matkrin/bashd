#!/usr/bin/env bash

shopt -s globstar

BUILD_CMD="make --no-print-directory build"

cyan() { printf "\e[36m%s\e[0m" "$1"; }
yellow_bright() { printf "\e[93m%s\e[0m" "$1"; }
black_on_green() { printf "\e[30;42m%s\e[0m" "$1"; }

WATCH_MSG=$(yellow_bright "Watching .go files")
BUILD_MSG=$(black_on_green "✓ Build successful")

recompile() {
    echo "$WATCH_MSG"
    echo "$(cyan "$changed_file") changed. Rebuilding..."
    $BUILD_CMD && echo "$BUILD_MSG"
}

watch_go_files_mac() {
    echo "$WATCH_MSG"
    fswatch ./**/*.go |
        while read -r changed_file; do
            clear
            recompile
        done
}

watch_go_files_linux() {
    echo "$WATCH_MSG"
    inotifywait \
        --monitor \
        --recursive \
        -e modify,create,delete,move \
        --exclude '(^|/)\.git/' \
        --format '%w%f' \
        ./**/*.go |
        while read -r changed_file; do
            clear
            recompile
        done
}

if [[ $(uname) == "Linux" ]]; then
    command -v inotifywait &>/dev/null || (echo "'inotifywait' needs to be installed" && exit)
    watch_go_files_linux
elif [[ $(uname) == "Darwin" ]]; then
    command -v fswatch &>/dev/null || (echo "'fswatch' needs to be installed" && exit)
    watch_go_files_mac
fi
