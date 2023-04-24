# go-oauth

このWebアプリケーションは、Google OAuth 2.0を使用してユーザーを認証し、指定された期間のGoogleカレンダーイベントを表示します。
Go言語で書かれており、GoogleカレンダーAPIを利用しています。

[OAuth2.0を使用したGoogle APIへのアクセス](https://developers.google.com/identity/protocols/oauth2?hl=ja)

## 前提条件

- Go 1.20がインストールされていること
- [Google API Console](https://console.developers.google.com/)で新しいプロジェクトを作成し、GoogleカレンダーAPIを有効にし、認証情報を作成してダウンロードする。ダウンロードしたJSONファイルを、アプリケーションのルートディレクトリに `client_secret.json` として保存してください。
- 作成したプロジェクトの認証情報に、リダイレクトURIとして `http://localhost:8080/callback` を追加してください。
- Oauth同意画面において自身が利用するgoogleアカウントのメールアドレスのテストユーザーを追加してください。
  - (このアプリはローカル環境で動かすことを想定しています。)

## 実行方法

1. アプリケーションを起動する:

```
go run main.go
```

2. Webブラウザで http://localhost:8080 にアクセスし、Googleアカウントでログインしてください。

3. GoogleカレンダーAPIへのアクセス許可を与えると、指定された期間のイベントが表示されます。期間を指定するには、クエリパラメータ `start` と `end` を使用してください。例: `http://localhost:8080/calendar?start=2023-05-01&end=2023-05-31`

