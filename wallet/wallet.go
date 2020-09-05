package wallet

import (
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"math/big"

	"github.com/everFinance/goar/client"
	"github.com/everFinance/goar/types"
	"github.com/everFinance/goar/utils"
	"github.com/everFinance/gojwk"
)

type Wallet struct {
	Client  *client.Client
	PubKey  *rsa.PublicKey
	PrvKey  *rsa.PrivateKey
	Address string
}

func NewFromPath(path string) (*Wallet, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return New(b)
}

func New(b []byte) (w *Wallet, err error) {
	key, err := gojwk.Unmarshal(b)
	if err != nil {
		return
	}

	pubKey, err := key.DecodePublicKey()
	if err != nil {
		return
	}
	pub, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		err = fmt.Errorf("pubKey type error")
		return
	}
	prvKey, err := key.DecodePrivateKey()
	if err != nil {
		return
	}
	prv, ok := prvKey.(*rsa.PrivateKey)
	if !ok {
		err = fmt.Errorf("prvKey type error")
		return
	}

	addr := sha256.Sum256(pub.N.Bytes())
	w = &Wallet{
		Client:  client.New("https://arweave.net"),
		PubKey:  pub,
		PrvKey:  prv,
		Address: utils.Base64Encode(addr[:]),
	}

	return
}
func (w *Wallet) SendAR(amount *big.Float, target string, tags []types.Tag) (id, status string, err error) {
	return w.SendWinston(utils.ARToWinston(amount), target, tags)
}

func (w *Wallet) SendWinston(amount *big.Int, target string, tags []types.Tag) (id, status string, err error) {
	reward, err := w.Client.GetTransactionPrice(nil, &target)
	if err != nil {
		return
	}

	tx := &types.Transaction{
		Format:   2,
		Target:   target,
		Quantity: amount.String(),
		Tags:     utils.TagsEncode(tags),
		Data:     "",
		DataSize: "0",
		Reward:   fmt.Sprintf("%d", reward),
	}

	return w.SendTransaction(tx)
}

func (w *Wallet) SendData(data []byte, tags []types.Tag) (id, status string, err error) {
	reward, err := w.Client.GetTransactionPrice(data, nil)
	if err != nil {
		return
	}

	tx := &types.Transaction{
		Format:   2,
		Target:   "",
		Quantity: "0",
		Tags:     utils.TagsEncode(tags),
		Data:     utils.Base64Encode(data),
		DataSize: fmt.Sprintf("%d", len(string(data))),
		Reward:   fmt.Sprintf("%d", reward),
	}

	return w.SendTransaction(tx)
}

// SendTransaction: if send success, should return pending
func (w *Wallet) SendTransaction(tx *types.Transaction) (id, status string, err error) {
	anchor, err := w.Client.GetTransactionAnchor()
	if err != nil {
		return
	}
	tx.LastTx = anchor

	if err = utils.SignTransaction(tx, w.PubKey, w.PrvKey); err != nil {
		return
	}

	id = tx.ID
	status, err = w.Client.SubmitTransaction(tx)
	if err != nil || status != "OK" {
		return
	}

	// no data & no target tx sent return OK but tx not really accept from ar node.
	// TODO: catch ar node not found tx
	_, status, err = w.Client.GetTransactionByID(id)

	return
}