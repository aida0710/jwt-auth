package tests_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

const (
	baseURL = "http://localhost:8080/api/v1"
)

// テスト用の構造体
type SignUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type AccountResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ProjectResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ヘルパー関数：HTTPリクエストを送信して詳細を表示
func sendRequest(t *testing.T, method, url string, body interface{}, headers map[string]string) (*http.Response, []byte) {
	t.Helper()

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("リクエストボディのマーシャルに失敗: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)

		// リクエストボディを表示
		fmt.Printf("\n📤 リクエストボディ:\n%s\n", prettyJSON(jsonBody))
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		t.Fatalf("リクエスト作成に失敗: %v", err)
	}

	// デフォルトヘッダー
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// カスタムヘッダーを追加
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// リクエスト詳細を表示
	fmt.Printf("\n🔗 リクエスト詳細:\n")
	fmt.Printf("  メソッド: %s\n", method)
	fmt.Printf("  URL: %s\n", url)
	fmt.Printf("  ヘッダー:\n")
	for key, values := range req.Header {
		fmt.Printf("    %s: %s\n", key, strings.Join(values, ", "))
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("リクエスト送信に失敗: %v", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("レスポンスボディの読み取りに失敗: %v", err)
	}

	// レスポンス詳細を表示
	fmt.Printf("\n📥 レスポンス詳細:\n")
	fmt.Printf("  ステータスコード: %d (%s)\n", resp.StatusCode, resp.Status)
	fmt.Printf("  ヘッダー:\n")
	for key, values := range resp.Header {
		fmt.Printf("    %s: %s\n", key, strings.Join(values, ", "))
	}
	fmt.Printf("\n📄 レスポンスボディ:\n%s\n", prettyJSON(respBody))

	return resp, respBody
}

// JSONを整形して表示
func prettyJSON(data []byte) string {
	var result bytes.Buffer
	if err := json.Indent(&result, data, "", "  "); err != nil {
		return string(data)
	}
	return result.String()
}

// JWTのペイロードをデコードして表示
func decodeJWTPayload(token string) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		fmt.Printf("  ⚠️ 無効なJWT形式\n")
		return
	}

	// Base64デコード（パディング追加）
	payload := parts[1]
	if l := len(payload) % 4; l > 0 {
		payload += strings.Repeat("=", 4-l)
	}

	decoded, err := base64URLDecode(payload)
	if err != nil {
		fmt.Printf("  ⚠️ JWTペイロードのデコードに失敗: %v\n", err)
		return
	}

	fmt.Printf("  🔐 JWTペイロード:\n%s\n", prettyJSON(decoded))
}

func base64URLDecode(s string) ([]byte, error) {
	// URL-safe Base64をstandard Base64に変換
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")

	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(s)))
	n, err := base64.StdEncoding.Decode(decoded, []byte(s))
	if err != nil {
		return nil, err
	}
	return decoded[:n], nil
}

func TestE2E_CompleteFlow(t *testing.T) {
	// テスト用のユニークなメールアドレスを生成
	timestamp := time.Now().Unix()
	email := fmt.Sprintf("test_%d@example.com", timestamp)
	password := "SecurePassword123!"
	name := "Test User"

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🧪 JWT認証 E2Eテスト開始")
	fmt.Println(strings.Repeat("=", 60))

	// 1. ヘルスチェック
	t.Run("ヘルスチェック", func(t *testing.T) {
		fmt.Println("\n📋 テスト1: ヘルスチェック")
		fmt.Println(strings.Repeat("-", 40))

		resp, body := sendRequest(t, "GET", baseURL+"/health", nil, nil)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("❌ ヘルスチェック失敗: ステータスコード %d", resp.StatusCode)
		} else {
			fmt.Println("✅ ヘルスチェック成功")
		}

		var healthResp map[string]string
		json.Unmarshal(body, &healthResp)
		if healthResp["status"] == "ok" {
			fmt.Println("✅ サービスは正常に動作しています")
		}
	})

	var accessToken string
	var refreshToken string
	var accountID string

	// 2. サインアップ
	t.Run("サインアップ", func(t *testing.T) {
		fmt.Println("\n📋 テスト2: 新規アカウント作成")
		fmt.Println(strings.Repeat("-", 40))

		signupReq := SignUpRequest{
			Email:    email,
			Password: password,
			Name:     name,
		}

		resp, body := sendRequest(t, "POST", baseURL+"/auth/signup", signupReq, nil)

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("❌ サインアップ失敗: ステータスコード %d", resp.StatusCode)
			return
		}

		var authResp AuthResponse
		if err := json.Unmarshal(body, &authResp); err != nil {
			t.Fatalf("❌ レスポンスのパースに失敗: %v", err)
		}

		accessToken = authResp.AccessToken
		refreshToken = authResp.RefreshToken

		fmt.Println("✅ サインアップ成功")
		fmt.Printf("  アクセストークン長: %d文字\n", len(accessToken))
		fmt.Printf("  リフレッシュトークン長: %d文字\n", len(refreshToken))
		fmt.Printf("  有効期限: %d秒\n", authResp.ExpiresIn)

		// JWTペイロードをデコード
		decodeJWTPayload(accessToken)
	})

	// 3. ログイン
	t.Run("ログイン", func(t *testing.T) {
		fmt.Println("\n📋 テスト3: ログイン")
		fmt.Println(strings.Repeat("-", 40))

		loginReq := LoginRequest{
			Email:    email,
			Password: password,
		}

		resp, body := sendRequest(t, "POST", baseURL+"/auth/login", loginReq, nil)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("❌ ログイン失敗: ステータスコード %d", resp.StatusCode)
			return
		}

		var authResp AuthResponse
		if err := json.Unmarshal(body, &authResp); err != nil {
			t.Fatalf("❌ レスポンスのパースに失敗: %v", err)
		}

		// トークンを更新
		accessToken = authResp.AccessToken
		refreshToken = authResp.RefreshToken

		fmt.Println("✅ ログイン成功")
		decodeJWTPayload(accessToken)
	})

	// 4. アカウント情報取得（認証付き）
	t.Run("アカウント情報取得", func(t *testing.T) {
		fmt.Println("\n📋 テスト4: アカウント情報取得")
		fmt.Println(strings.Repeat("-", 40))

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}

		resp, body := sendRequest(t, "GET", baseURL+"/accounts", nil, headers)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("❌ アカウント情報取得失敗: ステータスコード %d", resp.StatusCode)
			return
		}

		var accounts []AccountResponse
		if err := json.Unmarshal(body, &accounts); err != nil {
			t.Fatalf("❌ レスポンスのパースに失敗: %v", err)
		}

		if len(accounts) > 0 {
			accountID = accounts[0].ID
			fmt.Printf("✅ アカウント情報取得成功: %d件\n", len(accounts))
			fmt.Printf("  アカウントID: %s\n", accountID)
		}
	})

	// 5. プロジェクト作成
	var projectID string
	t.Run("プロジェクト作成", func(t *testing.T) {
		fmt.Println("\n📋 テスト5: プロジェクト作成")
		fmt.Println(strings.Repeat("-", 40))

		projectReq := ProjectRequest{
			Name:        "Test Project",
			Description: "This is a test project",
		}

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}

		resp, body := sendRequest(t, "POST", baseURL+"/projects", projectReq, headers)

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("❌ プロジェクト作成失敗: ステータスコード %d", resp.StatusCode)
			return
		}

		var project ProjectResponse
		if err := json.Unmarshal(body, &project); err != nil {
			t.Fatalf("❌ レスポンスのパースに失敗: %v", err)
		}

		projectID = project.ID
		fmt.Println("✅ プロジェクト作成成功")
		fmt.Printf("  プロジェクトID: %s\n", projectID)
		fmt.Printf("  プロジェクト名: %s\n", project.Name)
		fmt.Printf("  オーナーID: %s\n", project.OwnerID)
	})

	// 6. トークンリフレッシュ
	var newAccessToken string
	var newRefreshToken string
	t.Run("トークンリフレッシュ", func(t *testing.T) {
		fmt.Println("\n📋 テスト6: トークンリフレッシュ")
		fmt.Println(strings.Repeat("-", 40))

		refreshReq := RefreshRequest{
			RefreshToken: refreshToken,
		}

		resp, body := sendRequest(t, "POST", baseURL+"/auth/refresh", refreshReq, nil)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("❌ トークンリフレッシュ失敗: ステータスコード %d", resp.StatusCode)
			return
		}

		var authResp AuthResponse
		if err := json.Unmarshal(body, &authResp); err != nil {
			t.Fatalf("❌ レスポンスのパースに失敗: %v", err)
		}

		newAccessToken = authResp.AccessToken
		newRefreshToken = authResp.RefreshToken

		fmt.Println("✅ トークンリフレッシュ成功")
		fmt.Printf("  新しいアクセストークン長: %d文字\n", len(newAccessToken))
		fmt.Printf("  新しいリフレッシュトークン長: %d文字\n", len(newRefreshToken))
		decodeJWTPayload(newAccessToken)
	})

	// 7. セキュリティテスト：無効なトークン
	t.Run("セキュリティテスト_無効なトークン", func(t *testing.T) {
		fmt.Println("\n📋 テスト7: セキュリティテスト - 無効なトークン")
		fmt.Println(strings.Repeat("-", 40))

		headers := map[string]string{
			"Authorization": "Bearer invalid.token.here",
		}

		resp, _ := sendRequest(t, "GET", baseURL+"/accounts", nil, headers)

		if resp.StatusCode == http.StatusUnauthorized {
			fmt.Println("✅ 無効なトークンは正しく拒否されました")
		} else {
			t.Errorf("❌ セキュリティ問題: 無効なトークンが受け入れられました (ステータス: %d)", resp.StatusCode)
		}
	})

	// 8. セキュリティテスト：トークン再利用検出
	t.Run("セキュリティテスト_トークン再利用", func(t *testing.T) {
		fmt.Println("\n📋 テスト8: セキュリティテスト - リフレッシュトークン再利用検出")
		fmt.Println(strings.Repeat("-", 40))

		// 古いリフレッシュトークンを再利用
		refreshReq := RefreshRequest{
			RefreshToken: refreshToken, // 古いトークン
		}

		resp, body := sendRequest(t, "POST", baseURL+"/auth/refresh", refreshReq, nil)

		if resp.StatusCode == http.StatusUnauthorized {
			fmt.Println("✅ トークン再利用が正しく検出されました")

			var errResp ErrorResponse
			if err := json.Unmarshal(body, &errResp); err == nil {
				if strings.Contains(errResp.Message, "Security alert") || strings.Contains(errResp.Error, "security") {
					fmt.Println("✅ セキュリティアラートが発行されました")
				}
			}
		} else {
			t.Errorf("❌ セキュリティ問題: トークン再利用が検出されませんでした (ステータス: %d)", resp.StatusCode)
		}
	})

	// 9. セキュリティテスト：alg:noneアタック
	t.Run("セキュリティテスト_alg_none", func(t *testing.T) {
		fmt.Println("\n📋 テスト9: セキュリティテスト - alg:noneアタック")
		fmt.Println(strings.Repeat("-", 40))

		// alg:noneのJWTを作成
		headerNone := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
		parts := strings.Split(newAccessToken, ".")
		if len(parts) >= 2 {
			fakeToken := headerNone + "." + parts[1] + "."

			headers := map[string]string{
				"Authorization": "Bearer " + fakeToken,
			}

			resp, _ := sendRequest(t, "GET", baseURL+"/accounts", nil, headers)

			if resp.StatusCode == http.StatusUnauthorized {
				fmt.Println("✅ alg:noneアタックは正しくブロックされました")
			} else {
				t.Errorf("❌ セキュリティ問題: alg:noneアタックがブロックされませんでした (ステータス: %d)", resp.StatusCode)
			}
		}
	})

	// 10. パフォーマンステスト：並行リクエスト
	t.Run("パフォーマンステスト_並行リクエスト", func(t *testing.T) {
		fmt.Println("\n📋 テスト10: パフォーマンステスト - 並行リクエスト")
		fmt.Println(strings.Repeat("-", 40))

		// 新しいログインでトークンを取得
		loginReq := LoginRequest{
			Email:    email,
			Password: password,
		}

		resp, body := sendRequest(t, "POST", baseURL+"/auth/login", loginReq, nil)
		if resp.StatusCode != http.StatusOK {
			t.Skip("ログインに失敗したためパフォーマンステストをスキップ")
		}

		var authResp AuthResponse
		json.Unmarshal(body, &authResp)
		validToken := authResp.AccessToken

		// 10並行でリクエスト
		concurrency := 10
		done := make(chan bool, concurrency)
		startTime := time.Now()

		for i := 0; i < concurrency; i++ {
			go func(id int) {
				headers := map[string]string{
					"Authorization": "Bearer " + validToken,
				}

				client := &http.Client{Timeout: 30 * time.Second}
				req, _ := http.NewRequest("GET", baseURL+"/accounts", nil)
				req.Header.Set("Authorization", headers["Authorization"])

				resp, err := client.Do(req)
				if err != nil {
					fmt.Printf("  ⚠️ リクエスト%d失敗: %v\n", id, err)
				} else {
					fmt.Printf("  ✅ リクエスト%d完了 (ステータス: %d)\n", id, resp.StatusCode)
					resp.Body.Close()
				}

				done <- true
			}(i)
		}

		// 全てのリクエストが完了するまで待機
		for i := 0; i < concurrency; i++ {
			<-done
		}

		elapsed := time.Since(startTime)
		fmt.Printf("\n⏱️ %d並行リクエスト完了時間: %v\n", concurrency, elapsed)
		fmt.Printf("  平均レスポンス時間: %v\n", elapsed/time.Duration(concurrency))
	})

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎉 E2Eテスト完了")
	fmt.Println(strings.Repeat("=", 60))
}

// エラーレスポンスのテスト
func TestE2E_ErrorCases(t *testing.T) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🧪 エラーケースのE2Eテスト")
	fmt.Println(strings.Repeat("=", 60))

	// 1. 重複メールでのサインアップ
	t.Run("重複メールでのサインアップ", func(t *testing.T) {
		fmt.Println("\n📋 エラーテスト1: 重複メールでのサインアップ")
		fmt.Println(strings.Repeat("-", 40))

		email := fmt.Sprintf("duplicate_%d@example.com", time.Now().Unix())

		// 1回目のサインアップ
		signupReq := SignUpRequest{
			Email:    email,
			Password: "Password123!",
			Name:     "First User",
		}

		resp1, _ := sendRequest(t, "POST", baseURL+"/auth/signup", signupReq, nil)
		if resp1.StatusCode != http.StatusCreated {
			t.Skip("初回サインアップに失敗")
		}

		// 2回目のサインアップ（同じメール）
		fmt.Println("\n🔄 同じメールで再度サインアップを試みます...")
		resp2, body := sendRequest(t, "POST", baseURL+"/auth/signup", signupReq, nil)

		if resp2.StatusCode == http.StatusConflict {
			fmt.Println("✅ 重複メールは正しく拒否されました")

			var errResp ErrorResponse
			if err := json.Unmarshal(body, &errResp); err == nil {
				fmt.Printf("  エラーメッセージ: %s\n", errResp.Message)
			}
		} else {
			t.Errorf("❌ 重複メールが受け入れられました (ステータス: %d)", resp2.StatusCode)
		}
	})

	// 2. 弱いパスワードでのサインアップ
	t.Run("弱いパスワードでのサインアップ", func(t *testing.T) {
		fmt.Println("\n📋 エラーテスト2: 弱いパスワードでのサインアップ")
		fmt.Println(strings.Repeat("-", 40))

		weakPasswords := []string{
			"123456",    // 短すぎる
			"password",  // 数字なし
			"Password",  // 数字なし
			"password1", // 大文字なし
			"PASSWORD1", // 小文字なし
		}

		for i, password := range weakPasswords {
			fmt.Printf("\n  テスト %d: パスワード '%s'\n", i+1, password)

			signupReq := SignUpRequest{
				Email:    fmt.Sprintf("weak_%d_%d@example.com", time.Now().Unix(), i),
				Password: password,
				Name:     "Test User",
			}

			resp, body := sendRequest(t, "POST", baseURL+"/auth/signup", signupReq, nil)

			if resp.StatusCode == http.StatusBadRequest {
				fmt.Printf("    ✅ 弱いパスワードは拒否されました\n")

				var errResp ErrorResponse
				if err := json.Unmarshal(body, &errResp); err == nil {
					fmt.Printf("    エラー: %s\n", errResp.Message)
				}
			} else if resp.StatusCode == http.StatusCreated {
				fmt.Printf("    ⚠️ 弱いパスワードが受け入れられました\n")
			}
		}
	})

	// 3. 無効なメール形式
	t.Run("無効なメール形式", func(t *testing.T) {
		fmt.Println("\n📋 エラーテスト3: 無効なメール形式")
		fmt.Println(strings.Repeat("-", 40))

		invalidEmails := []string{
			"notanemail",
			"@example.com",
			"user@",
			"user @example.com",
			"user@.com",
		}

		for i, email := range invalidEmails {
			fmt.Printf("\n  テスト %d: メール '%s'\n", i+1, email)

			signupReq := SignUpRequest{
				Email:    email,
				Password: "ValidPassword123!",
				Name:     "Test User",
			}

			resp, body := sendRequest(t, "POST", baseURL+"/auth/signup", signupReq, nil)

			if resp.StatusCode == http.StatusBadRequest {
				fmt.Printf("    ✅ 無効なメール形式は拒否されました\n")

				var errResp ErrorResponse
				if err := json.Unmarshal(body, &errResp); err == nil {
					fmt.Printf("    エラー: %s\n", errResp.Message)
				}
			} else {
				fmt.Printf("    ❌ 無効なメール形式が受け入れられました (ステータス: %d)\n", resp.StatusCode)
			}
		}
	})

	// 4. 間違ったパスワードでのログイン
	t.Run("間違ったパスワードでのログイン", func(t *testing.T) {
		fmt.Println("\n📋 エラーテスト4: 間違ったパスワードでのログイン")
		fmt.Println(strings.Repeat("-", 40))

		// まず正しいアカウントを作成
		email := fmt.Sprintf("wrong_pass_%d@example.com", time.Now().Unix())
		correctPassword := "CorrectPassword123!"

		signupReq := SignUpRequest{
			Email:    email,
			Password: correctPassword,
			Name:     "Test User",
		}

		resp, _ := sendRequest(t, "POST", baseURL+"/auth/signup", signupReq, nil)
		if resp.StatusCode != http.StatusCreated {
			t.Skip("アカウント作成に失敗")
		}

		// 間違ったパスワードでログイン
		fmt.Println("\n🔐 間違ったパスワードでログインを試みます...")
		loginReq := LoginRequest{
			Email:    email,
			Password: "WrongPassword123!",
		}

		resp, body := sendRequest(t, "POST", baseURL+"/auth/login", loginReq, nil)

		if resp.StatusCode == http.StatusUnauthorized {
			fmt.Println("✅ 間違ったパスワードは正しく拒否されました")

			var errResp ErrorResponse
			if err := json.Unmarshal(body, &errResp); err == nil {
				fmt.Printf("  エラーメッセージ: %s\n", errResp.Message)
			}
		} else {
			t.Errorf("❌ 間違ったパスワードが受け入れられました (ステータス: %d)", resp.StatusCode)
		}
	})

	// 5. 存在しないユーザーでのログイン
	t.Run("存在しないユーザーでのログイン", func(t *testing.T) {
		fmt.Println("\n📋 エラーテスト5: 存在しないユーザーでのログイン")
		fmt.Println(strings.Repeat("-", 40))

		loginReq := LoginRequest{
			Email:    fmt.Sprintf("nonexistent_%d@example.com", time.Now().Unix()),
			Password: "Password123!",
		}

		resp, body := sendRequest(t, "POST", baseURL+"/auth/login", loginReq, nil)

		if resp.StatusCode == http.StatusUnauthorized {
			fmt.Println("✅ 存在しないユーザーは正しく拒否されました")

			var errResp ErrorResponse
			if err := json.Unmarshal(body, &errResp); err == nil {
				fmt.Printf("  エラーメッセージ: %s\n", errResp.Message)
			}
		} else {
			t.Errorf("❌ 存在しないユーザーでログインできました (ステータス: %d)", resp.StatusCode)
		}
	})

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎉 エラーケースのテスト完了")
	fmt.Println(strings.Repeat("=", 60))
}
