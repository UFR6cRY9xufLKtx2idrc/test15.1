/*
 * Copyright (C) 2020-2022, IrineSistiana
 *
 * This file is part of mosdns.
 *
 * mosdns is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * mosdns is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package client_ip

import (
	"github.com/UFR6cRY9xufLKtx2idrc/mosdns/main/pkg/matcher/netlist"
	"github.com/UFR6cRY9xufLKtx2idrc/mosdns/main/pkg/query_context"
	"github.com/UFR6cRY9xufLKtx2idrc/mosdns/main/plugin/executable/sequence"
	"github.com/UFR6cRY9xufLKtx2idrc/mosdns/main/plugin/matcher/base_ip"
)

const PluginType = "client_ip"

func init() {
	sequence.MustRegMatchQuickSetup(PluginType, QuickSetup)
}

type Args = base_ip.Args

func QuickSetup(bq sequence.BQ, s string) (sequence.Matcher, error) {
	return base_ip.NewMatcher(bq, base_ip.ParseQuickSetupArgs(s), matchClientAddr)
}

func matchClientAddr(qCtx *query_context.Context, m netlist.Matcher) (bool, error) {
	addr, _ := query_context.GetClientAddr(qCtx)
	if !addr.IsValid() {
		return false, nil
	}
	return m.Match(*addr), nil
}
