## Overview
---
This project is for engine whatsapp multi-session using library https://github.com/tulir/whatsmeow to use emulate whatsapp web.

### prerequisite
a. gcc (dev essential libs on linux) or using mingw on windows platform
b. golang version >= 1.22

### build up
a. windows platform 
 1. install mingw or gcc from trusted sources like choco or another package manager
 2. build on windows on this command
    ```shell
        set GOOS=windows
        set GOARCH=amd64
        set CGO_ENABLED=1
        go build -o bin/whatsapp_multi_session-windows-amd64.exe
    ```
b. linux or unix platform
 1. install gcc or using `sudo apt install build-essential` 
 2. build on linux to target linux based on your server
    ```shell
        env GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-w -s" -trimpath -o bin/whatsapp_multi_session_with_proxies-linux-amd64
    ```
    or linux to windows platform
    ```shell
        env GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-w -s" -trimpath -o bin/whatsapp_multi_session_with_proxies-windows-amd64.exe
    ```
 3. macOS (using mingw as gcc)
    build on macOS to windows platform
    ```shell
        env GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -ldflags="-w -s" -trimpath bin/whatsapp_multi_session_with_proxies-windows-amd64.exe
    ```
### running up
you can run the source code like
```shell
cd PROJECT_ROOT_FOLDER
```
```shell
cp config.local.yaml.example config.local.yaml
```
```shell
go run main.go
```

a. windows platform, open up a command prompt and just cd to the directory and execute via cmd prompt
```shell
  whatsapp_multi_session_with_proxies-windows-amd64.exe
```

a. linux platform, open up a terminal or tmux and just cd to the directory and execute via terminal
```shell
  ./whatsapp_multi_session_with_proxies-linux-amd64
```

a. unix (freebsd or darwin/macOS), open up a terminal or tmux and just cd to the directory and execute via terminal
```shell
  ./whatsapp_multi_session_with_proxies-freebsd-amd64
```
darwin
```shell
  ./whatsapp_multi_session_with_proxies-darwin-amd64
```
