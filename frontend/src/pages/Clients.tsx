import { useState, useEffect, useCallback } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { Plus, Pencil, Trash2, RefreshCw, X } from 'lucide-react';
import { listGroups, addGroup, updateGroup, removeGroup } from '@/lib/api/groups';
import { listPiholeClients, updatePiholeClient, removePiholeClient } from '@/lib/api/piholeClients';
import type { Group, ListGroupsResponse } from '@/types/group';
import type { PiholeClient, ListClientsResponse } from '@/types/pihole_client';
import styles from './Clients.module.scss';

// ── Consolidation ────────────────────────────────────────────────

function consolidateGroups(resp: ListGroupsResponse): Group[] {
	const map = new Map<number, Group>();
	for (const nr of Object.values(resp.nodes)) {
		for (const g of nr.groups) {
			if (!map.has(g.id)) map.set(g.id, g);
		}
	}
	return Array.from(map.values()).sort((a, b) => a.id - b.id);
}

function consolidateClients(resp: ListClientsResponse): PiholeClient[] {
	const map = new Map<number, PiholeClient>();
	for (const nr of Object.values(resp.nodes)) {
		for (const c of nr.clients) {
			if (!map.has(c.id)) map.set(c.id, c);
		}
	}
	return Array.from(map.values()).sort((a, b) => a.ip.localeCompare(b.ip));
}

// ── Types ─────────────────────────────────────────────────────────

type AddGroupForm = { name: string; description: string; enabled: boolean };
type EditGroupForm = { description: string; enabled: boolean };

const DEFAULT_ADD_GROUP: AddGroupForm = { name: '', description: '', enabled: true };

// ── Groups panel ─────────────────────────────────────────────────

type GroupsPanelProps = {
	groups: Group[];
	loading: boolean;
	error: string | null;
	onRefresh: () => void;
};

function GroupsPanel({ groups, loading, error, onRefresh }: GroupsPanelProps) {
	const [addOpen, setAddOpen] = useState(false);
	const [addForm, setAddForm] = useState<AddGroupForm>(DEFAULT_ADD_GROUP);
	const [addSubmitting, setAddSubmitting] = useState(false);
	const [addError, setAddError] = useState<string | null>(null);

	const [editGroup, setEditGroup] = useState<Group | null>(null);
	const [editForm, setEditForm] = useState<EditGroupForm>({ description: '', enabled: true });
	const [editSubmitting, setEditSubmitting] = useState(false);
	const [editError, setEditError] = useState<string | null>(null);

	const [removeConfirm, setRemoveConfirm] = useState<string | null>(null);
	const [removing, setRemoving] = useState<string | null>(null);
	const [removeError, setRemoveError] = useState<string | null>(null);

	async function handleAdd() {
		setAddError(null);
		if (!addForm.name.trim()) { setAddError('Name is required'); return; }
		setAddSubmitting(true);
		try {
			await addGroup(addForm.name.trim(), addForm.description.trim() || undefined, addForm.enabled);
			setAddOpen(false);
			setAddForm(DEFAULT_ADD_GROUP);
			onRefresh();
		} catch (err) {
			setAddError(err instanceof Error ? err.message : 'Failed to add group');
		} finally {
			setAddSubmitting(false);
		}
	}

	function openEdit(g: Group) {
		setEditGroup(g);
		setEditForm({ description: g.description ?? '', enabled: g.enabled });
		setEditError(null);
	}

	async function handleEdit() {
		if (!editGroup) return;
		setEditError(null);
		setEditSubmitting(true);
		try {
			await updateGroup(editGroup.name, {
				description: editForm.description.trim() || null,
				enabled: editForm.enabled,
			});
			setEditGroup(null);
			onRefresh();
		} catch (err) {
			setEditError(err instanceof Error ? err.message : 'Failed to update group');
		} finally {
			setEditSubmitting(false);
		}
	}

	async function handleRemove(name: string) {
		setRemoveError(null);
		setRemoving(name);
		try {
			await removeGroup(name);
			setRemoveConfirm(null);
			onRefresh();
		} catch (err) {
			setRemoveError(err instanceof Error ? err.message : 'Failed to remove group');
		} finally {
			setRemoving(null);
		}
	}

	return (
		<div className={styles.section}>
			<div className={styles.sectionHeader}>
				<h2 className={styles.sectionTitle}>Groups</h2>
				<div className={styles.headerActions}>
					<button
						type='button'
						className={styles.refreshBtn}
						onClick={onRefresh}
						disabled={loading}
						aria-label='Refresh groups'
					>
						<RefreshCw size={15} className={loading ? styles.spin : undefined} />
					</button>

					<Dialog.Root
						open={addOpen}
						onOpenChange={(next) => {
							setAddOpen(next);
							if (!next) { setAddForm(DEFAULT_ADD_GROUP); setAddError(null); }
						}}
					>
						<Dialog.Trigger asChild>
							<button type='button' className={styles.addBtn}>
								<Plus size={15} /> Add group
							</button>
						</Dialog.Trigger>
						<Dialog.Portal>
							<Dialog.Overlay className={styles.overlay} />
							<Dialog.Content className={styles.dialog}>
								<Dialog.Title className={styles.dialogTitle}>Add group</Dialog.Title>
								<div className={styles.field}>
									<label htmlFor='add-group-name' className={styles.fieldLabel}>Name</label>
									<input
										id='add-group-name'
										type='text'
										className={styles.input}
										placeholder='e.g. Kids Devices'
										value={addForm.name}
										onChange={(e) => setAddForm((f) => ({ ...f, name: e.target.value }))}
										disabled={addSubmitting}
									/>
								</div>
								<div className={styles.field}>
									<label htmlFor='add-group-desc' className={styles.fieldLabel}>
										Description <span className={styles.fieldHint}>(optional)</span>
									</label>
									<input
										id='add-group-desc'
										type='text'
										className={styles.input}
										placeholder='Short description'
										value={addForm.description}
										onChange={(e) => setAddForm((f) => ({ ...f, description: e.target.value }))}
										disabled={addSubmitting}
									/>
								</div>
								<div className={styles.field}>
									<label className={styles.checkboxField}>
										<input
											type='checkbox'
											checked={addForm.enabled}
											onChange={(e) => setAddForm((f) => ({ ...f, enabled: e.target.checked }))}
											disabled={addSubmitting}
										/>
										Enabled
									</label>
								</div>
								{addError && <p className={styles.dialogError}>{addError}</p>}
								<div className={styles.dialogActions}>
									<Dialog.Close asChild>
										<button type='button' className={styles.cancelBtn} disabled={addSubmitting}>Cancel</button>
									</Dialog.Close>
									<button
										type='button'
										className={styles.submitBtn}
										onClick={handleAdd}
										disabled={addSubmitting || !addForm.name.trim()}
										aria-busy={addSubmitting}
									>
										{addSubmitting ? <RefreshCw size={14} className={styles.spin} /> : null}
										Add
									</button>
								</div>
								<Dialog.Close asChild>
									<button className={styles.dialogClose} aria-label='Close'><X size={18} /></button>
								</Dialog.Close>
							</Dialog.Content>
						</Dialog.Portal>
					</Dialog.Root>
				</div>
			</div>

			{error && <div className={styles.error}>{error}</div>}
			{removeError && <div className={styles.error}>{removeError}</div>}

			{loading && groups.length === 0 && (
				<div className={styles.loadingState}><RefreshCw size={18} className={styles.spin} /> Loading…</div>
			)}
			{!loading && !error && groups.length === 0 && (
				<div className={styles.emptyState}>No groups configured.</div>
			)}

			{groups.length > 0 && (
				<>
					<div className={styles.tableInfo}>{groups.length} {groups.length === 1 ? 'group' : 'groups'}</div>
					<div className={styles.tableWrap}>
						<table className={styles.table}>
							<thead>
								<tr>
									<th>ID</th>
									<th>Name</th>
									<th>Description</th>
									<th>Status</th>
									<th aria-label='Actions' />
								</tr>
							</thead>
							<tbody>
								{groups.map((g) => {
									const isDefault = g.id === 0;
									const confirming = removeConfirm === g.name;
									const isRemoving = removing === g.name;
									return (
										<tr key={g.id}>
											<td>{g.id}</td>
											<td>{g.name}</td>
											<td>{g.description ?? '—'}</td>
											<td>
												<span className={styles.enabledBadge} data-enabled={String(g.enabled)}>
													{g.enabled ? 'Enabled' : 'Disabled'}
												</span>
											</td>
											<td className={styles.actionCell}>
												{confirming ? (
													<div className={styles.confirmRow}>
														<span className={styles.confirmText}>Remove?</span>
														<button
															type='button'
															className={styles.confirmYesBtn}
															onClick={() => handleRemove(g.name)}
															disabled={isRemoving}
														>
															{isRemoving ? <RefreshCw size={12} className={styles.spin} /> : 'Yes'}
														</button>
														<button
															type='button'
															className={styles.confirmNoBtn}
															onClick={() => { setRemoveConfirm(null); setRemoveError(null); }}
															disabled={isRemoving}
														>No</button>
													</div>
												) : (
													<div className={styles.actionRow}>
														<button
															type='button'
															className={styles.editBtn}
															onClick={() => openEdit(g)}
															aria-label={`Edit ${g.name}`}
														>
															<Pencil size={14} />
														</button>
														{!isDefault && (
															<button
																type='button'
																className={styles.removeBtn}
																onClick={() => setRemoveConfirm(g.name)}
																aria-label={`Remove ${g.name}`}
															>
																<Trash2 size={14} />
															</button>
														)}
													</div>
												)}
											</td>
										</tr>
									);
								})}
							</tbody>
						</table>
					</div>
				</>
			)}

			{/* Edit dialog */}
			<Dialog.Root
				open={editGroup !== null}
				onOpenChange={(next) => { if (!next) { setEditGroup(null); setEditError(null); } }}
			>
				<Dialog.Portal>
					<Dialog.Overlay className={styles.overlay} />
					<Dialog.Content className={styles.dialog}>
						<Dialog.Title className={styles.dialogTitle}>Edit group — {editGroup?.name}</Dialog.Title>
						<div className={styles.field}>
							<label htmlFor='edit-group-desc' className={styles.fieldLabel}>
								Description <span className={styles.fieldHint}>(optional)</span>
							</label>
							<input
								id='edit-group-desc'
								type='text'
								className={styles.input}
								placeholder='Short description'
								value={editForm.description}
								onChange={(e) => setEditForm((f) => ({ ...f, description: e.target.value }))}
								disabled={editSubmitting}
							/>
						</div>
						<div className={styles.field}>
							<label className={styles.checkboxField}>
								<input
									type='checkbox'
									checked={editForm.enabled}
									onChange={(e) => setEditForm((f) => ({ ...f, enabled: e.target.checked }))}
									disabled={editSubmitting}
								/>
								Enabled
							</label>
						</div>
						{editError && <p className={styles.dialogError}>{editError}</p>}
						<div className={styles.dialogActions}>
							<Dialog.Close asChild>
								<button type='button' className={styles.cancelBtn} disabled={editSubmitting}>Cancel</button>
							</Dialog.Close>
							<button
								type='button'
								className={styles.submitBtn}
								onClick={handleEdit}
								disabled={editSubmitting}
								aria-busy={editSubmitting}
							>
								{editSubmitting ? <RefreshCw size={14} className={styles.spin} /> : null}
								Save
							</button>
						</div>
						<Dialog.Close asChild>
							<button className={styles.dialogClose} aria-label='Close'><X size={18} /></button>
						</Dialog.Close>
					</Dialog.Content>
				</Dialog.Portal>
			</Dialog.Root>
		</div>
	);
}

// ── Clients panel ─────────────────────────────────────────────────

type ClientsPanelProps = {
	groups: Group[];
};

function ClientsPanel({ groups }: ClientsPanelProps) {
	const [clients, setClients] = useState<PiholeClient[]>([]);
	const [loading, setLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);

	const [editClient, setEditClient] = useState<PiholeClient | null>(null);
	const [editGroups, setEditGroups] = useState<number[]>([]);
	const [editSubmitting, setEditSubmitting] = useState(false);
	const [editError, setEditError] = useState<string | null>(null);

	const [removeConfirm, setRemoveConfirm] = useState<number | null>(null);
	const [removing, setRemoving] = useState<number | null>(null);
	const [removeError, setRemoveError] = useState<string | null>(null);

	const fetchClients = useCallback(async () => {
		setLoading(true);
		setError(null);
		try {
			const resp = await listPiholeClients();
			setClients(consolidateClients(resp));
		} catch (err) {
			setError(err instanceof Error ? err.message : 'Failed to load clients');
		} finally {
			setLoading(false);
		}
	}, []);

	useEffect(() => { fetchClients(); }, [fetchClients]);

	function openEdit(c: PiholeClient) {
		setEditClient(c);
		setEditGroups([...c.groups]);
		setEditError(null);
	}

	function toggleGroup(id: number) {
		setEditGroups((prev) => prev.includes(id) ? prev.filter((g) => g !== id) : [...prev, id]);
	}

	async function handleEditGroups() {
		if (!editClient) return;
		setEditError(null);
		setEditSubmitting(true);
		try {
			await updatePiholeClient(editClient.id, { groups: editGroups });
			setEditClient(null);
			await fetchClients();
		} catch (err) {
			setEditError(err instanceof Error ? err.message : 'Failed to update client');
		} finally {
			setEditSubmitting(false);
		}
	}

	async function handleRemove(id: number) {
		setRemoveError(null);
		setRemoving(id);
		try {
			await removePiholeClient(id);
			setRemoveConfirm(null);
			await fetchClients();
		} catch (err) {
			setRemoveError(err instanceof Error ? err.message : 'Failed to remove client');
		} finally {
			setRemoving(null);
		}
	}

	const groupNameById = new Map(groups.map((g) => [g.id, g.name]));

	return (
		<div className={styles.section}>
			<div className={styles.sectionHeader}>
				<h2 className={styles.sectionTitle}>Clients</h2>
				<div className={styles.headerActions}>
					<button
						type='button'
						className={styles.refreshBtn}
						onClick={fetchClients}
						disabled={loading}
						aria-label='Refresh clients'
					>
						<RefreshCw size={15} className={loading ? styles.spin : undefined} />
					</button>
				</div>
			</div>

			{error && <div className={styles.error}>{error}</div>}
			{removeError && <div className={styles.error}>{removeError}</div>}

			{loading && clients.length === 0 && (
				<div className={styles.loadingState}><RefreshCw size={18} className={styles.spin} /> Loading…</div>
			)}
			{!loading && !error && clients.length === 0 && (
				<div className={styles.emptyState}>No clients configured.</div>
			)}

			{clients.length > 0 && (
				<>
					<div className={styles.tableInfo}>{clients.length} {clients.length === 1 ? 'client' : 'clients'}</div>
					<div className={styles.tableWrap}>
						<table className={styles.table}>
							<thead>
								<tr>
									<th>IP</th>
									<th>Name</th>
									<th>Groups</th>
									<th aria-label='Actions' />
								</tr>
							</thead>
							<tbody>
								{clients.map((c) => {
									const confirming = removeConfirm === c.id;
									const isRemoving = removing === c.id;
									return (
										<tr key={c.id}>
											<td>{c.ip}</td>
											<td>{c.name || '—'}</td>
											<td>
												{c.groups.length === 0 ? (
													<span className={styles.noGroups}>—</span>
												) : (
													<div className={styles.groupBadges}>
														{c.groups.map((gid) => (
															<span key={gid} className={styles.groupBadge}>
																{groupNameById.get(gid) ?? `Group ${gid}`}
															</span>
														))}
													</div>
												)}
											</td>
											<td className={styles.actionCell}>
												{confirming ? (
													<div className={styles.confirmRow}>
														<span className={styles.confirmText}>Remove?</span>
														<button
															type='button'
															className={styles.confirmYesBtn}
															onClick={() => handleRemove(c.id)}
															disabled={isRemoving}
														>
															{isRemoving ? <RefreshCw size={12} className={styles.spin} /> : 'Yes'}
														</button>
														<button
															type='button'
															className={styles.confirmNoBtn}
															onClick={() => { setRemoveConfirm(null); setRemoveError(null); }}
															disabled={isRemoving}
														>No</button>
													</div>
												) : (
													<div className={styles.actionRow}>
														<button
															type='button'
															className={styles.editBtn}
															onClick={() => openEdit(c)}
															aria-label={`Assign groups for ${c.ip}`}
															title='Assign groups'
														>
															<Pencil size={14} />
														</button>
														<button
															type='button'
															className={styles.removeBtn}
															onClick={() => setRemoveConfirm(c.id)}
															aria-label={`Remove ${c.ip}`}
														>
															<Trash2 size={14} />
														</button>
													</div>
												)}
											</td>
										</tr>
									);
								})}
							</tbody>
						</table>
					</div>
				</>
			)}

			{/* Edit client groups dialog */}
			<Dialog.Root
				open={editClient !== null}
				onOpenChange={(next) => { if (!next) { setEditClient(null); setEditError(null); } }}
			>
				<Dialog.Portal>
					<Dialog.Overlay className={styles.overlay} />
					<Dialog.Content className={styles.dialog}>
						<Dialog.Title className={styles.dialogTitle}>
							Assign groups — {editClient?.ip}
						</Dialog.Title>
						<div className={styles.field}>
							<span className={styles.fieldLabel}>Groups</span>
							{groups.length === 0 ? (
								<p className={styles.fieldHint}>No groups available.</p>
							) : (
								<div className={styles.groupCheckboxList}>
									{groups.map((g) => (
										<label key={g.id} className={styles.groupCheckboxItem}>
											<input
												type='checkbox'
												checked={editGroups.includes(g.id)}
												onChange={() => toggleGroup(g.id)}
												disabled={editSubmitting}
											/>
											{g.name}{g.description ? ` — ${g.description}` : ''}
										</label>
									))}
								</div>
							)}
						</div>
						{editError && <p className={styles.dialogError}>{editError}</p>}
						<div className={styles.dialogActions}>
							<Dialog.Close asChild>
								<button type='button' className={styles.cancelBtn} disabled={editSubmitting}>Cancel</button>
							</Dialog.Close>
							<button
								type='button'
								className={styles.submitBtn}
								onClick={handleEditGroups}
								disabled={editSubmitting}
								aria-busy={editSubmitting}
							>
								{editSubmitting ? <RefreshCw size={14} className={styles.spin} /> : null}
								Save
							</button>
						</div>
						<Dialog.Close asChild>
							<button className={styles.dialogClose} aria-label='Close'><X size={18} /></button>
						</Dialog.Close>
					</Dialog.Content>
				</Dialog.Portal>
			</Dialog.Root>
		</div>
	);
}

// ── Page ──────────────────────────────────────────────────────────

export function Clients() {
	const [groups, setGroups] = useState<Group[]>([]);
	const [groupsLoading, setGroupsLoading] = useState(false);
	const [groupsError, setGroupsError] = useState<string | null>(null);

	const fetchGroups = useCallback(async () => {
		setGroupsLoading(true);
		setGroupsError(null);
		try {
			const resp = await listGroups();
			setGroups(consolidateGroups(resp));
		} catch (err) {
			setGroupsError(err instanceof Error ? err.message : 'Failed to load groups');
		} finally {
			setGroupsLoading(false);
		}
	}, []);

	useEffect(() => { fetchGroups(); }, [fetchGroups]);

	return (
		<div className={styles.page}>
			<GroupsPanel
				groups={groups}
				loading={groupsLoading}
				error={groupsError}
				onRefresh={fetchGroups}
			/>
			<ClientsPanel groups={groups} />
		</div>
	);
}
