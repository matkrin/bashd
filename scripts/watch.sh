#!/usr/bin/env bash

shopt -s globstar

WATCH_PATTERN="./**/*.go"
BUILD_CMD="make --no-print-directory build"

cyan() { printf "\e[36m%s\e[0m" "$1"; }
yellow_bright() { printf "\e[93m%s\e[0m" "$1"; }
black_on_green() { printf "\e[30;42m%s\e[0m" "$1"; }

WATCH_MSG=$(yellow_bright "Watching .go files")
BUILD_MSG=$(black_on_green "✓ Build successful")

recompile() {
    local changed_file="$1"
    echo "$WATCH_MSG"
    echo "$(cyan "$changed_file") changed. Rebuilding..."
    $BUILD_CMD && echo "$BUILD_MSG"
}

watch_go_files_mac() {
    echo "$WATCH_MSG"
    fswatch $WATCH_PATTERN |
        while read -r changed_file; do
            clear
            recompile "$changed_file"
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
        $WATCH_PATTERN |
        while read -r changed_file; do
            clear
            recompile "$changed_file"
        done
}

if [[ $(uname) == "Linux" ]]; then
    command -v inotifywait &>/dev/null || { echo "'inotifywait' needs to be installed" ; exit 1; }
    watch_go_files_linux
elif [[ $(uname) == "Darwin" ]]; then
    command -v fswatch &>/dev/null || { echo "'fswatch' needs to be installed"; exit 1; }
    watch_go_files_mac
fi
