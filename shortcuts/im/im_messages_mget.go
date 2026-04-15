// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package im

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/larksuite/cli/internal/output"
	"github.com/larksuite/cli/shortcuts/common"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

const maxMGetMessageIDs = 50

var ImMessagesMGet = common.Shortcut{
	Service:     "im",
	Command:     "+messages-mget",
	Description: "Batch get messages by IDs; user/bot; fetches up to 50 om_ message IDs, formats sender names, expands thread replies",
	Risk:        "read",
	Scopes:      []string{"im:message:readonly"},
	UserScopes:  []string{"im:message.group_msg:get_as_user", "im:message.p2p_msg:get_as_user", "contact:user.basic_profile:readonly"},
	BotScopes:   []string{"im:message.group_msg", "im:message.p2p_msg:readonly", "contact:user.base:readonly"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "message-ids", Desc: "message IDs, comma-separated (om_xxx,om_yyy)", Required: true},
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		ids := common.SplitCSV(runtime.Str("message-ids"))
		return common.NewDryRunAPI().GET(buildMGetURL(ids))
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		ids := common.SplitCSV(runtime.Str("message-ids"))
		if len(ids) == 0 {
			return output.ErrValidation("--message-ids is required (comma-separated om_xxx)")
		}
		if len(ids) > maxMGetMessageIDs {
			return output.ErrValidation("--message-ids supports at most %d IDs per request (got %d)", maxMGetMessageIDs, len(ids))
		}
		for _, id := range ids {
			if _, err := validateMessageID(id); err != nil {
				return err
			}
		}
		return nil
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		ids := common.SplitCSV(runtime.Str("message-ids"))
		mgetURL := buildMGetURL(ids)

		apiResp, err := runtime.DoAPI(&larkcore.ApiReq{
			HttpMethod: http.MethodGet,
			ApiPath:    mgetURL,
		})
		if err != nil {
			return err
		}
		if apiResp.StatusCode >= 400 {
			return output.ErrAPI(apiResp.StatusCode, string(apiResp.RawBody), nil)
		}

		var raw map[string]any
		if err := json.Unmarshal(apiResp.RawBody, &raw); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
		runtime.Out(raw, nil)
		return nil
	},
}
