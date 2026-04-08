# Firecrawl Reference Audit

## Purpose

这份文档只做一件事：

- 把 Firecrawl 的 `rawHtml -> clean html -> markdown` 参考实现说清楚
- 判断它哪里扎实，哪里只是能用
- 提炼出 `markmaton` 该借什么、不该背什么

## High-level verdict

Firecrawl 这条线的整体设计是清楚的。

它最值得学的不是“大而全”，而是这三个点：

1. 先清 HTML，再转 Markdown
2. Markdown 是核心中间层，后面的 JSON / summary / query 都是后处理
3. 清洗和转换都有 fallback，不押宝单一实现

但它也不是那种可以整套照搬的实现。

它的两个最重环节都不在同一层里：

- HTML 清洗主要靠 `@mendable/firecrawl-rs`
- HTML 转 Markdown 主要靠 `github.com/firecrawl/html-to-markdown`

所以 Firecrawl 上层更像一层 orchestration，而不是唯一的核心算法。

## Pipeline map

Firecrawl 的主链在这里：

- `/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/src/scraper/scrapeURL/transformers/index.ts`
- `/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/src/scraper/scrapeURL/lib/removeUnwantedElements.ts`
- `/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/src/lib/html-to-markdown.ts`

可以收成这一条：

```text
rawHtml
-> htmlTransform(clean html)
-> parseMarkdown(markdown)
-> postProcessMarkdown
-> 后续 summary / json / query / extract
```

更具体一点：

1. `deriveHTMLFromRawHTML`
   - 调 `htmlTransform(...)`
   - 做 HTML 清洗和内容裁剪

2. `deriveMarkdownFromHTML`
   - 调 `parseMarkdown(...)`
   - 把 clean HTML 变成 Markdown

3. `postProcessMarkdown`
   - 每条转换路径最后都会再跑一遍 Markdown 后处理

4. 后续 transformer
   - `performLLMExtract`
   - `performSummary`
   - `performQuery`

这说明一个很重要的边界：

**Firecrawl 的 Markdown 生成层和 LLM 处理层是分开的。**

这对 `markmaton` 很重要，因为我们也应该保持这个边界。

## HTML cleaning layer

HTML 清洗入口在：

- `/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/src/scraper/scrapeURL/lib/removeUnwantedElements.ts`

它的策略是：

- 优先调 `@mendable/firecrawl-rs` 的 `transformHtml`
- 失败后退到 `cheerio`

这层做的事情包括：

- 去掉 `head` / `meta` / `script` / `style` / `noscript`
- `onlyMainContent` 时移除导航、页脚、社交、cookie、广告类区域
- 支持 `includeTags`
- 支持 `excludeTags`
- 把相对链接和图片地址转绝对
- 处理 `img[srcset]`，尽量选最大图

### Clear strengths

- 职责明确，没把 Markdown 转换混进来
- `onlyMainContent`、`includeTags`、`excludeTags` 三种控制方式都在
- 链接和图片归一化做得早，后面模块更省心

### Weak spots

- 默认排除规则是硬编码列表，维护久了会越来越像补丁仓库
- `force include` 也是站点经验积累，不是通用模型
- `includeTags` 是直接克隆匹配节点，简单有效，但不够语义化
- 文件顶部自己就写着 `TODO: refactor`

### What markmaton should borrow

- `onlyMainContent`
- 显式的 `include/exclude selectors`
- 相对 URL 归一化
- `srcset` 主图选择
- main-content 失败后允许回退到 full-content

### What markmaton should avoid

- 一长串难以解释的硬编码站点特例
- 让清洗层承担太多“业务判断”
- 在清洗层里掺抓取逻辑或 LLM 逻辑

## Markdown conversion layer

转换入口在：

- `/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/src/lib/html-to-markdown.ts`

这里的顺序是：

1. 如果配置了 `HTML_TO_MARKDOWN_SERVICE_URL`
   - 走 HTTP service
2. 否则如果启用了 `USE_GO_MARKDOWN_PARSER`
   - 走 Go 动态库
3. 再不行
   - 退到 `turndown + joplin-turndown-plugin-gfm`

三条路最后都会跑：

- `postProcessMarkdown`

### What is good here

- fallback 顺序很实在
- 把重活从 Node 主线程挪开了
- 对调用方来说只有一个 `parseMarkdown(...)`

### What is not so good

- wrapper 本身比较薄，真正的行为分散在不同后端
- `TODO: add a timeout to the Go parser`
- 文件里有两个没接上的辅助函数，说明这里不是完全收干净的状态
- native 路径用 `process.cwd()` 拼，比较脆

## The real markdown core

真正做转换的，不是 Firecrawl API 层，而是这两个东西：

1. Go wrapper
   - `/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/sharedLibs/go-html-to-md/html-to-markdown.go`
   - `/Users/ac/dev/agents/firecrawl/firecrawl/apps/go-html-to-md-service/converter.go`

2. 外部 Go 库
   - `/Users/ac/dev/agents/firecrawl/html-to-markdown`

Go wrapper 本身很薄，只是：

- `md.NewConverter("", true, nil)`
- `Use(plugin.GitHubFlavored())`
- `Use(plugin.RobustCodeBlock())`
- `ConvertString(html)`

所以真正的规则引擎在：

- `github.com/firecrawl/html-to-markdown`

## What html-to-markdown gives Firecrawl

这个库本身比 Firecrawl 上层更像“真正的 parser core”。

从源码和测试看，它有几件事很值得参考：

- 基于 `goquery` / DOM，不是 regex 硬切
- Converter 有规则系统、插件系统、before/after hooks
- 默认会做一些基础 cleanup
- 有 golden-file 测试思路
- 有大输入 smoke perf tests
- `RobustCodeBlock` 专门补了代码块场景，处理 syntax highlighter 嵌套结构

尤其这两点很值钱：

1. **after hook**
   - 自动 trim
   - 合并多余空行
   - 去 trailing spaces

2. **RobustCodeBlock plugin**
   - 能处理 syntax highlighter 生成的复杂嵌套
   - 能跳过 gutter / line numbers
   - 能探测 language

这说明 Firecrawl 在“技术文章、文档、代码块”场景上，已经不是只靠普通 HTML 转换了。

## Testing signal

Firecrawl API 层自己的 Markdown 测试偏浅：

- `/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/src/lib/__tests__/html-to-markdown.test.ts`

主要测的是：

- 简单段落
- 简单列表
- 坏 HTML

但底层 `html-to-markdown` 仓库的测试味道更好：

- golden / regression 风格
- plugin tests
- code block tests
- perf smoke tests

这意味着：

**Firecrawl 真正靠谱的部分，更多沉在底层库，而不是 API 层 wrapper。**

## Bottom line for markmaton

`markmaton` 不该复制 Firecrawl 的整套结构。

应该借的是：

- 分层边界
- DOM-first 的清洗和转换思路
- Markdown 作为中间表示
- 真实页面 + golden tests
- 针对代码块的专门处理

不该借的是：

- 多后端混杂的 orchestration
- 太多运行时 fallback
- 清洗规则和业务逻辑搅在一起
- 解析层之外的重系统依赖

最适合 `markmaton` 的方向是：

```text
HTML input
-> cleanhtml
-> resolve
-> convert
-> postprocess
-> metadata / links / images / quality
```

然后把抓取和 LLM 全留在外面。
