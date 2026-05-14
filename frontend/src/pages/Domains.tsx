import { useState, useEffect, useCallback } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { Plus, Trash2, RefreshCw, X } from 'lucide-react';
import { listDomainRules, addDomainRule, removeDomainRule } from '@/lib/api/domainrules';
import type { ConsolidatedRule, RuleType, RuleKind, ListDomainRulesResponse } from '@/types/domainrule';
import styles from './Domains.module.scss';

function consolidate(resp: ListDomainRulesResponse): { rules: ConsolidatedRule[]; nodeNames: Map<number, string> } {
	const totalNodes = Object.keys(resp.nodes).length;
	const ruleMap = new Map<string, ConsolidatedRule>();
	const nodeNames = new Map<number, string>();

	for (const [idStr, nodeResult] of Object.entries(resp.nodes)) {
		const nodeId = Number(idStr);
		nodeNames.set(nodeId, nodeResult.node.name);

		for (const rule of nodeResult.rules) {
			const key = `${rule.domain}:${rule.type}:${rule.kind}`;
			if (!ruleMap.has(key)) {
				ruleMap.set(key, {
					key,
					domain: rule.domain,
					type: rule.type as RuleType,
					kind: rule.kind as RuleKind,
					enabled: rule.enabled,
					comment: rule.comment,
					nodeIds: [],
					totalNodes,
				});
			}
			ruleMap.get(key)!.nodeIds.push(nodeId);
		}
	}

	const rules = Array.from(ruleMap.values()).sort((a, b) => a.domain.localeCompare(b.domain));
	return { rules, nodeNames };
}

type TypeFilter = 'all' | 'allow' | 'deny';

type AddForm = {
	domains: string;
	type: RuleType;
	kind: RuleKind;
	comment: string;
};

const DEFAULT_ADD_FORM: AddForm = { domains: '', type: 'deny', kind: 'exact', comment: '' };

export function Domains() {
	const [rules, setRules] = useState<ConsolidatedRule[]>([]);
	const [nodeNames, setNodeNames] = useState<Map<number, string>>(new Map());
	const [loading, setLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [typeFilter, setTypeFilter] = useState<TypeFilter>('all');

	const [addOpen, setAddOpen] = useState(false);
	const [addForm, setAddForm] = useState<AddForm>(DEFAULT_ADD_FORM);
	const [addSubmitting, setAddSubmitting] = useState(false);
	const [addError, setAddError] = useState<string | null>(null);

	const [removeConfirm, setRemoveConfirm] = useState<string | null>(null);
	const [removing, setRemoving] = useState<string | null>(null);
	const [removeError, setRemoveError] = useState<string | null>(null);

	const fetchRules = useCallback(async () => {
		setLoading(true);
		setError(null);
		try {
			const resp = await listDomainRules();
			const { rules: consolidated, nodeNames: names } = consolidate(resp);
			setRules(consolidated);
			setNodeNames(names);
		} catch (err) {
			setError(err instanceof Error ? err.message : 'Failed to load domain rules');
		} finally {
			setLoading(false);
		}
	}, []);

	useEffect(() => { fetchRules(); }, [fetchRules]);

	async function handleAdd() {
		setAddError(null);
		const rawDomains = addForm.domains.split('\n').map(d => d.trim()).filter(Boolean);
		if (rawDomains.length === 0) {
			setAddError('Enter at least one domain');
			return;
		}
		setAddSubmitting(true);
		try {
			await addDomainRule(
				addForm.type,
				addForm.kind,
				rawDomains.length === 1 ? rawDomains[0] : rawDomains,
				addForm.comment || undefined,
			);
			setAddOpen(false);
			setAddForm(DEFAULT_ADD_FORM);
			await fetchRules();
		} catch (err) {
			setAddError(err instanceof Error ? err.message : 'Failed to add rule');
		} finally {
			setAddSubmitting(false);
		}
	}

	async function handleRemove(rule: ConsolidatedRule) {
		setRemoveError(null);
		setRemoving(rule.key);
		try {
			await removeDomainRule(rule.type, rule.kind, rule.domain);
			setRemoveConfirm(null);
			await fetchRules();
		} catch (err) {
			setRemoveError(err instanceof Error ? err.message : 'Failed to remove rule');
		} finally {
			setRemoving(null);
		}
	}

	const filtered = typeFilter === 'all' ? rules : rules.filter(r => r.type === typeFilter);

	return (
		<div className={styles.page}>
			<div className={styles.toolbar}>
				<div className={styles.typeFilter} role="group" aria-label="Filter by type">
					{(['all', 'allow', 'deny'] as const).map(t => (
						<button
							key={t}
							type="button"
							className={styles.filterBtn}
							data-active={typeFilter === t || undefined}
							data-type={t !== 'all' ? t : undefined}
							onClick={() => setTypeFilter(t)}
						>
							{t === 'all' ? 'All' : t === 'allow' ? 'Allowlist' : 'Blocklist'}
						</button>
					))}
				</div>

				<div className={styles.toolbarRight}>
					<button
						type="button"
						className={styles.refreshBtn}
						onClick={fetchRules}
						disabled={loading}
						aria-label="Refresh"
					>
						<RefreshCw size={16} className={loading ? styles.spin : undefined} />
					</button>

					<Dialog.Root open={addOpen} onOpenChange={(next) => { setAddOpen(next); if (!next) { setAddForm(DEFAULT_ADD_FORM); setAddError(null); } }}>
						<Dialog.Trigger asChild>
							<button type="button" className={styles.addBtn}>
								<Plus size={16} />
								Add rule
							</button>
						</Dialog.Trigger>
						<Dialog.Portal>
							<Dialog.Overlay className={styles.overlay} />
							<Dialog.Content className={styles.dialog}>
								<Dialog.Title className={styles.dialogTitle}>Add domain rule</Dialog.Title>

								<div className={styles.field}>
									<label htmlFor="add-domains" className={styles.fieldLabel}>
										Domains
										<span className={styles.fieldHint}>(one per line)</span>
									</label>
									<textarea
										id="add-domains"
										className={styles.textarea}
										rows={4}
										placeholder="example.com"
										value={addForm.domains}
										onChange={e => setAddForm(f => ({ ...f, domains: e.target.value }))}
										disabled={addSubmitting}
									/>
								</div>

								<div className={styles.fieldRow}>
									<div className={styles.field}>
										<label htmlFor="add-type" className={styles.fieldLabel}>Type</label>
										<select
											id="add-type"
											className={styles.select}
											value={addForm.type}
											onChange={e => setAddForm(f => ({ ...f, type: e.target.value as RuleType }))}
											disabled={addSubmitting}
										>
											<option value="deny">Blocklist</option>
											<option value="allow">Allowlist</option>
										</select>
									</div>

									<div className={styles.field}>
										<label htmlFor="add-kind" className={styles.fieldLabel}>Kind</label>
										<select
											id="add-kind"
											className={styles.select}
											value={addForm.kind}
											onChange={e => setAddForm(f => ({ ...f, kind: e.target.value as RuleKind }))}
											disabled={addSubmitting}
										>
											<option value="exact">Exact</option>
											<option value="regex">Regex</option>
										</select>
									</div>
								</div>

								<div className={styles.field}>
									<label htmlFor="add-comment" className={styles.fieldLabel}>
										Comment
										<span className={styles.fieldHint}>(optional)</span>
									</label>
									<input
										id="add-comment"
										type="text"
										className={styles.input}
										placeholder="e.g. Added manually"
										value={addForm.comment}
										onChange={e => setAddForm(f => ({ ...f, comment: e.target.value }))}
										disabled={addSubmitting}
									/>
								</div>

								{addError && <p className={styles.dialogError}>{addError}</p>}

								<div className={styles.dialogActions}>
									<Dialog.Close asChild>
										<button type="button" className={styles.cancelBtn} disabled={addSubmitting}>
											Cancel
										</button>
									</Dialog.Close>
									<button
										type="button"
										className={styles.submitBtn}
										onClick={handleAdd}
										disabled={addSubmitting || addForm.domains.trim() === ''}
										aria-busy={addSubmitting}
									>
										{addSubmitting ? <RefreshCw size={15} className={styles.spin} /> : null}
										Add
									</button>
								</div>

								<Dialog.Close asChild>
									<button className={styles.dialogClose} aria-label="Close">
										<X size={18} />
									</button>
								</Dialog.Close>
							</Dialog.Content>
						</Dialog.Portal>
					</Dialog.Root>
				</div>
			</div>

			{error && <div className={styles.error}>{error}</div>}
			{removeError && <div className={styles.error}>{removeError}</div>}

			{loading && rules.length === 0 && (
				<div className={styles.loadingState}>
					<RefreshCw size={20} className={styles.spin} />
					Loading…
				</div>
			)}

			{!loading && !error && filtered.length === 0 && (
				<div className={styles.emptyState}>
					{typeFilter === 'all'
						? 'No domain rules found across the cluster.'
						: `No ${typeFilter === 'allow' ? 'allowlist' : 'blocklist'} rules found.`}
				</div>
			)}

			{filtered.length > 0 && (
				<>
					<div className={styles.tableInfo}>
						{filtered.length} {filtered.length === 1 ? 'rule' : 'rules'}
						{typeFilter !== 'all' && ` in ${typeFilter === 'allow' ? 'allowlist' : 'blocklist'}`}
					</div>
					<div className={styles.tableWrap}>
						<table className={styles.table}>
							<thead>
								<tr>
									<th>Domain</th>
									<th>Type</th>
									<th>Kind</th>
									<th>Nodes</th>
									<th>Comment</th>
									<th aria-label="Actions" />
								</tr>
							</thead>
							<tbody>
								{filtered.map(rule => {
									const partial = rule.nodeIds.length < rule.totalNodes;
									const confirming = removeConfirm === rule.key;
									const isRemoving = removing === rule.key;

									const missingNodeIds = Array.from(nodeNames.keys()).filter(
										id => !rule.nodeIds.includes(id),
									);

									return (
										<tr key={rule.key} className={styles.row}>
											<td className={styles.domainCell}>{rule.domain}</td>
											<td>
												<span className={styles.typeBadge} data-type={rule.type}>
													{rule.type === 'allow' ? 'Allow' : 'Block'}
												</span>
											</td>
											<td>
												<span className={styles.kindBadge}>{rule.kind}</span>
											</td>
											<td>
												<span
													className={styles.nodesBadge}
													data-partial={partial || undefined}
													title={partial
														? `Missing from: ${missingNodeIds.map(id => nodeNames.get(id) ?? id).join(', ')}`
														: `Present on all ${rule.totalNodes} node(s)`}
												>
													{rule.nodeIds.length}/{rule.totalNodes}
												</span>
											</td>
											<td className={styles.commentCell}>{rule.comment || '—'}</td>
											<td className={styles.actionCell}>
												{confirming ? (
													<div className={styles.confirmRow}>
														<span className={styles.confirmText}>Remove?</span>
														<button
															type="button"
															className={styles.confirmYesBtn}
															onClick={() => handleRemove(rule)}
															disabled={isRemoving}
														>
															{isRemoving ? <RefreshCw size={13} className={styles.spin} /> : 'Yes'}
														</button>
														<button
															type="button"
															className={styles.confirmNoBtn}
															onClick={() => { setRemoveConfirm(null); setRemoveError(null); }}
															disabled={isRemoving}
														>
															No
														</button>
													</div>
												) : (
													<button
														type="button"
														className={styles.removeBtn}
														onClick={() => setRemoveConfirm(rule.key)}
														disabled={isRemoving}
														aria-label={`Remove ${rule.domain}`}
													>
														<Trash2 size={15} />
													</button>
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
		</div>
	);
}
