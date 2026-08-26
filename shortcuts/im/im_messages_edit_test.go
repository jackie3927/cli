// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package im

import (
	"context"
	"strings"
	"testing"

	"github.com/larksuite/cli/shortcuts/common"
	"github.com/spf13/cobra"
)

// newEditTestRuntimeContext builds a RuntimeContext wired with the +messages-edit
// flag surface, including the repeatable --set-attachments string-slice flag.
func newEditTestRuntimeContext(stringFlags map[string]string, sliceFlags map[string][]string) *common.RuntimeContext {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("message-id", "", "")
	cmd.Flags().String("msg-type", "text", "")
	cmd.Flags().String("content", "", "")
	cmd.Flags().String("text", "", "")
	cmd.Flags().String("markdown", "", "")
	cmd.Flags().StringSlice("set-attachments", nil, "")
	cmd.Flags().Bool("clear-attachments", false, "")
	for name, val := range stringFlags {
		_ = cmd.Flags().Set(name, val)
	}
	for name, vals := range sliceFlags {
		_ = cmd.Flags().Set(name, strings.Join(vals, ","))
	}
	return &common.RuntimeContext{Cmd: cmd}
}

func TestValidateEditContentFlags(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		markdown    string
		content     string
		attachments []attachmentItem
		wantErr     string
	}{
		{name: "text ok", text: "hi", wantErr: ""},
		{name: "markdown ok", markdown: "# hi", wantErr: ""},
		{name: "content ok", content: `{"text":"hi"}`, wantErr: ""},
		{name: "attachment only ok", attachments: []attachmentItem{{Key: "file_1"}}, wantErr: ""},
		{name: "text and markdown conflict", text: "hi", markdown: "# hi", wantErr: "cannot be specified together"},
		{name: "nothing specified", wantErr: "specify --content"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateEditContentFlags(tt.text, tt.markdown, tt.content, tt.attachments)
			if tt.wantErr == "" {
				if got != "" {
					t.Fatalf("validateEditContentFlags() = %q, want empty", got)
				}
				return
			}
			if !strings.Contains(got, tt.wantErr) {
				t.Fatalf("validateEditContentFlags() = %q, want substring %q", got, tt.wantErr)
			}
		})
	}
}

func TestImMessagesEditValidate(t *testing.T) {
	t.Run("markdown with attachment", func(t *testing.T) {
		rt := newEditTestRuntimeContext(map[string]string{
			"message-id": "om_123",
			"markdown":   "# hi",
		}, map[string][]string{"set-attachments": {"file_1"}})
		if err := ImMessagesEdit.Validate(context.Background(), rt); err != nil {
			t.Fatalf("ImMessagesEdit.Validate() unexpected error = %v", err)
		}
	})

	t.Run("content post with attachment", func(t *testing.T) {
		rt := newEditTestRuntimeContext(map[string]string{
			"message-id": "om_123",
			"msg-type":   "post",
			"content":    `{"zh_cn":{"content":[[{"tag":"text","text":"hi"}]]}}`,
		}, map[string][]string{"set-attachments": {"file_1"}})
		if err := ImMessagesEdit.Validate(context.Background(), rt); err != nil {
			t.Fatalf("ImMessagesEdit.Validate() unexpected error = %v", err)
		}
	})

	t.Run("text with attachment rejected (post only)", func(t *testing.T) {
		rt := newEditTestRuntimeContext(map[string]string{
			"message-id": "om_123",
			"text":       "hi",
		}, map[string][]string{"set-attachments": {"file_1"}})
		err := ImMessagesEdit.Validate(context.Background(), rt)
		if err == nil || !strings.Contains(err.Error(), "post message") {
			t.Fatalf("ImMessagesEdit.Validate() error = %v, want post-only hint", err)
		}
	})

	t.Run("missing message-id", func(t *testing.T) {
		rt := newEditTestRuntimeContext(map[string]string{"text": "hi"}, nil)
		err := ImMessagesEdit.Validate(context.Background(), rt)
		if err == nil || !strings.Contains(err.Error(), "--message-id") {
			t.Fatalf("ImMessagesEdit.Validate() error = %v, want --message-id error", err)
		}
	})

	t.Run("missing content", func(t *testing.T) {
		rt := newEditTestRuntimeContext(map[string]string{"message-id": "om_123"}, nil)
		err := ImMessagesEdit.Validate(context.Background(), rt)
		if err == nil || !strings.Contains(err.Error(), "specify --content") {
			t.Fatalf("ImMessagesEdit.Validate() error = %v, want content-required error", err)
		}
	})

	t.Run("clear-attachments with markdown", func(t *testing.T) {
		rt := newEditTestRuntimeContext(map[string]string{
			"message-id":        "om_123",
			"markdown":          "# hi",
			"clear-attachments": "true",
		}, nil)
		if err := ImMessagesEdit.Validate(context.Background(), rt); err != nil {
			t.Fatalf("ImMessagesEdit.Validate() unexpected error = %v", err)
		}
	})

	t.Run("clear-attachments with attachment rejected", func(t *testing.T) {
		rt := newEditTestRuntimeContext(map[string]string{
			"message-id":        "om_123",
			"markdown":          "# hi",
			"clear-attachments": "true",
		}, map[string][]string{"set-attachments": {"file_1"}})
		err := ImMessagesEdit.Validate(context.Background(), rt)
		if err == nil || !strings.Contains(err.Error(), "cannot be used with --set-attachments") {
			t.Fatalf("ImMessagesEdit.Validate() error = %v, want clear+set-attachments conflict", err)
		}
	})

	t.Run("clear-attachments without post type rejected", func(t *testing.T) {
		rt := newEditTestRuntimeContext(map[string]string{
			"message-id":        "om_123",
			"text":              "hi",
			"clear-attachments": "true",
		}, nil)
		err := ImMessagesEdit.Validate(context.Background(), rt)
		if err == nil || !strings.Contains(err.Error(), "post message") {
			t.Fatalf("ImMessagesEdit.Validate() error = %v, want post-only hint", err)
		}
	})
}
