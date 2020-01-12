package main

import (
    "encoding/json"
    "github.com/btcsuite/btcd/btcjson"
    "github.com/btcsuite/btcd/chaincfg"
    "github.com/btcsuite/btcd/chaincfg/chainhash"
    "github.com/btcsuite/btcd/rpcclient"
    "github.com/btcsuite/btcd/wire"
    "github.com/btcsuite/btcutil"
    "io/ioutil"
    "log"
    "path/filepath"
    "time"
)

// $ git clone git@github.com:btcsuite/btcd.git
// $ cd btcd
// $ git checkout -f v0.20.1-beta
// $ go install -v . ./cmd/...
// $ btcd --rpcuser=test --rpcpass=test --testnet

// See btcd/rpcclient/examples for more examples.

func main() {
    // Only override the handlers for notifications you care about.
    // Also note most of these handlers will only be called if you register
    // for notifications.  See the documentation of the rpcclient
    // NotificationHandlers type for more details about each handler.
    ntfnHandlers := rpcclient.NotificationHandlers{
        OnClientConnected: func() {
            log.Printf("OnClientConnected\n")
        },
        OnBlockConnected: func(hash *chainhash.Hash, height int32, t time.Time) {
            log.Printf("OnBlockConnected\n")
        },
        OnFilteredBlockConnected: func(height int32, header *wire.BlockHeader, txs []*btcutil.Tx) {
            log.Printf("OnFilteredBlockConnected: %v\n", txs)
        },
        OnBlockDisconnected: func(hash *chainhash.Hash, height int32, t time.Time) {
            log.Printf("OnBlockDisconnected\n")
        },
        OnFilteredBlockDisconnected: func(height int32, header *wire.BlockHeader) {
            log.Printf("OnFilteredBlockDisconnected\n")
        },
        OnRecvTx: func(transaction *btcutil.Tx, details *btcjson.BlockDetails) {
            log.Printf("OnRecvTx\n")
        },
        OnRedeemingTx: func(transaction *btcutil.Tx, details *btcjson.BlockDetails) {
            log.Printf("OnRedeemingTx\n")
        },
        OnRelevantTxAccepted: func(transaction []byte) {
            log.Printf("OnRelevantTxAccepted\n")
        },
        OnRescanFinished: func(hash *chainhash.Hash, height int32, blkTime time.Time) {
            log.Printf("OnRescanFinished\n")
        },
        OnRescanProgress: func(hash *chainhash.Hash, height int32, blkTime time.Time) {
            log.Printf("OnRescanProgress\n")
        },
        OnTxAccepted: func(hash *chainhash.Hash, amount btcutil.Amount) {
            log.Printf("OnTxAccepted\n")
        },
        OnTxAcceptedVerbose: func(txDetails *btcjson.TxRawResult) {
            log.Printf("OnTxAcceptedVerbose\n")
        },
        OnBtcdConnected: func(connected bool) {
            log.Printf("OnBtcdConnected\n")
        },
        OnAccountBalance: func(account string, balance btcutil.Amount, confirmed bool) {
            log.Printf("OnAccountBalance\n")
        },
        OnWalletLockState: func(locked bool) {
            log.Printf("OnWalletLockState\n")
        },
        OnUnknownNotification: func(method string, params []json.RawMessage) {
            log.Printf("OnUnknownNotification\n")
        },
    }
    // Connect to local btcd RPC server using websockets.
    btcdHomeDir := btcutil.AppDataDir("btcd", false)
    certs, err := ioutil.ReadFile(filepath.Join(btcdHomeDir, "rpc.cert"))
    if err != nil {
        log.Fatal(err)
    }
    connCfg := &rpcclient.ConnConfig{
        // 8334 for mainnet.
        Host:         "localhost:18334",
        Endpoint:     "ws",
        User:         "test",
        Pass:         "test",
        Certificates: certs,
    }
    client, err := rpcclient.New(connCfg, &ntfnHandlers)
    if err != nil {
        log.Fatal(err)
    }

    recvAddr, err := btcutil.DecodeAddress("moTY3Dk25jUo7n14VY94Z8aUCxiWuVs34h",
        &chaincfg.TestNet3Params)
    if err != nil {
        log.Fatal(recvAddr)
    }
    log.Println(recvAddr)
    filterAddrs := []btcutil.Address{recvAddr}
    if err := client.LoadTxFilter(true, filterAddrs, nil); err != nil {
        log.Fatal(err)
    }

    if err := client.NotifyBlocks(); err != nil {
        log.Fatal(err)
    }
    if err := client.NotifyNewTransactions(true); err != nil {
        log.Fatal(err)
    }

    // For this example gracefully shutdown the client after 10 seconds.
    // Ordinarily when to shutdown the client is highly application
    // specific.
    log.Println("Client shutdown in 6 hours...")
    time.AfterFunc(6*time.Hour, func() {
        log.Println("Client shutting down...")
        client.Shutdown()
        log.Println("Client shutdown complete.")
    })

    // Wait until the client either shuts down gracefully (or the user
    // terminates the process with Ctrl+C).
    client.WaitForShutdown()
}

func example(client *rpcclient.Client) {
    // list accounts
    accounts, err := client.ListAccounts()
    if err != nil {
        log.Fatalf("error listing accounts: %v", err)
    }
    // iterate over accounts (map[string]btcutil.Amount) and write to stdout
    for label, amount := range accounts {
        log.Printf("%s: %s", label, amount)
    }
    // prepare a sendMany transaction
    receiver1, err := btcutil.DecodeAddress("1someAddressThatIsActuallyReal", &chaincfg.MainNetParams)
    if err != nil {
        log.Fatalf("address receiver1 seems to be invalid: %v", err)
    }
    receiver2, err := btcutil.DecodeAddress("1anotherAddressThatsPrettyReal", &chaincfg.MainNetParams)
    if err != nil {
        log.Fatalf("address receiver2 seems to be invalid: %v", err)
    }
    receivers := map[btcutil.Address]btcutil.Amount{
        receiver1: 42, // 42 satoshi
        receiver2: 100, // 100 satoshi
    }
    // create and send the sendMany tx
    txSha, err := client.SendMany("some-account-label-from-which-to-send", receivers)
    if err != nil {
        log.Fatalf("error sendMany: %v", err)
    }
    log.Printf("sendMany completed! tx sha is: %s", txSha.String())
}