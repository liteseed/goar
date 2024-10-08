package wallet

import (
	"errors"
	"os"

	"github.com/liteseed/goar/client"
	"github.com/liteseed/goar/signer"
	"github.com/liteseed/goar/tag"
	"github.com/liteseed/goar/transaction"
	"github.com/liteseed/goar/transaction/bundle"
	"github.com/liteseed/goar/transaction/data_item"
	"github.com/liteseed/goar/uploader"
)

type Wallet struct {
	Client *client.Client
	Signer *signer.Signer
}

// New create a [Wallet] with a new [Signer] and [Client] using the gateway url
func New(gateway string) (w *Wallet, err error) {
	s, err := signer.New()
	if err != nil {
		return nil, err
	}
	return &Wallet{
		Client: client.New(gateway),
		Signer: s,
	}, nil
}

// FromPath create a [Wallet] with a [Signer] from the path and [Client] using the gateway url
func FromPath(path string, gateway string) (*Wallet, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return FromJWK(b, gateway)
}

// FromJWK create a [Wallet] with a [signer.Signer] from the JWK and [client.Client] using the gateway url
func FromJWK(jwk []byte, gateway string) (*Wallet, error) {
	s, err := signer.FromJWK(jwk)
	if err != nil {
		return nil, err
	}
	return &Wallet{
		Client: client.New(gateway),
		Signer: s,
	}, nil
}

// CreateTransaction create a [transaction.Transaction]
func (w *Wallet) CreateTransaction(data []byte, target string, quantity string, tags *[]tag.Tag) *transaction.Transaction {
	return transaction.New(data, target, quantity, tags)
}

// SignTransaction sign a [transaction.Transaction]
func (w *Wallet) SignTransaction(tx *transaction.Transaction) (*transaction.Transaction, error) {
	tx.Owner = w.Signer.Owner()

	anchor, err := w.Client.GetTransactionAnchor()
	if err != nil {
		return nil, err
	}
	tx.LastTx = anchor

	reward, err := w.Client.GetTransactionPrice(len(tx.Data), "")
	if err != nil {
		return nil, err
	}
	tx.Reward = reward

	if err = tx.Sign(w.Signer); err != nil {
		return nil, err
	}
	return tx, nil
}

// SendTransaction send a [transaction.Transaction] to an Arweave Gateway
func (w *Wallet) SendTransaction(tx *transaction.Transaction) error {
	if tx.ID == "" || tx.Signature == "" {
		return errors.New("transaction not signed")
	}
	tu, err := uploader.New(w.Client, tx)
	if err != nil {
		return err
	}
	if err = tu.PostTransaction(); err != nil {
		return err
	}
	return nil
}

// CreateDataItem create a [data_item.DataItem] which is a special format for uploading data.
func (w *Wallet) CreateDataItem(data []byte, target string, anchor string, tags *[]tag.Tag) *data_item.DataItem {
	return data_item.New(data, target, anchor, tags)
}

func (w *Wallet) SignDataItem(di *data_item.DataItem) (*data_item.DataItem, error) {
	if err := di.Sign(w.Signer); err != nil {
		return nil, err
	}
	return di, nil
}

// CreateBundle create a [bundle.Bundle] which is a sequence of [data_item.DataItem] defined in ANS-104
func (w *Wallet) CreateBundle(dataItems *[]data_item.DataItem) (*bundle.Bundle, error) {
	return bundle.New(dataItems)
}
