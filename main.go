package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type AccountID string
type SignupRequest struct {
	Id AccountID `json:"id"`
}
type PasswordResetRequest struct {
	Id AccountID `json:"id"`
}
type TransferRequest struct {
	RecipientID AccountID `json:"recipient_id"`
	Amount      string    `json:"amount"`
}
type UserInfo struct {
	Balance  int32
	Password string
}

var flag string
var users = map[AccountID]*UserInfo{}

func init() {
	// フラグ読み込み
	flagBytes, err := os.ReadFile("./flag.txt")
	if err != nil {
		panic(err)
	}
	flag = strings.ReplaceAll(string(flagBytes), "\n", "")

	// タイムゾーン設定
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
	time.Local = jst

	// ユーザーを追加
	users["alice"] = &UserInfo{Balance: 0, Password: generatePassword(time.Now().UnixNano())}
	users["bob"] = &UserInfo{Balance: 0, Password: generatePassword(time.Now().UnixNano())}
}

// パスワードをリセットする
func passwordReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "{\"error\": \"Invalid HTTP method\"}", http.StatusMethodNotAllowed)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var req PasswordResetRequest
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, "{\"error\": \"Invalid request body\"}", http.StatusBadRequest)
		return
	}

	if _, ok := users[req.Id]; !ok {
		http.Error(w, "{\"error\": \"User not found\"}", http.StatusBadRequest)
		return
	}

	users[req.Id].Password = generatePassword(time.Now().Unix())
	w.Write([]byte("{\"success\": true}")) // #nosec G104
}

// 残高表示
func balance(w http.ResponseWriter, r *http.Request) {
	accountID := AccountID(r.Header.Get("X-ID"))
	balance := users[accountID].Balance

	response := fmt.Sprintf("{\"balance\": \"%d\"}", balance)
	if balance > 9999999 {
		response = fmt.Sprintf("{\"balance\": \"%d\", \"flag\": \"%s\"}", balance, flag)
	}

	w.Write([]byte(response)) // #nosec G104
}

// 送金処理
func transfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "{\"error\": \"Invalid HTTP method\"}", http.StatusMethodNotAllowed)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var req TransferRequest
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, "{\"error\": \"Invalid request body\"}", http.StatusBadRequest)
		return
	}

	from := AccountID(r.Header.Get("X-ID"))
	to := req.RecipientID

	amount, err := strconv.Atoi(req.Amount)
	if err != nil {
		http.Error(w, "{\"error\": \"Invalid request body\"}", http.StatusBadRequest)
		return
	}

	// 残高チェック
	if int(users[from].Balance) < amount {
		http.Error(w, "{\"error\": \"Insufficient balance\"}", http.StatusMethodNotAllowed)
		return
	}

	// 送金額のバリデーション
	if int32(amount) < 0 || int32(amount) > 1000000 {
		msg := fmt.Sprintf("{\"error\": \"Amount validation failed: %d\"}", int32(amount))
		http.Error(w, msg, http.StatusMethodNotAllowed)
		return
	}

	// 上限チェック
	if users[to].Balance+int32(amount) > 9999999 {
		msg := fmt.Sprintf("{\"error\": \"Balance validation failed: %d\"}", int32(amount))
		http.Error(w, msg, http.StatusMethodNotAllowed)
		return
	}

	// ユーザーへの入金通知やその他の処理をシミュレート
	time.Sleep(1 * time.Second)

	// 残高の移動
	users[from].Balance = users[from].Balance - int32(amount)
	users[to].Balance = users[to].Balance + int32(amount)

	w.Write([]byte("{\"success\": true}")) // #nosec G104
}

// ユーザーのパスワード生成処理
func generatePassword(seed int64) string {
	rand.Seed(seed)
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	buf := make([]byte, 12)
	for i := range buf {
		buf[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(buf)
}

// 認証用ミドルウェア
func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		accountID := AccountID(r.Header.Get("X-ID"))
		password := r.Header.Get("X-Password")

		v, ok := users[accountID]
		if !ok || password != v.Password {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("{\"error\": \"Authentication failed\"}")) // #nosec G104
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()

	mux.Handle("/password-reset", http.HandlerFunc(passwordReset))
	mux.Handle("/balance", middleware(http.HandlerFunc(balance)))
	mux.Handle("/transfer", middleware(http.HandlerFunc(transfer)))

	http.ListenAndServe(":8080", mux)  // #nosec G104,G114
}
