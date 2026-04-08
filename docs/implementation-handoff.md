# Implementation Handoff

## What is fixed now

这些已经定下来了：

- `markmaton` 只负责 parser core
- 抓取器在外层
- Go 做解析内核
- Python 做 CLI / package / distribution
- Go 和 Python 之间走 JSON CLI 协议
- 先做 `html -> markdown + metadata + links + images + quality`
- 不把 LLM 功能塞进 v1

测试策略也定下来了：

- 自动化测试优先做 unit tests
- 模块边界优先用 fixtures 和 mocked process 来测
- integration / real-engine smoke 只保留手动路径，需要时再跑

打包边界也定下来了：

- Python CLI 是用户入口
- Go binary 是实际执行体
- local dev 先认 `./bin/markmaton-engine`
- future packaged installs 认 `markmaton/bin/markmaton-engine`

## What the next development plan should cover

下一份真正进入开发的 plan，应该只围绕实现，不再重复做边界定义。

它应该覆盖：

1. Go engine shell
2. request / response model
3. cleanhtml first pass
4. resolve first pass
5. convert first pass
6. postprocess first pass
7. metadata / links / images
8. quality heuristics
9. Python wrapper
10. CLI
11. fixtures + golden tests
12. packaging path

## Recommended milestone order

### Milestone 1 — engine skeleton

- 建 `go.mod`
- 建 `cmd/markmaton-engine/main.go`
- 支持 JSON stdin -> JSON stdout
- 先打通假数据流程

### Milestone 2 — clean + convert happy path

- 输入 HTML
- 做基础清洗
- 转出 Markdown
- 先不追复杂页面

### Milestone 3 — resolve + metadata

- `base href`
- 相对链接
- 图片 src/srcset
- title / description / canonical

### Milestone 4 — quality + fallback

- 做结果评分
- main-content 输出太差时允许 fallback

### Milestone 5 — Python wrapper

- `markmaton.engine`
- `markmaton.models`
- `markmaton.cli`

### Milestone 6 — golden fixtures

- 文章页
- 文档页
- 新闻页
- 论坛页
- 带代码块页
- 带表格页

### Milestone 7 — packaging

- 本地开发流程
- binary 发现逻辑
- PyPI 包装策略

## First build slice

如果只做第一刀，我建议是：

```text
input: article-like HTML
-> cleanhtml
-> convert
-> postprocess
-> markdown + title
```

先拿博客和文档页打穿。

不要一开始就碰：

- 表格
- 奇怪营销页
- 强动态 DOM
- 论坛线程
- 复杂代码高亮

这些留到第二圈。

## Main technical risks

### Risk 1

cleanhtml 太激进，正文被裁掉。

处理：

- 先保守
- 加 quality heuristics
- 支持 full-content fallback

### Risk 2

Markdown 转换层太黑箱。

处理：

- 转换前后都留中间输出
- fixtures + golden tests
- 先让 postprocess 可控

### Risk 3

Python 包装层和 Go 引擎一起长歪。

处理：

- Python 不写 parser 逻辑
- 只做调用和序列化

## What not to decide again

下一轮不要再重新讨论这些：

- 要不要把抓取塞进来
- 要不要先接 LLM
- 要不要上 FFI
- 要不要把它做成大平台

这些这轮已经定了。

下一轮只问：

- 先实现哪一块
- 每块验收标准是什么
- 测试怎么跟上

## Expected next artifact

下一轮应该产出：

- 一份 implementation plan
- 一份 implementation issues CSV

然后才进入正式 coding。
