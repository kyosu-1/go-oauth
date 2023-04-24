package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// CredentialFile は、Google API Console からダウンロードした OAuth クライアントID JSON ファイルのパスです
const CredentialFile = "client_secret.json"

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
	<title>Google Calendar Events</title>
</head>
<body>
	<h1>Events</h1>
	{{range .}}
		<h3>{{.Summary}}</h3>
		<p>{{.Start.DateTime}}</p>
	{{else}}
		<p>No events.</p>
	{{end}}
</body>
</html>
`

func main() {
	ctx := context.Background()

	// クレデンシャルファイルから設定を読み込む
	data, err := ioutil.ReadFile(CredentialFile)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// 認証情報を取得する
	config, err := google.ConfigFromJSON(data, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	client := getClient(ctx, config)

	// Google Calendar API サービスを作成
	srv, err := calendar.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	http.HandleFunc("/calendar", func(w http.ResponseWriter, r *http.Request) {

		startDate := r.URL.Query().Get("start")
		endDate := r.URL.Query().Get("end")

		// クエリパラメータが設定されていない場合、デフォルトとして今日の日付を使用
		if startDate == "" || endDate == "" {
			t := time.Now()
			startDate = t.Format("2006-01-02")
			endDate = startDate
		}

		// 指定期間のイベントを取得
		events, err := getEvents(ctx, srv, startDate, endDate)
		if err != nil {
			log.Printf("Error fetching events: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// HTML ページにイベント情報を埋め込む
		tmpl := template.Must(template.New("events").Parse(htmlTemplate))
		if err := tmpl.Execute(w, events); err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	})

	// OAuth2 認証フロー用のエンドポイント
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		url := config.AuthCodeURL("state", oauth2.AccessTypeOffline)
		http.Redirect(w, r, url, http.StatusFound)
	})

	// OAuth2 コールバック用のエンドポイント
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		token, err := config.Exchange(ctx, code)
		if err != nil {
			log.Printf("Unable to exchange code for token: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// トークンを保存
		// ここは、DBやKVSなどに保存するのがよいかも
		tokenJson, err := json.Marshal(token)
		if err != nil {
			log.Printf("Unable to marshal token to JSON: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if err := ioutil.WriteFile("token.json", tokenJson, 0600); err != nil {
			log.Printf("Unable to save token: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/calendar", http.StatusFound)
	})

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	// 保存されたトークンを読み込む
	token, err := ioutil.ReadFile("token.json")
	if err == nil {
		var t oauth2.Token
		if err := json.Unmarshal(token, &t); err == nil {
			return config.Client(ctx, &t)
		}
	}

	return &http.Client{}
}

func getEvents(ctx context.Context, srv *calendar.Service, startDate, endDate string) ([]*calendar.Event, error) {
	timeMin := startDate + "T00:00:00Z"
	timeMax := endDate + "T23:59:59Z"

	// 指定期間のイベントを取得
	events, err := srv.Events.List("primary").TimeMin(timeMin).TimeMax(timeMax).SingleEvents(true).OrderBy("startTime").Do()
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve events: %v", err)
	}

	return events.Items, nil
}