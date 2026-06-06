# LINE Adapter 日本語版

[English](line.md)

LINE adapter は、Messaging API、LINE Login、LIFF helper、MINI App service message、LINE Pay、Mini Dapp helper の local workflow を扱います。

## 対応範囲

- message send、content、signed webhook、rich menu、channel token workflow。
- OAuth code/token/profile と local profile lookup。LINE Login flow では authorize と token exchange の `client_id` を必須とし、token exchange の値は code 発行時と一致する必要があります。
- LIFF browser runtime、provider-driven webhook redelivery、quota enforcement、regional policy、Dapp Portal の完全再現は対象外です。

詳細な endpoint と known gap は英語版を正とします。
