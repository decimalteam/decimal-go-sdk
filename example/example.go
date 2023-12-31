package main

// Testing example for decimal go sdk

import (
	"bytes"
	//"io/fs"
	"os"

	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/go-resty/resty/v2" // use resty.Client for debugging

	"net/http"

	decapi "bitbucket.org/decimalteam/decimal-go-sdk/api"
	"bitbucket.org/decimalteam/decimal-go-sdk/wallet"
)

const (
	testMnemonicWords              = "repair furnace west loud peasant false six hockey poem tube now alien service phone hazard winter favorite away sand fuel describe version tragic vendor"
	testMnemonicPassphrase         = ""
	testSenderAddress              = "dx12k95ukkqzjhkm9d94866r4d9fwx7tsd82r8pjd"
	testReceiverAddress            = "dx1yzxrvpj807dzs5mapwpu77zuh4669lltjheqvv"
	testStakesMakerAddress         = "dx1g0gf9gdcyyrqk9rhnm23l7dsm7v5ex8c3ga98p"
	testValidatorAddress           = "dxvaloper16rr3cvdgj8jsywhx8lfteunn9uz0xg2czw6gx5"
	testMultisigParticipantAddress = "dx173lnn7jjuym5rwp23aufhnwshylrdemcswtcg5"
	testMultisigAddress            = "dx1nkujpc7fj72cfdyrtj7f090wgdakjvnyy6dak5"
	testCoin                       = "tdel"
	testTxHash                     = "F7BEE024F6EECD0909EF90B5C6A46FE6AFD7AEF061EE02BB73F800960EF57326"
	testNFTTokenId                 = "rt7c7255cd002f1595a8d8a00ce11ffce25a315t"

	testWrongSenderAddress              = "dx12k95ukkqzjhkm9d94866r4d9fwx7tsd82r8p00"
	testWrongReceiverAddress            = "dx1yzxrvpj807dzs5mapwpu77zuh4669lltjheq00"
	testWrongStakesMakerAddress         = "dx1dqx544dw3gfc2q2n0yv0ghdsjq79zlaf000000"
	testWrongValidatorAddress           = "dxvaloper16rr3cvdgj8jsywhx8lfteunn9uz0xg2czw6g00"
	testWrongMultisigParticipantAddress = "dx173lnn7jjuym5rwp23aufhnwshylrdemcswtc00"
	testWrongMultisigAddress            = "dx1kgnzuwwgzhecyk0dn62sxmp4wyuk0000000000"
	testWrongCoin                       = "tdel0"
	testWrongTxHash                     = "22EAE3E30713B1CC319FDDFCA0F47E94CC4BB94CC2052EBC1A255B53D27D0500"
	testWrongNFTTokenId                 = "n46sJWaSEgJ0Qyie3pelWci7jCI9mN1Wi0QFujHKSenbDAWuxFOjdCfhQmB02l00"
	testWrongProposalId                 = 100500
)

////////////////////////////////////////////////////////////////
// Decimal SDK example initializing
////////////////////////////////////////////////////////////////

/*
func init() {
	rand.Seed(time.Now().UnixNano())

	// Create Decimal API instance with default direct connection
	api = decapi.NewAPI(hostURL, &decapi.DirectConn{})

	// Create account from the mnemonic
	account, err = wallet.NewAccountFromMnemonicWords(testMnemonicWords, testMnemonicPassphrase)
	if err != nil {
		log.Println("Error while creating account")
		panic(err)
	}

	// Request chain Id
	if chainId, err := api.ChainId(); err == nil {
		account = account.WithChainId(chainId)
		log.Printf("Current chain Id: %s\n", chainId)
	} else {
		log.Println("Error while requesting chain id")
		panic(err)
	}

	// Request account number and sequence and update with received values
	if an, s, err := api.AccountNumberAndSequence(testSenderAddress); err == nil {
		account = account.WithAccountNumber(an).WithSequence(s)
		log.Printf("Account %s number: %d, sequence: %d\n", account.Address(), an, s)
	} else {
		log.Println("Error while account number and sequence")
		panic(err)
	}
}
*/

//resty logger implementation
type log2log struct{}

type Logger interface {
	Errorf(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Debugf(format string, v ...interface{})
}

func (l log2log) Errorf(format string, v ...interface{}) {
	log.Printf("L2LERR:"+format, v...)
}
func (l log2log) Warnf(format string, v ...interface{}) {
	log.Printf("L2LWRN:"+format, v...)
}
func (l log2log) Debugf(format string, v ...interface{}) {
	log.Printf("L2LDBG:"+format, v...)
}

////////////////////////////////////////////////////////////////
// Decimal SDK example running
////////////////////////////////////////////////////////////////

// possible api host+options combinations
var apiEndpoints = []struct {
	endpointId string
	hostURL    string
	dcOptions  *decapi.DirectConn
	netId      string
	baseCoin   string
}{
	{"testnet-gate", "https://testnet-gate.decimalchain.com/api", nil, "testnet", "tdel"},
	{"devnet-gate", "https://devnet-gate.decimalchain.com/api", nil, "devnet", "del"},
	{"devnet-local", "http://localhost", &decapi.DirectConn{}, "devnet", "del"}, // direct: RPC(port 26657)+REST(port 1317)
	{"testnet-local", "http://localhost",
		&decapi.DirectConn{PortRPC: ":26658", PortREST: ":1318"}, "testnet", "tdel"}, // direct: RPC(port 26658)+REST(port 1317)
}

//helper function
func formatAsJSON(obj interface{}) string {
	objStr, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s\n", objStr)
}

func main() {
	var api *decapi.API
	var endpointId = flag.String("id", "", "Predefined endpoint id for testing")
	var logfileName = flag.String("log", "example_testing.log", "path to logfile")
	var doLogRequests = flag.Bool("logreq", true, "write raw request/response to log")
	var checkCoins = flag.Bool("check-coins", false, "Check coins requests")
	var checkValidators = flag.Bool("check-validators", false, "Check validators requests")
	var checkProposals = flag.Bool("check-proposals", false, "Check proposals requests")
	var checkNFT = flag.Bool("check-nft", false, "Check NFT requests")
	var checkMultisig = flag.Bool("check-multisig", false, "Check Multisig requests")
	var checkStakes = flag.Bool("check-stakes", false, "Check stakes requests")
	var checkTransaction = flag.Bool("check-transaction", false, "Check transaction requests")
	var checkSend = flag.Bool("send", false, "Try to send transaction")
	var checkBurn = flag.Bool("burn", false, "Try to burn coin")
	var checkWallets bool = false
	flag.Parse()

	endpoint := apiEndpoints[0]
	for i := range apiEndpoints {
		if apiEndpoints[i].endpointId == *endpointId {
			endpoint = apiEndpoints[i]
		}
	}

	var i int
	rand.Seed(time.Now().UnixNano())

	log.SetFlags(log.Lshortfile | log.LstdFlags)
	//logfile, err := os.OpenFile(*logfileName, os.O_CREATE, 0644)
	logfile, err := os.Create(*logfileName)
	if err != nil {
		log.Fatalln("Cannot create log file")
	}
	defer logfile.Close()
	log.SetOutput(logfile)

	log.Printf("START: test endpoint %s with options %#v", endpoint.hostURL, endpoint.dcOptions)
	// 1
	// You can simply use
	// api := decapi.NewAPI(endpoint.hostURL, endpoint.dcOptions)
	if *doLogRequests {
		client1 := resty.New().SetDebug(true).SetLogger(log2log{}).SetTimeout(time.Second * 20)
		client2 := resty.New().SetDebug(true).SetLogger(log2log{}).SetTimeout(time.Second * 20)
		api = decapi.NewAPIWithClient(endpoint.hostURL, client1, client2, endpoint.dcOptions)
	} else {
		api = decapi.NewAPI(endpoint.hostURL, endpoint.dcOptions)
	}
	// 2
	chainId := exampleApiBlockchainInfo(api)
	////////////////////
	if checkWallets {
		// 3
		wallets := exampleApiCreateWallets()
		for i, _ = range wallets {
			wallets[i] = wallets[i].WithChainID(chainId)
		}
		// 3
		for i = 0; i < (len(wallets) - 1); i++ {
			fillWallet(api, wallets[i].Address(), endpoint.netId)
		}
		time.Sleep(time.Second * 10) //wait for transaction
		// 4
		for i = 0; i < len(wallets); i++ {
			log.Printf("Assign wallet #%d", i)
			// Request account number and sequence and update with received values
			if an, s, err := api.AccountNumberAndSequence(wallets[i].Address()); err == nil {
				wallets[i] = wallets[i].WithAccountNumber(an).WithSequence(s)
				log.Printf("Account %s number: %d, sequence: %d", wallets[i].Address(), an, s)
			} else {
				log.Printf("ERROR: while get account number and sequence: %s", err.Error())
			}
		}
		// 5
		for i = 0; i < len(wallets); i++ {
			log.Printf("Get info wallet #%d", i)
			adrInfo, err := api.Address(wallets[i].Address())
			if err != nil {
				log.Printf("ERROR: while get account info: %s", err.Error())
				continue
			}
			log.Printf("Account info: %s", formatAsJSON(adrInfo))
		}
	}
	////////////////////
	if *checkCoins {
		coinSymbols := []struct {
			symbol string
			valid  bool
		}{
			{endpoint.baseCoin, true},
			{"0del", false},
		}
		for _, symbol := range coinSymbols {
			log.Printf("Coint info: %s", symbol.symbol)
			coinInfo, err := api.Coin(symbol.symbol)
			if err != nil && symbol.valid {
				log.Printf("ERROR: while get coin info: %s", err.Error())
			}
			if err == nil && !symbol.valid {
				log.Printf("ERROR: expect error for wrong symbol")
			} else if err != nil && !symbol.valid {
				log.Printf("Wrong symbol message: %s", err.Error())
			}
			log.Printf("Coin info: %s", formatAsJSON(coinInfo))
		}
	}
	////////////////////
	if *checkValidators {
		log.Printf("Validators")
		validators, err := api.Validators()
		if err != nil {
			log.Printf("ERROR: while get validators: %s", err.Error())
		} else {
			log.Printf("Validators info:", formatAsJSON(validators))
			//
			log.Printf("get individual validator")
			validatorsChecks := []struct {
				address string
				valid   bool
			}{
				{validators[0].Address, true},
				{validators[0].Address + "0", false},
			}
			for _, vi := range validatorsChecks {
				log.Printf("try to get validator info address: %s, valid: %v", vi.address, vi.valid)
				val, err := api.Validator(vi.address)
				if err != nil && vi.valid {
					log.Printf("ERROR: while get validator: %s", err.Error())
				}
				if err == nil && !vi.valid {
					log.Printf("ERROR: you must get error")
				}
				log.Printf("Validator info: %s", formatAsJSON(val))
			}
		}
		//
		log.Printf("Candidates")
		cands, err := api.Candidates()
		if err != nil {
			log.Printf("ERROR: while get candidates: %s", err.Error())
		} else {
			log.Printf("Candidates info: %s", formatAsJSON(cands))
		}
	}
	////////////////////
	if *checkProposals {
		log.Printf("Proposals")
		props, err := api.Proposals()
		if err != nil {
			log.Printf("ERROR: while get proposals: %s", err.Error())
		}
		log.Printf("Proposals info: %s", formatAsJSON(props))
		//
		if len(props) > 0 {
			log.Printf("Try to get existing proposal")
			prop, err := api.Proposal(props[0].ProposalID)
			if err != nil {
				log.Printf("ERROR: while get proposal: %s", err.Error())
			}
			log.Printf("Proposal info: %s", formatAsJSON(prop))
		}
		//
		log.Printf("Try to get unexisting proposal")
		prop, err := api.Proposal(100500)
		if err == nil {
			log.Printf("ERROR: you must get error")
		}
		log.Printf("Proposal info: %s", formatAsJSON(prop))
	}
	////////////////////
	if *checkNFT {
		log.Printf("NFTs")
		log.Printf("Get NFT list")
		nfts, err := api.NFTList()
		if err != nil {
			log.Printf("ERROR: while get NFT list: %s", err.Error())
		}
		log.Printf("NFT list: %s", formatAsJSON(nfts))
		for i, nft := range nfts {
			if i > 3 {
				break
			}
			data, err := api.NFT(nft.Id)
			if err != nil {
				log.Printf("ERROR: while get NFT by id(%s): %s", nft.Id, err.Error())
			}
			log.Printf("NFT: %s", formatAsJSON(data))
		}
		for i, nft := range nfts {
			if i > 3 {
				break
			}
			data, err := api.NFTByAddress(nft.Creator)
			if err != nil {
				log.Printf("ERROR: while get NFT by address(%s): %s", nft.Creator, err.Error())
			}
			log.Printf("NFTByAddress: %s", formatAsJSON(data))
		}
	}
	////////////////////
	if *checkMultisig {

	}
	////////////////////
	if *checkStakes {
		log.Printf("Stakes")
		for _, adr := range []string{
			"dx18ag7adcd0qxrlfxw3f9v79lfvgh99xe50s63a3",
			"dx1w98j4vk6dkpyndjnv5dn2eemesq6a2c2j9depy",
			"dx19c7rudu8fs9kvhxyxuxer03058t78cxzvacgd4",
			"dx1wq40a4tzk226kymfzqfr0s96cjeka66j0xmlcr",
			"dx16mjgdzv8aq2jwdtrdgjh06233rdl2u52dk4kjz",
		} {
			stakes, err := api.Stakes(adr)
			if err != nil {
				log.Printf("ERROR: while get stakes: %s", err.Error())
			}
			log.Printf("Stakes info: %s", formatAsJSON(stakes))
		}
	}
	////////////////////
	if *checkSend {
		testSend(api, endpoint.baseCoin)
		//testInvalidSendCoin(api)
		//testInvalidSendSignature(api)
		//testGovProposal(api)
	}
	////////////////////
	if *checkTransaction {
		txs := []string{}
		last_block, err := api.GetHeight()
		if err != nil {
			log.Printf("ERROR: while get last block: %s", err.Error())
		}
		// try find at least 2 transactions in last 500 blocks
		for block := last_block; (len(txs) < 2) && (block > last_block-100); block-- {
			tmp, err := api.TransactionsByBlock(block)
			if err != nil {
				log.Printf("ERROR: while get txs: %s", err.Error())
			}
			txs = append(txs, tmp...)
		}
		// get tx info
		for _, hash := range txs {
			tx, err := api.Transaction(hash)
			if err != nil {
				log.Printf("ERROR: while get tx: hash=%s, error=%s", hash, err.Error())
			}
			log.Printf("Tx result: %s", formatAsJSON(tx))
		}
	}
	////////////////////
	if *checkBurn {
		testBurnCoin(api)
		//testInvalidSendCoin(api)
		//testInvalidSendSignature(api)
		//testGovProposal(api)
	}
	////////////////////
	log.Printf("END test endpoint")
	log.Println("--------------------------")

}

func exampleApiBlockchainInfo(api *decapi.API) string {
	log.Printf("Request ChainId")
	chainId, err := api.ChainID()
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
	} else {
		log.Printf("ChainId = %s", chainId)
	}
	return chainId
}

// create 3 wallets (accounts) for further testing
func exampleApiCreateWallets() []*wallet.Account {
	var res []*wallet.Account
	log.Printf("Create wallets")
	for i := 0; i < 3; i++ {
		log.Printf("Create wallet #%d", i)
		mnemonic, err := wallet.NewMnemonic(256, "")
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
		} else {
			log.Printf("Wallet #%d mnemonic %#v", i, mnemonic.Words())
		}
		w, err := wallet.NewAccountFromMnemonic(mnemonic)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
		} else {
			log.Printf("Wallet #%d address %s", i, w.Address())
		}
		res = append(res, w)
	}
	return res
}

//devnet/testnet 15k del request
func fillWallet(api *decapi.API, address string, network string) {
	log.Printf("Fill wallet %s on net %s", address, network)
	body := bytes.NewBufferString(fmt.Sprintf("{\"address\":\"%s\",\"network\":\"%s\"}", address, network))
	resp, err := http.Post("https://faucet.decimalchain.com/api/get", "application/json; charset=utf-8", body)
	if err != nil {
		log.Printf("ERROR: Fill requests failed: %s", err.Error())
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("ERROR: Fill requests failed: %s", resp.Status)
		return
	}
}

// TODO: full set with error handling
func bindAcc(api *decapi.API, acc *wallet.Account) {
	// check and bind wallets numbers and sequences; if zero - account without balance and transcations
	if accNumber, seq, err := api.AccountNumberAndSequence(acc.Address()); err == nil {
		acc = acc.WithAccountNumber(accNumber).WithSequence(seq)
		log.Printf("Account %s number: %d, sequence: %d", acc.Address(), accNumber, seq)
	} else {
		log.Printf("ERROR: get AccountNumberAndSequence error %s", err.Error())
	}
}
