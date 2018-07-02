package goterm

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/hiro4bbh/go-assert"
)

func TestIsTerminal(t *testing.T) {
	var buf bytes.Buffer
	goassert.New(t, false).Equal(IsTerminal(&buf))
	f := goassert.New(t).SucceedNew(ioutil.TempFile("", "is_terminal")).(*os.File)
	goassert.New(t, false).Equal(IsTerminal(f))
	// We cannot test the following case?
	//goassert.New(t, true).Equal(IsTerminal(os.Stdout))
}

func TestRuneWidth(t *testing.T) {
	goassert.New(t, 0).Equal(RuneWidth('\u0000'))
	goassert.New(t, 1).Equal(RuneWidth('A'))
	goassert.New(t, 1).Equal(RuneWidth('a'))
	goassert.New(t, 1).Equal(RuneWidth('0'))
	goassert.New(t, 1).Equal(RuneWidth('!'))
	goassert.New(t, 1).Equal(RuneWidth('?'))
	goassert.New(t, 1).Equal(RuneWidth('-'))
	goassert.New(t, 2).Equal(RuneWidth('　'))
	goassert.New(t, 2).Equal(RuneWidth('、'))
	goassert.New(t, 2).Equal(RuneWidth('。'))
	goassert.New(t, 2).Equal(RuneWidth('Ａ'))
	goassert.New(t, 2).Equal(RuneWidth('０'))
	goassert.New(t, 2).Equal(RuneWidth('！'))
	goassert.New(t, 2).Equal(RuneWidth('？'))
	goassert.New(t, 2).Equal(RuneWidth('ー'))
	goassert.New(t, 2).Equal(RuneWidth('あ'))
	goassert.New(t, 2).Equal(RuneWidth('ば'))
	goassert.New(t, 2).Equal(RuneWidth('ぴ'))
	goassert.New(t, 2).Equal(RuneWidth('ア'))
	goassert.New(t, 1).Equal(RuneWidth('ｱ'))
}
