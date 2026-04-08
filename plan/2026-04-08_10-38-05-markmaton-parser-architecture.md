---
mode: plan
task: markmaton parser architecture
created_at: "2026-04-08T10:38:05-04:00"
complexity: complex
---

# Plan: markmaton parser architecture

## Goal
- 把 `markmaton` 定成一套清楚、可发布、可迭代的结构：Go 做解析内核，Python 做 CLI 和 PyPI 分发，先把 `html -> clean markdown + metadata` 这条主链定义清楚。

## Scope
- In:
  - 研究 Firecrawl 的 HTML 清洗和 Markdown 转换链路
  - 明确 `markmaton` 的模块边界、输入输出、目录结构
  - 明确 Go/Python 的职责分工和通信方式
  - 产出 plan 和 issue CSV，作为后续实现依据
- Out:
  - 这一步不实现正式 parser
  - 不接 Playwright / fetch / no-driver 抓取
  - 不做 LLM extract / summary / json 抽取
  - 不做完整 PyPI 发布流程

## Assumptions / Dependencies
- `markmaton` 是独立 repo，工作区在 `/Users/ac/dev/agents/firecrawl/markmaton`
- `markmaton` 只吃外部已经拿到的 HTML，不负责抓取网页
- Go 内核和 Python wrapper 之间走 JSON stdin/stdout，不走 FFI
- Firecrawl 作为参考实现，不作为代码复制目标
- 参考源码主要来自：
  - `/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/src/scraper/scrapeURL/transformers/index.ts`
  - `/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/src/scraper/scrapeURL/lib/removeUnwantedElements.ts`
  - `/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/src/lib/html-to-markdown.ts`
  - `/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/sharedLibs/go-html-to-md/html-to-markdown.go`
  - `/Users/ac/dev/agents/firecrawl/html-to-markdown`

## Phases
1. Phase 1 — Firecrawl reference audit
   - 确认 Firecrawl 的 `html -> markdown` 主流程、fallback、测试覆盖、强项和短板
2. Phase 2 — markmaton architecture definition
   - 定义模块边界、目录结构、数据模型、Go/Python 职责分工
3. Phase 3 — execution contract
   - 定义 CLI 输入输出、Python 调用方式、测试与 golden fixtures 策略
4. Phase 4 — issue breakdown
   - 把架构工作拆成独立 issue，形成后续执行清单

## Tests & Verification
- Firecrawl 参考链条是否清楚 -> manual: 对照源码调用链与测试文件，确认 clean/convert/postprocess 边界
- markmaton 模块设计是否闭合 -> manual: 检查每个模块是否单一职责、无抓取/LLM 泄漏
- Go/Python 边界是否清楚 -> manual: 检查是否只通过 JSON CLI 协议通信
- issue CSV 是否可执行 -> `python3 /Users/ac/dev/agents/skills/agent-designer/skills/issue-driven-workflow/scripts/validate_issues_csv.py issues/2026-04-08_10-38-05-markmaton-parser-architecture.csv`

## Issue CSV
- Path: `issues/2026-04-08_10-38-05-markmaton-parser-architecture.csv`
- Must share the same timestamp/slug as this plan.
- Column spec: `references/issue-csv-spec.md`

## Acceptance Checklist
- [ ] Firecrawl 参考实现的主链、fallback、短板已经说清楚
- [ ] markmaton 的模块边界和目录结构已经定稿
- [ ] Go/Python 的通信方式和职责分工已经定稿
- [ ] issue CSV 已生成并能通过校验
- [ ] 每一行 issue 都能单独执行和验证

## Risks / Blockers
- Firecrawl 的底层依赖分散在 monorepo + 外部 Go repo，容易只看到 wrapper 看不到核心
- 如果一开始把抓取、解析、抽取混在一起，`markmaton` 很快会变重
- 如果 Python 和 Go 边界不硬，后面分发和测试会变乱

## References
- `/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/src/scraper/scrapeURL/transformers/index.ts`
- `/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/src/scraper/scrapeURL/lib/removeUnwantedElements.ts`
- `/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/src/lib/html-to-markdown.ts`
- `/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/src/lib/html-to-markdown-client.ts`
- `/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/sharedLibs/go-html-to-md/html-to-markdown.go`
- `/Users/ac/dev/agents/firecrawl/firecrawl/apps/go-html-to-md-service/converter.go`
- `/Users/ac/dev/agents/firecrawl/firecrawl/apps/go-html-to-md-service/handler_test.go`
- `/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/src/lib/__tests__/html-transformer.test.ts`
- `/Users/ac/dev/agents/firecrawl/html-to-markdown`

## Tools / MCP
- none

## Rollback / Recovery
- 这一步只写 plan 和 issue 文件，不碰实现代码
- 如果计划拆得不对，直接重写 plan/CSV 即可，不涉及代码回滚

## Checkpoints
- Commit after: plan file written
- Commit after: issues CSV generated and validated
- Commit after: architecture docs accepted
