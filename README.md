# 民俗舞蹈动作谱记复核台（task267-choreoreview）

面向舞蹈人类学研究者的动作谱记复核后端服务：导入田野影像转换出的舞段动作谱记、节拍锚点与队形关系，
服务校验动作连续性、节拍对齐与队形拓扑，生成异读候选，研究者裁决后发布不可变谱记版本。

## 业务闭环

1. 创建舞段项目（organizing）。
2. 导入动作单元（舞者、动作名、起止节拍、连接类型）与影像节拍锚点、队形关系。
3. 全量分析：动作连续性校验（断裂段）、节拍对齐（谱记节拍 vs 影像锚点偏移）、队形拓扑校验（互斥关系冲突）。
4. 物化异读候选，研究者裁决（确认/保留地方变体/否决）。
5. 创建谱记版本（草稿→共享→冻结），封存舞段。

## 标准命令

```bash
# 构建 / 静态检查 / 测试
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go vet   ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go test  ./...

# 端到端自检（创建→分析→裁决→发布→重开验证恢复）
go run ./cmd/task267-choreoreview --smoke-test --db smoke.db

# 启动服务
go run ./cmd/task267-choreoreview --addr :8080 --db choreoreview.db
```

## API 入口

- 舞段：`POST /api/dances`、`GET /api/dances`、`GET /api/dances/{id}`、`PATCH /api/dances/{id}/status`、`POST /api/dances/{id}/seal`
- 动作：`POST /api/dances/{id}/movements`、`GET /api/dances/{id}/movements`
- 节拍：`POST /api/dances/{id}/beats`、`GET /api/dances/{id}/beats`
- 队形：`POST /api/dances/{id}/formations`、`GET /api/dances/{id}/formations`
- 分析：`POST /api/dances/{id}/analyze`、`GET /api/dances/{id}/alignment`、`GET /api/dances/{id}/continuity`、`GET /api/dances/{id}/formation-check`
- 异读：`GET /api/dances/{id}/variants`、`GET /api/dances/{id}/variants/open`、`POST /api/variants/{id}/adjudicate`
- 版本：`POST /api/dances/{id}/versions`、`GET /api/dances/{id}/versions`、`POST /api/versions/{id}/freeze`、`POST /api/versions/{id}/supersede`
- 统计/健康：`GET /api/stats`、`GET /api/health`

## 持久化

SQLite（modernc.org/sqlite 纯 Go 驱动，CGO 无关）。建表：dances / movements / beat_anchors / formations / variants / notation_versions。
WAL 模式、外键开启、唯一约束见迁移；`--smoke-test` 关闭重开同一数据库验证重启恢复。

## 模块责任

- `internal/model`：实体与状态机常量。
- `internal/store`：SQLite 建表迁移与仓储访问。
- `internal/notation`：动作谱记连续性校验。
- `internal/beat`：节拍锚点对齐（节奏估算 + 逐拍偏移）。
- `internal/formation`：队形拓扑校验（互斥关系冲突检测）。
- `internal/variant`：异读候选物化与裁决语义。
- `internal/service`：编排层。
- `internal/httpapi`：HTTP API 层（/api 前缀）。
