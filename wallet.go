package goar

import (
	"context"
	"fmt"
	"os"

	"github.com/liteseed/goar/client"
	"github.com/liteseed/goar/signer"
	"github.com/liteseed/goar/tx"
)

type Wallet struct {
	Client *client.Client
	Signer *signer.Signer
}

func New(b []byte, url string) (w *Wallet, err error) {
	signer, err := signer.New(b)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		Client: client.New(url),
		Signer: signer,
	}, nil
}

func FromPath(path string, node string) (*Wallet, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return New(b, node)
}

func (w *Wallet) Owner() string {
	return w.Signer.Owner()
}

func (w *Wallet) SendData(data []byte, tags []tx.Tag) (tx.Transaction, error) {
	return w.SendDataSpeedUp(data, tags, 0)
}

// SendDataSpeedUp set speedFactor for speed up
// eg: speedFactor = 10, reward = 1.1 * reward
func (w *Wallet) SendDataSpeedUp(data []byte, tags []tx.Tag, speedFactor int64) (tx.Transaction, error) {
	reward, err := w.Client.GetTransactionPrice(len(data), nil)
	if err != nil {
		return tx.Transaction{}, err
	}

	tx := &tx.Transaction{
		Format:   2,
		Target:   "",
		Quantity: "0",
		Tags:     tags,
		Data:     data,
		DataSize: fmt.Sprintf("%d", len(data)),
		Reward:   fmt.Sprintf("%d", reward*(100+speedFactor)/100),
	}

	return w.SendTransaction(tx)
}

func (w *Wallet) SendDataConcurrentSpeedUp(ctx context.Context, concurrentNum int, data interface{}, tags []tx.Tag, speedFactor int64) (tx.Transaction, error) {
	var reward int64
	var dataLen int
	isByteArr := true
	if _, isByteArr = data.([]byte); isByteArr {
		dataLen = len(data.([]byte))
	} else {
		fileInfo, err := data.(*os.File).Stat()
		if err != nil {
			return tx.Transaction{}, err
		}
		dataLen = int(fileInfo.Size())
	}
	reward, err := w.Client.GetTransactionPrice(dataLen, nil)
	if err != nil {
		return tx.Transaction{}, err
	}

	tx := &tx.Transaction{
		Format:   2,
		Target:   "",
		Quantity: "0",
		Tags:     tags,
		DataSize: fmt.Sprintf("%d", dataLen),
		Reward:   fmt.Sprintf("%d", reward*(100+speedFactor)/100),
	}

	tx.Data = data.([]byte)

	return w.SendTransactionConcurrent(ctx, concurrentNum, tx)
}

// SendTransaction: if send success, should return pending
func (w *Wallet) SendTransaction(transaction *tx.Transaction) (tx.Transaction, error) {
	uploader, err := w.getUploader(transaction)
	if err != nil {
		return tx.Transaction{}, err
	}
	err = uploader.Once()
	return *transaction, err
}

func (w *Wallet) SendTransactionConcurrent(ctx context.Context, concurrentNum int, transaction *tx.Transaction) (tx.Transaction, error) {
	uploader, err := w.getUploader(transaction)
	if err != nil {
		return tx.Transaction{}, err
	}
	err = uploader.ConcurrentOnce(ctx, concurrentNum)
	return *transaction, err
}

func (w *Wallet) getUploader(transaction *tx.Transaction) (*client.TransactionUploader, error) {
	anchor, err := w.Client.GetTransactionAnchor()
	if err != nil {
		return nil, err
	}
	transaction.LastTx = anchor
	transaction.Owner = w.Owner()
	if err = w.Signer.SignTx(transaction); err != nil {
		return nil, err
	}
	return client.CreateUploader(w.Client, transaction, nil)
}
