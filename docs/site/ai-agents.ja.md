# AI Coding Agent

[English](ai-agents.md)

Mockport は、AI coding agent に実 provider credential を渡さずに integration code を
検証するための local API environment です。実キーを prompt、設定、Issue、log、fixture
へ含めないでください。

```bash
mockport init --adapter stripe --adapter openai
mockport agent-context --adapter stripe --adapter openai >> AGENTS.md
```

`agent-context` は stdout だけに出力し、既存ファイルを作成・上書きしません。出力を確認して
から `AGENTS.md`、`CLAUDE.md`、その他の instruction file へ追加できます。

実装後は `/_mockport/report` の `unsupported_endpoints`、`safety_warnings`、status code を
確認してください。未対応endpointを成功扱いにしたり、実providerへfallbackしたりしないでください。
