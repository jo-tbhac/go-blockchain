package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"projects/goBlockChain/config"
	"strconv"
	"time"
)

func init() {
	var initialHash []byte
	hash := sha256.Sum256(initialHash)
	CreateBlock(5, hex.EncodeToString(hash[:]), TransactionPool)
}

func getChainHandler(w http.ResponseWriter, r *http.Request) {
	templates := template.Must(template.ParseFiles("app/views/chain.html"))

	if err := templates.Execute(w, Chain); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func createWalletHandler(w http.ResponseWriter, r *http.Request) {
	toAllowAccess(w)
	jsonValue, _ := json.Marshal(CreateWallet())
	w.Write(jsonValue)
}

func transactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		templates := template.Must(template.ParseFiles("app/views/transaction.html"))

		if err := templates.Execute(w, TransactionPool); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	toAllowAccess(w)

	if r.Method == "POST" {
		length, err := strconv.Atoi(r.Header.Get("Content-Length"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		body := make([]byte, length)
		length, err = r.Body.Read(body)

		var tx Transaction

		if err = json.Unmarshal(body[:length], &tx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, isExist := WalletPool[tx.RecipientAddress]; !isExist {
			writeResponse(w, false)
			return
		}

		if wallet := WalletPool[tx.SenderAddress]; tx.AddTransaction(&wallet) {
			writeResponse(w, true)
		} else {
			writeResponse(w, false)
		}
	}
}

func calcTotalAmountHandler(w http.ResponseWriter, r *http.Request) {
	toAllowAccess(w)

	if r.Method == "POST" {
		length, err := strconv.Atoi(r.Header.Get("Content-Length"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		body := make([]byte, length)
		length, err = r.Body.Read(body)

		var wallet Wallet

		if err = json.Unmarshal(body[:length], &wallet); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		amount := CalculateTotalAmount(wallet.BlockchainAddress)
		jsonValue, _ := json.Marshal(map[string]float64{"result": amount})
		w.Write(jsonValue)
	}
}

func toAllowAccess(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
}

func writeResponse(w http.ResponseWriter, result bool) {
	jsonValue, _ := json.Marshal(map[string]bool{"result": result})
	w.Write(jsonValue)
}

// StartMining is mining every 10sec.
func StartMining(wallet *Wallet) {
	fmt.Println("mining to listen on")
	for {
		Mining(wallet)
		time.Sleep(10000 * time.Millisecond)
	}
}

// StartBlockchainServer start blockchain node.
func StartBlockchainServer() error {
	log.Println("Port:8080 to listen on")
	http.HandleFunc("/wallet", createWalletHandler)
	http.HandleFunc("/chain", getChainHandler)
	http.HandleFunc("/transaction", transactionHandler)
	http.HandleFunc("/calc", calcTotalAmountHandler)
	return http.ListenAndServe(fmt.Sprintf(":%d", config.Config.Port), nil)
}
