# remote-ops fork 的开发与验证

## 本次范围

上游为 `woodpecker-ci/woodpecker`，基线为 `v3.19.0`（`9f5568c20a1f7195d23c4194c49d53bd2580be41`）。开发分支为 `codex/remote-ops-parameters`。保留 Apache-2.0 及上游版权声明。参数基础实现参考 PR #6878 的 `4f663d2d4ec2e5b4744544d02591383ee2490451`；本 fork 适配 3.19.0 的分页与迁移合同，并扩展原生 Deploy。

只有一个 Woodpecker Server。Server 管参数定义、运行记录和调度；workflow 管发布步骤；Agent 执行步骤；Go `next` 保留原命令、校验、确认与状态职责。当前改动仅为控制面参数能力，不修改 TruthLoop 运行代码和 V3 配置。

参数界面使用 [Naive UI](https://www.naiveui.com/) `2.45.3`（MIT、Vue 3），通过具名导入按需使用 Checkbox、Select、Input、Form、Card 和 Modal。范围为原生 manual 页、Deploy 弹窗与参数设置；不另加业务页面。`OpsUiProvider` 接入既有明暗主题，关闭组件库 preflight，不引入全局样式重置；局部布局使用 `ops-` 类名。自由变量编辑通过原 `KeyValueEditor` 的可选渲染插槽更换控件，保留原增删行与重复键校验，其余调用方沿用原渲染。

## 调用与持久化

1. 仓库设置通过原鉴权链路读取、创建、更新或删除参数定义。
2. 手动页读取 manual 定义；Deploy 弹窗读取 deployment 定义及所选父记录。
3. 用户提交仍走原生 CreatePipeline / PostPipeline。
4. Server 校验当前入口的输入。Deploy 先合并父记录变量和显式输入，再写入新记录的 AdditionalVariables；原生目标、任务单独记录，不能只修改编译期变量。
5. 继续使用原配置服务、编译器和调度队列；不改 Parent/SHA，不把界面参数转换为新的 Plan 实体。

## 对抗式 E2E 功能集成测试设计

核心闭环由 `server/api/pipeline_parameters_integration_test.go` 自动化。真实入口是 HTTP handler，并复用实际 MustPush 中间件；使用真实 SQLite migration、仓库参数 CRUD、原 Pipeline 创建、配置持久化、编译与调度。forge 与队列传输采用既有 mock，执行端不接真实 Registry / K8s / PostgreSQL。

测试先创建参数并从 API 读回，提交多服务构建，再从同一记录 Deploy 到 testing，并从该部署记录 Deploy 到 production。断言新记录的 Parent、原 SHA、保存配置、逐次环境/任务/变量，以及真正排队任务包含覆盖后的服务。反向断言原构建未变、未知/重复/空服务与未知目标/任务不会入队、无 push 权限与 AllowDeploy 关闭时仍拒绝。仅响应 200、仅 mock 自己的入参、不检查持久记录或任务均不算闭环通过。

`web/src/components/layout/popups/DeployPipelinePopup.test.ts` 通过真实 Vue / Naive UI 组件验证父构建选择、服务继承、原生目标/任务的请求映射、失败后可纠正、配置加载失败不可提交，以及折叠变量区保留重复键错误、重新打开部署时清除上次校验状态。`OpsVariableEditor.test.ts` 验证换用组件库输入框后原重复键校验和纠正仍生效。`PipelineParameterFields.test.ts` 验证真实 Checkbox 多选、清空和下拉；`pipelineParameters.test.ts` 验证入口隔离和空值语义。后端参数 CRUD 测试覆盖仓库隔离，datastore 回归覆盖删除仓库时清理参数。

这些用例验证控制面输入闭环，不证明真实 Go 发布、同镜像晋级或线上可用；对应业务 FIT 仍留在 TruthLoop 的 remote-release-operations change。

## 测试命令

在根目录执行：

```sh
GOMAXPROCS=4 go test -p 2 -mod=readonly -tags test -count=1 ./server/model ./server/api ./server/store/datastore ./server/pipeline ./server/router
CGO_ENABLED=0 GOMAXPROCS=4 go generate cmd/server/openapi.go
```

在 web 目录执行：

```sh
pnpm install --frozen-lockfile
pnpm typecheck
pnpm exec vitest run --maxWorkers=2
pnpm build
```

首次 Go 路由检查前先完成 UI build，以提供 go:embed 所需的 dist。格式与 lint 仅针对修改文件，不扩大无关代码格式。

## 可能出错点与复核方式

| 风险 | 验证 |
| --- | --- |
| 默认服务被清空后又自动恢复 | 组件交互与 Go 显式空值负例 |
| Deploy 请求覆盖值未保存，记录与任务不一致 | HTTP → SQLite → 编译任务闭环与二次 Deploy |
| deployment 必填项挡住纯构建 | 同名参数按 event 分开；集成用例只带服务构建成功 |
| 参数来自错误仓库或无权修改 | CRUD 跨仓库负例与原 MustPush 权限链 |
| 旧数据库迁移或删除仓库遗漏 | 既有 datastore 回归、新表清理；现场验证只用原实例只读快照的独立副本 |
| 控件替换或折叠导致重复变量名通过、再打开后无法提交 | 真实 Naive UI 输入框、折叠/展开及关闭/重开回归，Deploy 请求与来源断言 |
| Modal 内容超过可视高度或主题不一致 | 隔离预览检查滚动、固定操作区；Provider 复用既有主题，局部样式不重置旧页面 |
| 误把 UI/API 成功当成 Go 发布成功 | 文档明确执行边界，未连接真实发布目标 |

## 尚未交付

当前选项为仓库级静态定义。没有分支联动、备份目录加载、Artifact 服务目录或条件表单框架。没有自动冻结配置、审批、重试或恢复框架。正式部署 fork、Go next 接入与生产启用不包含在本次本地参数开发中。

## 参数基础版本验证（2026-10-19，f8429579）

- 前端全套：12 个测试文件、88 项通过；最终控件文案调整后，相关 3 个文件、6 项再次通过。
- Vue 类型检查、修改文件 ESLint、production UI build、OpenAPI 生成与 Server 编译通过。
- model / api / datastore / pipeline 回归通过，router 编译通过。数据层覆盖参数清理与不同 event 下的同名参数。
- 在原 `18080` 官方测试库的只读快照上，以独立目录和 `18082` 启动 fork；migration 建新表、6 条参数可读写，非法目标返回 422，Pipeline 数量仍为 3。独立 fork 重启后记录保留。
- 浏览器实际检查原生 manual 页和成功构建 #2 的 Deploy 弹窗：服务多选、清空必填禁用、重新勾选恢复、production/cold 下拉可用。未点击发布。
- `127.0.0.1:18083` 仅是本机临时只读验证代理，服务器端复用此前用户通过官方 CLI 登录取得的 API token；不向页面注入 token，拒绝写请求。不是新增产品服务或认证实现，也不代替 OAuth 登录验收。
- 原 `18080` 实例继续响应 204，原控制配置/凭据文件摘要未变，原数据库没有 parameters 新表，也未新增 Pipeline。没有连接远端 V3 或正式发布目标。

本机原始 API / migration / 隔离检查结果及截图位于 `/tmp/woodpecker-remote-ops-next-current-path` 指向的独立临时目录；仓库不保存凭据或数据库副本。临时 fork 使用既有测试 OAuth 配置的独立副本，本轮没有改变 OAuth App 回调或再次宣称新入口登录通过。

## Naive UI 界面改造验证（2026-10-19）

- 使用 Naive UI 2.45.3 的公开组件 API，按需导入；没有改写依赖、全局注册整库或增加独立页面。Provider 使用 abstract，不向原页面插入布局容器，沿用原明暗主题且关闭 preflight。
- 前端全套 13 个测试文件、90 项通过；Vue 类型检查、修改文件 ESLint 和生产构建通过。Server 仅重新编译嵌入的前端资源，没有 Go 源码或数据库变更。
- 真实浏览器检查 manual 服务勾选、全部清空后禁用、重新选择后恢复、自由变量折叠/展开；检查参数设置的列表/编辑表单；检查来源 #2 的 Deploy 环境 production / 任务 cold 下拉、关闭与重新打开。没有提交构建或部署。
- 新增回归覆盖原重复变量名校验，以及折叠保留错误状态、重开 Deploy 重置旧校验；控件值继续映射为原变量与原生环境/任务字段。
- 依赖体积以全部 JS chunks 分别 gzip 后求和统计：引入前 441,210 字节，本轮约 564,700 字节，增加约 120.6 KiB。这是全站所有 JS 文件的总量，不是首页或单路由首包。记录在临时 evidence 的 bundle-before-naive.json / bundle-after-naive.json。
- 隔离检查记录为 naive-isolation.json；页面证据为 naive-manual.png、naive-settings.png、naive-deploy.png。原 18080 配置/凭据摘要和数据库保持不变，仍为 3 条 Pipeline；fork 保持 6 条参数、3 条 Pipeline。V3 和真实发布目标未接触。

暗色 Provider 合同和窄屏样式已接入；本次浏览器截图只覆盖浅色桌面视口，不把其他主题或移动端实测记作通过。
