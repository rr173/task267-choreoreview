# BENZHI 评测说明

基于 Go 实现的民俗舞蹈动作谱记复核后端服务，一款后端服务，完成舞段动作谱记导入、动作连续性与队形拓扑校验、影像节拍锚点对齐、异读候选裁决与不可变谱记版本发布。

## 启动

```bash
CGO_ENABLED=0 GOTOOLCHAIN=local go run ./cmd/task267-choreoreview --addr :8080 --db choreoreview.db
```

## 自检（不启动长驻服务）

```bash
go run ./cmd/task267-choreoreview --smoke-test
```

`--smoke-test` 会真实创建舞段、导入动作/锚点/队形、全量分析、裁决候选、冻结版本并封存舞段，关闭并重开同一数据库验证持久化与重启恢复，最后以 0 退出码结束。

## 构建门禁

```bash
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go vet   ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go test  ./...
go run ./cmd/task267-choreoreview --smoke-test
```

## HTTP API（前缀 /api）

舞段：`POST /api/dances`、`GET /api/dances`、`GET /api/dances/{id}`、`PATCH /api/dances/{id}/status`、`POST /api/dances/{id}/seal`
动作：`POST /api/dances/{id}/movements`、`GET /api/dances/{id}/movements`
节拍：`POST /api/dances/{id}/beats`、`GET /api/dances/{id}/beats`
队形：`POST /api/dances/{id}/formations`、`GET /api/dances/{id}/formations`
分析：`POST /api/dances/{id}/analyze`、`GET /api/dances/{id}/alignment`、`GET /api/dances/{id}/continuity`、`GET /api/dances/{id}/formation-check`
异读：`GET /api/dances/{id}/variants`、`GET /api/dances/{id}/variants/open`、`POST /api/variants/{id}/adjudicate`
版本：`POST /api/dances/{id}/versions`、`GET /api/dances/{id}/versions`、`POST /api/versions/{id}/freeze`、`POST /api/versions/{id}/supersede`
统计：`GET /api/stats`、`GET /api/health`

## 持久化

SQLite（modernc.org/sqlite）。表：dances、movements、beat_anchors、formations、variants、notation_versions。动作 (dance,dancer,start_beat) 与候选 (dance,type,ref_id) 幂等；封存舞段后拒绝一切写操作。
