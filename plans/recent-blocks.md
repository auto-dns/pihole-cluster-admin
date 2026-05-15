# Plan: Recent Blocks Page

## Context

Add a dedicated "Recent Blocks" troubleshooting page to pihole-cluster-admin. User story: wife reports a website is broken → open Recent Blocks → see aggregated blocked domains → click Allow → done. This is distinct from Query Logs (which shows all traffic) because it loads only blocked entries, aggregated by domain+qtype, with one-click allowlisting.

**User decisions:**
- Nav label: "Recent Blocks" | Icon: `ShieldAlert` (lucide-react)
- Default time window: last 1 hour
- Grouping: domain + record type (qtype), sorted by most recent block first
- Filters: time presets (1h/6h/24h/7d), domain search, client IP search
- Allow action: exact match by default; regex via three-dot (`MoreHorizontal`) dropdown menu

---

## Files to Create

| File | Purpose |
|------|---------|
| `frontend/src/utils/queryLogHelpers.ts` | Shared pure functions extracted from QueryLogs.tsx |
| `frontend/src/pages/RecentBlocks.tsx` | New page component |
| `frontend/src/pages/RecentBlocks.module.scss` | Page styles |

## Files to Modify

| File | Change |
|------|--------|
| `frontend/src/pages/QueryLogs.tsx` | Replace local helper definitions with imports from queryLogHelpers |
| `frontend/src/app/router.tsx` | Add `/recent-blocks` route (line 44, inside ProtectedRouteFullInit) |
| `frontend/src/components/Layout/Sidebar/Sidebar.tsx` | Add `ShieldAlert` import + nav link entry (line 23, after `/blocking`) |

---

## Step 1 — Extract `queryLogHelpers.ts`

Lift these pure functions out of `QueryLogs.tsx` into `frontend/src/utils/queryLogHelpers.ts`:
- `isBlockedStatus(status: string): boolean`
- `isForwardedStatus(status: string): boolean`
- `isCachedStatus(status: string): boolean`
- `mergeAndSort(resp: QueryLogResponse): MergedEntry[]`
- `presetRange(minutes: number): { from: string; until: string }`
- `shortStatus(status: string): string` (if it exists locally in QueryLogs.tsx)

Then update `QueryLogs.tsx` to import them from `@/utils/queryLogHelpers`. No behavior change.

---

## Step 2 — Router + Sidebar

**router.tsx** — add after the `/query` route (line 43):
```tsx
import { RecentBlocks } from '@/pages/RecentBlocks';
// ...
{
  path: '/recent-blocks',
  Component: RecentBlocks,
  handle: { layoutOptions: { pageTitle: 'Recent Blocks' } },
},
```

**Sidebar.tsx** — add `ShieldAlert` to lucide-react import, add to `links` array after `/blocking`:
```typescript
{ to: '/recent-blocks', label: 'Recent Blocks', icon: ShieldAlert },
```

---

## Step 3 — `RecentBlocks.tsx` State Shape

```typescript
type TimePreset = '1h' | '6h' | '24h' | '7d';

// Per-group allow/remove action state (Map keyed by group.key)
type GroupAllowState = {
  isAllowed: boolean;
  allowedKind: 'exact' | 'regex' | null;  // determines if Remove is available
  allowPending: boolean;
  allowFeedback: 'ok' | 'error' | null;
  removePending: boolean;
  removeFeedback: 'ok' | 'error' | null;
};
```

**State vars:**
- `preset: TimePreset = '1h'`
- `domainSearch: string`  — applied on Search click
- `clientIpSearch: string` — applied on Search click
- `appliedFilters: useRef<{ preset, domain, clientIp }>` — what was actually fetched (like QueryLogs)
- `entries: MergedEntry[]`
- `cursor: string`, `endOfResults: boolean`
- `loading: boolean`, `loadingMore: boolean`
- `error: string | null`, `nodeErrors: {id, name, err}[]`
- `expandedKeys: Set<string>` (immutable-update pattern)
- `allowState: Map<string, GroupAllowState>` (new Map on each update)
- `feedbackTimers: useRef<Map<string, NodeJS.Timeout>>`
- `regexModal: { open, domain, groupKey, pattern, submitting, error }`

---

## Step 4 — Data Fetching

**On mount:** fire two concurrent fetches:
1. `getQueryLogs({ from, until, length: 500 })` for the 1h preset → `setEntries(mergeAndSort(resp))`
2. `listDomainRules()` → build initial `allowState` from any `type==='allow' && kind==='exact'` rules

**Note on server-side status filtering:** The backend passes query params to Pi-hole. Pass no `status` filter to `getQueryLogs` — instead apply `isBlockedStatus` client-side during grouping. This avoids uncertainty about whether Pi-hole's API accepts an aggregate `"blocked"` value (QueryLogs.tsx applies status client-side for the same reason).

**On preset/domain/clientIp change:** Re-fetch from scratch; reset `entries`, `cursor`, `endOfResults`.

**Load more:** append mode, pass `{ cursor }` only (same pattern as QueryLogs).

**Cleanup:** `useEffect` cleanup clears all `feedbackTimers.current` entries.

---

## Step 5 — Grouping (useMemo)

```typescript
// Groups: computed from raw entries, sorted by most recent block first
const groups = useMemo<GroupedBase[]>(() => {
  const map = new Map<string, GroupedBase>();
  for (const entry of entries) {
    if (!isBlockedStatus(entry.status)) continue;
    const key = `${entry.domain}::${entry.qtype}`;
    const g = map.get(key) ?? { key, domain: entry.domain, qtype: entry.qtype, count: 0, lastBlockedAt: entry.time, entries: [] };
    g.count++;
    g.entries.push(entry);
    if (new Date(entry.time) > new Date(g.lastBlockedAt)) g.lastBlockedAt = entry.time;
    map.set(key, g);
  }
  return Array.from(map.values()).sort((a, b) => new Date(b.lastBlockedAt).getTime() - new Date(a.lastBlockedAt).getTime());
}, [entries]);

// Merge allow state overlay (separate memo so allow-state changes don't regroup)
const displayGroups = useMemo(() => {
  const DEFAULT_ALLOW: GroupAllowState = { isAllowed: false, allowedKind: null, allowPending: false, allowFeedback: null, removePending: false, removeFeedback: null };
  return groups.map(g => ({ ...g, ...(allowState.get(g.key) ?? DEFAULT_ALLOW) }));
}, [groups, allowState]);
```

Client-side filtering of `displayGroups` by `domainSearch` / `clientIpSearch` for the "no match" message is applied with a further `useMemo` (same pattern as QueryLogs uses for `displayedEntries`).

---

## Step 6 — Actions

**Allow (exact):**
```typescript
async function handleAllow(groupKey: string, domain: string) {
  setAllowState(prev => { const m = new Map(prev); m.set(groupKey, { ...getOrDefault(prev, groupKey), allowPending: true }); return m; });
  try {
    await addDomainRule('allow', 'exact', domain);
    updateAllowState(groupKey, { isAllowed: true, allowedKind: 'exact', allowPending: false, allowFeedback: 'ok' });
    scheduleFeedbackClear(groupKey, 'allowFeedback');
  } catch {
    updateAllowState(groupKey, { allowPending: false, allowFeedback: 'error' });
    scheduleFeedbackClear(groupKey, 'allowFeedback');
  }
}
```

**Allow (regex):** Opens `regexModal` with `pattern = domain`. On confirm: calls `addDomainRule('allow', 'regex', pattern)`. On success: sets `isAllowed: true, allowedKind: 'regex'` — but does NOT show "Remove" (regex rules are too ambiguous to remove from here).

**Remove (undo exact allow):** Calls `removeDomainRule('allow', 'exact', domain)`. Only shown when `allowedKind === 'exact'`.

**Three-dot menu:** Use `@radix-ui/react-dropdown-menu` (already in package.json). `MoreHorizontal` icon trigger → single item "Allow (Regex)".

**Feedback auto-clear:** `scheduleFeedbackClear` sets a 3s timeout via `feedbackTimers.current`, then nulls the feedback field via `updateAllowState`.

---

## Step 7 — Render Structure

```
<div.page>
  <div.filterBar>
    <div.filterRow>   ← presets: 1h | 6h | 24h | 7d
    <div.filterRow>   ← domain input, clientIp input, [Search btn], [Refresh btn]
  nodeErrors + error banners
  loading / empty state
  <div.tableInfo>  "N blocked domains • M entries total"
  <div.tableWrap>
    <table>
      thead: [expand] Domain  Type  Count  Last Blocked  [Allowed col]  Actions
      tbody:
        per displayGroup:
          <tr.groupRow [data-expanded] [data-allowed]>
            <td> expandBtn (Plus/Minus, padding:0 !important)
            <td.domainCell> domain
            <td> <span.qtypeBadge> qtype
            <td.countCell> count
            <td.timeCell> formatTime(lastBlockedAt) [title=full date]
            <td> {isAllowed && <span.allowedBadge>✓ Allowed</span>}
            <td.actionCell>
              ← feedback text OR buttons →
              if !isAllowed: <button.allowBtn> Allow + DropdownMenu (MoreHorizontal) with "Allow (Regex)"
              if isAllowed && allowedKind==='exact': <button.removeBtn> Remove
          {expanded && <tr.detailRow><td colSpan=7>
            {isAllowed && <div.allowedNotice>Allowlisted — entries below were blocked before the rule was added</div>}
            <div.detailEntries>
              {group.entries.map(e => <div.detailEntry> time | nodeName | clientIp | shortStatus(status) )}
  <div.paginationRow> Load more / end msg

<Dialog.Root> ← regex modal, same structure as Domains.tsx Add dialog
```

---

## Step 8 — Key SCSS Classes

Classes follow existing patterns from `QueryLogs.module.scss` and `Domains.module.scss`:

| Class | Notes |
|-------|-------|
| `.page` | `max-width: 76rem` (same as QueryLogs) |
| `.filterBar`, `.filterRow`, `.filterGroup`, `.filterLabel` | Identical to QueryLogs |
| `.presets`, `.presetBtn` | Identical to QueryLogs |
| `.filterInput`, `.filterActions`, `.searchBtn`, `.refreshBtn` | Identical to QueryLogs |
| `.tableWrap`, `.table`, `.tableInfo` | Identical to QueryLogs |
| `.groupRow` | `&[data-expanded]` → `background: var(--bg-header) !important`; `&[data-allowed]` → `border-left: 3px solid var(--accent-success)` |
| `.domainCell` | monospace, truncated (same as Domains) |
| `.qtypeBadge` | Same as `.kindBadge` in Domains |
| `.allowedBadge` | Green badge: `color: var(--accent-success); background: color-mix(in oklab, var(--accent-success) 12%, transparent)` |
| `.expandBtn` | `padding: 0 !important; background: none !important; color: var(--text-secondary) !important` (defeats global button reset) |
| `.allowBtn` | Same as QueryLogs: `color: var(--accent-success) !important; background: color-mix(...) !important; !important` on hover |
| `.removeBtn` | Danger variant: `color: var(--accent-danger) !important` etc |
| `.menuBtn` | Icon-only: `padding: 0 !important; background: none !important; width: 1.5rem; height: 1.5rem` |
| `.dropdownContent` | `background: var(--bg-card); border: 1px solid var(--border-card); border-radius: var(--card-radius); box-shadow: var(--card-shadow); padding: 0.25rem; z-index: 50` |
| `.dropdownItem` | `padding: 0.4rem 0.75rem; font-size: 0.85rem; cursor: pointer; &[data-highlighted]` → `background: var(--bg-muted)` |
| `.detailRow td` | `padding: 0.75rem 1rem 0.75rem 2.5rem; background: var(--bg-card-nested)` |
| `.detailEntries` | `flex-direction: column; gap: 0.35rem` |
| `.detailEntry` | `display: grid; grid-template-columns: 12rem 7rem 10rem 1fr; font-size: 0.8rem` |
| `.allowedNotice` | `font-size: 0.78rem; color: var(--accent-success); font-style: italic; background: color-mix(in oklab, var(--accent-success) 8%, transparent); border-radius: var(--border-radius); padding: 0.3rem 0.6rem; margin-bottom: 0.5rem` |
| `.actionDone` | `color: var(--accent-success); font-size: 0.8rem; font-weight: 600` |
| `.actionError` | `color: var(--accent-danger); font-size: 0.8rem; font-weight: 600` |
| `.countCell` | `font-variant-numeric: tabular-nums` |
| `.timeCell` | `font-family: monospace; font-size: 0.8rem; white-space: nowrap` |
| `.noFilterMatch` | Centered, secondary text (same as QueryLogs) |
| Modal classes | `.overlay`, `.dialog`, `.dialogTitle`, `.field`, `.fieldLabel`, `.input`, `.dialogError`, `.dialogActions`, `.cancelBtn`, `.submitBtn`, `.dialogClose` — copy verbatim from Domains.module.scss |
| `.loadingState`, `.emptyState`, `.error`, `.spin` | Identical to QueryLogs |
| `.paginationRow`, `.loadMoreBtn`, `.endMsg` | Identical to QueryLogs |

**Responsive (`@include respond-max($breakpoint-md)`):**
- `.page` → `padding: 1rem`
- `.filterGroup`, `.filterInput` → `width: 100%; min-width: 0`
- `.filterActions`, `.searchBtn`, `.refreshBtn` → `flex: 1; justify-content: center`
- Hide count column (nth-child selector) on small screens
- `.detailEntry` → `grid-template-columns: 1fr 1fr` (2-col)

---

## Critical Pitfalls

1. **Global button reset** (`globals.scss` sets `button { padding: 0.5rem 1rem }`): every fixed-size icon-only button MUST have `padding: 0 !important` in its module class or the icon will be invisible.

2. **Regex allow → no Remove**: When the user allows via regex, set `allowedKind: 'regex'` and do not render a Remove button (we can't reliably match the regex rule back to this domain from the client side).

3. **Pre-existing exact allow detection**: In the initial `listDomainRules()` response, scan all nodes for `type === 'allow' && kind === 'exact'`. Use the union of all nodes' rules (a domain is "allowed" if any node has the rule). Build a `Map<domain, 'exact' | 'regex'>` and initialize `allowState` from it.

4. **Map state updates**: Always `new Map(prev)` before `.set()` to trigger React re-render.

5. **Timer cleanup**: On unmount, clear all timers in `feedbackTimers.current.values()`.

6. **DropdownMenu z-index**: Radix portal renders at `<body>` level — ensure `.dropdownContent` has `z-index: 50`.

---

## Verification

1. Start dev server (`make dev` in `pihole-cluster-admin/`)
2. Navigate to "Recent Blocks" via sidebar
3. Confirm table loads immediately with last-hour blocked entries
4. Confirm grouping: same domain+qtype entries are collapsed into one row with correct count
5. Test expand/collapse: Plus → Minus, sub-entries visible, most recent first
6. Test "Allow" button: click → spinner → "✓ Allowed" feedback → row shows green "✓ Allowed" badge + "Remove" button
7. Test "Allow (Regex)" three-dot menu: modal appears, pre-filled, confirm → success → "✓ Allowed" badge, NO Remove button
8. Test "Remove": reverts allow state, feedback clears after 3s
9. Test loading rules on mount: if a domain is already on the allowlist, it shows "✓ Allowed" badge on page load
10. Test "allowlisted notice" in expanded rows
11. Test time preset switching: changes fetch window, re-groups data
12. Test domain + clientIp search filters
13. Run `cd frontend && npm run lint` — no new lint errors
