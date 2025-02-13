# 技術ドキュメント: Buy-the-Dips Bot

## 1. プロジェクト概要

### 1.1 プロジェクトの目的と用途
このプロジェクトは、Bybit取引所における自動押し目買いBotを開発することを目的としています。指定された下落率を検知すると自動でエントリーし、保有ポジションを自動で決済することで、効率的な取引を可能にします。

### 1.2 対象読者
- 開発者: Botの機能拡張や保守を行う担当者。
- DevOpsエンジニア: Botのデプロイや運用を行う担当者。
- トレーダー: Botの設定や利用方法を理解し、実際に運用するユーザー。

## 2. まず利用するために (Getting Started)

### 2.1 導入手順

1.  **環境変数の設定**: `.env.local` または `.env.production` ファイルを作成し、必要な環境変数を設定します。以下は必要な環境変数の例です。

    ```
    PORT=8080
    BYBIT=YOUR_BYBIT_API_KEY.YOUR_BYBIT_API_SECRET #ドット区切りでkey.secretと結合
    CATEGORY=linear
    SYMBOL=BTCUSDT
    IS_TEST=true  # テスト環境の場合
    IS_MARKET=false # Limit注文の場合
    IS_EXIT=true # ポジションを決済する場合
    DIPS_RATIO=-0.02 # 下落率（例: -2%）
    SIZEUSD=100 # 一度の取引額（USD）
    ORDERLINKID=your_unique_order_link_id
    ```

2.  **依存関係のインストール**: 以下のコマンドで必要なGoの依存関係をインストールします。 `go mod tidy`を実行してください。

    ```bash
    go mod tidy
    ```

3.  **アプリケーションの起動**: 以下のコマンドでアプリケーションを起動します。

    ```bash
    go run main.go
    ```

4.  **Cloud Runにデプロイする場合**: `deploy.md` に記述されている `gcloud run deploy btd --source .` コマンドを実行します。 事前にgcloud SDKをインストールして、初期設定を済ませてください。

### 2.2 初期設定のヒント

-   `IS_TEST` 環境変数を `true` に設定すると、取引所への実際の注文は行われず、ログに注文内容が出力されるだけになります。なお、設定IDへのキャンセル及び各種情報取得は行われます。
-   `ORDERLINKID` 環境変数は、当プログラムに依る注文を特定するための一意なIDです。同じIDを複数回使用すると、意図しない動作を引き起こす可能性があります。

## 3. ファイル構造図とディレクトリ説明

### 3.1 ファイル構造図 (ツリービュー)

```
├── funcs/
│   ├── buy-the-dips.go
│   ├── buy-the-dip_test.go
│   ├── client.go
│   ├── getter.go
│   └── order.go
├── main.go
├── utils/
│   └── utils.go
├── deploy.md
├── README.md
└── go.mod
```

### 3.2 ディレクトリとファイルの役割詳細

-   **`funcs/`**: 主要なビジネスロジックを含むディレクトリ。
    -   **`buy-the-dips.go`**: メインの処理関数 `BuyTheDips` が含まれています。指定された下落率に基づいて取引判断を行います。
    -   **`buy-the-dip_test.go`**: `buy-the-dips.go` のテストコードが含まれています。
    -   **`client.go`**: Bybit APIクライアントの設定や、環境変数の読み込みを行う関数が含まれています。
    -   **`getter.go`**: OHLCVデータやポジション、ティッカー情報を取得する関数が含まれています。
    -   **`order.go`**: 注文、決済、キャンセル処理を行う関数が含まれています。
-   **`main.go`**: アプリケーションのエントリーポイント。Echoフレームワークを使用してHTTPサーバーを起動し、APIエンドポイントを定義します。
-   **`utils/`**: ユーティリティ関数を含むディレクトリ。文字列操作やテスト用のログ出力など、汎用的な処理を行います。
-   **`deploy.md`**: Cloud Runへのデプロイ手順が記述されています。
-   **`README.md`**: 当ドキュメントです。
-   **`go.mod`**: Goのモジュール管理ファイル。依存関係が記述されています。

### 3.3 ファイル構造の意図

このファイル構造は、機能ごとにファイルを分割し、関心事の分離を重視しています。`funcs/` ディレクトリには取引ロジックが集中しており、`main.go` は取引ロジックを内包し、発火リクエストを受信するプログラムの主体です。`utils/` ディレクトリは、複数の場所で使用される共通関数を提供します。

## 4. 各ファイル・モジュールの詳細

### 4.1 `funcs/buy-the-dips.go`

-   **目的**: 指定値以上の下落を検知した場合にエントリーする主要な関数 `BuyTheDips` を提供します。
-   **主要な関数/クラス**:
    -   **`BuyTheDips(c echo.Context) error`**:
        -   **引数**: `c` (echo.Context): HTTPリクエストのコンテキスト。
        -   **返り値**: `error`: エラーが発生した場合。
        -   **使用例**: `/buy-the-dips` エンドポイントへのリクエストを処理し、下落率を計算してエントリーまたはイグジットの判断を行います。
        -   **エラー処理**: 設定値の取得失敗、APIリクエストの失敗、計算エラーなどを検出し、HTTPステータスコードとエラーメッセージを返します。
    -   **`ArgsForLogic.HasPosition(positions []map[string]any) (float64, bool)`**:
        -   **引数**: `positions` ([]map[string]any): ポジションデータ
        -   **返り値**: `float64`: ポジションサイズ, `bool`: ポジションの有無
        -   **使用例**: 現在のポジションサイズを取得します。
    -   **`ArgsForLogic.IsDip(targetNumber int, ohlcv []Ohlcv) bool`**:
        -   **引数**: `targetNumber` (int): 比較対象のローソク足の本数, `ohlcv` ([]Ohlcv): OHLCVデータ
        -   **返り値**: `bool`: 指定値以上の下落を検知した場合、trueを返します。
        -   **使用例**: 過去のローソク足との価格を比較して、下落率が指定値を超えているかどうかを判断します。
-   **依存関係**:
    -   `context`: タイムアウトやキャンセルを扱うためのコンテキスト。
    -   `fmt`: 文字列フォーマット。
    -   `net/http`: HTTPステータスコード。
    -   `strconv`: 文字列変換。
    -   `github.com/howlrs/buy-the-dips/utils`: ユーティリティ関数。
    -   `github.com/labstack/echo/v4`: HTTPフレームワーク。
-   **ポイント**:
    -   環境変数から設定を読み込み、取引ロジックを実行します。
    -   エラーハンドリングを徹底し、APIからのエラーや計算エラーを適切に処理します。
    -   Bybit APIとのインタラクションを集約し、取引に必要な情報を取得します。

### 4.2 `funcs/client.go`

-   **目的**: Bybit APIクライアントの設定と、環境変数の読み込みを行います。
-   **主要な関数/クラス**:
    -   **`NewArgsForLogic() (*ArgsForLogic, error)`**:
        -   **引数**: なし
        -   **返り値**: `*ArgsForLogic`: 設定値, `error`: エラーが発生した場合
        -   **使用例**: 環境変数から必要な設定値を読み込み、`ArgsForLogic` 構造体を初期化します。
        -   **エラー処理**: 必須の環境変数が設定されていない場合、エラーを返します。
-   **依存関係**:
    -   `fmt`: 文字列フォーマット。
    -   `os`: 環境変数へのアクセス。
    -   `strconv`: 文字列変換。
    -   `github.com/howlrs/buy-the-dips/utils`: ユーティリティ関数。
    -   `github.com/wuhewuhe/bybit.go.api`: Bybit APIクライアント。
-   **ポイント**:
    -   Bybit APIキーを環境変数から安全に読み込みます。
    -   APIクライアントをシングルトンとして提供し、リソースの浪費を防ぎます。

### 4.3 `funcs/getter.go`

-   **目的**: OHLCVデータ、ポジション、ティッカー情報を取得する関数を提供します。
-   **主要な関数/クラス**:
    -   **`ArgsForLogic.GetOHLCV(ctx context.Context, targetNumber int) ([]Ohlcv, error)`**:
        -   **引数**: `ctx` (context.Context): コンテキスト, `targetNumber` (int): 取得するローソク足の本数
        -   **返り値**: `[]Ohlcv`: OHLCVデータ, `error`: エラーが発生した場合
        -   **使用例**: Bybit APIからOHLCVデータを取得し、指定された本数分のデータを返します。
        -   **エラー処理**: APIエラー、データ形式のエラー、タイムアウトなどを検出し、エラーを返します。
    -   **`ArgsForLogic.GetPositions(ctx context.Context) ([]map[string]any, error)`**:
        -   **引数**: `ctx` (context.Context): コンテキスト
        -   **返り値**: `[]map[string]any`: ポジションデータ, `error`: エラーが発生した場合
        -   **使用例**: Bybit APIからポジション情報を取得し、未決済のポジションを返します。
    -   **`ArgsForLogic.GetTicker(ctx context.Context) (*Ticker, error)`**:
        -   **引数**: `ctx` (context.Context): コンテキスト
        -   **返り値**: `*Ticker`: ティッカーデータ, `error`: エラーが発生した場合
        -   **使用例**: Bybit APIからティッカー情報を取得し、最新の価格情報を返します。
-   **依存関係**:
    -   `context`: タイムアウトやキャンセルを扱うためのコンテキスト。
    -   `fmt`: 文字列フォーマット。
    -   `strconv`: 文字列変換。
    -   `time`: 時間操作。
    -   `github.com/howlrs/buy-the-dips/utils`: ユーティリティ関数。
-   **ポイント**:
    -   APIからのレスポンスを適切にパースし、必要なデータを取り出します。
    -   エラーハンドリングを徹底し、APIエラーやデータ形式のエラーを適切に処理します。

### 4.4 `funcs/order.go`

-   **目的**: 注文、決済、キャンセル処理を行う関数を提供します。
-   **主要な関数/クラス**:
    -   **`ArgsForLogic.Entry(ctx context.Context, bestBid string) error`**:
        -   **引数**: `ctx` (context.Context): コンテキスト, `bestBid` (string): 最良買い気配
        -   **返り値**: `error`: エラーが発生した場合
        -   **使用例**: Bybit APIを使用して買い注文を送信します。
        -   **エラー処理**: APIエラー、注文サイズのバリデーションエラーなどを検出し、エラーを返します。
    -   **`ArgsForLogic.Exit(ctx context.Context, tokenSize float64, bestAsk string) error`**:
        -   **引数**: `ctx` (context.Context): コンテキスト, `tokenSize` (string): 注文数量, `bestAsk` (string): 最良売り気配
        -   **返り値**: `error`: エラーが発生した場合
        -   **使用例**: Bybit APIを使用して売り注文を送信し、ポジションを決済します。
        -   **エラー処理**: APIエラー、注文サイズのバリデーションエラーなどを検出し、エラーを返します。
    -   **`ArgsForLogic.Cancel(ctx context.Context) error`**:
        -   **引数**: `ctx` (context.Context): コンテキスト
        -   **返り値**: `error`: エラーが発生した場合
        -   **使用例**: Bybit APIを使用して注文をキャンセルします。
        -   **エラー処理**: APIエラー、注文が見つからないエラーなどを検出し、エラーを返します。
-   **依存関係**:
    -   `context`: タイムアウトやキャンセルを扱うためのコンテキスト。
    -   `fmt`: 文字列フォーマット。
    -   `os`: 環境変数へのアクセス。
    -   `time`: 時間操作。
    -   `github.com/howlrs/buy-the-dips/utils`: ユーティリティ関数。
    -   `github.com/rs/zerolog/log`: ロギング。
-   **ポイント**:
    -   注文の種類（成行/指値）、注文価格、注文数量を適切に設定します。
    -   テスト環境では、APIリクエストを送信せずにログを出力します。

### 4.5 `main.go`

-   **目的**: アプリケーションのエントリーポイントです。
-   **主要な関数/クラス**:
    -   **`main()`**:
        -   **引数**: なし
        -   **返り値**: なし
        -   **使用例**: HTTPサーバーを起動し、APIエンドポイントを定義します。
        -   **エラー処理**: サーバー起動に失敗した場合、エラーログを出力して終了します。
-   **依存関係**:
    -   `fmt`: 文字列フォーマット。
    -   `net/http`: HTTPステータスコード。
    -   `os`: 環境変数へのアクセス。
    -   `runtime`: 実行環境。
    -   `github.com/howlrs/buy-the-dips/funcs`: 関数パッケージ。
    -   `github.com/labstack/echo/v4`: HTTPフレームワーク。
    -   `github.com/labstack/echo/v4/middleware`: ミドルウェア。
    -   `github.com/rs/zerolog`: ロギング。
    -   `github.com/rs/zerolog/log`: ロギング。
    -   `github.com/joho/godotenv`: 環境変数ファイル。
-   **ポイント**:
    -   HTTPサーバーを起動し、APIエンドポイントを定義します。
    -   CORSミドルウェアを有効にし、異なるオリジンからのリクエストを許可します。

### 4.6 `utils/utils.go`

-   **目的**: 汎用的なユーティリティ関数を提供します。
-   **主要な関数/クラス**:
    -   **`Split(env string) []string`**:
        -   **引数**: `env` (string): 分割する文字列
        -   **返り値**: `[]string`: 分割後の文字列
        -   **使用例**: 環境変数を分割します。
    -   **`GetInnerData(result any) map[string]any`**:
        -   **引数**: `result` (any): APIからのレスポンス
        -   **返り値**: `map[string]any`: 内部データ
        -   **使用例**: APIからのレスポンスから内部データを取り出します。
    -   **`GetInnerList(result any) []map[string]any`**:
        -   **引数**: `result` (any): APIからのレスポンス
        -   **返り値**: `[]map[string]any`: 内部リストデータ
        -   **使用例**: APIからのレスポンスから内部リストデータを取り出します。
    -   **`DipsRatio(numbers []float64) float64`**:
        -   **引数**: `numbers` ([]float64): 価格データ
        -   **返り値**: `float64`: 下落率
        -   **使用例**: 価格データから下落率を計算します。
    -   **`ToString(isMarket bool) string`**:
        -   **引数**: `isMarket` (bool): 成行注文かどうか
        -   **返り値**: `string`: 成行注文かどうか
        -   **使用例**: 真偽値から文字列に変換します。
    -   **`ToFloat64(s string) float64`**:
        -   **引数**: `s` (string): 変換する文字列
        -   **返り値**: `float64`: 変換後の浮動小数点数
        -   **使用例**: 文字列を浮動小数点数に変換します。
    -   **`TestLog(isTest bool, format string, a ...any)`**:
        -   **引数**: `isTest` (bool): テストモードかどうか, `format` (string): フォーマット文字列, `a` (...any): 引数
        -   **返り値**: なし
        -   **使用例**: テストモードの場合にログを出力します。
-   **依存関係**:
    -   `fmt`: 文字列フォーマット。
    -   `strconv`: 文字列変換。
    -   `strings`: 文字列操作。
    -   `github.com/rs/zerolog/log`: ロギング。

## 5. 設計アルゴリズムやパターン

### 5.1 主要なアルゴリズム

1.  **下落率計算**: `DipsRatio` 関数で計算された下落率が、`DIPS_RATIO` 環境変数で設定された閾値以下になった場合に買いエントリーを行います。
2.  **ポジション管理**: ポジションの有無をチェックし、ポジションを持っている場合は決済処理を行います。
3.  **注文処理**: Bybit APIを使用して、買い注文または売り注文を送信します。


## 6. 環境構築・セットアップ手順

### 6.1 必要な依存関係のインストール

以下の手順で必要な依存関係をインストールしてください。

1.  Goがインストールされていることを確認してください。
2.  プロジェクトディレクトリに移動します。

    ```bash
    cd [プロジェクトディレクトリ]
    ```

3.  依存関係をダウンロードします。

    ```bash
    go mod tidy
    ```

### 6.2 環境変数の設定

`.env.local` ファイルまたは `.env.production` ファイルを作成し、必要な環境変数を設定します。

### 6.3 開発環境における問題と解決策

-   **問題**: APIキーが正しく設定されていない。
    -   **解決策**: `.env.local` または `.env.production` ファイルに正しいAPIキーが設定されていることを確認してください。
-   **問題**: 依存関係が正しくインストールされていない。
    -   **解決策**: `go mod tidy` コマンドを実行し、依存関係を再インストールしてください。
-   **問題**: テスト環境でAPIリクエストが送信される。
    -   **解決策**: `IS_TEST` 環境変数を `true` に設定してください。

## 7. ベストプラクティスと拡張方法

### 7.1 ベストプラクティス

-   **ロギング**: 重要な処理やエラーが発生した場合には、適切なログを出力するようにしましょう。
-   **エラーハンドリング**: APIからのエラーや計算エラーを適切に処理し、アプリケーションが停止しないようにしましょう。
-   **テスト**: 単体テストや結合テストを記述し、コードの品質を確保しましょう。
-   **セキュリティ**: APIキーなどの機密情報は環境変数で管理し、コードに直接記述しないようにしましょう。

### 7.2 プロジェクトの拡張方法

-   **新しい取引戦略の追加**: ストラテジーパターンを利用して、新しい取引戦略を簡単に追加できるようにしましょう。
-   **異なる取引所への対応**: ファクトリーパターンを利用して、異なる取引所のAPIクライアントを切り替えられるようにしましょう。
-   **自動デプロイ**: CI/CDパイプラインを構築し、コードの変更を自動的にデプロイできるようにしましょう。

### 7.3 新規参加者向け追記

-   **依存ライブラリ**: このプロジェクトでは、`github.com/labstack/echo/v4` (HTTPフレームワーク), `github.com/wuhewuhe/bybit.go.api` (Bybit APIクライアント), `github.com/rs/zerolog` (ロギング) を使用しています。
-   **動作上の特殊な要件**: Cloud Runにデプロイする場合、`ORDERLINKID` 環境変数は必ず設定してください。また、Bybit APIを使用するため、APIキーが必要です。
-   **推奨される使用ケース**: `funcs/buy-the-dips.go` は、指定された下落率に基づいて取引判断を行う場合に利用します。`utils/utils.go` は、複数の場所で使用される共通関数を提供します。
