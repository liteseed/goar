package transaction

import (
	"errors"

	"github.com/liteseed/goar/crypto"
	"github.com/liteseed/goar/signer"
	"github.com/liteseed/goar/tag"
)

func New(data []byte, target string, quantity string, tags *[]tag.Tag) *Transaction {
	if tags == nil {
		tags = &[]tag.Tag{}
	}
	if quantity == "" {
		quantity = "0"
	}
	if data == nil {
		data = []byte("")
	}
	return &Transaction{
		Format:   2,
		Data:     crypto.Base64URLEncode(data),
		Target:   target,
		Quantity: quantity,
		Tags:     tag.ConvertToBase64(tags),
		DataSize: "0",
	}
}

func (tx *Transaction) Sign(s *signer.Signer) error {
	payload, err := tx.getSignatureData()
	if err != nil {
		return err
	}
	rawSignature, err := crypto.Sign(payload, s.PrivateKey)
	if err != nil {
		return err
	}
	tx.ID = crypto.Base64URLEncode(crypto.SHA256(rawSignature))
	tx.Signature = crypto.Base64URLEncode(rawSignature)
	return nil
}

func (tx *Transaction) Verify() error {
	signatureData, err := tx.getSignatureData()
	if err != nil {
		return err
	}
	rawSignature, err := crypto.Base64URLDecode(tx.Signature)
	if err != nil {
		return err
	}
	publicKey, err := crypto.GetPublicKeyFromOwner(tx.Owner)
	if err != nil {
		return err
	}
	return crypto.Verify(signatureData, rawSignature, publicKey)
}

func (tx *Transaction) getSignatureData() ([]byte, error) {
	if tx.Format != 2 {
		return nil, errors.New("only type 2 transaction supported")
	}
	rawOwner, err := crypto.Base64URLDecode(tx.Owner)
	if err != nil {
		return nil, err
	}
	rawTarget, err := crypto.Base64URLDecode(tx.Target)
	if err != nil {
		return nil, err
	}

	rawTags, err := tag.Decode(tx.Tags)
	if err != nil {
		return nil, err
	}

	rawLastTx, err := crypto.Base64URLDecode(tx.LastTx)
	if err != nil {
		return nil, err
	}

	data, err := crypto.Base64URLDecode(tx.Data)
	if err != nil {
		return nil, err
	}

	err = tx.PrepareChunks(data)
	if err != nil {
		return nil, err
	}

	rawDataRoot, err := crypto.Base64URLDecode(tx.DataRoot)
	if err != nil {
		return nil, err
	}

	chunks := []any{
		[]byte("2"),
		rawOwner,
		rawTarget,
		[]byte(tx.Quantity),
		[]byte(tx.Reward),
		rawLastTx,
		rawTags,
		[]byte(tx.DataSize),
		rawDataRoot,
	}

	deepHash := crypto.DeepHash(chunks)
	signatureData := deepHash[:]
	return signatureData, nil
}
