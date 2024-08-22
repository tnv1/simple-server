# Simple websocket server

## 1. Run server

The command runs a websocket server at default url `localhost:8080`

```bash
go run cmd/server/main.go
```

## 2. Run client

The command spawns to 10 connections connect to ws server and log to received id request

```bash
go run cmd/client/main.go
```
