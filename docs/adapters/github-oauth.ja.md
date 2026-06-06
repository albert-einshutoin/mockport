# GitHub OAuth Adapter 日本語版

[English](github-oauth.md)

GitHub OAuth adapter は、local OAuth flow の selected workflow を再現するための adapter です。認可 redirect、token exchange、user profile、emails、orgs などを fake state で扱います。

## 対応範囲

- authorization code flow の成功・失敗 path。
- authorize と token exchange は `client_id` を必須とし、token exchange の `client_id` は code 発行時と一致する必要があります。
- token expiry、scope、redirect URI mismatch などの client contract。`redirect_uri_mismatch` scenario は token exchange の mismatch を表します。
- GitHub policy、SSO、enterprise enforcement、repository permission の完全再現は対象外です。

詳細な endpoint、known gap、contract は英語版を正とします。
