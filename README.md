# Sway / 言和

Sway（中文名：言和）是一款面向聊天场景的表达增强工具，帮助用户把想说的话改得更自然、更有分寸、更贴近自己的语气。

项目当前处于方案设计阶段。前端、后端、iOS 主 App 与输入法扩展相关文档统一放在本仓库中。

## 文档

- [产品文档](docs/product-requirements.md)
- [产品设计基线与低保真原型](docs/product-design-baseline.md)
- [技术文档](docs/technical-design.md)
- [开发排期](docs/development-plan.md)
- [后端 API v1](docs/backend-api.md)
- [LLM Agent 回归用例](docs/llm-agent-cases.md)
- [iOS 联调说明](ios/README.md)

## 核心原则

- 只做表达增强，不替用户聊天。
- 只生成候选和插入文本，不自动发送。
- 只处理用户主动输入、复制、粘贴或导入的上下文。
- 历史和样本可保存 1 个月，App 需要提供清理入口；`ephemeral` 请求不保存原文。
