package testgen

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

var forkActive = map[string]func(*params.ChainConfig, *types.Header) bool{
	"Byzantium":      func(c *params.ChainConfig, h *types.Header) bool { return c.IsByzantium(h.Number) },
	"Constantinople": func(c *params.ChainConfig, h *types.Header) bool { return c.IsConstantinople(h.Number) },
	"Petersburg":     func(c *params.ChainConfig, h *types.Header) bool { return c.IsPetersburg(h.Number) },
	"Istanbul":       func(c *params.ChainConfig, h *types.Header) bool { return c.IsIstanbul(h.Number) },
	"Berlin":         func(c *params.ChainConfig, h *types.Header) bool { return c.IsBerlin(h.Number) },
	"London":         func(c *params.ChainConfig, h *types.Header) bool { return c.IsLondon(h.Number) },
	"Paris":          func(_ *params.ChainConfig, h *types.Header) bool { return h.Difficulty.Sign() == 0 },
	"Shanghai":       func(c *params.ChainConfig, h *types.Header) bool { return c.IsShanghai(h.Number, h.Time) },
	"Cancun":         func(c *params.ChainConfig, h *types.Header) bool { return c.IsCancun(h.Number, h.Time) },
	"Prague":         func(c *params.ChainConfig, h *types.Header) bool { return c.IsPrague(h.Number, h.Time) },
	"Osaka":          func(c *params.ChainConfig, h *types.Header) bool { return c.IsOsaka(h.Number, h.Time) },
	"Amsterdam":      func(c *params.ChainConfig, h *types.Header) bool { return c.IsAmsterdam(h.Number, h.Time) },
}

func firstForkBlock(c *Chain, fork string) uint64 {
	for _, b := range c.blocks {
		if forkActive[fork](c.Config(), b.Header()) {
			return b.NumberU64()
		}
	}
	panic("fork " + fork + " is not active in the chain")
}
