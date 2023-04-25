#  Standard library misconfiguration
標準ライブラリの利用ミスに関する脆弱性含むアプリケーションです。

## セットアップ
```
$ docker pull ghcr.io/kanmu/gocon2023-office-hour:latest
$ docker run -p 8080:8080 --rm ghcr.io/kanmu/gocon2023-office-hour:latest
```

## 達成条件
アカウントの残高が9,999,999より大きくなった状態で残高確認APIを呼び出すと正解のフラグが出力されます。

正解のフラグは、`kanmu_ctf_2023{xxxxx}`という形式です。

初期状態では、alice / bobの2つのアカウントがあり、残高はどちらも0です。

## アプリケーション説明
攻撃対象はGoで記述されたWeb APIです。
このアプリケーションでは、以下の操作が可能です。

**パスワードリセット ( `/password-reset` )**

指定したアカウントのパスワードがリセットされます。
```
$ curl http://localhost:8080/password-reset \
    -d '{"id": "alice"}'
{"success": true}
```

**残高の確認 ( `/balance` )**

ヘッダで指定したアカウントの残高が返却されます。
```
$ curl http://localhost:8080/balance \
    -H 'X-ID: alice' \
    -H 'X-Password: xxxx'
{"balance": "0"}
```

**残高の送金 ( `/transfer` )**

指定したアカウントから残高が送金されます。
recipient_idには送金先のaccount_idを指定してください。
```
$ curl http://localhost:8080/transfer \
    -H 'X-ID: alice' \
    -H 'X-Password: xxxx' \
    -d '{"recipient_id": "bob", "amount": "9999"}'
{"success": true}
```

## 注意事項
- 64bitプラットフォームを対象とした問題です

## 禁止事項
Dockerコンテナ内に正解のフラグが記述されたファイルがありますが、Dockerコンテナ内を確認してフラグをゲットすることは想定されていません。
curlと、同梱しているexploit.goを使って解けるように設定してあります。

## ヒント (1)

<details>
passwordReset関数 に注目！
`time.Now().Unix()` で初期化しているということは1秒以内に実行すると `rand.Intn` が同じ値になるかも？
</details>

## ヒント (2)

<details>
securego/gosec で何か分かるかも...？
`int32` と `int` が混在しているからオーバーフローの予感…？ ためしに `/transfer` に `amount = -9999999999999` を投げてみたらなにやらエラーがおかしいぞ？
</details>

## ヒント (3)

<details>
TOCTOU で検索！
`transfer` で `time.Sleep` しているところがあるな…？これ同時に叩いたらバグるかも？
</details>
