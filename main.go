name: Collector

on:
  workflow_dispatch:
  schedule:
    - cron: '0 */5 * * *'

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
    - name: Checkout repository
      uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: 1.17

    - name: Fetch Telegram messages
      run: go run ./path/to/fetch_telegram.go  # Update with correct path if necessary

    - name: Build and run Golang file
      run: go run ./path/to/main.go  # Update with correct path if necessary

    - name: List generated files
      run: ls -la ./  # List files created

    - name: Commit Changes
      run: |
        git config --local user.email "actions@github.com"
        git config --local user.name "GitHub Actions"

        if [[ -f ./vmess_iran.txt ]]; then git add ./vmess_iran.txt; fi
        if [[ -f ./trojan_iran.txt ]]; then git add ./trojan_iran.txt; fi
        if [[ -f ./vless_iran.txt ]]; then git add ./vless_iran.txt; fi
        if [[ -f ./ss_iran.txt ]]; then git add ./ss_iran.txt; fi
        if [[ -f ./mixed_iran.txt ]]; then git add ./mixed_iran.txt; fi
        
        git commit -m "✔️ $(date '+%Y-%m-%d %H:%M:%S') Collected" || echo "No changes to commit"

    - name: Push Changes
      uses: ad-m/github-push-action@v0.6.0

      with:
        branch: main
