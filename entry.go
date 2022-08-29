package fctrl

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/tigbox/fctrl/internal/lib"
)

type Entry struct {
	Input map[string]interface{}
	Mode  int   // Mode -2/-1/0/1 分别表示 仅读取不返回明细/仅读取返回明细/写入返回明细/写入不返回明细 默认是0
	Ts    int64 // Ts 其实就是timestamp
}

type EntryOption func(entry *Entry)

func OptionTs(ts int64) EntryOption {
	return func(entry *Entry) {
		// 对ts进行赋值
		entry.Ts = ts
	}
}

func (entry *Entry) OptionTs(ts int64) EntryOption {
	return func(entry *Entry) {
		// 对ts进行赋值
		entry.Ts = ts
	}
}

// OptionMode 1/2 分别表示 返回的时候带上明细数据/返回的时候不带明细数据
func OptionMode(mode int) EntryOption {
	return func(entry *Entry) {
		entry.Mode = mode
	}
}

// NewEntry ...
func NewEntry(input map[string]interface{}, options ...EntryOption) *Entry {
	entry := &Entry{
		Input: input,
		Mode:  default_mode,
	}
	for _, optionFunc := range options {
		optionFunc(entry)
	}
	return entry
}

func (entry *Entry) checkAndFormat(ctx context.Context) error {
	if entry == nil {
		return errors.New("entry is nil")
	}
	if !lib.ContainIntSlice(SupportModes, entry.Mode) {
		return fmt.Errorf("not support mode %v", entry.Mode)
	}
	if entry.Ts == 0 {
		OptionTs(TimeStamp())(entry)
	}
	return nil

}
