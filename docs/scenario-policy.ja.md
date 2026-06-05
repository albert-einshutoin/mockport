# Scenario Policy 日本語版

[English](scenario-policy.md)

scenario は Mockport が local integration test で再現する behavior の単位です。success だけでなく、failure、auth error、rate limit、timeout、webhook/callback を明示的に扱います。

## 方針

- scenario 名は adapter docs と test から追えるようにします。
- provider の全内部状態を再現せず、workflow 検証に必要な範囲に絞ります。
- unsupported behavior は known gap として残します。
