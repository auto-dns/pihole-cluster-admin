import { useState, useEffect, useCallback } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import {
	Pencil,
	Plus,
	Trash2,
	RefreshCw,
	X,
	CheckCircle,
	XCircle,
	AlertTriangle,
	RotateCcw,
} from 'lucide-react';
import classNames from 'classnames';
import {
	listAdlists,
	addAdlist,
	updateAdlist,
	removeAdlist,
	rebuildGravity,
} from '@/lib/api/adlists';
import { listGroups } from '@/lib/api/groups';
import type {
	ConsolidatedAdlist,
	AdlistType,
	ListAdlistsResponse,
	GravityRebuildResponse,
} from '@/types/adlist';
import type { Group } from '@/types/group';
import { formatCount } from '@/utils/formatters';
import styles from './Adlists.module.scss';

// Consolidate per-node lists into a single deduplicated view (by id).
// When the same adlist id appears on multiple nodes it's merged; when it
// appears on fewer nodes than total a parity badge is shown.
function consolidate(resp: ListAdlistsResponse): {
	adlists: ConsolidatedAdlist[];
	totalNodes: number;
} {
	const totalNodes = Object.keys(resp.nodes).length;
	const map = new Map<number, ConsolidatedAdlist>();

	for (const [idStr, nodeResult] of Object.entries(resp.nodes)) {
		const nodeId = Number(idStr);
		for (const list of nodeResult.lists) {
			if (!map.has(list.id)) {
				map.set(list.id, {
					id: list.id,
					address: list.address,
					type: list.type,
					enabled: list.enabled,
					comment: list.comment,
					groups: list.groups,
					number: list.number,
					invalidDomains: list.invalidDomains,
					dateUpdated: list.dateUpdated,
					nodeIds: [],
					totalNodes,
				});
			}
			map.get(list.id)!.nodeIds.push(nodeId);
		}
	}

	return {
		adlists: Array.from(map.values()).sort((a, b) => a.address.localeCompare(b.address)),
		totalNodes,
	};
}

type TypeFilter = 'all' | 'block' | 'allow';

type AddForm = {
	address: string;
	type: AdlistType;
	comment: string;
	groups: number[];
};

const DEFAULT_ADD_FORM: AddForm = { address: '', type: 'block', comment: '', groups: [] };

type GravityResult = {
	response: GravityRebuildResponse;
} | null;

export function Adlists() {
	const [adlists, setAdlists] = useState<ConsolidatedAdlist[]>([]);
	const [loading, setLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [typeFilter, setTypeFilter] = useState<TypeFilter>('all');
	const [gravityStale, setGravityStale] = useState(false);
	const [availableGroups, setAvailableGroups] = useState<Group[]>([]);

	// Add dialog
	const [addOpen, setAddOpen] = useState(false);
	const [addForm, setAddForm] = useState<AddForm>(DEFAULT_ADD_FORM);
	const [addSubmitting, setAddSubmitting] = useState(false);
	const [addError, setAddError] = useState<string | null>(null);

	// Edit groups dialog
	const [editGroupsAdlist, setEditGroupsAdlist] = useState<ConsolidatedAdlist | null>(null);
	const [editGroupsList, setEditGroupsList] = useState<number[]>([]);
	const [editGroupsSubmitting, setEditGroupsSubmitting] = useState(false);
	const [editGroupsError, setEditGroupsError] = useState<string | null>(null);

	// Remove confirm
	const [removeConfirm, setRemoveConfirm] = useState<number | null>(null);
	const [removing, setRemoving] = useState<number | null>(null);

	// Toggle (enable/disable)
	const [toggling, setToggling] = useState<number | null>(null);

	// Gravity rebuild
	const [rebuilding, setRebuilding] = useState(false);
	const [gravityResult, setGravityResult] = useState<GravityResult>(null);

	const load = useCallback(async () => {
		setLoading(true);
		setError(null);
		try {
			const resp = await listAdlists();
			const { adlists: consolidated } = consolidate(resp);
			setAdlists(consolidated);
		} catch (e) {
			setError(e instanceof Error ? e.message : 'Failed to load adlists');
		} finally {
			setLoading(false);
		}
	}, []);

	useEffect(() => {
		load();
	}, [load]);

	useEffect(() => {
		listGroups()
			.then((resp) => {
				const map = new Map<number, Group>();
				for (const nr of Object.values(resp.nodes)) {
					for (const g of nr.groups) {
						if (!map.has(g.id)) map.set(g.id, g);
					}
				}
				setAvailableGroups(Array.from(map.values()).sort((a, b) => a.id - b.id));
			})
			.catch(() => { /* non-fatal — group pickers stay empty */ });
	}, []);

	const filtered = adlists.filter((a) => typeFilter === 'all' || a.type === typeFilter);

	const groupNameById = new Map(availableGroups.map((g) => [g.id, g.name]));

	// Edit groups
	const openEditGroups = (adlist: ConsolidatedAdlist) => {
		setEditGroupsAdlist(adlist);
		setEditGroupsList([...adlist.groups]);
		setEditGroupsError(null);
	};

	const handleSaveGroups = async () => {
		if (!editGroupsAdlist) return;
		setEditGroupsSubmitting(true);
		setEditGroupsError(null);
		try {
			await updateAdlist(editGroupsAdlist.id, { groups: editGroupsList });
			setEditGroupsAdlist(null);
			await load();
		} catch (e) {
			setEditGroupsError(e instanceof Error ? e.message : 'Failed to update groups');
		} finally {
			setEditGroupsSubmitting(false);
		}
	};

	// Add
	const handleAdd = async () => {
		if (!addForm.address.trim()) {
			setAddError('Address is required');
			return;
		}
		setAddSubmitting(true);
		setAddError(null);
		try {
			await addAdlist(
				addForm.address.trim(),
				addForm.type,
				addForm.comment.trim() || undefined,
				addForm.groups.length > 0 ? addForm.groups : undefined,
			);
			setAddOpen(false);
			setAddForm(DEFAULT_ADD_FORM);
			setGravityStale(true);
			setGravityResult(null);
			await load();
		} catch (e) {
			setAddError(e instanceof Error ? e.message : 'Failed to add adlist');
		} finally {
			setAddSubmitting(false);
		}
	};

	// Toggle enabled
	const handleToggle = async (adlist: ConsolidatedAdlist) => {
		setToggling(adlist.id);
		try {
			await updateAdlist(adlist.id, { enabled: !adlist.enabled });
			setGravityStale(true);
			setGravityResult(null);
			await load();
		} catch (e) {
			setError(e instanceof Error ? e.message : 'Failed to update adlist');
		} finally {
			setToggling(null);
		}
	};

	// Remove
	const handleRemove = async (id: number) => {
		setRemoving(id);
		setRemoveConfirm(null);
		try {
			await removeAdlist(id);
			setGravityStale(true);
			setGravityResult(null);
			await load();
		} catch (e) {
			setError(e instanceof Error ? e.message : 'Failed to remove adlist');
		} finally {
			setRemoving(null);
		}
	};

	// Rebuild gravity
	const handleRebuild = async () => {
		setRebuilding(true);
		setGravityResult(null);
		try {
			const resp = await rebuildGravity();
			setGravityResult({ response: resp });
			if (resp.summary.failed === 0) {
				setGravityStale(false);
			}
			await load();
		} catch (e) {
			setError(e instanceof Error ? e.message : 'Failed to rebuild gravity');
		} finally {
			setRebuilding(false);
		}
	};

	return (
		<div className={styles.page}>
			<div className={styles.pageHeader}>
				<h1 className={styles.pageTitle}>Adlists</h1>
				<div className={styles.headerActions}>
					<button
						className={styles.refreshBtn}
						onClick={load}
						disabled={loading || rebuilding}
						aria-label='Refresh adlists'
					>
						<RefreshCw size={15} className={classNames({ [styles.spinning]: loading })} />
						Refresh
					</button>
					<button
						className={styles.addBtn}
						onClick={() => {
							setAddForm(DEFAULT_ADD_FORM);
							setAddError(null);
							setAddOpen(true);
						}}
						disabled={loading}
					>
						<Plus size={15} />
						Add adlist
					</button>
				</div>
			</div>

			{error && <div className={styles.errorBanner}>{error}</div>}

			{/* Stale gravity warning */}
			{gravityStale && !rebuilding && (
				<div className={styles.gravityWarning}>
					<AlertTriangle size={15} />
					Gravity not rebuilt — changes won&apos;t take effect until you rebuild.
					<button className={styles.rebuildBtn} onClick={handleRebuild} disabled={rebuilding}>
						<RotateCcw size={14} />
						Rebuild Gravity
					</button>
				</div>
			)}

			{/* Gravity rebuild result */}
			{rebuilding && (
				<div className={styles.gravityProgress}>
					<RefreshCw size={14} className={styles.spinning} />
					Rebuilding gravity on all nodes… this may take up to 2 minutes.
				</div>
			)}
			{gravityResult && !rebuilding && (
				<div
					className={classNames(styles.gravityResult, {
						[styles.gravityResultOk]: gravityResult.response.summary.failed === 0,
						[styles.gravityResultPartial]: gravityResult.response.summary.failed > 0,
					})}
				>
					<div className={styles.gravityResultTitle}>
						{gravityResult.response.summary.failed === 0
							? 'Gravity rebuilt successfully on all nodes'
							: `Gravity rebuild failed on ${gravityResult.response.summary.failed} of ${gravityResult.response.summary.total} nodes`}
					</div>
					<ul className={styles.gravityNodeList}>
						{Object.values(gravityResult.response.nodes).map((n) => (
							<li key={n.node.id} className={styles.gravityNodeItem}>
								{n.success ? (
									<CheckCircle size={13} className={styles.iconOk} />
								) : (
									<XCircle size={13} className={styles.iconErr} />
								)}
								<span>{n.node.name}</span>
								{n.error && <span className={styles.nodeError}>— {n.error}</span>}
							</li>
						))}
					</ul>
					{gravityResult.response.summary.failed === 0 ? null : (
						<button
							className={styles.rebuildBtn}
							onClick={handleRebuild}
							disabled={rebuilding}
							style={{ marginTop: '0.5rem' }}
						>
							<RotateCcw size={14} />
							Retry Rebuild
						</button>
					)}
				</div>
			)}

			{/* Rebuild gravity button (when no stale warning and no result shown) */}
			{!gravityStale && !gravityResult && !rebuilding && (
				<div className={styles.gravityBar}>
					<button className={styles.rebuildBtn} onClick={handleRebuild} disabled={rebuilding}>
						<RotateCcw size={14} />
						Rebuild Gravity
					</button>
				</div>
			)}

			{/* Type filter */}
			<div className={styles.filterBar}>
				<span className={styles.filterLabel}>Type:</span>
				<div className={styles.filterButtons} role='group' aria-label='Adlist type filter'>
					{(['all', 'block', 'allow'] as const).map((t) => (
						<button
							key={t}
							className={classNames(styles.filterBtn, { [styles.filterBtnActive]: typeFilter === t })}
							onClick={() => setTypeFilter(t)}
						>
							{t === 'all' ? 'All' : t === 'block' ? 'Blocklist' : 'Allowlist'}
						</button>
					))}
				</div>
			</div>

			{/* Table */}
			<div className={styles.tableCard}>
				{loading ? (
					<div className={styles.empty}>Loading…</div>
				) : filtered.length === 0 ? (
					<div className={styles.empty}>No adlists found.</div>
				) : (
					<table className={styles.table}>
						<thead>
							<tr>
								<th>Address</th>
								<th>Type</th>
								<th className={styles.numCol}>Entries</th>
								<th>Groups</th>
								<th>Last updated</th>
								<th>Enabled</th>
								<th aria-label='Actions' />
							</tr>
						</thead>
						<tbody>
							{filtered.map((adlist) => {
								const isPartial = adlist.nodeIds.length < adlist.totalNodes;
								return (
									<tr key={adlist.id} className={classNames({ [styles.rowPartial]: isPartial })}>
										<td className={styles.addressCell}>
											<span className={styles.address}>{adlist.address}</span>
											{adlist.comment && (
												<span className={styles.comment}>{adlist.comment}</span>
											)}
											{isPartial && (
												<span
													className={styles.parityBadge}
													title={`Present on ${adlist.nodeIds.length} of ${adlist.totalNodes} nodes`}
												>
													{adlist.nodeIds.length}/{adlist.totalNodes} nodes
												</span>
											)}
										</td>
										<td>
											<span
												className={classNames(styles.typeBadge, {
													[styles.typeBadgeBlock]: adlist.type === 'block',
													[styles.typeBadgeAllow]: adlist.type === 'allow',
												})}
											>
												{adlist.type === 'block' ? 'Blocklist' : 'Allowlist'}
											</span>
										</td>
										<td className={styles.numCol}>
											{formatCount(adlist.number)}
											{adlist.invalidDomains > 0 && (
												<span
													className={styles.invalidBadge}
													title={`${adlist.invalidDomains} invalid entries`}
												>
													{adlist.invalidDomains} invalid
												</span>
											)}
										</td>
										<td>
											{adlist.groups.length === 0 ? (
												<span className={styles.noGroups}>—</span>
											) : (
												<div className={styles.groupBadges}>
													{adlist.groups.map((gid) => (
														<span key={gid} className={styles.groupBadge}>
															{groupNameById.get(gid) ?? `Group ${gid}`}
														</span>
													))}
												</div>
											)}
										</td>
										<td className={styles.dateCell}>
											{adlist.dateUpdated
												? new Date(adlist.dateUpdated).toLocaleDateString()
												: '—'}
										</td>
										<td>
											<button
												className={classNames(styles.toggleBtn, {
													[styles.toggleBtnOn]: adlist.enabled,
													[styles.toggleBtnOff]: !adlist.enabled,
												})}
												onClick={() => handleToggle(adlist)}
												disabled={toggling === adlist.id}
												aria-label={adlist.enabled ? 'Disable adlist' : 'Enable adlist'}
												aria-pressed={adlist.enabled}
											>
												{adlist.enabled ? 'Enabled' : 'Disabled'}
											</button>
										</td>
										<td className={styles.actionsCell}>
											{removeConfirm === adlist.id ? (
												<span className={styles.confirmRow}>
													Remove?
													<button
														className={styles.confirmYes}
														onClick={() => handleRemove(adlist.id)}
														disabled={removing === adlist.id}
													>
														Yes
													</button>
													<button
														className={styles.confirmNo}
														onClick={() => setRemoveConfirm(null)}
													>
														No
													</button>
												</span>
											) : (
												<span className={styles.actionBtns}>
													<button
														className={styles.editGroupsBtn}
														onClick={() => openEditGroups(adlist)}
														aria-label='Assign groups'
														title='Assign groups'
													>
														<Pencil size={14} />
													</button>
													<button
														className={styles.removeBtn}
														onClick={() => setRemoveConfirm(adlist.id)}
														disabled={removing === adlist.id}
														aria-label='Remove adlist'
													>
														<Trash2 size={14} />
													</button>
												</span>
											)}
										</td>
									</tr>
								);
							})}
						</tbody>
					</table>
				)}
			</div>

			{/* Add adlist dialog */}
			<Dialog.Root open={addOpen} onOpenChange={setAddOpen}>
				<Dialog.Portal>
					<Dialog.Overlay className={styles.dialogOverlay} />
					<Dialog.Content className={styles.dialog} aria-describedby={undefined}>
						<div className={styles.dialogHeader}>
							<Dialog.Title className={styles.dialogTitle}>Add adlist</Dialog.Title>
							<Dialog.Close className={styles.dialogClose} aria-label='Close'>
								<X size={16} />
							</Dialog.Close>
						</div>

						{addError && <div className={styles.dialogError}>{addError}</div>}

						<div className={styles.formGroup}>
							<label className={styles.label} htmlFor='adlist-address'>
								URL
							</label>
							<input
								id='adlist-address'
								className={styles.input}
								type='url'
								placeholder='https://example.com/hosts.txt'
								value={addForm.address}
								onChange={(e) => setAddForm((f) => ({ ...f, address: e.target.value }))}
								autoFocus
							/>
						</div>

						<div className={styles.formGroup}>
							<label className={styles.label}>Type</label>
							<div className={styles.radioGroup}>
								{(['block', 'allow'] as const).map((t) => (
									<label key={t} className={styles.radioLabel}>
										<input
											type='radio'
											name='adlist-type'
											value={t}
											checked={addForm.type === t}
											onChange={() => setAddForm((f) => ({ ...f, type: t }))}
										/>
										{t === 'block' ? 'Blocklist' : 'Allowlist'}
									</label>
								))}
							</div>
						</div>

						<div className={styles.formGroup}>
							<label className={styles.label} htmlFor='adlist-comment'>
								Comment <span className={styles.optional}>(optional)</span>
							</label>
							<input
								id='adlist-comment'
								className={styles.input}
								type='text'
								placeholder='e.g. Main blocklist'
								value={addForm.comment}
								onChange={(e) => setAddForm((f) => ({ ...f, comment: e.target.value }))}
							/>
						</div>

						{availableGroups.length > 0 && (
							<div className={styles.formGroup}>
								<span className={styles.label}>
									Groups <span className={styles.optional}>(optional)</span>
								</span>
								<div className={styles.groupCheckboxList}>
									{availableGroups.map((g) => (
										<label key={g.id} className={styles.groupCheckboxItem}>
											<input
												type='checkbox'
												checked={addForm.groups.includes(g.id)}
												onChange={() =>
													setAddForm((f) => ({
														...f,
														groups: f.groups.includes(g.id)
															? f.groups.filter((x) => x !== g.id)
															: [...f.groups, g.id],
													}))
												}
												disabled={addSubmitting}
											/>
											{g.name}
										</label>
									))}
								</div>
							</div>
						)}

						<div className={styles.dialogFooter}>
							<Dialog.Close className={styles.cancelBtn} disabled={addSubmitting}>
								Cancel
							</Dialog.Close>
							<button
								className={styles.submitBtn}
								onClick={handleAdd}
								disabled={addSubmitting || !addForm.address.trim()}
							>
								{addSubmitting ? 'Adding…' : 'Add adlist'}
							</button>
						</div>
					</Dialog.Content>
				</Dialog.Portal>
			</Dialog.Root>

			{/* Edit groups dialog */}
			<Dialog.Root
				open={editGroupsAdlist !== null}
				onOpenChange={(next) => { if (!next) { setEditGroupsAdlist(null); setEditGroupsError(null); } }}
			>
				<Dialog.Portal>
					<Dialog.Overlay className={styles.dialogOverlay} />
					<Dialog.Content className={styles.dialog} aria-describedby={undefined}>
						<div className={styles.dialogHeader}>
							<Dialog.Title className={styles.dialogTitle}>
								Assign groups — {editGroupsAdlist?.address}
							</Dialog.Title>
							<Dialog.Close className={styles.dialogClose} aria-label='Close'>
								<X size={16} />
							</Dialog.Close>
						</div>

						{editGroupsError && <div className={styles.dialogError}>{editGroupsError}</div>}

						<div className={styles.formGroup}>
							{availableGroups.length === 0 ? (
								<p className={styles.optional}>No groups available.</p>
							) : (
								<div className={styles.groupCheckboxList}>
									{availableGroups.map((g) => (
										<label key={g.id} className={styles.groupCheckboxItem}>
											<input
												type='checkbox'
												checked={editGroupsList.includes(g.id)}
												onChange={() =>
													setEditGroupsList((prev) =>
														prev.includes(g.id)
															? prev.filter((x) => x !== g.id)
															: [...prev, g.id],
													)
												}
												disabled={editGroupsSubmitting}
											/>
											{g.name}{g.description ? ` — ${g.description}` : ''}
										</label>
									))}
								</div>
							)}
						</div>

						<div className={styles.dialogFooter}>
							<Dialog.Close className={styles.cancelBtn} disabled={editGroupsSubmitting}>
								Cancel
							</Dialog.Close>
							<button
								className={styles.submitBtn}
								onClick={handleSaveGroups}
								disabled={editGroupsSubmitting}
							>
								{editGroupsSubmitting ? 'Saving…' : 'Save'}
							</button>
						</div>
					</Dialog.Content>
				</Dialog.Portal>
			</Dialog.Root>
		</div>
	);
}
