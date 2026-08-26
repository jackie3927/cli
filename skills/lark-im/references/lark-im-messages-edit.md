# im +messages-edit

> **Prerequisite:** Read [`../lark-shared/SKILL.md`](../../lark-shared/SKILL.md) first to understand authentication, global parameters, and safety rules.

Edit an already-sent message's content. Supports both user identity (`--as user`) and bot identity (`--as bot`). Only messages the calling identity sent can be edited.

This skill maps to the shortcut: `lark-cli im +messages-edit` (internally calls `PUT /open-apis/im/v1/messages/:message_id`).

## Safety Constraints

Editing rewrites a message visible to other people. Before calling it, you **must** confirm with the user:

1. Which message to edit (its `message_id`)
2. The new content
3. Which identity to use (user or bot)

**Do not** edit a message without explicit user approval. The edited message must have been sent by the calling identity; editing someone else's message fails.

## Choose The Right Content Flag

| Need | Recommended flag | Why |
|------|------|------|
| Edit to headings, lists, links, summaries, or Markdown-looking content | `--markdown` | Best default for lightweight formatting; converted to Feishu `post` JSON |
| Edit to exact plain text | `--text` | Preserves literal text; no Markdown conversion |
| Precisely control the new payload | `--content` | You provide the exact JSON for `text` / `post` |
| Attach files/folders to the edited message's attachment zone | `--set-attachments` | Repeatable, as bare `file_key` (`file_xxx`); sets the post content's `files` array. Requires a post message (`--markdown` or `--msg-type post`). Name/metadata are filled by the server, not the client |
| Clear the edited message's attachment zone | `--clear-attachments` | Sets `files:[]` on the post content. Requires a post message; mutually exclusive with `--set-attachments` |

## Editing the Attachment Zone

`post` messages can carry an attachment zone — a top-level `files` array that renders files/folders under the rich-text body. To edit a message so it attaches (or re-attaches) files:

```bash
lark-cli im +messages-edit --message-id om_xxx --markdown "Updated content" --set-attachments file_xxx --set-attachments file_yyy
```

- `--set-attachments` accepts a bare file/folder key (`file_xxx`), and may be repeated.
- The server fills name/size/mime/is_folder from file service metadata; the client does not (and cannot) override the display name.
- When `--set-attachments` is present the effective `msg_type` is forced to `post`. Pair it with `--markdown` (or `--content` with post JSON plus `--msg-type post`); `--text` cannot carry an attachment zone.
- The edited content replaces the whole message content, so include every file you want to keep in the new attachment zone.

To **clear** the attachment zone entirely, pass `--clear-attachments` instead of `--set-attachments`:

```bash
lark-cli im +messages-edit --message-id om_xxx --markdown "Updated content" --clear-attachments
```

- `--clear-attachments` sets the post content's `files` array to `[]`, telling the server to remove all file/folder attachments.
- It cannot be used together with `--set-attachments`.
- Like `--set-attachments`, it forces the effective `msg_type` to `post`, so pair it with `--markdown` or `--msg-type post --content <post-json>`.

## Parameters

| Parameter | Required | Description |
|------|------|------|
| `--message-id <id>` | Yes | Message ID (`om_xxx`) to edit |
| `--text <string>` | One content option | Plain text content |
| `--markdown <string>` | One content option | Markdown text, converted to `post` JSON |
| `--content <json>` | One content option | Exact message content JSON; must match the effective `--msg-type` |
| `--set-attachments <key>` | One content option | Repeatable bare file/folder key (`file_xxx`); sets the post attachment zone. Name/size/mime/is_folder are filled by the server |
| `--clear-attachments` | One content option | Clear the post attachment zone by setting `files:[]` |
| `--msg-type <type>` | No | Message type (default `text`). When `--markdown`/`--set-attachments`/`--clear-attachments` is used the effective type is inferred automatically |
| `--as <identity>` | No | Identity type: `bot` or `user` (default `bot`) |
| `--dry-run` | No | Print the request only, do not execute it |

## Return Value

```json
{
  "message_id": "om_xxx",
  "chat_id": "oc_xxx",
  "update_time": "1234567890"
}
```

## Common Mistakes

- Editing a message the calling identity did not send — the API rejects it.
- Using `--set-attachments` with `--text`. The attachment zone only exists on `post` messages; use `--markdown` or `--msg-type post`.
- Supplying only the files you want to keep, then losing the text. Editing replaces the entire content; pass the full new content (text + attachments) in one call.
