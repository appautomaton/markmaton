# Markmaton Architecture Brief

## What markmaton is

`markmaton` 是一个很轻的 parser core。

它只做：

- 输入一份已经拿到的 HTML
- 清洗页面结构
- 转成干净 Markdown
- 顺手给出 metadata / links / images / quality signals

它不做：

- 网页抓取
- 浏览器控制
- 队列
- LLM 抽取
- summary / query / extract

## Design principles

1. **先把边界钉死**
   - `markmaton` 只吃 HTML
   - 抓取器永远在外面

2. **Go 做内核，Python 做壳**
   - Go：解析、转换、后处理
   - Python：CLI、包分发、调用体验

3. **Python 和 Go 只通过 JSON 通信**
   - 不走 FFI
   - 不走 Python extension
   - 不把逻辑写两份

4. **一条主链先做稳**
   - 先不要三套 fallback
   - 先有一条可靠主路径

5. **用真实页面打磨，不靠玩具输入自我感动**

## Runtime split

### Go responsibilities

- clean HTML
- resolve URLs
- choose image source
- convert HTML to Markdown
- post-process Markdown
- extract metadata
- extract links
- extract images
- compute lightweight quality signals

### Python responsibilities

- 调 Go binary
- 提供 Python API
- 提供 CLI
- 处理打包和分发
- 把结果映射成清楚的数据结构

## Contract

### Input

最小输入：

```json
{
  "url": "https://example.com/article",
  "html": "<html>...</html>"
}
```

语义约束：

- `only_main_content` 缺省时，默认走 `main-content-first`
- 显式传 `only_main_content: false` 时，必须保留调用方意图，不能被默认值逻辑覆盖
- `full-content mode` 不是“原样返回整页 HTML”，而是跳过 main-content 收窄，但仍做全局清洗
- automatic fallback 只是在默认路径结果太弱时触发，语义上不同于显式 full-content mode

可选输入：

```json
{
  "url": "https://example.com/article",
  "html": "<html>...</html>",
  "final_url": "https://www.example.com/article",
  "content_type": "text/html",
  "options": {
    "only_main_content": true,
    "include_selectors": [],
    "exclude_selectors": []
  }
}
```

### Output

```json
{
  "markdown": "...",
  "html_clean": "...",
  "metadata": {
    "title": "...",
    "description": "...",
    "canonical_url": "..."
  },
  "links": ["..."],
  "images": ["..."],
  "quality": {
    "text_length": 12345,
    "link_count": 42,
    "image_count": 7,
    "used_main_content": true,
    "fallback_used": false,
    "quality_score": 0.91
  }
}
```

## Module layout

推荐的 Go 目录是：

```text
cmd/markmaton-engine/
internal/cleanhtml/
internal/resolve/
internal/convert/
internal/postprocess/
internal/metadata/
internal/links/
internal/images/
internal/quality/
internal/model/
```

每一层职责如下。

### `internal/model`

- 请求结构
- 响应结构
- 中间文档对象

### `internal/cleanhtml`

- 删除无关节点
- `only_main_content`
- include / exclude selectors
- 保证后续转换拿到的是合理 HTML
- `only_main_content=false` 时仍保留全局清洗，不等于 raw HTML passthrough

### `internal/resolve`

- `base href`
- 相对链接转绝对
- 图片 `src` / `srcset`

### `internal/convert`

- clean HTML -> Markdown
- 只负责转换，不做业务判断

### `internal/postprocess`

- 空行
- trailing spaces
- 链接和代码块边角修整
- Markdown 语义修补

### `internal/metadata`

- title
- description
- canonical
- OG / Twitter / author / language 这类字段

### `internal/links`

- 页面链接提取

### `internal/images`

- 图片提取

### `internal/quality`

- 判断结果是不是太空
- 判断是否值得从 main-content 回退到 full-content
- 给调用方一个可观测的质量信号
- 不覆盖调用方显式选择的 `full-content mode`

## Recommended repo shape

```text
markmaton/
  pyproject.toml
  README.md
  go.mod
  go.sum

  cmd/
    markmaton-engine/
      main.go

  internal/
    ...

  markmaton/
    __init__.py
    cli.py
    engine.py
    models.py

  docs/
  tests/
  testdata/
```

## Why CLI + JSON

因为这是最稳的组合。

优点：

- Python 和 Go 边界清楚
- 不需要维护 FFI
- 本地开发简单
- CLI 天然可测试
- agent 也容易调
- 以后就算加别的 wrapper，也还是同一个 engine

坏处：

- 需要管理二进制分发

但这个坏处比 FFI 的复杂度轻得多。

## Distribution

主路径：

- Go engine 作为真正执行体
- Python 包作为分发壳
- 未来通过 PyPI / `uv tool` 提供安装体验

不要做的事：

- 不让 Python 侧复制一份转换逻辑
- 不做 npm-first
- 不做 HTTP service-first

## Quality policy

`markmaton` 不该只看“结果是不是空字符串”。

更好的质量信号至少包括：

- `text_length`
- `paragraph_count`
- `link_density`
- `image_count`
- `title_present`
- `quality_score`
- `used_main_content`
- `fallback_used`

这样外层调用方才知道这次结果是：

- 真抓好了
- 还是只是勉强有输出

## Non-goals for v1

- 不做站点级花哨规则系统
- 不做复杂插件市场
- 不做内置 fetch / Playwright
- 不做 LLM 结构化抽取
- 不做“全能 web platform”

v1 只求一件事：

**把一份 HTML 稳稳地变成值得喂给人和模型的 Markdown。**
