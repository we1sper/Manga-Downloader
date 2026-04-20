package mhg

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/we1sper/Manga-Downloader/mhg/script"
)

type Decoder struct {
	decryptExpression    string
	decompressExpression string
}

func NewDecoder() *Decoder {
	return &Decoder{
		decryptExpression:    "var result = decrypt(%s)",
		decompressExpression: "var result = LZString.decompressFromBase64('%s')",
	}
}

func (d *Decoder) Decrypt(cipherText string) (string, error) {
	return d.kernel(d.decryptExpression, cipherText)
}

func (d *Decoder) Decompress(compressedText string) (string, error) {
	return d.kernel(d.decompressExpression, compressedText)
}

func (d *Decoder) kernel(expression, text string) (string, error) {
	scriptExecuted := script.Get() + fmt.Sprintf(expression, text)

	vm := goja.New()

	_, err := vm.RunString(scriptExecuted)
	if err != nil {
		return "", err
	}

	return vm.Get("result").String(), nil
}
