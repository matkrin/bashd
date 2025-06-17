#!/bin/bash

black-on-green() {
    printf "\e[30;42m%s\e[0m" "$1"
}

yellow-bold() {
    printf "\e[93m%s\e[0m" "$1"
}

watch-go-files() {
    msg=$(yellow-bold "Watching .go files")
    echo "$msg"
    fswatch --exclude ".*" --include "\\.go$" . |
        while read -r; do
            clear
            echo "$msg"
            go build . && black-on-green "✓ Build successful"
        done
}

watch-go-files
